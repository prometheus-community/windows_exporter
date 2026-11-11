// SPDX-License-Identifier: MIT
//
// Copyright The Prometheus Authors

//go:build windows

package performancecounter_test

import (
	"testing"

	"github.com/alecthomas/kingpin/v2"
	"github.com/prometheus-community/windows_exporter/internal/collector/performancecounter"
	"github.com/prometheus-community/windows_exporter/internal/utils/testutils"
	"github.com/stretchr/testify/require"
)

func BenchmarkCollector(b *testing.B) {
	perfDataObjects := `[{"name":"processor_information","object":"Processor Information","instances":["*"],"instance_label":"core","counters":[{"name":"% Processor Time","metric":"windows_performancecounter_processor_information_processor_time","labels":{"state":"active"}},{"name":"% Idle Time","metric":"windows_performancecounter_processor_information_processor_time","labels":{"state":"idle"}}]},{"name":"memory","object":"Memory","counters":[{"name":"Cache Faults/sec","type":"counter"}]}]`

	testutils.FuncBenchmarkCollector(b, performancecounter.Name, performancecounter.NewWithFlags, func(app *kingpin.Application) {
		// The objects are only read by the application action, which runs on parse.
		_, err := app.Parse([]string{"--collector.performancecounter.objects", perfDataObjects})
		require.NoError(b, err)
	})
}
