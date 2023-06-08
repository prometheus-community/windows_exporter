//go:build windows
// +build windows

package collector

import (
	"errors"
	"fmt"
	"regexp"
	"runtime"
	"strings"

	"github.com/alecthomas/kingpin/v2"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	ole "github.com/go-ole/go-ole"
	"github.com/go-ole/go-ole/oleutil"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	FlagScheduledTaskOldExclude = "collector.scheduled_task.blacklist"
	FlagScheduledTaskOldInclude = "collector.scheduled_task.whitelist"

	FlagScheduledTaskExclude = "collector.scheduled_task.exclude"
	FlagScheduledTaskInclude = "collector.scheduled_task.include"
)

var (
	taskOldExclude *string
	taskOldInclude *string

	taskExclude *string
	taskInclude *string

	taskIncludeSet bool
	taskExcludeSet bool
)

type ScheduledTaskCollector struct {
	logger log.Logger

	LastResult *prometheus.Desc
	MissedRuns *prometheus.Desc
	State      *prometheus.Desc

	taskIncludePattern *regexp.Regexp
	taskExcludePattern *regexp.Regexp
}

// TaskState ...
// https://docs.microsoft.com/en-us/windows/desktop/api/taskschd/ne-taskschd-task_state
type TaskState uint

type TaskResult uint

const (
	TASK_STATE_UNKNOWN TaskState = iota
	TASK_STATE_DISABLED
	TASK_STATE_QUEUED
	TASK_STATE_READY
	TASK_STATE_RUNNING
	TASK_RESULT_SUCCESS TaskResult = 0x0
)

// RegisteredTask ...
type ScheduledTask struct {
	Name            string
	Path            string
	Enabled         bool
	State           TaskState
	MissedRunsCount float64
	LastTaskResult  TaskResult
}

type ScheduledTasks []ScheduledTask

// newScheduledTask ...
func newScheduledTaskFlags(app *kingpin.Application) {
	taskInclude = app.Flag(
		FlagScheduledTaskInclude,
		"Regexp of tasks to include. Task path must both match include and not match exclude to be included.",
	).Default(".+").PreAction(func(c *kingpin.ParseContext) error {
		taskIncludeSet = true
		return nil
	}).String()

	taskExclude = app.Flag(
		FlagScheduledTaskExclude,
		"Regexp of tasks to exclude. Task path must both match include and not match exclude to be included.",
	).Default("").PreAction(func(c *kingpin.ParseContext) error {
		taskExcludeSet = true
		return nil
	}).String()

	taskOldInclude = app.Flag(
		FlagScheduledTaskOldInclude,
		"DEPRECATED: Use --collector.scheduled_task.include",
	).Hidden().String()
	taskOldExclude = app.Flag(
		FlagScheduledTaskOldExclude,
		"DEPRECATED: Use --collector.scheduled_task.exclude",
	).Hidden().String()
}

// newScheduledTask ...
func newScheduledTask(logger log.Logger) (Collector, error) {
	const subsystem = "scheduled_task"
	logger = log.With(logger, "collector", subsystem)

	if *taskOldExclude != "" {
		if !taskExcludeSet {
			_ = level.Warn(logger).Log("msg", "--collector.scheduled_task.blacklist is DEPRECATED and will be removed in a future release, use --collector.scheduled_task.exclude")
			*taskExclude = *taskOldExclude
		} else {
			return nil, errors.New("--collector.scheduled_task.blacklist and --collector.scheduled_task.exclude are mutually exclusive")
		}
	}
	if *taskOldInclude != "" {
		if !taskIncludeSet {
			_ = level.Warn(logger).Log("msg", "--collector.scheduled_task.whitelist is DEPRECATED and will be removed in a future release, use --collector.scheduled_task.include")
			*taskInclude = *taskOldInclude
		} else {
			return nil, errors.New("--collector.scheduled_task.whitelist and --collector.scheduled_task.include are mutually exclusive")
		}
	}

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	err := ole.CoInitializeEx(0, ole.COINIT_MULTITHREADED)
	if err != nil {
		code := err.(*ole.OleError).Code()
		if code != ole.S_OK && code != S_FALSE {
			return nil, err
		}
	}
	defer ole.CoUninitialize()

	return &ScheduledTaskCollector{
		logger: logger,
		LastResult: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "last_result"),
			"The result that was returned the last time the registered task was run",
			[]string{"task"},
			nil,
		),

		MissedRuns: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "missed_runs"),
			"The number of times the registered task missed a scheduled run",
			[]string{"task"},
			nil,
		),

		State: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "state"),
			"The current state of a scheduled task",
			[]string{"task", "state"},
			nil,
		),

		taskIncludePattern: regexp.MustCompile(fmt.Sprintf("^(?:%s)$", *taskInclude)),
		taskExcludePattern: regexp.MustCompile(fmt.Sprintf("^(?:%s)$", *taskExclude)),
	}, nil
}

func (c *ScheduledTaskCollector) Collect(ctx *ScrapeContext, ch chan<- prometheus.Metric) error {
	if desc, err := c.collect(ch); err != nil {
		_ = level.Error(c.logger).Log("failed collecting user metrics", "desc", desc, "err", err)
		return err
	}

	return nil
}

var TASK_STATES = []string{"disabled", "queued", "ready", "running", "unknown"}

func (c *ScheduledTaskCollector) collect(ch chan<- prometheus.Metric) (*prometheus.Desc, error) {
	scheduledTasks, err := getScheduledTasks()
	if err != nil {
		return nil, err
	}

	for _, task := range scheduledTasks {
		if c.taskExcludePattern.MatchString(task.Path) ||
			!c.taskIncludePattern.MatchString(task.Path) {
			continue
		}

		lastResult := 0.0
		if task.LastTaskResult == TASK_RESULT_SUCCESS {
			lastResult = 1.0
		}

		ch <- prometheus.MustNewConstMetric(
			c.LastResult,
			prometheus.GaugeValue,
			lastResult,
			task.Path,
		)

		ch <- prometheus.MustNewConstMetric(
			c.MissedRuns,
			prometheus.GaugeValue,
			task.MissedRunsCount,
			task.Path,
		)

		for _, state := range TASK_STATES {
			var stateValue float64

			if strings.ToLower(task.State.String()) == state {
				stateValue = 1.0
			}

			ch <- prometheus.MustNewConstMetric(
				c.State,
				prometheus.GaugeValue,
				stateValue,
				task.Path,
				state,
			)
		}
	}

	return nil, nil
}

const SCHEDULED_TASK_PROGRAM_ID = "Schedule.Service.1"

// S_FALSE is returned by CoInitialize if it was already called on this thread.
const S_FALSE = 0x00000001

func getScheduledTasks() (scheduledTasks ScheduledTasks, err error) {
	schedClassID, err := ole.ClassIDFrom(SCHEDULED_TASK_PROGRAM_ID)
	if err != nil {
		return scheduledTasks, err
	}

	taskSchedulerObj, err := ole.CreateInstance(schedClassID, nil)
	if err != nil || taskSchedulerObj == nil {
		return scheduledTasks, err
	}
	defer taskSchedulerObj.Release()

	taskServiceObj := taskSchedulerObj.MustQueryInterface(ole.IID_IDispatch)
	_, err = oleutil.CallMethod(taskServiceObj, "Connect")
	if err != nil {
		return scheduledTasks, err
	}
	defer taskServiceObj.Release()

	res, err := oleutil.CallMethod(taskServiceObj, "GetFolder", `\`)
	if err != nil {
		return scheduledTasks, err
	}

	rootFolderObj := res.ToIDispatch()
	defer rootFolderObj.Release()

	err = fetchTasksRecursively(rootFolderObj, &scheduledTasks)

	return scheduledTasks, err
}

func fetchTasksInFolder(folder *ole.IDispatch, scheduledTasks *ScheduledTasks) error {
	res, err := oleutil.CallMethod(folder, "GetTasks", 1)
	if err != nil {
		return err
	}

	tasks := res.ToIDispatch()
	defer tasks.Release()

	err = oleutil.ForEach(tasks, func(v *ole.VARIANT) error {
		task := v.ToIDispatch()
		defer task.Release()

		parsedTask, err := parseTask(task)
		if err != nil {
			return err
		}

		*scheduledTasks = append(*scheduledTasks, parsedTask)

		return nil
	})

	return err
}

func fetchTasksRecursively(folder *ole.IDispatch, scheduledTasks *ScheduledTasks) error {
	if err := fetchTasksInFolder(folder, scheduledTasks); err != nil {
		return err
	}

	res, err := oleutil.CallMethod(folder, "GetFolders", 1)
	if err != nil {
		return err
	}

	subFolders := res.ToIDispatch()
	defer subFolders.Release()

	err = oleutil.ForEach(subFolders, func(v *ole.VARIANT) error {
		subFolder := v.ToIDispatch()
		defer subFolder.Release()
		return fetchTasksRecursively(subFolder, scheduledTasks)
	})

	return err
}

func parseTask(task *ole.IDispatch) (scheduledTask ScheduledTask, err error) {
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
	scheduledTask.Enabled = taskEnabledVar.Value().(bool)
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
