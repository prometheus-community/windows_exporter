//go:build windows

package exchange

import (
	"errors"
	"fmt"

	"github.com/prometheus-community/windows_exporter/internal/perfdata"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	activeUserCount = "Active User Count"
)

func (c *Collector) buildMapiHttpEmsmdb() error {
	counters := []string{
		activeUserCount,
	}

	var err error

	c.perfDataCollectorMapiHttpEmsmdb, err = perfdata.NewCollector("MSExchange MapiHttp Emsmdb", perfdata.InstanceAll, counters)
	if err != nil {
		return fmt.Errorf("failed to create MSExchange MapiHttp Emsmdb: %w", err)
	}

	return nil
}

func (c *Collector) collectMapiHttpEmsmdb(ch chan<- prometheus.Metric) error {
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
