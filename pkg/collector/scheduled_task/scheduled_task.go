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
	Name                        = "scheduled_task"
	FlagScheduledTaskOldExclude = "collector.scheduled_task.blacklist"
	FlagScheduledTaskOldInclude = "collector.scheduled_task.whitelist"

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

type collector struct {
	logger log.Logger

	taskOldExclude *string
	taskOldInclude *string

	taskExclude *string
	taskInclude *string

	taskIncludeSet bool
	taskExcludeSet bool

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

func New(logger log.Logger, config *Config) types.Collector {
	if config == nil {
		config = &ConfigDefaults
	}

	c := &collector{
		taskExclude: &config.TaskExclude,
		taskInclude: &config.TaskInclude,
	}
	c.SetLogger(logger)
	return c
}

func NewWithFlags(app *kingpin.Application) types.Collector {
	c := &collector{}

	c.taskInclude = app.Flag(
		FlagScheduledTaskInclude,
		"Regexp of tasks to include. Task path must both match include and not match exclude to be included.",
	).Default(ConfigDefaults.TaskInclude).PreAction(func(_ *kingpin.ParseContext) error {
		c.taskIncludeSet = true
		return nil
	}).String()

	c.taskExclude = app.Flag(
		FlagScheduledTaskExclude,
		"Regexp of tasks to exclude. Task path must both match include and not match exclude to be included.",
	).Default(ConfigDefaults.TaskExclude).PreAction(func(_ *kingpin.ParseContext) error {
		c.taskExcludeSet = true
		return nil
	}).String()

	c.taskOldInclude = app.Flag(
		FlagScheduledTaskOldInclude,
		"DEPRECATED: Use --collector.scheduled_task.include",
	).Hidden().String()
	c.taskOldExclude = app.Flag(
		FlagScheduledTaskOldExclude,
		"DEPRECATED: Use --collector.scheduled_task.exclude",
	).Hidden().String()

	return c
}

func (c *collector) GetName() string {
	return Name
}

func (c *collector) SetLogger(logger log.Logger) {
	c.logger = log.With(logger, "collector", Name)
}

func (c *collector) GetPerfCounter() ([]string, error) {
	return []string{}, nil
}

func (c *collector) Build() error {
	if *c.taskOldExclude != "" {
		if !c.taskExcludeSet {
			_ = level.Warn(c.logger).Log("msg", "--collector.scheduled_task.blacklist is DEPRECATED and will be removed in a future release, use --collector.scheduled_task.exclude")
			*c.taskExclude = *c.taskOldExclude
		} else {
			return errors.New("--collector.scheduled_task.blacklist and --collector.scheduled_task.exclude are mutually exclusive")
		}
	}
	if *c.taskOldInclude != "" {
		if !c.taskIncludeSet {
			_ = level.Warn(c.logger).Log("msg", "--collector.scheduled_task.whitelist is DEPRECATED and will be removed in a future release, use --collector.scheduled_task.include")
			*c.taskInclude = *c.taskOldInclude
		} else {
			return errors.New("--collector.scheduled_task.whitelist and --collector.scheduled_task.include are mutually exclusive")
		}
	}

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	err := ole.CoInitializeEx(0, ole.COINIT_MULTITHREADED)
	if err != nil {
		code := err.(*ole.OleError).Code()
		if code != ole.S_OK && code != S_FALSE {
			return err
		}
	}
	defer ole.CoUninitialize()

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

func (c *collector) Collect(_ *types.ScrapeContext, ch chan<- prometheus.Metric) error {
	if desc, err := c.collect(ch); err != nil {
		_ = level.Error(c.logger).Log("failed collecting user metrics", "desc", desc, "err", err)
		return err
	}

	return nil
}

var TASK_STATES = []string{"disabled", "queued", "ready", "running", "unknown"}

func (c *collector) collect(ch chan<- prometheus.Metric) (*prometheus.Desc, error) {
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
