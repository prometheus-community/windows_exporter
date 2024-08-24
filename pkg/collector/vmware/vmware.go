//go:build windows

package vmware

import (
	"errors"

	"github.com/alecthomas/kingpin/v2"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/prometheus-community/windows_exporter/pkg/perflib"
	"github.com/prometheus-community/windows_exporter/pkg/types"
	"github.com/prometheus-community/windows_exporter/pkg/wmi"
	"github.com/prometheus/client_golang/prometheus"
)

const Name = "vmware"

type Config struct{}

var ConfigDefaults = Config{}

// A Collector is a Prometheus Collector for WMI Win32_PerfRawData_vmGuestLib_VMem/Win32_PerfRawData_vmGuestLib_VCPU metrics.
type Collector struct {
	config Config

	memActive      *prometheus.Desc
	memBallooned   *prometheus.Desc
	memLimit       *prometheus.Desc
	memMapped      *prometheus.Desc
	memOverhead    *prometheus.Desc
	memReservation *prometheus.Desc
	memShared      *prometheus.Desc
	memSharedSaved *prometheus.Desc
	memShares      *prometheus.Desc
	memSwapped     *prometheus.Desc
	memTargetSize  *prometheus.Desc
	memUsed        *prometheus.Desc

	cpuLimitMHz           *prometheus.Desc
	cpuReservationMHz     *prometheus.Desc
	cpuShares             *prometheus.Desc
	cpuStolenTotal        *prometheus.Desc
	cpuTimeTotal          *prometheus.Desc
	effectiveVMSpeedMHz   *prometheus.Desc
	hostProcessorSpeedMHz *prometheus.Desc
}

func New(config *Config) *Collector {
	if config == nil {
		config = &ConfigDefaults
	}

	c := &Collector{
		config: *config,
	}

	return c
}

func NewWithFlags(_ *kingpin.Application) *Collector {
	return &Collector{}
}

func (c *Collector) GetName() string {
	return Name
}

func (c *Collector) GetPerfCounter(_ log.Logger) ([]string, error) {
	return []string{}, nil
}

func (c *Collector) Close() error {
	return nil
}

func (c *Collector) Build(_ log.Logger) error {
	c.memActive = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "mem_active_bytes"),
		"(MemActiveMB)",
		nil,
		nil,
	)
	c.memBallooned = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "mem_ballooned_bytes"),
		"(MemBalloonedMB)",
		nil,
		nil,
	)
	c.memLimit = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "mem_limit_bytes"),
		"(MemLimitMB)",
		nil,
		nil,
	)
	c.memMapped = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "mem_mapped_bytes"),
		"(MemMappedMB)",
		nil,
		nil,
	)
	c.memOverhead = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "mem_overhead_bytes"),
		"(MemOverheadMB)",
		nil,
		nil,
	)
	c.memReservation = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "mem_reservation_bytes"),
		"(MemReservationMB)",
		nil,
		nil,
	)
	c.memShared = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "mem_shared_bytes"),
		"(MemSharedMB)",
		nil,
		nil,
	)
	c.memSharedSaved = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "mem_shared_saved_bytes"),
		"(MemSharedSavedMB)",
		nil,
		nil,
	)
	c.memShares = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "mem_shares"),
		"(MemShares)",
		nil,
		nil,
	)
	c.memSwapped = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "mem_swapped_bytes"),
		"(MemSwappedMB)",
		nil,
		nil,
	)
	c.memTargetSize = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "mem_target_size_bytes"),
		"(MemTargetSizeMB)",
		nil,
		nil,
	)
	c.memUsed = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "mem_used_bytes"),
		"(MemUsedMB)",
		nil,
		nil,
	)

	c.cpuLimitMHz = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "cpu_limit_mhz"),
		"(CpuLimitMHz)",
		nil,
		nil,
	)
	c.cpuReservationMHz = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "cpu_reservation_mhz"),
		"(CpuReservationMHz)",
		nil,
		nil,
	)
	c.cpuShares = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "cpu_shares"),
		"(CpuShares)",
		nil,
		nil,
	)
	c.cpuStolenTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "cpu_stolen_seconds_total"),
		"(CpuStolenMs)",
		nil,
		nil,
	)
	c.cpuTimeTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "cpu_time_seconds_total"),
		"(CpuTimePercents)",
		nil,
		nil,
	)
	c.effectiveVMSpeedMHz = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "effective_vm_speed_mhz"),
		"(EffectiveVMSpeedMHz)",
		nil,
		nil,
	)
	c.hostProcessorSpeedMHz = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "host_processor_speed_mhz"),
		"(HostProcessorSpeedMHz)",
		nil,
		nil,
	)
	return nil
}

// Collect sends the metric values for each metric
// to the provided prometheus Metric channel.
func (c *Collector) Collect(_ *types.ScrapeContext, logger log.Logger, ch chan<- prometheus.Metric) error {
	logger = log.With(logger, "collector", Name)
	if err := c.collectMem(logger, ch); err != nil {
		_ = level.Error(logger).Log("msg", "failed collecting vmware memory metrics", "err", err)
		return err
	}
	if err := c.collectCpu(logger, ch); err != nil {
		_ = level.Error(logger).Log("msg", "failed collecting vmware cpu metrics", "err", err)
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

func (c *Collector) collectMem(logger log.Logger, ch chan<- prometheus.Metric) error {
	var dst []Win32_PerfRawData_vmGuestLib_VMem
	q := wmi.QueryAll(&dst, logger)
	if err := wmi.Query(q, &dst); err != nil {
		return err
	}
	if len(dst) == 0 {
		return errors.New("WMI query returned empty result set")
	}

	ch <- prometheus.MustNewConstMetric(
		c.memActive,
		prometheus.GaugeValue,
		mbToBytes(dst[0].MemActiveMB),
	)

	ch <- prometheus.MustNewConstMetric(
		c.memBallooned,
		prometheus.GaugeValue,
		mbToBytes(dst[0].MemBalloonedMB),
	)

	ch <- prometheus.MustNewConstMetric(
		c.memLimit,
		prometheus.GaugeValue,
		mbToBytes(dst[0].MemLimitMB),
	)

	ch <- prometheus.MustNewConstMetric(
		c.memMapped,
		prometheus.GaugeValue,
		mbToBytes(dst[0].MemMappedMB),
	)

	ch <- prometheus.MustNewConstMetric(
		c.memOverhead,
		prometheus.GaugeValue,
		mbToBytes(dst[0].MemOverheadMB),
	)

	ch <- prometheus.MustNewConstMetric(
		c.memReservation,
		prometheus.GaugeValue,
		mbToBytes(dst[0].MemReservationMB),
	)

	ch <- prometheus.MustNewConstMetric(
		c.memShared,
		prometheus.GaugeValue,
		mbToBytes(dst[0].MemSharedMB),
	)

	ch <- prometheus.MustNewConstMetric(
		c.memSharedSaved,
		prometheus.GaugeValue,
		mbToBytes(dst[0].MemSharedSavedMB),
	)

	ch <- prometheus.MustNewConstMetric(
		c.memShares,
		prometheus.GaugeValue,
		float64(dst[0].MemShares),
	)

	ch <- prometheus.MustNewConstMetric(
		c.memSwapped,
		prometheus.GaugeValue,
		mbToBytes(dst[0].MemSwappedMB),
	)

	ch <- prometheus.MustNewConstMetric(
		c.memTargetSize,
		prometheus.GaugeValue,
		mbToBytes(dst[0].MemTargetSizeMB),
	)

	ch <- prometheus.MustNewConstMetric(
		c.memUsed,
		prometheus.GaugeValue,
		mbToBytes(dst[0].MemUsedMB),
	)

	return nil
}

func mbToBytes(mb uint64) float64 {
	return float64(mb * 1024 * 1024)
}

func (c *Collector) collectCpu(logger log.Logger, ch chan<- prometheus.Metric) error {
	var dst []Win32_PerfRawData_vmGuestLib_VCPU
	q := wmi.QueryAll(&dst, logger)
	if err := wmi.Query(q, &dst); err != nil {
		return err
	}
	if len(dst) == 0 {
		return errors.New("WMI query returned empty result set")
	}

	ch <- prometheus.MustNewConstMetric(
		c.cpuLimitMHz,
		prometheus.GaugeValue,
		float64(dst[0].CpuLimitMHz),
	)

	ch <- prometheus.MustNewConstMetric(
		c.cpuReservationMHz,
		prometheus.GaugeValue,
		float64(dst[0].CpuReservationMHz),
	)

	ch <- prometheus.MustNewConstMetric(
		c.cpuShares,
		prometheus.GaugeValue,
		float64(dst[0].CpuShares),
	)

	ch <- prometheus.MustNewConstMetric(
		c.cpuStolenTotal,
		prometheus.CounterValue,
		float64(dst[0].CpuStolenMs)*perflib.TicksToSecondScaleFactor,
	)

	ch <- prometheus.MustNewConstMetric(
		c.cpuTimeTotal,
		prometheus.CounterValue,
		float64(dst[0].CpuTimePercents)*perflib.TicksToSecondScaleFactor,
	)

	ch <- prometheus.MustNewConstMetric(
		c.effectiveVMSpeedMHz,
		prometheus.GaugeValue,
		float64(dst[0].EffectiveVMSpeedMHz),
	)

	ch <- prometheus.MustNewConstMetric(
		c.hostProcessorSpeedMHz,
		prometheus.GaugeValue,
		float64(dst[0].HostProcessorSpeedMHz),
	)

	return nil
}
