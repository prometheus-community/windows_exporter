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
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"slices"
	"strings"
	"sync"
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
)

// nativeSubCollectors read through ClusAPI and do not need an MI session.
//
//nolint:gochecknoglobals
var nativeSubCollectors = []string{subCollectorCluster, subCollectorNetwork, subCollectorNode, subCollectorResource, subCollectorResourceGroup, subCollectorSharedVolumes}

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

	lifecycleMu sync.Mutex
	config      Config
	miSession   *mi.Session
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
	c.lifecycleMu.Lock()
	defer c.lifecycleMu.Unlock()

	return c.closeSources()
}

// closeSources releases the ClusAPI sources. A source whose Close fails is
// dropped as well: the clusapi.Cluster has already released its handle, and a
// later Build must be able to create new sources (Grafana Alloy rebuilds
// collectors in place).
func (c *Collector) closeSources() error {
	var errs []error

	closeSource := func(source io.Closer, release func()) {
		if source == nil {
			return
		}

		if err := source.Close(); err != nil {
			errs = append(errs, err)
		}

		release()
	}

	closeSource(c.clusterSource, func() { c.clusterSource = nil })
	closeSource(c.networkSource, func() { c.networkSource = nil })
	closeSource(c.nodeSource, func() { c.nodeSource = nil })
	closeSource(c.resourceSource, func() { c.resourceSource = nil })
	closeSource(c.resourceGroupSource, func() { c.resourceGroupSource = nil })
	closeSource(c.sharedVolumesSource, func() { c.sharedVolumesSource = nil })

	return errors.Join(errs...)
}

// errNotBuilt reports a sub-collector without a source, for example when a
// library consumer calls Collect after a failed Build.
func errNotBuilt(name string) error {
	return fmt.Errorf("%s collector is not built", name)
}

// callSource runs a ClusAPI read in its own goroutine and waits at most for
// the scrape budget. ClusAPI RPCs cannot be cancelled: a read that outlives
// the budget keeps running, and its source refuses new reads with
// clusapi.ErrBusy until the RPC returns. Returning here keeps one stuck
// sub-collector from holding Collect, and with it every other sub-collector.
// Late results are dropped; the read never writes to the metric channel.
//
//nolint:ireturn // T is the concrete result type of each source.
func callSource[T any](budget time.Duration, read func(deadline time.Time) (T, error)) (T, error) {
	var deadline time.Time
	if budget > 0 {
		deadline = time.Now().Add(budget)
	}

	type result struct {
		value T
		err   error
	}

	done := make(chan result, 1)

	go func() {
		defer func() {
			if r := recover(); r != nil {
				done <- result{err: fmt.Errorf("panic in ClusAPI read: %v", r)}
			}
		}()

		value, err := read(deadline)
		done <- result{value: value, err: err}
	}()

	if budget <= 0 {
		r := <-done

		return r.value, r.err
	}

	timer := time.NewTimer(budget)
	defer timer.Stop()

	select {
	case r := <-done:
		return r.value, r.err
	case <-timer.C:
		var zero T

		return zero, fmt.Errorf("ClusAPI read did not return within %s: %w", budget, context.DeadlineExceeded)
	}
}

// remainingBudget returns the part of the scrape budget that is left for a
// sub-collector started after another one. Zero means no limit.
func remainingBudget(maxScrapeDuration time.Duration, scrapeStarted time.Time) (time.Duration, error) {
	if maxScrapeDuration <= 0 {
		return maxScrapeDuration, nil
	}

	budget := maxScrapeDuration - time.Since(scrapeStarted)
	if budget <= 0 {
		return 0, context.DeadlineExceeded
	}

	return budget, nil
}

func (c *Collector) Build(_ *slog.Logger, miSession *mi.Session) error {
	c.lifecycleMu.Lock()
	defer c.lifecycleMu.Unlock()

	// A failed close is reported, but the old sources are dropped and rebuilt.
	closeErr := c.closeSources()

	if len(c.config.CollectorsEnabled) == 0 {
		return closeErr
	}

	subCollectors := []string{
		subCollectorCluster,
		subCollectorNetwork,
		subCollectorNode,
		subCollectorResource,
		subCollectorResourceGroup,
		subCollectorSharedVolumes,
	}

	for _, name := range c.config.CollectorsEnabled {
		if !slices.Contains(subCollectors, name) {
			return errors.Join(closeErr, fmt.Errorf("unknown sub collector: %s. Possible values: %s", name,
				strings.Join(subCollectors, ", "),
			))
		}
	}

	if miSession == nil && slices.ContainsFunc(c.config.CollectorsEnabled, func(name string) bool { return !slices.Contains(nativeSubCollectors, name) }) {
		return errors.Join(closeErr, errors.New("miSession is nil"))
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

	if len(errs) != 0 {
		errs = append(errs, c.closeSources())
	}

	return errors.Join(append(errs, closeErr)...)
}

// Collect sends the metric values for each metric
// to the provided prometheus Metric channel.
func (c *Collector) Collect(ch chan<- prometheus.Metric, maxScrapeDuration time.Duration) error {
	c.lifecycleMu.Lock()
	defer c.lifecycleMu.Unlock()

	scrapeStarted := time.Now()

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
				budget, err := remainingBudget(maxScrapeDuration, scrapeStarted)
				if err != nil {
					return fmt.Errorf("failed to collect resource metrics: %w", err)
				}

				if err := c.collectResource(ch, budget, nodeNames); err != nil {
					return fmt.Errorf("failed to collect resource metrics: %w", err)
				}

				return nil
			})
		}

		if slices.Contains(c.config.CollectorsEnabled, subCollectorResourceGroup) {
			g.Go(func() error {
				budget, err := remainingBudget(maxScrapeDuration, scrapeStarted)
				if err != nil {
					return fmt.Errorf("failed to collect resource group metrics: %w", err)
				}

				if err := c.collectResourceGroup(ch, budget, nodeNames); err != nil {
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

	return g.Wait()
}
