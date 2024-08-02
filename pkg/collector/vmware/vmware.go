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

// A Collector is a Prometheus Collector for WMI Win32_PerfRawData_vmGuestLib_VMem/Win32_PerfRawData_vmGuestLib_VCPU metrics
type Collector struct {
	logger log.Logger

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

func New(logger log.Logger, _ *Config) *Collector {
	c := &Collector{}
	c.SetLogger(logger)

	return c
}

func NewWithFlags(_ *kingpin.Application) *Collector {
	return &Collector{}
}

func (c *Collector) GetName() string {
	return Name
}

func (c *Collector) SetLogger(logger log.Logger) {
	c.logger = log.With(logger, "collector", Name)
}

func (c *Collector) GetPerfCounter() ([]string, error) {
	return []string{}, nil
}

func (c *Collector) Close() error {
	return nil
}

func (c *Collector) Build() error {
	c.MemActive = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "mem_active_bytes"),
		"(MemActiveMB)",
		nil,
		nil,
	)
	c.MemBallooned = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "mem_ballooned_bytes"),
		"(MemBalloonedMB)",
		nil,
		nil,
	)
	c.MemLimit = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "mem_limit_bytes"),
		"(MemLimitMB)",
		nil,
		nil,
	)
	c.MemMapped = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "mem_mapped_bytes"),
		"(MemMappedMB)",
		nil,
		nil,
	)
	c.MemOverhead = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "mem_overhead_bytes"),
		"(MemOverheadMB)",
		nil,
		nil,
	)
	c.MemReservation = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "mem_reservation_bytes"),
		"(MemReservationMB)",
		nil,
		nil,
	)
	c.MemShared = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "mem_shared_bytes"),
		"(MemSharedMB)",
		nil,
		nil,
	)
	c.MemSharedSaved = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "mem_shared_saved_bytes"),
		"(MemSharedSavedMB)",
		nil,
		nil,
	)
	c.MemShares = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "mem_shares"),
		"(MemShares)",
		nil,
		nil,
	)
	c.MemSwapped = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "mem_swapped_bytes"),
		"(MemSwappedMB)",
		nil,
		nil,
	)
	c.MemTargetSize = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "mem_target_size_bytes"),
		"(MemTargetSizeMB)",
		nil,
		nil,
	)
	c.MemUsed = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "mem_used_bytes"),
		"(MemUsedMB)",
		nil,
		nil,
	)

	c.CpuLimitMHz = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "cpu_limit_mhz"),
		"(CpuLimitMHz)",
		nil,
		nil,
	)
	c.CpuReservationMHz = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "cpu_reservation_mhz"),
		"(CpuReservationMHz)",
		nil,
		nil,
	)
	c.CpuShares = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "cpu_shares"),
		"(CpuShares)",
		nil,
		nil,
	)
	c.CpuStolenTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "cpu_stolen_seconds_total"),
		"(CpuStolenMs)",
		nil,
		nil,
	)
	c.CpuTimeTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "cpu_time_seconds_total"),
		"(CpuTimePercents)",
		nil,
		nil,
	)
	c.EffectiveVMSpeedMHz = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "effective_vm_speed_mhz"),
		"(EffectiveVMSpeedMHz)",
		nil,
		nil,
	)
	c.HostProcessorSpeedMHz = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "host_processor_speed_mhz"),
		"(HostProcessorSpeedMHz)",
		nil,
		nil,
	)
	return nil
}

// Collect sends the metric values for each metric
// to the provided prometheus Metric channel.
func (c *Collector) Collect(_ *types.ScrapeContext, ch chan<- prometheus.Metric) error {
	if err := c.collectMem(ch); err != nil {
		_ = level.Error(c.logger).Log("msg", "failed collecting vmware memory metrics", "err", err)
		return err
	}
	if err := c.collectCpu(ch); err != nil {
		_ = level.Error(c.logger).Log("msg", "failed collecting vmware cpu metrics", "err", err)
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

func (c *Collector) collectMem(ch chan<- prometheus.Metric) error {
	var dst []Win32_PerfRawData_vmGuestLib_VMem
	q := wmi.QueryAll(&dst, c.logger)
	if err := wmi.Query(q, &dst); err != nil {
		return err
	}
	if len(dst) == 0 {
		return errors.New("WMI query returned empty result set")
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

	return nil
}

func mbToBytes(mb uint64) float64 {
	return float64(mb * 1024 * 1024)
}

func (c *Collector) collectCpu(ch chan<- prometheus.Metric) error {
	var dst []Win32_PerfRawData_vmGuestLib_VCPU
	q := wmi.QueryAll(&dst, c.logger)
	if err := wmi.Query(q, &dst); err != nil {
		return err
	}
	if len(dst) == 0 {
		return errors.New("WMI query returned empty result set")
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
		float64(dst[0].CpuStolenMs)*perflib.TicksToSecondScaleFactor,
	)

	ch <- prometheus.MustNewConstMetric(
		c.CpuTimeTotal,
		prometheus.CounterValue,
		float64(dst[0].CpuTimePercents)*perflib.TicksToSecondScaleFactor,
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

	return nil
}
