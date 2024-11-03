//go:build windows

package vmware

import (
	"errors"
	"fmt"
	"log/slog"

	"github.com/alecthomas/kingpin/v2"
	"github.com/prometheus-community/windows_exporter/internal/mi"
	"github.com/prometheus-community/windows_exporter/internal/perfdata/perftypes"
	"github.com/prometheus-community/windows_exporter/internal/types"
	"github.com/prometheus/client_golang/prometheus"
)

const Name = "vmware"

type Config struct{}

var ConfigDefaults = Config{}

// A Collector is a Prometheus Collector for WMI Win32_PerfRawData_vmGuestLib_VMem/Win32_PerfRawData_vmGuestLib_VCPU metrics.
type Collector struct {
	config     Config
	miSession  *mi.Session
	miQueryCPU mi.Query
	miQueryMem mi.Query

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

func (c *Collector) GetPerfCounter(_ *slog.Logger) ([]string, error) {
	return []string{}, nil
}

func (c *Collector) Close(_ *slog.Logger) error {
	return nil
}

func (c *Collector) Build(_ *slog.Logger, miSession *mi.Session) error {
	if miSession == nil {
		return errors.New("miSession is nil")
	}

	miQuery, err := mi.NewQuery("SELECT * FROM Win32_PerfRawData_vmGuestLib_VCPU")
	if err != nil {
		return fmt.Errorf("failed to create WMI query: %w", err)
	}

	c.miQueryCPU = miQuery

	miQuery, err = mi.NewQuery("SELECT * FROM Win32_PerfRawData_vmGuestLib_VMem")
	if err != nil {
		return fmt.Errorf("failed to create WMI query: %w", err)
	}

	c.miQueryMem = miQuery
	c.miSession = miSession

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
func (c *Collector) Collect(_ *types.ScrapeContext, logger *slog.Logger, ch chan<- prometheus.Metric) error {
	logger = logger.With(slog.String("collector", Name))
	if err := c.collectMem(ch); err != nil {
		logger.Error("failed collecting vmware memory metrics",
			slog.Any("err", err),
		)

		return err
	}

	if err := c.collectCpu(ch); err != nil {
		logger.Error("failed collecting vmware cpu metrics",
			slog.Any("err", err),
		)

		return err
	}

	return nil
}

type Win32_PerfRawData_vmGuestLib_VMem struct {
	MemActiveMB      uint64 `mi:"MemActiveMB"`
	MemBalloonedMB   uint64 `mi:"MemBalloonedMB"`
	MemLimitMB       uint64 `mi:"MemLimitMB"`
	MemMappedMB      uint64 `mi:"MemMappedMB"`
	MemOverheadMB    uint64 `mi:"MemOverheadMB"`
	MemReservationMB uint64 `mi:"MemReservationMB"`
	MemSharedMB      uint64 `mi:"MemSharedMB"`
	MemSharedSavedMB uint64 `mi:"MemSharedSavedMB"`
	MemShares        uint64 `mi:"MemShares"`
	MemSwappedMB     uint64 `mi:"MemSwappedMB"`
	MemTargetSizeMB  uint64 `mi:"MemTargetSizeMB"`
	MemUsedMB        uint64 `mi:"MemUsedMB"`
}

type Win32_PerfRawData_vmGuestLib_VCPU struct {
	CpuLimitMHz           uint64 `mi:"CpuLimitMHz"`
	CpuReservationMHz     uint64 `mi:"CpuReservationMHz"`
	CpuShares             uint64 `mi:"CpuShares"`
	CpuStolenMs           uint64 `mi:"CpuStolenMs"`
	CpuTimePercents       uint64 `mi:"CpuTimePercents"`
	EffectiveVMSpeedMHz   uint64 `mi:"EffectiveVMSpeedMHz"`
	HostProcessorSpeedMHz uint64 `mi:"HostProcessorSpeedMHz"`
}

func (c *Collector) collectMem(ch chan<- prometheus.Metric) error {
	var dst []Win32_PerfRawData_vmGuestLib_VMem
	if err := c.miSession.Query(&dst, mi.NamespaceRootCIMv2, c.miQueryMem); err != nil {
		return fmt.Errorf("WMI query failed: %w", err)
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

// mbToBytes moved to utils package
func mbToBytes(mb uint64) float64 {
	return float64(mb * 1024 * 1024)
}

func (c *Collector) collectCpu(ch chan<- prometheus.Metric) error {
	var dst []Win32_PerfRawData_vmGuestLib_VCPU
	if err := c.miSession.Query(&dst, mi.NamespaceRootCIMv2, c.miQueryCPU); err != nil {
		return fmt.Errorf("WMI query failed: %w", err)
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
		float64(dst[0].CpuStolenMs)*perftypes.TicksToSecondScaleFactor,
	)

	ch <- prometheus.MustNewConstMetric(
		c.cpuTimeTotal,
		prometheus.CounterValue,
		float64(dst[0].CpuTimePercents)*perftypes.TicksToSecondScaleFactor,
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
