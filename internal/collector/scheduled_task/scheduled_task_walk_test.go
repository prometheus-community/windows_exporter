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
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// fakeFolder is a folder of a fake tree read by walkTaskFolders.
type fakeFolder struct {
	tasks      []string
	subfolders []string
	delay      time.Duration
	openErr    error
	err        error
	panic      bool
}

// fakeReaders returns a newReader for walkTaskFolders that reads tree and
// counts started and released readers.
func fakeReaders(tree map[string]fakeFolder, startErrs ...error) (func() (taskFolderReader, func(), error), *atomic.Int32, *atomic.Int32) {
	var started, released atomic.Int32

	newReader := func() (taskFolderReader, func(), error) {
		n := int(started.Add(1)) - 1
		if n < len(startErrs) && startErrs[n] != nil {
			return nil, nil, startErrs[n]
		}

		read := func(path string) (taskFolderContents, error) {
			folder, ok := tree[path]
			if !ok {
				return taskFolderContents{}, fmt.Errorf("unknown folder %s", path)
			}

			time.Sleep(folder.delay)

			if folder.panic {
				panic("fake panic")
			}

			if folder.openErr != nil {
				return taskFolderContents{}, folder.openErr
			}

			contents := taskFolderContents{subfolders: folder.subfolders, err: folder.err}
			for _, task := range folder.tasks {
				contents.tasks = append(contents.tasks, ScheduledTask{Path: task})
			}

			return contents, nil
		}

		return read, func() { released.Add(1) }, nil
	}

	return newReader, &started, &released
}

func taskPaths(tasks ScheduledTasks) []string {
	paths := make([]string, 0, len(tasks))
	for _, task := range tasks {
		paths = append(paths, task.Path)
	}

	return paths
}

func TestWalkTaskFoldersOrder(t *testing.T) {
	t.Parallel()

	// Folders that are read late must not change the depth-first order.
	tree := map[string]fakeFolder{
		`\`:      {tasks: []string{"/T0"}, subfolders: []string{`\A`, `\B`, `\C`}},
		`\A`:     {tasks: []string{"/A/T1"}, subfolders: []string{`\A\X`, `\A\Y`}, delay: 10 * time.Millisecond},
		`\A\X`:   {tasks: []string{"/A/X/T2", "/A/X/T3"}, subfolders: []string{`\A\X\Z`}, delay: 20 * time.Millisecond},
		`\A\X\Z`: {tasks: []string{"/A/X/Z/T4"}},
		`\A\Y`:   {},
		`\B`:     {tasks: []string{"/B/T5"}},
		`\C`:     {tasks: []string{"/C/T6"}, delay: 5 * time.Millisecond},
	}

	for _, workers := range []int{1, 2, 4, 8} {
		t.Run(fmt.Sprintf("workers=%d", workers), func(t *testing.T) {
			t.Parallel()

			newReader, started, released := fakeReaders(tree)

			tasks, err := walkTaskFolders(workers, newReader)
			require.NoError(t, err)
			require.Equal(t, []string{"/T0", "/A/T1", "/A/X/T2", "/A/X/T3", "/A/X/Z/T4", "/B/T5", "/C/T6"}, taskPaths(tasks))
			require.Equal(t, int32(workers), started.Load())
			require.Equal(t, int32(workers), released.Load())
		})
	}
}

func TestWalkTaskFoldersErrors(t *testing.T) {
	t.Parallel()

	errAccess := errors.New("access denied")
	errTask := errors.New("parse task")

	tree := map[string]fakeFolder{
		`\`:  {tasks: []string{"/T0"}, subfolders: []string{`\A`, `\B`, `\C`}},
		`\A`: {tasks: []string{"/A/T1"}, err: errTask},
		`\B`: {openErr: errAccess},
		`\C`: {tasks: []string{"/C/T2"}},
	}

	newReader, _, _ := fakeReaders(tree)

	tasks, err := walkTaskFolders(4, newReader)
	require.ErrorIs(t, err, errTasksSkipped)
	require.ErrorIs(t, err, errAccess)
	require.ErrorIs(t, err, errTask)
	require.ErrorContains(t, err, `folder \A: parse task`)
	require.ErrorContains(t, err, `folder \B: get folder: access denied`)
	require.Equal(t, []string{"/T0", "/A/T1", "/C/T2"}, taskPaths(tasks))
}

func TestWalkTaskFoldersRootError(t *testing.T) {
	t.Parallel()

	errAccess := errors.New("access denied")
	newReader, _, _ := fakeReaders(map[string]fakeFolder{`\`: {openErr: errAccess}})

	tasks, err := walkTaskFolders(4, newReader)
	require.ErrorIs(t, err, errAccess)
	require.NotErrorIs(t, err, errTasksSkipped)
	require.ErrorContains(t, err, "get root task folder")
	require.Nil(t, tasks)
}

func TestWalkTaskFoldersStartErrors(t *testing.T) {
	t.Parallel()

	tree := map[string]fakeFolder{
		`\`:  {tasks: []string{"/T0"}, subfolders: []string{`\A`}},
		`\A`: {tasks: []string{"/A/T1"}},
	}
	errConnect := errors.New("connect")

	t.Run("some workers", func(t *testing.T) {
		t.Parallel()

		newReader, _, released := fakeReaders(tree, errConnect, nil, errConnect)

		tasks, err := walkTaskFolders(3, newReader)
		require.NoError(t, err)
		require.Equal(t, []string{"/T0", "/A/T1"}, taskPaths(tasks))
		require.Equal(t, int32(1), released.Load())
	})

	t.Run("all workers", func(t *testing.T) {
		t.Parallel()

		newReader, _, released := fakeReaders(tree, errConnect, errConnect)

		tasks, err := walkTaskFolders(2, newReader)
		require.ErrorIs(t, err, errConnect)
		require.NotErrorIs(t, err, errTasksSkipped)
		require.Nil(t, tasks)
		require.Zero(t, released.Load())
	})
}

func TestWalkTaskFoldersPanic(t *testing.T) {
	t.Parallel()

	tree := map[string]fakeFolder{
		`\`:  {tasks: []string{"/T0"}, subfolders: []string{`\A`, `\B`}},
		`\A`: {panic: true, subfolders: []string{`\A\X`}},
		`\B`: {tasks: []string{"/B/T1"}, delay: 10 * time.Millisecond},
	}

	t.Run("other workers continue", func(t *testing.T) {
		t.Parallel()

		newReader, _, released := fakeReaders(tree)

		tasks, err := walkTaskFolders(2, newReader)
		require.ErrorIs(t, err, errTasksSkipped)
		require.ErrorContains(t, err, `folder \A: get folder: panic: fake panic`)
		require.Equal(t, []string{"/T0", "/B/T1"}, taskPaths(tasks))
		require.Equal(t, int32(2), released.Load())
	})

	t.Run("last worker", func(t *testing.T) {
		t.Parallel()

		newReader, _, released := fakeReaders(tree)

		tasks, err := walkTaskFolders(1, newReader)
		require.ErrorIs(t, err, errTasksSkipped)
		require.ErrorContains(t, err, `folder \A: get folder: panic: fake panic`)
		require.ErrorContains(t, err, `folder \B: not read`)
		require.Equal(t, []string{"/T0"}, taskPaths(tasks))
		require.Equal(t, int32(1), released.Load())
	})
}
