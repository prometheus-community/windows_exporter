// SPDX-License-Identifier: MIT
//
// Copyright The Prometheus Authors

//go:build windows

package pdh_test

import (
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/prometheus-community/windows_exporter/internal/pdh"
	"github.com/stretchr/testify/require"
)

func TestIsTotalInstance(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		object string
		name   string
		want   bool
	}{
		{"Process", "_Total", true},
		{"Process", "worker_Total", false},
		{"SQLServer:Databases", "Orders_Total", false},
		{"SQLServer:Databases", "0,_Total", false},
		{"Processor Information", "_Total", true},
		{"Processor Information", "0,_Total", true},
		{"Processor Information", "0,1", false},
		{"Processor Information", "worker_Total", false},
	} {
		require.Equal(t, tc.want, pdh.IsTotalInstance(tc.object, tc.name), "%s/%s", tc.object, tc.name)
	}
}

func TestTotalSuffixProcess(t *testing.T) {
	if os.Getenv("WINDOWS_EXPORTER_TEST_TOTAL_CHILD") == "1" {
		time.Sleep(time.Minute)

		return
	}

	t.Parallel()

	executable, err := os.Executable()
	require.NoError(t, err)
	contents, err := os.ReadFile(executable)
	require.NoError(t, err)
	childPath := filepath.Join(t.TempDir(), "worker_Total.exe")
	//nolint:gosec // The destination is a fixed filename in a test-owned temporary directory.
	require.NoError(t, os.WriteFile(childPath, contents, 0o600))
	child := exec.CommandContext(t.Context(), childPath, "-test.run=^TestTotalSuffixProcess$")

	child.Env = append(os.Environ(), "WINDOWS_EXPORTER_TEST_TOTAL_CHILD=1")
	require.NoError(t, child.Start())
	t.Cleanup(func() { _ = child.Process.Kill(); _ = child.Wait() })

	type processValues struct {
		Name string
		ID   float64 `perfdata:"ID Process"`
	}

	collector, err := pdh.NewCollector[processValues](slog.New(slog.DiscardHandler), pdh.CounterTypeRaw, "Process", []string{"worker_Total"})
	require.NoError(t, err)
	t.Cleanup(collector.Close)

	require.Eventually(t, func() bool {
		var rows []processValues
		if collector.Collect(&rows) != nil {
			return false
		}

		for _, row := range rows {
			if row.ID == float64(child.Process.Pid) && row.Name == "worker_Total" {
				return true
			}
		}

		return false
	}, time.Second*5, time.Millisecond*100)
}
