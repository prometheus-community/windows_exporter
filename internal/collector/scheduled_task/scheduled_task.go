// Copyright 2024 The Prometheus Authors
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
	"strings"
	"sync"

	"github.com/alecthomas/kingpin/v2"
	"github.com/go-ole/go-ole"
	"github.com/go-ole/go-ole/oleutil"
	"github.com/prometheus-community/windows_exporter/internal/headers/schedule_service"
	"github.com/prometheus-community/windows_exporter/internal/mi"
	"github.com/prometheus-community/windows_exporter/internal/types"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	Name = "scheduled_task"

	workerCount = 4
)

type Config struct {
	TaskExclude *regexp.Regexp `yaml:"task_exclude"`
	TaskInclude *regexp.Regexp `yaml:"task_include"`
}

//nolint:gochecknoglobals
var ConfigDefaults = Config{
	TaskExclude: types.RegExpEmpty,
	TaskInclude: types.RegExpAny,
}

type Collector struct {
	config Config

	logger *slog.Logger

	scheduledTasksReqCh  chan struct{}
	scheduledTasksWorker chan scheduledTaskWorkerRequest
	scheduledTasksCh     chan scheduledTaskResults

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
	tasks []scheduledTask
	err   error
}

type scheduledTaskWorkerRequest struct {
	folderPath string
	results    chan<- scheduledTaskResults
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

func (c *Collector) Build(logger *slog.Logger, _ *mi.Session) error {
	c.logger = logger.With(slog.String("collector", Name))

	initErrCh := make(chan error)
	c.scheduledTasksReqCh = make(chan struct{})
	c.scheduledTasksCh = make(chan scheduledTaskResults)
	c.scheduledTasksWorker = make(chan scheduledTaskWorkerRequest, 100)

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

	return scheduledTasks.tasks, scheduledTasks.err
}

func (c *Collector) initializeScheduleService(initErrCh chan<- error) {
	service := schedule_service.New()
	if err := service.Connect(); err != nil {
		initErrCh <- fmt.Errorf("failed to connect to schedule service: %w", err)

		return
	}

	defer service.Close()

	errs := make([]error, 0, workerCount)

	for range workerCount {
		errCh := make(chan error, workerCount)

		go c.collectWorker(errCh)

		if err := <-errCh; err != nil {
			errs = append(errs, err)
		}
	}

	if err := errors.Join(errs...); err != nil {
		initErrCh <- err

		return
	}

	close(initErrCh)

	taskServiceObj := service.GetOLETaskServiceObj()
	scheduledTasks := make([]scheduledTask, 0, 500)

	for range c.scheduledTasksReqCh {
		func() {
			// Clear the slice to avoid memory leaks
			clear(scheduledTasks)
			scheduledTasks = scheduledTasks[:0]

			res, err := oleutil.CallMethod(taskServiceObj, "GetFolder", `\`)
			if err != nil {
				c.scheduledTasksCh <- scheduledTaskResults{err: err}
			}

			rootFolderObj := res.ToIDispatch()
			defer rootFolderObj.Release()

			errs := make([]error, 0)
			scheduledTasksWorkerResults := make(chan scheduledTaskResults)

			wg := &sync.WaitGroup{}

			go func() {
				for workerResults := range scheduledTasksWorkerResults {
					wg.Done()

					if workerResults.err != nil {
						errs = append(errs, workerResults.err)
					}

					if workerResults.tasks != nil {
						errs = append(errs, workerResults.err)

						scheduledTasks = append(scheduledTasks, workerResults.tasks...)
					}
				}
			}()

			if err := c.fetchRecursively(rootFolderObj, wg, scheduledTasksWorkerResults); err != nil {
				errs = append(errs, err)
			}

			wg.Wait()

			close(scheduledTasksWorkerResults)

			c.scheduledTasksCh <- scheduledTaskResults{tasks: scheduledTasks, err: errors.Join(errs...)}
		}()
	}

	close(c.scheduledTasksCh)
	close(c.scheduledTasksWorker)

	c.scheduledTasksCh = nil
	c.scheduledTasksWorker = nil
}

func (c *Collector) collectWorker(errCh chan<- error) {
	defer func() {
		if r := recover(); r != nil {
			c.logger.Error("worker panic",
				slog.Any("panic", r),
			)

			errCh := make(chan error, 1)
			// Restart the collectWorker
			go c.collectWorker(errCh)

			if err := <-errCh; err != nil {
				c.logger.Error("failed to restart worker",
					slog.Any("err", err),
				)
			}
		}
	}()

	service := schedule_service.New()
	if err := service.Connect(); err != nil {
		errCh <- fmt.Errorf("failed to connect to schedule service: %w", err)

		return
	}

	close(errCh)

	defer service.Close()

	taskServiceObj := service.GetOLETaskServiceObj()

	for task := range c.scheduledTasksWorker {
		scheduledTasks, err := fetchTasksInFolder(taskServiceObj, task.folderPath)

		task.results <- scheduledTaskResults{tasks: scheduledTasks, err: err}
	}
}

func (c *Collector) fetchRecursively(folder *ole.IDispatch, wg *sync.WaitGroup, results chan<- scheduledTaskResults) error {
	folderPathVariant, err := oleutil.GetProperty(folder, "Path")
	if err != nil {
		return fmt.Errorf("failed to get folder path: %w", err)
	}

	folderPath := folderPathVariant.ToString()

	wg.Add(1)
	c.scheduledTasksWorker <- scheduledTaskWorkerRequest{folderPath: folderPath, results: results}

	res, err := oleutil.CallMethod(folder, "GetFolders", 1)
	if err != nil {
		return err
	}

	subFolders := res.ToIDispatch()
	defer subFolders.Release()

	return oleutil.ForEach(subFolders, func(v *ole.VARIANT) error {
		subFolder := v.ToIDispatch()
		defer subFolder.Release()

		return c.fetchRecursively(subFolder, wg, results)
	})
}

func fetchTasksInFolder(taskServiceObj *ole.IDispatch, folderPath string) ([]scheduledTask, error) {
	folderObjRes, err := oleutil.CallMethod(taskServiceObj, "GetFolder", folderPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get folder %s: %w", folderPath, err)
	}

	folderObj := folderObjRes.ToIDispatch()
	defer folderObj.Release()

	tasksRes, err := oleutil.CallMethod(folderObj, "GetTasks", 1)
	if err != nil {
		return nil, fmt.Errorf("failed to get tasks in folder %s: %w", folderPath, err)
	}

	tasks := tasksRes.ToIDispatch()
	defer tasks.Release()

	// Get task count
	countVariant, err := oleutil.GetProperty(tasks, "Count")
	if err != nil {
		return nil, fmt.Errorf("failed to get task count: %w", err)
	}

	taskCount := int(countVariant.Val)

	scheduledTasks := make([]scheduledTask, 0, taskCount)

	err = oleutil.ForEach(tasks, func(v *ole.VARIANT) error {
		task := v.ToIDispatch()
		defer task.Release()

		parsedTask, err := parseTask(task)
		if err != nil {
			return err
		}

		scheduledTasks = append(scheduledTasks, parsedTask)

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to iterate over tasks: %w", err)
	}

	return scheduledTasks, nil
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
