package hyperv

import (
	"errors"
	"fmt"
	"strings"

	"github.com/prometheus-community/windows_exporter/internal/perfdata"
	"github.com/prometheus-community/windows_exporter/internal/types"
	"github.com/prometheus/client_golang/prometheus"
)

// collectorHypervisorLogicalProcessor Hyper-V Hypervisor Logical Processor metrics
type collectorHypervisorLogicalProcessor struct {
	perfDataCollectorHypervisorLogicalProcessor perfdata.Collector

	// \Hyper-V Hypervisor Logical Processor(*)\% Guest Run Time
	// \Hyper-V Hypervisor Logical Processor(*)\% Total Run Time
	// \Hyper-V Hypervisor Logical Processor(*)\% Hypervisor Run Time
	// \Hyper-V Hypervisor Logical Processor(*)\% Idle Time
	hypervisorLogicalProcessorRunTimeTotal    *prometheus.Desc
	hypervisorLogicalProcessorContextSwitches *prometheus.Desc // \Hyper-V Hypervisor Logical Processor(*)\Context Switches/sec
}

const (
	hypervisorLogicalProcessorGuestRunTimePercent      = "% Guest Run Time"
	hypervisorLogicalProcessorHypervisorRunTimePercent = "% Hypervisor Run Time"
	hypervisorLogicalProcessorTotalRunTimePercent      = "% Total Run Time"
	hypervisorLogicalProcessorIdleRunTimePercent       = "% Idle Time"
	hypervisorLogicalProcessorContextSwitches          = "Context Switches/sec"
)

func (c *Collector) buildHypervisorLogicalProcessor() error {
	var err error

	c.perfDataCollectorHypervisorLogicalProcessor, err = perfdata.NewCollector(perfdata.V2, "Hyper-V Hypervisor Logical Processor", perfdata.AllInstances, []string{
		hypervisorLogicalProcessorGuestRunTimePercent,
		hypervisorLogicalProcessorHypervisorRunTimePercent,
		hypervisorLogicalProcessorTotalRunTimePercent,
		hypervisorLogicalProcessorIdleRunTimePercent,
		hypervisorLogicalProcessorContextSwitches,
	})

	if err != nil {
		return fmt.Errorf("failed to create Hyper-V Hypervisor Logical Processor collector: %w", err)
	}

	c.hypervisorLogicalProcessorRunTimeTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "hypervisor_logical_processor_time_total"),
		"Time that processor spent in different modes (hypervisor, guest, idle)",
		[]string{"core", "state"},
		nil,
	)

	c.hypervisorLogicalProcessorContextSwitches = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "hypervisor_logical_processor_context_switches_total"),
		"The rate of virtual processor context switches on the processor.",
		[]string{"core"},
		nil,
	)

	return nil
}

func (c *Collector) collectHypervisorLogicalProcessor(ch chan<- prometheus.Metric) error {
	data, err := c.perfDataCollectorHypervisorLogicalProcessor.Collect()
	if err != nil {
		return fmt.Errorf("failed to collect Hyper-V Hypervisor Logical Processor metrics: %w", err)
	} else if len(data) == 0 {
		return errors.New("no data returned from Hyper-V Hypervisor Root Partition")
	}

	for coreName, coreData := range data {
		// The name format is Hv LP <core id>
		parts := strings.Split(coreName, " ")
		if len(parts) != 3 {
			return fmt.Errorf("unexpected Hyper-V Hypervisor Logical Processor name format: %s", coreName)
		}

		coreId := parts[2]

		ch <- prometheus.MustNewConstMetric(
			c.hypervisorLogicalProcessorRunTimeTotal,
			prometheus.CounterValue,
			coreData[hypervisorLogicalProcessorGuestRunTimePercent].FirstValue,
			coreId, "guest",
		)

		ch <- prometheus.MustNewConstMetric(
			c.hypervisorLogicalProcessorRunTimeTotal,
			prometheus.CounterValue,
			coreData[hypervisorLogicalProcessorHypervisorRunTimePercent].FirstValue,
			coreId, "hypervisor",
		)

		ch <- prometheus.MustNewConstMetric(
			c.hypervisorLogicalProcessorRunTimeTotal,
			prometheus.CounterValue,
			coreData[hypervisorLogicalProcessorIdleRunTimePercent].FirstValue,
			coreId, "idle",
		)

		ch <- prometheus.MustNewConstMetric(
			c.hypervisorLogicalProcessorRunTimeTotal,
			prometheus.CounterValue,
			coreData[hypervisorLogicalProcessorTotalRunTimePercent].FirstValue,
			coreId, "total",
		)

		ch <- prometheus.MustNewConstMetric(
			c.hypervisorLogicalProcessorContextSwitches,
			prometheus.CounterValue,
			coreData[hypervisorLogicalProcessorContextSwitches].FirstValue,
			coreId,
		)
	}

	return nil
}
