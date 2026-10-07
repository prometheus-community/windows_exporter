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

package file_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/prometheus-community/windows_exporter/internal/collector/file"
	"github.com/prometheus-community/windows_exporter/internal/utils/testutils"
	"github.com/stretchr/testify/require"
)

func BenchmarkCollector(b *testing.B) {
	testutils.FuncBenchmarkCollector(b, file.Name, file.NewWithFlags)
}

func TestCollector(t *testing.T) {
	path := filepath.Join(t.TempDir(), "test.txt")
	content := []byte("windows_exporter test")
	require.NoError(t, os.WriteFile(path, content, 0o600))

	metrics := testutils.TestCollector(t, file.New, &file.Config{
		FilePatterns: []string{path},
	})
	require.Contains(t, metrics, "windows_file_size_bytes")
	require.Len(t, metrics["windows_file_size_bytes"].GetMetric(), 1)
	require.InDelta(t, len(content), metrics["windows_file_size_bytes"].GetMetric()[0].GetGauge().GetValue(), 0)
	require.Contains(t, metrics, "windows_file_mtime_timestamp_seconds")
	require.Positive(t, metrics["windows_file_mtime_timestamp_seconds"].GetMetric()[0].GetGauge().GetValue())
}

func TestCollectorPatterns(t *testing.T) {
	t.Parallel()

	directory := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(directory, "first.txt"), []byte("first"), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(directory, "second.log"), []byte("second"), 0o600))

	for _, tc := range []struct {
		name    string
		pattern string
		count   int
	}{
		{name: "all files", pattern: "*", count: 2},
		{name: "extension", pattern: "*.txt", count: 1},
		{name: "case insensitive", pattern: "FIRST.TXT", count: 1},
		{name: "missing file", pattern: "missing.txt", count: 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			metrics := testutils.TestCollector(t, file.New, &file.Config{
				FilePatterns: []string{filepath.Join(directory, tc.pattern)},
			})
			if tc.count == 0 {
				require.Empty(t, metrics)

				return
			}

			require.Contains(t, metrics, "windows_file_size_bytes")
			require.Len(t, metrics["windows_file_size_bytes"].GetMetric(), tc.count)
		})
	}
}
