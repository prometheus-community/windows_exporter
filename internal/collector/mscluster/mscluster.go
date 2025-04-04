// Copyright 2024 The Prometheus Authors
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

//go:build windows

package mscluster

import (
	"errors"
	"fmt"
	"log/slog"
	"slices"
	"strings"
	"sync"

	"github.com/alecthomas/kingpin/v2"
	"github.com/prometheus-community/windows_exporter/internal/mi"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	Name = "mscluster"

	subCollectorCluster       = "cluster"
	subCollectorNetwork       = "network"
	subCollectorNode          = "node"
	subCollectorResource      = "resource"
	subCollectorResourceGroup = "resourcegroup"
)

type Config struct {
	CollectorsEnabled []string `yaml:"collectors_enabled"`
}

//nolint:gochecknoglobals
var ConfigDefaults = Config{
	CollectorsEnabled: []string{
		subCollectorCluster,
		subCollectorNetwork,
		subCollectorNode,
		subCollectorResource,
		subCollectorResourceGroup,
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

	errs := make([]error, 0)

	if slices.Contains(c.config.CollectorsEnabled, subCollectorCluster) {
		if err := c.buildCluster(); err != nil {
			errs = append(errs, fmt.Errorf("failed to build cluster collector: %w", err))
		}
	}

	if slices.Contains(c.config.CollectorsEnabled, subCollectorNetwork) {
		if err := c.buildNetwork(); err != nil {
			errs = append(errs, fmt.Errorf("failed to build network collector: %w", err))
		}
	}

	if slices.Contains(c.config.CollectorsEnabled, subCollectorNode) {
		if err := c.buildNode(); err != nil {
			errs = append(errs, fmt.Errorf("failed to build node collector: %w", err))
		}
	}

	if slices.Contains(c.config.CollectorsEnabled, subCollectorResource) {
		if err := c.buildResource(); err != nil {
			errs = append(errs, fmt.Errorf("failed to build resource collector: %w", err))
		}
	}

	if slices.Contains(c.config.CollectorsEnabled, subCollectorResourceGroup) {
		if err := c.buildResourceGroup(); err != nil {
			errs = append(errs, fmt.Errorf("failed to build resource group collector: %w", err))
		}
	}

	return errors.Join(errs...)
}

// Collect sends the metric values for each metric
// to the provided prometheus Metric channel.
func (c *Collector) Collect(ch chan<- prometheus.Metric) error {
	if len(c.config.CollectorsEnabled) == 0 {
		return nil
	}

	errCh := make(chan error, 5)

	wg := sync.WaitGroup{}
	wg.Add(5)

	go func() {
		defer wg.Done()

		if slices.Contains(c.config.CollectorsEnabled, subCollectorCluster) {
			if err := c.collectCluster(ch); err != nil {
				errCh <- fmt.Errorf("failed to collect cluster metrics: %w", err)
			}
		}
	}()

	go func() {
		defer wg.Done()

		if slices.Contains(c.config.CollectorsEnabled, subCollectorNetwork) {
			if err := c.collectNetwork(ch); err != nil {
				errCh <- fmt.Errorf("failed to collect network metrics: %w", err)
			}
		}
	}()

	go func() {
		defer wg.Done()

		nodeNames := make([]string, 0)

		if slices.Contains(c.config.CollectorsEnabled, subCollectorNode) {
			var err error

			nodeNames, err = c.collectNode(ch)
			if err != nil {
				errCh <- fmt.Errorf("failed to collect node metrics: %w", err)
			}
		}

		go func() {
			defer wg.Done()

			if slices.Contains(c.config.CollectorsEnabled, subCollectorResource) {
				if err := c.collectResource(ch, nodeNames); err != nil {
					errCh <- fmt.Errorf("failed to collect resource metrics: %w", err)
				}
			}
		}()

		go func() {
			defer wg.Done()

			if slices.Contains(c.config.CollectorsEnabled, subCollectorResourceGroup) {
				if err := c.collectResourceGroup(ch, nodeNames); err != nil {
					errCh <- fmt.Errorf("failed to collect resource group metrics: %w", err)
				}
			}
		}()
	}()

	wg.Wait()
	close(errCh)

	errs := make([]error, 0)

	for err := range errCh {
		errs = append(errs, err)
	}

	return errors.Join(errs...)
}
