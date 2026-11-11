// SPDX-License-Identifier: MIT
//
// Copyright The Prometheus Authors

//go:build windows

package ad_test

import (
	"testing"

	"github.com/prometheus-community/windows_exporter/internal/collector/ad"
	"github.com/prometheus-community/windows_exporter/internal/utils/testutils"
)

func BenchmarkCollector(b *testing.B) {
	testutils.FuncBenchmarkCollector(b, ad.Name, ad.NewWithFlags)
}

func TestCollector(t *testing.T) {
	testutils.TestCollector(t, ad.New, nil)
}
