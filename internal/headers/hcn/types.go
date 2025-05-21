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
	"github.com/prometheus-community/windows_exporter/internal/headers/guid"
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
