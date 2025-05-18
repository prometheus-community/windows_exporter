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

package hcs

import (
	"encoding/json"
	"fmt"

	"github.com/prometheus-community/windows_exporter/internal/utils"
	"golang.org/x/sys/windows"
)

//nolint:gochecknoglobals
var (
	ContainerQuery  = utils.Must(windows.UTF16PtrFromString(`{"Types":["Container"]}`))
	StatisticsQuery = utils.Must(windows.UTF16PtrFromString(`{"PropertyTypes":["Statistics"]}`))
)

func GetContainers() ([]Properties, error) {
	operation, err := CreateOperation()
	if err != nil {
		return nil, fmt.Errorf("failed to create operation: %w", err)
	}

	defer CloseOperation(operation)

	if err := EnumerateComputeSystems(ContainerQuery, operation); err != nil {
		return nil, fmt.Errorf("failed to enumerate compute systems: %w", err)
	}

	resultDocument, err := WaitForOperationResult(operation, 1000)
	if err != nil {
		return nil, fmt.Errorf("failed to wait and get for operation result: %w - %s", err, resultDocument)
	} else if resultDocument == "" {
		return nil, ErrEmptyResultDocument
	}

	var computeSystems []Properties
	if err := json.Unmarshal([]byte(resultDocument), &computeSystems); err != nil {
		return nil, fmt.Errorf("failed to unmarshal compute systems: %w", err)
	}

	return computeSystems, nil
}

func GetContainerStatistics(containerID string) (Statistics, error) {
	computeSystem, err := OpenComputeSystem(containerID)
	if err != nil {
		return Statistics{}, fmt.Errorf("failed to open compute system: %w", err)
	}

	defer CloseComputeSystem(computeSystem)

	operation, err := CreateOperation()
	if err != nil {
		return Statistics{}, fmt.Errorf("failed to create operation: %w", err)
	}

	defer CloseOperation(operation)

	if err := GetComputeSystemProperties(computeSystem, operation, StatisticsQuery); err != nil {
		return Statistics{}, fmt.Errorf("failed to enumerate compute systems: %w", err)
	}

	resultDocument, err := WaitForOperationResult(operation, 1000)
	if err != nil {
		return Statistics{}, fmt.Errorf("failed to get compute system properties: %w", err)
	} else if resultDocument == "" {
		return Statistics{}, ErrEmptyResultDocument
	}

	var properties Properties
	if err := json.Unmarshal([]byte(resultDocument), &properties); err != nil {
		return Statistics{}, fmt.Errorf("failed to unmarshal system properties: %w", err)
	}

	if properties.Statistics == nil {
		return Statistics{}, fmt.Errorf("no statistics found for container %s", containerID)
	}

	return *properties.Statistics, nil
}
