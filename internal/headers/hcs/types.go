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
	"errors"
	"time"

	"golang.org/x/sys/windows"
)

var (
	ErrEmptyResultDocument = errors.New("empty result document")
	ErrIDNotFound          = windows.Errno(2151088398)
)

type (
	Operation     = windows.Handle
	ComputeSystem = windows.Handle
)

type Properties struct {
	ID          string           `json:"Id,omitempty"`
	SystemType  string           `json:"SystemType,omitempty"`
	Owner       string           `json:"Owner,omitempty"`
	State       string           `json:"State,omitempty"`
	Statistics  *Statistics      `json:"Statistics,omitempty"`
	ProcessList []ProcessDetails `json:"ProcessList,omitempty"`
}

type ProcessDetails struct {
	ProcessId                    int32     `json:"ProcessId,omitempty"`
	ImageName                    string    `json:"ImageName,omitempty"`
	CreateTimestamp              time.Time `json:"CreateTimestamp"`
	UserTime100ns                int32     `json:"UserTime100ns,omitempty"`
	KernelTime100ns              int32     `json:"KernelTime100ns,omitempty"`
	MemoryCommitBytes            int32     `json:"MemoryCommitBytes,omitempty"`
	MemoryWorkingSetPrivateBytes int32     `json:"MemoryWorkingSetPrivateBytes,omitempty"`
	MemoryWorkingSetSharedBytes  int32     `json:"MemoryWorkingSetSharedBytes,omitempty"`
}

type Statistics struct {
	Timestamp          time.Time       `json:"Timestamp"`
	ContainerStartTime time.Time       `json:"ContainerStartTime"`
	Uptime100ns        uint64          `json:"Uptime100ns,omitempty"`
	Processor          *ProcessorStats `json:"Processor,omitempty"`
	Memory             *MemoryStats    `json:"Memory,omitempty"`
	Storage            *StorageStats   `json:"Storage,omitempty"`
}

type ProcessorStats struct {
	TotalRuntime100ns  uint64 `json:"TotalRuntime100ns,omitempty"`
	RuntimeUser100ns   uint64 `json:"RuntimeUser100ns,omitempty"`
	RuntimeKernel100ns uint64 `json:"RuntimeKernel100ns,omitempty"`
}

type MemoryStats struct {
	MemoryUsageCommitBytes            uint64 `json:"MemoryUsageCommitBytes,omitempty"`
	MemoryUsageCommitPeakBytes        uint64 `json:"MemoryUsageCommitPeakBytes,omitempty"`
	MemoryUsagePrivateWorkingSetBytes uint64 `json:"MemoryUsagePrivateWorkingSetBytes,omitempty"`
}

type StorageStats struct {
	ReadCountNormalized  uint64 `json:"ReadCountNormalized,omitempty"`
	ReadSizeBytes        uint64 `json:"ReadSizeBytes,omitempty"`
	WriteCountNormalized uint64 `json:"WriteCountNormalized,omitempty"`
	WriteSizeBytes       uint64 `json:"WriteSizeBytes,omitempty"`
}
