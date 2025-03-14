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

package hyperv

import (
	"errors"
	"fmt"
	"log/slog"
	"sort"
	"strings"
	"sync"

	"github.com/Microsoft/hcsshim/osversion"
	"github.com/alecthomas/kingpin/v2"
	"github.com/prometheus-community/windows_exporter/internal/mi"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	Name = "hyperv"

	subCollectorDataStore                        = "datastore"
	subCollectorDynamicMemoryBalancer            = "dynamic_memory_balancer"
	subCollectorDynamicMemoryVM                  = "dynamic_memory_vm"
	subCollectorHypervisorLogicalProcessor       = "hypervisor_logical_processor"
	subCollectorHypervisorRootPartition          = "hypervisor_root_partition"
	subCollectorHypervisorRootVirtualProcessor   = "hypervisor_root_virtual_processor"
	subCollectorHypervisorVirtualProcessor       = "hypervisor_virtual_processor"
	subCollectorLegacyNetworkAdapter             = "legacy_network_adapter"
	subCollectorVirtualMachineHealthSummary      = "virtual_machine_health_summary"
	subCollectorVirtualMachineVidPartition       = "virtual_machine_vid_partition"
	subCollectorVirtualNetworkAdapter            = "virtual_network_adapter"
	subCollectorVirtualNetworkAdapterDropReasons = "virtual_network_adapter_drop_reasons"
	subCollectorVirtualSMB                       = "virtual_smb"
	subCollectorVirtualStorageDevice             = "virtual_storage_device"
	subCollectorVirtualSwitch                    = "virtual_switch"
)

type Config struct {
	CollectorsEnabled []string `yaml:"collectors_enabled"`
}

//nolint:gochecknoglobals
var ConfigDefaults = Config{
	CollectorsEnabled: []string{
		subCollectorDataStore,
		subCollectorDynamicMemoryBalancer,
		subCollectorDynamicMemoryVM,
		subCollectorHypervisorLogicalProcessor,
		subCollectorHypervisorRootPartition,
		subCollectorHypervisorRootVirtualProcessor,
		subCollectorHypervisorVirtualProcessor,
		subCollectorLegacyNetworkAdapter,
		subCollectorVirtualMachineHealthSummary,
		subCollectorVirtualMachineVidPartition,
		subCollectorVirtualNetworkAdapter,
		subCollectorVirtualNetworkAdapterDropReasons,
		subCollectorVirtualSMB,
		subCollectorVirtualStorageDevice,
		subCollectorVirtualSwitch,
	},
}

// Collector is a Prometheus Collector for hyper-v.
type Collector struct {
	config Config

	collectorFns []func(ch chan<- prometheus.Metric) error
	closeFns     []func()

	collectorDataStore
	collectorDynamicMemoryBalancer
	collectorDynamicMemoryVM
	collectorHypervisorLogicalProcessor
	collectorHypervisorRootPartition
	collectorHypervisorRootVirtualProcessor
	collectorHypervisorVirtualProcessor
	collectorLegacyNetworkAdapter
	collectorVirtualMachineHealthSummary
	collectorVirtualMachineVidPartition
	collectorVirtualNetworkAdapter
	collectorVirtualNetworkAdapterDropReasons
	collectorVirtualSMB
	collectorVirtualStorageDevice
	collectorVirtualSwitch
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

	return nil
}

func (c *Collector) Build(logger *slog.Logger, _ *mi.Session) error {
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
			close:          c.perfDataCollectorDataStore.Close,
			minBuildNumber: osversion.LTSC2022,
		},
		subCollectorDynamicMemoryBalancer: {
			build:   c.buildDynamicMemoryBalancer,
			collect: c.collectDynamicMemoryBalancer,
			close:   c.perfDataCollectorDynamicMemoryBalancer.Close,
		},
		subCollectorDynamicMemoryVM: {
			build:   c.buildDynamicMemoryVM,
			collect: c.collectDynamicMemoryVM,
			close:   c.perfDataCollectorDynamicMemoryVM.Close,
		},
		subCollectorHypervisorLogicalProcessor: {
			build:   c.buildHypervisorLogicalProcessor,
			collect: c.collectHypervisorLogicalProcessor,
			close:   c.perfDataCollectorHypervisorLogicalProcessor.Close,
		},
		subCollectorHypervisorRootPartition: {
			build:   c.buildHypervisorRootPartition,
			collect: c.collectHypervisorRootPartition,
			close:   c.perfDataCollectorHypervisorRootPartition.Close,
		},
		subCollectorHypervisorRootVirtualProcessor: {
			build:   c.buildHypervisorRootVirtualProcessor,
			collect: c.collectHypervisorRootVirtualProcessor,
			close:   c.perfDataCollectorHypervisorRootVirtualProcessor.Close,
		},
		subCollectorHypervisorVirtualProcessor: {
			build:   c.buildHypervisorVirtualProcessor,
			collect: c.collectHypervisorVirtualProcessor,
			close:   c.perfDataCollectorHypervisorVirtualProcessor.Close,
		},
		subCollectorLegacyNetworkAdapter: {
			build:   c.buildLegacyNetworkAdapter,
			collect: c.collectLegacyNetworkAdapter,
			close:   c.perfDataCollectorLegacyNetworkAdapter.Close,
		},
		subCollectorVirtualMachineHealthSummary: {
			build:   c.buildVirtualMachineHealthSummary,
			collect: c.collectVirtualMachineHealthSummary,
			close:   c.perfDataCollectorVirtualMachineHealthSummary.Close,
		},
		subCollectorVirtualMachineVidPartition: {
			build:   c.buildVirtualMachineVidPartition,
			collect: c.collectVirtualMachineVidPartition,
			close:   c.perfDataCollectorVirtualMachineVidPartition.Close,
		},
		subCollectorVirtualNetworkAdapter: {
			build:   c.buildVirtualNetworkAdapter,
			collect: c.collectVirtualNetworkAdapter,
			close:   c.perfDataCollectorVirtualNetworkAdapter.Close,
		},
		subCollectorVirtualNetworkAdapterDropReasons: {
			build:   c.buildVirtualNetworkAdapterDropReasons,
			collect: c.collectVirtualNetworkAdapterDropReasons,
			close:   c.perfDataCollectorVirtualNetworkAdapterDropReasons.Close,
		},
		subCollectorVirtualSMB: {
			build:          c.buildVirtualSMB,
			collect:        c.collectVirtualSMB,
			close:          c.perfDataCollectorVirtualSMB.Close,
			minBuildNumber: osversion.LTSC2022,
		},
		subCollectorVirtualStorageDevice: {
			build:   c.buildVirtualStorageDevice,
			collect: c.collectVirtualStorageDevice,
			close:   c.perfDataCollectorVirtualStorageDevice.Close,
		},
		subCollectorVirtualSwitch: {
			build:   c.buildVirtualSwitch,
			collect: c.collectVirtualSwitch,
			close:   c.perfDataCollectorVirtualSwitch.Close,
		},
	}

	buildNumber := osversion.Build()

	// Result must order, to prevent test failures.
	sort.Strings(c.config.CollectorsEnabled)

	errs := make([]error, 0, len(c.config.CollectorsEnabled))

	for _, name := range c.config.CollectorsEnabled {
		if _, ok := subCollectors[name]; !ok {
			return fmt.Errorf("unknown collector: %s", name)
		}

		if buildNumber < subCollectors[name].minBuildNumber {
			logger.Warn(fmt.Sprintf(
				"collector %s requires windows build version %d. Current build version: %d",
				name, subCollectors[name].minBuildNumber, buildNumber,
			), slog.String("collector", name))

			continue
		}

		if err := subCollectors[name].build(); err != nil {
			errs = append(errs, fmt.Errorf("failed to build %s collector: %w", name, err))

			continue
		}

		c.collectorFns = append(c.collectorFns, subCollectors[name].collect)
		c.closeFns = append(c.closeFns, subCollectors[name].close)
	}

	return errors.Join(errs...)
}

// Collect sends the metric values for each metric
// to the provided prometheus Metric channel.
func (c *Collector) Collect(ch chan<- prometheus.Metric) error {
	errCh := make(chan error, len(c.collectorFns))
	errs := make([]error, 0, len(c.collectorFns))

	wg := sync.WaitGroup{}

	for _, fn := range c.collectorFns {
		wg.Add(1)

		go func(fn func(ch chan<- prometheus.Metric) error) {
			defer wg.Done()

			if err := fn(ch); err != nil {
				errCh <- err
			}
		}(fn)
	}

	wg.Wait()

	close(errCh)

	for err := range errCh {
		errs = append(errs, err)
	}

	return errors.Join(errs...)
}
