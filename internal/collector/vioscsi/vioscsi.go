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

package vioscsi

import (
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"time"

	"github.com/alecthomas/kingpin/v2"
	"github.com/prometheus-community/windows_exporter/internal/mi"
	"github.com/prometheus-community/windows_exporter/internal/types"
	"github.com/prometheus/client_golang/prometheus"
)

const Name = "vioscsi"

type Config struct{}

//nolint:gochecknoglobals
var ConfigDefaults = Config{}

type Collector struct {
	config Config
	logger *slog.Logger

	miSession *mi.Session
	miQuery   mi.Query

	info           *prometheus.Desc
	queueDepth     *prometheus.Desc
	queuesCount    *prometheus.Desc
	physicalBreaks *prometheus.Desc
	responseTime   *prometheus.Desc
}

// VioScsiExtendedInfoGuid from vioscsi.mof
// See: kvm-guest-drivers-windows/vioscsi/vioscsi.mof
type vioScsiExtendedInfo struct {
	InstanceName            string `mi:"InstanceName"`
	QueueDepth              uint32 `mi:"QueueDepth"`
	QueuesCount             uint8  `mi:"QueuesCount"` // MOF: uint8
	Indirect                bool   `mi:"Indirect"`
	EventIndex              bool   `mi:"EventIndex"`
	DpcRedirection          bool   `mi:"DpcRedirection"`
	ConcurrentChannels      bool   `mi:"ConcurrentChannels"`
	InterruptMsgRanges      bool   `mi:"InterruptMsgRanges"`
	CompletionDuringStartIo bool   `mi:"CompletionDuringStartIo"`
	RingPacked              bool   `mi:"RingPacked"`
	PhysicalBreaks          uint32 `mi:"PhysicalBreaks"`
	ResponseTime            uint32 `mi:"ResponseTime"` // resp_time threshold in milliseconds
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

	c.miQuery, err = mi.NewQuery("SELECT InstanceName, QueueDepth, QueuesCount, Indirect, EventIndex, DpcRedirection, ConcurrentChannels, InterruptMsgRanges, CompletionDuringStartIo, RingPacked, PhysicalBreaks, ResponseTime FROM VioScsiExtendedInfoGuid")
	if err != nil {
		return fmt.Errorf("failed to create VioScsiExtendedInfoGuid query: %w", err)
	}

	c.info = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "info"),
		"VirtIO SCSI adapter information.",
		[]string{"adapter", "indirect", "event_index", "dpc_redirection", "concurrent_channels", "interrupt_msg_ranges", "completion_during_start_io", "ring_packed"},
		nil,
	)
	c.queueDepth = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "queue_depth"),
		"VirtIO SCSI queue depth.",
		[]string{"adapter"},
		nil,
	)
	c.queuesCount = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "queues_count"),
		"Number of VirtIO SCSI queues.",
		[]string{"adapter"},
		nil,
	)
	c.physicalBreaks = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "physical_breaks"),
		"Maximum number of scatter-gather segments.",
		[]string{"adapter"},
		nil,
	)
	c.responseTime = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "response_time_threshold_seconds"),
		"VirtIO SCSI adapter response time tracing threshold in seconds.",
		[]string{"adapter"},
		nil,
	)

	var dst []vioScsiExtendedInfo
	if err := c.miSession.Query(&dst, mi.NamespaceRootWMI, c.miQuery, 0); err != nil {
		return fmt.Errorf("failed to query VioScsiExtendedInfoGuid: %w", err)
	}

	return nil
}

func (c *Collector) Collect(ch chan<- prometheus.Metric, maxScrapeDuration time.Duration) error {
	var dst []vioScsiExtendedInfo

	if err := c.miSession.Query(&dst, mi.NamespaceRootWMI, c.miQuery, maxScrapeDuration); err != nil {
		return fmt.Errorf("failed to query VioScsiExtendedInfoGuid: %w", err)
	}

	for _, data := range dst {
		adapter := data.InstanceName

		ch <- prometheus.MustNewConstMetric(
			c.info,
			prometheus.GaugeValue,
			1.0,
			adapter,
			strconv.FormatBool(data.Indirect),
			strconv.FormatBool(data.EventIndex),
			strconv.FormatBool(data.DpcRedirection),
			strconv.FormatBool(data.ConcurrentChannels),
			strconv.FormatBool(data.InterruptMsgRanges),
			strconv.FormatBool(data.CompletionDuringStartIo),
			strconv.FormatBool(data.RingPacked),
		)
		ch <- prometheus.MustNewConstMetric(c.queueDepth, prometheus.GaugeValue, float64(data.QueueDepth), adapter)
		ch <- prometheus.MustNewConstMetric(c.queuesCount, prometheus.GaugeValue, float64(data.QueuesCount), adapter)
		ch <- prometheus.MustNewConstMetric(c.physicalBreaks, prometheus.GaugeValue, float64(data.PhysicalBreaks), adapter)
		// ResponseTime is in milliseconds (vioscsi driver compares it against time_msec)
		ch <- prometheus.MustNewConstMetric(c.responseTime, prometheus.GaugeValue, float64(data.ResponseTime)/1e3, adapter)
	}

	return nil
}
