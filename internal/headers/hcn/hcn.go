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
	"fmt"

	"github.com/prometheus-community/windows_exporter/internal/headers/guid"
	"github.com/prometheus-community/windows_exporter/internal/utils"
	"golang.org/x/sys/windows"
)

//nolint:gochecknoglobals
var (
	defaultQuery = utils.Must(windows.UTF16PtrFromString(`{"SchemaVersion":{"Major": 2,"Minor": 0},"Flags":"None"}`))
)

func GetEndpointProperties(endpointID guid.GUID) (EndpointProperties, error) {
	endpoint, err := OpenEndpoint(endpointID)
	if err != nil {
		return EndpointProperties{}, fmt.Errorf("failed to open endpoint: %w", err)
	}

	defer CloseEndpoint(endpoint)

	result, err := QueryEndpointProperties(endpoint, defaultQuery)
	if err != nil {
		return EndpointProperties{}, fmt.Errorf("failed to query endpoint properties: %w", err)
	}

	return result, nil
}
