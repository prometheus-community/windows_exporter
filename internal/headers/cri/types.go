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

package cri

import (
	"time"

	"google.golang.org/protobuf/encoding/protowire"
)

// ContainerState is the runtime.v1.ContainerState enum.
type ContainerState int32

const (
	ContainerCreated ContainerState = 0
	ContainerRunning ContainerState = 1
	ContainerExited  ContainerState = 2
	ContainerUnknown ContainerState = 3
)

// PodSandboxState is the runtime.v1.PodSandboxState enum.
type PodSandboxState int32

const (
	SandboxReady    PodSandboxState = 0
	SandboxNotReady PodSandboxState = 1
)

// Version is the subset of runtime.v1.VersionResponse used by the exporter.
type Version struct {
	// RuntimeName is the runtime name, e.g. containerd. Kubernetes uses it as container ID prefix.
	RuntimeName       string
	RuntimeVersion    string
	RuntimeAPIVersion string
}

// Container is the subset of runtime.v1.Container used by the exporter.
type Container struct {
	ID           string
	PodSandboxID string
	// Name is the container name from the Kubernetes pod spec.
	Name  string
	State ContainerState
	// CreatedAt is the creation time of the container. It is zero if the runtime did not set it.
	CreatedAt time.Time
}

// PodSandbox is the subset of runtime.v1.PodSandbox used by the exporter.
type PodSandbox struct {
	ID        string
	Name      string
	UID       string
	Namespace string
	State     PodSandboxState
}

// ContainerStats is the subset of runtime.v1.ContainerStats used by the exporter.
// Nil fields were not reported by the runtime.
//
// On Windows, containerd fills them from the HCS statistics of the container,
// which the shim also reads for Hyper-V isolated containers.
// https://github.com/containerd/containerd/blob/v2.4.1/internal/cri/server/container_stats_list.go
type ContainerStats struct {
	ID            string
	CPU           *CPUUsage
	Memory        *MemoryUsage
	WritableLayer *FilesystemUsage
}

// CPUUsage is the subset of runtime.v1.CpuUsage used by the exporter.
type CPUUsage struct {
	// Timestamp is the time of the sample in nanoseconds since the Unix epoch.
	Timestamp int64
	// UsageCoreNanoSeconds is the cumulative CPU time of the container.
	UsageCoreNanoSeconds *uint64
}

// MemoryUsage is the subset of runtime.v1.MemoryUsage used by the exporter.
type MemoryUsage struct {
	// Timestamp is the time of the sample in nanoseconds since the Unix epoch.
	Timestamp int64
	// WorkingSetBytes is the private working set on Windows.
	WorkingSetBytes *uint64
	// UsageBytes is the commit size on Windows.
	UsageBytes *uint64
}

// FilesystemUsage is the subset of runtime.v1.FilesystemUsage used by the exporter.
type FilesystemUsage struct {
	// Timestamp is the time of the sample in nanoseconds since the Unix epoch.
	// containerd sends 0 if it has not measured the usage yet.
	Timestamp int64
	UsedBytes *uint64
}

// field is a decoded protobuf field. value holds varints, bytes length-delimited values.
type field struct {
	num   protowire.Number
	typ   protowire.Type
	value uint64
	bytes []byte
}

func (f field) is(num protowire.Number, typ protowire.Type) bool {
	return f.num == num && f.typ == typ
}

// eachField calls fn for each field of the protobuf message b.
func eachField(b []byte, fn func(field) error) error {
	for len(b) > 0 {
		num, typ, n := protowire.ConsumeTag(b)
		if n < 0 {
			return protowire.ParseError(n)
		}

		b = b[n:]
		f := field{num: num, typ: typ}

		switch typ {
		case protowire.VarintType:
			f.value, n = protowire.ConsumeVarint(b)
		case protowire.BytesType:
			f.bytes, n = protowire.ConsumeBytes(b)
		default:
			n = protowire.ConsumeFieldValue(num, typ, b)
		}

		if n < 0 {
			return protowire.ParseError(n)
		}

		b = b[n:]

		if err := fn(f); err != nil {
			return err
		}
	}

	return nil
}

func decodeVersion(b []byte) (Version, error) {
	var version Version

	err := eachField(b, func(f field) error {
		switch {
		case f.is(2, protowire.BytesType):
			version.RuntimeName = string(f.bytes)
		case f.is(3, protowire.BytesType):
			version.RuntimeVersion = string(f.bytes)
		case f.is(4, protowire.BytesType):
			version.RuntimeAPIVersion = string(f.bytes)
		}

		return nil
	})

	return version, err
}

func decodeListContainersResponse(b []byte) ([]Container, error) {
	var containers []Container

	err := eachField(b, func(f field) error {
		if !f.is(1, protowire.BytesType) {
			return nil
		}

		container, err := decodeContainer(f.bytes)
		if err != nil {
			return err
		}

		containers = append(containers, container)

		return nil
	})

	return containers, err
}

func decodeContainer(b []byte) (Container, error) {
	var container Container

	err := eachField(b, func(f field) error {
		switch {
		case f.is(1, protowire.BytesType):
			container.ID = string(f.bytes)
		case f.is(2, protowire.BytesType):
			container.PodSandboxID = string(f.bytes)
		case f.is(3, protowire.BytesType): // ContainerMetadata
			return eachField(f.bytes, func(f field) error {
				if f.is(1, protowire.BytesType) {
					container.Name = string(f.bytes)
				}

				return nil
			})
		case f.is(6, protowire.VarintType):
			container.State = ContainerState(f.value)
		case f.is(7, protowire.VarintType): // created_at, int64 in nanoseconds since the Unix epoch
			if f.value != 0 {
				container.CreatedAt = time.Unix(0, int64(f.value))
			}
		}

		return nil
	})

	return container, err
}

func decodeListPodSandboxResponse(b []byte) ([]PodSandbox, error) {
	var sandboxes []PodSandbox

	err := eachField(b, func(f field) error {
		if !f.is(1, protowire.BytesType) {
			return nil
		}

		sandbox, err := decodePodSandbox(f.bytes)
		if err != nil {
			return err
		}

		sandboxes = append(sandboxes, sandbox)

		return nil
	})

	return sandboxes, err
}

func decodePodSandbox(b []byte) (PodSandbox, error) {
	var sandbox PodSandbox

	err := eachField(b, func(f field) error {
		switch {
		case f.is(1, protowire.BytesType):
			sandbox.ID = string(f.bytes)
		case f.is(2, protowire.BytesType): // PodSandboxMetadata
			return eachField(f.bytes, func(f field) error {
				switch {
				case f.is(1, protowire.BytesType):
					sandbox.Name = string(f.bytes)
				case f.is(2, protowire.BytesType):
					sandbox.UID = string(f.bytes)
				case f.is(3, protowire.BytesType):
					sandbox.Namespace = string(f.bytes)
				}

				return nil
			})
		case f.is(3, protowire.VarintType):
			sandbox.State = PodSandboxState(f.value)
		}

		return nil
	})

	return sandbox, err
}

func decodeListContainerStatsResponse(b []byte) ([]ContainerStats, error) {
	var stats []ContainerStats

	err := eachField(b, func(f field) error {
		if !f.is(1, protowire.BytesType) {
			return nil
		}

		s, err := decodeContainerStats(f.bytes)
		if err != nil {
			return err
		}

		stats = append(stats, s)

		return nil
	})

	return stats, err
}

func decodeContainerStats(b []byte) (ContainerStats, error) {
	var stats ContainerStats

	err := eachField(b, func(f field) error {
		var err error

		switch {
		case f.is(1, protowire.BytesType): // ContainerAttributes
			err = eachField(f.bytes, func(f field) error {
				if f.is(1, protowire.BytesType) {
					stats.ID = string(f.bytes)
				}

				return nil
			})
		case f.is(2, protowire.BytesType):
			stats.CPU = &CPUUsage{}
			err = eachField(f.bytes, func(f field) error {
				var err error

				switch {
				case f.is(1, protowire.VarintType):
					stats.CPU.Timestamp = int64(f.value)
				case f.is(2, protowire.BytesType):
					stats.CPU.UsageCoreNanoSeconds, err = decodeUInt64Value(f.bytes)
				}

				return err
			})
		case f.is(3, protowire.BytesType):
			stats.Memory = &MemoryUsage{}
			err = eachField(f.bytes, func(f field) error {
				var err error

				switch {
				case f.is(1, protowire.VarintType):
					stats.Memory.Timestamp = int64(f.value)
				case f.is(2, protowire.BytesType):
					stats.Memory.WorkingSetBytes, err = decodeUInt64Value(f.bytes)
				case f.is(4, protowire.BytesType):
					stats.Memory.UsageBytes, err = decodeUInt64Value(f.bytes)
				}

				return err
			})
		case f.is(4, protowire.BytesType):
			stats.WritableLayer = &FilesystemUsage{}
			err = eachField(f.bytes, func(f field) error {
				var err error

				switch {
				case f.is(1, protowire.VarintType):
					stats.WritableLayer.Timestamp = int64(f.value)
				case f.is(3, protowire.BytesType):
					stats.WritableLayer.UsedBytes, err = decodeUInt64Value(f.bytes)
				}

				return err
			})
		}

		return err
	})

	return stats, err
}

// decodeUInt64Value decodes a runtime.v1.UInt64Value. An empty message is the value 0.
func decodeUInt64Value(b []byte) (*uint64, error) {
	var value uint64

	err := eachField(b, func(f field) error {
		if f.is(1, protowire.VarintType) {
			value = f.value
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return &value, nil
}

// encodeListRunningContainersRequest encodes a ListContainersRequest
// with filter.state.state set to CONTAINER_RUNNING.
func encodeListRunningContainersRequest() []byte {
	var state []byte // ContainerStateValue

	state = protowire.AppendTag(state, 1, protowire.VarintType)
	state = protowire.AppendVarint(state, uint64(ContainerRunning))

	var filter []byte // ContainerFilter

	filter = protowire.AppendTag(filter, 2, protowire.BytesType)
	filter = protowire.AppendBytes(filter, state)

	var req []byte // ListContainersRequest

	req = protowire.AppendTag(req, 1, protowire.BytesType)
	req = protowire.AppendBytes(req, filter)

	return req
}
