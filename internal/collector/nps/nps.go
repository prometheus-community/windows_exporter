// Copyright 2024 The Prometheus Authors
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

package nps

import (
	"errors"
	"fmt"
	"log/slog"

	"github.com/alecthomas/kingpin/v2"
	"github.com/prometheus-community/windows_exporter/internal/mi"
	"github.com/prometheus-community/windows_exporter/internal/pdh"
	"github.com/prometheus-community/windows_exporter/internal/types"
	"github.com/prometheus/client_golang/prometheus"
)

const Name = "nps"

type Config struct{}

//nolint:gochecknoglobals
var ConfigDefaults = Config{}

type Collector struct {
	config Config

	accessPerfDataCollector *pdh.Collector
	accessPerfDataObject    []perfDataCounterValuesAccess
	accessAccepts           *prometheus.Desc
	accessChallenges        *prometheus.Desc
	accessRejects           *prometheus.Desc
	accessRequests          *prometheus.Desc
	accessBadAuthenticators *prometheus.Desc
	accessDroppedPackets    *prometheus.Desc
	accessInvalidRequests   *prometheus.Desc
	accessMalformedPackets  *prometheus.Desc
	accessPacketsReceived   *prometheus.Desc
	accessPacketsSent       *prometheus.Desc
	accessServerResetTime   *prometheus.Desc
	accessServerUpTime      *prometheus.Desc
	accessUnknownType       *prometheus.Desc

	accountingPerfDataCollector *pdh.Collector
	accountingPerfDataObject    []perfDataCounterValuesAccounting
	accountingRequests          *prometheus.Desc
	accountingResponses         *prometheus.Desc
	accountingBadAuthenticators *prometheus.Desc
	accountingDroppedPackets    *prometheus.Desc
	accountingInvalidRequests   *prometheus.Desc
	accountingMalformedPackets  *prometheus.Desc
	accountingNoRecord          *prometheus.Desc
	accountingPacketsReceived   *prometheus.Desc
	accountingPacketsSent       *prometheus.Desc
	accountingServerResetTime   *prometheus.Desc
	accountingServerUpTime      *prometheus.Desc
	accountingUnknownType       *prometheus.Desc
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

func (c *Collector) Close() error {
	return nil
}

func (c *Collector) Build(_ *slog.Logger, _ *mi.Session) error {
	c.accessAccepts = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "access_accepts"),
		"(AccessAccepts)",
		nil,
		nil,
	)
	c.accessChallenges = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "access_challenges"),
		"(AccessChallenges)",
		nil,
		nil,
	)
	c.accessRejects = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "access_rejects"),
		"(AccessRejects)",
		nil,
		nil,
	)
	c.accessRequests = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "access_requests"),
		"(AccessRequests)",
		nil,
		nil,
	)
	c.accessBadAuthenticators = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "access_bad_authenticators"),
		"(BadAuthenticators)",
		nil,
		nil,
	)
	c.accessDroppedPackets = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "access_dropped_packets"),
		"(DroppedPackets)",
		nil,
		nil,
	)
	c.accessInvalidRequests = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "access_invalid_requests"),
		"(InvalidRequests)",
		nil,
		nil,
	)
	c.accessMalformedPackets = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "access_malformed_packets"),
		"(MalformedPackets)",
		nil,
		nil,
	)
	c.accessPacketsReceived = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "access_packets_received"),
		"(PacketsReceived)",
		nil,
		nil,
	)
	c.accessPacketsSent = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "access_packets_sent"),
		"(PacketsSent)",
		nil,
		nil,
	)
	c.accessServerResetTime = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "access_server_reset_time"),
		"(ServerResetTime)",
		nil,
		nil,
	)
	c.accessServerUpTime = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "access_server_up_time"),
		"(ServerUpTime)",
		nil,
		nil,
	)
	c.accessUnknownType = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "access_unknown_type"),
		"(UnknownType)",
		nil,
		nil,
	)

	c.accountingRequests = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "accounting_requests"),
		"(AccountingRequests)",
		nil,
		nil,
	)
	c.accountingResponses = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "accounting_responses"),
		"(AccountingResponses)",
		nil,
		nil,
	)
	c.accountingBadAuthenticators = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "accounting_bad_authenticators"),
		"(BadAuthenticators)",
		nil,
		nil,
	)
	c.accountingDroppedPackets = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "accounting_dropped_packets"),
		"(DroppedPackets)",
		nil,
		nil,
	)
	c.accountingInvalidRequests = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "accounting_invalid_requests"),
		"(InvalidRequests)",
		nil,
		nil,
	)
	c.accountingMalformedPackets = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "accounting_malformed_packets"),
		"(MalformedPackets)",
		nil,
		nil,
	)
	c.accountingNoRecord = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "accounting_no_record"),
		"(NoRecord)",
		nil,
		nil,
	)
	c.accountingPacketsReceived = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "accounting_packets_received"),
		"(PacketsReceived)",
		nil,
		nil,
	)
	c.accountingPacketsSent = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "accounting_packets_sent"),
		"(PacketsSent)",
		nil,
		nil,
	)
	c.accountingServerResetTime = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "accounting_server_reset_time"),
		"(ServerResetTime)",
		nil,
		nil,
	)
	c.accountingServerUpTime = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "accounting_server_up_time"),
		"(ServerUpTime)",
		nil,
		nil,
	)
	c.accountingUnknownType = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "accounting_unknown_type"),
		"(UnknownType)",
		nil,
		nil,
	)

	var err error

	errs := make([]error, 0)

	c.accessPerfDataCollector, err = pdh.NewCollector[perfDataCounterValuesAccess](pdh.CounterTypeRaw, "NPS Authentication Server", nil)
	if err != nil {
		errs = append(errs, fmt.Errorf("failed to create NPS Authentication Server collector: %w", err))
	}

	c.accountingPerfDataCollector, err = pdh.NewCollector[perfDataCounterValuesAccounting](pdh.CounterTypeRaw, "NPS Accounting Server", nil)
	if err != nil {
		errs = append(errs, fmt.Errorf("failed to create NPS Accounting Server collector: %w", err))
	}

	return errors.Join(errs...)
}

// Collect sends the metric values for each metric
// to the provided prometheus Metric channel.
func (c *Collector) Collect(ch chan<- prometheus.Metric) error {
	errs := make([]error, 0)

	if err := c.collectAccept(ch); err != nil {
		errs = append(errs, fmt.Errorf("failed collecting NPS accept data: %w", err))
	}

	if err := c.collectAccounting(ch); err != nil {
		errs = append(errs, fmt.Errorf("failed collecting NPS accounting data: %w", err))
	}

	return errors.Join(errs...)
}

// collectAccept sends the metric values for each metric
// to the provided prometheus Metric channel.
func (c *Collector) collectAccept(ch chan<- prometheus.Metric) error {
	err := c.accessPerfDataCollector.Collect(&c.accessPerfDataObject)
	if err != nil {
		return fmt.Errorf("failed to collect NPS Authentication Server metrics: %w", err)
	}

	ch <- prometheus.MustNewConstMetric(
		c.accessAccepts,
		prometheus.CounterValue,
		c.accessPerfDataObject[0].AccessAccepts,
	)

	ch <- prometheus.MustNewConstMetric(
		c.accessChallenges,
		prometheus.CounterValue,
		c.accessPerfDataObject[0].AccessChallenges,
	)

	ch <- prometheus.MustNewConstMetric(
		c.accessRejects,
		prometheus.CounterValue,
		c.accessPerfDataObject[0].AccessRejects,
	)

	ch <- prometheus.MustNewConstMetric(
		c.accessRequests,
		prometheus.CounterValue,
		c.accessPerfDataObject[0].AccessRequests,
	)

	ch <- prometheus.MustNewConstMetric(
		c.accessBadAuthenticators,
		prometheus.CounterValue,
		c.accessPerfDataObject[0].AccessBadAuthenticators,
	)

	ch <- prometheus.MustNewConstMetric(
		c.accessDroppedPackets,
		prometheus.CounterValue,
		c.accessPerfDataObject[0].AccessDroppedPackets,
	)

	ch <- prometheus.MustNewConstMetric(
		c.accessInvalidRequests,
		prometheus.CounterValue,
		c.accessPerfDataObject[0].AccessInvalidRequests,
	)

	ch <- prometheus.MustNewConstMetric(
		c.accessMalformedPackets,
		prometheus.CounterValue,
		c.accessPerfDataObject[0].AccessMalformedPackets,
	)

	ch <- prometheus.MustNewConstMetric(
		c.accessPacketsReceived,
		prometheus.CounterValue,
		c.accessPerfDataObject[0].AccessPacketsReceived,
	)

	ch <- prometheus.MustNewConstMetric(
		c.accessPacketsSent,
		prometheus.CounterValue,
		c.accessPerfDataObject[0].AccessPacketsSent,
	)

	ch <- prometheus.MustNewConstMetric(
		c.accessServerResetTime,
		prometheus.CounterValue,
		c.accessPerfDataObject[0].AccessServerResetTime,
	)

	ch <- prometheus.MustNewConstMetric(
		c.accessServerUpTime,
		prometheus.CounterValue,
		c.accessPerfDataObject[0].AccessServerUpTime,
	)

	ch <- prometheus.MustNewConstMetric(
		c.accessUnknownType,
		prometheus.CounterValue,
		c.accessPerfDataObject[0].AccessUnknownType,
	)

	return nil
}

func (c *Collector) collectAccounting(ch chan<- prometheus.Metric) error {
	err := c.accountingPerfDataCollector.Collect(&c.accountingPerfDataObject)
	if err != nil {
		return fmt.Errorf("failed to collect NPS Accounting Server metrics: %w", err)
	}

	ch <- prometheus.MustNewConstMetric(
		c.accountingRequests,
		prometheus.CounterValue,
		c.accountingPerfDataObject[0].AccountingRequests,
	)

	ch <- prometheus.MustNewConstMetric(
		c.accountingResponses,
		prometheus.CounterValue,
		c.accountingPerfDataObject[0].AccountingResponses,
	)

	ch <- prometheus.MustNewConstMetric(
		c.accountingBadAuthenticators,
		prometheus.CounterValue,
		c.accountingPerfDataObject[0].AccountingBadAuthenticators,
	)

	ch <- prometheus.MustNewConstMetric(
		c.accountingDroppedPackets,
		prometheus.CounterValue,
		c.accountingPerfDataObject[0].AccountingDroppedPackets,
	)

	ch <- prometheus.MustNewConstMetric(
		c.accountingInvalidRequests,
		prometheus.CounterValue,
		c.accountingPerfDataObject[0].AccountingInvalidRequests,
	)

	ch <- prometheus.MustNewConstMetric(
		c.accountingMalformedPackets,
		prometheus.CounterValue,
		c.accountingPerfDataObject[0].AccountingMalformedPackets,
	)

	ch <- prometheus.MustNewConstMetric(
		c.accountingNoRecord,
		prometheus.CounterValue,
		c.accountingPerfDataObject[0].AccountingNoRecord,
	)

	ch <- prometheus.MustNewConstMetric(
		c.accountingPacketsReceived,
		prometheus.CounterValue,
		c.accountingPerfDataObject[0].AccountingPacketsReceived,
	)

	ch <- prometheus.MustNewConstMetric(
		c.accountingPacketsSent,
		prometheus.CounterValue,
		c.accountingPerfDataObject[0].AccountingPacketsSent,
	)

	ch <- prometheus.MustNewConstMetric(
		c.accountingServerResetTime,
		prometheus.CounterValue,
		c.accountingPerfDataObject[0].AccountingServerResetTime,
	)

	ch <- prometheus.MustNewConstMetric(
		c.accountingServerUpTime,
		prometheus.CounterValue,
		c.accountingPerfDataObject[0].AccountingServerUpTime,
	)

	ch <- prometheus.MustNewConstMetric(
		c.accountingUnknownType,
		prometheus.CounterValue,
		c.accountingPerfDataObject[0].AccountingUnknownType,
	)

	return nil
}
