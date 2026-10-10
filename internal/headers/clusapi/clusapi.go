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
	stateUnknown                  = ^uint32(0) // ClusterResourceStateUnknown
	propertyListBufferSize        = 4 << 10
	typeBufferSize                = 512
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

// ErrBusy reports that an earlier call on the same Cluster has not returned
// yet. ClusAPI RPCs cannot be cancelled; a call that outlived its scrape keeps
// the cluster handle until the RPC returns, and later calls fail fast instead
// of queuing behind it.
var ErrBusy = errors.New("previous ClusAPI collection is still running")

// nativeAPI holds the ClusAPI calls that drive the handle lifecycle and the
// enumeration. Tests substitute it to exercise these paths without a cluster.
type nativeAPI struct {
	openCluster   func() (uintptr, error)
	closeCluster  func(cluster uintptr) error
	openEnum      func(cluster uintptr, enumType uint32) (uintptr, error)
	closeEnum     func(enum uintptr) error
	enumName      func(enum uintptr, index, enumType uint32, deadline time.Time) (objectName, error)
	openResource  func(cluster uintptr, name objectName) (uintptr, error)
	closeResource func(resource uintptr) error
	readResource  func(resource uintptr, out *Resource, deadline time.Time) error
}

//nolint:gochecknoglobals
var clusAPI = nativeAPI{
	openCluster: func() (uintptr, error) {
		handle, _, err := openCluster.Call(0, windows.GENERIC_READ, 0)
		if handle == 0 {
			return 0, callError(err)
		}

		return handle, nil
	},
	closeCluster: func(cluster uintptr) error {
		if result, _, err := closeCluster.Call(cluster); result == 0 {
			return callError(err)
		}

		return nil
	},
	openEnum: func(cluster uintptr, enumType uint32) (uintptr, error) {
		enum, _, err := openEnum.Call(cluster, uintptr(enumType))
		if enum == 0 {
			return 0, callError(err)
		}

		return enum, nil
	},
	closeEnum: func(enum uintptr) error {
		if status, _, _ := closeEnum.Call(enum); status != 0 {
			return windows.Errno(status)
		}

		return nil
	},
	enumName: enumName,
	openResource: func(cluster uintptr, name objectName) (uintptr, error) {
		return openObject(openResource, cluster, name)
	},
	closeResource: func(resource uintptr) error {
		return closeObject(closeResource, resource)
	},
	readResource: readResource,
}

// callError returns the last error of a failed handle or BOOL call. Go clears
// the thread's last error before each call, so 0 means the API set none.
func callError(err error) error {
	if errors.Is(err, windows.Errno(0)) {
		return errors.New("failed without setting a Windows error code")
	}

	return err
}

// openObject calls one of the OpenCluster*Ex functions, which share a signature.
func openObject(proc *windows.LazyProc, cluster uintptr, name objectName) (uintptr, error) {
	handle, _, err := proc.Call(cluster, uintptr(unsafe.Pointer(&name.units[0])), windows.GENERIC_READ, 0)
	if handle == 0 {
		return 0, callError(err)
	}

	return handle, nil
}

// closeObject calls one of the CloseCluster* functions, which return a BOOL.
func closeObject(proc *windows.LazyProc, handle uintptr) error {
	if result, _, err := proc.Call(handle); result == 0 {
		return callError(err)
	}

	return nil
}

// Cluster owns a cluster handle. Only one call runs at a time: while a call
// runs it owns the handle, and Close leaves the handle to that call, which
// closes it when its RPC returns. This keeps a hung RPC from blocking Close.
type Cluster struct {
	api *nativeAPI

	mu      sync.Mutex // guards closed and running, and handle when not running
	handle  uintptr
	closed  bool
	running bool
}

type Resource struct {
	Name          string
	Type          string
	OwnerGroup    string
	OwnerNode     string
	IdentityValid bool
	Values        map[string]uint32
}

// Open checks that the local node is a configured cluster member. A missing
// clusapi.dll or export reports errors.ErrUnsupported, like a node that is not
// part of a cluster.
func Open() (*Cluster, error) {
	for _, proc := range []*windows.LazyProc{getClusterState, openCluster, closeCluster, openEnum, nextEnum, closeEnum, openResource, closeResource, resourceControl, resourceState} {
		if err := proc.Find(); err != nil {
			return nil, fmt.Errorf("load ClusAPI: %w: %w", errors.ErrUnsupported, err)
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
	return &Cluster{api: &clusAPI}, nil
}

func (c *Cluster) native() *nativeAPI {
	if c.api == nil {
		return &clusAPI
	}

	return c.api
}

// Close releases the cluster handle. A handle whose CloseCluster fails is
// dropped anyway, so a later Build starts from a clean state. If a call is
// still running, it closes the handle when it returns.
func (c *Cluster) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.closed = true

	if c.running || c.handle == 0 {
		return nil
	}

	handle := c.handle
	c.handle = 0

	if err := c.native().closeCluster(handle); err != nil {
		return fmt.Errorf("CloseCluster: %w", err)
	}

	return nil
}

// begin marks a call as running. A call that is still running makes it fail
// with ErrBusy. A successful begin must be paired with end.
func (c *Cluster) begin(deadline time.Time) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	switch {
	case c.closed:
		return errors.New("cluster is closed")
	case c.running:
		return ErrBusy
	}

	if err := checkDeadline(deadline); err != nil {
		return err
	}

	c.running = true

	return nil
}

// open opens the cluster handle on first use. The running call owns the
// handle, so the RPC runs without the lock.
func (c *Cluster) open() error {
	if c.handle != 0 {
		return nil
	}

	handle, err := c.native().openCluster()
	if err != nil {
		return fmt.Errorf("OpenClusterEx: %w", err)
	}

	c.handle = handle

	return nil
}

// end finishes a call started by begin. If Close ran meanwhile, the handle is
// released here and the CloseCluster error is returned.
func (c *Cluster) end() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.running = false

	if !c.closed || c.handle == 0 {
		return nil
	}

	handle := c.handle
	c.handle = 0

	if err := c.native().closeCluster(handle); err != nil {
		return fmt.Errorf("CloseCluster: %w", err)
	}

	return nil
}

// resetHandle drops a handle whose RPC binding failed, for example after a
// cluster service restart, so the next call reopens the cluster. The caller
// owns the handle through begin. The close error is irrelevant: the call
// reports the error that made the handle unusable.
func (c *Cluster) resetHandle() {
	_ = c.native().closeCluster(c.handle)
	c.handle = 0
}

// errObjectDeleted marks an object that was deleted after ClusterOpenEnum took
// its snapshot. The enumeration skips it without an error.
var errObjectDeleted = errors.New("object was deleted during the enumeration")

// enumerate calls visit for every object name of a CLUSTER_ENUM type. Visit
// errors are joined per object; the walk stops at the deadline. The caller
// owns the handle through begin.
func (c *Cluster) enumerate(enumType uint32, kind string, deadline time.Time, visit func(name objectName) error) (resultErr error) {
	if err := checkDeadline(deadline); err != nil {
		return err
	}

	api := c.native()

	enum, err := api.openEnum(c.handle, enumType)
	if err != nil {
		c.resetHandle()

		return fmt.Errorf("ClusterOpenEnum: %w", err)
	}
	defer func() {
		if err := api.closeEnum(enum); err != nil {
			resultErr = errors.Join(resultErr, fmt.Errorf("ClusterCloseEnum: %w", err))
		}
	}()

	for index := uint32(0); ; index++ {
		name, err := api.enumName(enum, index, enumType, deadline)
		if errors.Is(err, windows.ERROR_NO_MORE_ITEMS) {
			return resultErr
		}

		if _, ok := errors.AsType[invalidEnumEntryError](err); ok {
			// One malformed entry must not hide the objects after it.
			resultErr = errors.Join(resultErr, fmt.Errorf("ClusterEnum index %d: %w", index, err))

			continue
		}

		if err != nil {
			return errors.Join(resultErr, fmt.Errorf("ClusterEnum: %w", err))
		}

		err = visit(name)
		if errors.Is(err, errObjectDeleted) {
			continue
		}

		if err != nil {
			resultErr = errors.Join(resultErr, fmt.Errorf("%s %q: %w", kind, name.name, err))
		}

		if errors.Is(err, context.DeadlineExceeded) {
			return resultErr
		}
	}
}

// Resources retains successful resource/property results alongside joined failures.
// Windows ClusAPI RPCs cannot be cancelled. Deadline checks prevent starting
// another RPC after the supplied budget; a call that still outlives it makes
// later calls fail with ErrBusy until it returns.
func (c *Cluster) Resources(deadline time.Time) (_ []Resource, resultErr error) {
	if err := c.begin(deadline); err != nil {
		return nil, err
	}
	defer func() { resultErr = errors.Join(resultErr, c.end()) }()

	if err := c.open(); err != nil {
		return nil, err
	}

	api := c.native()

	var resources []Resource

	err := c.enumerate(enumResource, "resource", deadline, func(name objectName) error {
		if err := checkDeadline(deadline); err != nil {
			return err
		}

		handle, err := api.openResource(c.handle, name)
		if errors.Is(err, windows.ERROR_RESOURCE_NOT_FOUND) {
			return errObjectDeleted
		}

		resource := Resource{Name: name.name, Values: make(map[string]uint32)}

		if err != nil {
			resources = append(resources, resource)

			return fmt.Errorf("OpenClusterResourceEx: %w", err)
		}

		err = api.readResource(handle, &resource, deadline)
		if closeErr := api.closeResource(handle); closeErr != nil {
			err = errors.Join(err, fmt.Errorf("CloseClusterResource: %w", closeErr))
		}

		resources = append(resources, resource)

		return err
	})

	return resources, err
}

// invalidEnumEntryError reports a ClusterEnum entry that violates the API
// contract. Enumeration continues with the next index.
type invalidEnumEntryError string

func (e invalidEnumEntryError) Error() string { return string(e) }

// objectName keeps the NUL-terminated native name for opening the resource
// next to the decoded label value.
type objectName struct {
	units []uint16
	name  string
}

func enumName(enum uintptr, index, enumType uint32, deadline time.Time) (objectName, error) {
	buffer := make([]uint16, 256)

	for range 16 {
		if err := checkDeadline(deadline); err != nil {
			return objectName{}, err
		}

		length := uint32(len(buffer))

		var objectType uint32

		status, _, _ := nextEnum.Call(enum, uintptr(index), uintptr(unsafe.Pointer(&objectType)), uintptr(unsafe.Pointer(&buffer[0])), uintptr(unsafe.Pointer(&length)))
		if windows.Errno(status) == windows.ERROR_MORE_DATA {
			if length >= maxBufferSize/2 || int(length)+1 <= len(buffer) {
				return objectName{}, errors.New("invalid ClusterEnum buffer size")
			}

			buffer = make([]uint16, int(length)+1)

			continue
		}

		if status != 0 {
			return objectName{}, windows.Errno(status)
		}

		if objectType != enumType || length >= uint32(len(buffer)) {
			return objectName{}, invalidEnumEntryError(fmt.Sprintf("invalid ClusterEnum entry: type %d, length %d", objectType, length))
		}

		units := make([]uint16, length+1)
		copy(units, buffer[:length])

		return objectName{units: units, name: decodeUnits(buffer[:length])}, nil
	}

	return objectName{}, errors.New("ClusterEnum buffer did not stabilize")
}

// readResource fills the properties of an open resource. Failed properties are
// omitted; the identity is only valid once the state and owner are known.
func readResource(handle uintptr, resource *Resource, deadline time.Time) error {
	var resultErr error

	hasType := false

	for _, code := range []uint32{resourceGetROCommonProperties, resourceGetCommonProperties} {
		data, err := resourceBuffer(handle, code, propertyListBufferSize, deadline)
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

		// The read-only common Type property is the same resource type name
		// that CLUSCTL_RESOURCE_GET_RESOURCE_TYPE returns.
		if values := properties["Type"]; len(values) == 1 {
			if value, err := values[0].String(); err == nil {
				resource.Type = value
				hasType = true
			}
		}
	}

	for _, field := range []struct {
		name string
		code uint32
	}{{"Characteristics", resourceGetCharacteristics}, {"Flags", resourceGetFlags}} {
		data, err := resourceBuffer(handle, field.code, 4, deadline)
		switch {
		case err != nil:
			resultErr = errors.Join(resultErr, fmt.Errorf("%s: %w", field.name, err))
		case len(data) != 4:
			resultErr = errors.Join(resultErr, fmt.Errorf("%s: invalid DWORD size", field.name))
		default:
			resource.Values[field.name] = binary.LittleEndian.Uint32(data)
		}
	}

	data, err := resourceBuffer(handle, resourceGetClassInfo, 8, deadline)
	switch {
	case err != nil:
		resultErr = errors.Join(resultErr, fmt.Errorf("class information: %w", err))
	case len(data) != 8:
		resultErr = errors.Join(resultErr, errors.New("invalid resource class information size"))
	default:
		resource.Values["ResourceClass"] = binary.LittleEndian.Uint32(data)
		resource.Values["Subclass"] = binary.LittleEndian.Uint32(data[4:])
	}

	// The property lists normally carry the type; only ask separately without it.
	if !hasType {
		data, err = resourceBuffer(handle, resourceGetType, typeBufferSize, deadline)
		if err != nil {
			return errors.Join(resultErr, fmt.Errorf("resource type: %w", err))
		}

		resource.Type = decodeString(data)
	}

	state, owner, group, err := readResourceState(handle, deadline)
	if err != nil {
		return errors.Join(resultErr, fmt.Errorf("GetClusterResourceState: %w", err))
	}

	resource.Values["State"] = state
	resource.OwnerNode = owner
	resource.OwnerGroup = group
	resource.IdentityValid = true

	return resultErr
}

func resourceBuffer(handle uintptr, code uint32, size int, deadline time.Time) ([]byte, error) {
	return controlBuffer(deadline, size, func(buffer []byte) (uint32, error) {
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
		// Only ERROR_MORE_DATA matters: the caller retries with larger buffers.
		if uint32(state) == stateUnknown && errors.Is(err, windows.ERROR_MORE_DATA) {
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

		// ClusterResourceStateUnknown is a state WMI published as 4294967295
		// together with the other properties. A failed query may leave the
		// names and lengths untouched; the buffers start zeroed, so decoding up
		// to the first NUL yields whatever names were returned.
		if state == stateUnknown {
			return state, decodeUnits(node), decodeUnits(group), nil
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

// controlBuffer starts with a buffer of the expected size, so most controls need
// one RPC. ERROR_MORE_DATA still grows the buffer to the reported size; size 0
// starts with a NULL size probe.
func controlBuffer(deadline time.Time, size int, call func([]byte) (uint32, error)) ([]byte, error) {
	var buffer []byte
	if size > 0 {
		buffer = make([]byte, size)
	}

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
