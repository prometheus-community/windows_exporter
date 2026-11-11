// SPDX-License-Identifier: MIT
//
// Copyright The Prometheus Authors

//go:build windows

package textfile_test

import (
	"fmt"
	"log/slog"
	"strings"
	"testing"

	"github.com/prometheus-community/windows_exporter/internal/collector/textfile"
	"github.com/prometheus-community/windows_exporter/pkg/collector"
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"github.com/stretchr/testify/require"
)

//nolint:gochecknoglobals
var baseDir = "../../../tools/textfile-test"

//nolint:paralleltest
func TestMultipleDirectories(t *testing.T) {
	logger := slog.New(slog.DiscardHandler)
	testDir := baseDir + "/multiple-dirs"
	testDirs := fmt.Sprintf("%[1]s/dir1,%[1]s/dir2,%[1]s/dir3", testDir)

	textFileCollector := textfile.New(&textfile.Config{
		TextFileDirectories: strings.Split(testDirs, ","),
	})

	collectors := collector.New(map[string]collector.Collector{textfile.Name: textFileCollector})
	require.NoError(t, collectors.Build(t.Context(), logger))

	metrics := make(chan prometheus.Metric)
	got := strings.Builder{}

	errCh := make(chan error, 1)

	go func() {
		errCh <- textFileCollector.Collect(metrics, 0)

		close(metrics)
	}()

	for val := range metrics {
		var metric dto.Metric

		err := val.Write(&metric)
		require.NoError(t, err)

		got.WriteString(metric.String())
	}

	require.NoError(t, <-errCh)

	for _, f := range []string{"dir1", "dir2", "dir3", "dir3sub"} {
		require.Contains(t, got.String(), f)
	}
}

//nolint:paralleltest
func TestDuplicateFileName(t *testing.T) {
	logger := slog.New(slog.DiscardHandler)
	testDir := baseDir + "/duplicate-filename"
	textFileCollector := textfile.New(&textfile.Config{
		TextFileDirectories: []string{testDir},
	})

	collectors := collector.New(map[string]collector.Collector{textfile.Name: textFileCollector})
	require.NoError(t, collectors.Build(t.Context(), logger))

	metrics := make(chan prometheus.Metric)
	got := strings.Builder{}

	errCh := make(chan error, 1)

	go func() {
		errCh <- textFileCollector.Collect(metrics, 0)

		close(metrics)
	}()

	for val := range metrics {
		var metric dto.Metric

		err := val.Write(&metric)
		require.NoError(t, err)

		got.WriteString(metric.String())
	}

	require.ErrorContains(t, <-errCh, "duplicate filename detected")

	require.Contains(t, got.String(), "file")
	require.NotContains(t, got.String(), "sub_file")
}
