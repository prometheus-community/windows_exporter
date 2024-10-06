//go:build windows

package testutils

import (
	"io"
	"log/slog"
	"sync"
	"testing"
	"time"

	"github.com/alecthomas/kingpin/v2"
	"github.com/prometheus-community/windows_exporter/pkg/collector"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/require"
	"github.com/yusufpapurcu/wmi"
)

func FuncBenchmarkCollector[C collector.Collector](b *testing.B, name string, collectFunc collector.BuilderWithFlags[C]) {
	b.Helper()

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	c := collectFunc(kingpin.CommandLine)
	collectors := collector.New(map[string]collector.Collector{name: c})
	require.NoError(b, collectors.Build(logger))

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
		require.NoError(b, c.Collect(scrapeContext, logger, metrics))
	}
}

func TestCollector[C collector.Collector, V interface{}](t *testing.T, fn func(*V) C, conf *V) {
	t.Helper()
	t.Setenv("WINDOWS_EXPORTER_PERF_COUNTERS_ENGINE", "pdh")

	var (
		metrics []prometheus.Metric
		err     error
	)

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	c := fn(conf)
	ch := make(chan prometheus.Metric, 10000)

	wmiClient := &wmi.Client{
		AllowMissingFields: true,
	}
	wmiClient.SWbemServicesClient, err = wmi.InitializeSWbemServices(wmiClient)
	require.NoError(t, err)

	t.Cleanup(func() {
		require.NoError(t, c.Close(logger))
	})

	wg := sync.WaitGroup{}
	wg.Add(1)

	go func() {
		defer wg.Done()

		for metric := range ch {
			metrics = append(metrics, metric)
		}
	}()

	require.NoError(t, c.Build(logger, wmiClient))

	time.Sleep(1 * time.Second)

	require.NoError(t, c.Collect(nil, logger, ch))

	close(ch)

	wg.Wait()

	require.NotEmpty(t, metrics)
}
