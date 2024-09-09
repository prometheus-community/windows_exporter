//go:build windows

package terminal_services

import (
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"strings"

	"github.com/alecthomas/kingpin/v2"
	"github.com/prometheus-community/windows_exporter/pkg/headers/wtsapi32"
	"github.com/prometheus-community/windows_exporter/pkg/perflib"
	"github.com/prometheus-community/windows_exporter/pkg/types"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/yusufpapurcu/wmi"
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

func isConnectionBrokerServer(logger *slog.Logger, wmiClient *wmi.Client) bool {
	var dst []Win32_ServerFeature
	if err := wmiClient.Query("SELECT * FROM Win32_ServerFeature", &dst); err != nil {
		return false
	}

	for _, d := range dst {
		if d.ID == ConnectionBrokerFeatureID {
			return true
		}
	}

	logger.Debug("host is not a connection broker skipping Connection Broker performance metrics.")

	return false
}

// A Collector is a Prometheus Collector for WMI
// Win32_PerfRawData_LocalSessionManager_TerminalServices &  Win32_PerfRawData_TermService_TerminalServicesSession  metrics
// https://docs.microsoft.com/en-us/previous-versions/aa394344(v%3Dvs.85)
// https://wutils.com/wmi/root/cimv2/win32_perfrawdata_localsessionmanager_terminalservices/
type Collector struct {
	config Config

	connectionBrokerEnabled bool

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

func (c *Collector) GetPerfCounter(_ *slog.Logger) ([]string, error) {
	return []string{
		"Terminal Services Session",
		"Remote Desktop Connection Broker Counterset",
	}, nil
}

func (c *Collector) Close(_ *slog.Logger) error {
	err := wtsapi32.WTSCloseServer(c.hServer)
	if err != nil {
		return fmt.Errorf("failed to close WTS server: %w", err)
	}

	return nil
}

func (c *Collector) Build(logger *slog.Logger, wmiClient *wmi.Client) error {
	logger = logger.With(slog.String("collector", Name))

	c.connectionBrokerEnabled = isConnectionBrokerServer(logger, wmiClient)

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

	var err error

	c.hServer, err = wtsapi32.WTSOpenServer("")
	if err != nil {
		return fmt.Errorf("failed to open WTS server: %w", err)
	}

	return nil
}

// Collect sends the metric values for each metric
// to the provided prometheus Metric channel.
func (c *Collector) Collect(ctx *types.ScrapeContext, logger *slog.Logger, ch chan<- prometheus.Metric) error {
	logger = logger.With(slog.String("collector", Name))
	if err := c.collectWTSSessions(logger, ch); err != nil {
		logger.Error("failed collecting terminal services session infos",
			slog.Any("err", err),
		)

		return err
	}

	if err := c.collectTSSessionCounters(ctx, logger, ch); err != nil {
		logger.Error("failed collecting terminal services session count metrics",
			slog.Any("err", err),
		)

		return err
	}

	// only collect CollectionBrokerPerformance if host is a Connection Broker
	if c.connectionBrokerEnabled {
		if err := c.collectCollectionBrokerPerformanceCounter(ctx, logger, ch); err != nil {
			logger.Error("failed collecting Connection Broker performance metrics",
				slog.Any("err", err),
			)

			return err
		}
	}

	return nil
}

type perflibTerminalServicesSession struct {
	Name                  string
	HandleCount           float64 `perflib:"Handle Count"`
	PageFaultsPersec      float64 `perflib:"Page Faults/sec"`
	PageFileBytes         float64 `perflib:"Page File Bytes"`
	PageFileBytesPeak     float64 `perflib:"Page File Bytes Peak"`
	PercentPrivilegedTime float64 `perflib:"% Privileged Time"`
	PercentProcessorTime  float64 `perflib:"% Processor Time"`
	PercentUserTime       float64 `perflib:"% User Time"`
	PoolNonpagedBytes     float64 `perflib:"Pool Nonpaged Bytes"`
	PoolPagedBytes        float64 `perflib:"Pool Paged Bytes"`
	PrivateBytes          float64 `perflib:"Private Bytes"`
	ThreadCount           float64 `perflib:"Thread Count"`
	VirtualBytes          float64 `perflib:"Virtual Bytes"`
	VirtualBytesPeak      float64 `perflib:"Virtual Bytes Peak"`
	WorkingSet            float64 `perflib:"Working Set"`
	WorkingSetPeak        float64 `perflib:"Working Set Peak"`
}

func (c *Collector) collectTSSessionCounters(ctx *types.ScrapeContext, logger *slog.Logger, ch chan<- prometheus.Metric) error {
	logger = logger.With(slog.String("collector", Name))
	dst := make([]perflibTerminalServicesSession, 0)

	err := perflib.UnmarshalObject(ctx.PerfObjects["Terminal Services Session"], &dst, logger)
	if err != nil {
		return err
	}

	names := make(map[string]bool)

	for _, d := range dst {
		// only connect metrics for remote named sessions
		n := strings.ToLower(d.Name)
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
			d.HandleCount,
			d.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.pageFaultsPerSec,
			prometheus.CounterValue,
			d.PageFaultsPersec,
			d.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.pageFileBytes,
			prometheus.GaugeValue,
			d.PageFileBytes,
			d.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.pageFileBytesPeak,
			prometheus.GaugeValue,
			d.PageFileBytesPeak,
			d.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.percentCPUTime,
			prometheus.CounterValue,
			d.PercentPrivilegedTime,
			d.Name,
			"privileged",
		)
		ch <- prometheus.MustNewConstMetric(
			c.percentCPUTime,
			prometheus.CounterValue,
			d.PercentProcessorTime,
			d.Name,
			"processor",
		)
		ch <- prometheus.MustNewConstMetric(
			c.percentCPUTime,
			prometheus.CounterValue,
			d.PercentUserTime,
			d.Name,
			"user",
		)
		ch <- prometheus.MustNewConstMetric(
			c.poolNonPagedBytes,
			prometheus.GaugeValue,
			d.PoolNonpagedBytes,
			d.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.poolPagedBytes,
			prometheus.GaugeValue,
			d.PoolPagedBytes,
			d.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.privateBytes,
			prometheus.GaugeValue,
			d.PrivateBytes,
			d.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.threadCount,
			prometheus.GaugeValue,
			d.ThreadCount,
			d.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.virtualBytes,
			prometheus.GaugeValue,
			d.VirtualBytes,
			d.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.virtualBytesPeak,
			prometheus.GaugeValue,
			d.VirtualBytesPeak,
			d.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.workingSet,
			prometheus.GaugeValue,
			d.WorkingSet,
			d.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.workingSetPeak,
			prometheus.GaugeValue,
			d.WorkingSetPeak,
			d.Name,
		)
	}

	return nil
}

type perflibRemoteDesktopConnectionBrokerCounterset struct {
	SuccessfulConnections float64 `perflib:"Successful Connections"`
	PendingConnections    float64 `perflib:"Pending Connections"`
	FailedConnections     float64 `perflib:"Failed Connections"`
}

func (c *Collector) collectCollectionBrokerPerformanceCounter(ctx *types.ScrapeContext, logger *slog.Logger, ch chan<- prometheus.Metric) error {
	logger = logger.With(slog.String("collector", Name))
	dst := make([]perflibRemoteDesktopConnectionBrokerCounterset, 0)

	err := perflib.UnmarshalObject(ctx.PerfObjects["Remote Desktop Connection Broker Counterset"], &dst, logger)
	if err != nil {
		return err
	}

	if len(dst) == 0 {
		return errors.New("WMI query returned empty result set")
	}

	ch <- prometheus.MustNewConstMetric(
		c.connectionBrokerPerformance,
		prometheus.CounterValue,
		dst[0].SuccessfulConnections,
		"Successful",
	)

	ch <- prometheus.MustNewConstMetric(
		c.connectionBrokerPerformance,
		prometheus.CounterValue,
		dst[0].PendingConnections,
		"Pending",
	)

	ch <- prometheus.MustNewConstMetric(
		c.connectionBrokerPerformance,
		prometheus.CounterValue,
		dst[0].FailedConnections,
		"Failed",
	)

	return nil
}

func (c *Collector) collectWTSSessions(logger *slog.Logger, ch chan<- prometheus.Metric) error {
	sessions, err := wtsapi32.WTSEnumerateSessionsEx(c.hServer, logger)
	if err != nil {
		return fmt.Errorf("failed to enumerate WTS sessions: %w", err)
	}

	for _, session := range sessions {
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
				strings.ReplaceAll(session.SessionName, "#", " "),
				userName,
				session.HostName,
				stateName,
				strconv.Itoa(int(session.SessionID)),
			)
		}
	}

	return nil
}
