//go:build windows
// +build windows

package collector

import (
	"errors"

	"github.com/StackExchange/wmi"
	"github.com/prometheus-community/windows_exporter/log"
	"github.com/prometheus/client_golang/prometheus"
)

func init() {
	registerCollector("vmware", NewVmwareCollector)
}

// A VmwareCollector is a Prometheus collector for WMI Win32_PerfRawData_vmGuestLib_VMem/Win32_PerfRawData_vmGuestLib_VCPU metrics
type VmwareCollector struct {
	MemActive      *prometheus.Desc
	MemBallooned   *prometheus.Desc
	MemLimit       *prometheus.Desc
	MemMapped      *prometheus.Desc
	MemOverhead    *prometheus.Desc
	MemReservation *prometheus.Desc
	MemShared      *prometheus.Desc
	MemSharedSaved *prometheus.Desc
	MemShares      *prometheus.Desc
	MemSwapped     *prometheus.Desc
	MemTargetSize  *prometheus.Desc
	MemUsed        *prometheus.Desc

	CpuLimitMHz           *prometheus.Desc
	CpuReservationMHz     *prometheus.Desc
	CpuShares             *prometheus.Desc
	CpuStolenTotal        *prometheus.Desc
	CpuTimeTotal          *prometheus.Desc
	EffectiveVMSpeedMHz   *prometheus.Desc
	HostProcessorSpeedMHz *prometheus.Desc
}

// NewVmwareCollector constructs a new VmwareCollector
func NewVmwareCollector() (Collector, error) {
	const subsystem = "vmware"
	return &VmwareCollector{
		MemActive: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "mem_active_bytes"),
			"(MemActiveMB)",
			nil,
			nil,
		),
		MemBallooned: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "mem_ballooned_bytes"),
			"(MemBalloonedMB)",
			nil,
			nil,
		),
		MemLimit: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "mem_limit_bytes"),
			"(MemLimitMB)",
			nil,
			nil,
		),
		MemMapped: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "mem_mapped_bytes"),
			"(MemMappedMB)",
			nil,
			nil,
		),
		MemOverhead: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "mem_overhead_bytes"),
			"(MemOverheadMB)",
			nil,
			nil,
		),
		MemReservation: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "mem_reservation_bytes"),
			"(MemReservationMB)",
			nil,
			nil,
		),
		MemShared: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "mem_shared_bytes"),
			"(MemSharedMB)",
			nil,
			nil,
		),
		MemSharedSaved: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "mem_shared_saved_bytes"),
			"(MemSharedSavedMB)",
			nil,
			nil,
		),
		MemShares: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "mem_shares"),
			"(MemShares)",
			nil,
			nil,
		),
		MemSwapped: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "mem_swapped_bytes"),
			"(MemSwappedMB)",
			nil,
			nil,
		),
		MemTargetSize: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "mem_target_size_bytes"),
			"(MemTargetSizeMB)",
			nil,
			nil,
		),
		MemUsed: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "mem_used_bytes"),
			"(MemUsedMB)",
			nil,
			nil,
		),

		CpuLimitMHz: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "cpu_limit_mhz"),
			"(CpuLimitMHz)",
			nil,
			nil,
		),
		CpuReservationMHz: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "cpu_reservation_mhz"),
			"(CpuReservationMHz)",
			nil,
			nil,
		),
		CpuShares: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "cpu_shares"),
			"(CpuShares)",
			nil,
			nil,
		),
		CpuStolenTotal: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "cpu_stolen_seconds_total"),
			"(CpuStolenMs)",
			nil,
			nil,
		),
		CpuTimeTotal: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "cpu_time_seconds_total"),
			"(CpuTimePercents)",
			nil,
			nil,
		),
		EffectiveVMSpeedMHz: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "effective_vm_speed_mhz"),
			"(EffectiveVMSpeedMHz)",
			nil,
			nil,
		),
		HostProcessorSpeedMHz: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "host_processor_speed_mhz"),
			"(HostProcessorSpeedMHz)",
			nil,
			nil,
		),
	}, nil
}

// Collect sends the metric values for each metric
// to the provided prometheus Metric channel.
func (c *VmwareCollector) Collect(ctx *ScrapeContext, ch chan<- prometheus.Metric) error {
	if desc, err := c.collectMem(ch); err != nil {
		log.Error("failed collecting vmware memory metrics:", desc, err)
		return err
	}
	if desc, err := c.collectCpu(ch); err != nil {
		log.Error("failed collecting vmware cpu metrics:", desc, err)
		return err
	}
	return nil
}

type Win32_PerfRawData_vmGuestLib_VMem struct {
	MemActiveMB      uint64
	MemBalloonedMB   uint64
	MemLimitMB       uint64
	MemMappedMB      uint64
	MemOverheadMB    uint64
	MemReservationMB uint64
	MemSharedMB      uint64
	MemSharedSavedMB uint64
	MemShares        uint64
	MemSwappedMB     uint64
	MemTargetSizeMB  uint64
	MemUsedMB        uint64
}

type Win32_PerfRawData_vmGuestLib_VCPU struct {
	CpuLimitMHz           uint64
	CpuReservationMHz     uint64
	CpuShares             uint64
	CpuStolenMs           uint64
	CpuTimePercents       uint64
	EffectiveVMSpeedMHz   uint64
	HostProcessorSpeedMHz uint64
}

func (c *VmwareCollector) collectMem(ch chan<- prometheus.Metric) (*prometheus.Desc, error) {
	var dst []Win32_PerfRawData_vmGuestLib_VMem
	q := queryAll(&dst)
	if err := wmi.Query(q, &dst); err != nil {
		return nil, err
	}
	if len(dst) == 0 {
		return nil, errors.New("WMI query returned empty result set")
	}

	ch <- prometheus.MustNewConstMetric(
		c.MemActive,
		prometheus.GaugeValue,
		mbToBytes(dst[0].MemActiveMB),
	)

	ch <- prometheus.MustNewConstMetric(
		c.MemBallooned,
		prometheus.GaugeValue,
		mbToBytes(dst[0].MemBalloonedMB),
	)

	ch <- prometheus.MustNewConstMetric(
		c.MemLimit,
		prometheus.GaugeValue,
		mbToBytes(dst[0].MemLimitMB),
	)

	ch <- prometheus.MustNewConstMetric(
		c.MemMapped,
		prometheus.GaugeValue,
		mbToBytes(dst[0].MemMappedMB),
	)

	ch <- prometheus.MustNewConstMetric(
		c.MemOverhead,
		prometheus.GaugeValue,
		mbToBytes(dst[0].MemOverheadMB),
	)

	ch <- prometheus.MustNewConstMetric(
		c.MemReservation,
		prometheus.GaugeValue,
		mbToBytes(dst[0].MemReservationMB),
	)

	ch <- prometheus.MustNewConstMetric(
		c.MemShared,
		prometheus.GaugeValue,
		mbToBytes(dst[0].MemSharedMB),
	)

	ch <- prometheus.MustNewConstMetric(
		c.MemSharedSaved,
		prometheus.GaugeValue,
		mbToBytes(dst[0].MemSharedSavedMB),
	)

	ch <- prometheus.MustNewConstMetric(
		c.MemShares,
		prometheus.GaugeValue,
		float64(dst[0].MemShares),
	)

	ch <- prometheus.MustNewConstMetric(
		c.MemSwapped,
		prometheus.GaugeValue,
		mbToBytes(dst[0].MemSwappedMB),
	)

	ch <- prometheus.MustNewConstMetric(
		c.MemTargetSize,
		prometheus.GaugeValue,
		mbToBytes(dst[0].MemTargetSizeMB),
	)

	ch <- prometheus.MustNewConstMetric(
		c.MemUsed,
		prometheus.GaugeValue,
		mbToBytes(dst[0].MemUsedMB),
	)

	return nil, nil
}

func mbToBytes(mb uint64) float64 {
	return float64(mb * 1024 * 1024)
}

func (c *VmwareCollector) collectCpu(ch chan<- prometheus.Metric) (*prometheus.Desc, error) {
	var dst []Win32_PerfRawData_vmGuestLib_VCPU
	q := queryAll(&dst)
	if err := wmi.Query(q, &dst); err != nil {
		return nil, err
	}
	if len(dst) == 0 {
		return nil, errors.New("WMI query returned empty result set")
	}

	ch <- prometheus.MustNewConstMetric(
		c.CpuLimitMHz,
		prometheus.GaugeValue,
		float64(dst[0].CpuLimitMHz),
	)

	ch <- prometheus.MustNewConstMetric(
		c.CpuReservationMHz,
		prometheus.GaugeValue,
		float64(dst[0].CpuReservationMHz),
	)

	ch <- prometheus.MustNewConstMetric(
		c.CpuShares,
		prometheus.GaugeValue,
		float64(dst[0].CpuShares),
	)

	ch <- prometheus.MustNewConstMetric(
		c.CpuStolenTotal,
		prometheus.CounterValue,
		float64(dst[0].CpuStolenMs)*ticksToSecondsScaleFactor,
	)

	ch <- prometheus.MustNewConstMetric(
		c.CpuTimeTotal,
		prometheus.CounterValue,
		float64(dst[0].CpuTimePercents)*ticksToSecondsScaleFactor,
	)

	ch <- prometheus.MustNewConstMetric(
		c.EffectiveVMSpeedMHz,
		prometheus.GaugeValue,
		float64(dst[0].EffectiveVMSpeedMHz),
	)

	ch <- prometheus.MustNewConstMetric(
		c.HostProcessorSpeedMHz,
		prometheus.GaugeValue,
		float64(dst[0].HostProcessorSpeedMHz),
	)

	return nil, nil
}
