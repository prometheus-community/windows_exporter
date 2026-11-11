// SPDX-License-Identifier: MIT
//
// Copyright The Prometheus Authors

//go:build windows

package exchange_test

import (
	"testing"

	"github.com/prometheus-community/windows_exporter/internal/collector/exchange"
	"github.com/prometheus-community/windows_exporter/internal/utils/testutils"
)

func BenchmarkCollector(b *testing.B) {
	testutils.FuncBenchmarkCollector(b, exchange.Name, exchange.NewWithFlags)
}

func TestCollector(t *testing.T) {
	testutils.TestCollector(t, exchange.New, nil)
}
