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
	"github.com/prometheus-community/windows_exporter/config"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/yusufpapurcu/wmi"
)

const (
	FlagProcessOldExclude = "collector.process.blacklist"
	FlagProcessOldInclude = "collector.process.whitelist"

	FlagProcessExclude = "collector.process.exclude"
	FlagProcessInclude = "collector.process.include"
)

type ProcessDef struct {
	Name         string
	Include      *regexp.Regexp
	Exclude      *regexp.Regexp
	CustomLabels map[string]string
	Count        int
}

var (
	processOldInclude *string
	processOldExclude *string

	processInclude *string
	processExclude *string

	processIncludeSet bool
	processExcludeSet bool

	enableWorkerProcess *bool
	processes           = make(map[string]*ProcessDef)
	processes_labels    = make([]string, 0)
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
	ProcessGroupCount *prometheus.Desc

	Processes       map[string]*ProcessDef
	ProcessesLabels []string
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

func ProcessBuildHook() map[string]config.CfgHook {
	config_hooks := &config.CfgHook{
		ConfigAttrs: []string{"collector", "process", "processes"},
		Hook:        ProcessBuildMap,
	}
	entry := make(map[string]config.CfgHook)
	entry["processes-list"] = *config_hooks
	return entry
}

func ProcessBuildMap(logger log.Logger, data interface{}) map[string]string {
	ret := make(map[string]string)
	switch typed := data.(type) {
	case map[interface{}]interface{}:
		ret = flatten(data)

	// form is a dict of service's name maybe with custom labels
	case map[string]interface{}:
		for name, raw_labels := range typed {
			var (
				include, exclude *regexp.Regexp
			)
			labels := flatten(raw_labels)
			if inc, ok := labels["include"]; ok {
				if inc != "" {
					include = regexp.MustCompile(fmt.Sprintf("^(?:%s)$", inc))
				}
				delete(labels, "include")
			}
			if exc, ok := labels["exclude"]; ok {
				if exc != "" {
					exclude = regexp.MustCompile(fmt.Sprintf("^(?:%s)$", exc))
				}
				delete(labels, "exclude")
			}
			process := &ProcessDef{
				Name:         name,
				Include:      include,
				Exclude:      exclude,
				CustomLabels: labels,
			}

			processes[strings.ToLower(process.Name)] = process
		}
	default:
		ret["unknown_parameter"] = "1"
	}

	// build a list of all custom labels for each component
	exists := make(map[string]bool)
	for _, proc := range processes {
		// fill the exists map with all labels' names
		for name := range proc.CustomLabels {
			exists[name] = true
		}
	}
	// check custom labels: each name must be present for each component
	for _, proc := range processes {
		for name := range proc.CustomLabels {
			if _, ok := exists[name]; !ok {
				_ = level.Warn(logger).Log("errmsg", "label: %s not present for process pattern '%s'", name, proc.Name)
				exists[name] = false
			}
		}
	}
	processes_labels = make([]string, 0, len(exists))
	for name := range exists {
		processes_labels = append(processes_labels, name)
	}
	return ret
}

// check if process name is matching one of include or exclude pattern of processDef
func (c *processCollector) Match(proc_name string) (bool, *ProcessDef) {
	var (
		ok   = false
		proc *ProcessDef
	)
	for _, p := range c.Processes {
		// when include pattern is defined check if proc_name match include pattern
		if p.Include != nil && p.Include.MatchString(proc_name) {
			ok = true
			proc = p
		}

		if p.Exclude != nil {
			if p.Exclude.MatchString(proc_name) {
				ok = false
				proc = nil
			}
		} else if ok {
			// exclude not defined and proc is in include we got procDef
			break
		}
	}
	return ok, proc
}
func (pdef *ProcessDef) Debug(logger log.Logger) {
	inc_pat := "_"
	exc_pat := "_"
	if pdef.Include != nil {
		inc_pat = pdef.Include.String()
	}
	if pdef.Exclude != nil {
		exc_pat = pdef.Exclude.String()
	}
	_ = level.Debug(logger).Log("msg", fmt.Sprintf("process Matcher '%s' include: '%s' exclude: '%s'", pdef.Name, inc_pat, exc_pat))
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
	if *processInclude == ".*" && *processExclude == "" && len(processes) <= 0 {
		_ = level.Warn(logger).Log("msg", "No filters specified for process collector. This will generate a very large number of metrics!")
	} else {
		if *processInclude != ".*" || *processExclude != "" {
			var include, exclude *regexp.Regexp
			if *processInclude != "" {
				include = regexp.MustCompile(fmt.Sprintf("^(?:%s)$", *processInclude))
			}
			if *processExclude != "" {
				exclude = regexp.MustCompile(fmt.Sprintf("^(?:%s)$", *processExclude))
			}
			processes["default"] = &ProcessDef{
				Name:         "default",
				Include:      include,
				Exclude:      exclude,
				CustomLabels: nil,
			}
			pdef := processes["default"]
			pdef.Debug(logger)
		} else {
			if len(processes) > 0 {
				for _, pdef := range processes {
					pdef.Debug(logger)
				}
			}
		}
	}

	var var_labels [4][]string
	var_labels[0] = []string{"process", "process_id", "creating_process_id"}
	var_labels[1] = []string{"process", "process_id", "creating_process_id", "mode"}
	var_labels[2] = []string{"process", "process_id", "creating_process_id", "pool"}
	var_labels[3] = []string{"group"}
	if len(processes_labels) > 0 {
		for _, label := range processes_labels {
			for idx, metric_label := range var_labels {
				var_labels[idx] = append(metric_label, label)
			}
		}
	}
	return &processCollector{
		logger: logger,
		StartTime: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "start_time"),
			"Time of process start.",
			var_labels[0],
			nil,
		),
		CPUTimeTotal: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "cpu_time_total"),
			"Returns elapsed time that all of the threads of this process used the processor to execute instructions by mode (privileged, user).",
			var_labels[1],
			nil,
		),
		HandleCount: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "handles"),
			"Total number of handles the process has open. This number is the sum of the handles currently open by each thread in the process.",
			var_labels[0],
			nil,
		),
		IOBytesTotal: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "io_bytes_total"),
			"Bytes issued to I/O operations in different modes (read, write, other).",
			var_labels[1],
			nil,
		),
		IOOperationsTotal: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "io_operations_total"),
			"I/O operations issued in different modes (read, write, other).",
			var_labels[1],
			nil,
		),
		PageFaultsTotal: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "page_faults_total"),
			"Page faults by the threads executing in this process.",
			var_labels[0],
			nil,
		),
		PageFileBytes: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "page_file_bytes"),
			"Current number of bytes this process has used in the paging file(s).",
			var_labels[0],
			nil,
		),
		PoolBytes: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "pool_bytes"),
			"Pool Bytes is the last observed number of bytes in the paged or nonpaged pool.",
			var_labels[2],
			nil,
		),
		PriorityBase: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "priority_base"),
			"Current base priority of this process. Threads within a process can raise and lower their own base priority relative to the process base priority of the process.",
			var_labels[0],
			nil,
		),
		PrivateBytes: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "private_bytes"),
			"Current number of bytes this process has allocated that cannot be shared with other processes.",
			var_labels[0],
			nil,
		),
		ThreadCount: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "threads"),
			"Number of threads currently active in this process.",
			var_labels[0],
			nil,
		),
		VirtualBytes: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "virtual_bytes"),
			"Current size, in bytes, of the virtual address space that the process is using.",
			var_labels[0],
			nil,
		),
		WorkingSetPrivate: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "working_set_private_bytes"),
			"Size of the working set, in bytes, that is use for this process only and not shared nor shareable by other processes.",
			var_labels[0],
			nil,
		),
		WorkingSetPeak: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "working_set_peak_bytes"),
			"Maximum size, in bytes, of the Working Set of this process at any point in time. The Working Set is the set of memory pages touched recently by the threads in the process.",
			var_labels[0],
			nil,
		),
		WorkingSet: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "working_set_bytes"),
			"Maximum number of bytes in the working set of this process at any point in time. The working set is the set of memory pages touched recently by the threads in the process.",
			var_labels[0],
			nil,
		),
		ProcessGroupCount: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "group_count"),
			"Number of processes found for the matching patterns.",
			var_labels[3],
			nil,
		),

		Processes:       processes,
		ProcessesLabels: processes_labels,
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

	var (
		dst_wp        []WorkerProcess
		pdef          *ProcessDef
		custom_labels []string
	)
	if *enableWorkerProcess {
		q_wp := queryAll(&dst_wp, c.logger)
		if err := wmi.QueryNamespace(q_wp, &dst_wp, "root\\WebAdministration"); err != nil {
			_ = level.Debug(c.logger).Log(fmt.Sprintf("Could not query WebAdministration namespace for IIS worker processes: %v. Skipping\n", err))
		}
	}

	// we have 3 models for labels
	//var_labels[0] = []string{"process", "process_id", "creating_process_id"}
	//var_labels[1] = []string{"process", "process_id", "creating_process_id", "mode"}
	//var_labels[2] = []string{"process", "process_id", "creating_process_id", "pool"}
	// for processs_group_count
	//var_labels[3] = []string{"group"}

	// reset counter for processes_group
	for _, procdef := range c.Processes {
		procdef.Count = 0
	}

	for _, process := range data {
		if process.Name == "_Total" {
			continue
		}
		if ok, procdef := c.Match(process.Name); !ok {
			continue
		} else {
			pdef = procdef
			// increment group counter
			if pdef != nil {
				pdef.Count++
			}
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
		labels_value := make([]string, 3)
		labels_value[0] = processName
		labels_value[1] = pid
		labels_value[2] = cpid
		if pdef != nil && len(c.ProcessesLabels) > 0 {
			custom_labels = make([]string, len(c.ProcessesLabels))
			// we find the service name in service definition list
			for idx, label := range c.ProcessesLabels {
				if val, tst := pdef.CustomLabels[label]; tst {
					custom_labels[idx] = val
				}
			}

		} else {
			// we don't find the service name !?!? not possible but...
			for idx := range services_labels {
				custom_labels[idx] = ""
			}
		}

		labels := append(labels_value, custom_labels...)

		// metrics with var_labels[0] format
		ch <- prometheus.MustNewConstMetric(
			c.StartTime,
			prometheus.GaugeValue,
			process.ElapsedTime,
			labels...,
		)

		ch <- prometheus.MustNewConstMetric(
			c.HandleCount,
			prometheus.GaugeValue,
			process.HandleCount,
			labels...,
		)

		ch <- prometheus.MustNewConstMetric(
			c.PageFaultsTotal,
			prometheus.CounterValue,
			process.PageFaultsPerSec,
			labels...,
		)

		ch <- prometheus.MustNewConstMetric(
			c.PageFileBytes,
			prometheus.GaugeValue,
			process.PageFileBytes,
			labels...,
		)

		ch <- prometheus.MustNewConstMetric(
			c.PriorityBase,
			prometheus.GaugeValue,
			process.PriorityBase,
			labels...,
		)

		ch <- prometheus.MustNewConstMetric(
			c.PrivateBytes,
			prometheus.GaugeValue,
			process.PrivateBytes,
			labels...,
		)

		ch <- prometheus.MustNewConstMetric(
			c.ThreadCount,
			prometheus.GaugeValue,
			process.ThreadCount,
			labels...,
		)

		ch <- prometheus.MustNewConstMetric(
			c.VirtualBytes,
			prometheus.GaugeValue,
			process.VirtualBytes,
			labels...,
		)

		ch <- prometheus.MustNewConstMetric(
			c.WorkingSetPrivate,
			prometheus.GaugeValue,
			process.WorkingSetPrivate,
			labels...,
		)

		ch <- prometheus.MustNewConstMetric(
			c.WorkingSetPeak,
			prometheus.GaugeValue,
			process.WorkingSetPeak,
			labels...,
		)

		ch <- prometheus.MustNewConstMetric(
			c.WorkingSet,
			prometheus.GaugeValue,
			process.WorkingSet,
			labels...,
		)

		// metrics with var_labels[1] format
		labels_value = append(labels_value, "privileged")
		labels = append(labels_value, custom_labels...)
		ch <- prometheus.MustNewConstMetric(
			c.CPUTimeTotal,
			prometheus.CounterValue,
			process.PercentPrivilegedTime,
			labels...,
		)

		labels_value[3] = "user"
		labels = append(labels_value, custom_labels...)
		ch <- prometheus.MustNewConstMetric(
			c.CPUTimeTotal,
			prometheus.CounterValue,
			process.PercentUserTime,
			labels...,
		)

		labels_value[3] = "other"
		labels = append(labels_value, custom_labels...)
		ch <- prometheus.MustNewConstMetric(
			c.IOBytesTotal,
			prometheus.CounterValue,
			process.IOOtherBytesPerSec,
			labels...,
		)

		ch <- prometheus.MustNewConstMetric(
			c.IOOperationsTotal,
			prometheus.CounterValue,
			process.IOOtherOperationsPerSec,
			labels...,
		)

		labels_value[3] = "read"
		labels = append(labels_value, custom_labels...)
		ch <- prometheus.MustNewConstMetric(
			c.IOBytesTotal,
			prometheus.CounterValue,
			process.IOReadBytesPerSec,
			labels...,
		)

		ch <- prometheus.MustNewConstMetric(
			c.IOOperationsTotal,
			prometheus.CounterValue,
			process.IOReadOperationsPerSec,
			labels...,
		)

		labels_value[3] = "write"
		labels = append(labels_value, custom_labels...)
		ch <- prometheus.MustNewConstMetric(
			c.IOBytesTotal,
			prometheus.CounterValue,
			process.IOWriteBytesPerSec,
			labels...,
		)

		ch <- prometheus.MustNewConstMetric(
			c.IOOperationsTotal,
			prometheus.CounterValue,
			process.IOWriteOperationsPerSec,
			labels...,
		)

		labels_value[3] = "nonpaged"
		labels = append(labels_value, custom_labels...)
		ch <- prometheus.MustNewConstMetric(
			c.PoolBytes,
			prometheus.GaugeValue,
			process.PoolNonpagedBytes,
			labels...,
		)

		labels_value[3] = "paged"
		labels = append(labels_value, custom_labels...)
		ch <- prometheus.MustNewConstMetric(
			c.PoolBytes,
			prometheus.GaugeValue,
			process.PoolPagedBytes,
			labels...,
		)

	}
	// add metric for processes_group
	for _, procdef := range c.Processes {
		labels_value := []string{procdef.Name}

		custom_labels = make([]string, len(c.ProcessesLabels))
		// we find the service name in service definition list
		for idx, label := range c.ProcessesLabels {
			if val, tst := procdef.CustomLabels[label]; tst {
				custom_labels[idx] = val
			}
		}
		labels := append(labels_value, custom_labels...)
		ch <- prometheus.MustNewConstMetric(
			c.ProcessGroupCount,
			prometheus.GaugeValue,
			float64(procdef.Count),
			labels...,
		)
	}

	return nil
}
