// SPDX-License-Identifier: MIT
//
// Copyright The Prometheus Authors

//go:build windows

package service_test

import (
	"testing"

	"github.com/alecthomas/kingpin/v2"
	"github.com/prometheus-community/windows_exporter/internal/collector/service"
	"github.com/prometheus-community/windows_exporter/internal/utils/testutils"
	"github.com/stretchr/testify/require"
)

func BenchmarkCollector(b *testing.B) {
	testutils.FuncBenchmarkCollector(b, service.Name, service.NewWithFlags)
}

func TestCollector(t *testing.T) {
	testutils.TestCollector(t, service.New, nil)
}

func TestNewWithFlagsReportsRegexpFlagName(t *testing.T) {
	t.Parallel()

	for _, flagName := range []string{"collector.service.exclude", "collector.service.include"} {
		t.Run(flagName, func(t *testing.T) {
			t.Parallel()

			app := kingpin.New("test", "")
			service.NewWithFlags(app)

			_, err := app.Parse([]string{"--" + flagName + "=("})
			require.ErrorContains(t, err, flagName+":")
		})
	}
}
