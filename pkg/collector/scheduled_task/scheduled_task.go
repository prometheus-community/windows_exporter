//go:build windows

package scheduled_task

import (
	"errors"
	"fmt"
	"log/slog"
	"regexp"
	"runtime"
	"strings"

	"github.com/alecthomas/kingpin/v2"
	"github.com/go-ole/go-ole"
	"github.com/go-ole/go-ole/oleutil"
	"github.com/prometheus-community/windows_exporter/pkg/types"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/yusufpapurcu/wmi"
)

const Name = "scheduled_task"

type Config struct {
	TaskExclude *regexp.Regexp `yaml:"task_exclude"`
	TaskInclude *regexp.Regexp `yaml:"task_include"`
}

var ConfigDefaults = Config{
	TaskExclude: types.RegExpEmpty,
	TaskInclude: types.RegExpAny,
}

type Collector struct {
	config Config

	lastResult *prometheus.Desc
	missedRuns *prometheus.Desc
	state      *prometheus.Desc
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
)

const (
	SCHED_S_SUCCESS          TaskResult = 0x0
	SCHED_S_TASK_HAS_NOT_RUN TaskResult = 0x00041303
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
	).Default(c.config.TaskExclude.String()).StringVar(&taskExclude)

	app.Flag(
		"collector.scheduled_task.include",
		"Regexp of tasks to include. Task path must both match include and not match exclude to be included.",
	).Default(c.config.TaskInclude.String()).StringVar(&taskInclude)

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

func (c *Collector) GetPerfCounter(_ *slog.Logger) ([]string, error) {
	return []string{}, nil
}

func (c *Collector) Close(_ *slog.Logger) error {
	return nil
}

func (c *Collector) Build(_ *slog.Logger, _ *wmi.Client) error {
	c.lastResult = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "last_result"),
		"The result that was returned the last time the registered task was run",
		[]string{"task"},
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

func (c *Collector) Collect(_ *types.ScrapeContext, logger *slog.Logger, ch chan<- prometheus.Metric) error {
	logger = logger.With(slog.String("collector", Name))
	if err := c.collect(ch); err != nil {
		logger.Error("failed collecting user metrics",
			slog.Any("err", err),
		)

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

	return nil
}

const SCHEDULED_TASK_PROGRAM_ID = "Schedule.Service.1"

// S_FALSE is returned by CoInitialize if it was already called on this thread.
const S_FALSE = 0x00000001

func getScheduledTasks() (ScheduledTasks, error) {
	var scheduledTasks ScheduledTasks

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
