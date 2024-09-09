//go:build windows

package smb

import (
	"log/slog"
	"strings"

	"github.com/alecthomas/kingpin/v2"
	"github.com/prometheus-community/windows_exporter/pkg/perflib"
	"github.com/prometheus-community/windows_exporter/pkg/types"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/yusufpapurcu/wmi"
)

const Name = "smb"

type Config struct{}

var ConfigDefaults = Config{}

type Collector struct {
	config Config

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

func (c *Collector) GetPerfCounter(_ *slog.Logger) ([]string, error) {
	return []string{
		"SMB Server Shares",
	}, nil
}

func (c *Collector) Close(_ *slog.Logger) error {
	return nil
}

func (c *Collector) Build(_ *slog.Logger, _ *wmi.Client) error {
	// desc creates a new prometheus description
	desc := func(metricName string, description string, labels ...string) *prometheus.Desc {
		return prometheus.NewDesc(
			prometheus.BuildFQName(types.Namespace, "smb", metricName),
			description,
			labels,
			nil,
		)
	}

	c.currentOpenFileCount = desc("server_shares_current_open_file_count", "Current total count open files on the SMB Server")
	c.treeConnectCount = desc("server_shares_tree_connect_count", "Count of user connections to the SMB Server")

	return nil
}

// Collect collects smb metrics and sends them to prometheus.
func (c *Collector) Collect(ctx *types.ScrapeContext, logger *slog.Logger, ch chan<- prometheus.Metric) error {
	logger = logger.With(slog.String("collector", Name))
	if err := c.collectServerShares(ctx, logger, ch); err != nil {
		logger.Error("failed to collect server share metrics",
			slog.Any("err", err),
		)

		return err
	}

	return nil
}

// Perflib: SMB Server Shares.
type perflibServerShares struct {
	Name string

	CurrentOpenFileCount float64 `perflib:"Current Open File Count"`
	TreeConnectCount     float64 `perflib:"Tree Connect Count"`
}

func (c *Collector) collectServerShares(ctx *types.ScrapeContext, logger *slog.Logger, ch chan<- prometheus.Metric) error {
	logger = logger.With(slog.String("collector", Name))

	var data []perflibServerShares

	if err := perflib.UnmarshalObject(ctx.PerfObjects["SMB Server Shares"], &data, logger); err != nil {
		return err
	}

	for _, instance := range data {
		labelName := c.toLabelName(instance.Name)
		if !strings.HasSuffix(labelName, "_total") {
			continue
		}

		ch <- prometheus.MustNewConstMetric(
			c.currentOpenFileCount,
			prometheus.CounterValue,
			instance.CurrentOpenFileCount,
		)

		ch <- prometheus.MustNewConstMetric(
			c.treeConnectCount,
			prometheus.CounterValue,
			instance.TreeConnectCount,
		)
	}

	return nil
}

// toLabelName converts strings to lowercase and replaces all whitespaces and dots with underscores.
func (c *Collector) toLabelName(name string) string {
	s := strings.ReplaceAll(strings.Join(strings.Fields(strings.ToLower(name)), "_"), ".", "_")
	s = strings.ReplaceAll(s, "__", "_")

	return s
}
