// SPDX-License-Identifier: MIT
//
// Copyright The Prometheus Authors

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
