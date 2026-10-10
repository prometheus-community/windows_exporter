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
	"errors"
	"fmt"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"
)

// CLUS_OBJECT_CLUSTER, used for CLUSCTL_CLUSTER_* control codes.
const objectCluster = 7

// GetClusterInformation, ClusterControl and GetClusterQuorumResource are
// available since Windows Server 2008.
//
//nolint:gochecknoglobals
var (
	clusterInformation = dll.NewProc("GetClusterInformation")
	clusterControl     = dll.NewProc("ClusterControl")
	quorumResource     = dll.NewProc("GetClusterQuorumResource")
)

// Properties reads the cluster name and the 32-bit cluster common properties.
// QuorumLogFileSize falls back to GetClusterQuorumResource when it is not a
// common property. A failed name read returns no object.
func (c *Cluster) Properties(deadline time.Time) (Object, error) {
	for _, proc := range []*windows.LazyProc{openCluster, clusterInformation, clusterControl, quorumResource} {
		if err := proc.Find(); err != nil {
			return Object{}, fmt.Errorf("load ClusAPI: %w", err)
		}
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if err := c.openLocked(deadline); err != nil {
		return Object{}, err
	}

	name, err := readClusterName(c.handle, deadline)
	if err != nil {
		return Object{}, fmt.Errorf("GetClusterInformation: %w", err)
	}

	object := Object{Name: name, Values: make(map[string]uint32)}

	var resultErr error

	if err := readCommonProperties(clusterControl, c.handle, objectCluster, object.Values, deadline); err != nil {
		resultErr = errors.Join(resultErr, err)

		if errors.Is(err, context.DeadlineExceeded) {
			return object, resultErr
		}
	}

	if _, exists := object.Values["QuorumLogFileSize"]; !exists {
		size, err := readQuorumLogSize(c.handle, deadline)
		if err != nil {
			return object, errors.Join(resultErr, fmt.Errorf("GetClusterQuorumResource: %w", err))
		}

		object.Values["QuorumLogFileSize"] = size
	}

	return object, resultErr
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

// readQuorumLogSize returns the maximum quorum log size reported by
// GetClusterQuorumResource. The resource and device names are not used.
func readQuorumLogSize(handle uintptr, deadline time.Time) (uint32, error) {
	size, _, _, err := stateBuffers(deadline, func(resource, device []uint16) (uint32, uint32, uint32, error) {
		var maxLogSize uint32

		resourceLength, deviceLength := uint32(len(resource)), uint32(len(device))

		status, _, _ := quorumResource.Call(handle, uintptr(unsafe.Pointer(&resource[0])), uintptr(unsafe.Pointer(&resourceLength)), uintptr(unsafe.Pointer(&device[0])), uintptr(unsafe.Pointer(&deviceLength)), uintptr(unsafe.Pointer(&maxLogSize)))
		if status != 0 {
			return 0, resourceLength, deviceLength, windows.Errno(status)
		}

		return maxLogSize, resourceLength, deviceLength, nil
	})

	return size, err
}
