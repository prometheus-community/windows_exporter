// Copyright 2024 The Prometheus Authors
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

package pdh_test

import (
	"testing"

	v2 "github.com/prometheus-community/windows_exporter/internal/pdh"
	"github.com/stretchr/testify/require"
)

func BenchmarkTestCollector(b *testing.B) {
	counters := []string{
		"% Processor Time",
		"% Privileged Time",
		"% User Time",
		"Creating Process ID",
		"Elapsed Time",
		"Handle Count",
		"ID Process",
		"IO Data Bytes/sec",
		"IO Data Operations/sec",
		"IO Other Bytes/sec",
		"IO Other Operations/sec",
		"IO Read Bytes/sec",
		"IO Read Operations/sec",
		"IO Write Bytes/sec",
		"IO Write Operations/sec",
		"Page Faults/sec",
		"Page File Bytes Peak",
		"Page File Bytes",
		"Pool Nonpaged Bytes",
		"Pool Paged Bytes",
		"Priority Base",
		"Private Bytes",
		"Thread Count",
		"Virtual Bytes Peak",
		"Virtual Bytes",
		"Working Set - Private",
		"Working Set Peak",
		"Working Set",
	}
	performanceData, err := v2.NewCollector("Process", []string{"*"}, counters, false)
	require.NoError(b, err)

	for i := 0; i < b.N; i++ {
		_, _ = performanceData.Collect()
	}

	performanceData.Close()

	b.ReportAllocs()
}
