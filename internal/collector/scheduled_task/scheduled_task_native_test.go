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
	"testing"

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

			err := fetchTasksInFolder(folder, &tasks)
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

	err := fetchTasksRecursively(root, `\`, &tasks)
	require.ErrorIs(t, err, ole.HRESULT(0x80070005))
	require.Equal(t, []int64{1, 2, 3}, *attempted)
	require.Len(t, tasks, 2)
	require.Equal(t, TaskState(1), tasks[0].State)
	require.Equal(t, TaskState(3), tasks[1].State)
	require.Equal(t, 2, tasksReleased)
	require.Equal(t, 2, foldersReleased)
	require.Equal(t, 6, collectionsReleased)
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
	var methods [15]uintptr

	methods[2] = windows.NewCallback(func(uintptr) uintptr {
		if released != nil {
			*released++
		}

		return 0
	})
	methods[8] = windows.NewCallback(func(_ uintptr, out *uintptr) uintptr {
		*out = 0 // A null BSTR is a valid empty string.

		return 0
	})
	methods[10] = windows.NewCallback(func(_ uintptr, _ uintptr, out **ole.Object) uintptr {
		*out = folders

		return 0
	})
	methods[14] = windows.NewCallback(func(_ uintptr, _ uintptr, out **ole.Object) uintptr {
		*out = tasks

		return 0
	})

	return &taskschd.TaskFolder{VTable: &methods[0]}
}

func fakeScheduledTask(state int32, released *int) *ole.Object {
	var methods [18]uintptr

	methods[2] = windows.NewCallback(func(uintptr) uintptr {
		*released++

		return 0
	})
	stringGetter := windows.NewCallback(func(_ uintptr, out *uintptr) uintptr {
		*out = 0

		return 0
	})
	methods[7], methods[8] = stringGetter, stringGetter
	methods[9] = windows.NewCallback(func(_ uintptr, out *int32) uintptr {
		*out = state

		return 0
	})
	methods[10] = windows.NewCallback(func(_ uintptr, out *int16) uintptr {
		*out = 0

		return 0
	})
	intGetter := windows.NewCallback(func(_ uintptr, out *int32) uintptr {
		*out = 0

		return 0
	})
	methods[16], methods[17] = intGetter, intGetter

	return &ole.Object{VTable: &methods[0]}
}
