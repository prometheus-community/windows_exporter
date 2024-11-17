//go:build windows

package testutils

import (
	"io"
	"log/slog"
	"sync"
	"testing"
	"time"

	"github.com/alecthomas/kingpin/v2"
	"github.com/prometheus-community/windows_exporter/internal/mi"
	"github.com/prometheus-community/windows_exporter/pkg/collector"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/require"
)

func FuncBenchmarkCollector[C collector.Collector](b *testing.B, name string, collectFunc collector.BuilderWithFlags[C]) {
	b.Helper()

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	c := collectFunc(kingpin.CommandLine)
	collectors := collector.New(map[string]collector.Collector{name: c})
	require.NoError(b, collectors.Build(logger))

	metrics := make(chan prometheus.Metric)

	go func() {
		for {
			<-metrics
		}
	}()

	for i := 0; i < b.N; i++ {
		require.NoError(b, c.Collect(metrics))
	}
}

func TestCollector[C collector.Collector, V interface{}](t *testing.T, fn func(*V) C, conf *V) {
	t.Helper()

	var (
		metrics []prometheus.Metric
		err     error
	)

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	c := fn(conf)
	ch := make(chan prometheus.Metric, 10000)

	miApp, err := mi.Application_Initialize()
	require.NoError(t, err)

	miSession, err := miApp.NewSession(nil)
	require.NoError(t, err)

	t.Cleanup(func() {
		require.NoError(t, c.Close())
		require.NoError(t, miSession.Close())
		require.NoError(t, miApp.Close())
	})

	wg := sync.WaitGroup{}
	wg.Add(1)

	go func() {
		defer wg.Done()

		for metric := range ch {
			metrics = append(metrics, metric)
		}
	}()

	require.NoError(t, c.Build(logger, miSession))

	time.Sleep(1 * time.Second)

	require.NoError(t, c.Collect(ch))

	close(ch)

	wg.Wait()

	require.NotEmpty(t, metrics)
}
