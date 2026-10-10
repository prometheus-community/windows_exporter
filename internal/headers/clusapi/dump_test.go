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
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"slices"
	"strings"
	"testing"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"
)

const (
	resourceGetROPrivateProperties    = 0x0100007d
	resourceGetPrivateProperties      = 0x01000081
	resourceTypeGetROCommonProperties = 0x02000055
	resourceTypeGetCommonProperties   = 0x02000059
	resourceTypeGetPrivateProperties  = 0x02000081
)

//nolint:gochecknoglobals
var resourceTypeControl = dll.NewProc("ClusterResourceTypeControl")

// TestDumpPropertyLists logs raw property lists from a live cluster as hex, so
// parser fixtures can be built from real cluster service output. Set
// WINDOWS_EXPORTER_TEST_CLUSAPI_DUMP=1 on a cluster node to enable it.
func TestDumpPropertyLists(t *testing.T) {
	if os.Getenv("WINDOWS_EXPORTER_TEST_CLUSAPI_DUMP") == "" {
		t.Skip("WINDOWS_EXPORTER_TEST_CLUSAPI_DUMP is not set")
	}

	cluster, _, err := openCluster.Call(0, windows.GENERIC_READ, 0)
	if cluster == 0 {
		t.Fatalf("OpenClusterEx: %v", err)
	}

	defer closeCluster.Call(cluster) //nolint:errcheck

	deadline := time.Now().Add(time.Minute)

	for _, code := range []uint32{objectCluster<<24 | ctlGetROCommonProperties, objectCluster<<24 | ctlGetCommonProperties} {
		data, err := controlBuffer(deadline, 0, func(buffer []byte) (uint32, error) {
			var (
				size    uint32
				pointer *byte
			)
			if len(buffer) != 0 {
				pointer = &buffer[0]
			}

			status, _, _ := clusterControl.Call(cluster, 0, uintptr(code), 0, 0, uintptr(unsafe.Pointer(pointer)), uintptr(len(buffer)), uintptr(unsafe.Pointer(&size)))
			if status != 0 {
				return size, windows.Errno(status)
			}

			return size, nil
		})
		logPropertyList(t, fmt.Sprintf("cluster %#x", code), data, err)
	}

	// Resource type lists carry EXPAND_SZ values such as DllName, which the
	// cluster service returns as an EXPAND_SZ and EXPANDED_SZ value pair.
	for _, typeName := range []string{"Generic Service", "IP Address"} {
		typePtr, err := windows.UTF16PtrFromString(typeName)
		if err != nil {
			t.Fatal(err)
		}

		for _, code := range []uint32{resourceTypeGetROCommonProperties, resourceTypeGetCommonProperties, resourceTypeGetPrivateProperties} {
			data, err := controlBuffer(deadline, 0, func(buffer []byte) (uint32, error) {
				var (
					size    uint32
					pointer *byte
				)
				if len(buffer) != 0 {
					pointer = &buffer[0]
				}

				status, _, _ := resourceTypeControl.Call(cluster, uintptr(unsafe.Pointer(typePtr)), 0, uintptr(code), 0, 0, uintptr(unsafe.Pointer(pointer)), uintptr(len(buffer)), uintptr(unsafe.Pointer(&size)))
				if status != 0 {
					return size, windows.Errno(status)
				}

				return size, nil
			})
			logPropertyList(t, fmt.Sprintf("resource type %q %#x", typeName, code), data, err)
		}
	}

	enum, _, err := openEnum.Call(cluster, 4)
	if enum == 0 {
		t.Fatalf("ClusterOpenEnum: %v", err)
	}

	defer closeEnum.Call(enum) //nolint:errcheck

	for index := uint32(0); ; index++ {
		name, err := enumName(enum, index, enumResource, deadline)
		if errors.Is(err, windows.ERROR_NO_MORE_ITEMS) {
			return
		}

		if err != nil {
			t.Fatalf("ClusterEnum: %v", err)
		}

		handle, _, err := openResource.Call(cluster, uintptr(unsafe.Pointer(&name.units[0])), windows.GENERIC_READ, 0)
		if handle == 0 {
			t.Errorf("OpenClusterResourceEx %q: %v", name.name, err)

			continue
		}

		for _, code := range []uint32{resourceGetROCommonProperties, resourceGetCommonProperties, resourceGetROPrivateProperties, resourceGetPrivateProperties} {
			data, err := resourceBuffer(handle, code, 0, deadline)
			logPropertyList(t, fmt.Sprintf("resource %q %#x", name.name, code), data, err)
		}

		data, err := resourceBuffer(handle, resourceGetType, 0, deadline)
		t.Logf("resource %q GET_RESOURCE_TYPE: %s err=%v", name.name, hex.EncodeToString(data), err)

		closeResource.Call(handle) //nolint:errcheck
	}
}

func logPropertyList(t *testing.T, label string, data []byte, err error) {
	t.Helper()

	if err != nil {
		t.Logf("%s: %v", label, err)

		return
	}

	properties, parseErr := ParseProperties(data)

	formats := make([]string, 0, len(properties))
	for name, values := range properties {
		valueFormats := make([]string, 0, len(values))
		for _, value := range values {
			valueFormats = append(valueFormats, fmt.Sprintf("%d/%d", value.Format, len(value.Data)))
		}

		formats = append(formats, name+"="+strings.Join(valueFormats, "+"))
	}

	slices.Sort(formats)

	t.Logf("%s: parse error=%v formats=[%s]\n%s", label, parseErr, strings.Join(formats, " "), hex.EncodeToString(data))
}
