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
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/require"
)

func TestSplitFamilyWithoutHelp(t *testing.T) {
	t.Parallel()

	directory := t.TempDir()

	for _, source := range []string{"a", "b"} {
		content := fmt.Sprintf("probe{source=%q} 1\n", source)
		require.NoError(t, os.WriteFile(filepath.Join(directory, source+".prom"), []byte(content), 0o600))
	}

	c := New(&Config{TextFileDirectories: []string{directory}})
	require.NoError(t, c.Build(slog.New(slog.DiscardHandler), nil))
	wrapper := &helpTestCollector{collector: c}
	registry := prometheus.NewRegistry()
	registry.MustRegister(wrapper)
	families, err := registry.Gather()
	require.NoError(t, err)
	require.NoError(t, wrapper.err)

	found := false

	for _, family := range families {
		if family.GetName() == "probe" {
			found = true

			require.Len(t, family.GetMetric(), 2)
			require.NotContains(t, family.GetHelp(), directory)
		}
	}

	require.True(t, found)
}

type helpTestCollector struct {
	collector *Collector
	err       error
}

func (c *helpTestCollector) Describe(_ chan<- *prometheus.Desc) {}

func (c *helpTestCollector) Collect(ch chan<- prometheus.Metric) {
	c.err = c.collector.Collect(ch, 0)
}
