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

const (
	// CLUS_OBJECT_CLUSTER, used for CLUSCTL_CLUSTER_* control codes.
	objectCluster = 7

	// CLUS_RESCLASS_STORAGE.
	resourceClassStorage = 1
)

// GetClusterInformation, ClusterControl and GetClusterQuorumResource are
// available since Windows Server 2008.
//
//nolint:gochecknoglobals
var (
	clusterInformation = dll.NewProc("GetClusterInformation")
	clusterControl     = dll.NewProc("ClusterControl")
	quorumResource     = dll.NewProc("GetClusterQuorumResource")
)

// Properties reads the cluster name, the 32-bit cluster common properties and
// the quorum values (see readQuorum). A failed name read returns no object.
func (c *Cluster) Properties(deadline time.Time) (_ Object, resultErr error) {
	for _, proc := range []*windows.LazyProc{openCluster, clusterInformation, clusterControl, quorumResource, openResource, closeResource, resourceControl} {
		if err := proc.Find(); err != nil {
			return Object{}, fmt.Errorf("load ClusAPI: %w: %w", errors.ErrUnsupported, err)
		}
	}

	if err := c.begin(deadline); err != nil {
		return Object{}, err
	}
	defer func() { resultErr = errors.Join(resultErr, c.end()) }()

	if err := c.open(); err != nil {
		return Object{}, err
	}

	name, err := readClusterName(c.handle, deadline)
	if err != nil {
		// The name is the first RPC on the handle; a broken binding fails here.
		c.resetHandle()

		return Object{}, fmt.Errorf("GetClusterInformation: %w", err)
	}

	object := Object{Name: name, Values: make(map[string]uint32)}

	if err := readCommonProperties(clusterControl, c.handle, objectCluster, object.Values, deadline); err != nil {
		resultErr = errors.Join(resultErr, err)

		if errors.Is(err, context.DeadlineExceeded) {
			return object, resultErr
		}
	}

	// The common property is spelled ClusterEnforcedAntiaffinity; WMI and the
	// metric use ClusterEnforcedAntiAffinity.
	if value, exists := object.Values["ClusterEnforcedAntiaffinity"]; exists {
		object.Values["ClusterEnforcedAntiAffinity"] = value
	}

	if err := c.readQuorum(object.Values, deadline); err != nil {
		resultErr = errors.Join(resultErr, err)
	}

	return object, resultErr
}

// Quorum type values of MSCluster_Cluster.QuorumTypeValue.
const (
	quorumTypeNode             = 1
	quorumTypeFileShareWitness = 2
	quorumTypeStorage          = 3
)

// readQuorum derives QuorumLogFileSize and QuorumTypeValue from
// GetClusterQuorumResource. The log size parameter carries the quorum type
// (CLUS_NODE_MAJORITY_QUORUM, CLUS_HYBRID_QUORUM or CLUS_LEGACY_QUORUM). A
// cluster without a witness has no quorum resource; otherwise the witness
// resource class decides between a disk and a file share or cloud witness.
// https://learn.microsoft.com/en-us/windows/win32/api/clusapi/nf-clusapi-setclusterquorumresource
func (c *Cluster) readQuorum(values map[string]uint32, deadline time.Time) error {
	resource, logSize, err := readQuorumResource(c.handle, deadline)
	if err != nil {
		return fmt.Errorf("GetClusterQuorumResource: %w", err)
	}

	values["QuorumLogFileSize"] = logSize

	if resource == "" {
		values["QuorumTypeValue"] = quorumTypeNode

		return nil
	}

	class, err := c.resourceClass(resource, deadline)
	if err != nil {
		return fmt.Errorf("quorum resource %q: %w", resource, err)
	}

	if class == resourceClassStorage {
		values["QuorumTypeValue"] = quorumTypeStorage
	} else {
		values["QuorumTypeValue"] = quorumTypeFileShareWitness
	}

	return nil
}

// resourceClass returns the CLUS_RESOURCE_CLASS of a resource.
func (c *Cluster) resourceClass(name string, deadline time.Time) (_ uint32, resultErr error) {
	if err := checkDeadline(deadline); err != nil {
		return 0, err
	}

	namePtr, err := windows.UTF16PtrFromString(name)
	if err != nil {
		return 0, err
	}

	handle, _, err := openResource.Call(c.handle, uintptr(unsafe.Pointer(namePtr)), windows.GENERIC_READ, 0)
	if handle == 0 {
		return 0, fmt.Errorf("OpenClusterResourceEx: %w", err)
	}
	defer func() {
		result, _, err := closeResource.Call(handle)
		if result == 0 {
			resultErr = errors.Join(resultErr, fmt.Errorf("CloseClusterResource: %w", err))
		}
	}()

	data, err := resourceBuffer(handle, resourceGetClassInfo, 8, deadline)
	if err != nil {
		return 0, fmt.Errorf("class information: %w", err)
	}

	if len(data) != 8 {
		return 0, errors.New("invalid resource class information size")
	}

	return binary.LittleEndian.Uint32(data), nil
}

// readClusterName calls GetClusterInformation without CLUSTERVERSIONINFO.
// https://learn.microsoft.com/en-us/windows/win32/api/clusapi/nf-clusapi-getclusterinformation
func readClusterName(handle uintptr, deadline time.Time) (string, error) {
	_, name, _, err := stateBuffers(deadline, func(name, _ []uint16) (uint32, uint32, uint32, error) {
		length := uint32(len(name))

		status, _, _ := clusterInformation.Call(handle, uintptr(unsafe.Pointer(&name[0])), uintptr(unsafe.Pointer(&length)), 0)
		if status != 0 {
			return 0, length, 0, windows.Errno(status)
		}

		return 0, length, 0, nil
	})

	return name, err
}

// readQuorumResource returns the quorum resource name and the
// lpdwMaxQuorumLogSize value of GetClusterQuorumResource.
func readQuorumResource(handle uintptr, deadline time.Time) (string, uint32, error) {
	return quorumBuffers(deadline, func(resource, device []uint16) (uint32, uint32, uint32, error) {
		var maxLogSize uint32

		resourceLength, deviceLength := uint32(len(resource)), uint32(len(device))

		status, _, _ := quorumResource.Call(handle, uintptr(unsafe.Pointer(&resource[0])), uintptr(unsafe.Pointer(&resourceLength)), uintptr(unsafe.Pointer(&device[0])), uintptr(unsafe.Pointer(&deviceLength)), uintptr(unsafe.Pointer(&maxLogSize)))
		if status != 0 {
			return 0, resourceLength, deviceLength, windows.Errno(status)
		}

		return maxLogSize, resourceLength, deviceLength, nil
	})
}

// quorumBuffers grows the name buffers on ERROR_MORE_DATA. On a node majority
// cluster the CI fixture showed returned lengths that cover NUL characters, so
// the resource name ends at its first NUL within the reported length.
func quorumBuffers(deadline time.Time, call func([]uint16, []uint16) (uint32, uint32, uint32, error)) (string, uint32, error) {
	resource, device := make([]uint16, 256), make([]uint16, 256)

	for range 16 {
		if err := checkDeadline(deadline); err != nil {
			return "", 0, err
		}

		logSize, resourceLength, deviceLength, err := call(resource, device)
		if errors.Is(err, windows.ERROR_MORE_DATA) {
			if resourceLength >= maxBufferSize/2 || deviceLength >= maxBufferSize/2 {
				return "", 0, errors.New("invalid quorum resource name size")
			}

			grew := false

			if int(resourceLength)+1 > len(resource) {
				resource = make([]uint16, int(resourceLength)+1)
				grew = true
			}

			if int(deviceLength)+1 > len(device) {
				device = make([]uint16, int(deviceLength)+1)
				grew = true
			}

			if !grew {
				return "", 0, errors.New("quorum resource name buffer did not grow")
			}

			continue
		}

		if err != nil {
			return "", 0, err
		}

		if uint64(resourceLength) >= uint64(len(resource)) {
			return "", 0, fmt.Errorf("quorum resource name length %d exceeds buffer", resourceLength)
		}

		units := resource[:resourceLength]
		for index, unit := range units {
			if unit == 0 {
				units = units[:index]

				break
			}
		}

		return decodeUnits(units), logSize, nil
	}

	return "", 0, errors.New("quorum resource buffers did not stabilize")
}
