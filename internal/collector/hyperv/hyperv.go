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

package hyperv

import (
	"errors"
	"fmt"
	"log/slog"
	"maps"
	"slices"
	"strings"
	"time"

	"github.com/alecthomas/kingpin/v2"
	"github.com/prometheus-community/windows_exporter/internal/mi"
	"github.com/prometheus-community/windows_exporter/internal/osversion"
	"github.com/prometheus-community/windows_exporter/internal/utils/recovery"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	Name = "hyperv"

	subCollectorDataStore                        = "datastore"
	subCollectorDynamicMemoryBalancer            = "dynamic_memory_balancer"
	subCollectorDynamicMemoryVM                  = "dynamic_memory_vm"
	subCollectorHost                             = "host"
	subCollectorHypervisorLogicalProcessor       = "hypervisor_logical_processor"
	subCollectorHypervisorRootPartition          = "hypervisor_root_partition"
	subCollectorHypervisorRootVirtualProcessor   = "hypervisor_root_virtual_processor"
	subCollectorHypervisorVirtualProcessor       = "hypervisor_virtual_processor"
	subCollectorLegacyNetworkAdapter             = "legacy_network_adapter"
	subCollectorReplicaVM                        = "replica_vm"
	subCollectorVirtualMachineHealthSummary      = "virtual_machine_health_summary"
	subCollectorVirtualMachineVidPartition       = "virtual_machine_vid_partition"
	subCollectorVirtualNetworkAdapter            = "virtual_network_adapter"
	subCollectorVirtualNetworkAdapterDropReasons = "virtual_network_adapter_drop_reasons"
	subCollectorVirtualSMB                       = "virtual_smb"
	subCollectorVirtualStorageDevice             = "virtual_storage_device"
	subCollectorVirtualSwitch                    = "virtual_switch"
	subCollectorWMIHealth                        = "wmi_health"
)

type Config struct {
	CollectorsEnabled []string `yaml:"enabled"`
}

//nolint:gochecknoglobals
var ConfigDefaults = Config{
	CollectorsEnabled: []string{
		subCollectorDataStore,
		subCollectorDynamicMemoryBalancer,
		subCollectorDynamicMemoryVM,
		subCollectorHost,
		subCollectorHypervisorLogicalProcessor,
		subCollectorHypervisorRootPartition,
		subCollectorHypervisorRootVirtualProcessor,
		subCollectorHypervisorVirtualProcessor,
		subCollectorLegacyNetworkAdapter,
		subCollectorReplicaVM,
		subCollectorVirtualMachineHealthSummary,
		subCollectorVirtualMachineVidPartition,
		subCollectorVirtualNetworkAdapter,
		subCollectorVirtualNetworkAdapterDropReasons,
		subCollectorVirtualSMB,
		subCollectorVirtualStorageDevice,
		subCollectorVirtualSwitch,
		subCollectorWMIHealth,
	},
}

// Collector is a Prometheus Collector for hyper-v.
type Collector struct {
	collectorDataStore
	collectorDynamicMemoryBalancer
	collectorDynamicMemoryVM
	collectorHost
	collectorHypervisorLogicalProcessor
	collectorHypervisorRootPartition
	collectorHypervisorRootVirtualProcessor
	collectorHypervisorVirtualProcessor
	collectorLegacyNetworkAdapter
	collectorReplicaVM
	collectorVirtualMachineHealthSummary
	collectorVirtualMachineVidPartition
	collectorVirtualNetworkAdapter
	collectorVirtualNetworkAdapterDropReasons
	collectorVirtualSMB
	collectorVirtualStorageDevice
	collectorVirtualSwitch
	collectorWMIHealth

	config Config
	logger *slog.Logger

	miSession *mi.Session

	collectorFns []func(ch chan<- prometheus.Metric) error
	closeFns     []func()
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
		"collector.hyperv.enabled",
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
	for _, fn := range c.closeFns {
		fn()
	}

	c.closeFns = nil

	return nil
}

func (c *Collector) Build(logger *slog.Logger, miSession *mi.Session) error {
	c.miSession = miSession
	c.logger = logger.With(slog.String("collector", Name))
	c.collectorFns = make([]func(ch chan<- prometheus.Metric) error, 0, len(c.config.CollectorsEnabled))
	c.closeFns = make([]func(), 0, len(c.config.CollectorsEnabled))

	if len(c.config.CollectorsEnabled) == 0 {
		return nil
	}

	subCollectors := map[string]struct {
		build          func() error
		collect        func(ch chan<- prometheus.Metric) error
		close          func()
		minBuildNumber uint16
	}{
		subCollectorDataStore: {
			build:          c.buildDataStore,
			collect:        c.collectDataStore,
			close:          func() { c.perfDataCollectorDataStore.Close() },
			minBuildNumber: osversion.LTSC2022,
		},
		subCollectorDynamicMemoryBalancer: {
			build:   c.buildDynamicMemoryBalancer,
			collect: c.collectDynamicMemoryBalancer,
			close:   func() { c.perfDataCollectorDynamicMemoryBalancer.Close() },
		},
		subCollectorDynamicMemoryVM: {
			build:   c.buildDynamicMemoryVM,
			collect: c.collectDynamicMemoryVM,
			close:   func() { c.perfDataCollectorDynamicMemoryVM.Close() },
		},
		subCollectorHost: {
			build:   c.buildHost,
			collect: c.collectHost,
			close:   func() { c.perfDataCollectorLogicalProcessor.Close() },
		},
		subCollectorHypervisorLogicalProcessor: {
			build:   c.buildHypervisorLogicalProcessor,
			collect: c.collectHypervisorLogicalProcessor,
			close:   func() { c.perfDataCollectorHypervisorLogicalProcessor.Close() },
		},
		subCollectorHypervisorRootPartition: {
			build:   c.buildHypervisorRootPartition,
			collect: c.collectHypervisorRootPartition,
			close:   func() { c.perfDataCollectorHypervisorRootPartition.Close() },
		},
		subCollectorHypervisorRootVirtualProcessor: {
			build:   c.buildHypervisorRootVirtualProcessor,
			collect: c.collectHypervisorRootVirtualProcessor,
			close:   func() { c.perfDataCollectorHypervisorRootVirtualProcessor.Close() },
		},
		subCollectorHypervisorVirtualProcessor: {
			build:   c.buildHypervisorVirtualProcessor,
			collect: c.collectHypervisorVirtualProcessor,
			close:   func() { c.perfDataCollectorHypervisorVirtualProcessor.Close() },
		},
		subCollectorLegacyNetworkAdapter: {
			build:   c.buildLegacyNetworkAdapter,
			collect: c.collectLegacyNetworkAdapter,
			close:   func() { c.perfDataCollectorLegacyNetworkAdapter.Close() },
		},
		subCollectorReplicaVM: {
			build:   c.buildReplicaVM,
			collect: c.collectReplicaVM,
			// Close the collector created by build, not the nil one bound at this point.
			close: func() { c.perfDataCollectorReplicaVM.Close() },
		},
		subCollectorVirtualMachineHealthSummary: {
			build:   c.buildVirtualMachineHealthSummary,
			collect: c.collectVirtualMachineHealthSummary,
			close:   func() { c.perfDataCollectorVirtualMachineHealthSummary.Close() },
		},
		subCollectorVirtualMachineVidPartition: {
			build:   c.buildVirtualMachineVidPartition,
			collect: c.collectVirtualMachineVidPartition,
			close:   func() { c.perfDataCollectorVirtualMachineVidPartition.Close() },
		},
		subCollectorVirtualNetworkAdapter: {
			build:   c.buildVirtualNetworkAdapter,
			collect: c.collectVirtualNetworkAdapter,
			close:   func() { c.perfDataCollectorVirtualNetworkAdapter.Close() },
		},
		subCollectorVirtualNetworkAdapterDropReasons: {
			build:   c.buildVirtualNetworkAdapterDropReasons,
			collect: c.collectVirtualNetworkAdapterDropReasons,
			close:   func() { c.perfDataCollectorVirtualNetworkAdapterDropReasons.Close() },
		},
		subCollectorVirtualSMB: {
			build:          c.buildVirtualSMB,
			collect:        c.collectVirtualSMB,
			close:          func() { c.perfDataCollectorVirtualSMB.Close() },
			minBuildNumber: osversion.LTSC2022,
		},
		subCollectorVirtualStorageDevice: {
			build:   c.buildVirtualStorageDevice,
			collect: c.collectVirtualStorageDevice,
			close:   func() { c.perfDataCollectorVirtualStorageDevice.Close() },
		},
		subCollectorVirtualSwitch: {
			build:   c.buildVirtualSwitch,
			collect: c.collectVirtualSwitch,
			close:   func() { c.perfDataCollectorVirtualSwitch.Close() },
		},
		subCollectorWMIHealth: {
			build:   c.buildWMIHealth,
			collect: c.collectWMIHealth,
			close:   func() {},
		},
	}

	buildNumber := osversion.Build()

	// Sort a copy, to not modify the slice of the caller or ConfigDefaults.
	// Result must order, to prevent test failures.
	collectorsEnabled := slices.Compact(slices.Sorted(slices.Values(c.config.CollectorsEnabled)))

	for _, name := range collectorsEnabled {
		if _, ok := subCollectors[name]; !ok {
			return fmt.Errorf("unknown sub collector: %s. Possible values: %s", name,
				strings.Join(slices.Sorted(maps.Keys(subCollectors)), ", "),
			)
		}
	}

	errs := make([]error, 0, len(collectorsEnabled))

	for _, name := range collectorsEnabled {
		if buildNumber < subCollectors[name].minBuildNumber {
			c.logger.Warn(fmt.Sprintf(
				"collector %s requires windows build version %d. Current build version: %d",
				name, subCollectors[name].minBuildNumber, buildNumber,
			))

			continue
		}

		// Register close before build, so that Close also releases a partially built sub collector.
		c.closeFns = append(c.closeFns, subCollectors[name].close)

		if err := subCollectors[name].build(); err != nil {
			errs = append(errs, fmt.Errorf("failed to build %s collector: %w", name, err))

			continue
		}

		c.collectorFns = append(c.collectorFns, subCollectors[name].collect)
	}

	return errors.Join(errs...)
}

// Collect sends the metric values for each metric
// to the provided prometheus Metric channel.
func (c *Collector) Collect(ch chan<- prometheus.Metric, _ time.Duration) error {
	var g recovery.Group

	for _, fn := range c.collectorFns {
		g.Go(func() error {
			return fn(ch)
		})
	}

	return g.Wait()
}
