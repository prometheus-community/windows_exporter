//go:build windows

package perfdata_test

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"

	"github.com/prometheus-community/windows_exporter/internal/collector/perfdata"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type collectorAdapter struct {
	perfdata.Collector
}

// Describe implements the prometheus.Collector interface.
func (a collectorAdapter) Describe(_ chan<- *prometheus.Desc) {}

// Collect implements the prometheus.Collector interface.
func (a collectorAdapter) Collect(ch chan<- prometheus.Metric) {
	if err := a.Collector.Collect(ch); err != nil {
		panic(fmt.Sprintf("failed to update collector: %v", err))
	}
}

func TestCollector(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		object          string
		instances       []string
		counters        map[string]perfdata.Counter
		expectedMetrics *regexp.Regexp
	}{
		{
			object:          "Memory",
			instances:       nil,
			counters:        map[string]perfdata.Counter{"Available Bytes": {Type: "gauge"}},
			expectedMetrics: regexp.MustCompile(`^# HELP windows_perfdata_memory_available_bytes Performance data for \\\\Memory\\\\Available Bytes\s*# TYPE windows_perfdata_memory_available_bytes gauge\s*windows_perfdata_memory_available_bytes \d`),
		},
		{
			object:          "Process",
			instances:       []string{"*"},
			counters:        map[string]perfdata.Counter{"Thread Count": {Type: "counter"}},
			expectedMetrics: regexp.MustCompile(`^# HELP windows_perfdata_process_thread_count Performance data for \\\\Process\\\\Thread Count\s*# TYPE windows_perfdata_process_thread_count counter\s*windows_perfdata_process_thread_count\{instance=".+"} \d`),
		},
	} {
		t.Run(tc.object, func(t *testing.T) {
			t.Parallel()

			perfDataCollector := perfdata.New(&perfdata.Config{
				Objects: []perfdata.Object{
					{
						Object:    tc.object,
						Instances: tc.instances,
						Counters:  tc.counters,
					},
				},
			})

			logger := slog.New(slog.NewTextHandler(io.Discard, nil))
			err := perfDataCollector.Build(logger, nil)
			require.NoError(t, err)

			registry := prometheus.NewRegistry()
			registry.MustRegister(collectorAdapter{*perfDataCollector})

			rw := httptest.NewRecorder()
			promhttp.HandlerFor(registry, promhttp.HandlerOpts{ErrorHandling: promhttp.ContinueOnError}).ServeHTTP(rw, &http.Request{})
			got := rw.Body.String()

			assert.NotEmpty(t, got)
			assert.Regexp(t, tc.expectedMetrics, got)
		})
	}
}
