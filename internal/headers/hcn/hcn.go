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

package hcn

import (
	"encoding/json"
	"fmt"
	"unsafe"

	"github.com/prometheus-community/windows_exporter/internal/headers/hcs"
	"github.com/prometheus-community/windows_exporter/internal/utils"
	"golang.org/x/sys/windows"
)

//nolint:gochecknoglobals
var (
	modvmcompute = windows.NewLazySystemDLL("vmcompute.dll")
	procHNSCall  = modvmcompute.NewProc("HNSCall")
)

//nolint:gochecknoglobals
var (
	defaultQuery = utils.Must(windows.UTF16PtrFromString(`{"SchemaVersion":{"Major": 2,"Minor": 0},"Flags":"None"}`))

	hcnBodyEmpty         = utils.Must(windows.UTF16PtrFromString(""))
	hcnMethodGet         = utils.Must(windows.UTF16PtrFromString("GET"))
	hcnPathEndpoints     = utils.Must(windows.UTF16PtrFromString("/endpoints/"))
	hcnPathEndpointStats = utils.Must(windows.UTF16FromString("/endpointstats/"))
)

func ListEndpoints() ([]EndpointProperties, error) {
	var responseJSON *uint16

	r1, _, _ := procHNSCall.Call(
		uintptr(unsafe.Pointer(hcnMethodGet)),
		uintptr(unsafe.Pointer(hcnPathEndpoints)),
		uintptr(unsafe.Pointer(hcnBodyEmpty)),
		uintptr(unsafe.Pointer(&responseJSON)),
	)

	result := windows.UTF16PtrToString(responseJSON)
	windows.CoTaskMemFree(unsafe.Pointer(responseJSON))

	if r1 != 0 {
		return nil, fmt.Errorf("HNSCall failed: HRESULT 0x%X: %w", r1, hcs.Win32FromHResult(r1))
	}

	var endpoints struct {
		Success bool
		Error   string
		Output  []EndpointProperties
	}

	if err := json.Unmarshal([]byte(result), &endpoints); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON %s: %w", result, err)
	}

	if !endpoints.Success {
		return nil, fmt.Errorf("HNSCall failed: %s", endpoints.Error)
	}

	return endpoints.Output, nil
}

func GetHNSEndpointStats(endpointID string) (EndpointStats, error) {
	var responseJSON *uint16

	endpointIDUTF16, err := windows.UTF16FromString(endpointID)
	if err != nil {
		return EndpointStats{}, fmt.Errorf("failed to convert endpoint ID to UTF16: %w", err)
	}

	path := append(hcnPathEndpointStats[:len(hcnPathEndpointStats)-1], endpointIDUTF16...)

	r1, _, _ := procHNSCall.Call(
		uintptr(unsafe.Pointer(hcnMethodGet)),
		uintptr(unsafe.Pointer(&path[0])),
		uintptr(unsafe.Pointer(hcnBodyEmpty)),
		uintptr(unsafe.Pointer(&responseJSON)),
	)

	if r1 != 0 {
		return EndpointStats{}, fmt.Errorf("HNSCall failed: HRESULT 0x%X: %w", r1, hcs.Win32FromHResult(r1))
	}

	result := windows.UTF16PtrToString(responseJSON)
	windows.CoTaskMemFree(unsafe.Pointer(responseJSON))

	var stats EndpointStats

	if err := json.Unmarshal([]byte(result), &stats); err != nil {
		return EndpointStats{}, fmt.Errorf("failed to unmarshal JSON %s: %w", result, err)
	}

	return stats, nil
}
