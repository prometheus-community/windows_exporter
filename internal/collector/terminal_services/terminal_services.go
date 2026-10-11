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

package terminal_services

import (
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/alecthomas/kingpin/v2"
	"github.com/prometheus-community/windows_exporter/internal/headers/wtsapi32"
	"github.com/prometheus-community/windows_exporter/internal/labelcache"
	"github.com/prometheus-community/windows_exporter/internal/metriccache"
	"github.com/prometheus-community/windows_exporter/internal/mi"
	"github.com/prometheus-community/windows_exporter/internal/pdh"
	"github.com/prometheus-community/windows_exporter/internal/types"
	"github.com/prometheus/client_golang/prometheus"
	"golang.org/x/sys/windows"
)

const (
	Name = "terminal_services"

	// connectionBrokerServiceName is the "Remote Desktop Connection Broker"
	// service, installed by the RDS-Connection-Broker role.
	connectionBrokerServiceName = "Tssdis"
)

type Config struct{}

//nolint:gochecknoglobals
var ConfigDefaults = Config{}

// isConnectionBrokerServer reports whether the Remote Desktop Connection Broker
// role is installed by looking up its service in the service control manager.
func isConnectionBrokerServer() bool {
	scm, err := windows.OpenSCManager(nil, nil, windows.SC_MANAGER_CONNECT)
	if err != nil {
		return false
	}

	defer windows.CloseServiceHandle(scm) //nolint:errcheck

	service, err := windows.OpenService(scm, windows.StringToUTF16Ptr(connectionBrokerServiceName), windows.SERVICE_QUERY_STATUS)
	if err != nil {
		return false
	}

	_ = windows.CloseServiceHandle(service)

	return true
}

// A Collector is a Prometheus Collector for WMI
// Win32_PerfRawData_LocalSessionManager_TerminalServices &  Win32_PerfRawData_TermService_TerminalServicesSession  metrics
// https://docs.microsoft.com/en-us/previous-versions/aa394344(v%3Dvs.85)
// https://wutils.com/wmi/root/cimv2/win32_perfrawdata_localsessionmanager_terminalservices/
type Collector struct {
	config Config

	logger *slog.Logger

	connectionBrokerEnabled bool

	// perfDataQuery samples the session and Connection Broker objects together,
	// so the collector waits only once for the other performance counter queries.
	perfDataQuery                            *pdh.Query
	perfDataCollectorTerminalServicesSession *pdh.Collector[perfDataCounterValuesTerminalServicesSession]
	perfDataCollectorBroker                  *pdh.Collector[perfDataCounterValuesBroker]

	perfDataObjectTerminalServicesSession []perfDataCounterValuesTerminalServicesSession
	perfDataObjectBroker                  []perfDataCounterValuesBroker

	hServer windows.Handle

	// sessionCache reuses the session_info state series of sessions whose
	// name, user, host and state did not change.
	sessionCache metriccache.Cache[uint32, sessionValues]
	// labels reuses the label pairs of the per-session counter series.
	labels labelcache.Cache

	sessionInfo                        *prometheus.Desc
	connectionBrokerPerformance        *prometheus.Desc
	connectionBrokerPendingConnections *prometheus.Desc
	handleCount                        *prometheus.Desc
	pageFaultsPerSec                   *prometheus.Desc
	pageFileBytes                      *prometheus.Desc
	pageFileBytesPeak                  *prometheus.Desc
	percentCPUTime                     *prometheus.Desc
	poolNonPagedBytes                  *prometheus.Desc
	poolPagedBytes                     *prometheus.Desc
	privateBytes                       *prometheus.Desc
	threadCount                        *prometheus.Desc
	virtualBytes                       *prometheus.Desc
	virtualBytesPeak                   *prometheus.Desc
	workingSet                         *prometheus.Desc
	workingSetPeak                     *prometheus.Desc
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
	c.perfDataCollectorTerminalServicesSession.Close()
	c.perfDataCollectorBroker.Close()
	c.perfDataQuery.Close()

	c.sessionCache.Reset()
	c.labels.Reset()

	// hServer is 0 if Build failed before opening the server.
	if c.hServer == 0 || c.hServer == windows.InvalidHandle {
		return nil
	}

	wtsapi32.WTSCloseServer(c.hServer)
	c.hServer = 0

	return nil
}

func (c *Collector) Build(logger *slog.Logger, miSession *mi.Session) error {
	c.logger = logger.With(slog.String("collector", Name))

	c.sessionCache.Reset()
	c.labels.Reset()

	c.sessionInfo = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "session_info"),
		"Terminal Services sessions info",
		[]string{"session_name", "user", "host", "state", "id"},
		nil,
	)
	c.connectionBrokerPerformance = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "connection_broker_performance_total"),
		"The number of connections handled by the Connection Broker since the service started, by result. connection=\"Pending\" is deprecated: it is the current number of pending connections, use windows_terminal_services_connection_broker_pending_connections instead.",
		[]string{"connection"},
		nil,
	)
	c.connectionBrokerPendingConnections = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "connection_broker_pending_connections"),
		"The number of redirected connections pending completion from the Remote Desktop Session Host servers in the farm.",
		nil,
		nil,
	)
	c.handleCount = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "handles"),
		"Total number of handles currently opened by this process. This number is the sum of the handles currently opened by each thread in this process.",
		[]string{"session_name"},
		nil,
	)
	c.pageFaultsPerSec = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "page_fault_total"),
		"Rate at which page faults occur in the threads executing in this process. A page fault occurs when a thread refers to a virtual memory page that is not in its working set in main memory. The page may not be retrieved from disk if it is on the standby list and therefore already in main memory. The page also may not be retrieved if it is in use by another process which shares the page.",
		[]string{"session_name"},
		nil,
	)
	c.pageFileBytes = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "page_file_bytes"),
		"Current number of bytes this process has used in the paging file(s). Paging files are used to store pages of memory used by the process that are not contained in other files. Paging files are shared by all processes, and lack of space in paging files can prevent other processes from allocating memory.",
		[]string{"session_name"},
		nil,
	)
	c.pageFileBytesPeak = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "page_file_bytes_peak"),
		"Maximum number of bytes this process has used in the paging file(s). Paging files are used to store pages of memory used by the process that are not contained in other files. Paging files are shared by all processes, and lack of space in paging files can prevent other processes from allocating memory.",
		[]string{"session_name"},
		nil,
	)
	c.percentCPUTime = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "cpu_time_seconds_total"),
		"Total elapsed time that this process's threads have spent executing code.",
		[]string{"mode", "session_name"},
		nil,
	)
	c.poolNonPagedBytes = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "pool_non_paged_bytes"),
		"Number of bytes in the non-paged pool, an area of system memory (physical memory used by the operating system) for objects that cannot be written to disk, but must remain in physical memory as long as they are allocated. This property displays the last observed value only; it is not an average.",
		[]string{"session_name"},
		nil,
	)
	c.poolPagedBytes = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "pool_paged_bytes"),
		"Number of bytes in the paged pool, an area of system memory (physical memory used by the operating system) for objects that can be written to disk when they are not being used. This property displays the last observed value only; it is not an average.",
		[]string{"session_name"},
		nil,
	)
	c.privateBytes = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "private_bytes"),
		"Current number of bytes this process has allocated that cannot be shared with other processes.",
		[]string{"session_name"},
		nil,
	)
	c.threadCount = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "threads"),
		"Number of threads currently active in this process. An instruction is the basic unit of execution in a processor, and a thread is the object that executes instructions. Every running process has at least one thread.",
		[]string{"session_name"},
		nil,
	)
	c.virtualBytes = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "virtual_bytes"),
		"Current size, in bytes, of the virtual address space the process is using. Use of virtual address space does not necessarily imply corresponding use of either disk or main memory pages. Virtual space is finite and, by using too much, the process can limit its ability to load libraries.",
		[]string{"session_name"},
		nil,
	)
	c.virtualBytesPeak = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "virtual_bytes_peak"),
		"Maximum number of bytes of virtual address space the process has used at any one time. Use of virtual address space does not necessarily imply corresponding use of either disk or main memory pages. Virtual space is finite and, by using too much, the process might limit its ability to load libraries.",
		[]string{"session_name"},
		nil,
	)
	c.workingSet = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "working_set_bytes"),
		"Current number of bytes in the working set of this process. The working set is the set of memory pages touched recently by the threads in the process. If free memory in the computer is above a threshold, pages are left in the working set of a process even if they are not in use. When free memory falls below a threshold, pages are trimmed from working sets. If they are needed, they are then soft-faulted back into the working set before they leave main memory.",
		[]string{"session_name"},
		nil,
	)
	c.workingSetPeak = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "working_set_bytes_peak"),
		"Maximum number of bytes in the working set of this process at any point in time. The working set is the set of memory pages touched recently by the threads in the process. If free memory in the computer is above a threshold, pages are left in the working set of a process even if they are not in use. When free memory falls below a threshold, pages are trimmed from working sets. If they are needed, they are then soft-faulted back into the working set before they leave main memory.",
		[]string{"session_name"},
		nil,
	)

	var err error

	c.perfDataQuery, err = pdh.NewQuery()
	if err != nil {
		return fmt.Errorf("failed to open Terminal Services performance counter query: %w", err)
	}

	c.connectionBrokerEnabled = isConnectionBrokerServer()

	if c.connectionBrokerEnabled {
		c.perfDataCollectorBroker, err = pdh.NewQueryCollector[perfDataCounterValuesBroker](c.logger, c.perfDataQuery, pdh.CounterTypeRaw, "Remote Desktop Connection Broker Counterset", pdh.InstancesAll)
		if err != nil {
			// The service can be installed while its counters are not registered or
			// unreadable. Keep the session metrics and skip only the broker metrics.
			c.logger.Warn("failed to create Remote Desktop Connection Broker Counterset collector, skipping Connection Broker performance metrics",
				slog.Any("err", err),
			)

			c.perfDataCollectorBroker.Close()
			c.perfDataCollectorBroker = nil
			c.connectionBrokerEnabled = false
		}
	} else {
		c.logger.Debug("host is not a connection broker skipping Connection Broker performance metrics.")
	}

	c.hServer, err = wtsapi32.WTSOpenServer("")
	if err != nil {
		return fmt.Errorf("failed to open WTS server: %w", err)
	}

	c.perfDataCollectorTerminalServicesSession, err = pdh.NewQueryCollector[perfDataCounterValuesTerminalServicesSession](c.logger, c.perfDataQuery, pdh.CounterTypeRaw, "Terminal Services Session", pdh.InstancesAll)
	if err != nil {
		return fmt.Errorf("failed to create Terminal Services Session collector: %w", err)
	}

	return nil
}

// Collect sends the metric values for each metric
// to the provided prometheus Metric channel.
func (c *Collector) Collect(ch chan<- prometheus.Metric, _ time.Duration) error {
	errs := make([]error, 0)

	if err := c.collectWTSSessions(ch); err != nil {
		errs = append(errs, fmt.Errorf("failed collecting terminal services session infos: %w", err))
	}

	// The session and Connection Broker counters read this sample. Unlike
	// netframework, a failed sample, including pdh.ErrNoData, skips both, so
	// they don't publish the values of an earlier sample again.
	if err := c.perfDataQuery.Collect(); err != nil {
		errs = append(errs, fmt.Errorf("failed to sample session and Connection Broker counters: %w", err))

		return errors.Join(errs...)
	}

	if err := c.collectTSSessionCounters(ch); err != nil {
		errs = append(errs, fmt.Errorf("failed collecting terminal services session count metrics: %w", err))
	}

	// only collect CollectionBrokerPerformance if host is a Connection Broker
	if c.connectionBrokerEnabled {
		if err := c.collectCollectionBrokerPerformanceCounter(ch); err != nil {
			errs = append(errs, fmt.Errorf("failed collecting Connection Broker performance metrics: %w", err))
		}
	}

	return errors.Join(errs...)
}

func (c *Collector) collectTSSessionCounters(ch chan<- prometheus.Metric) error {
	err := c.perfDataCollectorTerminalServicesSession.Collect(&c.perfDataObjectTerminalServicesSession)
	if err != nil {
		return fmt.Errorf("failed to collect Terminal Services Session metrics: %w", err)
	}

	scrape := c.labels.Begin()
	defer scrape.Commit()

	names := make(map[string]bool)

	for _, data := range c.perfDataObjectTerminalServicesSession {
		// only connect metrics for remote named sessions
		n := strings.ToLower(data.Name)
		if n == "" || n == "services" || n == "console" {
			continue
		}
		// don't add name already present in labels list
		if _, ok := names[n]; ok {
			continue
		}

		names[n] = true

		ch <- scrape.MustNewConstMetric(
			c.handleCount,
			prometheus.GaugeValue,
			data.HandleCount,
			data.Name,
		)

		ch <- scrape.MustNewConstMetric(
			c.pageFaultsPerSec,
			prometheus.CounterValue,
			data.PageFaultsPersec,
			data.Name,
		)

		ch <- scrape.MustNewConstMetric(
			c.pageFileBytes,
			prometheus.GaugeValue,
			data.PageFileBytes,
			data.Name,
		)

		ch <- scrape.MustNewConstMetric(
			c.pageFileBytesPeak,
			prometheus.GaugeValue,
			data.PageFileBytesPeak,
			data.Name,
		)

		ch <- scrape.MustNewConstMetric(
			c.percentCPUTime,
			prometheus.CounterValue,
			data.PercentPrivilegedTime,
			"privileged",
			data.Name,
		)

		ch <- scrape.MustNewConstMetric(
			c.percentCPUTime,
			prometheus.CounterValue,
			data.PercentProcessorTime,
			"processor",
			data.Name,
		)

		ch <- scrape.MustNewConstMetric(
			c.percentCPUTime,
			prometheus.CounterValue,
			data.PercentUserTime,
			"user",
			data.Name,
		)

		ch <- scrape.MustNewConstMetric(
			c.poolNonPagedBytes,
			prometheus.GaugeValue,
			data.PoolNonpagedBytes,
			data.Name,
		)

		ch <- scrape.MustNewConstMetric(
			c.poolPagedBytes,
			prometheus.GaugeValue,
			data.PoolPagedBytes,
			data.Name,
		)

		ch <- scrape.MustNewConstMetric(
			c.privateBytes,
			prometheus.GaugeValue,
			data.PrivateBytes,
			data.Name,
		)

		ch <- scrape.MustNewConstMetric(
			c.threadCount,
			prometheus.GaugeValue,
			data.ThreadCount,
			data.Name,
		)

		ch <- scrape.MustNewConstMetric(
			c.virtualBytes,
			prometheus.GaugeValue,
			data.VirtualBytes,
			data.Name,
		)

		ch <- scrape.MustNewConstMetric(
			c.virtualBytesPeak,
			prometheus.GaugeValue,
			data.VirtualBytesPeak,
			data.Name,
		)

		ch <- scrape.MustNewConstMetric(
			c.workingSet,
			prometheus.GaugeValue,
			data.WorkingSet,
			data.Name,
		)

		ch <- scrape.MustNewConstMetric(
			c.workingSetPeak,
			prometheus.GaugeValue,
			data.WorkingSetPeak,
			data.Name,
		)
	}

	return nil
}

func (c *Collector) collectCollectionBrokerPerformanceCounter(ch chan<- prometheus.Metric) error {
	err := c.perfDataCollectorBroker.Collect(&c.perfDataObjectBroker)
	if err != nil {
		return fmt.Errorf("failed to collect Remote Desktop Connection Broker Counterset metrics: %w", err)
	}

	ch <- prometheus.MustNewConstMetric(
		c.connectionBrokerPerformance,
		prometheus.CounterValue,
		c.perfDataObjectBroker[0].SuccessfulConnections,
		"Successful",
	)

	ch <- prometheus.MustNewConstMetric(
		c.connectionBrokerPendingConnections,
		prometheus.GaugeValue,
		c.perfDataObjectBroker[0].PendingConnections,
	)

	// Deprecated: Pending Connections is a current count, not a counter.
	ch <- prometheus.MustNewConstMetric(
		c.connectionBrokerPerformance,
		prometheus.CounterValue,
		c.perfDataObjectBroker[0].PendingConnections,
		"Pending",
	)

	ch <- prometheus.MustNewConstMetric(
		c.connectionBrokerPerformance,
		prometheus.CounterValue,
		c.perfDataObjectBroker[0].FailedConnections,
		"Failed",
	)

	return nil
}

func (c *Collector) collectWTSSessions(ch chan<- prometheus.Metric) error {
	sessions, err := wtsapi32.WTSEnumerateSessionsEx(c.hServer, c.logger)
	if err != nil {
		return fmt.Errorf("failed to enumerate WTS sessions: %w", err)
	}

	scrape := c.sessionCache.Begin()
	// Sessions that were not collected in this scrape are dropped from the cache.
	defer scrape.Commit()

	for _, session := range sessions {
		// only connect metrics for remote named sessions
		n := strings.ReplaceAll(session.SessionName, "#", " ")
		if n == "Services" {
			continue
		}

		userName := session.UserName
		if session.DomainName != "" {
			userName = session.DomainName + "\\" + session.UserName
		}

		values := sessionValues{
			name:  n,
			user:  userName,
			host:  session.HostName,
			state: session.State,
		}

		for _, metric := range scrape.Get(session.SessionID, values, func() []prometheus.Metric {
			return c.sessionInfoMetrics(session.SessionID, values)
		}) {
			ch <- metric
		}
	}

	return nil
}

// sessionValues are the inputs of the session_info series of a session.
type sessionValues struct {
	name  string
	user  string
	host  string
	state wtsapi32.WTSConnectState
}

// sessionInfoMetrics builds the session_info state series of a session.
func (c *Collector) sessionInfoMetrics(sessionID uint32, values sessionValues) []prometheus.Metric {
	id := strconv.FormatUint(uint64(sessionID), 10)
	metrics := make([]prometheus.Metric, 0, len(wtsapi32.WTSSessionStates))

	for stateID, stateName := range wtsapi32.WTSSessionStates {
		isState := 0.0
		if values.state == stateID {
			isState = 1.0
		}

		metrics = append(metrics, prometheus.MustNewConstMetric(
			c.sessionInfo,
			prometheus.GaugeValue,
			isState,
			values.name,
			values.user,
			values.host,
			stateName,
			id,
		))
	}

	return metrics
}
