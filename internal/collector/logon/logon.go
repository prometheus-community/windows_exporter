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

package logon

import (
	"fmt"
	"log/slog"

	"github.com/alecthomas/kingpin/v2"
	"github.com/prometheus-community/windows_exporter/internal/headers/secur32"
	"github.com/prometheus-community/windows_exporter/internal/mi"
	"github.com/prometheus-community/windows_exporter/internal/types"
	"github.com/prometheus/client_golang/prometheus"
)

const Name = "logon"

type Config struct{}

//nolint:gochecknoglobals
var ConfigDefaults = Config{}

// A Collector is a Prometheus Collector for WMI metrics.
// Deprecated: Use windows_terminal_services_session_info instead.
type Collector struct {
	config Config

	sessionInfo *prometheus.Desc
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

func (c *Collector) Build(logger *slog.Logger, _ *mi.Session) error {
	logger.Warn("The logon collector will be removed mid 2025. "+
		"See https://github.com/prometheus-community/windows_exporter/pull/1957 for more information. If you see values in this collector"+
		" that you need, please open an issue to discuss how to get them into the new collector.",
		slog.String("collector", Name),
	)

	c.sessionInfo = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "session_logon_timestamp_seconds"),
		"Deprecated. Use windows_terminal_services_session_info instead.",
		[]string{"id", "username", "domain", "type"},
		nil,
	)

	return nil
}

// Collect sends the metric values for each metric
// to the provided prometheus Metric channel.
func (c *Collector) Collect(ch chan<- prometheus.Metric) error {
	logonSessions, err := secur32.GetLogonSessions()
	if err != nil {
		return fmt.Errorf("failed to get logon sessions: %w", err)
	}

	for _, session := range logonSessions {
		ch <- prometheus.MustNewConstMetric(
			c.sessionInfo,
			prometheus.GaugeValue,
			float64(session.LogonTime.Unix()),
			session.LogonId.String(), session.UserName, session.LogonDomain, session.LogonType.String(),
		)
	}

	return nil
}
