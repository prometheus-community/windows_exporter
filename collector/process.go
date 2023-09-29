//go:build windows
// +build windows

package collector

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/alecthomas/kingpin/v2"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/yusufpapurcu/wmi"
)

const (
	FlagProcessOldExclude = "collector.process.blacklist"
	FlagProcessOldInclude = "collector.process.whitelist"

	FlagProcessExclude = "collector.process.exclude"
	FlagProcessInclude = "collector.process.include"
)

var (
	processOldInclude *string
	processOldExclude *string

	processInclude *string
	processExclude *string

	processIncludeSet bool
	processExcludeSet bool

	enableWorkerProcess *bool
)

type processCollector struct {
	logger log.Logger

	StartTime         *prometheus.Desc
	CPUTimeTotal      *prometheus.Desc
	HandleCount       *prometheus.Desc
	IOBytesTotal      *prometheus.Desc
	IOOperationsTotal *prometheus.Desc
	PageFaultsTotal   *prometheus.Desc
	PageFileBytes     *prometheus.Desc
	PoolBytes         *prometheus.Desc
	PriorityBase      *prometheus.Desc
	PrivateBytes      *prometheus.Desc
	ThreadCount       *prometheus.Desc
	VirtualBytes      *prometheus.Desc
	WorkingSetPrivate *prometheus.Desc
	WorkingSetPeak    *prometheus.Desc
	WorkingSet        *prometheus.Desc

	processIncludePattern *regexp.Regexp
	processExcludePattern *regexp.Regexp
}

// newProcessCollectorFlags ...
func newProcessCollectorFlags(app *kingpin.Application) {
	processInclude = app.Flag(
		FlagProcessInclude,
		"Regexp of processes to include. Process name must both match include and not match exclude to be included.",
	).Default(".*").PreAction(func(c *kingpin.ParseContext) error {
		processIncludeSet = true
		return nil
	}).String()

	processExclude = app.Flag(
		FlagProcessExclude,
		"Regexp of processes to exclude. Process name must both match include and not match exclude to be included.",
	).Default("").PreAction(func(c *kingpin.ParseContext) error {
		processExcludeSet = true
		return nil
	}).String()

	enableWorkerProcess = kingpin.Flag(
		"collector.process.iis",
		"Enable IIS worker process name queries. May cause the collector to leak memory.",
	).Default("false").Bool()

	processOldInclude = app.Flag(
		FlagProcessOldInclude,
		"DEPRECATED: Use --collector.process.include",
	).Hidden().String()
	processOldExclude = app.Flag(
		FlagProcessOldExclude,
		"DEPRECATED: Use --collector.process.exclude",
	).Hidden().String()
}

// NewProcessCollector ...
func newProcessCollector(logger log.Logger) (Collector, error) {
	const subsystem = "process"
	logger = log.With(logger, "collector", subsystem)

	if *processOldExclude != "" {
		if !processExcludeSet {
			_ = level.Warn(logger).Log("msg", "--collector.process.blacklist is DEPRECATED and will be removed in a future release, use --collector.process.exclude")
			*processExclude = *processOldExclude
		} else {
			return nil, errors.New("--collector.process.blacklist and --collector.process.exclude are mutually exclusive")
		}
	}
	if *processOldInclude != "" {
		if !processIncludeSet {
			_ = level.Warn(logger).Log("msg", "--collector.process.whitelist is DEPRECATED and will be removed in a future release, use --collector.process.include")
			*processInclude = *processOldInclude
		} else {
			return nil, errors.New("--collector.process.whitelist and --collector.process.include are mutually exclusive")
		}
	}

	if *processInclude == ".*" && *processExclude == "" {
		_ = level.Warn(logger).Log("msg", "No filters specified for process collector. This will generate a very large number of metrics!")
	}

	return &processCollector{
		logger: logger,
		StartTime: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "start_time"),
			"Time of process start.",
			[]string{"process", "process_id", "creating_process_id"},
			nil,
		),
		CPUTimeTotal: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "cpu_time_total"),
			"Returns elapsed time that all of the threads of this process used the processor to execute instructions by mode (privileged, user).",
			[]string{"process", "process_id", "creating_process_id", "mode"},
			nil,
		),
		HandleCount: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "handles"),
			"Total number of handles the process has open. This number is the sum of the handles currently open by each thread in the process.",
			[]string{"process", "process_id", "creating_process_id"},
			nil,
		),
		IOBytesTotal: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "io_bytes_total"),
			"Bytes issued to I/O operations in different modes (read, write, other).",
			[]string{"process", "process_id", "creating_process_id", "mode"},
			nil,
		),
		IOOperationsTotal: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "io_operations_total"),
			"I/O operations issued in different modes (read, write, other).",
			[]string{"process", "process_id", "creating_process_id", "mode"},
			nil,
		),
		PageFaultsTotal: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "page_faults_total"),
			"Page faults by the threads executing in this process.",
			[]string{"process", "process_id", "creating_process_id"},
			nil,
		),
		PageFileBytes: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "page_file_bytes"),
			"Current number of bytes this process has used in the paging file(s).",
			[]string{"process", "process_id", "creating_process_id"},
			nil,
		),
		PoolBytes: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "pool_bytes"),
			"Pool Bytes is the last observed number of bytes in the paged or nonpaged pool.",
			[]string{"process", "process_id", "creating_process_id", "pool"},
			nil,
		),
		PriorityBase: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "priority_base"),
			"Current base priority of this process. Threads within a process can raise and lower their own base priority relative to the process base priority of the process.",
			[]string{"process", "process_id", "creating_process_id"},
			nil,
		),
		PrivateBytes: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "private_bytes"),
			"Current number of bytes this process has allocated that cannot be shared with other processes.",
			[]string{"process", "process_id", "creating_process_id"},
			nil,
		),
		ThreadCount: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "threads"),
			"Number of threads currently active in this process.",
			[]string{"process", "process_id", "creating_process_id"},
			nil,
		),
		VirtualBytes: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "virtual_bytes"),
			"Current size, in bytes, of the virtual address space that the process is using.",
			[]string{"process", "process_id", "creating_process_id"},
			nil,
		),
		WorkingSetPrivate: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "working_set_private_bytes"),
			"Size of the working set, in bytes, that is use for this process only and not shared nor shareable by other processes.",
			[]string{"process", "process_id", "creating_process_id"},
			nil,
		),
		WorkingSetPeak: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "working_set_peak_bytes"),
			"Maximum size, in bytes, of the Working Set of this process at any point in time. The Working Set is the set of memory pages touched recently by the threads in the process.",
			[]string{"process", "process_id", "creating_process_id"},
			nil,
		),
		WorkingSet: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "working_set_bytes"),
			"Maximum number of bytes in the working set of this process at any point in time. The working set is the set of memory pages touched recently by the threads in the process.",
			[]string{"process", "process_id", "creating_process_id"},
			nil,
		),
		processIncludePattern: regexp.MustCompile(fmt.Sprintf("^(?:%s)$", *processInclude)),
		processExcludePattern: regexp.MustCompile(fmt.Sprintf("^(?:%s)$", *processExclude)),
	}, nil
}

type perflibProcess struct {
	Name                    string
	PercentProcessorTime    float64 `perflib:"% Processor Time"`
	PercentPrivilegedTime   float64 `perflib:"% Privileged Time"`
	PercentUserTime         float64 `perflib:"% User Time"`
	CreatingProcessID       float64 `perflib:"Creating Process ID"`
	ElapsedTime             float64 `perflib:"Elapsed Time"`
	HandleCount             float64 `perflib:"Handle Count"`
	IDProcess               float64 `perflib:"ID Process"`
	IODataBytesPerSec       float64 `perflib:"IO Data Bytes/sec"`
	IODataOperationsPerSec  float64 `perflib:"IO Data Operations/sec"`
	IOOtherBytesPerSec      float64 `perflib:"IO Other Bytes/sec"`
	IOOtherOperationsPerSec float64 `perflib:"IO Other Operations/sec"`
	IOReadBytesPerSec       float64 `perflib:"IO Read Bytes/sec"`
	IOReadOperationsPerSec  float64 `perflib:"IO Read Operations/sec"`
	IOWriteBytesPerSec      float64 `perflib:"IO Write Bytes/sec"`
	IOWriteOperationsPerSec float64 `perflib:"IO Write Operations/sec"`
	PageFaultsPerSec        float64 `perflib:"Page Faults/sec"`
	PageFileBytesPeak       float64 `perflib:"Page File Bytes Peak"`
	PageFileBytes           float64 `perflib:"Page File Bytes"`
	PoolNonpagedBytes       float64 `perflib:"Pool Nonpaged Bytes"`
	PoolPagedBytes          float64 `perflib:"Pool Paged Bytes"`
	PriorityBase            float64 `perflib:"Priority Base"`
	PrivateBytes            float64 `perflib:"Private Bytes"`
	ThreadCount             float64 `perflib:"Thread Count"`
	VirtualBytesPeak        float64 `perflib:"Virtual Bytes Peak"`
	VirtualBytes            float64 `perflib:"Virtual Bytes"`
	WorkingSetPrivate       float64 `perflib:"Working Set - Private"`
	WorkingSetPeak          float64 `perflib:"Working Set Peak"`
	WorkingSet              float64 `perflib:"Working Set"`
}

type WorkerProcess struct {
	AppPoolName string
	ProcessId   uint64
}

func (c *processCollector) Collect(ctx *ScrapeContext, ch chan<- prometheus.Metric) error {
	data := make([]perflibProcess, 0)
	err := unmarshalObject(ctx.perfObjects["Process"], &data, c.logger)
	if err != nil {
		return err
	}

	var dst_wp []WorkerProcess
	if *enableWorkerProcess {
		q_wp := queryAll(&dst_wp, c.logger)
		if err := wmi.QueryNamespace(q_wp, &dst_wp, "root\\WebAdministration"); err != nil {
			_ = level.Debug(c.logger).Log(fmt.Sprintf("Could not query WebAdministration namespace for IIS worker processes: %v. Skipping\n", err))
		}
	}

	for _, process := range data {
		if process.Name == "_Total" ||
			c.processExcludePattern.MatchString(process.Name) ||
			!c.processIncludePattern.MatchString(process.Name) {
			continue
		}
		// Duplicate processes are suffixed # and an index number. Remove those.
		processName := strings.Split(process.Name, "#")[0]
		pid := strconv.FormatUint(uint64(process.IDProcess), 10)
		cpid := strconv.FormatUint(uint64(process.CreatingProcessID), 10)

		if *enableWorkerProcess {
			for _, wp := range dst_wp {
				if wp.ProcessId == uint64(process.IDProcess) {
					processName = strings.Join([]string{processName, wp.AppPoolName}, "_")
					break
				}
			}
		}

		ch <- prometheus.MustNewConstMetric(
			c.StartTime,
			prometheus.GaugeValue,
			process.ElapsedTime,
			processName,
			pid,
			cpid,
		)

		ch <- prometheus.MustNewConstMetric(
			c.HandleCount,
			prometheus.GaugeValue,
			process.HandleCount,
			processName,
			pid,
			cpid,
		)

		ch <- prometheus.MustNewConstMetric(
			c.CPUTimeTotal,
			prometheus.CounterValue,
			process.PercentPrivilegedTime,
			processName,
			pid,
			cpid,
			"privileged",
		)

		ch <- prometheus.MustNewConstMetric(
			c.CPUTimeTotal,
			prometheus.CounterValue,
			process.PercentUserTime,
			processName,
			pid,
			cpid,
			"user",
		)

		ch <- prometheus.MustNewConstMetric(
			c.IOBytesTotal,
			prometheus.CounterValue,
			process.IOOtherBytesPerSec,
			processName,
			pid,
			cpid,
			"other",
		)

		ch <- prometheus.MustNewConstMetric(
			c.IOOperationsTotal,
			prometheus.CounterValue,
			process.IOOtherOperationsPerSec,
			processName,
			pid,
			cpid,
			"other",
		)

		ch <- prometheus.MustNewConstMetric(
			c.IOBytesTotal,
			prometheus.CounterValue,
			process.IOReadBytesPerSec,
			processName,
			pid,
			cpid,
			"read",
		)

		ch <- prometheus.MustNewConstMetric(
			c.IOOperationsTotal,
			prometheus.CounterValue,
			process.IOReadOperationsPerSec,
			processName,
			pid,
			cpid,
			"read",
		)

		ch <- prometheus.MustNewConstMetric(
			c.IOBytesTotal,
			prometheus.CounterValue,
			process.IOWriteBytesPerSec,
			processName,
			pid,
			cpid,
			"write",
		)

		ch <- prometheus.MustNewConstMetric(
			c.IOOperationsTotal,
			prometheus.CounterValue,
			process.IOWriteOperationsPerSec,
			processName,
			pid,
			cpid,
			"write",
		)

		ch <- prometheus.MustNewConstMetric(
			c.PageFaultsTotal,
			prometheus.CounterValue,
			process.PageFaultsPerSec,
			processName,
			pid,
			cpid,
		)

		ch <- prometheus.MustNewConstMetric(
			c.PageFileBytes,
			prometheus.GaugeValue,
			process.PageFileBytes,
			processName,
			pid,
			cpid,
		)

		ch <- prometheus.MustNewConstMetric(
			c.PoolBytes,
			prometheus.GaugeValue,
			process.PoolNonpagedBytes,
			processName,
			pid,
			cpid,
			"nonpaged",
		)

		ch <- prometheus.MustNewConstMetric(
			c.PoolBytes,
			prometheus.GaugeValue,
			process.PoolPagedBytes,
			processName,
			pid,
			cpid,
			"paged",
		)

		ch <- prometheus.MustNewConstMetric(
			c.PriorityBase,
			prometheus.GaugeValue,
			process.PriorityBase,
			processName,
			pid,
			cpid,
		)

		ch <- prometheus.MustNewConstMetric(
			c.PrivateBytes,
			prometheus.GaugeValue,
			process.PrivateBytes,
			processName,
			pid,
			cpid,
		)

		ch <- prometheus.MustNewConstMetric(
			c.ThreadCount,
			prometheus.GaugeValue,
			process.ThreadCount,
			processName,
			pid,
			cpid,
		)

		ch <- prometheus.MustNewConstMetric(
			c.VirtualBytes,
			prometheus.GaugeValue,
			process.VirtualBytes,
			processName,
			pid,
			cpid,
		)

		ch <- prometheus.MustNewConstMetric(
			c.WorkingSetPrivate,
			prometheus.GaugeValue,
			process.WorkingSetPrivate,
			processName,
			pid,
			cpid,
		)

		ch <- prometheus.MustNewConstMetric(
			c.WorkingSetPeak,
			prometheus.GaugeValue,
			process.WorkingSetPeak,
			processName,
			pid,
			cpid,
		)

		ch <- prometheus.MustNewConstMetric(
			c.WorkingSet,
			prometheus.GaugeValue,
			process.WorkingSet,
			processName,
			pid,
			cpid,
		)
	}

	return nil
}
