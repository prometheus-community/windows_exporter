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
	"sync"
	"time"

	"github.com/alecthomas/kingpin/v2"
	"github.com/prometheus-community/windows_exporter/internal/mi"
	"github.com/prometheus-community/windows_exporter/internal/ole"
	"github.com/prometheus-community/windows_exporter/internal/ole/taskschd"
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

// ScheduledTask holds the task properties published as metrics.
type ScheduledTask struct {
	Path            string
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
	scheduledTasks, err := getScheduledTasks(c.includeTask)
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

// includeTask reports whether a task path matches include and does not match exclude.
func (c *Collector) includeTask(path string) bool {
	return c.config.TaskInclude.MatchString(path) && !c.config.TaskExclude.MatchString(path)
}

func (c *Collector) collectMetrics(ch chan<- prometheus.Metric, scheduledTasks ScheduledTasks) {
	for _, task := range scheduledTasks {
		for _, state := range TASK_STATES {
			var stateValue float64

			if strings.EqualFold(task.State.String(), state) {
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

// errTasksSkipped accompanies readable tasks when some tasks or folders could not be read.
var errTasksSkipped = errors.New("tasks skipped")

// taskFolderWorkers is the number of folders read concurrently. Almost all of
// a scrape is spent waiting for the Task Scheduler service, which answers
// requests from several connections in parallel.
const taskFolderWorkers = 4

// getScheduledTasks returns the tasks whose slash-separated paths pass include.
// Folders are read concurrently, but tasks are returned in the order of a
// depth-first walk, as if the folders had been read one after another.
func getScheduledTasks(include func(path string) bool) (ScheduledTasks, error) {
	return walkTaskFolders(taskFolderWorkers, func() (taskFolderReader, func(), error) {
		return newTaskFolderReader(include)
	})
}

// taskFolderReader opens the folder at path and reads it. An error means the
// folder could not be opened; errors reading its tasks and subfolders are
// reported in taskFolderContents.
type taskFolderReader func(path string) (taskFolderContents, error)

type taskFolderContents struct {
	tasks      ScheduledTasks
	subfolders []string
	err        error
}

// newTaskFolderReader connects to the Task Scheduler service on the calling
// goroutine's OS thread. The returned reader must be used, and release called,
// on the same goroutine.
func newTaskFolderReader(include func(path string) bool) (taskFolderReader, func(), error) {
	// COM initialization and every interface call stay on the same OS thread.
	runtime.LockOSThread()

	if err := ole.Initialize(); err != nil {
		runtime.UnlockOSThread()

		return nil, nil, err
	}

	service, err := taskschd.NewTaskService()
	if err != nil {
		ole.Uninitialize()
		runtime.UnlockOSThread()

		return nil, nil, fmt.Errorf("create Task Scheduler service: %w", err)
	}

	release := func() {
		service.Release()
		ole.Uninitialize()
		runtime.UnlockOSThread()
	}

	if err := service.Connect(); err != nil {
		release()

		return nil, nil, fmt.Errorf("connect Task Scheduler service: %w", err)
	}

	read := func(path string) (taskFolderContents, error) {
		folder, err := service.Folder(path)
		if err != nil {
			return taskFolderContents{}, err
		}
		defer folder.Release()

		return readTaskFolder(folder, include), nil
	}

	return read, release, nil
}

// readTaskFolder reads the tasks and the subfolder paths of a folder, and
// reports errors after reading the remaining tasks and subfolders.
func readTaskFolder(folder *taskschd.TaskFolder, include func(path string) bool) taskFolderContents {
	var contents taskFolderContents

	errs := []error{}
	if err := fetchTasksInFolder(folder, include, &contents.tasks); err != nil {
		errs = append(errs, err)
	}

	folders, err := folder.Folders()
	if err != nil {
		contents.err = errors.Join(append(errs, fmt.Errorf("get sub folders: %w", err))...)

		return contents
	}
	defer folders.Release()

	for subfolder, err := range folders.All() {
		if err != nil {
			errs = append(errs, fmt.Errorf("enumerate sub folders: %w", err))

			continue
		}

		path, err := subfolder.Path()
		if err != nil {
			errs = append(errs, fmt.Errorf("get sub folder path: %w", err))

			continue
		}

		contents.subfolders = append(contents.subfolders, path)
	}

	contents.err = errors.Join(errs...)

	return contents
}

// fetchTasksInFolder appends readable tasks and reports errors after reading the remaining tasks.
func fetchTasksInFolder(folder *taskschd.TaskFolder, include func(path string) bool, scheduledTasks *ScheduledTasks) error {
	tasks, err := folder.Tasks()
	if err != nil {
		return fmt.Errorf("get tasks: %w", err)
	}
	defer tasks.Release()

	errs := []error{}

	for task, err := range tasks.All() {
		if err != nil {
			errs = append(errs, fmt.Errorf("enumerate tasks: %w", err))

			continue
		}

		parsedTask, ok, err := parseTask(task, include)
		if err != nil {
			errs = append(errs, fmt.Errorf("parse task: %w", err))

			continue
		}

		if !ok {
			continue
		}

		*scheduledTasks = append(*scheduledTasks, parsedTask)
	}

	return errors.Join(errs...)
}

// taskFolderNode is a folder of the walk. Only the worker reading the folder
// writes its fields, before its subfolders are queued.
type taskFolderNode struct {
	path     string
	read     bool
	openErr  error
	contents taskFolderContents
	children []*taskFolderNode
}

// taskFolderQueue holds the folders waiting for a worker.
type taskFolderQueue struct {
	mu    sync.Mutex
	ready sync.Cond
	queue []*taskFolderNode
	// pending counts the folders that are queued or being read.
	pending int
}

// walkTaskFolders reads the folder tree from the root folder with up to
// workers goroutines. newReader runs on each worker goroutine. Workers that
// fail to start don't take part; their errors are returned only if no worker
// read the root folder.
func walkTaskFolders(workers int, newReader func() (taskFolderReader, func(), error)) (ScheduledTasks, error) {
	root := &taskFolderNode{path: `\`}
	queue := &taskFolderQueue{queue: []*taskFolderNode{root}, pending: 1}
	queue.ready.L = &queue.mu

	startErrs := make([]error, workers)

	var wg sync.WaitGroup

	for i := range workers {
		wg.Go(func() {
			startErrs[i] = queue.work(newReader)
		})
	}

	wg.Wait()

	if !root.read {
		return nil, errors.Join(startErrs...)
	}

	if root.openErr != nil {
		return nil, fmt.Errorf("get root task folder: %w", root.openErr)
	}

	tasks := ScheduledTasks{}
	errs := []error{}

	root.collect(&tasks, &errs)

	if err := errors.Join(errs...); err != nil {
		return tasks, fmt.Errorf("%w: %w", errTasksSkipped, err)
	}

	return tasks, nil
}

// work reads queued folders until all folders are read. A panic stops only
// this worker; the folder it was reading is reported as failed.
func (q *taskFolderQueue) work(newReader func() (taskFolderReader, func(), error)) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic in task folder worker: %v", r)
		}
	}()

	read, release, err := newReader()
	if err != nil {
		return err
	}
	defer release()

	for {
		node := q.next()
		if node == nil {
			return nil
		}

		// After a panic, the state of this worker's COM objects is unknown.
		if err := q.read(node, read); err != nil {
			return err
		}
	}
}

// next returns the next folder to read, or nil when all folders are read.
func (q *taskFolderQueue) next() *taskFolderNode {
	q.mu.Lock()
	defer q.mu.Unlock()

	for len(q.queue) == 0 && q.pending > 0 {
		q.ready.Wait()
	}

	if len(q.queue) == 0 {
		return nil
	}

	node := q.queue[0]
	q.queue = q.queue[1:]

	return node
}

// read reads a folder and queues its subfolders. A panic in read is recorded
// as the folder's error and returned.
func (q *taskFolderQueue) read(node *taskFolderNode, read taskFolderReader) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic: %v", r)
			node.openErr = err
			node.children = nil
		}

		node.read = true

		q.mu.Lock()
		q.queue = append(q.queue, node.children...)
		q.pending += len(node.children) - 1
		q.mu.Unlock()
		q.ready.Broadcast()
	}()

	node.contents, node.openErr = read(node.path)

	node.children = make([]*taskFolderNode, 0, len(node.contents.subfolders))
	for _, path := range node.contents.subfolders {
		node.children = append(node.children, &taskFolderNode{path: path})
	}

	return nil
}

// collect appends the tasks and errors of the folder and its subfolders in depth-first order.
func (n *taskFolderNode) collect(tasks *ScheduledTasks, errs *[]error) {
	switch {
	case !n.read:
		*errs = append(*errs, fmt.Errorf("folder %s: not read", n.path))

		return
	case n.openErr != nil:
		*errs = append(*errs, fmt.Errorf("folder %s: get folder: %w", n.path, n.openErr))

		return
	case n.contents.err != nil:
		*errs = append(*errs, fmt.Errorf("folder %s: %w", n.path, n.contents.err))
	}

	*tasks = append(*tasks, n.contents.tasks...)

	for _, child := range n.children {
		child.collect(tasks, errs)
	}
}

// parseTask reads a task that passes include and reports false for filtered
// tasks. Reading a task property other than Path is a round trip to the Task
// Scheduler service, so filtered tasks are skipped before those reads, and only
// properties that collectMetrics publishes are read.
func parseTask(task *taskschd.RegisteredTask, include func(path string) bool) (ScheduledTask, bool, error) {
	path, err := task.Path()
	if err != nil {
		return ScheduledTask{}, false, fmt.Errorf("get task path: %w", err)
	}

	parsedTask := ScheduledTask{Path: strings.ReplaceAll(path, "\\", "/")}
	if !include(parsedTask.Path) {
		return ScheduledTask{}, false, nil
	}

	state, err := task.State()
	if err != nil {
		return ScheduledTask{}, false, fmt.Errorf("get task state: %w", err)
	}

	parsedTask.State = TaskState(state)

	result, err := task.LastTaskResult()
	if err != nil {
		return ScheduledTask{}, false, fmt.Errorf("get task last result: %w", err)
	}

	parsedTask.LastTaskResult = TaskResult(result)

	// collectMetrics omits missed runs for tasks that have not run.
	if parsedTask.LastTaskResult != SCHED_S_TASK_HAS_NOT_RUN {
		missed, err := task.NumberOfMissedRuns()
		if err != nil {
			return ScheduledTask{}, false, fmt.Errorf("get task missed runs: %w", err)
		}

		parsedTask.MissedRunsCount = float64(missed)
	}

	return parsedTask, true, nil
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
