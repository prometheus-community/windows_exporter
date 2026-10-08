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
	"runtime"
	"testing"
	"unsafe"

	"github.com/stretchr/testify/require"
	"golang.org/x/sys/windows"
)

func TestInitialize(t *testing.T) {
	runtime.LockOSThread()

	defer runtime.UnlockOSThread()

	require.NoError(t, Initialize())

	defer Uninitialize()
	// The second initialization returns S_FALSE; it still needs balancing.
	require.NoError(t, Initialize())

	defer Uninitialize()
}

func TestBSTR(t *testing.T) {
	for _, value := range []string{"", "windows_exporter", "before\x00after", "Unicode \U0001F600"} {
		t.Run(value, func(t *testing.T) {
			str, err := newBSTR(value)
			require.NoError(t, err)

			defer str.free()

			require.Equal(t, value, str.string())
		})
	}

	require.Empty(t, (bstr{}).string())
}

func TestNativeTaskGetters(t *testing.T) {
	var methods [18]uintptr

	// IRegisteredTask::get_State slot from the Windows SDK.
	methods[9] = windows.NewCallback(func(_ uintptr, out *int32) uintptr {
		*out = 4

		return 1 // S_FALSE is a successful HRESULT.
	})
	// IRegisteredTask::get_Enabled slot from the Windows SDK.
	methods[10] = windows.NewCallback(func(_ uintptr, out *int16) uintptr {
		*out = -1

		return 0
	})
	// IRegisteredTask::get_LastTaskResult slot from the Windows SDK.
	methods[16] = windows.NewCallback(func(_ uintptr, out *int32) uintptr {
		*out = -2147216629 // 0x8004130B; preserve the native signed LONG.

		return 0
	})
	// IRegisteredTask::get_NumberOfMissedRuns slot from the Windows SDK.
	methods[17] = windows.NewCallback(func(_ uintptr, _ *int32) uintptr {
		return 0x80070005
	})
	task := &RegisteredTask{vtable: &methods[0]}
	state, err := task.State()
	require.NoError(t, err)
	require.Equal(t, int32(4), state)

	enabled, err := task.Enabled()
	require.NoError(t, err)
	require.True(t, enabled)

	result, err := task.LastTaskResult()
	require.NoError(t, err)
	require.Equal(t, uint32(0x8004130b), uint32(result))

	_, err = task.NumberOfMissedRuns()
	require.ErrorIs(t, err, HRESULT(0x80070005))
}

func TestNativeConnect(t *testing.T) {
	var methods [11]uintptr

	valid := false
	// ITaskService::Connect slot from the Windows SDK.
	methods[10] = windows.NewCallback(func(_ uintptr, server, user, domain, password *variant) uintptr {
		valid = true
		for _, value := range []*variant{server, user, domain, password} {
			valid = valid && *value == (variant{}) && uintptr(unsafe.Pointer(value))%16 == 0
		}

		return 0
	})
	service := &TaskService{vtable: &methods[0]}
	require.NoError(t, service.Connect())
	require.True(t, valid, "Connect arguments must be aligned VT_EMPTY variants")
}

func TestNativeTaskCollection(t *testing.T) {
	var taskMethods [3]uintptr

	released := 0
	taskMethods[2] = windows.NewCallback(func(uintptr) uintptr {
		released++

		return 0
	})
	task := &RegisteredTask{vtable: &taskMethods[0]}

	var methods [9]uintptr

	// IRegisteredTaskCollection::get_Count slot from the Windows SDK.
	methods[7] = windows.NewCallback(func(_ uintptr, out *int32) uintptr {
		*out = 2

		return 0
	})

	var (
		index   int64
		aligned bool
	)

	// IRegisteredTaskCollection::get_Item slot from the Windows SDK.
	methods[8] = windows.NewCallback(func(_ uintptr, value *variant, out **RegisteredTask) uintptr {
		index = value.value
		aligned = value.vt == 3 && uintptr(unsafe.Pointer(value))%16 == 0
		*out = task

		return 0
	})

	collection := &taskCollection[RegisteredTask]{vtable: &methods[0]}
	for item, err := range collection.All() {
		require.NoError(t, err)
		require.Same(t, task, item)
		require.Zero(t, released)

		break
	}

	require.Equal(t, int64(1), index)
	require.True(t, aligned)
	require.Equal(t, 1, released)
}

func TestNativeUpdateCollection(t *testing.T) {
	var methods [11]uintptr

	// IUpdateCollection::get_Count slot from the Windows SDK.
	methods[10] = windows.NewCallback(func(_ uintptr, out *int32) uintptr {
		*out = 1

		return 0
	})

	var index uintptr

	// IUpdateCollection::get_Item slot from the Windows SDK.
	methods[7] = windows.NewCallback(func(_ uintptr, i uintptr, _ **Update) uintptr {
		index = i

		return 0x80070005
	})
	collection := &updateCollection[Update]{vtable: &methods[0]}

	errors := 0
	for item, err := range collection.All() {
		errors++

		require.Nil(t, item)
		require.ErrorIs(t, err, HRESULT(0x80070005))
	}

	require.Equal(t, 1, errors)
	require.Zero(t, index)
}

func TestVariantLayout(t *testing.T) {
	require.Equal(t, uintptr(24), unsafe.Sizeof(variant{}))
	require.Equal(t, uintptr(8), unsafe.Offsetof(variant{}.value))
}
