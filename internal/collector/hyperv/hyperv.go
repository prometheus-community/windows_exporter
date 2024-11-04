//go:build windows

package hyperv

import (
	"errors"
	"log/slog"
	"slices"
	"strings"
	"sync"

	"github.com/alecthomas/kingpin/v2"
	"github.com/prometheus-community/windows_exporter/internal/mi"
	"github.com/prometheus-community/windows_exporter/internal/types"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	Name = "hyperv"

	SubCollectorDataStore                        = "datastore"
	SubCollectorDynamicMemoryBalancer            = "dynamic_memory_balancer"
	SubCollectorDynamicMemoryVM                  = "dynamic_memory_vm"
	SubCollectorHypervisorLogicalProcessor       = "hypervisor_logical_processor"
	SubCollectorHypervisorRootPartition          = "hypervisor_root_partition"
	SubCollectorHypervisorRootVirtualProcessor   = "hypervisor_root_virtual_processor"
	SubCollectorHypervisorVirtualProcessor       = "hypervisor_virtual_processor"
	SubCollectorLegacyNetworkAdapter             = "legacy_network_adapter"
	SubCollectorVirtualMachineHealthSummary      = "virtual_machine_health_summary"
	SubCollectorVirtualMachineVidPartition       = "virtual_machine_vid_partition"
	SubCollectorVirtualNetworkAdapter            = "virtual_network_adapter"
	SubCollectorVirtualNetworkAdapterDropReasons = "virtual_network_adapter_drop_reasons"
	SubCollectorVirtualSMB                       = "virtual_smb"
	SubCollectorVirtualStorageDevice             = "virtual_storage_device"
	SubCollectorVirtualSwitch                    = "virtual_switch"
)

type Config struct {
	CollectorsEnabled []string `yaml:"collectors_enabled"`
}

var ConfigDefaults = Config{
	CollectorsEnabled: []string{
		SubCollectorDataStore,
		SubCollectorDynamicMemoryBalancer,
		SubCollectorDynamicMemoryVM,
		SubCollectorHypervisorLogicalProcessor,
		SubCollectorHypervisorRootPartition,
		SubCollectorHypervisorRootVirtualProcessor,
		SubCollectorHypervisorVirtualProcessor,
		SubCollectorLegacyNetworkAdapter,
		SubCollectorVirtualMachineHealthSummary,
		SubCollectorVirtualMachineVidPartition,
		SubCollectorVirtualNetworkAdapter,
		SubCollectorVirtualNetworkAdapterDropReasons,
		SubCollectorVirtualSMB,
		SubCollectorVirtualStorageDevice,
		SubCollectorVirtualSwitch,
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

func (c *Collector) GetPerfCounter(_ *slog.Logger) ([]string, error) {
	return []string{}, nil
}

func (c *Collector) Close(_ *slog.Logger) error {
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

	for _, collector := range c.config.CollectorsEnabled {
		if !slices.Contains([]string{
			SubCollectorDataStore,
			SubCollectorDynamicMemoryBalancer,
			SubCollectorDynamicMemoryVM,
			SubCollectorHypervisorLogicalProcessor,
			SubCollectorHypervisorRootPartition,
			SubCollectorHypervisorRootVirtualProcessor,
			SubCollectorHypervisorVirtualProcessor,
			SubCollectorLegacyNetworkAdapter,
			SubCollectorVirtualMachineHealthSummary,
			SubCollectorVirtualMachineVidPartition,
			SubCollectorVirtualNetworkAdapter,
			SubCollectorVirtualNetworkAdapterDropReasons,
			SubCollectorVirtualSMB,
			SubCollectorVirtualStorageDevice,
			SubCollectorVirtualSwitch,
		}, collector) {
			return errors.New("invalid collector: " + collector)
		}
	}

	if slices.Contains(c.config.CollectorsEnabled, SubCollectorDataStore) {
		logger.Info("Hyper-V datastore collector is in an experimental state! Metrics for this collector have not been tested.",
			slog.String("collector", Name),
		)

		if err := c.buildDataStore(); err != nil {
			return err
		}

		c.collectorFns = append(c.collectorFns, c.collectDataStore)
		c.closeFns = append(c.closeFns, c.perfDataCollectorDataStore.Close)
	}

	if slices.Contains(c.config.CollectorsEnabled, SubCollectorDynamicMemoryBalancer) {
		if err := c.buildDynamicMemoryBalancer(); err != nil {
			return err
		}

		c.collectorFns = append(c.collectorFns, c.collectDynamicMemoryBalancer)
		c.closeFns = append(c.closeFns, c.perfDataCollectorDynamicMemoryBalancer.Close)
	}

	if slices.Contains(c.config.CollectorsEnabled, SubCollectorDynamicMemoryVM) {
		if err := c.buildDynamicMemoryVM(); err != nil {
			return err
		}

		c.collectorFns = append(c.collectorFns, c.collectDynamicMemoryVM)
		c.closeFns = append(c.closeFns, c.perfDataCollectorDynamicMemoryVM.Close)
	}

	if slices.Contains(c.config.CollectorsEnabled, SubCollectorHypervisorLogicalProcessor) {
		if err := c.buildHypervisorLogicalProcessor(); err != nil {
			return err
		}

		c.collectorFns = append(c.collectorFns, c.collectHypervisorLogicalProcessor)
		c.closeFns = append(c.closeFns, c.perfDataCollectorHypervisorLogicalProcessor.Close)
	}

	if slices.Contains(c.config.CollectorsEnabled, SubCollectorHypervisorRootPartition) {
		if err := c.buildHypervisorRootPartition(); err != nil {
			return err
		}

		c.collectorFns = append(c.collectorFns, c.collectHypervisorRootPartition)
		c.closeFns = append(c.closeFns, c.perfDataCollectorHypervisorRootPartition.Close)
	}

	if slices.Contains(c.config.CollectorsEnabled, SubCollectorHypervisorRootVirtualProcessor) {
		if err := c.buildHypervisorRootVirtualProcessor(); err != nil {
			return err
		}

		c.collectorFns = append(c.collectorFns, c.collectHypervisorRootVirtualProcessor)
		c.closeFns = append(c.closeFns, c.perfDataCollectorHypervisorRootVirtualProcessor.Close)
	}

	if slices.Contains(c.config.CollectorsEnabled, SubCollectorHypervisorVirtualProcessor) {
		if err := c.buildHypervisorVirtualProcessor(); err != nil {
			return err
		}

		c.collectorFns = append(c.collectorFns, c.collectHypervisorVirtualProcessor)
		c.closeFns = append(c.closeFns, c.perfDataCollectorHypervisorVirtualProcessor.Close)
	}

	if slices.Contains(c.config.CollectorsEnabled, SubCollectorLegacyNetworkAdapter) {
		if err := c.buildLegacyNetworkAdapter(); err != nil {
			return err
		}

		c.collectorFns = append(c.collectorFns, c.collectLegacyNetworkAdapter)
		c.closeFns = append(c.closeFns, c.perfDataCollectorLegacyNetworkAdapter.Close)
	}

	if slices.Contains(c.config.CollectorsEnabled, SubCollectorVirtualMachineHealthSummary) {
		if err := c.buildVirtualMachineHealthSummary(); err != nil {
			return err
		}

		c.collectorFns = append(c.collectorFns, c.collectVirtualMachineHealthSummary)
		c.closeFns = append(c.closeFns, c.perfDataCollectorVirtualMachineHealthSummary.Close)
	}

	if slices.Contains(c.config.CollectorsEnabled, SubCollectorVirtualMachineVidPartition) {
		if err := c.buildVirtualMachineVidPartition(); err != nil {
			return err
		}

		c.collectorFns = append(c.collectorFns, c.collectVirtualMachineVidPartition)
		c.closeFns = append(c.closeFns, c.perfDataCollectorVirtualMachineVidPartition.Close)
	}

	if slices.Contains(c.config.CollectorsEnabled, SubCollectorVirtualNetworkAdapter) {
		if err := c.buildVirtualNetworkAdapter(); err != nil {
			return err
		}

		c.collectorFns = append(c.collectorFns, c.collectVirtualNetworkAdapter)
		c.closeFns = append(c.closeFns, c.perfDataCollectorVirtualNetworkAdapter.Close)
	}

	if slices.Contains(c.config.CollectorsEnabled, SubCollectorVirtualNetworkAdapterDropReasons) {
		if err := c.buildVirtualNetworkAdapterDropReasons(); err != nil {
			return err
		}

		c.collectorFns = append(c.collectorFns, c.collectVirtualNetworkAdapterDropReasons)
		c.closeFns = append(c.closeFns, c.perfDataCollectorVirtualNetworkAdapterDropReasons.Close)
	}

	if slices.Contains(c.config.CollectorsEnabled, SubCollectorVirtualSMB) {
		logger.Info("Hyper-V Virtual SMB collector is in an experimental state! Metrics for this collector have not been tested.",
			slog.String("collector", Name),
		)

		if err := c.buildVirtualSMB(); err != nil {
			return err
		}

		c.collectorFns = append(c.collectorFns, c.collectVirtualSMB)
		c.closeFns = append(c.closeFns, c.perfDataCollectorVirtualSMB.Close)
	}

	if slices.Contains(c.config.CollectorsEnabled, SubCollectorVirtualStorageDevice) {
		if err := c.buildVirtualStorageDevice(); err != nil {
			return err
		}

		c.collectorFns = append(c.collectorFns, c.collectVirtualStorageDevice)
		c.closeFns = append(c.closeFns, c.perfDataCollectorVirtualStorageDevice.Close)
	}

	if slices.Contains(c.config.CollectorsEnabled, SubCollectorVirtualSwitch) {
		if err := c.buildVirtualSwitch(); err != nil {
			return err
		}

		c.collectorFns = append(c.collectorFns, c.collectVirtualSwitch)
		c.closeFns = append(c.closeFns, c.perfDataCollectorVirtualSwitch.Close)
	}

	return nil
}

// Collect sends the metric values for each metric
// to the provided prometheus Metric channel.
func (c *Collector) Collect(_ *types.ScrapeContext, _ *slog.Logger, ch chan<- prometheus.Metric) error {
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
