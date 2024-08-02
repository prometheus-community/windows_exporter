//go:build windows

package scheduled_task

import (
	"errors"
	"fmt"
	"regexp"
	"runtime"
	"strings"

	"github.com/alecthomas/kingpin/v2"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/go-ole/go-ole"
	"github.com/go-ole/go-ole/oleutil"
	"github.com/prometheus-community/windows_exporter/pkg/types"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	Name = "scheduled_task"

	FlagScheduledTaskExclude = "collector.scheduled_task.exclude"
	FlagScheduledTaskInclude = "collector.scheduled_task.include"
)

type Config struct {
	TaskExclude string `yaml:"task_exclude"`
	TaskInclude string `yaml:"task_include"`
}

var ConfigDefaults = Config{
	TaskExclude: "",
	TaskInclude: ".+",
}

type Collector struct {
	logger log.Logger

	taskExclude *string
	taskInclude *string

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

type ScheduledTask struct {
	Name            string
	Path            string
	Enabled         bool
	State           TaskState
	MissedRunsCount float64
	LastTaskResult  TaskResult
}

type ScheduledTasks []ScheduledTask

func New(logger log.Logger, config *Config) *Collector {
	if config == nil {
		config = &ConfigDefaults
	}

	c := &Collector{
		taskExclude: &config.TaskExclude,
		taskInclude: &config.TaskInclude,
	}
	c.SetLogger(logger)

	return c
}

func NewWithFlags(app *kingpin.Application) *Collector {
	c := &Collector{
		taskInclude: app.Flag(
			FlagScheduledTaskInclude,
			"Regexp of tasks to include. Task path must both match include and not match exclude to be included.",
		).Default(ConfigDefaults.TaskInclude).String(),

		taskExclude: app.Flag(
			FlagScheduledTaskExclude,
			"Regexp of tasks to exclude. Task path must both match include and not match exclude to be included.",
		).Default(ConfigDefaults.TaskExclude).String(),
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
	return []string{}, nil
}

func (c *Collector) Close() error {
	return nil
}

func (c *Collector) Build() error {
	c.LastResult = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "last_result"),
		"The result that was returned the last time the registered task was run",
		[]string{"task"},
		nil,
	)

	c.MissedRuns = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "missed_runs"),
		"The number of times the registered task missed a scheduled run",
		[]string{"task"},
		nil,
	)

	c.State = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "state"),
		"The current state of a scheduled task",
		[]string{"task", "state"},
		nil,
	)

	var err error

	c.taskIncludePattern, err = regexp.Compile(fmt.Sprintf("^(?:%s)$", *c.taskInclude))
	if err != nil {
		return err
	}

	c.taskExcludePattern, err = regexp.Compile(fmt.Sprintf("^(?:%s)$", *c.taskExclude))
	if err != nil {
		return err
	}

	return nil
}

func (c *Collector) Collect(_ *types.ScrapeContext, ch chan<- prometheus.Metric) error {
	if err := c.collect(ch); err != nil {
		_ = level.Error(c.logger).Log("msg", "failed collecting user metrics", "err", err)
		return err
	}

	return nil
}

var TASK_STATES = []string{"disabled", "queued", "ready", "running", "unknown"}

func (c *Collector) collect(ch chan<- prometheus.Metric) error {
	scheduledTasks, err := getScheduledTasks()
	if err != nil {
		return err
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

	return nil
}

const SCHEDULED_TASK_PROGRAM_ID = "Schedule.Service.1"

// S_FALSE is returned by CoInitialize if it was already called on this thread.
const S_FALSE = 0x00000001

func getScheduledTasks() (scheduledTasks ScheduledTasks, err error) {
	// The only way to run WMI queries in parallel while being thread-safe is to
	// ensure the CoInitialize[Ex]() call is bound to its current OS thread.
	// Otherwise, attempting to initialize and run parallel queries across
	// goroutines will result in protected memory errors.
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	if err := ole.CoInitializeEx(0, ole.COINIT_MULTITHREADED); err != nil {
		var oleCode *ole.OleError
		if errors.As(err, &oleCode) && oleCode.Code() != ole.S_OK && oleCode.Code() != S_FALSE {
			return nil, err
		}
	}
	defer ole.CoUninitialize()

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
