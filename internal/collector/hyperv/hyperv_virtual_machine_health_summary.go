package hyperv

import (
	"errors"
	"fmt"

	"github.com/prometheus-community/windows_exporter/internal/perfdata"
	"github.com/prometheus-community/windows_exporter/internal/types"
	"github.com/prometheus/client_golang/prometheus"
)

// collectorVirtualMachineHealthSummary Hyper-V Virtual Machine Health Summary metrics
type collectorVirtualMachineHealthSummary struct {
	perfDataCollectorVirtualMachineHealthSummary *perfdata.Collector

	// \Hyper-V Virtual Machine Health Summary\Health Critical
	// \Hyper-V Virtual Machine Health Summary\Health Ok
	health *prometheus.Desc
}

const (
	// Hyper-V Virtual Machine Health Summary
	healthCritical = "Health Critical"
	healthOk       = "Health Ok"
)

func (c *Collector) buildVirtualMachineHealthSummary() error {
	var err error

	c.perfDataCollectorVirtualMachineHealthSummary, err = perfdata.NewCollector("Hyper-V Virtual Machine Health Summary", nil, []string{
		healthCritical,
		healthOk,
	})
	if err != nil {
		return fmt.Errorf("failed to create Hyper-V Virtual Machine Health Summary collector: %w", err)
	}

	c.health = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "virtual_machine_health_total_count"),
		"Represents the number of virtual machines with critical health",
		[]string{"state"},
		nil,
	)

	return nil
}

func (c *Collector) collectVirtualMachineHealthSummary(ch chan<- prometheus.Metric) error {
	data, err := c.perfDataCollectorVirtualMachineHealthSummary.Collect()
	if err != nil {
		return fmt.Errorf("failed to collect Hyper-V Virtual Machine Health Summary metrics: %w", err)
	}

	healthData, ok := data[perfdata.EmptyInstance]
	if !ok {
		return errors.New("no data returned for Hyper-V Virtual Machine Health Summary")
	}

	ch <- prometheus.MustNewConstMetric(
		c.health,
		prometheus.GaugeValue,
		healthData[healthCritical].FirstValue,
		"critical",
	)

	ch <- prometheus.MustNewConstMetric(
		c.health,
		prometheus.GaugeValue,
		healthData[healthOk].FirstValue,
		"ok",
	)

	return nil
}
