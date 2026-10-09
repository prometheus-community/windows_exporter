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

package main

import (
	"math"
	"runtime/debug"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestProcessMemoryLimit(t *testing.T) {
	// This setting is process-wide, so the cases must run sequentially.
	previous := debug.SetMemoryLimit(-1)

	t.Cleanup(func() { debug.SetMemoryLimit(previous) })

	require.NoError(t, setProcessMemoryLimit(0))
	require.Equal(t, int64(math.MaxInt64), debug.SetMemoryLimit(-1))
	require.NoError(t, setProcessMemoryLimit(1<<30))
	require.Equal(t, int64(1<<30), debug.SetMemoryLimit(-1))
	require.ErrorContains(t, setProcessMemoryLimit(-1), "non-negative")
	require.Equal(t, int64(1<<30), debug.SetMemoryLimit(-1))
}
