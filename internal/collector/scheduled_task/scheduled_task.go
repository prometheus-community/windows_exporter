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

package scheduled_task

import (
	"errors"
	"fmt"
	"log/slog"
	"regexp"
	"runtime"
	"strings"
	"time"

	"github.com/alecthomas/kingpin/v2"
	"github.com/go-ole/go-ole"
	"github.com/go-ole/go-ole/oleutil"
	"github.com/prometheus-community/windows_exporter/internal/mi"
	"github.com/prometheus-community/windows_exporter/internal/types"
	"github.com/prometheus/client_golang/prometheus"
)

const Name = "scheduled_task"

type Config struct {
	TaskExclude *regexp.Regexp `yaml:"exclude"`
	TaskInclude *regexp.Regexp `yaml:"include"`
}

//nolint:gochecknoglobals
var ConfigDefaults = Config{
	TaskExclude: types.RegExpEmpty,
	TaskInclude: types.RegExpAny,
}

type Collector struct {
	config Config
	logger *slog.Logger

	lastResult       *prometheus.Desc
	lastResultStatus *prometheus.Desc
	missedRuns       *prometheus.Desc
	state            *prometheus.Desc
}

// TaskState ...
// https://docs.microsoft.com/en-us/windows/desktop/api/taskschd/ne-taskschd-task_state
type TaskState uint

// TaskResult preserves the unsigned 32-bit representation of a result code,
// including HRESULTs returned as signed LONG values by Task Scheduler.
type TaskResult uint32

const (
	TASK_STATE_UNKNOWN TaskState = iota
	TASK_STATE_DISABLED
	TASK_STATE_QUEUED
	TASK_STATE_READY
	TASK_STATE_RUNNING
)

const (
	SCHED_S_SUCCESS                TaskResult = 0x0
	SCHED_S_TASK_READY             TaskResult = 0x00041300
	SCHED_S_TASK_RUNNING           TaskResult = 0x00041301
	SCHED_S_TASK_DISABLED          TaskResult = 0x00041302
	SCHED_S_TASK_HAS_NOT_RUN       TaskResult = 0x00041303
	SCHED_S_TASK_NO_MORE_RUNS      TaskResult = 0x00041304
	SCHED_S_TASK_NOT_SCHEDULED     TaskResult = 0x00041305
	SCHED_S_TASK_TERMINATED        TaskResult = 0x00041306
	SCHED_S_TASK_NO_VALID_TRIGGERS TaskResult = 0x00041307
	SCHED_S_EVENT_TRIGGER          TaskResult = 0x00041308
	SCHED_S_TASK_QUEUED            TaskResult = 0x00041325
)

type ScheduledTask struct {
	Name            string
	Path            string
	Enabled         bool
	State           TaskState
	MissedRunsCount float64
	LastTaskResult  TaskResult
}

type ScheduledTasks []ScheduledTask

func New(config *Config) *Collector {
	if config == nil {
		config = &ConfigDefaults
	}

	if config.TaskExclude == nil {
		config.TaskExclude = ConfigDefaults.TaskExclude
	}

	if config.TaskInclude == nil {
		config.TaskInclude = ConfigDefaults.TaskInclude
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

	var taskExclude, taskInclude string

	app.Flag(
		"collector.scheduled_task.exclude",
		"Regexp of tasks to exclude. Task path must both match include and not match exclude to be included.",
	).Default("").StringVar(&taskExclude)

	app.Flag(
		"collector.scheduled_task.include",
		"Regexp of tasks to include. Task path must both match include and not match exclude to be included.",
	).Default(".+").StringVar(&taskInclude)

	app.Action(func(*kingpin.ParseContext) error {
		var err error

		c.config.TaskExclude, err = regexp.Compile(fmt.Sprintf("^(?:%s)$", taskExclude))
		if err != nil {
			return fmt.Errorf("collector.scheduled_task.exclude: %w", err)
		}

		c.config.TaskInclude, err = regexp.Compile(fmt.Sprintf("^(?:%s)$", taskInclude))
		if err != nil {
			return fmt.Errorf("collector.scheduled_task.include: %w", err)
		}

		return nil
	})

	return c
}

func (c *Collector) GetName() string {
	return Name
}

func (c *Collector) Close() error {
	return nil
}

func (c *Collector) Build(logger *slog.Logger, _ *mi.Session) error {
	c.logger = logger.With(slog.String("collector", Name))

	c.lastResult = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "last_result"),
		"DEPRECATED: use windows_scheduled_task_last_result_status. "+
			"1 if the last result code of the registered task is zero, 0 otherwise",
		[]string{"task"},
		nil,
	)

	c.lastResultStatus = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "last_result_status"),
		"The last result status of a scheduled task, 1 if the current status, 0 otherwise",
		[]string{"task", "status"},
		nil,
	)

	c.missedRuns = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "missed_runs"),
		"The number of times the registered task missed a scheduled run",
		[]string{"task"},
		nil,
	)

	c.state = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "state"),
		"The current state of a scheduled task",
		[]string{"task", "state"},
		nil,
	)

	return nil
}

func (c *Collector) Collect(ch chan<- prometheus.Metric, _ time.Duration) error {
	return c.collect(ch)
}

//nolint:gochecknoglobals
var TASK_STATES = []string{"disabled", "queued", "ready", "running", "unknown"}

//nolint:gochecknoglobals
var TASK_RESULT_STATUSES = []string{
	"success", "ready", "running", "disabled", "has_not_run", "no_more_runs",
	"not_scheduled", "terminated", "no_valid_triggers", "event_trigger", "queued", "unknown",
}

func (c *Collector) collect(ch chan<- prometheus.Metric) error {
	scheduledTasks, err := getScheduledTasks()
	if errors.Is(err, errTasksSkipped) {
		// Tasks and folders that can't be read are skipped, the other tasks are still collected.
		c.logger.Warn("failed to read some scheduled tasks",
			slog.Any("err", err),
		)
	} else if err != nil {
		return fmt.Errorf("get scheduled tasks: %w", err)
	}

	c.collectMetrics(ch, scheduledTasks)

	return nil
}

func (c *Collector) collectMetrics(ch chan<- prometheus.Metric, scheduledTasks ScheduledTasks) {
	for _, task := range scheduledTasks {
		if c.config.TaskExclude.MatchString(task.Path) ||
			!c.config.TaskInclude.MatchString(task.Path) {
			continue
		}

		for _, state := range TASK_STATES {
			var stateValue float64

			if strings.ToLower(task.State.String()) == state {
				stateValue = 1.0
			}

			ch <- prometheus.MustNewConstMetric(
				c.state,
				prometheus.GaugeValue,
				stateValue,
				task.Path,
				state,
			)
		}

		resultStatus := task.LastTaskResult.String()

		for _, status := range TASK_RESULT_STATUSES {
			var statusValue float64

			if resultStatus == status {
				statusValue = 1
			}

			ch <- prometheus.MustNewConstMetric(
				c.lastResultStatus,
				prometheus.GaugeValue,
				statusValue,
				task.Path,
				status,
			)
		}

		if task.LastTaskResult == SCHED_S_TASK_HAS_NOT_RUN {
			continue
		}

		lastResult := 0.0
		if task.LastTaskResult == SCHED_S_SUCCESS {
			lastResult = 1.0
		}

		ch <- prometheus.MustNewConstMetric(
			c.lastResult,
			prometheus.GaugeValue,
			lastResult,
			task.Path,
		)

		ch <- prometheus.MustNewConstMetric(
			c.missedRuns,
			prometheus.GaugeValue,
			task.MissedRunsCount,
			task.Path,
		)
	}
}

// CLSID_TaskScheduler {0F87369F-A4E5-4CFC-BD3E-73E6154572DD}: https://github.com/microsoft/win32metadata/blob/76c04c2021ef4a831a6f1e06d9566002d746139b/generation/WinSDK/RecompiledIdlHeaders/um/taskschd.h#L10359
//
//nolint:gochecknoglobals // CLSID, not ProgID: go-ole's CLSIDFromProgID can free the ProgID buffer mid-call
var taskSchedulerCLSID = ole.GUID{
	Data1: 0x0F87369F,
	Data2: 0xA4E5,
	Data3: 0x4CFC,
	Data4: [8]byte{0xBD, 0x3E, 0x73, 0xE6, 0x15, 0x45, 0x72, 0xDD},
}

// S_FALSE is returned by CoInitialize if it was already called on this thread.
const S_FALSE = 0x00000001

// errTasksSkipped is returned with the readable tasks if some tasks or folders could not be read.
var errTasksSkipped = errors.New("tasks skipped")

func getScheduledTasks() (ScheduledTasks, error) {
	var scheduledTasks ScheduledTasks

	// The only way to run WMI queries in parallel while being thread-safe is to
	// ensure the CoInitialize[Ex]() call is bound to its current OS thread.
	// Otherwise, attempting to initialize and run parallel queries across
	// goroutines will result in protected memory errors.
	runtime.LockOSThread()

	defer runtime.UnlockOSThread()

	if err := ole.CoInitializeEx(0, ole.COINIT_APARTMENTTHREADED|ole.COINIT_DISABLE_OLE1DDE); err != nil {
		if oleCode, ok := errors.AsType[*ole.OleError](err); ok && oleCode.Code() != ole.S_OK && oleCode.Code() != S_FALSE {
			return nil, err
		}
	}

	defer ole.CoUninitialize()

	taskSchedulerObj, err := ole.CreateInstance(&taskSchedulerCLSID, nil)
	if err != nil || taskSchedulerObj == nil {
		return scheduledTasks, err
	}
	defer taskSchedulerObj.Release()

	taskServiceObj, err := taskSchedulerObj.QueryInterface(ole.IID_IDispatch)
	if err != nil {
		return scheduledTasks, fmt.Errorf("IID_IDispatch: %w", err)
	}

	defer taskServiceObj.Release()

	_, err = oleutil.CallMethod(taskServiceObj, "Connect")
	if err != nil {
		return scheduledTasks, err
	}

	res, err := oleutil.CallMethod(taskServiceObj, "GetFolder", `\`)
	if err != nil {
		return scheduledTasks, err
	}

	rootFolderObj := res.ToIDispatch()
	defer rootFolderObj.Release()

	if err = fetchTasksRecursively(rootFolderObj, `\`, &scheduledTasks); err != nil {
		return scheduledTasks, fmt.Errorf("%w: %w", errTasksSkipped, err)
	}

	return scheduledTasks, nil
}

// fetchTasksInFolder appends the tasks of folder to scheduledTasks.
// A task that can't be read is skipped and its error is returned after all other tasks are read.
func fetchTasksInFolder(folder *ole.IDispatch, scheduledTasks *ScheduledTasks) error {
	res, err := oleutil.CallMethod(folder, "GetTasks", 1)
	if err != nil {
		return fmt.Errorf("get tasks: %w", err)
	}

	tasks := res.ToIDispatch()
	defer tasks.Release()

	errs := make([]error, 0)

	err = oleutil.ForEach(tasks, func(v *ole.VARIANT) error {
		task := v.ToIDispatch()
		defer task.Release()

		parsedTask, err := parseTask(task)
		if err != nil {
			errs = append(errs, fmt.Errorf("parse task: %w", err))

			return nil
		}

		*scheduledTasks = append(*scheduledTasks, parsedTask)

		return nil
	})
	if err != nil {
		errs = append(errs, fmt.Errorf("enumerate tasks: %w", err))
	}

	return errors.Join(errs...)
}

// fetchTasksRecursively appends the tasks of folder and its sub folders to scheduledTasks.
// A folder or task that can't be read is skipped and its error is returned after all other
// folders are read.
func fetchTasksRecursively(folder *ole.IDispatch, folderPath string, scheduledTasks *ScheduledTasks) error {
	errs := make([]error, 0)

	if err := fetchTasksInFolder(folder, scheduledTasks); err != nil {
		errs = append(errs, fmt.Errorf("folder %s: %w", folderPath, err))
	}

	res, err := oleutil.CallMethod(folder, "GetFolders", 1)
	if err != nil {
		return errors.Join(append(errs, fmt.Errorf("folder %s: get sub folders: %w", folderPath, err))...)
	}

	subFolders := res.ToIDispatch()
	defer subFolders.Release()

	err = oleutil.ForEach(subFolders, func(v *ole.VARIANT) error {
		subFolder := v.ToIDispatch()
		defer subFolder.Release()

		subFolderPath := folderPath

		if pathVar, err := oleutil.GetProperty(subFolder, "Path"); err == nil {
			subFolderPath = pathVar.ToString()
			_ = pathVar.Clear()
		}

		if err := fetchTasksRecursively(subFolder, subFolderPath, scheduledTasks); err != nil {
			errs = append(errs, err)
		}

		return nil
	})
	if err != nil {
		errs = append(errs, fmt.Errorf("folder %s: enumerate sub folders: %w", folderPath, err))
	}

	return errors.Join(errs...)
}

func parseTask(task *ole.IDispatch) (ScheduledTask, error) {
	var scheduledTask ScheduledTask

	taskNameVar, err := oleutil.GetProperty(task, "Name")
	if err != nil {
		return scheduledTask, err
	}

	defer func() {
		if tempErr := taskNameVar.Clear(); tempErr != nil {
			err = tempErr
		}
	}()

	taskPathVar, err := oleutil.GetProperty(task, "Path")
	if err != nil {
		return scheduledTask, err
	}

	defer func() {
		if tempErr := taskPathVar.Clear(); tempErr != nil {
			err = tempErr
		}
	}()

	taskEnabledVar, err := oleutil.GetProperty(task, "Enabled")
	if err != nil {
		return scheduledTask, err
	}

	defer func() {
		if tempErr := taskEnabledVar.Clear(); tempErr != nil {
			err = tempErr
		}
	}()

	taskStateVar, err := oleutil.GetProperty(task, "State")
	if err != nil {
		return scheduledTask, err
	}

	defer func() {
		if tempErr := taskStateVar.Clear(); tempErr != nil {
			err = tempErr
		}
	}()

	taskNumberOfMissedRunsVar, err := oleutil.GetProperty(task, "NumberOfMissedRuns")
	if err != nil {
		return scheduledTask, err
	}

	defer func() {
		if tempErr := taskNumberOfMissedRunsVar.Clear(); tempErr != nil {
			err = tempErr
		}
	}()

	taskLastTaskResultVar, err := oleutil.GetProperty(task, "LastTaskResult")
	if err != nil {
		return scheduledTask, err
	}

	defer func() {
		if tempErr := taskLastTaskResultVar.Clear(); tempErr != nil {
			err = tempErr
		}
	}()

	scheduledTask.Name = taskNameVar.ToString()
	scheduledTask.Path = strings.ReplaceAll(taskPathVar.ToString(), "\\", "/")

	if val, ok := taskEnabledVar.Value().(bool); ok {
		scheduledTask.Enabled = val
	}

	scheduledTask.State = TaskState(taskStateVar.Val)
	scheduledTask.MissedRunsCount = float64(taskNumberOfMissedRunsVar.Val)
	scheduledTask.LastTaskResult = TaskResult(taskLastTaskResultVar.Val)

	return scheduledTask, err
}

func (t TaskState) String() string {
	switch t {
	case TASK_STATE_UNKNOWN:
		return "Unknown"
	case TASK_STATE_DISABLED:
		return "Disabled"
	case TASK_STATE_QUEUED:
		return "Queued"
	case TASK_STATE_READY:
		return "Ready"
	case TASK_STATE_RUNNING:
		return "Running"
	default:
		return ""
	}
}

func (t TaskResult) String() string {
	switch t {
	case SCHED_S_SUCCESS:
		return "success"
	case SCHED_S_TASK_READY:
		return "ready"
	case SCHED_S_TASK_RUNNING:
		return "running"
	case SCHED_S_TASK_DISABLED:
		return "disabled"
	case SCHED_S_TASK_HAS_NOT_RUN:
		return "has_not_run"
	case SCHED_S_TASK_NO_MORE_RUNS:
		return "no_more_runs"
	case SCHED_S_TASK_NOT_SCHEDULED:
		return "not_scheduled"
	case SCHED_S_TASK_TERMINATED:
		return "terminated"
	case SCHED_S_TASK_NO_VALID_TRIGGERS:
		return "no_valid_triggers"
	case SCHED_S_EVENT_TRIGGER:
		return "event_trigger"
	case SCHED_S_TASK_QUEUED:
		return "queued"
	default:
		return "unknown"
	}
}
