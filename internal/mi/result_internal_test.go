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

package mi

import (
	"testing"
	"unsafe"

	"github.com/stretchr/testify/require"
)

func TestResultError_String(t *testing.T) {
	t.Parallel()

	require.Equal(t, "MI_RESULT_OK", MI_RESULT_OK.String())
	require.Equal(t, "MI_RESULT_NO_SUCH_PROPERTY", MI_RESULT_NO_SUCH_PROPERTY.Error())
	require.Equal(t, "MI_RESULT_SERVER_IS_SHUTTING_DOWN", MI_RESULT_SERVER_IS_SHUTTING_DOWN.String())
	require.Equal(t, "MI_RESULT_UNKNOWN(1000)", ResultError(1000).String())
}

// TestValueType_Size guards the MI_Type out-parameter of MI_Instance_GetElement,
// which the native side writes as a 32-bit enum.
func TestValueType_Size(t *testing.T) {
	t.Parallel()

	require.Equal(t, uintptr(4), unsafe.Sizeof(ValueType(0)))
}

func TestElement_GetValue_Null(t *testing.T) {
	t.Parallel()

	for _, valueType := range []ValueType{ValueTypeUINT32, ValueTypeSTRING, ValueTypeDATETIME} {
		value, err := newElement(valueType, [5]uint64{42}, flagNull).GetValue()
		require.NoError(t, err)
		require.Nil(t, value)
	}
}
