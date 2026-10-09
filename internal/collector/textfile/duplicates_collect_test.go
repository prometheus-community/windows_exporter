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

package textfile

import (
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/require"
)

func TestDuplicateAcrossFilesPreservesOtherMetrics(t *testing.T) {
	t.Parallel()

	directory := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(directory, "a.prom"), []byte("# HELP probe Shared probe.\nprobe{source=\"same\"} 1\n"), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(directory, "b.prom"), []byte("# HELP probe Shared probe.\nprobe{source=\"same\"} 2\nother 9\n"), 0o600))
	c := New(&Config{TextFileDirectories: []string{directory}})
	require.NoError(t, c.Build(slog.New(slog.DiscardHandler), nil))
	wrapper := &duplicateTestCollector{collector: c}
	registry := prometheus.NewRegistry()
	registry.MustRegister(wrapper)
	families, err := registry.Gather()
	require.NoError(t, err)
	require.ErrorContains(t, wrapper.err, "duplicate metric")
	require.ErrorContains(t, wrapper.err, "probe")
	require.ErrorContains(t, wrapper.err, "a.prom")
	require.ErrorContains(t, wrapper.err, "b.prom")

	seen := make(map[string]bool)
	for _, family := range families {
		seen[family.GetName()] = true
		switch family.GetName() {
		case "other":
			require.InDelta(t, 9, family.GetMetric()[0].GetUntyped().GetValue(), 1e-9)
		case "probe":
			require.Len(t, family.GetMetric(), 1)
			require.InDelta(t, 1, family.GetMetric()[0].GetUntyped().GetValue(), 1e-9)
		case "windows_textfile_mtime_seconds":
			require.Len(t, family.GetMetric(), 1)
			require.Equal(t, "a.prom", family.GetMetric()[0].GetLabel()[0].GetValue())
		}
	}

	require.True(t, seen["other"])
	require.True(t, seen["probe"])
	require.True(t, seen["windows_textfile_mtime_seconds"])
}

type duplicateTestCollector struct {
	collector *Collector
	err       error
}

func (c *duplicateTestCollector) Describe(_ chan<- *prometheus.Desc) {}

func (c *duplicateTestCollector) Collect(ch chan<- prometheus.Metric) {
	c.err = c.collector.Collect(ch, 0)
}
