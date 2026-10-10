// SPDX-License-Identifier: Apache-2.0
//
// Copyright The Prometheus Authors
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

//go:build windows

// Package clusapi exposes read-only access to local failover cluster resources.
package clusapi

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"sync"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"
)

const (
	resourceGetCharacteristics    = 0x01000005
	resourceGetFlags              = 0x01000009
	resourceGetROCommonProperties = 0x01000055
	resourceGetCommonProperties   = 0x01000059
	resourceGetType               = 0x0100002d
	resourceGetClassInfo          = 0x0100000d
	maxBufferSize                 = 64 << 20
)

//nolint:gochecknoglobals
var (
	dll             = windows.NewLazySystemDLL("clusapi.dll")
	getClusterState = dll.NewProc("GetNodeClusterState")
	openCluster     = dll.NewProc("OpenClusterEx")
	closeCluster    = dll.NewProc("CloseCluster")
	openEnum        = dll.NewProc("ClusterOpenEnum")
	nextEnum        = dll.NewProc("ClusterEnum")
	closeEnum       = dll.NewProc("ClusterCloseEnum")
	openResource    = dll.NewProc("OpenClusterResourceEx")
	closeResource   = dll.NewProc("CloseClusterResource")
	resourceControl = dll.NewProc("ClusterResourceControl")
	resourceState   = dll.NewProc("GetClusterResourceState")
)

// Cluster owns a cluster handle. Operations and closure are serialized because
// an RPC that outlives a scrape timeout must finish before its handle is freed.
type Cluster struct {
	mu     sync.Mutex
	handle uintptr
	closed bool
}

type Resource struct {
	Name          string
	Type          string
	OwnerGroup    string
	OwnerNode     string
	IdentityValid bool
	Values        map[string]uint32
}

func Open() (*Cluster, error) {
	for _, proc := range []*windows.LazyProc{getClusterState, openCluster, closeCluster, openEnum, nextEnum, closeEnum, openResource, closeResource, resourceControl, resourceState} {
		if err := proc.Find(); err != nil {
			return nil, fmt.Errorf("load ClusAPI: %w", err)
		}
	}

	var state uint32

	status, _, _ := getClusterState.Call(0, uintptr(unsafe.Pointer(&state)))
	if status != 0 {
		return nil, fmt.Errorf("GetNodeClusterState: %w", windows.Errno(status))
	}
	// Installed and configured are separate flags. A configured but stopped
	// service remains an operational failure rather than unsupported.
	if state&3 != 3 {
		return nil, fmt.Errorf("local failover cluster is not configured: %w", errors.ErrUnsupported)
	}

	// Defer the potentially blocking RPC until collection. A transient service
	// outage must not turn resource initialization into an exporter startup failure.
	return &Cluster{}, nil
}

func (c *Cluster) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.handle == 0 {
		c.closed = true

		return nil
	}

	result, _, err := closeCluster.Call(c.handle)
	if result == 0 {
		return fmt.Errorf("CloseCluster: %w", err)
	}

	c.handle = 0
	c.closed = true

	return nil
}

// Resources retains successful resource/property results alongside joined failures.
// Windows ClusAPI RPCs cannot be cancelled. Deadline checks prevent starting
// another RPC after the supplied budget; the caller must drain a late result.
func (c *Cluster) Resources(deadline time.Time) (_ []Resource, resultErr error) {
	var resources []Resource

	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return nil, errors.New("cluster is closed")
	}

	if err := checkDeadline(deadline); err != nil {
		return nil, err
	}

	if c.handle == 0 {
		handle, _, err := openCluster.Call(0, windows.GENERIC_READ, 0)
		if handle == 0 {
			return nil, fmt.Errorf("OpenClusterEx: %w", err)
		}

		c.handle = handle
	}

	if err := checkDeadline(deadline); err != nil {
		return nil, err
	}

	enum, _, err := openEnum.Call(c.handle, 4) // CLUSTER_ENUM_RESOURCE
	if enum == 0 {
		return nil, fmt.Errorf("ClusterOpenEnum: %w", err)
	}
	defer func() {
		status, _, _ := closeEnum.Call(enum)
		if status != 0 {
			resultErr = errors.Join(resultErr, fmt.Errorf("ClusterCloseEnum: %w", windows.Errno(status)))
		}
	}()

	for index := uint32(0); ; index++ {
		name, err := enumName(enum, index, deadline)
		if errors.Is(err, windows.ERROR_NO_MORE_ITEMS) {
			return resources, resultErr
		}

		var invalid invalidEnumEntryError
		if errors.As(err, &invalid) {
			// One malformed entry must not hide the resources after it.
			resultErr = errors.Join(resultErr, fmt.Errorf("ClusterEnum index %d: %w", index, err))

			continue
		}

		if err != nil {
			return resources, errors.Join(resultErr, fmt.Errorf("ClusterEnum: %w", err))
		}

		resource, err := c.readResource(name, deadline)
		resources = append(resources, resource)

		if err != nil {
			resultErr = errors.Join(resultErr, fmt.Errorf("resource %q: %w", name.name, err))
		}

		if errors.Is(err, context.DeadlineExceeded) {
			return resources, resultErr
		}
	}
}

// invalidEnumEntryError reports a ClusterEnum entry that violates the API
// contract. Enumeration continues with the next index.
type invalidEnumEntryError string

func (e invalidEnumEntryError) Error() string { return string(e) }

// resourceName keeps the NUL-terminated native name for opening the resource
// next to the decoded label value.
type resourceName struct {
	units []uint16
	name  string
}

func enumName(enum uintptr, index uint32, deadline time.Time) (resourceName, error) {
	buffer := make([]uint16, 256)

	for range 16 {
		if err := checkDeadline(deadline); err != nil {
			return resourceName{}, err
		}

		length := uint32(len(buffer))

		var objectType uint32

		status, _, _ := nextEnum.Call(enum, uintptr(index), uintptr(unsafe.Pointer(&objectType)), uintptr(unsafe.Pointer(&buffer[0])), uintptr(unsafe.Pointer(&length)))
		if windows.Errno(status) == windows.ERROR_MORE_DATA {
			if length >= maxBufferSize/2 || int(length)+1 <= len(buffer) {
				return resourceName{}, errors.New("invalid ClusterEnum buffer size")
			}

			buffer = make([]uint16, int(length)+1)

			continue
		}

		if status != 0 {
			return resourceName{}, windows.Errno(status)
		}

		if objectType != 4 || length >= uint32(len(buffer)) {
			return resourceName{}, invalidEnumEntryError(fmt.Sprintf("invalid ClusterEnum entry: type %d, length %d", objectType, length))
		}

		units := make([]uint16, length+1)
		copy(units, buffer[:length])

		return resourceName{units: units, name: decodeUnits(buffer[:length])}, nil
	}

	return resourceName{}, errors.New("ClusterEnum buffer did not stabilize")
}

func (c *Cluster) readResource(name resourceName, deadline time.Time) (_ Resource, resultErr error) {
	resource := Resource{Name: name.name, Values: make(map[string]uint32)}
	if err := checkDeadline(deadline); err != nil {
		return resource, err
	}

	handle, _, err := openResource.Call(c.handle, uintptr(unsafe.Pointer(&name.units[0])), windows.GENERIC_READ, 0)
	if handle == 0 {
		return resource, fmt.Errorf("OpenClusterResourceEx: %w", err)
	}
	defer func() {
		result, _, err := closeResource.Call(handle)
		if result == 0 {
			resultErr = errors.Join(resultErr, fmt.Errorf("CloseClusterResource: %w", err))
		}
	}()

	for _, code := range []uint32{resourceGetROCommonProperties, resourceGetCommonProperties} {
		data, err := controlBuffer(deadline, func(buffer []byte) (uint32, error) {
			var (
				size    uint32
				pointer *byte
			)
			if len(buffer) != 0 {
				pointer = &buffer[0]
			}

			status, _, _ := resourceControl.Call(handle, 0, uintptr(code), 0, 0, uintptr(unsafe.Pointer(pointer)), uintptr(len(buffer)), uintptr(unsafe.Pointer(&size)))
			if status != 0 {
				return size, windows.Errno(status)
			}

			return size, nil
		})
		if err != nil {
			resultErr = errors.Join(resultErr, fmt.Errorf("ClusterResourceControl %#x: %w", code, err))

			continue
		}

		properties, err := ParseProperties(data)
		if err != nil {
			resultErr = errors.Join(resultErr, err)

			continue
		}

		for name, values := range properties {
			if len(values) != 1 {
				continue
			}

			if value, err := values[0].DWORD(); err == nil {
				resource.Values[name] = value
			}
		}
	}

	for _, field := range []struct {
		name string
		code uint32
	}{{"Characteristics", resourceGetCharacteristics}, {"Flags", resourceGetFlags}} {
		data, err := resourceBuffer(handle, field.code, deadline)
		switch {
		case err != nil:
			resultErr = errors.Join(resultErr, fmt.Errorf("%s: %w", field.name, err))
		case len(data) != 4:
			resultErr = errors.Join(resultErr, fmt.Errorf("%s: invalid DWORD size", field.name))
		default:
			resource.Values[field.name] = binary.LittleEndian.Uint32(data)
		}
	}

	data, err := resourceBuffer(handle, resourceGetClassInfo, deadline)
	switch {
	case err != nil:
		resultErr = errors.Join(resultErr, fmt.Errorf("class information: %w", err))
	case len(data) != 8:
		resultErr = errors.Join(resultErr, errors.New("invalid resource class information size"))
	default:
		resource.Values["ResourceClass"] = binary.LittleEndian.Uint32(data)
		resource.Values["Subclass"] = binary.LittleEndian.Uint32(data[4:])
	}

	data, err = resourceBuffer(handle, resourceGetType, deadline)
	if err != nil {
		return resource, errors.Join(resultErr, fmt.Errorf("resource type: %w", err))
	}

	resource.Type = decodeString(data)

	state, owner, group, err := readResourceState(handle, deadline)
	if err != nil {
		return resource, errors.Join(resultErr, fmt.Errorf("GetClusterResourceState: %w", err))
	}

	resource.Values["State"] = state
	resource.OwnerNode = owner
	resource.OwnerGroup = group
	resource.IdentityValid = true

	return resource, resultErr
}

func resourceBuffer(handle uintptr, code uint32, deadline time.Time) ([]byte, error) {
	return controlBuffer(deadline, func(buffer []byte) (uint32, error) {
		var (
			size    uint32
			pointer *byte
		)
		if len(buffer) != 0 {
			pointer = &buffer[0]
		}

		status, _, _ := resourceControl.Call(handle, 0, uintptr(code), 0, 0, uintptr(unsafe.Pointer(pointer)), uintptr(len(buffer)), uintptr(unsafe.Pointer(&size)))
		if status != 0 {
			return size, windows.Errno(status)
		}

		return size, nil
	})
}

func readResourceState(handle uintptr, deadline time.Time) (uint32, string, string, error) {
	return stateBuffers(deadline, func(node, group []uint16) (uint32, uint32, uint32, error) {
		nodeLength, groupLength := uint32(len(node)), uint32(len(group))

		state, _, err := resourceState.Call(handle, uintptr(unsafe.Pointer(&node[0])), uintptr(unsafe.Pointer(&nodeLength)), uintptr(unsafe.Pointer(&group[0])), uintptr(unsafe.Pointer(&groupLength)))
		if uint32(state) == ^uint32(0) {
			return uint32(state), nodeLength, groupLength, err
		}

		return uint32(state), nodeLength, groupLength, nil
	})
}

func stateBuffers(deadline time.Time, call func([]uint16, []uint16) (uint32, uint32, uint32, error)) (uint32, string, string, error) {
	node, group := make([]uint16, 256), make([]uint16, 256)

	for range 16 {
		if err := checkDeadline(deadline); err != nil {
			return 0, "", "", err
		}

		state, nodeLength, groupLength, err := call(node, group)
		if errors.Is(err, windows.ERROR_MORE_DATA) {
			if nodeLength >= maxBufferSize/2 || groupLength >= maxBufferSize/2 {
				return 0, "", "", errors.New("invalid resource state name size")
			}

			grew := false

			if int(nodeLength)+1 > len(node) {
				node = make([]uint16, int(nodeLength)+1)
				grew = true
			}

			if int(groupLength)+1 > len(group) {
				group = make([]uint16, int(groupLength)+1)
				grew = true
			}

			if !grew {
				return 0, "", "", errors.New("resource state name buffer did not grow")
			}

			continue
		}

		if err != nil {
			return state, "", "", err
		}

		if uint64(nodeLength) >= uint64(len(node)) || uint64(groupLength) >= uint64(len(group)) {
			return 0, "", "", errors.New("invalid resource state names")
		}

		return state, decodeUnits(node[:nodeLength]), decodeUnits(group[:groupLength]), nil
	}

	return 0, "", "", errors.New("resource state buffers did not stabilize")
}

func checkDeadline(deadline time.Time) error {
	if !deadline.IsZero() && !time.Now().Before(deadline) {
		return context.DeadlineExceeded
	}

	return nil
}

func controlBuffer(deadline time.Time, call func([]byte) (uint32, error)) ([]byte, error) {
	var buffer []byte

	for range 16 {
		if err := checkDeadline(deadline); err != nil {
			return nil, err
		}

		size, err := call(buffer)
		if errors.Is(err, windows.ERROR_MORE_DATA) {
			if size > maxBufferSize || int(size) <= len(buffer) {
				return nil, errors.New("invalid ClusterResourceControl buffer size")
			}

			buffer = make([]byte, size)

			continue
		}

		if err != nil {
			return nil, err
		}
		// NULL/zero-size probes may succeed with the required buffer size.
		if len(buffer) == 0 && size != 0 {
			if size > maxBufferSize {
				return nil, errors.New("invalid ClusterResourceControl buffer size")
			}

			buffer = make([]byte, size)

			continue
		}

		if uint64(size) > uint64(len(buffer)) {
			return nil, errors.New("ClusterResourceControl returned an oversized buffer")
		}

		return buffer[:size], nil
	}

	return nil, errors.New("ClusterResourceControl buffer did not stabilize")
}
