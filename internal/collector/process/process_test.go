package process_test

import (
	"io"
	"log/slog"
	"sync"
	"testing"
	"time"

	"github.com/alecthomas/kingpin/v2"
	"github.com/prometheus-community/windows_exporter/internal/collector/process"
	"github.com/prometheus-community/windows_exporter/internal/testutils"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/require"
	"github.com/yusufpapurcu/wmi"
)

func BenchmarkProcessCollector(b *testing.B) {
	// PrinterInclude is not set in testing context (kingpin flags not parsed), causing the collector to skip all processes.
	localProcessInclude := ".+"
	kingpin.CommandLine.GetArg("collector.process.include").StringVar(&localProcessInclude)
	// No context name required as collector source is WMI
	testutils.FuncBenchmarkCollector(b, process.Name, process.NewWithFlags)
}

func TestProcessCollector(t *testing.T) {
	t.Setenv("WINDOWS_EXPORTER_PERF_COUNTERS_ENGINE", "v2")

	var (
		metrics []prometheus.Metric
		err     error
	)

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	c := process.New(nil)
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
