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

//go:build windows && (amd64 || arm64)

package scheduled_task

import (
	"fmt"
	"log/slog"
	"regexp"
	"slices"
	"strings"
	"testing"
	"unsafe"

	"github.com/prometheus-community/windows_exporter/internal/ole"
	"github.com/prometheus-community/windows_exporter/internal/ole/taskschd"
	"github.com/stretchr/testify/require"
	"golang.org/x/sys/windows"
)

func TestFetchTasksInFolderPartialCollection(t *testing.T) {
	for _, tc := range []struct {
		name       string
		countError bool
	}{
		{name: "skip unreadable item"},
		{name: "count failure", countError: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var released, collectionsReleased int

			items := []*ole.Object{fakeScheduledTask(1, &released), nil, fakeScheduledTask(3, &released)}
			collection, attempted := fakeTaskCollection(items, tc.countError, &collectionsReleased)
			folder := fakeTaskFolder(collection, nil, nil)
			tasks := ScheduledTasks{}

			err := readAllTasks.fetchTasksInFolder(folder, &tasks)
			require.ErrorIs(t, err, ole.HRESULT(0x80070005))
			require.Equal(t, 1, collectionsReleased)

			if tc.countError {
				require.Empty(t, *attempted)
				require.Empty(t, tasks)
				require.Zero(t, released)

				return
			}

			require.Equal(t, []int64{1, 2, 3}, *attempted)
			require.Len(t, tasks, 2)
			require.Equal(t, TaskState(1), tasks[0].State)
			require.Equal(t, TaskState(3), tasks[1].State)
			require.Equal(t, 2, released)
		})
	}
}

func TestFetchTasksRecursivelyPartialFolderCollection(t *testing.T) {
	var tasksReleased, foldersReleased, collectionsReleased int

	children := make([]*ole.Object, 0, 2)

	for _, state := range []int32{1, 3} {
		tasks, _ := fakeTaskCollection([]*ole.Object{fakeScheduledTask(state, &tasksReleased)}, false, &collectionsReleased)
		folders, _ := fakeTaskCollection(nil, false, &collectionsReleased)
		child := fakeTaskFolder(tasks, folders, &foldersReleased)
		children = append(children, &child.Object)
	}

	folders, attempted := fakeTaskCollection([]*ole.Object{children[0], nil, children[1]}, false, &collectionsReleased)
	emptyTasks, _ := fakeTaskCollection(nil, false, &collectionsReleased)
	root := fakeTaskFolder(emptyTasks, folders, nil)
	tasks := ScheduledTasks{}

	err := readAllTasks.fetchTasksRecursively(root, `\`, &tasks)
	require.ErrorIs(t, err, ole.HRESULT(0x80070005))
	require.Equal(t, []int64{1, 2, 3}, *attempted)
	require.Len(t, tasks, 2)
	require.Equal(t, TaskState(1), tasks[0].State)
	require.Equal(t, TaskState(3), tasks[1].State)
	require.Equal(t, 2, tasksReleased)
	require.Equal(t, 2, foldersReleased)
	require.Equal(t, 6, collectionsReleased)
}

func TestFetchTasksRecursivelySkipsFilteredFolders(t *testing.T) {
	for _, tc := range []struct {
		name            string
		config          *Config
		paths           []string
		listedMicrosoft int
	}{
		{name: "default exclude", paths: []string{"/Task", "/Vendor/Task"}},
		{
			name:   "empty exclude",
			config: &Config{TaskExclude: regexp.MustCompile(`^(?:)$`)},
			paths:  []string{"/Task", "/Microsoft/Task", "/Vendor/Task"}, listedMicrosoft: 2,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var released, collectionsReleased, foldersReleased, listedMicrosoft, listedVendor int

			fakes := []*fakeTask{
				{path: `\Microsoft\Task`, state: int32(TASK_STATE_READY)},
				{path: `\Vendor\Task`, state: int32(TASK_STATE_READY)},
				{path: `\Task`, state: int32(TASK_STATE_READY)},
			}

			folder := func(path string, task *fakeTask, listed *int) *ole.Object {
				tasks, _ := fakeTaskCollection([]*ole.Object{task.object(&released)}, false, &collectionsReleased)
				folders, _ := fakeTaskCollection(nil, false, &collectionsReleased)

				return &fakeTaskFolderAt(path, tasks, folders, &foldersReleased, listed).Object
			}

			subfolders, _ := fakeTaskCollection([]*ole.Object{
				folder(`\Microsoft`, fakes[0], &listedMicrosoft),
				folder(`\Vendor`, fakes[1], &listedVendor),
			}, false, &collectionsReleased)
			rootTasks, _ := fakeTaskCollection([]*ole.Object{fakes[2].object(&released)}, false, &collectionsReleased)
			root := fakeTaskFolderAt(`\`, rootTasks, subfolders, nil, nil)

			c := New(tc.config)
			require.NoError(t, c.Build(slog.New(slog.DiscardHandler), nil))

			tasks := ScheduledTasks{}

			require.NoError(t, c.taskReader().fetchTasksRecursively(root, `\`, &tasks))

			paths := make([]string, 0, len(tasks))
			for _, task := range tasks {
				require.Equal(t, TASK_STATE_READY, task.State)

				paths = append(paths, task.Path)
			}

			require.ElementsMatch(t, tc.paths, paths)

			// An excluded folder is released without listing its tasks or subfolders.
			require.Equal(t, tc.listedMicrosoft, listedMicrosoft)
			require.Equal(t, 2, listedVendor)
			require.Equal(t, 2, foldersReleased)
			require.Equal(t, tc.listedMicrosoft/2, fakes[0].reads[taskSlotPath])
			require.Equal(t, 1, fakes[1].reads[taskSlotState])
		})
	}
}

func TestFetchTasksFilters(t *testing.T) {
	for _, tc := range []struct {
		name   string
		config *Config
		paths  []string
	}{
		{name: "defaults", config: &Config{}, paths: []string{"/Included/Task", "/Included/Excluded", "/Other/Task"}},
		{
			name:   "include",
			config: &Config{TaskInclude: regexp.MustCompile(`^/Included/`)},
			paths:  []string{"/Included/Task", "/Included/Excluded"},
		},
		{
			name: "exclude takes precedence",
			config: &Config{
				TaskInclude: regexp.MustCompile(`^/Included/`),
				TaskExclude: regexp.MustCompile(`Excluded$`),
			},
			paths: []string{"/Included/Task"},
		},
		{
			name:   "all excluded",
			config: &Config{TaskExclude: regexp.MustCompile(`.*`)},
			paths:  []string{},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var released, collectionsReleased int

			fakes := []*fakeTask{
				{path: `\Included\Task`, state: int32(TASK_STATE_READY)},
				{path: `\Included\Excluded`, state: int32(TASK_STATE_READY)},
				{path: `\Other\Task`, state: int32(TASK_STATE_READY)},
			}
			items := make([]*ole.Object, 0, len(fakes))

			for _, task := range fakes {
				items = append(items, task.object(&released))
			}

			collection, _ := fakeTaskCollection(items, false, &collectionsReleased)
			tasks := ScheduledTasks{}

			tc.config.MissedRuns = true

			err := New(tc.config).taskReader().fetchTasksInFolder(fakeTaskFolder(collection, nil, nil), &tasks)
			require.NoError(t, err)
			require.Equal(t, len(fakes), released)

			// Filtered tasks are skipped before the Task Scheduler service is queried.
			for _, task := range fakes {
				reads := 0
				if slices.Contains(tc.paths, strings.ReplaceAll(task.path, `\`, "/")) {
					reads = 1
				}

				require.Equal(t, 1, task.reads[taskSlotPath], task.path)
				require.Equal(t, reads, task.reads[taskSlotState], task.path)
				require.Equal(t, reads, task.reads[taskSlotLastResult], task.path)
				require.Equal(t, reads, task.reads[taskSlotMissedRuns], task.path)
			}

			families := gatherTaskMetrics(t, tc.config, tasks)
			if len(tc.paths) == 0 {
				require.Empty(t, families)

				return
			}

			require.Len(t, families, 4)

			for _, family := range families {
				paths := make([]string, 0, len(family.GetMetric()))

				for _, metric := range family.GetMetric() {
					for _, label := range metric.GetLabel() {
						if label.GetName() == "task" {
							paths = append(paths, label.GetValue())
						}
					}
				}

				samplesPerTask := 1
				if family.GetName() == "windows_scheduled_task_state" {
					samplesPerTask = 5
				}

				if family.GetName() == "windows_scheduled_task_last_result_status" {
					samplesPerTask = 12
				}

				expected := make([]string, 0, samplesPerTask*len(tc.paths))
				for range samplesPerTask {
					expected = append(expected, tc.paths...)
				}

				require.ElementsMatch(t, expected, paths, "%s", family.GetName())
			}
		})
	}
}

func TestFetchTasksInFolderReadsPublishedProperties(t *testing.T) {
	var released, collectionsReleased int

	ran := &fakeTask{
		path: `\Folder\Ran`, name: "Ran", state: int32(TASK_STATE_READY), enabled: -1, result: 1, missed: 2,
	}
	notRun := &fakeTask{
		path: `\NotRun`, name: "NotRun", state: int32(TASK_STATE_DISABLED),
		result: int32(SCHED_S_TASK_HAS_NOT_RUN), missed: 4,
	}
	collection, _ := fakeTaskCollection(
		[]*ole.Object{ran.object(&released), notRun.object(&released)}, false, &collectionsReleased,
	)
	tasks := ScheduledTasks{}

	require.NoError(t, readAllTasks.fetchTasksInFolder(fakeTaskFolder(collection, nil, nil), &tasks))
	require.Equal(t, ScheduledTasks{
		{Path: "/Folder/Ran", State: TASK_STATE_READY, LastTaskResult: 1, MissedRunsCount: 2},
		{Path: "/NotRun", State: TASK_STATE_DISABLED, LastTaskResult: SCHED_S_TASK_HAS_NOT_RUN},
	}, tasks)

	// Each published property is read once. Name and Enabled are not published,
	// and missed runs are not published for tasks that have not run.
	for _, task := range []*fakeTask{ran, notRun} {
		missedRuns := 1
		if task == notRun {
			missedRuns = 0
		}

		require.Equal(t, 0, task.reads[taskSlotName], task.path)
		require.Equal(t, 1, task.reads[taskSlotPath], task.path)
		require.Equal(t, 1, task.reads[taskSlotState], task.path)
		require.Equal(t, 0, task.reads[taskSlotEnabled], task.path)
		require.Equal(t, 1, task.reads[taskSlotLastResult], task.path)
		require.Equal(t, missedRuns, task.reads[taskSlotMissedRuns], task.path)
	}

	missedRuns := gatherTaskMetrics(t, &Config{MissedRuns: true}, tasks)["windows_scheduled_task_missed_runs"]
	require.NotNil(t, missedRuns)
	require.Len(t, missedRuns.GetMetric(), 1)
	require.Equal(t, "/Folder/Ran", missedRuns.GetMetric()[0].GetLabel()[0].GetValue())
	require.InDelta(t, 2, missedRuns.GetMetric()[0].GetGauge().GetValue(), 0)
}

// Missed runs are not read or published by default.
func TestFetchTasksInFolderSkipsMissedRuns(t *testing.T) {
	var released, collectionsReleased int

	task := &fakeTask{path: `\Task`, state: int32(TASK_STATE_READY), result: 1, missed: 2}
	collection, _ := fakeTaskCollection([]*ole.Object{task.object(&released)}, false, &collectionsReleased)
	config := &Config{}
	tasks := ScheduledTasks{}

	require.NoError(t, New(config).taskReader().fetchTasksInFolder(fakeTaskFolder(collection, nil, nil), &tasks))
	require.Equal(t, ScheduledTasks{{Path: "/Task", State: TASK_STATE_READY, LastTaskResult: 1}}, tasks)
	require.Equal(t, 1, task.reads[taskSlotLastResult])
	require.Zero(t, task.reads[taskSlotMissedRuns])

	families := gatherTaskMetrics(t, config, tasks)
	require.Contains(t, families, "windows_scheduled_task_last_result")
	require.NotContains(t, families, "windows_scheduled_task_missed_runs")
}

// The fake native collections use the Task Scheduler's one-based VARIANT indices.
func fakeTaskCollection(items []*ole.Object, countError bool, released *int) (*ole.Object, *[]int64) {
	var methods [9]uintptr

	attempted := []int64{}
	methods[2] = windows.NewCallback(func(uintptr) uintptr {
		*released++

		return 0
	})
	methods[7] = windows.NewCallback(func(_ uintptr, out *int32) uintptr {
		if countError {
			return 0x80070005
		}

		*out = int32(len(items))

		return 0
	})
	methods[8] = windows.NewCallback(func(_ uintptr, index *ole.Variant, out **ole.Object) uintptr {
		attempted = append(attempted, index.Value)

		item := items[index.Value-1]
		if item == nil {
			return 0x80070005
		}

		*out = item

		return 0
	})

	return &ole.Object{VTable: &methods[0]}, &attempted
}

func fakeTaskFolder(tasks, folders *ole.Object, released *int) *taskschd.TaskFolder {
	return fakeTaskFolderAt("", tasks, folders, released, nil)
}

// fakeTaskFolderAt returns a folder with the given path that counts the
// listings of its tasks and subfolders in listed when it is not nil.
func fakeTaskFolderAt(path string, tasks, folders *ole.Object, released, listed *int) *taskschd.TaskFolder {
	var methods [15]uintptr

	count := func() {
		if listed != nil {
			*listed++
		}
	}

	methods[2] = windows.NewCallback(func(uintptr) uintptr {
		if released != nil {
			*released++
		}

		return 0
	})
	methods[8] = (&fakeTask{}).stringGetter(8, &path) // A null BSTR is a valid empty string.
	methods[10] = windows.NewCallback(func(_ uintptr, _ uintptr, out **ole.Object) uintptr {
		count()

		*out = folders

		return 0
	})
	methods[14] = windows.NewCallback(func(_ uintptr, _ uintptr, out **ole.Object) uintptr {
		count()

		*out = tasks

		return 0
	})

	return &taskschd.TaskFolder{VTable: &methods[0]}
}

func fakeScheduledTask(state int32, released *int) *ole.Object {
	return (&fakeTask{state: state}).object(released)
}

// fakeTask is a native IRegisteredTask that counts the reads of each vtable slot.
type fakeTask struct {
	path, name     string
	state          int32
	enabled        int16
	result, missed int32
	reads          [18]int
}

func (f *fakeTask) object(released *int) *ole.Object {
	var methods [18]uintptr

	methods[2] = windows.NewCallback(func(uintptr) uintptr {
		*released++

		return 0
	})
	methods[taskSlotName] = f.stringGetter(taskSlotName, &f.name)
	methods[taskSlotPath] = f.stringGetter(taskSlotPath, &f.path)
	methods[taskSlotState] = f.int32Getter(taskSlotState, &f.state)
	methods[taskSlotEnabled] = windows.NewCallback(func(_ uintptr, out *int16) uintptr {
		f.reads[taskSlotEnabled]++
		*out = f.enabled

		return 0
	})
	methods[taskSlotLastResult] = f.int32Getter(taskSlotLastResult, &f.result)
	methods[taskSlotMissedRuns] = f.int32Getter(taskSlotMissedRuns, &f.missed)

	return &ole.Object{VTable: &methods[0]}
}

// stringGetter returns a caller-owned BSTR, or a null BSTR for an empty value.
func (f *fakeTask) stringGetter(slot int, value *string) uintptr {
	return windows.NewCallback(func(_ uintptr, out *uintptr) uintptr {
		f.reads[slot]++
		*out = 0

		if *value == "" {
			return 0
		}

		str, err := windows.UTF16PtrFromString(*value)
		if err != nil {
			return 0x80070057 // E_INVALIDARG
		}

		*out, _, _ = sysAllocString.Call(uintptr(unsafe.Pointer(str)))
		if *out == 0 {
			return 0x8007000e // E_OUTOFMEMORY
		}

		return 0
	})
}

func (f *fakeTask) int32Getter(slot int, value *int32) uintptr {
	return windows.NewCallback(func(_ uintptr, out *int32) uintptr {
		f.reads[slot]++
		*out = *value

		return 0
	})
}

// IRegisteredTask vtable slots, including the seven IUnknown/IDispatch slots.
const (
	taskSlotName       = 7
	taskSlotPath       = 8
	taskSlotState      = 9
	taskSlotEnabled    = 10
	taskSlotLastResult = 16
	taskSlotMissedRuns = 17
)

//nolint:gochecknoglobals
var sysAllocString = windows.NewLazySystemDLL("oleaut32.dll").NewProc("SysAllocString")

func includeAllTasks(string) bool { return true }

//nolint:gochecknoglobals
var readAllTasks = taskReader{includeTask: includeAllTasks, includeFolder: includeAllTasks, missedRuns: true}

// BenchmarkGetScheduledTasks measures the Task Scheduler reads without metric creation.
func BenchmarkGetScheduledTasks(b *testing.B) {
	for _, bc := range []struct {
		name    string
		exclude string
	}{
		{name: "all"},
		{name: "exclude Microsoft", exclude: "/Microsoft/.+"},
	} {
		b.Run(bc.name, func(b *testing.B) {
			c := New(&Config{TaskExclude: regexp.MustCompile(fmt.Sprintf("^(?:%s)$", bc.exclude))})
			require.NoError(b, c.Build(slog.New(slog.DiscardHandler), nil))

			for b.Loop() {
				if _, err := c.taskReader().getScheduledTasks(); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
