//go:build windows

package smb

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
	Name                     = "smb"
	FlagSmbListAllCollectors = "collectors.smb.list"
	FlagSmbCollectorsEnabled = "collectors.smb.enabled"
)

type Config struct {
	CollectorsEnabled string `yaml:"collectors_enabled"`
}

var ConfigDefaults = Config{
	CollectorsEnabled: "",
}

type collector struct {
	logger log.Logger

	smbListAllCollectors *bool
	smbCollectorsEnabled *string

	TreeConnectCount     *prometheus.Desc
	CurrentOpenFileCount *prometheus.Desc

	enabledCollectors []string
}

// All available collector functions
var smbAllCollectorNames = []string{
	"ServerShares",
	"ServerSessions",
}

func New(logger log.Logger, config *Config) types.Collector {
	if config == nil {
		config = &ConfigDefaults
	}

	smbListAllCollectors := false
	c := &collector{
		smbCollectorsEnabled: &config.CollectorsEnabled,
		smbListAllCollectors: &smbListAllCollectors,
	}
	c.SetLogger(logger)
	return c
}

func NewWithFlags(app *kingpin.Application) types.Collector {
	return &collector{
		smbListAllCollectors: app.Flag(
			FlagSmbListAllCollectors,
			"List the collectors along with their perflib object name/ids",
		).Bool(),

		smbCollectorsEnabled: app.Flag(
			FlagSmbCollectorsEnabled,
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
		"SMB Server Sessions",
		"SMB Server Shares",
	}, nil
}

func (c *collector) Build() error {
	// desc creates a new prometheus description
	desc := func(metricName string, description string, labels ...string) *prometheus.Desc {
		return prometheus.NewDesc(
			prometheus.BuildFQName(types.Namespace, "smb", metricName),
			description,
			labels,
			nil,
		)
	}

	c.CurrentOpenFileCount = desc("server_shares_current_open_file_count", "Current total count open files on the SMB Server")
	c.TreeConnectCount = desc("server_session_tree_connect_count", "Tree connect count to SMB Server")

	c.enabledCollectors = make([]string, 0, len(smbAllCollectorNames))

	collectorDesc := map[string]string{
		"ServerShares":   "SMB Server Shares",
		"ServerSessions": "SMB Server Sessions",
	}

	if *c.smbListAllCollectors {
		fmt.Printf("%-32s %-32s\n", "Collector Name", "Perflib Object")
		for _, cname := range smbAllCollectorNames {
			fmt.Printf("%-32s %-32s\n", cname, collectorDesc[cname])
		}
		os.Exit(0)
	}

	if *c.smbCollectorsEnabled == "" {
		for _, collectorName := range smbAllCollectorNames {
			c.enabledCollectors = append(c.enabledCollectors, collectorName)
		}
	} else {
		for _, collectorName := range strings.Split(*c.smbCollectorsEnabled, ",") {
			if slices.Contains(smbAllCollectorNames, collectorName) {
				c.enabledCollectors = append(c.enabledCollectors, collectorName)
			} else {
				return fmt.Errorf("unknown smb collector: %s", collectorName)
			}
		}
	}

	return nil
}

// Collect collects smb metrics and sends them to prometheus
func (c *collector) Collect(ctx *types.ScrapeContext, ch chan<- prometheus.Metric) error {
	collectorFuncs := map[string]func(ctx *types.ScrapeContext, ch chan<- prometheus.Metric) error{
		"ServerShares":   c.collectServerShares,
		"ServerSessions": c.collectServerSessions,
	}

	for _, collectorName := range c.enabledCollectors {
		if err := collectorFuncs[collectorName](ctx, ch); err != nil {
			_ = level.Error(c.logger).Log("msg", "Error in "+collectorName, "err", err)
			return err
		}
	}
	return nil
}

// Perflib: SMB Server Shares
type perflibServerShares struct {
	Name string

	CurrentOpenFileCount float64 `perflib:"Current Open File Count"`
}

func (c *collector) collectServerShares(ctx *types.ScrapeContext, ch chan<- prometheus.Metric) error {
	var data []perflibServerShares
	if err := perflib.UnmarshalObject(ctx.PerfObjects["SMB Server Shares"], &data, c.logger); err != nil {
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

	}
	return nil
}

// Perflib: SMB Server Sessions
type perflibServerSession struct {
	TreeConnectCount float64 `perflib:"Tree Connect Count"`
}

func (c *collector) collectServerSessions(ctx *types.ScrapeContext, ch chan<- prometheus.Metric) error {
	var data []perflibServerSession
	if err := perflib.UnmarshalObject(ctx.PerfObjects["SMB Server Sessions"], &data, c.logger); err != nil {
		return err
	}

	for _, instance := range data {

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
