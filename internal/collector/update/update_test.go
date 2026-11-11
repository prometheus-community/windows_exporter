// SPDX-License-Identifier: MIT
//
// Copyright The Prometheus Authors

//go:build windows

package update_test

import (
	"testing"

	"github.com/prometheus-community/windows_exporter/internal/collector/update"
	"github.com/prometheus-community/windows_exporter/internal/utils/testutils"
)

func BenchmarkCollector(b *testing.B) {
	testutils.FuncBenchmarkCollector(b, update.Name, update.NewWithFlags)
}

func TestCollector(t *testing.T) {
	testutils.TestCollector(t, update.New, nil)
}
