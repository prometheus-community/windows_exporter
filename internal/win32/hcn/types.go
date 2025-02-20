package hcn

import (
	"github.com/prometheus-community/windows_exporter/internal/win32/guid"
	"golang.org/x/sys/windows"
)

type Endpoint = windows.Handle

// EndpointProperties contains the properties of an HCN endpoint.
//
// https://learn.microsoft.com/en-us/virtualization/api/hcn/hns_schema#HostComputeEndpoint
type EndpointProperties struct {
	ID               string                      `json:"ID"`
	State            int                         `json:"State"`
	SharedContainers []string                    `json:"SharedContainers"`
	Resources        EndpointPropertiesResources `json:"Resources"`
}

type EndpointPropertiesResources struct {
	Allocators []EndpointPropertiesAllocators `json:"Allocators"`
}
type EndpointPropertiesAllocators struct {
	AdapterNetCfgInstanceId *guid.GUID `json:"AdapterNetCfgInstanceId"`
}

type EndpointStats struct {
	BytesReceived          uint64 `json:"BytesReceived"`
	BytesSent              uint64 `json:"BytesSent"`
	DroppedPacketsIncoming uint64 `json:"DroppedPacketsIncoming"`
	DroppedPacketsOutgoing uint64 `json:"DroppedPacketsOutgoing"`
	EndpointID             string `json:"EndpointId"`
	InstanceID             string `json:"InstanceId"`
	PacketsReceived        uint64 `json:"PacketsReceived"`
	PacketsSent            uint64 `json:"PacketsSent"`
}
