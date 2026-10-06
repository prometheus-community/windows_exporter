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

	"github.com/stretchr/testify/require"

	"github.com/prometheus-community/windows_exporter/internal/collector/file"
	"github.com/prometheus-community/windows_exporter/internal/utils/testutils"
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
	require.Equal(t, float64(len(content)), metrics["windows_file_size_bytes"].GetMetric()[0].GetGauge().GetValue())
	require.Contains(t, metrics, "windows_file_mtime_timestamp_seconds")
	require.Positive(t, metrics["windows_file_mtime_timestamp_seconds"].GetMetric()[0].GetGauge().GetValue())
}
