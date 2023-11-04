//go:build windows

package collector

import (
	"slices"
	"strings"

	"github.com/alecthomas/kingpin/v2"
	"github.com/go-kit/log"
	"github.com/prometheus-community/windows_exporter/pkg/perflib"
	"github.com/prometheus-community/windows_exporter/pkg/types"
)

type Collectors struct {
	logger log.Logger

	collectors       map[string]types.Collector
	perfCounterQuery string
}

// NewWithFlags To be called by the exporter for collector initialization before running kingpin.Parse
func NewWithFlags(app *kingpin.Application) Collectors {
	collectors := map[string]types.Collector{}

	for name, builder := range Map {
		collectors[name] = builder(app)
	}

	return New(collectors)
}

// New To be called by the external libraries for collector initialization without running kingpin.Parse
func New(collectors map[string]types.Collector) Collectors {
	return Collectors{
		collectors: collectors,
	}
}

func (c *Collectors) SetLogger(logger log.Logger) {
	c.logger = logger

	for _, collector := range c.collectors {
		collector.SetLogger(logger)
	}
}

func (c *Collectors) SetPerfCounterQuery() error {
	var (
		err error

		perfCounterDependencies []string
		perfCounterNames        []string
		perfIndicies            []string
	)

	for _, collector := range c.collectors {
		perfCounterNames, err = collector.GetPerfCounter()
		if err != nil {
			return err
		}

		perfIndicies = make([]string, 0, len(perfCounterNames))
		for _, cn := range perfCounterNames {
			perfIndicies = append(perfIndicies, perflib.MapCounterToIndex(cn))
		}

		perfCounterDependencies = append(perfCounterDependencies, strings.Join(perfIndicies, " "))
	}

	c.perfCounterQuery = strings.Join(perfCounterDependencies, " ")

	return nil
}

// Enable removes all collectors that not enabledCollectors
func (c *Collectors) Enable(enabledCollectors []string) {
	for name := range c.collectors {
		if !slices.Contains(enabledCollectors, name) {
			delete(c.collectors, name)
		}
	}
}

// Build To be called by the exporter for collector initialization
func (c *Collectors) Build() error {
	var err error
	for _, collector := range c.collectors {
		if err = collector.Build(); err != nil {
			return err
		}
	}

	return nil
}

// PrepareScrapeContext creates a ScrapeContext to be used during a single scrape
func (c *Collectors) PrepareScrapeContext() (*types.ScrapeContext, error) {
	objs, err := perflib.GetPerflibSnapshot(c.perfCounterQuery)
	if err != nil {
		return nil, err
	}

	return &types.ScrapeContext{PerfObjects: objs}, nil
}
