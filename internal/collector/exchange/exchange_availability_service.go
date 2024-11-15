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

// Perflib: [24914] MSExchange Availability Service.
type perflibAvailabilityService struct {
	RequestsSec float64 `perflib:"Availability Requests (sec)"`
}

func (c *Collector) buildAvailabilityService() error {
	counters := []string{}

	var err error

	c.perfDataCollectorAvailabilityService, err = perfdata.NewCollector(perfdata.V2, "MSExchange Availability Service", perfdata.AllInstances, counters)
	if err != nil {
		return fmt.Errorf("failed to create MSExchange Availability Service collector: %w", err)
	}

	return nil
}

func (c *Collector) collectAvailabilityService(ctx *types.ScrapeContext, logger *slog.Logger, ch chan<- prometheus.Metric) error {
	var data []perflibAvailabilityService

	if err := v1.UnmarshalObject(ctx.PerfObjects["MSExchange Availability Service"], &data, logger); err != nil {
		return err
	}

	for _, availservice := range data {
		ch <- prometheus.MustNewConstMetric(
			c.availabilityRequestsSec,
			prometheus.CounterValue,
			availservice.RequestsSec,
		)
	}

	return nil
}

func (c *Collector) collectPDHAvailabilityService(ch chan<- prometheus.Metric) error {
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
