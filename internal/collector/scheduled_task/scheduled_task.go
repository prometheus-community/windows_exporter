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
	"github.com/prometheus-community/windows_exporter/internal/mi"
	"github.com/prometheus-community/windows_exporter/internal/types"
	"github.com/prometheus/client_golang/prometheus"
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

	scheduledTasksReqCh chan struct{}
	scheduledTasksCh    chan *scheduledTaskResults

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

var taskStates = []string{"disabled", "queued", "ready", "running", "unknown"}

type scheduledTask struct {
	Name            string
	Path            string
	Enabled         bool
	State           TaskState
	MissedRunsCount float64
	LastTaskResult  TaskResult
}

type scheduledTaskResults struct {
	scheduledTasks []scheduledTask
	err            error
}

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
	close(c.scheduledTasksReqCh)

	c.scheduledTasksReqCh = nil

	return nil
}

func (c *Collector) Build(_ *slog.Logger, _ *mi.Session) error {
	initErrCh := make(chan error)
	c.scheduledTasksReqCh = make(chan struct{})
	c.scheduledTasksCh = make(chan *scheduledTaskResults)

	go c.initializeScheduleService(initErrCh)

	if err := <-initErrCh; err != nil {
		return fmt.Errorf("initialize schedule service: %w", err)
	}

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

func (c *Collector) Collect(ch chan<- prometheus.Metric) error {
	scheduledTasks, err := c.getScheduledTasks()
	if err != nil {
		return fmt.Errorf("get scheduled tasks: %w", err)
	}

	for _, task := range scheduledTasks {
		if c.config.TaskExclude.MatchString(task.Path) ||
			!c.config.TaskInclude.MatchString(task.Path) {
			continue
		}

		for _, state := range taskStates {
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

func (c *Collector) getScheduledTasks() ([]scheduledTask, error) {
	c.scheduledTasksReqCh <- struct{}{}

	scheduledTasks, ok := <-c.scheduledTasksCh

	if !ok {
		return []scheduledTask{}, nil
	}

	if scheduledTasks == nil {
		return nil, errors.New("scheduled tasks channel is nil")
	}

	if scheduledTasks.err != nil {
		return nil, scheduledTasks.err
	}

	return scheduledTasks.scheduledTasks, scheduledTasks.err
}

func (c *Collector) initializeScheduleService(initErrCh chan<- error) {
	// The only way to run WMI queries in parallel while being thread-safe is to
	// ensure the CoInitialize[Ex]() call is bound to its current OS thread.
	// Otherwise, attempting to initialize and run parallel queries across
	// goroutines will result in protected memory errors.
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	if err := ole.CoInitializeEx(0, ole.COINIT_MULTITHREADED); err != nil {
		var oleCode *ole.OleError
		if errors.As(err, &oleCode) && oleCode.Code() != ole.S_OK && oleCode.Code() != 0x00000001 {
			initErrCh <- err

			return
		}
	}

	defer ole.CoUninitialize()

	scheduleClassID, err := ole.ClassIDFrom("Schedule.Service.1")
	if err != nil {
		initErrCh <- err

		return
	}

	taskSchedulerObj, err := ole.CreateInstance(scheduleClassID, nil)
	if err != nil || taskSchedulerObj == nil {
		initErrCh <- err

		return
	}
	defer taskSchedulerObj.Release()

	taskServiceObj := taskSchedulerObj.MustQueryInterface(ole.IID_IDispatch)
	defer taskServiceObj.Release()

	taskService, err := oleutil.CallMethod(taskServiceObj, "Connect")
	if err != nil {
		initErrCh <- err

		return
	}

	defer func(taskService *ole.VARIANT) {
		_ = taskService.Clear()
	}(taskService)

	close(initErrCh)

	scheduledTasks := make([]scheduledTask, 0, 100)

	for range c.scheduledTasksReqCh {
		func() {
			// Clear the slice to avoid memory leaks
			clear(scheduledTasks)
			scheduledTasks = scheduledTasks[:0]

			res, err := oleutil.CallMethod(taskServiceObj, "GetFolder", `\`)
			if err != nil {
				c.scheduledTasksCh <- &scheduledTaskResults{err: err}

				return
			}

			rootFolderObj := res.ToIDispatch()
			defer rootFolderObj.Release()

			err = fetchTasksRecursively(rootFolderObj, &scheduledTasks)

			c.scheduledTasksCh <- &scheduledTaskResults{scheduledTasks: scheduledTasks, err: err}
		}()
	}

	close(c.scheduledTasksCh)

	c.scheduledTasksCh = nil
}

func fetchTasksRecursively(folder *ole.IDispatch, scheduledTasks *[]scheduledTask) error {
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

func fetchTasksInFolder(folder *ole.IDispatch, scheduledTasks *[]scheduledTask) error {
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

func parseTask(task *ole.IDispatch) (scheduledTask, error) {
	var scheduledTask scheduledTask

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
