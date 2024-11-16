//go:build windows

package smb

import (
	"errors"
	"fmt"
	"log/slog"

	"github.com/alecthomas/kingpin/v2"
	"github.com/prometheus-community/windows_exporter/internal/mi"
	"github.com/prometheus-community/windows_exporter/internal/perfdata"
	"github.com/prometheus-community/windows_exporter/internal/types"
	"github.com/prometheus/client_golang/prometheus"
)

const Name = "smb"

type Config struct{}

var ConfigDefaults = Config{}

type Collector struct {
	config Config

	perfDataCollector *perfdata.Collector

	treeConnectCount     *prometheus.Desc
	currentOpenFileCount *prometheus.Desc
}

func New(config *Config) *Collector {
	if config == nil {
		config = &ConfigDefaults
	}

	c := &Collector{
		config: *config,
	}

	return c
}

func NewWithFlags(_ *kingpin.Application) *Collector {
	return &Collector{}
}

func (c *Collector) GetName() string {
	return Name
}

func (c *Collector) Close() error {
	c.perfDataCollector.Close()

	return nil
}

func (c *Collector) Build(_ *slog.Logger, _ *mi.Session) error {
	var err error

	c.perfDataCollector, err = perfdata.NewCollector("SMB Server Shares", nil, []string{
		currentOpenFileCount,
		treeConnectCount,
	})
	if err != nil {
		return fmt.Errorf("failed to create SMB Server Shares collector: %w", err)
	}

	c.currentOpenFileCount = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "server_shares_current_open_file_count"),
		"Current total count open files on the SMB Server",
		nil,
		nil,
	)
	c.treeConnectCount = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "server_shares_tree_connect_count"),
		"Count of user connections to the SMB Server",
		nil,
		nil,
	)

	return nil
}

// Collect collects smb metrics and sends them to prometheus.
func (c *Collector) Collect(ch chan<- prometheus.Metric) error {
	perfData, err := c.perfDataCollector.Collect()
	if err != nil {
		return fmt.Errorf("failed to collect SMB Server Shares metrics: %w", err)
	}

	data, ok := perfData[perfdata.EmptyInstance]
	if !ok {
		return errors.New("query for SMB Server Shares returned empty result set")
	}

	ch <- prometheus.MustNewConstMetric(
		c.currentOpenFileCount,
		prometheus.CounterValue,
		data[currentOpenFileCount].FirstValue,
	)

	ch <- prometheus.MustNewConstMetric(
		c.treeConnectCount,
		prometheus.CounterValue,
		data[treeConnectCount].FirstValue,
	)

	return nil
}
