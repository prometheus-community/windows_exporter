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

	"github.com/go-ole/go-ole"
	"github.com/prometheus-community/windows_exporter/internal/headers/hcs"
	"golang.org/x/sys/windows"
)

//nolint:gochecknoglobals
var (
	modComputeNetwork = windows.NewLazySystemDLL("computenetwork.dll")

	procHcnEnumerateEndpoints      = modComputeNetwork.NewProc("HcnEnumerateEndpoints")
	procHcnOpenEndpoint            = modComputeNetwork.NewProc("HcnOpenEndpoint")
	procHcnQueryEndpointProperties = modComputeNetwork.NewProc("HcnQueryEndpointProperties")
	procHcnCloseEndpoint           = modComputeNetwork.NewProc("HcnCloseEndpoint")
)

// EnumerateEndpoints enumerates the endpoints.
//
// https://learn.microsoft.com/en-us/virtualization/api/hcn/reference/hcnenumerateendpoints
func EnumerateEndpoints() ([]*ole.GUID, error) {
	var (
		endpointsJSON *uint16
		errorRecord   *uint16
	)

	r1, _, _ := procHcnEnumerateEndpoints.Call(
		0,
		uintptr(unsafe.Pointer(&endpointsJSON)),
		uintptr(unsafe.Pointer(&errorRecord)),
	)

	result := windows.UTF16PtrToString(endpointsJSON)
	windows.CoTaskMemFree(unsafe.Pointer(endpointsJSON))
	windows.CoTaskMemFree(unsafe.Pointer(errorRecord))

	if r1 != 0 {
		return nil, fmt.Errorf("HcnEnumerateEndpoints failed: HRESULT 0x%X: %w", r1, hcs.Win32FromHResult(r1))
	}

	var endpointStringIDs []string

	if err := json.Unmarshal([]byte(result), &endpointStringIDs); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON %s: %w", result, err)
	}

	var endpoints []*ole.GUID

	for _, id := range endpointStringIDs {
		guid := ole.NewGUID(id)
		endpoints = append(endpoints, guid)
	}

	return endpoints, nil
}

// OpenEndpoint opens an endpoint.
//
// https://learn.microsoft.com/en-us/virtualization/api/hcn/reference/hcnopenendpoint
func OpenEndpoint(endpointID *ole.GUID) (Endpoint, error) {
	var (
		endpoint    Endpoint
		errorRecord *uint16
	)

	r1, _, _ := procHcnOpenEndpoint.Call(
		uintptr(unsafe.Pointer(endpointID)),
		uintptr(unsafe.Pointer(&endpoint)),
		uintptr(unsafe.Pointer(&errorRecord)),
	)

	windows.CoTaskMemFree(unsafe.Pointer(errorRecord))

	if r1 != 0 {
		return 0, fmt.Errorf("HcnOpenEndpoint failed: HRESULT 0x%X: %w", r1, hcs.Win32FromHResult(r1))
	}

	return endpoint, nil
}

// QueryEndpointProperties queries the properties of an endpoint.
//
// https://learn.microsoft.com/en-us/virtualization/api/hcn/reference/hcnqueryendpointproperties
func QueryEndpointProperties(endpoint Endpoint, propertyQuery *uint16) (EndpointProperties, error) {
	var (
		resultDocument *uint16
		errorRecord    *uint16
	)

	r1, _, _ := procHcnQueryEndpointProperties.Call(
		uintptr(endpoint),
		uintptr(unsafe.Pointer(&propertyQuery)),
		uintptr(unsafe.Pointer(&resultDocument)),
		uintptr(unsafe.Pointer(&errorRecord)),
	)

	result := windows.UTF16PtrToString(resultDocument)
	windows.CoTaskMemFree(unsafe.Pointer(resultDocument))
	windows.CoTaskMemFree(unsafe.Pointer(errorRecord))

	if r1 != 0 {
		return EndpointProperties{}, fmt.Errorf("HcsGetComputeSystemProperties failed: HRESULT 0x%X: %w", r1, hcs.Win32FromHResult(r1))
	}

	var properties EndpointProperties

	if err := json.Unmarshal([]byte(result), &properties); err != nil {
		return EndpointProperties{}, fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	return properties, nil
}

// CloseEndpoint close a handle to an Endpoint.
//
// https://learn.microsoft.com/en-us/virtualization/api/hcn/reference/hcncloseendpoint
func CloseEndpoint(endpoint Endpoint) {
	_, _, _ = procHcnCloseEndpoint.Call(uintptr(endpoint))
}
