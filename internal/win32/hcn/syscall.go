package hcn

import (
	"encoding/json"
	"fmt"
	"unsafe"

	"github.com/prometheus-community/windows_exporter/internal/win32/guid"
	"github.com/prometheus-community/windows_exporter/internal/win32/hcs"
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
func EnumerateEndpoints() ([]guid.GUID, error) {
	var (
		endpointsJSON *uint16
		errorRecord   *uint16
	)

	r1, _, _ := procHcnEnumerateEndpoints.Call(
		0,
		uintptr(unsafe.Pointer(&endpointsJSON)),
		uintptr(unsafe.Pointer(&errorRecord)),
	)

	windows.CoTaskMemFree(unsafe.Pointer(errorRecord))
	result := windows.UTF16PtrToString(endpointsJSON)

	if r1 != 0 {
		return nil, fmt.Errorf("HcnEnumerateEndpoints failed: HRESULT 0x%X: %w", r1, hcs.Win32FromHResult(r1))
	}

	var endpoints []guid.GUID

	if err := json.Unmarshal([]byte(result), &endpoints); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	return endpoints, nil
}

// OpenEndpoint opens an endpoint.
//
// https://learn.microsoft.com/en-us/virtualization/api/hcn/reference/hcnopenendpoint
func OpenEndpoint(id guid.GUID) (Endpoint, error) {
	var (
		endpoint    Endpoint
		errorRecord *uint16
	)

	r1, _, _ := procHcnOpenEndpoint.Call(
		uintptr(unsafe.Pointer(&id)),
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

	windows.CoTaskMemFree(unsafe.Pointer(errorRecord))

	result := windows.UTF16PtrToString(resultDocument)
	windows.CoTaskMemFree(unsafe.Pointer(resultDocument))

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
