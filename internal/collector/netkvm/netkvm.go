// SPDX-License-Identifier: Apache-2.0
//
// Copyright The Prometheus Authors
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

package netkvm

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"time"

	"github.com/alecthomas/kingpin/v2"
	"github.com/prometheus-community/windows_exporter/internal/mi"
	"github.com/prometheus-community/windows_exporter/internal/types"
	"github.com/prometheus-community/windows_exporter/internal/utils"
	"github.com/prometheus/client_golang/prometheus"
)

const Name = "netkvm"

type Config struct{}

//nolint:gochecknoglobals
var ConfigDefaults = Config{}

type Collector struct {
	config Config
	logger *slog.Logger

	miSession   *mi.Session
	miQueryConf mi.Query
	miQueryDiag mi.Query

	// Config metrics
	info                     *prometheus.Desc
	queues                   *prometheus.Desc
	rxFreeBuffers            *prometheus.Desc
	txQueueSize              *prometheus.Desc
	memoryAllocatedBytes     *prometheus.Desc
	initDurationSeconds      *prometheus.Desc
	lazyAllocDurationSeconds *prometheus.Desc

	// Tx diagnostics
	txLargeOffloadTotal    *prometheus.Desc
	txUDPOffloadTotal      *prometheus.Desc
	txChecksumOffloadTotal *prometheus.Desc
	txCopiedTotal          *prometheus.Desc
	txDroppedTotal         *prometheus.Desc
	txMinFreeBuffers       *prometheus.Desc

	// Rx diagnostics
	rxCoalescedWindowsTotal *prometheus.Desc
	rxCoalescedHostTotal    *prometheus.Desc
	rxChecksumOKTotal       *prometheus.Desc
	rxPriorityTotal         *prometheus.Desc
	rxLowResourcesTotal     *prometheus.Desc
	rxMinFreeBuffers        *prometheus.Desc

	// RSS diagnostics
	rssDeviceSupported   *prometheus.Desc
	rssDeviceHashSupport *prometheus.Desc
	rssActive            *prometheus.Desc
	rssHitsTotal         *prometheus.Desc
	rssMissesTotal       *prometheus.Desc
	rssUnclassifiedTotal *prometheus.Desc
	rssErrorsTotal       *prometheus.Desc

	// Control diagnostics
	ctrlCommandsTotal         *prometheus.Desc
	ctrlCommandsTimedOutTotal *prometheus.Desc
	ctrlCommandsFailedTotal   *prometheus.Desc
}

// NetKvm_Config from netkvm.mof — fields match MOF sint32/uint32/boolean types.
// See: kvm-guest-drivers-windows/NetKVM/Common/netkvm.mof
type netkvmConfig struct {
	InstanceName string `mi:"InstanceName"`
	NumOfQueues  uint32 `mi:"NumOfQueues"`
	// WMI field is named RxQueueSize but the NetKVM driver populates it with
	// GetFreeRxBuffers() (ParaNdis6_Oid.cpp), not the queue capacity.
	RxFreeBuffers   uint32 `mi:"RxQueueSize"`
	TxQueueSize     uint32 `mi:"TxQueueSize"`
	RscEnabledv4    bool   `mi:"RscEnabledv4"`
	RscEnabledv6    bool   `mi:"RscEnabledv6"`
	Standby         bool   `mi:"Standby"`
	MemoryKB        uint32 `mi:"MemoryKB"`        // allocatedSharedMemory / 1024 (ULONG)
	InitTimeMs      int32  `mi:"InitTimeMs"`      // MOF: sint32; source: ULONG fastInitTime
	LazyAllocTimeMs int32  `mi:"LazyAllocTimeMs"` // MOF: sint32; source: LONG lazyAllocTime; negative = not completed
	UsoEnabledv4    int32  `mi:"UsoEnabledv4"`    // MOF: sint32; source: int fUsov4 : 1 (bitfield)
	UsoEnabledv6    int32  `mi:"UsoEnabledv6"`    // MOF: sint32; source: int fUsov6 : 1 (bitfield)
}

// NetKvm_Tx from netkvm.mof — all uint32 counters; they wrap at ~4 billion.
type netkvmTx struct {
	LargeOffload    uint32 `mi:"LargeOffload"`
	UdpOffload      uint32 `mi:"UdpOffload"`
	ChecksumOffload uint32 `mi:"ChecksumOffload"`
	MinFreeBuffers  uint32 `mi:"MinFreeBuffers"`
	Copied          uint32 `mi:"Copied"`
	Dropped         uint32 `mi:"Dropped"`
}

// NetKvm_Rx from netkvm.mof — all uint32 counters.
type netkvmRx struct {
	CoalescedWin   uint32 `mi:"CoalescedWin"`
	CoalescedHost  uint32 `mi:"CoalescedHost"`
	ChecksumOK     uint32 `mi:"ChecksumOK"`
	Priority       uint32 `mi:"Priority"`
	MinFreeBuffers uint32 `mi:"MinFreeBuffers"`
	LowResources   uint32 `mi:"LowResources"`
}

// NetKvm_Rss from netkvm.mof
type netkvmRss struct {
	DeviceRssSupport  bool   `mi:"DeviceRssSupport"`
	DeviceHashSupport bool   `mi:"DeviceHashSupport"`
	DeviceRssOn       bool   `mi:"DeviceRssOn"`
	Hits              uint32 `mi:"Hits"`
	Misses            uint32 `mi:"Misses"`
	Unclassified      uint32 `mi:"Unclassified"`
	Errors            uint32 `mi:"Errors"`
}

// NetKvm_Ctrl from netkvm.mof
type netkvmCtrl struct {
	Commands         uint32 `mi:"Commands"`
	CommandsTimedOut uint32 `mi:"CommandsTimedOut"`
	CommandsFailed   uint32 `mi:"CommandsFailed"`
}

// NetKvm_Diag from netkvm.mof — embeds Tx, Rx, Rss, Ctrl sub-instances.
type netkvmDiag struct {
	InstanceName string     `mi:"InstanceName"`
	Tx           netkvmTx   `mi:"tx"`
	Rx           netkvmRx   `mi:"rx"`
	Rss          netkvmRss  `mi:"rss"`
	Ctrl         netkvmCtrl `mi:"ctrl"`
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
	return &Collector{
		config: ConfigDefaults,
	}
}

func (c *Collector) GetName() string {
	return Name
}

func (c *Collector) Close() error {
	return nil
}

func (c *Collector) Build(logger *slog.Logger, miSession *mi.Session) error {
	c.logger = logger.With(slog.String("collector", Name))

	if miSession == nil {
		return errors.New("miSession is nil")
	}

	c.miSession = miSession

	var err error

	c.miQueryConf, err = mi.NewQuery("SELECT InstanceName, NumOfQueues, RxQueueSize, TxQueueSize, RscEnabledv4, RscEnabledv6, Standby, MemoryKB, InitTimeMs, LazyAllocTimeMs, UsoEnabledv4, UsoEnabledv6 FROM NetKvm_Config")
	if err != nil {
		return fmt.Errorf("failed to create NetKvm_Config query: %w", err)
	}

	c.miQueryDiag, err = mi.NewQuery("SELECT InstanceName, tx, rx, rss, ctrl FROM NetKvm_Diag")
	if err != nil {
		return fmt.Errorf("failed to create NetKvm_Diag query: %w", err)
	}

	// Config metrics
	c.info = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "info"),
		"NetKVM adapter information.",
		[]string{"adapter", "standby", "rsc_v4", "rsc_v6", "uso_v4", "uso_v6"},
		nil,
	)
	c.queues = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "queues"),
		"Number of virtio queues.",
		[]string{"adapter"},
		nil,
	)
	c.rxFreeBuffers = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "rx_free_buffers"),
		"Number of currently free receive buffers in the first receive queue.",
		[]string{"adapter"},
		nil,
	)
	c.txQueueSize = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "tx_queue_size"),
		"Transmit queue size.",
		[]string{"adapter"},
		nil,
	)
	c.memoryAllocatedBytes = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "memory_allocated_bytes"),
		"Allocated shared memory in bytes.",
		[]string{"adapter"},
		nil,
	)
	c.initDurationSeconds = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "init_duration_seconds"),
		"Adapter initialization time in seconds.",
		[]string{"adapter"},
		nil,
	)
	c.lazyAllocDurationSeconds = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "lazy_alloc_duration_seconds"),
		"Lazy memory allocation time in seconds (-1 if not completed).",
		[]string{"adapter"},
		nil,
	)

	// Tx diagnostics
	c.txLargeOffloadTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "tx_large_offload_total"),
		"Number of LSO offloaded transmit frames.",
		[]string{"adapter"},
		nil,
	)
	c.txUDPOffloadTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "tx_udp_offload_total"),
		"Number of USO offloaded transmit frames.",
		[]string{"adapter"},
		nil,
	)
	c.txChecksumOffloadTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "tx_checksum_offload_total"),
		"Number of checksum offloaded transmit frames.",
		[]string{"adapter"},
		nil,
	)
	c.txCopiedTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "tx_copied_total"),
		"Number of copied transmit packets.",
		[]string{"adapter"},
		nil,
	)
	c.txDroppedTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "tx_dropped_total"),
		"Number of dropped transmit packets.",
		[]string{"adapter"},
		nil,
	)
	c.txMinFreeBuffers = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "tx_min_free_buffers"),
		"Transmit minimum free buffer watermark.",
		[]string{"adapter"},
		nil,
	)

	// Rx diagnostics
	c.rxCoalescedWindowsTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "rx_coalesced_windows_total"),
		"Number of receive frames coalesced by Windows RSC.",
		[]string{"adapter"},
		nil,
	)
	c.rxCoalescedHostTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "rx_coalesced_host_total"),
		"Number of receive frames coalesced by the host.",
		[]string{"adapter"},
		nil,
	)
	c.rxChecksumOKTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "rx_checksum_ok_total"),
		"Number of receive frames with hardware-verified checksum.",
		[]string{"adapter"},
		nil,
	)
	c.rxPriorityTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "rx_priority_total"),
		"Number of priority-tagged receive frames.",
		[]string{"adapter"},
		nil,
	)
	c.rxLowResourcesTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "rx_low_resources_total"),
		"Number of receive indications with low-resource flag.",
		[]string{"adapter"},
		nil,
	)
	c.rxMinFreeBuffers = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "rx_min_free_buffers"),
		"Receive minimum free buffer watermark.",
		[]string{"adapter"},
		nil,
	)

	// RSS diagnostics
	c.rssDeviceSupported = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "rss_device_supported"),
		"Whether the device supports RSS (0 or 1).",
		[]string{"adapter"},
		nil,
	)
	c.rssDeviceHashSupport = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "rss_device_hash_supported"),
		"Whether the device reports hash (0 or 1).",
		[]string{"adapter"},
		nil,
	)
	c.rssActive = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "rss_active"),
		"Whether RSS is currently active (0 or 1).",
		[]string{"adapter"},
		nil,
	)
	c.rssHitsTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "rss_hits_total"),
		"Number of RSS hash hits.",
		[]string{"adapter"},
		nil,
	)
	c.rssMissesTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "rss_misses_total"),
		"Number of RSS hash misses.",
		[]string{"adapter"},
		nil,
	)
	c.rssUnclassifiedTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "rss_unclassified_total"),
		"Number of RSS unclassified packets.",
		[]string{"adapter"},
		nil,
	)
	c.rssErrorsTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "rss_errors_total"),
		"Number of RSS errors.",
		[]string{"adapter"},
		nil,
	)

	// Control diagnostics
	c.ctrlCommandsTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "ctrl_commands_total"),
		"Number of virtio control commands sent.",
		[]string{"adapter"},
		nil,
	)
	c.ctrlCommandsTimedOutTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "ctrl_commands_timed_out_total"),
		"Number of virtio control commands that timed out.",
		[]string{"adapter"},
		nil,
	)
	c.ctrlCommandsFailedTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "ctrl_commands_failed_total"),
		"Number of virtio control commands that failed.",
		[]string{"adapter"},
		nil,
	)

	var dstConf []netkvmConfig
	if err := c.miSession.Query(&dstConf, mi.NamespaceRootWMI, c.miQueryConf, 0); err != nil {
		return fmt.Errorf("failed to query NetKvm_Config: %w", err)
	}

	var dstDiag []netkvmDiag
	if err := c.miSession.Query(&dstDiag, mi.NamespaceRootWMI, c.miQueryDiag, 0); err != nil {
		return fmt.Errorf("failed to query NetKvm_Diag: %w", err)
	}

	return nil
}

func (c *Collector) Collect(ch chan<- prometheus.Metric, maxScrapeDuration time.Duration) error {
	startTime := time.Now()

	var dstConf []netkvmConfig

	if err := c.miSession.Query(&dstConf, mi.NamespaceRootWMI, c.miQueryConf, maxScrapeDuration); err != nil {
		return fmt.Errorf("failed to query NetKvm_Config: %w", err)
	}

	var dstDiag []netkvmDiag

	remaining := maxScrapeDuration
	if maxScrapeDuration > 0 {
		remaining -= time.Since(startTime)
		if remaining <= 0 {
			return fmt.Errorf("scrape budget exhausted before querying NetKvm_Diag: %w", context.DeadlineExceeded)
		}
	}

	if err := c.miSession.Query(&dstDiag, mi.NamespaceRootWMI, c.miQueryDiag, remaining); err != nil {
		return fmt.Errorf("failed to query NetKvm_Diag: %w", err)
	}

	for _, cfg := range dstConf {
		adapter := cfg.InstanceName

		ch <- prometheus.MustNewConstMetric(
			c.info,
			prometheus.GaugeValue,
			1.0,
			adapter,
			strconv.FormatBool(cfg.Standby),
			strconv.FormatBool(cfg.RscEnabledv4),
			strconv.FormatBool(cfg.RscEnabledv6),
			strconv.FormatBool(cfg.UsoEnabledv4 != 0),
			strconv.FormatBool(cfg.UsoEnabledv6 != 0),
		)
		ch <- prometheus.MustNewConstMetric(c.queues, prometheus.GaugeValue, float64(cfg.NumOfQueues), adapter)
		ch <- prometheus.MustNewConstMetric(c.rxFreeBuffers, prometheus.GaugeValue, float64(cfg.RxFreeBuffers), adapter)
		ch <- prometheus.MustNewConstMetric(c.txQueueSize, prometheus.GaugeValue, float64(cfg.TxQueueSize), adapter)
		ch <- prometheus.MustNewConstMetric(c.memoryAllocatedBytes, prometheus.GaugeValue, float64(cfg.MemoryKB)*1024, adapter)
		ch <- prometheus.MustNewConstMetric(c.initDurationSeconds, prometheus.GaugeValue, float64(cfg.InitTimeMs)/1000, adapter)
		lazyAllocSeconds := float64(cfg.LazyAllocTimeMs) / 1000
		if cfg.LazyAllocTimeMs < 0 {
			lazyAllocSeconds = -1
		}

		ch <- prometheus.MustNewConstMetric(c.lazyAllocDurationSeconds, prometheus.GaugeValue, lazyAllocSeconds, adapter)
	}

	for _, diag := range dstDiag {
		adapter := diag.InstanceName

		// Tx
		ch <- prometheus.MustNewConstMetric(c.txLargeOffloadTotal, prometheus.CounterValue, float64(diag.Tx.LargeOffload), adapter)
		ch <- prometheus.MustNewConstMetric(c.txUDPOffloadTotal, prometheus.CounterValue, float64(diag.Tx.UdpOffload), adapter)
		ch <- prometheus.MustNewConstMetric(c.txChecksumOffloadTotal, prometheus.CounterValue, float64(diag.Tx.ChecksumOffload), adapter)
		ch <- prometheus.MustNewConstMetric(c.txCopiedTotal, prometheus.CounterValue, float64(diag.Tx.Copied), adapter)
		ch <- prometheus.MustNewConstMetric(c.txDroppedTotal, prometheus.CounterValue, float64(diag.Tx.Dropped), adapter)
		ch <- prometheus.MustNewConstMetric(c.txMinFreeBuffers, prometheus.GaugeValue, float64(diag.Tx.MinFreeBuffers), adapter)

		// Rx
		ch <- prometheus.MustNewConstMetric(c.rxCoalescedWindowsTotal, prometheus.CounterValue, float64(diag.Rx.CoalescedWin), adapter)
		ch <- prometheus.MustNewConstMetric(c.rxCoalescedHostTotal, prometheus.CounterValue, float64(diag.Rx.CoalescedHost), adapter)
		ch <- prometheus.MustNewConstMetric(c.rxChecksumOKTotal, prometheus.CounterValue, float64(diag.Rx.ChecksumOK), adapter)
		ch <- prometheus.MustNewConstMetric(c.rxPriorityTotal, prometheus.CounterValue, float64(diag.Rx.Priority), adapter)
		ch <- prometheus.MustNewConstMetric(c.rxLowResourcesTotal, prometheus.CounterValue, float64(diag.Rx.LowResources), adapter)
		ch <- prometheus.MustNewConstMetric(c.rxMinFreeBuffers, prometheus.GaugeValue, float64(diag.Rx.MinFreeBuffers), adapter)

		// RSS
		ch <- prometheus.MustNewConstMetric(c.rssDeviceSupported, prometheus.GaugeValue, utils.BoolToFloat(diag.Rss.DeviceRssSupport), adapter)
		ch <- prometheus.MustNewConstMetric(c.rssDeviceHashSupport, prometheus.GaugeValue, utils.BoolToFloat(diag.Rss.DeviceHashSupport), adapter)
		ch <- prometheus.MustNewConstMetric(c.rssActive, prometheus.GaugeValue, utils.BoolToFloat(diag.Rss.DeviceRssOn), adapter)
		ch <- prometheus.MustNewConstMetric(c.rssHitsTotal, prometheus.CounterValue, float64(diag.Rss.Hits), adapter)
		ch <- prometheus.MustNewConstMetric(c.rssMissesTotal, prometheus.CounterValue, float64(diag.Rss.Misses), adapter)
		ch <- prometheus.MustNewConstMetric(c.rssUnclassifiedTotal, prometheus.CounterValue, float64(diag.Rss.Unclassified), adapter)
		ch <- prometheus.MustNewConstMetric(c.rssErrorsTotal, prometheus.CounterValue, float64(diag.Rss.Errors), adapter)

		// Control
		ch <- prometheus.MustNewConstMetric(c.ctrlCommandsTotal, prometheus.CounterValue, float64(diag.Ctrl.Commands), adapter)
		ch <- prometheus.MustNewConstMetric(c.ctrlCommandsTimedOutTotal, prometheus.CounterValue, float64(diag.Ctrl.CommandsTimedOut), adapter)
		ch <- prometheus.MustNewConstMetric(c.ctrlCommandsFailedTotal, prometheus.CounterValue, float64(diag.Ctrl.CommandsFailed), adapter)
	}

	return nil
}
