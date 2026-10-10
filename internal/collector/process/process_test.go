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

package process_test

import (
	"os"
	"regexp"
	"strconv"
	"testing"

	"github.com/alecthomas/kingpin/v2"
	"github.com/prometheus-community/windows_exporter/internal/collector/process"
	"github.com/prometheus-community/windows_exporter/internal/utils/testutils"
	"github.com/stretchr/testify/require"
)

func BenchmarkProcessCollector(b *testing.B) {
	// PrinterInclude is not set in testing context (kingpin flags not parsed), causing the collector to skip all processes.
	localProcessInclude := ".+"
	// No context name required as collector source is WMI
	testutils.FuncBenchmarkCollector(b, process.Name, process.NewWithFlags, func(app *kingpin.Application) {
		app.GetFlag("collector.process.include").StringVar(&localProcessInclude)
	})
}

func TestCollector(t *testing.T) {
	testutils.TestCollector(t, process.New, nil)
}

func TestCollectorCurrentProcess(t *testing.T) {
	config := process.ConfigDefaults
	config.ProcessInclude = regexp.MustCompile(`^process\.test$`)

	families := testutils.TestCollector(t, process.New, &config)
	pid := strconv.Itoa(os.Getpid())

	infos := families["windows_process_info"].GetMetric()
	require.Len(t, infos, 1)

	labels := map[string]string{}
	for _, label := range infos[0].GetLabel() {
		labels[label.GetName()] = label.GetValue()
	}

	require.Equal(t, "process.test", labels["process"])
	require.Equal(t, pid, labels["process_id"])
	require.Equal(t, strconv.Itoa(os.Getppid()), labels["creating_process_id"])
	require.NotEmpty(t, labels["owner"])
	require.Contains(t, labels["cmdline"], "process.test")

	for _, name := range []string{
		"windows_process_start_time_seconds_timestamp",
		"windows_process_handles",
		"windows_process_threads",
		"windows_process_virtual_bytes",
		"windows_process_working_set_bytes",
		"windows_process_private_bytes",
	} {
		require.Positive(t, testutils.MetricValuesByLabel(families, name, "process_id")[pid], name)
	}

	cpu := families["windows_process_cpu_time_total"].GetMetric()
	require.Len(t, cpu, 2)
	require.NotNil(t, cpu[0].GetCounter())
}
