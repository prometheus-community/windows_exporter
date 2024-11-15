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
	currentUniqueUsers = "Current Unique Users"
	// requestsPerSec     = "Requests/sec"
)

// Perflib: [24618] MSExchange OWA.
type perflibOWA struct {
	CurrentUniqueUsers float64 `perflib:"Current Unique Users"`
	RequestsPerSec     float64 `perflib:"Requests/sec"`
}

func (c *Collector) buildOWA() error {
	counters := []string{
		currentUniqueUsers,
		requestsPerSec,
	}

	var err error

	c.perfDataCollectorOWA, err = perfdata.NewCollector(perfdata.V2, "MSExchange OWA", perfdata.AllInstances, counters)
	if err != nil {
		return fmt.Errorf("failed to create MSExchange OWA collector: %w", err)
	}

	return nil
}

func (c *Collector) collectOWA(ctx *types.ScrapeContext, logger *slog.Logger, ch chan<- prometheus.Metric) error {
	var data []perflibOWA

	if err := v1.UnmarshalObject(ctx.PerfObjects["MSExchange OWA"], &data, logger); err != nil {
		return err
	}

	for _, owa := range data {
		ch <- prometheus.MustNewConstMetric(
			c.currentUniqueUsers,
			prometheus.GaugeValue,
			owa.CurrentUniqueUsers,
		)
		ch <- prometheus.MustNewConstMetric(
			c.owaRequestsPerSec,
			prometheus.CounterValue,
			owa.RequestsPerSec,
		)
	}

	return nil
}

func (c *Collector) collectPDHOWA(ch chan<- prometheus.Metric) error {
	perfData, err := c.perfDataCollectorOWA.Collect()
	if err != nil {
		return fmt.Errorf("failed to collect MSExchange OWA metrics: %w", err)
	}

	if len(perfData) == 0 {
		return errors.New("perflib query for MSExchange OWA returned empty result set")
	}

	for _, data := range perfData {
		ch <- prometheus.MustNewConstMetric(
			c.currentUniqueUsers,
			prometheus.GaugeValue,
			data[currentUniqueUsers].FirstValue,
		)
		ch <- prometheus.MustNewConstMetric(
			c.owaRequestsPerSec,
			prometheus.CounterValue,
			data[requestsPerSec].FirstValue,
		)
	}

	return nil
}
