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

// [29240] MSExchangeAutodiscover.
type perflibAutodiscover struct {
	RequestsPerSec float64 `perflib:"Requests/sec"`
}

func (c *Collector) buildAutoDiscover() error {
	counters := []string{
		requestsPerSec,
	}

	var err error

	c.perfDataCollectorAutoDiscover, err = perfdata.NewCollector(perfdata.V1, "MSExchange Autodiscover", perfdata.AllInstances, counters)
	if err != nil {
		return fmt.Errorf("failed to create MSExchange Autodiscover collector: %w", err)
	}

	return nil
}

func (c *Collector) collectAutoDiscover(ctx *types.ScrapeContext, logger *slog.Logger, ch chan<- prometheus.Metric) error {
	var data []perflibAutodiscover

	if err := v1.UnmarshalObject(ctx.PerfObjects["MSExchangeAutodiscover"], &data, logger); err != nil {
		return err
	}

	for _, autodisc := range data {
		ch <- prometheus.MustNewConstMetric(
			c.autoDiscoverRequestsPerSec,
			prometheus.CounterValue,
			autodisc.RequestsPerSec,
		)
	}

	return nil
}

func (c *Collector) collectPDHAutoDiscover(ch chan<- prometheus.Metric) error {
	perfData, err := c.perfDataCollectorAutoDiscover.Collect()
	if err != nil {
		return fmt.Errorf("failed to collect MSExchange Autodiscover metrics: %w", err)
	}

	if len(perfData) == 0 {
		return errors.New("perflib query for MSExchange Autodiscover returned empty result set")
	}

	for _, data := range perfData {
		ch <- prometheus.MustNewConstMetric(
			c.autoDiscoverRequestsPerSec,
			prometheus.CounterValue,
			data[requestsPerSec].FirstValue,
		)
	}

	return nil
}
