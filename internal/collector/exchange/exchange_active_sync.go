//go:build windows

package exchange

import (
	"errors"
	"fmt"

	"github.com/prometheus-community/windows_exporter/internal/perfdata"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	requestsPerSec      = "Requests/sec"
	pingCommandsPending = "Ping Commands Pending"
	syncCommandsPerSec  = "Sync Commands/sec"
)

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

func (c *Collector) collectActiveSync(ch chan<- prometheus.Metric) error {
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
