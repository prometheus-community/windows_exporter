package exchange

import (
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/prometheus-community/windows_exporter/internal/perfdata"
	v1 "github.com/prometheus-community/windows_exporter/internal/perfdata/v1"
	"github.com/prometheus-community/windows_exporter/internal/types"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	activeTasks    = "ActiveTasks"
	completedTasks = "CompletedTasks"
	queuedTasks    = "QueuedTasks"
	yieldedTasks   = "YieldedTasks"
	isActive       = "Active"
)

// Perflib: [19430] MSExchange WorkloadManagement Workloads.
type perflibWorkloadManagementWorkloads struct {
	Name string

	ActiveTasks    float64 `perflib:"ActiveTasks"`
	CompletedTasks float64 `perflib:"CompletedTasks"`
	QueuedTasks    float64 `perflib:"QueuedTasks"`
	YieldedTasks   float64 `perflib:"YieldedTasks"`
	IsActive       float64 `perflib:"Active"`
}

func (c *Collector) buildWorkloadManagementWorkloads() error {
	counters := []string{
		activeTasks,
		completedTasks,
		queuedTasks,
		yieldedTasks,
		isActive,
	}

	var err error

	c.perfDataCollectorWorkloadManagementWorkloads, err = perfdata.NewCollector(perfdata.V2, "MSExchange WorkloadManagement Workloads", perfdata.AllInstances, counters)
	if err != nil {
		return fmt.Errorf("failed to create MSExchange WorkloadManagement Workloads collector: %w", err)
	}

	return nil
}

func (c *Collector) collectWorkloadManagementWorkloads(ctx *types.ScrapeContext, logger *slog.Logger, ch chan<- prometheus.Metric) error {
	var data []perflibWorkloadManagementWorkloads

	if err := v1.UnmarshalObject(ctx.PerfObjects["MSExchange WorkloadManagement Workloads"], &data, logger); err != nil {
		return err
	}

	for _, instance := range data {
		labelName := c.toLabelName(instance.Name)
		if strings.HasSuffix(labelName, "_total") {
			continue
		}
		ch <- prometheus.MustNewConstMetric(
			c.activeTasks,
			prometheus.GaugeValue,
			instance.ActiveTasks,
			labelName,
		)
		ch <- prometheus.MustNewConstMetric(
			c.completedTasks,
			prometheus.CounterValue,
			instance.CompletedTasks,
			labelName,
		)
		ch <- prometheus.MustNewConstMetric(
			c.queuedTasks,
			prometheus.CounterValue,
			instance.QueuedTasks,
			labelName,
		)
		ch <- prometheus.MustNewConstMetric(
			c.yieldedTasks,
			prometheus.CounterValue,
			instance.YieldedTasks,
			labelName,
		)
		ch <- prometheus.MustNewConstMetric(
			c.isActive,
			prometheus.GaugeValue,
			instance.IsActive,
			labelName,
		)
	}

	return nil
}

func (c *Collector) collectPDHWorkloadManagementWorkloads(ch chan<- prometheus.Metric) error {
	perfData, err := c.perfDataCollectorWorkloadManagementWorkloads.Collect()
	if err != nil {
		return fmt.Errorf("failed to collect MSExchange WorkloadManagement Workloads: %w", err)
	}

	if len(perfData) == 0 {
		return errors.New("perflib query for MSExchange WorkloadManagement Workloads returned empty result set")
	}

	for name, data := range perfData {
		labelName := c.toLabelName(name)

		ch <- prometheus.MustNewConstMetric(
			c.activeTasks,
			prometheus.GaugeValue,
			data[activeTasks].FirstValue,
			labelName,
		)
		ch <- prometheus.MustNewConstMetric(
			c.completedTasks,
			prometheus.CounterValue,
			data[completedTasks].FirstValue,
			labelName,
		)
		ch <- prometheus.MustNewConstMetric(
			c.queuedTasks,
			prometheus.CounterValue,
			data[queuedTasks].FirstValue,
			labelName,
		)
		ch <- prometheus.MustNewConstMetric(
			c.yieldedTasks,
			prometheus.CounterValue,
			data[yieldedTasks].FirstValue,
			labelName,
		)
		ch <- prometheus.MustNewConstMetric(
			c.isActive,
			prometheus.GaugeValue,
			data[isActive].FirstValue,
			labelName,
		)
	}

	return nil
}
