// SPDX-License-Identifier: MIT
//
// Copyright The Prometheus Authors

//go:build windows

package hcn

import (
	"golang.org/x/sys/windows"
)

type Endpoint = windows.Handle

// EndpointProperties contains the properties of an HCN endpoint.
//
// https://learn.microsoft.com/en-us/virtualization/api/hcn/hns_schema#HostComputeEndpoint
type EndpointProperties struct {
	ID               string   `json:"ID"`
	State            int      `json:"State"`
	SharedContainers []string `json:"SharedContainers"`
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
