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

package setupapi_test

import (
	"strings"
	"testing"
	"unsafe"

	"github.com/prometheus-community/windows_exporter/internal/headers/setupapi"
	"github.com/stretchr/testify/require"
	"golang.org/x/sys/windows"
)

func TestStructSizes(t *testing.T) {
	t.Parallel()

	if unsafe.Sizeof(uintptr(0)) == 8 {
		require.Equal(t, uintptr(32), unsafe.Sizeof(setupapi.SP_DEVICE_INTERFACE_DATA{}))
		require.Equal(t, uintptr(32), unsafe.Sizeof(setupapi.SP_DEVINFO_DATA{}))
	} else {
		require.Equal(t, uintptr(28), unsafe.Sizeof(setupapi.SP_DEVICE_INTERFACE_DATA{}))
		require.Equal(t, uintptr(28), unsafe.Sizeof(setupapi.SP_DEVINFO_DATA{}))
	}
}

func TestGetDeviceInterfaces(t *testing.T) {
	t.Parallel()

	interfaces, err := setupapi.GetDeviceInterfaces(&setupapi.GUID_DEVINTERFACE_DISK)
	require.NoError(t, err)

	if len(interfaces) == 0 {
		t.Skip("no disk device interfaces present")
	}

	for _, deviceInterface := range interfaces {
		require.True(t, strings.HasPrefix(deviceInterface.Path, `\\?\`), deviceInterface.Path)
		require.True(t, strings.HasSuffix(strings.ToLower(deviceInterface.Path), "{53f56307-b6bf-11d0-94f2-00a0c91efb8b}"), deviceInterface.Path)

		var status, problem uint32

		require.NoError(t, windows.CM_Get_DevNode_Status(&status, &problem, windows.DEVINST(deviceInterface.DevInst), 0))
	}
}
