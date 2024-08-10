//go:build windows

package process

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"syscall"

	"github.com/alecthomas/kingpin/v2"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/prometheus-community/windows_exporter/pkg/perflib"
	"github.com/prometheus-community/windows_exporter/pkg/types"
	"github.com/prometheus-community/windows_exporter/pkg/utils"
	"github.com/prometheus-community/windows_exporter/pkg/wmi"
	"github.com/prometheus/client_golang/prometheus"
	"golang.org/x/sys/windows"
)

const Name = "process"

type Config struct {
	ProcessInclude      string `yaml:"process_include"`
	ProcessExclude      string `yaml:"process_exclude"`
	EnableWorkerProcess bool   `yaml:"enable_iis_worker_process"`
	EnableReportOwner   bool   `yaml:"enable_report_owner"`
}

var ConfigDefaults = Config{
	ProcessInclude:      ".+",
	ProcessExclude:      "",
	EnableWorkerProcess: false,
	EnableReportOwner:   false,
}

type Collector struct {
	logger log.Logger

	enableWorkerProcess *bool
	enableReportOwner   *bool

	processInclude *string
	processExclude *string

	processIncludePattern *regexp.Regexp
	processExcludePattern *regexp.Regexp

	lookupCache map[string]string

	cpuTimeTotal      *prometheus.Desc
	handleCount       *prometheus.Desc
	ioBytesTotal      *prometheus.Desc
	ioOperationsTotal *prometheus.Desc
	pageFaultsTotal   *prometheus.Desc
	pageFileBytes     *prometheus.Desc
	poolBytes         *prometheus.Desc
	priorityBase      *prometheus.Desc
	privateBytes      *prometheus.Desc
	startTime         *prometheus.Desc
	threadCount       *prometheus.Desc
	virtualBytes      *prometheus.Desc
	workingSet        *prometheus.Desc
	workingSetPeak    *prometheus.Desc
	workingSetPrivate *prometheus.Desc
}

func New(logger log.Logger, config *Config) *Collector {
	if config == nil {
		config = &ConfigDefaults
	}

	c := &Collector{
		processExclude:      &config.ProcessExclude,
		processInclude:      &config.ProcessInclude,
		enableWorkerProcess: &config.EnableWorkerProcess,
	}
	c.SetLogger(logger)

	return c
}

func NewWithFlags(app *kingpin.Application) *Collector {
	c := &Collector{
		processInclude: app.Flag(
			"collector.process.include",
			"Regexp of processes to include. Process name must both match include and not match exclude to be included.",
		).Default(ConfigDefaults.ProcessInclude).String(),

		processExclude: app.Flag(
			"collector.process.exclude",
			"Regexp of processes to exclude. Process name must both match include and not match exclude to be included.",
		).Default(ConfigDefaults.ProcessExclude).String(),

		enableWorkerProcess: app.Flag(
			"collector.process.iis",
			"Enable IIS worker process name queries. May cause the collector to leak memory.",
		).Default("false").Bool(),

		enableReportOwner: app.Flag(
			"collector.process.report-owner",
			"Enable reporting of process owner.",
		).Default("false").Bool(),
	}
	return c
}

func (c *Collector) GetName() string {
	return Name
}

func (c *Collector) SetLogger(logger log.Logger) {
	c.logger = log.With(logger, "collector", Name)
}

func (c *Collector) GetPerfCounter() ([]string, error) {
	return []string{"Process"}, nil
}

func (c *Collector) Close() error {
	return nil
}

func (c *Collector) Build() error {
	if c.processInclude != nil && *c.processInclude == ".*" && utils.IsEmpty(c.processExclude) {
		_ = level.Warn(c.logger).Log("msg", "No filters specified for process collector. This will generate a very large number of metrics!")
	}

	commonLabels := make([]string, 0)
	if *c.enableReportOwner {
		commonLabels = []string{"owner"}
	}

	c.startTime = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "start_time"),
		"Time of process start.",
		append(commonLabels, "process", "process_id", "creating_process_id"),
		nil,
	)
	c.cpuTimeTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "cpu_time_total"),
		"Returns elapsed time that all of the threads of this process used the processor to execute instructions by mode (privileged, user).",
		append(commonLabels, "process", "process_id", "creating_process_id", "mode"),
		nil,
	)
	c.handleCount = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "handles"),
		"Total number of handles the process has open. This number is the sum of the handles currently open by each thread in the process.",
		append(commonLabels, "process", "process_id", "creating_process_id"),
		nil,
	)
	c.ioBytesTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "io_bytes_total"),
		"Bytes issued to I/O operations in different modes (read, write, other).",
		append(commonLabels, "process", "process_id", "creating_process_id", "mode"),
		nil,
	)
	c.ioOperationsTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "io_operations_total"),
		"I/O operations issued in different modes (read, write, other).",
		append(commonLabels, "process", "process_id", "creating_process_id", "mode"),
		nil,
	)
	c.pageFaultsTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "page_faults_total"),
		"Page faults by the threads executing in this process.",
		append(commonLabels, "process", "process_id", "creating_process_id"),
		nil,
	)
	c.pageFileBytes = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "page_file_bytes"),
		"Current number of bytes this process has used in the paging file(s).",
		append(commonLabels, "process", "process_id", "creating_process_id"),
		nil,
	)
	c.poolBytes = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "pool_bytes"),
		"Pool Bytes is the last observed number of bytes in the paged or nonpaged pool.",
		append(commonLabels, "process", "process_id", "creating_process_id", "pool"),
		nil,
	)
	c.priorityBase = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "priority_base"),
		"Current base priority of this process. Threads within a process can raise and lower their own base priority relative to the process base priority of the process.",
		append(commonLabels, "process", "process_id", "creating_process_id"),
		nil,
	)
	c.privateBytes = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "private_bytes"),
		"Current number of bytes this process has allocated that cannot be shared with other processes.",
		append(commonLabels, "process", "process_id", "creating_process_id"),
		nil,
	)
	c.threadCount = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "threads"),
		"Number of threads currently active in this process.",
		append(commonLabels, "process", "process_id", "creating_process_id"),
		nil,
	)
	c.virtualBytes = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "virtual_bytes"),
		"Current size, in bytes, of the virtual address space that the process is using.",
		append(commonLabels, "process", "process_id", "creating_process_id"),
		nil,
	)
	c.workingSetPrivate = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "working_set_private_bytes"),
		"Size of the working set, in bytes, that is use for this process only and not shared nor shareable by other processes.",
		append(commonLabels, "process", "process_id", "creating_process_id"),
		nil,
	)
	c.workingSetPeak = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "working_set_peak_bytes"),
		"Maximum size, in bytes, of the Working Set of this process at any point in time. The Working Set is the set of memory pages touched recently by the threads in the process.",
		append(commonLabels, "process", "process_id", "creating_process_id"),
		nil,
	)
	c.workingSet = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "working_set_bytes"),
		"Maximum number of bytes in the working set of this process at any point in time. The working set is the set of memory pages touched recently by the threads in the process.",
		append(commonLabels, "process", "process_id", "creating_process_id"),
		nil,
	)

	c.lookupCache = make(map[string]string)

	var err error

	c.processIncludePattern, err = regexp.Compile(fmt.Sprintf("^(?:%s)$", *c.processInclude))
	if err != nil {
		return err
	}

	c.processExcludePattern, err = regexp.Compile(fmt.Sprintf("^(?:%s)$", *c.processExclude))
	if err != nil {
		return err
	}

	return nil
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

func (c *Collector) Collect(ctx *types.ScrapeContext, ch chan<- prometheus.Metric) error {
	data := make([]perflibProcess, 0)
	err := perflib.UnmarshalObject(ctx.PerfObjects["Process"], &data, c.logger)
	if err != nil {
		return err
	}

	var dst_wp []WorkerProcess
	if *c.enableWorkerProcess {
		q_wp := wmi.QueryAll(&dst_wp, c.logger)
		if err := wmi.QueryNamespace(q_wp, &dst_wp, "root\\WebAdministration"); err != nil {
			_ = level.Debug(c.logger).Log("msg", "Could not query WebAdministration namespace for IIS worker processes", "err", err)
		}
	}

	var owner string

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

		if *c.enableWorkerProcess {
			for _, wp := range dst_wp {
				if wp.ProcessId == uint64(process.IDProcess) {
					processName = strings.Join([]string{processName, wp.AppPoolName}, "_")
					break
				}
			}
		}

		labels := make([]string, 0, 4)

		if *c.enableReportOwner {
			owner, err = c.getProcessOwner(int(process.IDProcess))
			if err != nil {
				owner = "unknown"
			}

			labels = []string{owner}
		}

		labels = append(labels, processName, pid, cpid)

		ch <- prometheus.MustNewConstMetric(
			c.startTime,
			prometheus.GaugeValue,
			process.ElapsedTime,
			labels...,
		)

		ch <- prometheus.MustNewConstMetric(
			c.handleCount,
			prometheus.GaugeValue,
			process.HandleCount,
			labels...,
		)

		ch <- prometheus.MustNewConstMetric(
			c.cpuTimeTotal,
			prometheus.CounterValue,
			process.PercentPrivilegedTime,
			append(labels, "privileged")...,
		)

		ch <- prometheus.MustNewConstMetric(
			c.cpuTimeTotal,
			prometheus.CounterValue,
			process.PercentUserTime,
			append(labels, "user")...,
		)

		ch <- prometheus.MustNewConstMetric(
			c.ioBytesTotal,
			prometheus.CounterValue,
			process.IOOtherBytesPerSec,
			append(labels, "other")...,
		)

		ch <- prometheus.MustNewConstMetric(
			c.ioOperationsTotal,
			prometheus.CounterValue,
			process.IOOtherOperationsPerSec,
			append(labels, "other")...,
		)

		ch <- prometheus.MustNewConstMetric(
			c.ioBytesTotal,
			prometheus.CounterValue,
			process.IOReadBytesPerSec,
			append(labels, "read")...,
		)

		ch <- prometheus.MustNewConstMetric(
			c.ioOperationsTotal,
			prometheus.CounterValue,
			process.IOReadOperationsPerSec,
			append(labels, "read")...,
		)

		ch <- prometheus.MustNewConstMetric(
			c.ioBytesTotal,
			prometheus.CounterValue,
			process.IOWriteBytesPerSec,
			append(labels, "write")...,
		)

		ch <- prometheus.MustNewConstMetric(
			c.ioOperationsTotal,
			prometheus.CounterValue,
			process.IOWriteOperationsPerSec,
			append(labels, "write")...,
		)

		ch <- prometheus.MustNewConstMetric(
			c.pageFaultsTotal,
			prometheus.CounterValue,
			process.PageFaultsPerSec,
			labels...,
		)

		ch <- prometheus.MustNewConstMetric(
			c.pageFileBytes,
			prometheus.GaugeValue,
			process.PageFileBytes,
			labels...,
		)

		ch <- prometheus.MustNewConstMetric(
			c.poolBytes,
			prometheus.GaugeValue,
			process.PoolNonpagedBytes,
			append(labels, "nonpaged")...,
		)

		ch <- prometheus.MustNewConstMetric(
			c.poolBytes,
			prometheus.GaugeValue,
			process.PoolPagedBytes,
			append(labels, "paged")...,
		)

		ch <- prometheus.MustNewConstMetric(
			c.priorityBase,
			prometheus.GaugeValue,
			process.PriorityBase,
			labels...,
		)

		ch <- prometheus.MustNewConstMetric(
			c.privateBytes,
			prometheus.GaugeValue,
			process.PrivateBytes,
			labels...,
		)

		ch <- prometheus.MustNewConstMetric(
			c.threadCount,
			prometheus.GaugeValue,
			process.ThreadCount,
			labels...,
		)

		ch <- prometheus.MustNewConstMetric(
			c.virtualBytes,
			prometheus.GaugeValue,
			process.VirtualBytes,
			labels...,
		)

		ch <- prometheus.MustNewConstMetric(
			c.workingSetPrivate,
			prometheus.GaugeValue,
			process.WorkingSetPrivate,
			labels...,
		)

		ch <- prometheus.MustNewConstMetric(
			c.workingSetPeak,
			prometheus.GaugeValue,
			process.WorkingSetPeak,
			labels...,
		)

		ch <- prometheus.MustNewConstMetric(
			c.workingSet,
			prometheus.GaugeValue,
			process.WorkingSet,
			labels...,
		)
	}

	return nil
}

// ref: https://github.com/microsoft/hcsshim/blob/8beabacfc2d21767a07c20f8dd5f9f3932dbf305/internal/uvm/stats.go#L25
func (c *Collector) getProcessOwner(pid int) (string, error) {
	p, err := windows.OpenProcess(windows.PROCESS_QUERY_LIMITED_INFORMATION, false, uint32(pid))
	if errors.Is(err, syscall.Errno(0x57)) { // invalid parameter, for PIDs that don't exist
		return "", errors.New("process not found")
	}

	if err != nil {
		return "", fmt.Errorf("OpenProcess: %T %w", err, err)
	}

	defer windows.Close(p)

	var tok windows.Token
	if err = windows.OpenProcessToken(p, windows.TOKEN_QUERY, &tok); err != nil {
		return "", fmt.Errorf("OpenProcessToken: %w", err)
	}

	tokenUser, err := tok.GetTokenUser()
	if err != nil {
		return "", fmt.Errorf("GetTokenUser: %w", err)
	}

	sid := tokenUser.User.Sid.String()
	if owner, ok := c.lookupCache[sid]; ok {
		return owner, nil
	}

	account, domain, _, err := tokenUser.User.Sid.LookupAccount("")
	if err != nil {
		c.lookupCache[sid] = sid
	} else {
		c.lookupCache[sid] = fmt.Sprintf(`%s\%s`, account, domain)
	}

	return c.lookupCache[sid], nil
}
