//go:build windows

package testutils

import (
	"testing"

	"github.com/alecthomas/kingpin/v2"
	"github.com/go-kit/log"
	"github.com/prometheus-community/windows_exporter/pkg/collector"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/require"
)

func FuncBenchmarkCollector[C collector.Collector](b *testing.B, name string, collectFunc collector.BuilderWithFlags[C]) {
	c := collectFunc(kingpin.CommandLine)
	collectors := collector.New(map[string]collector.Collector{name: c})
	require.NoError(b, collectors.Build())
	collectors.SetLogger(log.NewNopLogger())

	// Create perflib scrape context.
	// Some perflib collectors required a correct context,
	// or will fail during benchmark.
	scrapeContext, err := collectors.PrepareScrapeContext()
	require.NoError(b, err)

	metrics := make(chan prometheus.Metric)
	go func() {
		for {
			<-metrics
		}
	}()

	for i := 0; i < b.N; i++ {
		require.NoError(b, c.Collect(scrapeContext, metrics))
	}
}
