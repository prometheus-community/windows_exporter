//go:build windows

package process

import (
	"errors"
	"fmt"
	"log/slog"
	"regexp"
	"strconv"
	"strings"
	"unsafe"

	"github.com/alecthomas/kingpin/v2"
	"github.com/prometheus-community/windows_exporter/internal/perfdata"
	"github.com/prometheus-community/windows_exporter/internal/perflib"
	"github.com/prometheus-community/windows_exporter/internal/types"
	"github.com/prometheus-community/windows_exporter/internal/utils"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/yusufpapurcu/wmi"
	"golang.org/x/sys/windows"
)

const Name = "process"

type Config struct {
	ProcessInclude       *regexp.Regexp `yaml:"process_include"`
	ProcessExclude       *regexp.Regexp `yaml:"process_exclude"`
	EnableWorkerProcess  bool           `yaml:"enable_iis_worker_process"` //nolint:tagliatelle
	PerfCounterInstances []string       `yaml:"perf_counter_instances"`
}

var ConfigDefaults = Config{
	ProcessInclude:       types.RegExpAny,
	ProcessExclude:       types.RegExpEmpty,
	EnableWorkerProcess:  false,
	PerfCounterInstances: []string{"*"},
}

type Collector struct {
	config    Config
	wmiClient *wmi.Client

	perfDataCollector *perfdata.Collector

	lookupCache map[string]string

	info              *prometheus.Desc
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

func New(config *Config) *Collector {
	if config == nil {
		config = &ConfigDefaults
	}

	if config.ProcessExclude == nil {
		config.ProcessExclude = ConfigDefaults.ProcessExclude
	}

	if config.ProcessInclude == nil {
		config.ProcessInclude = ConfigDefaults.ProcessInclude
	}

	if config.PerfCounterInstances == nil {
		config.PerfCounterInstances = ConfigDefaults.PerfCounterInstances
	}

	c := &Collector{
		config: *config,
	}

	return c
}

func NewWithFlags(app *kingpin.Application) *Collector {
	c := &Collector{
		config: ConfigDefaults,
	}

	var processExclude, processInclude, perfCounterInstances string

	app.Flag(
		"collector.process.exclude",
		"Regexp of processes to exclude. Process name must both match include and not match exclude to be included.",
	).Default(c.config.ProcessExclude.String()).StringVar(&processExclude)

	app.Flag(
		"collector.process.include",
		"Regexp of processes to include. Process name must both match include and not match exclude to be included.",
	).Default(c.config.ProcessInclude.String()).StringVar(&processInclude)

	app.Flag(
		"collector.process.iis",
		"Enable IIS worker process name queries. May cause the collector to leak memory.",
	).Default(strconv.FormatBool(c.config.EnableWorkerProcess)).BoolVar(&c.config.EnableWorkerProcess)

	app.Flag(
		"collector.process.perf-counter-instance",
		"Advanced: List of process performance counter instances to query. If not set, all instances are queried.",
	).Default(strings.Join(c.config.PerfCounterInstances, ",")).StringVar(&perfCounterInstances)

	app.Action(func(*kingpin.ParseContext) error {
		c.config.PerfCounterInstances = strings.Split(perfCounterInstances, ",")

		var err error

		c.config.ProcessExclude, err = regexp.Compile(fmt.Sprintf("^(?:%s)$", processExclude))
		if err != nil {
			return fmt.Errorf("collector.process.exclude: %w", err)
		}

		c.config.ProcessInclude, err = regexp.Compile(fmt.Sprintf("^(?:%s)$", processInclude))
		if err != nil {
			return fmt.Errorf("collector.process.include: %w", err)
		}

		return nil
	})

	return c
}

func (c *Collector) GetName() string {
	return Name
}

func (c *Collector) GetPerfCounter(_ *slog.Logger) ([]string, error) {
	if utils.PDHEnabled() {
		return []string{}, nil
	}

	return []string{"Process"}, nil
}

func (c *Collector) Close(_ *slog.Logger) error {
	return nil
}

func (c *Collector) Build(logger *slog.Logger, wmiClient *wmi.Client) error {
	logger = logger.With(slog.String("collector", Name))

	if wmiClient == nil || wmiClient.SWbemServicesClient == nil {
		return errors.New("wmiClient or SWbemServicesClient is nil")
	}

	c.wmiClient = wmiClient

	if utils.PDHEnabled() {
		counters := []string{
			percentProcessorTime,
			percentPrivilegedTime,
			percentUserTime,
			creatingProcessID,
			elapsedTime,
			handleCount,
			ioDataBytesPerSec,
			ioDataOperationsPerSec,
			ioOtherBytesPerSec,
			ioOtherOperationsPerSec,
			ioReadBytesPerSec,
			ioReadOperationsPerSec,
			ioWriteBytesPerSec,
			ioWriteOperationsPerSec,
			pageFaultsPerSec,
			pageFileBytesPeak,
			pageFileBytes,
			poolNonPagedBytes,
			poolPagedBytes,
			processID,
			priorityBase,
			privateBytes,
			threadCount,
			virtualBytesPeak,
			virtualBytes,
			workingSetPrivate,
			workingSetPeak,
			workingSet,
		}

		var err error

		c.perfDataCollector, err = perfdata.NewCollector("Process V2", c.config.PerfCounterInstances, counters)
		if errors.Is(err, perfdata.NewPdhError(perfdata.PdhNoData)) {
			counters = []string{
				percentProcessorTime,
				percentPrivilegedTime,
				percentUserTime,
				creatingProcessID,
				elapsedTime,
				handleCount,
				idProcess,
				ioDataBytesPerSec,
				ioDataOperationsPerSec,
				ioOtherBytesPerSec,
				ioOtherOperationsPerSec,
				ioReadBytesPerSec,
				ioReadOperationsPerSec,
				ioWriteBytesPerSec,
				ioWriteOperationsPerSec,
				pageFaultsPerSec,
				pageFileBytesPeak,
				pageFileBytes,
				poolNonPagedBytes,
				poolPagedBytes,
				priorityBase,
				privateBytes,
				threadCount,
				virtualBytesPeak,
				virtualBytes,
				workingSetPrivate,
				workingSetPeak,
				workingSet,
			}

			c.perfDataCollector, err = perfdata.NewCollector("Process", c.config.PerfCounterInstances, counters)
		}

		if err != nil {
			return fmt.Errorf("failed to create Process collector: %w", err)
		}
	}

	if c.config.ProcessInclude.String() == "^(?:.*)$" && c.config.ProcessExclude.String() == "^(?:)$" {
		logger.Warn("No filters specified for process collector. This will generate a very large number of metrics!")
	}

	c.info = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "info"),
		"Process information.",
		[]string{"process", "process_id", "creating_process_id", "process_group_id", "owner", "cmdline"},
		nil,
	)

	c.startTime = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "start_time"),
		"Time of process start.",
		[]string{"process", "process_id"},
		nil,
	)
	c.cpuTimeTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "cpu_time_total"),
		"Returns elapsed time that all of the threads of this process used the processor to execute instructions by mode (privileged, user).",
		[]string{"process", "process_id", "mode"},
		nil,
	)
	c.handleCount = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "handles"),
		"Total number of handles the process has open. This number is the sum of the handles currently open by each thread in the process.",
		[]string{"process", "process_id"},
		nil,
	)
	c.ioBytesTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "io_bytes_total"),
		"Bytes issued to I/O operations in different modes (read, write, other).",
		[]string{"process", "process_id", "mode"},
		nil,
	)
	c.ioOperationsTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "io_operations_total"),
		"I/O operations issued in different modes (read, write, other).",
		[]string{"process", "process_id", "mode"},
		nil,
	)
	c.pageFaultsTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "page_faults_total"),
		"Page faults by the threads executing in this process.",
		[]string{"process", "process_id"},
		nil,
	)
	c.pageFileBytes = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "page_file_bytes"),
		"Current number of bytes this process has used in the paging file(s).",
		[]string{"process", "process_id"},
		nil,
	)
	c.poolBytes = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "pool_bytes"),
		"Pool Bytes is the last observed number of bytes in the paged or nonpaged pool.",
		[]string{"process", "process_id", "pool"},
		nil,
	)
	c.priorityBase = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "priority_base"),
		"Current base priority of this process. Threads within a process can raise and lower their own base priority relative to the process base priority of the process.",
		[]string{"process", "process_id"},
		nil,
	)
	c.privateBytes = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "private_bytes"),
		"Current number of bytes this process has allocated that cannot be shared with other processes.",
		[]string{"process", "process_id"},
		nil,
	)
	c.threadCount = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "threads"),
		"Number of threads currently active in this process.",
		[]string{"process", "process_id"},
		nil,
	)
	c.virtualBytes = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "virtual_bytes"),
		"Current size, in bytes, of the virtual address space that the process is using.",
		[]string{"process", "process_id"},
		nil,
	)
	c.workingSetPrivate = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "working_set_private_bytes"),
		"Size of the working set, in bytes, that is use for this process only and not shared nor shareable by other processes.",
		[]string{"process", "process_id"},
		nil,
	)
	c.workingSetPeak = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "working_set_peak_bytes"),
		"Maximum size, in bytes, of the Working Set of this process at any point in time. The Working Set is the set of memory pages touched recently by the threads in the process.",
		[]string{"process", "process_id"},
		nil,
	)
	c.workingSet = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "working_set_bytes"),
		"Maximum number of bytes in the working set of this process at any point in time. The working set is the set of memory pages touched recently by the threads in the process.",
		[]string{"process", "process_id"},
		nil,
	)

	c.lookupCache = make(map[string]string)

	return nil
}

type WorkerProcess struct {
	AppPoolName string
	ProcessId   uint64
}

func (c *Collector) Collect(ctx *types.ScrapeContext, logger *slog.Logger, ch chan<- prometheus.Metric) error {
	if utils.PDHEnabled() {
		return c.collectPDH(logger, ch)
	}

	logger = logger.With(slog.String("collector", Name))
	if err := c.collect(ctx, logger, ch); err != nil {
		logger.Error("failed collecting metrics",
			slog.Any("err", err),
		)

		return err
	}

	return nil
}

func (c *Collector) collect(ctx *types.ScrapeContext, logger *slog.Logger, ch chan<- prometheus.Metric) error {
	data := make([]perflibProcess, 0)

	err := perflib.UnmarshalObject(ctx.PerfObjects["Process"], &data, logger)
	if err != nil {
		return err
	}

	var workerProcesses []WorkerProcess
	if c.config.EnableWorkerProcess {
		if err := c.wmiClient.Query("SELECT * FROM WorkerProcess", &workerProcesses, nil, "root\\WebAdministration"); err != nil {
			logger.Debug("Could not query WebAdministration namespace for IIS worker processes",
				slog.Any("err", err),
			)
		}
	}

	for _, process := range data {
		if process.Name == "_Total" ||
			c.config.ProcessExclude.MatchString(process.Name) ||
			!c.config.ProcessInclude.MatchString(process.Name) {
			continue
		}

		// Duplicate processes are suffixed #, and an index number. Remove those.
		processName, _, _ := strings.Cut(process.Name, "#")
		pid := strconv.FormatUint(uint64(process.IDProcess), 10)
		parentPID := strconv.FormatUint(uint64(process.CreatingProcessID), 10)

		if c.config.EnableWorkerProcess {
			for _, wp := range workerProcesses {
				if wp.ProcessId == uint64(process.IDProcess) {
					processName = strings.Join([]string{processName, wp.AppPoolName}, "_")

					break
				}
			}
		}

		cmdLine, processOwner, processGroupID, err := c.getProcessInformation(logger, uint32(process.IDProcess))
		if err != nil {
			logger.Debug("Failed to get process information",
				slog.String("pid", pid),
				slog.Any("err", err),
			)
		}

		ch <- prometheus.MustNewConstMetric(
			c.info,
			prometheus.GaugeValue,
			1.0,
			processName, pid, parentPID, strconv.Itoa(int(processGroupID)), processOwner, cmdLine,
		)

		ch <- prometheus.MustNewConstMetric(
			c.startTime,
			prometheus.GaugeValue,
			process.ElapsedTime,
			processName, pid,
		)

		ch <- prometheus.MustNewConstMetric(
			c.handleCount,
			prometheus.GaugeValue,
			process.HandleCount,
			processName, pid,
		)

		ch <- prometheus.MustNewConstMetric(
			c.cpuTimeTotal,
			prometheus.CounterValue,
			process.PercentPrivilegedTime,
			processName, pid, "privileged",
		)

		ch <- prometheus.MustNewConstMetric(
			c.cpuTimeTotal,
			prometheus.CounterValue,
			process.PercentUserTime,
			processName, pid, "user",
		)

		ch <- prometheus.MustNewConstMetric(
			c.ioBytesTotal,
			prometheus.CounterValue,
			process.IOOtherBytesPerSec,
			processName, pid, "other",
		)

		ch <- prometheus.MustNewConstMetric(
			c.ioOperationsTotal,
			prometheus.CounterValue,
			process.IOOtherOperationsPerSec,
			processName, pid, "other",
		)

		ch <- prometheus.MustNewConstMetric(
			c.ioBytesTotal,
			prometheus.CounterValue,
			process.IOReadBytesPerSec,
			processName, pid, "read",
		)

		ch <- prometheus.MustNewConstMetric(
			c.ioOperationsTotal,
			prometheus.CounterValue,
			process.IOReadOperationsPerSec,
			processName, pid, "read",
		)

		ch <- prometheus.MustNewConstMetric(
			c.ioBytesTotal,
			prometheus.CounterValue,
			process.IOWriteBytesPerSec,
			processName, pid, "write",
		)

		ch <- prometheus.MustNewConstMetric(
			c.ioOperationsTotal,
			prometheus.CounterValue,
			process.IOWriteOperationsPerSec,
			processName, pid, "write",
		)

		ch <- prometheus.MustNewConstMetric(
			c.pageFaultsTotal,
			prometheus.CounterValue,
			process.PageFaultsPerSec,
			processName, pid,
		)

		ch <- prometheus.MustNewConstMetric(
			c.pageFileBytes,
			prometheus.GaugeValue,
			process.PageFileBytes,
			processName, pid,
		)

		ch <- prometheus.MustNewConstMetric(
			c.poolBytes,
			prometheus.GaugeValue,
			process.PoolNonPagedBytes,
			processName, pid, "nonpaged",
		)

		ch <- prometheus.MustNewConstMetric(
			c.poolBytes,
			prometheus.GaugeValue,
			process.PoolPagedBytes,
			processName, pid, "paged",
		)

		ch <- prometheus.MustNewConstMetric(
			c.priorityBase,
			prometheus.GaugeValue,
			process.PriorityBase,
			processName, pid,
		)

		ch <- prometheus.MustNewConstMetric(
			c.privateBytes,
			prometheus.GaugeValue,
			process.PrivateBytes,
			processName, pid,
		)

		ch <- prometheus.MustNewConstMetric(
			c.threadCount,
			prometheus.GaugeValue,
			process.ThreadCount,
			processName, pid,
		)

		ch <- prometheus.MustNewConstMetric(
			c.virtualBytes,
			prometheus.GaugeValue,
			process.VirtualBytes,
			processName, pid,
		)

		ch <- prometheus.MustNewConstMetric(
			c.workingSetPrivate,
			prometheus.GaugeValue,
			process.WorkingSetPrivate,
			processName, pid,
		)

		ch <- prometheus.MustNewConstMetric(
			c.workingSetPeak,
			prometheus.GaugeValue,
			process.WorkingSetPeak,
			processName, pid,
		)

		ch <- prometheus.MustNewConstMetric(
			c.workingSet,
			prometheus.GaugeValue,
			process.WorkingSet,
			processName, pid,
		)
	}

	return nil
}

func (c *Collector) collectPDH(logger *slog.Logger, ch chan<- prometheus.Metric) error {
	perfData, err := c.perfDataCollector.Collect()
	if err != nil {
		return fmt.Errorf("failed to collect metrics: %w", err)
	}

	if len(perfData) == 0 {
		return errors.New("perflib query returned empty result set")
	}

	var workerProcesses []WorkerProcess
	if c.config.EnableWorkerProcess {
		if err := c.wmiClient.Query("SELECT * FROM WorkerProcess", &workerProcesses, nil, "root\\WebAdministration"); err != nil {
			logger.Debug("Could not query WebAdministration namespace for IIS worker processes",
				slog.Any("err", err),
			)
		}
	}

	for name, process := range perfData {
		// Duplicate processes are suffixed #, and an index number. Remove those.
		name, _, _ = strings.Cut(name, "#")

		if name == "_Total" ||
			c.config.ProcessExclude.MatchString(name) ||
			!c.config.ProcessInclude.MatchString(name) {
			continue
		}

		var pid uint64

		if v, ok := process[processID]; ok {
			pid = uint64(v.FirstValue)
		} else if v, ok = process[idProcess]; ok {
			pid = uint64(v.FirstValue)
		}

		parentPID := strconv.FormatUint(uint64(process[creatingProcessID].FirstValue), 10)

		if c.config.EnableWorkerProcess {
			for _, wp := range workerProcesses {
				if wp.ProcessId == pid {
					name = strings.Join([]string{name, wp.AppPoolName}, "_")

					break
				}
			}
		}

		cmdLine, processOwner, processGroupID, err := c.getProcessInformation(logger, uint32(pid))
		if err != nil {
			logger.Debug("Failed to get process information",
				slog.Uint64("pid", pid),
				slog.Any("err", err),
			)
		}

		pidString := strconv.FormatUint(pid, 10)

		ch <- prometheus.MustNewConstMetric(
			c.info,
			prometheus.GaugeValue,
			1.0,
			name, pidString, parentPID, strconv.Itoa(int(processGroupID)), processOwner, cmdLine,
		)

		ch <- prometheus.MustNewConstMetric(
			c.startTime,
			prometheus.GaugeValue,
			process[elapsedTime].FirstValue,
			name, pidString,
		)

		ch <- prometheus.MustNewConstMetric(
			c.handleCount,
			prometheus.GaugeValue,
			process[handleCount].FirstValue,
			name, pidString,
		)

		ch <- prometheus.MustNewConstMetric(
			c.cpuTimeTotal,
			prometheus.CounterValue,
			process[percentPrivilegedTime].FirstValue,
			name, pidString, "privileged",
		)

		ch <- prometheus.MustNewConstMetric(
			c.cpuTimeTotal,
			prometheus.CounterValue,
			process[percentUserTime].FirstValue,
			name, pidString, "user",
		)

		ch <- prometheus.MustNewConstMetric(
			c.ioBytesTotal,
			prometheus.CounterValue,
			process[ioOtherBytesPerSec].FirstValue,
			name, pidString, "other",
		)

		ch <- prometheus.MustNewConstMetric(
			c.ioOperationsTotal,
			prometheus.CounterValue,
			process[ioOtherOperationsPerSec].FirstValue,
			name, pidString, "other",
		)

		ch <- prometheus.MustNewConstMetric(
			c.ioBytesTotal,
			prometheus.CounterValue,
			process[ioReadBytesPerSec].FirstValue,
			name, pidString, "read",
		)

		ch <- prometheus.MustNewConstMetric(
			c.ioOperationsTotal,
			prometheus.CounterValue,
			process[ioReadOperationsPerSec].FirstValue,
			name, pidString, "read",
		)

		ch <- prometheus.MustNewConstMetric(
			c.ioBytesTotal,
			prometheus.CounterValue,
			process[ioWriteBytesPerSec].FirstValue,
			name, pidString, "write",
		)

		ch <- prometheus.MustNewConstMetric(
			c.ioOperationsTotal,
			prometheus.CounterValue,
			process[ioWriteOperationsPerSec].FirstValue,
			name, pidString, "write",
		)

		ch <- prometheus.MustNewConstMetric(
			c.pageFaultsTotal,
			prometheus.CounterValue,
			process[pageFaultsPerSec].FirstValue,
			name, pidString,
		)

		ch <- prometheus.MustNewConstMetric(
			c.pageFileBytes,
			prometheus.GaugeValue,
			process[pageFileBytes].FirstValue,
			name, pidString,
		)

		ch <- prometheus.MustNewConstMetric(
			c.poolBytes,
			prometheus.GaugeValue,
			process[poolNonPagedBytes].FirstValue,
			name, pidString, "nonpaged",
		)

		ch <- prometheus.MustNewConstMetric(
			c.poolBytes,
			prometheus.GaugeValue,
			process[poolPagedBytes].FirstValue,
			name, pidString, "paged",
		)

		ch <- prometheus.MustNewConstMetric(
			c.priorityBase,
			prometheus.GaugeValue,
			process[priorityBase].FirstValue,
			name, pidString,
		)

		ch <- prometheus.MustNewConstMetric(
			c.privateBytes,
			prometheus.GaugeValue,
			process[privateBytes].FirstValue,
			name, pidString,
		)

		ch <- prometheus.MustNewConstMetric(
			c.threadCount,
			prometheus.GaugeValue,
			process[threadCount].FirstValue,
			name, pidString,
		)

		ch <- prometheus.MustNewConstMetric(
			c.virtualBytes,
			prometheus.GaugeValue,
			process[virtualBytes].FirstValue,
			name, pidString,
		)

		ch <- prometheus.MustNewConstMetric(
			c.workingSetPrivate,
			prometheus.GaugeValue,
			process[workingSetPrivate].FirstValue,
			name, pidString,
		)

		ch <- prometheus.MustNewConstMetric(
			c.workingSetPeak,
			prometheus.GaugeValue,
			process[workingSetPeak].FirstValue,
			name, pidString,
		)

		ch <- prometheus.MustNewConstMetric(
			c.workingSet,
			prometheus.GaugeValue,
			process[workingSet].FirstValue,
			name, pidString,
		)
	}

	return nil
}

// ref: https://github.com/microsoft/hcsshim/blob/8beabacfc2d21767a07c20f8dd5f9f3932dbf305/internal/uvm/stats.go#L25
func (c *Collector) getProcessInformation(logger *slog.Logger, pid uint32) (string, string, uint32, error) {
	if pid == 0 {
		return "", "", 0, nil
	}

	hProcess, vmReadAccess, err := c.openProcess(pid)
	if err != nil {
		if errors.Is(err, windows.ERROR_ACCESS_DENIED) {
			return "", "", 0, nil
		}

		return "", "", 0, err
	}

	defer func(hProcess windows.Handle) {
		if err := windows.CloseHandle(hProcess); err != nil {
			logger.Warn("CloseHandle failed",
				slog.Any("err", err),
			)
		}
	}(hProcess)

	owner, err := c.getProcessOwner(logger, hProcess)
	if err != nil {
		return "", "", 0, err
	}

	var (
		cmdLine        string
		processGroupID uint32
	)

	if vmReadAccess {
		cmdLine, processGroupID, err = c.getExtendedProcessInformation(hProcess)
		if err != nil {
			return "", owner, processGroupID, err
		}
	}

	return cmdLine, owner, processGroupID, nil
}

func (c *Collector) getExtendedProcessInformation(hProcess windows.Handle) (string, uint32, error) {
	// Get the process environment block (PEB) address
	var pbi windows.PROCESS_BASIC_INFORMATION

	retLen := uint32(unsafe.Sizeof(pbi))
	if err := windows.NtQueryInformationProcess(hProcess, windows.ProcessBasicInformation, unsafe.Pointer(&pbi), retLen, &retLen); err != nil {
		return "", 0, fmt.Errorf("failed to query process basic information: %w", err)
	}

	peb := windows.PEB{}

	err := windows.ReadProcessMemory(hProcess,
		uintptr(unsafe.Pointer(pbi.PebBaseAddress)),
		(*byte)(unsafe.Pointer(&peb)),
		unsafe.Sizeof(peb),
		nil,
	)
	if err != nil {
		return "", 0, fmt.Errorf("failed to read process memory: %w", err)
	}

	processParameters := windows.RTL_USER_PROCESS_PARAMETERS{}

	err = windows.ReadProcessMemory(hProcess,
		uintptr(unsafe.Pointer(peb.ProcessParameters)),
		(*byte)(unsafe.Pointer(&processParameters)),
		unsafe.Sizeof(processParameters),
		nil,
	)
	if err != nil {
		return "", 0, fmt.Errorf("failed to read process memory: %w", err)
	}

	cmdLineUTF16 := make([]uint16, processParameters.CommandLine.Length)

	err = windows.ReadProcessMemory(hProcess,
		uintptr(unsafe.Pointer(processParameters.CommandLine.Buffer)),
		(*byte)(unsafe.Pointer(&cmdLineUTF16[0])),
		uintptr(processParameters.CommandLine.Length),
		nil,
	)
	if err != nil {
		return "", processParameters.ProcessGroupId, fmt.Errorf("failed to read process memory: %w", err)
	}

	return strings.TrimSpace(windows.UTF16ToString(cmdLineUTF16)), processParameters.ProcessGroupId, nil
}

func (c *Collector) getProcessOwner(logger *slog.Logger, hProcess windows.Handle) (string, error) {
	var tok windows.Token

	if err := windows.OpenProcessToken(hProcess, windows.TOKEN_QUERY, &tok); err != nil {
		if errors.Is(err, windows.ERROR_ACCESS_DENIED) {
			return "", nil
		}

		return "", fmt.Errorf("failed to open process token: %w", err)
	}

	defer func(tok windows.Token) {
		if err := tok.Close(); err != nil {
			logger.Warn("Token close failed",
				slog.Any("err", err),
			)
		}
	}(tok)

	tokenUser, err := tok.GetTokenUser()
	if err != nil {
		return "", fmt.Errorf("failed to get token user: %w", err)
	}

	sid := tokenUser.User.Sid.String()

	owner, ok := c.lookupCache[sid]
	if !ok {
		account, domain, _, err := tokenUser.User.Sid.LookupAccount("")
		if err != nil {
			owner = sid
		} else {
			owner = fmt.Sprintf(`%s\%s`, account, domain)
		}

		c.lookupCache[sid] = owner
	}

	return owner, nil
}

func (c *Collector) openProcess(pid uint32) (windows.Handle, bool, error) {
	// Open the process with QUERY_INFORMATION and VM_READ permissions
	hProcess, err := windows.OpenProcess(windows.PROCESS_QUERY_INFORMATION|windows.PROCESS_VM_READ, false, pid)
	if err == nil {
		return hProcess, true, nil
	}

	if !errors.Is(err, windows.ERROR_ACCESS_DENIED) {
		return 0, false, fmt.Errorf("failed to open process: %w", err)
	}

	if errors.Is(err, windows.Errno(0x57)) { // invalid parameter, for PIDs that don't exist
		return 0, false, errors.New("process not found")
	}

	hProcess, err = windows.OpenProcess(windows.PROCESS_QUERY_LIMITED_INFORMATION, false, pid)
	if err != nil {
		return 0, false, fmt.Errorf("failed to open process with limited permissions: %w", err)
	}

	return hProcess, false, nil
}
