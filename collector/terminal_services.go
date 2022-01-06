//go:build windows
// +build windows

package collector

import (
	"errors"
	"strings"

	"github.com/StackExchange/wmi"
	"github.com/prometheus-community/windows_exporter/log"
	"github.com/prometheus/client_golang/prometheus"
)

const ConnectionBrokerFeatureID uint32 = 133

func init() {
	registerCollector("terminal_services", NewTerminalServicesCollector, "Terminal Services", "Terminal Services Session", "Remote Desktop Connection Broker Counterset")
}

var (
	connectionBrokerEnabled = isConnectionBrokerServer()
)

type Win32_ServerFeature struct {
	ID uint32
}

func isConnectionBrokerServer() bool {
	var dst []Win32_ServerFeature
	q := queryAll(&dst)
	if err := wmi.Query(q, &dst); err != nil {
		return false
	}
	for _, d := range dst {
		if d.ID == ConnectionBrokerFeatureID {
			return true
		}
	}
	log.Debug("host is not a connection broker skipping Connection Broker performance metrics.")
	return false
}

// A TerminalServicesCollector is a Prometheus collector for WMI
// Win32_PerfRawData_LocalSessionManager_TerminalServices &  Win32_PerfRawData_TermService_TerminalServicesSession  metrics
// https://docs.microsoft.com/en-us/previous-versions/aa394344(v%3Dvs.85)
// https://wutils.com/wmi/root/cimv2/win32_perfrawdata_localsessionmanager_terminalservices/
type TerminalServicesCollector struct {
	LocalSessionCount           *prometheus.Desc
	ConnectionBrokerPerformance *prometheus.Desc
	HandleCount                 *prometheus.Desc
	PageFaultsPersec            *prometheus.Desc
	PageFileBytes               *prometheus.Desc
	PageFileBytesPeak           *prometheus.Desc
	PercentPrivilegedTime       *prometheus.Desc
	PercentProcessorTime        *prometheus.Desc
	PercentUserTime             *prometheus.Desc
	PoolNonpagedBytes           *prometheus.Desc
	PoolPagedBytes              *prometheus.Desc
	PrivateBytes                *prometheus.Desc
	ThreadCount                 *prometheus.Desc
	VirtualBytes                *prometheus.Desc
	VirtualBytesPeak            *prometheus.Desc
	WorkingSet                  *prometheus.Desc
	WorkingSetPeak              *prometheus.Desc
}

// NewTerminalServicesCollector ...
func NewTerminalServicesCollector() (Collector, error) {
	const subsystem = "terminal_services"
	return &TerminalServicesCollector{
		LocalSessionCount: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "local_session_count"),
			"Number of Terminal Services sessions",
			[]string{"session"},
			nil,
		),
		ConnectionBrokerPerformance: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "connection_broker_performance_total"),
			"The total number of connections handled by the Connection Brokers since the service started.",
			[]string{"connection"},
			nil,
		),
		HandleCount: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "handles"),
			"Total number of handles currently opened by this process. This number is the sum of the handles currently opened by each thread in this process.",
			[]string{"session_name"},
			nil,
		),
		PageFaultsPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "page_fault_total"),
			"Rate at which page faults occur in the threads executing in this process. A page fault occurs when a thread refers to a virtual memory page that is not in its working set in main memory. The page may not be retrieved from disk if it is on the standby list and therefore already in main memory. The page also may not be retrieved if it is in use by another process which shares the page.",
			[]string{"session_name"},
			nil,
		),
		PageFileBytes: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "page_file_bytes"),
			"Current number of bytes this process has used in the paging file(s). Paging files are used to store pages of memory used by the process that are not contained in other files. Paging files are shared by all processes, and lack of space in paging files can prevent other processes from allocating memory.",
			[]string{"session_name"},
			nil,
		),
		PageFileBytesPeak: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "page_file_bytes_peak"),
			"Maximum number of bytes this process has used in the paging file(s). Paging files are used to store pages of memory used by the process that are not contained in other files. Paging files are shared by all processes, and lack of space in paging files can prevent other processes from allocating memory.",
			[]string{"session_name"},
			nil,
		),
		PercentPrivilegedTime: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "privileged_time_seconds_total"),
			"Total elapsed time that the threads of the process have spent executing code in privileged mode.",
			[]string{"session_name"},
			nil,
		),
		PercentProcessorTime: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "processor_time_seconds_total"),
			"Total elapsed time that all of the threads of this process used the processor to execute instructions.",
			[]string{"session_name"},
			nil,
		),
		PercentUserTime: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "user_time_seconds_total"),
			"Total elapsed time that this process's threads have spent executing code in user mode. Applications, environment subsystems, and integral subsystems execute in user mode.",
			[]string{"session_name"},
			nil,
		),
		PoolNonpagedBytes: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "pool_non_paged_bytes"),
			"Number of bytes in the non-paged pool, an area of system memory (physical memory used by the operating system) for objects that cannot be written to disk, but must remain in physical memory as long as they are allocated. This property displays the last observed value only; it is not an average.",
			[]string{"session_name"},
			nil,
		),
		PoolPagedBytes: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "pool_paged_bytes"),
			"Number of bytes in the paged pool, an area of system memory (physical memory used by the operating system) for objects that can be written to disk when they are not being used. This property displays the last observed value only; it is not an average.",
			[]string{"session_name"},
			nil,
		),
		PrivateBytes: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "private_bytes"),
			"Current number of bytes this process has allocated that cannot be shared with other processes.",
			[]string{"session_name"},
			nil,
		),
		ThreadCount: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "threads"),
			"Number of threads currently active in this process. An instruction is the basic unit of execution in a processor, and a thread is the object that executes instructions. Every running process has at least one thread.",
			[]string{"session_name"},
			nil,
		),
		VirtualBytes: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "virtual_bytes"),
			"Current size, in bytes, of the virtual address space the process is using. Use of virtual address space does not necessarily imply corresponding use of either disk or main memory pages. Virtual space is finite and, by using too much, the process can limit its ability to load libraries.",
			[]string{"session_name"},
			nil,
		),
		VirtualBytesPeak: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "virtual_bytes_peak"),
			"Maximum number of bytes of virtual address space the process has used at any one time. Use of virtual address space does not necessarily imply corresponding use of either disk or main memory pages. Virtual space is finite and, by using too much, the process might limit its ability to load libraries.",
			[]string{"session_name"},
			nil,
		),
		WorkingSet: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "working_set_bytes"),
			"Current number of bytes in the working set of this process. The working set is the set of memory pages touched recently by the threads in the process. If free memory in the computer is above a threshold, pages are left in the working set of a process even if they are not in use. When free memory falls below a threshold, pages are trimmed from working sets. If they are needed, they are then soft-faulted back into the working set before they leave main memory.",
			[]string{"session_name"},
			nil,
		),
		WorkingSetPeak: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "working_set_bytes_peak"),
			"Maximum number of bytes in the working set of this process at any point in time. The working set is the set of memory pages touched recently by the threads in the process. If free memory in the computer is above a threshold, pages are left in the working set of a process even if they are not in use. When free memory falls below a threshold, pages are trimmed from working sets. If they are needed, they are then soft-faulted back into the working set before they leave main memory.",
			[]string{"session_name"},
			nil,
		),
	}, nil
}

// Collect sends the metric values for each metric
// to the provided prometheus Metric channel.
func (c *TerminalServicesCollector) Collect(ctx *ScrapeContext, ch chan<- prometheus.Metric) error {
	if desc, err := c.collectTSSessionCount(ctx, ch); err != nil {
		log.Error("failed collecting terminal services session count metrics:", desc, err)
		return err
	}
	if desc, err := c.collectTSSessionCounters(ctx, ch); err != nil {
		log.Error("failed collecting terminal services session count metrics:", desc, err)
		return err
	}

	// only collect CollectionBrokerPerformance if host is a Connection Broker
	if connectionBrokerEnabled {
		if desc, err := c.collectCollectionBrokerPerformanceCounter(ctx, ch); err != nil {
			log.Error("failed collecting Connection Broker performance metrics:", desc, err)
			return err
		}
	}
	return nil
}

type perflibTerminalServices struct {
	ActiveSessions   float64 `perflib:"Active Sessions"`
	InactiveSessions float64 `perflib:"Inactive Sessions"`
	TotalSessions    float64 `perflib:"Total Sessions"`
}

func (c *TerminalServicesCollector) collectTSSessionCount(ctx *ScrapeContext, ch chan<- prometheus.Metric) (*prometheus.Desc, error) {
	dst := make([]perflibTerminalServices, 0)
	err := unmarshalObject(ctx.perfObjects["Terminal Services"], &dst)
	if err != nil {
		return nil, err
	}
	if len(dst) == 0 {
		return nil, errors.New("WMI query returned empty result set")
	}

	ch <- prometheus.MustNewConstMetric(
		c.LocalSessionCount,
		prometheus.GaugeValue,
		dst[0].ActiveSessions,
		"active",
	)

	ch <- prometheus.MustNewConstMetric(
		c.LocalSessionCount,
		prometheus.GaugeValue,
		dst[0].InactiveSessions,
		"inactive",
	)

	ch <- prometheus.MustNewConstMetric(
		c.LocalSessionCount,
		prometheus.GaugeValue,
		dst[0].TotalSessions,
		"total",
	)

	return nil, nil
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

func (c *TerminalServicesCollector) collectTSSessionCounters(ctx *ScrapeContext, ch chan<- prometheus.Metric) (*prometheus.Desc, error) {
	dst := make([]perflibTerminalServicesSession, 0)
	err := unmarshalObject(ctx.perfObjects["Terminal Services Session"], &dst)
	if err != nil {
		return nil, err
	}

	for _, d := range dst {
		// only connect metrics for remote named sessions
		n := strings.ToLower(d.Name)
		if n == "" || n == "services" || n == "console" {
			continue
		}
		ch <- prometheus.MustNewConstMetric(
			c.HandleCount,
			prometheus.GaugeValue,
			d.HandleCount,
			d.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.PageFaultsPersec,
			prometheus.CounterValue,
			d.PageFaultsPersec,
			d.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.PageFileBytes,
			prometheus.GaugeValue,
			d.PageFileBytes,
			d.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.PageFileBytesPeak,
			prometheus.GaugeValue,
			d.PageFileBytesPeak,
			d.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.PercentPrivilegedTime,
			prometheus.CounterValue,
			d.PercentPrivilegedTime,
			d.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.PercentProcessorTime,
			prometheus.CounterValue,
			d.PercentProcessorTime,
			d.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.PercentUserTime,
			prometheus.CounterValue,
			d.PercentUserTime,
			d.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.PoolNonpagedBytes,
			prometheus.GaugeValue,
			d.PoolNonpagedBytes,
			d.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.PoolPagedBytes,
			prometheus.GaugeValue,
			d.PoolPagedBytes,
			d.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.PrivateBytes,
			prometheus.GaugeValue,
			d.PrivateBytes,
			d.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.ThreadCount,
			prometheus.GaugeValue,
			d.ThreadCount,
			d.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.VirtualBytes,
			prometheus.GaugeValue,
			d.VirtualBytes,
			d.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.VirtualBytesPeak,
			prometheus.GaugeValue,
			d.VirtualBytesPeak,
			d.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.WorkingSet,
			prometheus.GaugeValue,
			d.WorkingSet,
			d.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.WorkingSetPeak,
			prometheus.GaugeValue,
			d.WorkingSetPeak,
			d.Name,
		)
	}
	return nil, nil
}

type perflibRemoteDesktopConnectionBrokerCounterset struct {
	SuccessfulConnections float64 `perflib:"Successful Connections"`
	PendingConnections    float64 `perflib:"Pending Connections"`
	FailedConnections     float64 `perflib:"Failed Connections"`
}

func (c *TerminalServicesCollector) collectCollectionBrokerPerformanceCounter(ctx *ScrapeContext, ch chan<- prometheus.Metric) (*prometheus.Desc, error) {

	dst := make([]perflibRemoteDesktopConnectionBrokerCounterset, 0)
	err := unmarshalObject(ctx.perfObjects["Remote Desktop Connection Broker Counterset"], &dst)
	if err != nil {
		return nil, err
	}
	if len(dst) == 0 {
		return nil, errors.New("WMI query returned empty result set")
	}

	ch <- prometheus.MustNewConstMetric(
		c.ConnectionBrokerPerformance,
		prometheus.CounterValue,
		dst[0].SuccessfulConnections,
		"Successful",
	)

	ch <- prometheus.MustNewConstMetric(
		c.ConnectionBrokerPerformance,
		prometheus.CounterValue,
		dst[0].PendingConnections,
		"Pending",
	)

	ch <- prometheus.MustNewConstMetric(
		c.ConnectionBrokerPerformance,
		prometheus.CounterValue,
		dst[0].FailedConnections,
		"Failed",
	)

	return nil, nil
}
