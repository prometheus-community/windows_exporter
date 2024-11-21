//go:build windows

package exchange

import (
	"errors"
	"fmt"

	"github.com/prometheus-community/windows_exporter/internal/perfdata"
	"github.com/prometheus-community/windows_exporter/internal/types"
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

	c.activeUserCountMapiHttpEmsMDB = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "mapihttp_emsmdb_active_user_count"),
		"Number of unique outlook users that have shown some kind of activity in the last 2 minutes",
		nil,
		nil,
	)

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
