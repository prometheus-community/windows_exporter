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
	activeUserCount = "Active User Count"
)

// perflib [26463] MSExchange MapiHttp Emsmdb.
type perflibMapiHttpEmsmdb struct {
	ActiveUserCount float64 `perflib:"Active User Count"`
}

func (c *Collector) buildMapiHttpEmsmdb() error {
	counters := []string{
		activeUserCount,
	}

	var err error

	c.perfDataCollectorMapiHttpEmsmdb, err = perfdata.NewCollector(perfdata.V1, "MSExchange MapiHttp Emsmdb", perfdata.AllInstances, counters)
	if err != nil {
		return fmt.Errorf("failed to create MSExchange MapiHttp Emsmdb: %w", err)
	}

	return nil
}

func (c *Collector) collectMapiHttpEmsmdb(ctx *types.ScrapeContext, logger *slog.Logger, ch chan<- prometheus.Metric) error {
	var data []perflibMapiHttpEmsmdb

	if err := v1.UnmarshalObject(ctx.PerfObjects["MSExchange MapiHttp Emsmdb"], &data, logger); err != nil {
		return err
	}

	for _, mapihttp := range data {
		ch <- prometheus.MustNewConstMetric(
			c.activeUserCountMapiHttpEmsMDB,
			prometheus.GaugeValue,
			mapihttp.ActiveUserCount,
		)
	}

	return nil
}

func (c *Collector) collectPDHMapiHttpEmsmdb(ch chan<- prometheus.Metric) error {
	perfData, err := c.perfDataCollectorMapiHttpEmsmdb.Collect()
	if err != nil {
		return fmt.Errorf("failed to collect MSExchange MapiHttp Emsmdb metrics: %w", err)
	}

	if len(perfData) == 0 {
		return errors.New("perflib query for MSExchange MapiHttp Emsmdb returned empty result set")
	}

	for _, data := range perfData {
		ch <- prometheus.MustNewConstMetric(
			c.activeUserCountMapiHttpEmsMDB,
			prometheus.GaugeValue,
			data[activeUserCount].FirstValue,
		)
	}

	return nil
}
