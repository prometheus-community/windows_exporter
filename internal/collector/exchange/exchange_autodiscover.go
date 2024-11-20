//go:build windows

package exchange

import (
	"errors"
	"fmt"

	"github.com/prometheus-community/windows_exporter/internal/perfdata"
	"github.com/prometheus-community/windows_exporter/internal/types"
	"github.com/prometheus/client_golang/prometheus"
)

func (c *Collector) buildAutoDiscover() error {
	counters := []string{
		requestsPerSec,
	}

	var err error

	c.perfDataCollectorAutoDiscover, err = perfdata.NewCollector("MSExchange Autodiscover", perfdata.InstanceAll, counters)
	if err != nil {
		return fmt.Errorf("failed to create MSExchange Autodiscover collector: %w", err)
	}

	c.autoDiscoverRequestsPerSec = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "autodiscover_requests_total"),
		"Number of autodiscover service requests processed each second",
		nil,
		nil,
	)

	return nil
}

func (c *Collector) collectAutoDiscover(ch chan<- prometheus.Metric) error {
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
