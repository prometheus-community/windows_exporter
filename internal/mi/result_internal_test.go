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

	"github.com/stretchr/testify/require"
)

// TestResultError_Values pins the constants to the native MI_Result values,
// which have no 18 and 19.
func TestResultError_Values(t *testing.T) {
	t.Parallel()

	require.EqualValues(t, 17, MI_RESULT_METHOD_NOT_FOUND)
	require.EqualValues(t, 20, MI_RESULT_NAMESPACE_NOT_EMPTY)
	require.EqualValues(t, 22, MI_RESULT_INVALID_OPERATION_TIMEOUT)
	require.EqualValues(t, 28, MI_RESULT_SERVER_IS_SHUTTING_DOWN)
}

func TestResultError_String(t *testing.T) {
	t.Parallel()

	require.Equal(t, "MI_RESULT_OK", MI_RESULT_OK.String())
	require.Equal(t, "MI_RESULT_NO_SUCH_PROPERTY", MI_RESULT_NO_SUCH_PROPERTY.Error())
	require.Equal(t, "MI_RESULT_INVALID_OPERATION_TIMEOUT", ResultError(22).String())
	require.Equal(t, "MI_RESULT_SERVER_IS_SHUTTING_DOWN", ResultError(28).String())
	require.Equal(t, "MI_RESULT_UNKNOWN(18)", ResultError(18).String())
	require.Equal(t, "MI_RESULT_UNKNOWN(1000)", ResultError(1000).String())
}
