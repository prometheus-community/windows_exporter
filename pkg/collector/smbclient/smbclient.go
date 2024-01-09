//go:build windows

package smbclient

import (
	"fmt"
	"os"
	"slices"
	"strings"

	"github.com/alecthomas/kingpin/v2"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/prometheus-community/windows_exporter/pkg/perflib"
	"github.com/prometheus-community/windows_exporter/pkg/types"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	Name                           = "smbclient"
	FlagSmbClientListAllCollectors = "collectors.smbclient.list"
	FlagSmbClientCollectorsEnabled = "collectors.smbclient.enabled"
)

type Config struct {
	CollectorsEnabled string `yaml:"collectors_enabled"`
}

var ConfigDefaults = Config{
	CollectorsEnabled: "",
}

type collector struct {
	logger log.Logger

	smbclientListAllCollectors *bool
	smbclientCollectorsEnabled *string

	TreeConnectCount     *prometheus.Desc
	CurrentOpenFileCount *prometheus.Desc

	enabledCollectors []string
}

// All available collector functions
var smbclientAllCollectorNames = []string{
	"ServerShares",
}

func New(logger log.Logger, config *Config) types.Collector {
	if config == nil {
		config = &ConfigDefaults
	}

	smbclientListAllCollectors := false
	c := &collector{
		smbclientCollectorsEnabled: &config.CollectorsEnabled,
		smbclientListAllCollectors: &smbclientListAllCollectors,
	}
	c.SetLogger(logger)
	return c
}

func NewWithFlags(app *kingpin.Application) types.Collector {
	return &collector{
		smbclientListAllCollectors: app.Flag(
			FlagSmbClientListAllCollectors,
			"List the collectors along with their perflib object name/ids",
		).Bool(),

		smbclientCollectorsEnabled: app.Flag(
			FlagSmbClientCollectorsEnabled,
			"Comma-separated list of collectors to use. Defaults to all, if not specified.",
		).Default(ConfigDefaults.CollectorsEnabled).String(),
	}
}

func (c *collector) GetName() string {
	return Name
}

func (c *collector) SetLogger(logger log.Logger) {
	c.logger = log.With(logger, "collector", Name)
}

func (c *collector) GetPerfCounter() ([]string, error) {
	return []string{
		"SMB Client Shares",
	}, nil
}

func (c *collector) Build() error {
	// desc creates a new prometheus description
	desc := func(metricName string, description string, labels ...string) *prometheus.Desc {
		return prometheus.NewDesc(
			prometheus.BuildFQName(types.Namespace, "smbclient", metricName),
			description,
			labels,
			nil,
		)
	}

	//	c.CurrentOpenFileCount = desc("server_shares_current_open_file_count", "Current total count open files on the SMB Server")
	//	c.TreeConnectCount = desc("server_shares_tree_connect_count", "Count of user connections to the SMB Server")

	c.AvgSecPerRead = desc("client_shares_avg_sec_per_read", "The average latency between the time a read request is sent and when its response is received.")
	c.AvgSecPerRead = desc("client_shares_avg_sec_per_write", "The average latency between the time a write request is sent and when its response is received.")

	c.enabledCollectors = make([]string, 0, len(smbclientAllCollectorNames))

	collectorDesc := map[string]string{
		"ClientShares": "SMB Client Shares",
	}

	if *c.smbclientListAllCollectors {
		fmt.Printf("%-32s %-32s\n", "Collector Name", "Perflib Object")
		for _, cname := range smbclientAllCollectorNames {
			fmt.Printf("%-32s %-32s\n", cname, collectorDesc[cname])
		}
		os.Exit(0)
	}

	if *c.smbclientCollectorsEnabled == "" {
		for _, collectorName := range smbclientAllCollectorNames {
			c.enabledCollectors = append(c.enabledCollectors, collectorName)
		}
	} else {
		for _, collectorName := range strings.Split(*c.smbclientCollectorsEnabled, ",") {
			if slices.Contains(smbclientAllCollectorNames, collectorName) {
				c.enabledCollectors = append(c.enabledCollectors, collectorName)
			} else {
				return fmt.Errorf("unknown smbclient collector: %s", collectorName)
			}
		}
	}

	return nil
}

// Collect collects smb client metrics and sends them to prometheus
func (c *collector) Collect(ctx *types.ScrapeContext, ch chan<- prometheus.Metric) error {
	collectorFuncs := map[string]func(ctx *types.ScrapeContext, ch chan<- prometheus.Metric) error{
		"ServerShares": c.collectServerShares,
	}

	for _, collectorName := range c.enabledCollectors {
		if err := collectorFuncs[collectorName](ctx, ch); err != nil {
			_ = level.Error(c.logger).Log("msg", "Error in "+collectorName, "err", err)
			return err
		}
	}
	return nil
}

// Perflib: SMB Client Shares
type perflibServerShares struct {
	Name string

	AvgSecPerRead  float64 `perflib:"Avg. sec/Read"`
	AvgSecPerWrite float64 `perflib:"Avg. sec/Write"`
}

func (c *collector) collectServerShares(ctx *types.ScrapeContext, ch chan<- prometheus.Metric) error {
	var data []perflibServerShares
	if err := perflib.UnmarshalObject(ctx.PerfObjects["SMB Client Shares"], &data, c.logger); err != nil {
		return err
	}
	for _, instance := range data {
		labelName := c.toLabelName(instance.Name)
		if !strings.HasSuffix(labelName, "_total") {
			continue
		}

		ch <- prometheus.MustNewConstMetric(
			c.CurrentOpenFileCount,
			prometheus.CounterValue,
			instance.CurrentOpenFileCount,
		)

		ch <- prometheus.MustNewConstMetric(
			c.TreeConnectCount,
			prometheus.CounterValue,
			instance.TreeConnectCount,
		)

	}
	return nil
}

// toLabelName converts strings to lowercase and replaces all whitespaces and dots with underscores
func (c *collector) toLabelName(name string) string {
	s := strings.ReplaceAll(strings.Join(strings.Fields(strings.ToLower(name)), "_"), ".", "_")
	s = strings.ReplaceAll(s, "__", "_")
	return s
}
