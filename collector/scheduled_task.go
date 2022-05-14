//go:build windows
// +build windows

package collector

import (
	"fmt"
	"regexp"
	"runtime"
	"strings"

	ole "github.com/go-ole/go-ole"
	"github.com/go-ole/go-ole/oleutil"
	"github.com/prometheus-community/windows_exporter/log"
	"github.com/prometheus/client_golang/prometheus"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	taskWhitelist = kingpin.Flag(
		"collector.scheduled_task.whitelist",
		"Regexp of tasks to whitelist. Task path must both match whitelist and not match blacklist to be included.",
	).Default(".+").String()
	taskBlacklist = kingpin.Flag(
		"collector.scheduled_task.blacklist",
		"Regexp of tasks to blacklist. Task path must both match whitelist and not match blacklist to be included.",
	).String()
)

type ScheduledTaskCollector struct {
	LastResult *prometheus.Desc
	MissedRuns *prometheus.Desc
	State      *prometheus.Desc

	taskWhitelistPattern *regexp.Regexp
	taskBlacklistPattern *regexp.Regexp
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

func init() {
	registerCollector("scheduled_task", NewScheduledTask)
}

// NewScheduledTask ...
func NewScheduledTask() (Collector, error) {
	const subsystem = "scheduled_task"

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

		taskWhitelistPattern: regexp.MustCompile(fmt.Sprintf("^(?:%s)$", *taskWhitelist)),
		taskBlacklistPattern: regexp.MustCompile(fmt.Sprintf("^(?:%s)$", *taskBlacklist)),
	}, nil
}

func (c *ScheduledTaskCollector) Collect(ctx *ScrapeContext, ch chan<- prometheus.Metric) error {
	if desc, err := c.collect(ch); err != nil {
		log.Error("failed collecting user metrics:", desc, err)
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
		if c.taskBlacklistPattern.MatchString(task.Path) ||
			!c.taskWhitelistPattern.MatchString(task.Path) {
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
		return fetchTasksRecursively(subFolder, scheduledTasks)
	})

	return err
}

func parseTask(task *ole.IDispatch) (scheduledTask ScheduledTask, err error) {
	taskNameVar, err := oleutil.GetProperty(task, "Name")
	if err != nil {
		return scheduledTask, err
	}

	taskPathVar, err := oleutil.GetProperty(task, "Path")
	if err != nil {
		return scheduledTask, err
	}

	taskEnabledVar, err := oleutil.GetProperty(task, "Enabled")
	if err != nil {
		return scheduledTask, err
	}

	taskStateVar, err := oleutil.GetProperty(task, "State")
	if err != nil {
		return scheduledTask, err
	}

	taskNumberOfMissedRunsVar, err := oleutil.GetProperty(task, "NumberOfMissedRuns")
	if err != nil {
		return scheduledTask, err
	}

	taskLastTaskResultVar, err := oleutil.GetProperty(task, "LastTaskResult")
	if err != nil {
		return scheduledTask, err
	}

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
