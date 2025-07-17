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

	hcnBodyEmpty         = utils.Must(windows.UTF16PtrFromString(""))
	hcnMethodGet         = utils.Must(windows.UTF16PtrFromString("GET"))
	hcnPathEndpoints     = utils.Must(windows.UTF16PtrFromString("/endpoints/"))
	hcnPathEndpointStats = utils.Must(windows.UTF16FromString("/endpointstats/"))
)

func ListEndpoints() ([]EndpointProperties, error) {
	result, err := hnsCall(hcnMethodGet, hcnPathEndpoints, hcnBodyEmpty)
	if err != nil {
		return nil, err
	}

	var endpoints struct {
		Success bool                 `json:"success"`
		Error   string               `json:"error"`
		Output  []EndpointProperties `json:"output"`
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
	endpointIDUTF16, err := windows.UTF16FromString(endpointID)
	if err != nil {
		return EndpointStats{}, fmt.Errorf("failed to convert endpoint ID to UTF16: %w", err)
	}

	path := hcnPathEndpointStats[:len(hcnPathEndpointStats)-1]
	path = append(path, endpointIDUTF16...)

	result, err := hnsCall(hcnMethodGet, &path[0], hcnBodyEmpty)
	if err != nil {
		return EndpointStats{}, err
	}

	var stats EndpointStats

	if err := json.Unmarshal([]byte(result), &stats); err != nil {
		return EndpointStats{}, fmt.Errorf("failed to unmarshal JSON %s: %w", result, err)
	}

	return stats, nil
}

func hnsCall(method, path, body *uint16) (string, error) {
	var responseJSON *uint16

	r1, _, _ := procHNSCall.Call(
		uintptr(unsafe.Pointer(method)),
		uintptr(unsafe.Pointer(path)),
		uintptr(unsafe.Pointer(body)),
		uintptr(unsafe.Pointer(&responseJSON)),
	)

	response := windows.UTF16PtrToString(responseJSON)
	windows.CoTaskMemFree(unsafe.Pointer(responseJSON))

	if r1 != 0 {
		return "", fmt.Errorf("HNSCall failed: HRESULT 0x%X: %w", r1, hcs.Win32FromHResult(r1))
	}

	return response, nil
}
