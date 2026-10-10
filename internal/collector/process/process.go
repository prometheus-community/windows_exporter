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

package process

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"regexp"
	"slices"
	"strconv"
	"sync"
	"time"

	"github.com/alecthomas/kingpin/v2"
	"github.com/prometheus-community/windows_exporter/internal/mi"
	"github.com/prometheus-community/windows_exporter/internal/types"
	"github.com/prometheus/client_golang/prometheus"
	"golang.org/x/sys/windows"
)

const Name = "process"

const (
	// windowsEpoch is the FILETIME of the Unix epoch, in 100ns intervals since 1601.
	windowsEpoch int64 = 116444736000000000
	// ticksPerSecond is the number of 100ns intervals per second.
	ticksPerSecond = 1e7
)

type Config struct {
	ProcessInclude      *regexp.Regexp `yaml:"include"`
	ProcessExclude      *regexp.Regexp `yaml:"exclude"`
	EnableWorkerProcess bool           `yaml:"iis"`
	EnableCMDLine       bool           `yaml:"cmdline"`
}

//nolint:gochecknoglobals
var ConfigDefaults = Config{
	ProcessInclude:      types.RegExpAny,
	ProcessExclude:      types.RegExpEmpty,
	EnableWorkerProcess: false,
	EnableCMDLine:       true,
}

type Collector struct {
	config Config

	logger *slog.Logger

	// miSession is set when root\WebAdministration answered during Build. It's
	// the fallback for worker processes whose command line can't be read.
	miSession                 *mi.Session
	workerProcessMIQueryQuery mi.Query

	snapshot  snapshotBuffer
	processes []processSnapshot
	// infoCache holds the windows_process_info values of the processes of the last scrape, keyed by PID.
	infoCache map[uint32]processInfo

	lookupCache sync.Map

	mu sync.RWMutex

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

	c := &Collector{
		config: *config,
	}

	return c
}

func NewWithFlags(app *kingpin.Application) *Collector {
	c := &Collector{
		config: ConfigDefaults,
	}

	var processExclude, processInclude string

	app.Flag(
		"collector.process.exclude",
		"Regexp of processes to exclude. Process name must both match include and not match exclude to be included.",
	).Default("").StringVar(&processExclude)

	app.Flag(
		"collector.process.include",
		"Regexp of processes to include. Process name must both match include and not match exclude to be included.",
	).Default(".+").StringVar(&processInclude)

	app.Flag(
		"collector.process.iis",
		"Append the IIS application pool name to the process name of IIS worker processes (w3wp).",
	).Default(strconv.FormatBool(c.config.EnableWorkerProcess)).BoolVar(&c.config.EnableWorkerProcess)

	app.Flag(
		"collector.process.cmdline",
		"If enabled, the full cmdline is exposed to the windows_process_info metrics.",
	).Default(strconv.FormatBool(c.config.EnableCMDLine)).BoolVar(&c.config.EnableCMDLine)

	app.Action(func(*kingpin.ParseContext) error {
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

func (c *Collector) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.snapshot = snapshotBuffer{}
	c.processes = nil
	c.infoCache = nil

	// The session belongs to the caller of Build.
	c.miSession = nil

	return nil
}

func (c *Collector) Build(logger *slog.Logger, miSession *mi.Session) error {
	c.logger = logger.With(slog.String("collector", Name))

	c.infoCache = nil
	c.lookupCache = sync.Map{}

	// The flags wrap the expressions in ^(?:...)$, the defaults of [New] are [types.RegExpAny] and [types.RegExpEmpty].
	if slices.Contains([]string{"^(?:.+)$", "^(?:.*)$", types.RegExpAny.String()}, c.config.ProcessInclude.String()) &&
		slices.Contains([]string{"^(?:)$", types.RegExpEmpty.String()}, c.config.ProcessExclude.String()) {
		logger.Warn("No filters specified for process collector. This will generate a very large number of metrics!")
	}

	c.info = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "info"),
		"Process information.",
		[]string{"process", "process_id", "creating_process_id", "process_group_id", "owner", "cmdline"},
		nil,
	)

	c.startTime = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "start_time_seconds_timestamp"),
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

	c.miSession = nil

	if c.config.EnableWorkerProcess {
		if err := c.buildWorkerProcessWMI(miSession); err != nil {
			return err
		}
	}

	return nil
}

// buildWorkerProcessWMI enables the root\WebAdministration fallback for application pool names
// if the IIS WMI provider is installed. Without it, only the w3wp command line is used.
func (c *Collector) buildWorkerProcessWMI(miSession *mi.Session) error {
	if miSession == nil {
		c.logger.LogAttrs(context.Background(), slog.LevelDebug,
			"No MI session, reading IIS application pool names from the w3wp command line only",
		)

		return nil
	}

	miQuery, err := mi.NewQuery("SELECT AppPoolName, ProcessId FROM WorkerProcess")
	if err != nil {
		return fmt.Errorf("failed to create WMI query: %w", err)
	}

	var workerProcesses []WorkerProcess

	if err = miSession.Query(&workerProcesses, mi.NamespaceRootWebAdministration, miQuery, 0); err != nil {
		c.logger.LogAttrs(context.Background(), slog.LevelDebug,
			`root\WebAdministration isn't available, reading IIS application pool names from the w3wp command line only`,
			slog.Any("err", err),
		)

		return nil
	}

	c.workerProcessMIQueryQuery = miQuery
	c.miSession = miSession

	return nil
}

func (c *Collector) Collect(ch chan<- prometheus.Metric, maxScrapeDuration time.Duration) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	var err error

	c.processes, err = c.snapshot.query(c.processes)
	if err != nil {
		return fmt.Errorf("failed to query processes: %w", err)
	}

	processes := make([]*processSnapshot, 0, len(c.processes))

	var workerProcessIDs []uint32

	for i := range c.processes {
		process := &c.processes[i]

		if c.config.ProcessExclude.MatchString(process.name) || !c.config.ProcessInclude.MatchString(process.name) {
			continue
		}

		if c.config.EnableWorkerProcess && isWorkerProcess(process.name) {
			workerProcessIDs = append(workerProcessIDs, process.pid)
		}

		processes = append(processes, process)
	}

	var appPools map[uint32]string

	if len(workerProcessIDs) > 0 {
		appPools, err = resolveAppPools(c.logger, workerProcessIDs, workerProcessAppPool, c.workerProcessWMIQuery(maxScrapeDuration))
	}

	infos := c.resolveProcessInfo(processes)

	for i, process := range processes {
		name := process.name
		if appPoolName := appPools[process.pid]; appPoolName != "" {
			name = name + "_" + appPoolName
		}

		c.collectProcess(ch, name, process, &infos[i])
	}

	return err
}

// workerProcessWMIQuery returns the root\WebAdministration query, or nil if it was unavailable during Build.
func (c *Collector) workerProcessWMIQuery(maxScrapeDuration time.Duration) func() ([]WorkerProcess, error) {
	miSession, miQuery := c.miSession, c.workerProcessMIQueryQuery
	if miSession == nil {
		return nil
	}

	return func() ([]WorkerProcess, error) {
		var workerProcesses []WorkerProcess

		err := miSession.Query(&workerProcesses, mi.NamespaceRootWebAdministration, miQuery, maxScrapeDuration)

		return workerProcesses, err
	}
}

// resolveAppPools maps IIS worker process IDs to application pool names.
//
// The command line of the worker process is the primary source: it needs no
// optional feature and no WMI round trip. WMI is queried at most once, for the
// worker processes whose command line couldn't be read or has no -ap argument,
// and only if queryWMI isn't nil. Unresolved processes have no entry.
func resolveAppPools(
	logger *slog.Logger,
	pids []uint32,
	readCommandLine func(pid uint32) (string, error),
	queryWMI func() ([]WorkerProcess, error),
) (map[uint32]string, error) {
	appPools := make(map[uint32]string, len(pids))

	var unresolved []uint32

	for _, pid := range pids {
		appPool, err := readCommandLine(pid)
		if err == nil {
			appPools[pid] = appPool

			continue
		}

		logger.LogAttrs(context.Background(), slog.LevelDebug, "Failed to read the IIS application pool from the worker process command line",
			slog.Uint64("pid", uint64(pid)),
			slog.Any("err", err),
		)

		// OpenProcess fails with ERROR_INVALID_PARAMETER for processes that exited after the
		// process snapshot. WAS no longer reports them either.
		if !errors.Is(err, windows.ERROR_INVALID_PARAMETER) {
			unresolved = append(unresolved, pid)
		}
	}

	if len(unresolved) == 0 || queryWMI == nil {
		return appPools, nil
	}

	workerProcesses, err := queryWMI()
	if err != nil {
		return appPools, fmt.Errorf("WMI query for collector.process.iis failed: %w", err)
	}

	byPID := make(map[uint64]string, len(workerProcesses))
	for _, wp := range workerProcesses {
		if _, ok := byPID[wp.ProcessId]; !ok && wp.AppPoolName != "" {
			byPID[wp.ProcessId] = wp.AppPoolName
		}
	}

	for _, pid := range unresolved {
		if appPool, ok := byPID[uint64(pid)]; ok {
			appPools[pid] = appPool
		}
	}

	return appPools, nil
}

func (c *Collector) collectProcess(ch chan<- prometheus.Metric, name string, process *processSnapshot, info *processInfo) {
	pid := strconv.FormatUint(uint64(process.pid), 10)

	ch <- prometheus.MustNewConstMetric(
		c.info,
		prometheus.GaugeValue,
		1.0,
		name, pid, strconv.FormatUint(uint64(process.parentPID), 10), strconv.FormatUint(uint64(info.processGroupID), 10), info.owner, info.cmdLine,
	)

	// Perflib reports whole seconds for the Elapsed Time counter of the Process counter set.
	ch <- prometheus.MustNewConstMetric(
		c.startTime,
		prometheus.GaugeValue,
		float64((process.createTime-windowsEpoch)/ticksPerSecond),
		name, pid,
	)

	ch <- prometheus.MustNewConstMetric(
		c.handleCount,
		prometheus.GaugeValue,
		float64(process.handleCount),
		name, pid,
	)

	ch <- prometheus.MustNewConstMetric(
		c.cpuTimeTotal,
		prometheus.CounterValue,
		float64(process.kernelTime)/ticksPerSecond,
		name, pid, "privileged",
	)

	ch <- prometheus.MustNewConstMetric(
		c.cpuTimeTotal,
		prometheus.CounterValue,
		float64(process.userTime)/ticksPerSecond,
		name, pid, "user",
	)

	ch <- prometheus.MustNewConstMetric(
		c.ioBytesTotal,
		prometheus.CounterValue,
		float64(process.otherTransferCount),
		name, pid, "other",
	)

	ch <- prometheus.MustNewConstMetric(
		c.ioOperationsTotal,
		prometheus.CounterValue,
		float64(process.otherOperationCount),
		name, pid, "other",
	)

	ch <- prometheus.MustNewConstMetric(
		c.ioBytesTotal,
		prometheus.CounterValue,
		float64(process.readTransferCount),
		name, pid, "read",
	)

	ch <- prometheus.MustNewConstMetric(
		c.ioOperationsTotal,
		prometheus.CounterValue,
		float64(process.readOperationCount),
		name, pid, "read",
	)

	ch <- prometheus.MustNewConstMetric(
		c.ioBytesTotal,
		prometheus.CounterValue,
		float64(process.writeTransferCount),
		name, pid, "write",
	)

	ch <- prometheus.MustNewConstMetric(
		c.ioOperationsTotal,
		prometheus.CounterValue,
		float64(process.writeOperationCount),
		name, pid, "write",
	)

	ch <- prometheus.MustNewConstMetric(
		c.pageFaultsTotal,
		prometheus.CounterValue,
		float64(process.pageFaultCount),
		name, pid,
	)

	ch <- prometheus.MustNewConstMetric(
		c.pageFileBytes,
		prometheus.GaugeValue,
		float64(process.pageFileUsage),
		name, pid,
	)

	ch <- prometheus.MustNewConstMetric(
		c.poolBytes,
		prometheus.GaugeValue,
		float64(process.nonPagedPoolUsage),
		name, pid, "nonpaged",
	)

	ch <- prometheus.MustNewConstMetric(
		c.poolBytes,
		prometheus.GaugeValue,
		float64(process.pagedPoolUsage),
		name, pid, "paged",
	)

	ch <- prometheus.MustNewConstMetric(
		c.priorityBase,
		prometheus.GaugeValue,
		float64(process.basePriority),
		name, pid,
	)

	ch <- prometheus.MustNewConstMetric(
		c.privateBytes,
		prometheus.GaugeValue,
		float64(process.privatePageCount),
		name, pid,
	)

	ch <- prometheus.MustNewConstMetric(
		c.threadCount,
		prometheus.GaugeValue,
		float64(process.threadCount),
		name, pid,
	)

	ch <- prometheus.MustNewConstMetric(
		c.virtualBytes,
		prometheus.GaugeValue,
		float64(process.virtualSize),
		name, pid,
	)

	ch <- prometheus.MustNewConstMetric(
		c.workingSetPrivate,
		prometheus.GaugeValue,
		float64(process.workingSetPrivate),
		name, pid,
	)

	ch <- prometheus.MustNewConstMetric(
		c.workingSetPeak,
		prometheus.GaugeValue,
		float64(process.workingSetPeak),
		name, pid,
	)

	ch <- prometheus.MustNewConstMetric(
		c.workingSet,
		prometheus.GaugeValue,
		float64(process.workingSetSize),
		name, pid,
	)
}
