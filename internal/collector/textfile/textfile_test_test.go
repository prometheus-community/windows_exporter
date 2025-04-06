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

package textfile_test

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"strings"
	"testing"

	"github.com/prometheus-community/windows_exporter/internal/collector/textfile"
	"github.com/prometheus-community/windows_exporter/pkg/collector"
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//nolint:gochecknoglobals
var baseDir = "../../../tools/textfile-test"

//nolint:paralleltest
func TestMultipleDirectories(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	testDir := baseDir + "/multiple-dirs"
	testDirs := fmt.Sprintf("%[1]s/dir1,%[1]s/dir2,%[1]s/dir3", testDir)

	textFileCollector := textfile.New(&textfile.Config{
		TextFileDirectories: strings.Split(testDirs, ","),
	})

	collectors := collector.New(map[string]collector.Collector{textfile.Name: textFileCollector})
	require.NoError(t, collectors.Build(context.Background(), logger))

	metrics := make(chan prometheus.Metric)
	got := ""

	errCh := make(chan error, 1)
	go func() {
		errCh <- textFileCollector.Collect(metrics)

		close(metrics)
	}()

	for val := range metrics {
		var metric dto.Metric

		err := val.Write(&metric)
		require.NoError(t, err)

		got += metric.String()
	}

	require.NoError(t, <-errCh)

	for _, f := range []string{"dir1", "dir2", "dir3", "dir3sub"} {
		assert.Contains(t, got, f)
	}
}

//nolint:paralleltest
func TestDuplicateFileName(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	testDir := baseDir + "/duplicate-filename"
	textFileCollector := textfile.New(&textfile.Config{
		TextFileDirectories: []string{testDir},
	})

	collectors := collector.New(map[string]collector.Collector{textfile.Name: textFileCollector})
	require.NoError(t, collectors.Build(context.Background(), logger))

	metrics := make(chan prometheus.Metric)
	got := ""

	errCh := make(chan error, 1)
	go func() {
		errCh <- textFileCollector.Collect(metrics)

		close(metrics)
	}()

	for val := range metrics {
		var metric dto.Metric

		err := val.Write(&metric)
		require.NoError(t, err)

		got += metric.String()
	}

	require.ErrorContains(t, <-errCh, "duplicate filename detected")

	assert.Contains(t, got, "file")
	assert.NotContains(t, got, "sub_file")
}
