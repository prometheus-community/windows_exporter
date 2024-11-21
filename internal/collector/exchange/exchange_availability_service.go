//go:build windows

package exchange

import (
	"errors"
	"fmt"

	"github.com/prometheus-community/windows_exporter/internal/perfdata"
	"github.com/prometheus-community/windows_exporter/internal/types"
	"github.com/prometheus/client_golang/prometheus"
)

func (c *Collector) buildAvailabilityService() error {
	counters := []string{
		requestsPerSec,
	}

	var err error

	c.perfDataCollectorAvailabilityService, err = perfdata.NewCollector("MSExchange Availability Service", perfdata.InstanceAll, counters)
	if err != nil {
		return fmt.Errorf("failed to create MSExchange Availability Service collector: %w", err)
	}

	c.availabilityRequestsSec = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "avail_service_requests_per_sec"),
		"Number of requests serviced per second",
		nil,
		nil,
	)

	return nil
}

func (c *Collector) collectAvailabilityService(ch chan<- prometheus.Metric) error {
	perfData, err := c.perfDataCollectorAvailabilityService.Collect()
	if err != nil {
		return fmt.Errorf("failed to collect MSExchange Availability Service metrics: %w", err)
	}

	if len(perfData) == 0 {
		return errors.New("perflib query for MSExchange Availability Service returned empty result set")
	}

	for _, data := range perfData {
		ch <- prometheus.MustNewConstMetric(
			c.availabilityRequestsSec,
			prometheus.CounterValue,
			data[requestsPerSec].FirstValue,
		)
	}

	return nil
}
