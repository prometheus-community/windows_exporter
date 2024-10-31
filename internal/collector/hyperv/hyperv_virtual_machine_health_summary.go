package hyperv

import (
	"errors"
	"fmt"

	"github.com/prometheus-community/windows_exporter/internal/perfdata"
	"github.com/prometheus-community/windows_exporter/internal/perfdata/perftypes"
	"github.com/prometheus-community/windows_exporter/internal/types"
	"github.com/prometheus/client_golang/prometheus"
)

// collectorVirtualMachineHealthSummary Hyper-V Virtual Machine Health Summary metrics
type collectorVirtualMachineHealthSummary struct {
	perfDataCollectorVirtualMachineHealthSummary perfdata.Collector
	healthCritical                               *prometheus.Desc // \Hyper-V Virtual Machine Health Summary\Health Critical
	healthOk                                     *prometheus.Desc // \Hyper-V Virtual Machine Health Summary\Health Ok
}

const (
	// Hyper-V Virtual Machine Health Summary
	healthCritical = "Health Critical"
	healthOk       = "Health Ok"
)

func (c *Collector) buildVirtualMachineHealthSummary() error {
	var err error

	c.perfDataCollectorVirtualMachineHealthSummary, err = perfdata.NewCollector(perfdata.V2, "Hyper-V Virtual Machine Health Summary", perfdata.AllInstances, []string{
		healthCritical,
		healthOk,
	})

	if err != nil {
		return fmt.Errorf("failed to create Hyper-V Virtual Machine Health Summary collector: %w", err)
	}

	c.healthCritical = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "health_critical"),
		"This counter represents the number of virtual machines with critical health",
		nil,
		nil,
	)
	c.healthOk = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "health_ok"),
		"This counter represents the number of virtual machines with ok health",
		nil,
		nil,
	)

	return nil
}

func (c *Collector) collectVirtualMachineHealthSummary(ch chan<- prometheus.Metric) error {
	data, err := c.perfDataCollectorVirtualMachineHealthSummary.Collect()
	if err != nil {
		return fmt.Errorf("failed to collect Hyper-V Virtual Machine Health Summary metrics: %w", err)
	} else if len(data) == 0 {
		return errors.New("perflib query for Hyper-V Virtual Machine Health Summary returned empty result set")
	}

	healthData, ok := data[perftypes.EmptyInstance]
	if !ok {
		return errors.New("no data returned for Hyper-V Virtual Machine Health Summary")
	}

	ch <- prometheus.MustNewConstMetric(
		c.healthCritical,
		prometheus.GaugeValue,
		healthData[healthCritical].FirstValue,
	)

	ch <- prometheus.MustNewConstMetric(
		c.healthOk,
		prometheus.GaugeValue,
		healthData[healthOk].FirstValue,
	)

	return nil
}
