// +build windows

package collector

import (
	"errors"

	"github.com/StackExchange/wmi"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
)

const ConnectionBrokerFeatureID uint32 = 133

func init() {
	registerCollector("terminal_services", NewTerminalServicesCollector)
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
			prometheus.BuildFQName(Namespace, subsystem, "handle_count"),
			"Total number of handles currently opened by this process. This number is the sum of the handles currently opened by each thread in this process.",
			[]string{"session_name"},
			nil,
		),
		PageFaultsPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "page_fault_per_sec"),
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
			prometheus.BuildFQName(Namespace, subsystem, "percent_privileged_time"),
			"Percentage of elapsed time that the threads of the process have spent executing code in privileged mode.",
			[]string{"session_name"},
			nil,
		),
		PercentProcessorTime: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "percent_processor_time"),
			"Percentage of elapsed time that all of the threads of this process used the processor to execute instructions.",
			[]string{"session_name"},
			nil,
		),
		PercentUserTime: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "percent_user_time"),
			"Percentage of elapsed time that this process's threads have spent executing code in user mode. Applications, environment subsystems, and integral subsystems execute in user mode.",
			[]string{"session_name"},
			nil,
		),
		PoolNonpagedBytes: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "pool_non_paged_Bytes"),
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
			prometheus.BuildFQName(Namespace, subsystem, "thread_count"),
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
			prometheus.BuildFQName(Namespace, subsystem, "workingset"),
			"Current number of bytes in the working set of this process. The working set is the set of memory pages touched recently by the threads in the process. If free memory in the computer is above a threshold, pages are left in the working set of a process even if they are not in use. When free memory falls below a threshold, pages are trimmed from working sets. If they are needed, they are then soft-faulted back into the working set before they leave main memory.",
			[]string{"session_name"},
			nil,
		),
		WorkingSetPeak: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "workingset_peak"),
			"Maximum number of bytes in the working set of this process at any point in time. The working set is the set of memory pages touched recently by the threads in the process. If free memory in the computer is above a threshold, pages are left in the working set of a process even if they are not in use. When free memory falls below a threshold, pages are trimmed from working sets. If they are needed, they are then soft-faulted back into the working set before they leave main memory.",
			[]string{"session_name"},
			nil,
		),
	}, nil
}

// Collect sends the metric values for each metric
// to the provided prometheus Metric channel.
func (c *TerminalServicesCollector) Collect(ctx *ScrapeContext, ch chan<- prometheus.Metric) error {
	if desc, err := c.collectTSSessionCount(ch); err != nil {
		log.Error("failed collecting terminal services session count metrics:", desc, err)
		return err
	}
	if desc, err := c.collectTSSessionCounters(ch); err != nil {
		log.Error("failed collecting terminal services session count metrics:", desc, err)
		return err
	}

	// only collect CollectionBrokerPerformance if host is a Connection Broker
	if connectionBrokerEnabled {
		if desc, err := c.collectCollectionBrokerPerformanceCounter(ch); err != nil {
			log.Error("failed collecting Connection Broker performance metrics:", desc, err)
			return err
		}
	}
	return nil
}

type Win32_PerfRawData_LocalSessionManager_TerminalServices struct {
	ActiveSessions   uint32
	InactiveSessions uint32
	TotalSessions    uint32
}

func (c *TerminalServicesCollector) collectTSSessionCount(ch chan<- prometheus.Metric) (*prometheus.Desc, error) {
	var dst []Win32_PerfRawData_LocalSessionManager_TerminalServices
	q := queryAll(&dst)
	if err := wmi.Query(q, &dst); err != nil {
		return nil, err
	}
	if len(dst) == 0 {
		return nil, errors.New("WMI query returned empty result set")
	}

	ch <- prometheus.MustNewConstMetric(
		c.LocalSessionCount,
		prometheus.GaugeValue,
		float64(dst[0].ActiveSessions),
		"active",
	)

	ch <- prometheus.MustNewConstMetric(
		c.LocalSessionCount,
		prometheus.GaugeValue,
		float64(dst[0].InactiveSessions),
		"inactive",
	)

	ch <- prometheus.MustNewConstMetric(
		c.LocalSessionCount,
		prometheus.GaugeValue,
		float64(dst[0].TotalSessions),
		"total",
	)

	return nil, nil
}

type Win32_PerfRawData_TermService_TerminalServicesSession struct {
	Name                  string
	HandleCount           uint32
	PageFaultsPersec      uint32
	PageFileBytes         uint64
	PageFileBytesPeak     uint64
	PercentPrivilegedTime uint64
	PercentProcessorTime  uint64
	PercentUserTime       uint64
	PoolNonpagedBytes     uint32
	PoolPagedBytes        uint32
	PrivateBytes          uint64
	ThreadCount           uint32
	VirtualBytes          uint64
	VirtualBytesPeak      uint64
	WorkingSet            uint64
	WorkingSetPeak        uint64
}

func (c *TerminalServicesCollector) collectTSSessionCounters(ch chan<- prometheus.Metric) (*prometheus.Desc, error) {
	var dst []Win32_PerfRawData_TermService_TerminalServicesSession
	q := queryAll(&dst)
	if err := wmi.Query(q, &dst); err != nil {
		return nil, err
	}
	if len(dst) == 0 {
		return nil, errors.New("WMI query returned empty result set")
	}

	for _, d := range dst {
		ch <- prometheus.MustNewConstMetric(
			c.HandleCount,
			prometheus.GaugeValue,
			float64(d.HandleCount),
			d.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.PageFaultsPersec,
			prometheus.GaugeValue,
			float64(d.PageFaultsPersec),
			d.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.PageFileBytes,
			prometheus.GaugeValue,
			float64(d.PageFileBytes),
			d.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.PageFileBytesPeak,
			prometheus.GaugeValue,
			float64(d.PageFileBytesPeak),
			d.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.PercentPrivilegedTime,
			prometheus.GaugeValue,
			float64(d.PercentPrivilegedTime),
			d.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.PercentProcessorTime,
			prometheus.GaugeValue,
			float64(d.PercentProcessorTime),
			d.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.PercentUserTime,
			prometheus.GaugeValue,
			float64(d.PercentUserTime),
			d.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.PoolNonpagedBytes,
			prometheus.GaugeValue,
			float64(d.PoolNonpagedBytes),
			d.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.PoolPagedBytes,
			prometheus.GaugeValue,
			float64(d.PoolPagedBytes),
			d.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.PrivateBytes,
			prometheus.GaugeValue,
			float64(d.PrivateBytes),
			d.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.ThreadCount,
			prometheus.GaugeValue,
			float64(d.ThreadCount),
			d.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.VirtualBytes,
			prometheus.GaugeValue,
			float64(d.VirtualBytes),
			d.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.VirtualBytesPeak,
			prometheus.GaugeValue,
			float64(d.VirtualBytesPeak),
			d.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.WorkingSet,
			prometheus.GaugeValue,
			float64(d.WorkingSet),
			d.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.WorkingSetPeak,
			prometheus.GaugeValue,
			float64(d.WorkingSetPeak),
			d.Name,
		)
	}
	return nil, nil
}

type Win32_PerfRawData_RemoteDesktopConnectionBrokerPerformanceCounterProvider_RemoteDesktopConnectionBrokerCounterset struct {
	SuccessfulConnections uint64
	PendingConnections    uint64
	FailedConnections     uint64
}

func (c *TerminalServicesCollector) collectCollectionBrokerPerformanceCounter(ch chan<- prometheus.Metric) (*prometheus.Desc, error) {

	var dst []Win32_PerfRawData_RemoteDesktopConnectionBrokerPerformanceCounterProvider_RemoteDesktopConnectionBrokerCounterset
	q := queryAll(&dst)
	if err := wmi.Query(q, &dst); err != nil {
		return nil, err
	}
	if len(dst) == 0 {
		return nil, errors.New("WMI query returned empty result set")
	}

	ch <- prometheus.MustNewConstMetric(
		c.ConnectionBrokerPerformance,
		prometheus.CounterValue,
		float64(dst[0].SuccessfulConnections),
		"Successful",
	)

	ch <- prometheus.MustNewConstMetric(
		c.ConnectionBrokerPerformance,
		prometheus.CounterValue,
		float64(dst[0].PendingConnections),
		"Pending",
	)

	ch <- prometheus.MustNewConstMetric(
		c.ConnectionBrokerPerformance,
		prometheus.CounterValue,
		float64(dst[0].FailedConnections),
		"Failed",
	)

	return nil, nil
}
