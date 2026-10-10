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

package clusapi

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"
)

// CLUSTER_ENUM values for ClusterOpenEnum.
// https://learn.microsoft.com/en-us/windows/win32/api/clusapi/ne-clusapi-cluster_enum
const (
	enumNode     = 0x01
	enumResource = 0x04
	enumGroup    = 0x08
	enumNetwork  = 0x10
)

// CLUS_OBJECT_* identifiers and CLCTL_* operations. A CLUSCTL_* control code is
// (object << 24) | operation, for example CLUSCTL_GROUP_GET_FLAGS = 0x03000009.
// https://learn.microsoft.com/en-us/previous-versions/windows/desktop/mscs/control-code-architecture
const (
	objectGroup   = 3
	objectNode    = 4
	objectNetwork = 5

	ctlGetCharacteristics    = 0x05
	ctlGetFlags              = 0x09
	ctlGetROCommonProperties = 0x55
	ctlGetCommonProperties   = 0x59
)

// The Ex open functions and the control/state functions below are available
// since Windows Server 2008 R2, so they cover Windows Server 2012 R2 and 2016.
// ClusterOpenEnumEx and the per-type *OpenEnumEx functions (Windows Server 2016+)
// are deliberately not used.
//
//nolint:gochecknoglobals
var (
	openGroup    = dll.NewProc("OpenClusterGroupEx")
	closeGroup   = dll.NewProc("CloseClusterGroup")
	groupControl = dll.NewProc("ClusterGroupControl")
	groupState   = dll.NewProc("GetClusterGroupState")

	openNode    = dll.NewProc("OpenClusterNodeEx")
	closeNode   = dll.NewProc("CloseClusterNode")
	nodeControl = dll.NewProc("ClusterNodeControl")
	nodeState   = dll.NewProc("GetClusterNodeState")

	openNetwork    = dll.NewProc("OpenClusterNetworkEx")
	closeNetwork   = dll.NewProc("CloseClusterNetwork")
	networkControl = dll.NewProc("ClusterNetworkControl")
	networkState   = dll.NewProc("GetClusterNetworkState")
)

// Object is a cluster group, node or network. Values contains the 32-bit
// common properties plus Characteristics, Flags and State. A missing key means
// the value could not be read; it is never substituted with zero.
type Object struct {
	Name string
	// OwnerNode is the hosting node of a group. OwnerNodeValid is false when
	// the group state could not be read or the object has no owner concept.
	OwnerNode      string
	OwnerNodeValid bool
	Values         map[string]uint32
}

type objectAPI struct {
	kind     string
	enumType uint32
	object   uint32
	open     *windows.LazyProc
	close    *windows.LazyProc
	control  *windows.LazyProc
	// state reads the object state. It returns the owner node name for groups.
	state func(handle uintptr, deadline time.Time) (uint32, string, bool, error)
}

//nolint:gochecknoglobals
var groupAPI = objectAPI{
	kind:     "group",
	enumType: enumGroup,
	object:   objectGroup,
	open:     openGroup,
	close:    closeGroup,
	control:  groupControl,
	state:    readGroupState,
}

//nolint:gochecknoglobals
var nodeAPI = objectAPI{
	kind:     "node",
	enumType: enumNode,
	object:   objectNode,
	open:     openNode,
	close:    closeNode,
	control:  nodeControl,
	state:    simpleState(nodeState),
}

//nolint:gochecknoglobals
var networkAPI = objectAPI{
	kind:     "network",
	enumType: enumNetwork,
	object:   objectNetwork,
	open:     openNetwork,
	close:    closeNetwork,
	control:  networkControl,
	state:    simpleState(networkState),
}

// Networks reads every cluster network; see Groups for the result contract.
func (c *Cluster) Networks(deadline time.Time) ([]Object, error) {
	return c.objects(&networkAPI, deadline)
}

// Nodes reads every cluster node; see Groups for the result contract.
func (c *Cluster) Nodes(deadline time.Time) ([]Object, error) {
	return c.objects(&nodeAPI, deadline)
}

// Groups reads every cluster group. Successful groups and properties are
// retained alongside joined failures; see Resources for the deadline contract.
func (c *Cluster) Groups(deadline time.Time) ([]Object, error) {
	return c.objects(&groupAPI, deadline)
}

func (c *Cluster) objects(api *objectAPI, deadline time.Time) ([]Object, error) {
	for _, proc := range []*windows.LazyProc{openCluster, openEnum, nextEnum, closeEnum, api.open, api.close, api.control} {
		if err := proc.Find(); err != nil {
			return nil, fmt.Errorf("load ClusAPI: %w", err)
		}
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if err := c.openLocked(deadline); err != nil {
		return nil, err
	}

	var objects []Object

	err := c.enumerateLocked(api.enumType, api.kind, deadline, func(name string) error {
		object, err := c.readObject(api, name, deadline)
		objects = append(objects, object)

		return err
	})

	return objects, err
}

// enumerateLocked calls visit for every object name of a CLUSTER_ENUM type.
// Visit errors are joined per object; the walk stops at the deadline. The
// caller holds c.mu and has opened the cluster handle.
func (c *Cluster) enumerateLocked(enumType uint32, kind string, deadline time.Time, visit func(name string) error) (resultErr error) {
	if err := checkDeadline(deadline); err != nil {
		return err
	}

	enum, _, err := openEnum.Call(c.handle, uintptr(enumType))
	if enum == 0 {
		return fmt.Errorf("ClusterOpenEnum: %w", err)
	}
	defer func() {
		status, _, _ := closeEnum.Call(enum)
		if status != 0 {
			resultErr = errors.Join(resultErr, fmt.Errorf("ClusterCloseEnum: %w", windows.Errno(status)))
		}
	}()

	for index := uint32(0); ; index++ {
		name, err := enumName(enum, index, enumType, deadline)
		if errors.Is(err, windows.ERROR_NO_MORE_ITEMS) {
			return resultErr
		}

		if err != nil {
			return errors.Join(resultErr, fmt.Errorf("ClusterEnum: %w", err))
		}

		err = visit(name)
		if err != nil {
			resultErr = errors.Join(resultErr, fmt.Errorf("%s %q: %w", kind, name, err))
		}

		if errors.Is(err, context.DeadlineExceeded) {
			return resultErr
		}
	}
}

// openLocked opens the cluster handle on first use. The caller holds c.mu.
func (c *Cluster) openLocked(deadline time.Time) error {
	if c.closed {
		return errors.New("cluster is closed")
	}

	if err := checkDeadline(deadline); err != nil {
		return err
	}

	if c.handle != 0 {
		return nil
	}

	handle, _, err := openCluster.Call(0, windows.GENERIC_READ, 0)
	if handle == 0 {
		return fmt.Errorf("OpenClusterEx: %w", err)
	}

	c.handle = handle

	return nil
}

func (c *Cluster) readObject(api *objectAPI, name string, deadline time.Time) (_ Object, resultErr error) {
	object := Object{Name: name, Values: make(map[string]uint32)}
	if err := checkDeadline(deadline); err != nil {
		return object, err
	}

	namePtr, err := windows.UTF16PtrFromString(name)
	if err != nil {
		return object, err
	}

	handle, _, err := api.open.Call(c.handle, uintptr(unsafe.Pointer(namePtr)), windows.GENERIC_READ, 0)
	if handle == 0 {
		return object, fmt.Errorf("%s: %w", api.open.Name, err)
	}
	defer func() {
		result, _, err := api.close.Call(handle)
		if result == 0 {
			resultErr = errors.Join(resultErr, fmt.Errorf("%s: %w", api.close.Name, err))
		}
	}()

	if err := readCommonProperties(api.control, handle, api.object, object.Values, deadline); err != nil {
		resultErr = errors.Join(resultErr, err)

		if errors.Is(err, context.DeadlineExceeded) {
			return object, resultErr
		}
	}

	for _, field := range []struct {
		name      string
		operation uint32
	}{{"Characteristics", ctlGetCharacteristics}, {"Flags", ctlGetFlags}} {
		data, err := objectBuffer(api.control, handle, api.object<<24|field.operation, deadline)

		switch {
		case err != nil:
			resultErr = errors.Join(resultErr, fmt.Errorf("%s: %w", field.name, err))

			if errors.Is(err, context.DeadlineExceeded) {
				return object, resultErr
			}
		case len(data) != 4:
			resultErr = errors.Join(resultErr, fmt.Errorf("%s: invalid DWORD size %d", field.name, len(data)))
		default:
			object.Values[field.name] = binary.LittleEndian.Uint32(data)
		}
	}

	state, owner, ownerValid, err := api.state(handle, deadline)
	if err != nil {
		return object, errors.Join(resultErr, fmt.Errorf("state: %w", err))
	}

	object.Values["State"] = state
	object.OwnerNode = owner
	object.OwnerNodeValid = ownerValid

	return object, resultErr
}

// readCommonProperties stores the single 32-bit values of the read-only and
// read/write common property lists in values. Other formats are skipped.
func readCommonProperties(control *windows.LazyProc, handle uintptr, object uint32, values map[string]uint32, deadline time.Time) error {
	var resultErr error

	for _, operation := range []uint32{ctlGetROCommonProperties, ctlGetCommonProperties} {
		code := object<<24 | operation

		data, err := objectBuffer(control, handle, code, deadline)
		if err != nil {
			resultErr = errors.Join(resultErr, fmt.Errorf("%s %#x: %w", control.Name, code, err))

			if errors.Is(err, context.DeadlineExceeded) {
				return resultErr
			}

			continue
		}

		properties, err := ParseProperties(data)
		if err != nil {
			resultErr = errors.Join(resultErr, fmt.Errorf("%s %#x: %w", control.Name, code, err))

			continue
		}

		for name, list := range properties {
			if len(list) != 1 {
				continue
			}

			if value, err := list[0].Value32(); err == nil {
				values[name] = value
			}
		}
	}

	return resultErr
}

// objectBuffer calls a Cluster{Group,Node,Network}Control function. They share
// the ClusterResourceControl signature and return a DWORD status code.
func objectBuffer(proc *windows.LazyProc, handle uintptr, code uint32, deadline time.Time) ([]byte, error) {
	return controlBuffer(deadline, func(buffer []byte) (uint32, error) {
		var (
			size    uint32
			pointer *byte
		)
		if len(buffer) != 0 {
			pointer = &buffer[0]
		}

		status, _, _ := proc.Call(handle, 0, uintptr(code), 0, 0, uintptr(unsafe.Pointer(pointer)), uintptr(len(buffer)), uintptr(unsafe.Pointer(&size)))
		if status != 0 {
			return size, windows.Errno(status)
		}

		return size, nil
	})
}

// readGroupState calls GetClusterGroupState, which returns
// ClusterGroupStateUnknown (-1) and sets the last error on failure.
// https://learn.microsoft.com/en-us/windows/win32/api/clusapi/nf-clusapi-getclustergroupstate
func readGroupState(handle uintptr, deadline time.Time) (uint32, string, bool, error) {
	if err := groupState.Find(); err != nil {
		return 0, "", false, err
	}

	// stateBuffers validates a second name; groups have only an owner node, so
	// the unused buffer stays empty and terminated.
	state, node, _, err := stateBuffers(deadline, func(node, _ []uint16) (uint32, uint32, uint32, error) {
		nodeLength := uint32(len(node))

		state, _, err := groupState.Call(handle, uintptr(unsafe.Pointer(&node[0])), uintptr(unsafe.Pointer(&nodeLength)))

		return uint32(state), nodeLength, 0, unknownStateError(uint32(state), err)
	})
	if err != nil {
		return 0, "", false, err
	}

	return state, node, true, nil
}

// simpleState wraps GetClusterNodeState and GetClusterNetworkState. Both take
// only the object handle and return -1 (unknown) with the last error on failure.
func simpleState(proc *windows.LazyProc) func(uintptr, time.Time) (uint32, string, bool, error) {
	return func(handle uintptr, deadline time.Time) (uint32, string, bool, error) {
		if err := proc.Find(); err != nil {
			return 0, "", false, err
		}

		if err := checkDeadline(deadline); err != nil {
			return 0, "", false, err
		}

		state, _, err := proc.Call(handle)
		if err := unknownStateError(uint32(state), err); err != nil {
			return 0, "", false, err
		}

		return uint32(state), "", false, nil
	}
}

// unknownStateError converts the last error of a Get*State call that returned
// the documented failure value -1. A success code still marks a failed call.
func unknownStateError(state uint32, err error) error {
	if state != ^uint32(0) {
		return nil
	}

	var errno windows.Errno
	if errors.As(err, &errno) && errno != 0 {
		return errno
	}

	return errors.New("state is unknown")
}
