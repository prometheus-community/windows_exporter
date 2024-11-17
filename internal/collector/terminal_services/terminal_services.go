//go:build windows

package terminal_services

import (
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"strings"

	"github.com/alecthomas/kingpin/v2"
	"github.com/prometheus-community/windows_exporter/internal/headers/wtsapi32"
	"github.com/prometheus-community/windows_exporter/internal/mi"
	"github.com/prometheus-community/windows_exporter/internal/perfdata"
	"github.com/prometheus-community/windows_exporter/internal/types"
	"github.com/prometheus-community/windows_exporter/internal/utils"
	"github.com/prometheus/client_golang/prometheus"
	"golang.org/x/sys/windows"
)

const (
	Name                             = "terminal_services"
	ConnectionBrokerFeatureID uint32 = 133
)

type Config struct{}

var ConfigDefaults = Config{}

type Win32_ServerFeature struct {
	ID uint32
}

func isConnectionBrokerServer(miSession *mi.Session) bool {
	var dst []Win32_ServerFeature
	if err := miSession.Query(&dst, mi.NamespaceRootCIMv2, utils.Must(mi.NewQuery("SELECT * FROM Win32_ServerFeature"))); err != nil {
		return false
	}

	for _, d := range dst {
		if d.ID == ConnectionBrokerFeatureID {
			return true
		}
	}

	return false
}

// A Collector is a Prometheus Collector for WMI
// Win32_PerfRawData_LocalSessionManager_TerminalServices &  Win32_PerfRawData_TermService_TerminalServicesSession  metrics
// https://docs.microsoft.com/en-us/previous-versions/aa394344(v%3Dvs.85)
// https://wutils.com/wmi/root/cimv2/win32_perfrawdata_localsessionmanager_terminalservices/
type Collector struct {
	config Config

	logger *slog.Logger

	connectionBrokerEnabled bool

	perfDataCollectorTerminalServicesSession *perfdata.Collector
	perfDataCollectorBroker                  *perfdata.Collector

	hServer windows.Handle

	sessionInfo                 *prometheus.Desc
	connectionBrokerPerformance *prometheus.Desc
	handleCount                 *prometheus.Desc
	pageFaultsPerSec            *prometheus.Desc
	pageFileBytes               *prometheus.Desc
	pageFileBytesPeak           *prometheus.Desc
	percentCPUTime              *prometheus.Desc
	poolNonPagedBytes           *prometheus.Desc
	poolPagedBytes              *prometheus.Desc
	privateBytes                *prometheus.Desc
	threadCount                 *prometheus.Desc
	virtualBytes                *prometheus.Desc
	virtualBytesPeak            *prometheus.Desc
	workingSet                  *prometheus.Desc
	workingSetPeak              *prometheus.Desc
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
	err := wtsapi32.WTSCloseServer(c.hServer)
	if err != nil {
		return fmt.Errorf("failed to close WTS server: %w", err)
	}

	c.perfDataCollectorTerminalServicesSession.Close()

	if c.connectionBrokerEnabled {
		c.perfDataCollectorBroker.Close()
	}

	return nil
}

func (c *Collector) Build(logger *slog.Logger, miSession *mi.Session) error {
	if miSession == nil {
		return errors.New("miSession is nil")
	}

	c.logger = logger.With(slog.String("collector", Name))

	counters := []string{
		handleCount,
		pageFaultsPersec,
		pageFileBytes,
		pageFileBytesPeak,
		percentPrivilegedTime,
		percentProcessorTime,
		percentUserTime,
		poolNonpagedBytes,
		poolPagedBytes,
		privateBytes,
		threadCount,
		virtualBytes,
		virtualBytesPeak,
		workingSet,
		workingSetPeak,
	}

	var err error

	c.perfDataCollectorTerminalServicesSession, err = perfdata.NewCollector("Terminal Services Session", perfdata.InstanceAll, counters)
	if err != nil {
		return fmt.Errorf("failed to create Terminal Services Session collector: %w", err)
	}

	c.connectionBrokerEnabled = isConnectionBrokerServer(miSession)

	if c.connectionBrokerEnabled {
		counters = []string{
			successfulConnections,
			pendingConnections,
			failedConnections,
		}

		var err error

		c.perfDataCollectorBroker, err = perfdata.NewCollector("Remote Desktop Connection Broker Counterset", perfdata.InstanceAll, counters)
		if err != nil {
			return fmt.Errorf("failed to create Remote Desktop Connection Broker Counterset collector: %w", err)
		}
	} else {
		logger.Debug("host is not a connection broker skipping Connection Broker performance metrics.")
	}

	c.sessionInfo = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "session_info"),
		"Terminal Services sessions info",
		[]string{"session_name", "user", "host", "state", "id"},
		nil,
	)
	c.connectionBrokerPerformance = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "connection_broker_performance_total"),
		"The total number of connections handled by the Connection Brokers since the service started.",
		[]string{"connection"},
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

	c.hServer, err = wtsapi32.WTSOpenServer("")
	if err != nil {
		return fmt.Errorf("failed to open WTS server: %w", err)
	}

	return nil
}

// Collect sends the metric values for each metric
// to the provided prometheus Metric channel.
func (c *Collector) Collect(ch chan<- prometheus.Metric) error {
	errs := make([]error, 0, 3)

	if err := c.collectWTSSessions(ch); err != nil {
		errs = append(errs, fmt.Errorf("failed collecting terminal services session infos: %w", err))
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
	perfData, err := c.perfDataCollectorTerminalServicesSession.Collect()
	if err != nil {
		return fmt.Errorf("failed to collect Terminal Services Session metrics: %w", err)
	}

	names := make(map[string]bool)

	for name, data := range perfData {
		// only connect metrics for remote named sessions
		n := strings.ToLower(name)
		if n == "" || n == "services" || n == "console" {
			continue
		}
		// don't add name already present in labels list
		if _, ok := names[n]; ok {
			continue
		}

		names[n] = true

		ch <- prometheus.MustNewConstMetric(
			c.handleCount,
			prometheus.GaugeValue,
			data[handleCount].FirstValue,
			name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.pageFaultsPerSec,
			prometheus.CounterValue,
			data[pageFaultsPersec].FirstValue,
			name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.pageFileBytes,
			prometheus.GaugeValue,
			data[pageFileBytes].FirstValue,
			name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.pageFileBytesPeak,
			prometheus.GaugeValue,
			data[pageFileBytesPeak].FirstValue,
			name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.percentCPUTime,
			prometheus.CounterValue,
			data[percentPrivilegedTime].FirstValue,
			name,
			"privileged",
		)
		ch <- prometheus.MustNewConstMetric(
			c.percentCPUTime,
			prometheus.CounterValue,
			data[percentProcessorTime].FirstValue,
			name,
			"processor",
		)
		ch <- prometheus.MustNewConstMetric(
			c.percentCPUTime,
			prometheus.CounterValue,
			data[percentUserTime].FirstValue,
			name,
			"user",
		)
		ch <- prometheus.MustNewConstMetric(
			c.poolNonPagedBytes,
			prometheus.GaugeValue,
			data[poolNonpagedBytes].FirstValue,
			name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.poolPagedBytes,
			prometheus.GaugeValue,
			data[poolPagedBytes].FirstValue,
			name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.privateBytes,
			prometheus.GaugeValue,
			data[privateBytes].FirstValue,
			name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.threadCount,
			prometheus.GaugeValue,
			data[threadCount].FirstValue,
			name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.virtualBytes,
			prometheus.GaugeValue,
			data[virtualBytes].FirstValue,
			name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.virtualBytesPeak,
			prometheus.GaugeValue,
			data[virtualBytesPeak].FirstValue,
			name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.workingSet,
			prometheus.GaugeValue,
			data[workingSet].FirstValue,
			name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.workingSetPeak,
			prometheus.GaugeValue,
			data[workingSetPeak].FirstValue,
			name,
		)
	}

	return nil
}

func (c *Collector) collectCollectionBrokerPerformanceCounter(ch chan<- prometheus.Metric) error {
	perfData, err := c.perfDataCollectorBroker.Collect()
	if err != nil {
		return fmt.Errorf("failed to collect Remote Desktop Connection Broker Counterset metrics: %w", err)
	}

	data, ok := perfData[perfdata.EmptyInstance]
	if !ok {
		return errors.New("query for Remote Desktop Connection Broker Counterset returned empty result set")
	}

	ch <- prometheus.MustNewConstMetric(
		c.connectionBrokerPerformance,
		prometheus.CounterValue,
		data[successfulConnections].FirstValue,
		"Successful",
	)

	ch <- prometheus.MustNewConstMetric(
		c.connectionBrokerPerformance,
		prometheus.CounterValue,
		data[pendingConnections].FirstValue,
		"Pending",
	)

	ch <- prometheus.MustNewConstMetric(
		c.connectionBrokerPerformance,
		prometheus.CounterValue,
		data[failedConnections].FirstValue,
		"Failed",
	)

	return nil
}

func (c *Collector) collectWTSSessions(ch chan<- prometheus.Metric) error {
	sessions, err := wtsapi32.WTSEnumerateSessionsEx(c.hServer, c.logger)
	if err != nil {
		return fmt.Errorf("failed to enumerate WTS sessions: %w", err)
	}

	for _, session := range sessions {
		// only connect metrics for remote named sessions
		n := strings.ReplaceAll(session.SessionName, "#", " ")
		if n == "" || n == "Services" || n == "Console" {
			continue
		}

		userName := session.UserName
		if session.DomainName != "" {
			userName = fmt.Sprintf("%s\\%s", session.DomainName, session.UserName)
		}

		for stateID, stateName := range wtsapi32.WTSSessionStates {
			isState := 0.0
			if session.State == stateID {
				isState = 1.0
			}
			ch <- prometheus.MustNewConstMetric(
				c.sessionInfo,
				prometheus.GaugeValue,
				isState,
				n,
				userName,
				session.HostName,
				stateName,
				strconv.Itoa(int(session.SessionID)),
			)
		}
	}

	return nil
}
