//go:build windows

package mscluster

import (
	"errors"
	"fmt"
	"log/slog"
	"slices"
	"strings"

	"github.com/alecthomas/kingpin/v2"
	"github.com/prometheus-community/windows_exporter/internal/mi"
	"github.com/prometheus/client_golang/prometheus"
)

const Name = "mscluster"

type Config struct {
	CollectorsEnabled []string `yaml:"collectors_enabled"`
}

var ConfigDefaults = Config{
	CollectorsEnabled: []string{
		"cluster",
		"network",
		"node",
		"resource",
		"resourcegroup",
	},
}

// A Collector is a Prometheus Collector for WMI MSCluster_Cluster metrics.
type Collector struct {
	config    Config
	miSession *mi.Session

	collectorCluster
	collectorNetwork
	collectorNode
	collectorResource
	collectorResourceGroup
}

func New(config *Config) *Collector {
	if config == nil {
		config = &ConfigDefaults
	}

	if config.CollectorsEnabled == nil {
		config.CollectorsEnabled = ConfigDefaults.CollectorsEnabled
	}

	c := &Collector{
		config: *config,
	}

	return c
}

func NewWithFlags(app *kingpin.Application) *Collector {
	c := &Collector{
		config: ConfigDefaults,
	}
	c.config.CollectorsEnabled = make([]string, 0)

	var collectorsEnabled string

	app.Flag(
		"collector.mscluster.enabled",
		"Comma-separated list of collectors to use.",
	).Default(strings.Join(ConfigDefaults.CollectorsEnabled, ",")).StringVar(&collectorsEnabled)

	app.Action(func(*kingpin.ParseContext) error {
		c.config.CollectorsEnabled = strings.Split(collectorsEnabled, ",")

		return nil
	})

	return c
}

func (c *Collector) GetName() string {
	return Name
}

func (c *Collector) GetPerfCounter(_ *slog.Logger) ([]string, error) {
	return []string{"Memory"}, nil
}

func (c *Collector) Close() error {
	return nil
}

func (c *Collector) Build(_ *slog.Logger, miSession *mi.Session) error {
	if len(c.config.CollectorsEnabled) == 0 {
		return nil
	}

	if miSession == nil {
		return errors.New("miSession is nil")
	}

	c.miSession = miSession

	if slices.Contains(c.config.CollectorsEnabled, "cluster") {
		c.buildCluster()
	}

	if slices.Contains(c.config.CollectorsEnabled, "network") {
		c.buildNetwork()
	}

	if slices.Contains(c.config.CollectorsEnabled, "node") {
		c.buildNode()
	}

	if slices.Contains(c.config.CollectorsEnabled, "resource") {
		c.buildResource()
	}

	if slices.Contains(c.config.CollectorsEnabled, "resourcegroup") {
		c.buildResourceGroup()
	}

	return nil
}

// Collect sends the metric values for each metric
// to the provided prometheus Metric channel.
func (c *Collector) Collect(ch chan<- prometheus.Metric) error {
	if len(c.config.CollectorsEnabled) == 0 {
		return nil
	}

	var (
		err       error
		errs      []error
		nodeNames []string
	)

	if slices.Contains(c.config.CollectorsEnabled, "cluster") {
		if err = c.collectCluster(ch); err != nil {
			errs = append(errs, fmt.Errorf("failed to collect cluster metrics: %w", err))
		}
	}

	if slices.Contains(c.config.CollectorsEnabled, "network") {
		if err = c.collectNetwork(ch); err != nil {
			errs = append(errs, fmt.Errorf("failed to collect network metrics: %w", err))
		}
	}

	if slices.Contains(c.config.CollectorsEnabled, "node") {
		if nodeNames, err = c.collectNode(ch); err != nil {
			errs = append(errs, fmt.Errorf("failed to collect node metrics: %w", err))
		}
	}

	if slices.Contains(c.config.CollectorsEnabled, "resource") {
		if err = c.collectResource(ch, nodeNames); err != nil {
			errs = append(errs, fmt.Errorf("failed to collect resource metrics: %w", err))
		}
	}

	if slices.Contains(c.config.CollectorsEnabled, "resourcegroup") {
		if err = c.collectResourceGroup(ch, nodeNames); err != nil {
			errs = append(errs, fmt.Errorf("failed to collect resource group metrics: %w", err))
		}
	}

	return errors.Join(errs...)
}
