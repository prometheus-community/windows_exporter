// SPDX-License-Identifier: MIT
//
// Copyright The Prometheus Authors

//go:build windows

package netframework_test

import (
	"testing"

	"github.com/prometheus-community/windows_exporter/internal/collector/netframework"
	"github.com/prometheus-community/windows_exporter/internal/utils/testutils"
)

func BenchmarkCollector(b *testing.B) {
	testutils.FuncBenchmarkCollector(b, netframework.Name, netframework.NewWithFlags)
}

func TestCollector(t *testing.T) {
	testutils.TestCollector(t, netframework.New, nil)
}
