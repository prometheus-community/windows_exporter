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
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestParseStatistics(t *testing.T) {
	t.Parallel()

	// A result document of the Statistics and ProcessList query, as HCS returns it for a process-isolated container.
	properties, err := parseStatistics("abc", `{
	"Id": "abc",
	"SystemType": "Container",
	"Statistics": {
		"Timestamp": "2026-10-08T10:00:05.1234567Z",
		"ContainerStartTime": "2026-10-08T10:00:00.5Z",
		"Uptime100ns": 50000000,
		"Processor": {"TotalRuntime100ns": 30000000, "RuntimeUser100ns": 20000000, "RuntimeKernel100ns": 10000000},
		"Memory": {"MemoryUsageCommitBytes": 1, "MemoryUsageCommitPeakBytes": 2, "MemoryUsagePrivateWorkingSetBytes": 3},
		"Storage": {"ReadCountNormalized": 4, "ReadSizeBytes": 5, "WriteCountNormalized": 6, "WriteSizeBytes": 7}
	},
	"ProcessList": [
		{"CreateTimestamp": "2026-10-08T10:00:00.6Z", "ImageName": "smss.exe", "ProcessId": 4096},
		{
			"CreateTimestamp": "2026-10-08T10:00:01Z",
			"ImageName": "app.exe",
			"KernelTime100ns": 9000000000,
			"MemoryCommitBytes": 8589934592,
			"MemoryWorkingSetPrivateBytes": 4294967296,
			"MemoryWorkingSetSharedBytes": 4096,
			"ProcessId": 4294967292,
			"UserTime100ns": 9000000000
		}
	]
}`)
	require.NoError(t, err)
	require.NotNil(t, properties.Statistics)
	require.Equal(t, time.Date(2026, 10, 8, 10, 0, 0, 500_000_000, time.UTC), properties.Statistics.ContainerStartTime)
	require.Equal(t, uint64(30000000), properties.Statistics.Processor.TotalRuntime100ns)
	require.Equal(t, uint64(7), properties.Statistics.Storage.WriteSizeBytes)
	require.Len(t, properties.ProcessList, 2)
	require.Equal(t, uint32(4294967292), properties.ProcessList[1].ProcessId)
	require.Equal(t, uint64(8589934592), properties.ProcessList[1].MemoryCommitBytes)
	require.Equal(t, uint64(9000000000), properties.ProcessList[1].UserTime100ns)
}

func TestParseStatisticsMissing(t *testing.T) {
	t.Parallel()

	_, err := parseStatistics("abc", `{"Id": "abc", "ProcessList": []}`)
	require.ErrorContains(t, err, "no statistics found for container abc")

	_, err = parseStatistics("abc", `{"Statistics": `)
	require.Error(t, err)
}
