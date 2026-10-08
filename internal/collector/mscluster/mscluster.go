// SPDX-License-Identifier: Apache-2.0
//
// Copyright The Prometheus Authors
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
	"time"

	"github.com/alecthomas/kingpin/v2"
	"github.com/prometheus-community/windows_exporter/internal/mi"
	"github.com/prometheus-community/windows_exporter/internal/utils/recovery"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	Name = "mscluster"

	subCollectorCluster       = "cluster"
	subCollectorNetwork       = "network"
	subCollectorNode          = "node"
	subCollectorResource      = "resource"
	subCollectorResourceGroup = "resourcegroup"
	subCollectorSharedVolumes = "shared_volumes"
	subCollectorVirtualDisk   = "virtualdisk"
	subCollectorStoragePool   = "storagepool"
)

type Config struct {
	CollectorsEnabled []string `yaml:"enabled"`
}

//nolint:gochecknoglobals
var ConfigDefaults = Config{
	CollectorsEnabled: []string{
		subCollectorCluster,
		subCollectorNetwork,
		subCollectorNode,
		subCollectorResource,
		subCollectorResourceGroup,
		subCollectorSharedVolumes,
		subCollectorVirtualDisk,
		subCollectorStoragePool,
	},
}

// A Collector is a Prometheus Collector for WMI MSCluster_Cluster metrics.
type Collector struct {
	collectorCluster
	collectorNetwork
	collectorNode
	collectorResource
	collectorResourceGroup
	collectorSharedVolumes
	collectorVirtualDisk
	collectorStoragePool

	config    Config
	miSession *mi.Session
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

	subCollectors := []string{
		subCollectorCluster,
		subCollectorNetwork,
		subCollectorNode,
		subCollectorResource,
		subCollectorResourceGroup,
		subCollectorSharedVolumes,
		subCollectorVirtualDisk,
		subCollectorStoragePool,
	}

	for _, name := range c.config.CollectorsEnabled {
		if !slices.Contains(subCollectors, name) {
			return fmt.Errorf("unknown sub collector: %s. Possible values: %s", name,
				strings.Join(subCollectors, ", "),
			)
		}
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

	if slices.Contains(c.config.CollectorsEnabled, subCollectorSharedVolumes) {
		if err := c.buildSharedVolumes(); err != nil {
			errs = append(errs, fmt.Errorf("failed to build shared_volumes collector: %w", err))
		}
	}

	if slices.Contains(c.config.CollectorsEnabled, subCollectorVirtualDisk) {
		if err := c.buildVirtualDisk(); err != nil {
			errs = append(errs, fmt.Errorf("failed to build virtualdisk collector: %w", err))
		}
	}

	if slices.Contains(c.config.CollectorsEnabled, subCollectorStoragePool) {
		if err := c.buildStoragePool(); err != nil {
			errs = append(errs, fmt.Errorf("failed to build storagepool collector: %w", err))
		}
	}

	return errors.Join(errs...)
}

// Collect sends the metric values for each metric
// to the provided prometheus Metric channel.
func (c *Collector) Collect(ch chan<- prometheus.Metric, maxScrapeDuration time.Duration) error {
	if len(c.config.CollectorsEnabled) == 0 {
		return nil
	}

	var g recovery.Group

	if slices.Contains(c.config.CollectorsEnabled, subCollectorCluster) {
		g.Go(func() error {
			if err := c.collectCluster(ch, maxScrapeDuration); err != nil {
				return fmt.Errorf("failed to collect cluster metrics: %w", err)
			}

			return nil
		})
	}

	if slices.Contains(c.config.CollectorsEnabled, subCollectorNetwork) {
		g.Go(func() error {
			if err := c.collectNetwork(ch, maxScrapeDuration); err != nil {
				return fmt.Errorf("failed to collect network metrics: %w", err)
			}

			return nil
		})
	}

	// The resource and resource group collectors need the node names.
	g.Go(func() error {
		var (
			nodeNames []string
			err       error
		)

		if slices.Contains(c.config.CollectorsEnabled, subCollectorNode) {
			nodeNames, err = c.collectNode(ch, maxScrapeDuration)
			if err != nil {
				err = fmt.Errorf("failed to collect node metrics: %w", err)
			}
		}

		if slices.Contains(c.config.CollectorsEnabled, subCollectorResource) {
			g.Go(func() error {
				if err := c.collectResource(ch, maxScrapeDuration, nodeNames); err != nil {
					return fmt.Errorf("failed to collect resource metrics: %w", err)
				}

				return nil
			})
		}

		if slices.Contains(c.config.CollectorsEnabled, subCollectorResourceGroup) {
			g.Go(func() error {
				if err := c.collectResourceGroup(ch, maxScrapeDuration, nodeNames); err != nil {
					return fmt.Errorf("failed to collect resource group metrics: %w", err)
				}

				return nil
			})
		}

		return err
	})

	if slices.Contains(c.config.CollectorsEnabled, subCollectorSharedVolumes) {
		g.Go(func() error {
			if err := c.collectSharedVolumes(ch, maxScrapeDuration); err != nil {
				return fmt.Errorf("failed to collect shared_volumes metrics: %w", err)
			}

			return nil
		})
	}

	if slices.Contains(c.config.CollectorsEnabled, subCollectorVirtualDisk) {
		g.Go(func() error {
			if err := c.collectVirtualDisk(ch, maxScrapeDuration); err != nil {
				return fmt.Errorf("failed to collect virtualdisk metrics: %w", err)
			}

			return nil
		})
	}

	if slices.Contains(c.config.CollectorsEnabled, subCollectorStoragePool) {
		g.Go(func() error {
			if err := c.collectStoragePool(ch, maxScrapeDuration); err != nil {
				return fmt.Errorf("failed to collect storagepool metrics: %w", err)
			}

			return nil
		})
	}

	return g.Wait()
}
