package exchange

import (
	"errors"
	"fmt"
	"log/slog"

	"github.com/prometheus-community/windows_exporter/internal/perfdata"
	v1 "github.com/prometheus-community/windows_exporter/internal/perfdata/v1"
	"github.com/prometheus-community/windows_exporter/internal/types"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	requestsPerSec      = "Requests/sec"
	pingCommandsPending = "Ping Commands Pending"
	syncCommandsPerSec  = "Sync Commands/sec"
)

// Perflib: [25138] MSExchange ActiveSync.
type perflibActiveSync struct {
	RequestsPerSec      float64 `perflib:"Requests/sec"`
	PingCommandsPending float64 `perflib:"Ping Commands Pending"`
	SyncCommandsPerSec  float64 `perflib:"Sync Commands/sec"`
}

func (c *Collector) buildActiveSync() error {
	counters := []string{
		requestsPerSec,
		pingCommandsPending,
		syncCommandsPerSec,
	}

	var err error

	c.perfDataCollectorActiveSync, err = perfdata.NewCollector("MSExchange ActiveSync", perfdata.InstanceAll, counters)
	if err != nil {
		return fmt.Errorf("failed to create MSExchange ActiveSync collector: %w", err)
	}

	return nil
}

func (c *Collector) collectActiveSync(ctx *types.ScrapeContext, logger *slog.Logger, ch chan<- prometheus.Metric) error {
	var data []perflibActiveSync

	if err := v1.UnmarshalObject(ctx.PerfObjects["MSExchange ActiveSync"], &data, logger); err != nil {
		return err
	}

	for _, instance := range data {
		ch <- prometheus.MustNewConstMetric(
			c.activeSyncRequestsPerSec,
			prometheus.CounterValue,
			instance.RequestsPerSec,
		)
		ch <- prometheus.MustNewConstMetric(
			c.pingCommandsPending,
			prometheus.GaugeValue,
			instance.PingCommandsPending,
		)
		ch <- prometheus.MustNewConstMetric(
			c.syncCommandsPerSec,
			prometheus.CounterValue,
			instance.SyncCommandsPerSec,
		)
	}

	return nil
}

func (c *Collector) collectPDHActiveSync(ch chan<- prometheus.Metric) error {
	perfData, err := c.perfDataCollectorActiveSync.Collect()
	if err != nil {
		return fmt.Errorf("failed to collect MSExchange ActiveSync metrics: %w", err)
	}

	if len(perfData) == 0 {
		return errors.New("perflib query for MSExchange ActiveSync returned empty result set")
	}

	for _, data := range perfData {
		ch <- prometheus.MustNewConstMetric(
			c.activeSyncRequestsPerSec,
			prometheus.CounterValue,
			data[requestsPerSec].FirstValue,
		)
		ch <- prometheus.MustNewConstMetric(
			c.pingCommandsPending,
			prometheus.GaugeValue,
			data[pingCommandsPending].FirstValue,
		)
		ch <- prometheus.MustNewConstMetric(
			c.syncCommandsPerSec,
			prometheus.CounterValue,
			data[syncCommandsPerSec].FirstValue,
		)
	}

	return nil
}
