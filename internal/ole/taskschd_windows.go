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

package ole

import (
	"iter"
	"runtime"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

// Native method slots, including the seven inherited IUnknown/IDispatch slots.
// https://learn.microsoft.com/en-us/windows/win32/api/taskschd/
const (
	taskServiceGetFolder     = 7
	taskServiceConnect       = 10
	taskFolderPath           = 8
	taskFolderGetFolders     = 10
	taskFolderGetTasks       = 14
	taskCollectionCount      = 7
	taskCollectionItem       = 8
	registeredTaskName       = 7
	registeredTaskPath       = 8
	registeredTaskState      = 9
	registeredTaskEnabled    = 10
	registeredTaskLastResult = 16
	registeredTaskMissedRuns = 17
)

type (
	TaskService    struct{ object }
	TaskFolder     struct{ object }
	RegisteredTask struct{ object }
)

func NewTaskService() (*TaskService, error) {
	return create[TaskService](
		windows.GUID{Data1: 0x0f87369f, Data2: 0xa4e5, Data3: 0x4cfc, Data4: [8]byte{0xbd, 0x3e, 0x73, 0xe6, 0x15, 0x45, 0x72, 0xdd}},
		windows.GUID{Data1: 0x2faba4c7, Data2: 0x4da9, Data3: 0x4013, Data4: [8]byte{0x96, 0x97, 0x20, 0xcc, 0x3f, 0xd4, 0x0f, 0x85}},
	)
}

// Connect connects locally using the current security token. Empty VARIANTs
// supply the four optional arguments. The 24-byte values are passed indirectly
// on Windows amd64 and arm64, each using distinct caller-owned storage.
func (s *TaskService) Connect() error {
	var server, user, domain, password variantStorage

	hr, _, _ := syscall.SyscallN(
		s.method(taskServiceConnect),
		uintptr(unsafe.Pointer(s)),
		uintptr(unsafe.Pointer(server.variant())),
		uintptr(unsafe.Pointer(user.variant())),
		uintptr(unsafe.Pointer(domain.variant())),
		uintptr(unsafe.Pointer(password.variant())),
	)
	runtime.KeepAlive(s)

	return resultError(hr)
}

func (s *TaskService) Folder(path string) (*TaskFolder, error) {
	value, err := newBSTR(path)
	if err != nil {
		return nil, err
	}
	defer value.free()

	return s.getArg[*TaskFolder](taskServiceGetFolder, uintptr(unsafe.Pointer(value.ptr)))
}

// Folders includes all subfolders. Its reserved flags argument must be zero.
func (f *TaskFolder) Folders() (*taskCollection[TaskFolder], error) {
	return f.getArg[*taskCollection[TaskFolder]](taskFolderGetFolders, 0)
}

// Tasks includes hidden tasks (TASK_ENUM_HIDDEN).
func (f *TaskFolder) Tasks() (*taskCollection[RegisteredTask], error) {
	return f.getArg[*taskCollection[RegisteredTask]](taskFolderGetTasks, 1)
}

func (f *TaskFolder) Path() (string, error) { return f.string(taskFolderPath) }

func (t *RegisteredTask) Name() (string, error) { return t.string(registeredTaskName) }
func (t *RegisteredTask) Path() (string, error) { return t.string(registeredTaskPath) }
func (t *RegisteredTask) State() (int32, error) { return t.get[int32](registeredTaskState) }
func (t *RegisteredTask) Enabled() (bool, error) {
	value, err := t.get[int16](registeredTaskEnabled)

	return value != 0, err
}

func (t *RegisteredTask) LastTaskResult() (int32, error) {
	return t.get[int32](registeredTaskLastResult)
}

func (t *RegisteredTask) NumberOfMissedRuns() (int32, error) {
	return t.get[int32](registeredTaskMissedRuns)
}

type taskCollection[T TaskFolder | RegisteredTask] struct{ object }

func (c *taskCollection[T]) count() (int32, error) { return c.get[int32](taskCollectionCount) }

func (c *taskCollection[T]) item(index int32) (*T, error) {
	var storage variantStorage

	value := storage.variant()
	*value = variant{vt: 3, value: int64(index) + 1} // VT_I4, one-based indices

	var item *T

	hr, _, _ := syscall.SyscallN(
		c.method(taskCollectionItem),
		uintptr(unsafe.Pointer(c)),
		uintptr(unsafe.Pointer(value)),
		uintptr(unsafe.Pointer(&item)),
	)
	runtime.KeepAlive(c)

	return item, resultError(hr)
}

// All yields borrowed interfaces, valid only in the loop body. It releases each
// item on advance, early exit, or panic; callers must not release the items.
// The collection remains owned by the caller and must be released separately.
func (c *taskCollection[T]) All() iter.Seq2[*T, error] {
	return borrowedItems(c.count, c.item, func(item *T) { (*object)(unsafe.Pointer(item)).Release() })
}
