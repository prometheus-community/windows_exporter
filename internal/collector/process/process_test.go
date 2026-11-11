// SPDX-License-Identifier: MIT
//
// Copyright The Prometheus Authors

//go:build windows

package process_test

import (
	"os"
	"regexp"
	"strconv"
	"testing"

	"github.com/prometheus-community/windows_exporter/internal/collector/process"
	"github.com/prometheus-community/windows_exporter/internal/utils/testutils"
	"github.com/stretchr/testify/require"
)

func BenchmarkProcessCollector(b *testing.B) {
	testutils.FuncBenchmarkCollector(b, process.Name, process.NewWithFlags)
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

	cpu := families["windows_process_cpu_time_seconds_total"].GetMetric()
	require.Len(t, cpu, 2)
	require.NotNil(t, cpu[0].GetCounter())
}
