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

package kernel32_test

import (
	"os/exec"
	"testing"

	"github.com/prometheus-community/windows_exporter/internal/headers/kernel32"
	"github.com/stretchr/testify/require"
	"golang.org/x/sys/windows"
)

// startJobProcesses starts count idle processes and assigns them to a new job object.
func startJobProcesses(t *testing.T, count int) (windows.Handle, []uint32) {
	t.Helper()

	job, err := windows.CreateJobObject(nil, nil)
	require.NoError(t, err)

	t.Cleanup(func() {
		_ = windows.TerminateJobObject(job, 0)
		_ = windows.CloseHandle(job)
	})

	pids := make([]uint32, 0, count)

	for range count {
		// cmd.exe blocks on the open stdin pipe until the job is terminated.
		cmd := exec.CommandContext(t.Context(), "cmd.exe", "/c", "pause")

		stdin, err := cmd.StdinPipe()
		require.NoError(t, err)
		require.NoError(t, cmd.Start())

		t.Cleanup(func() {
			_ = stdin.Close()
			_ = cmd.Process.Kill()
			_ = cmd.Wait()
		})

		process, err := windows.OpenProcess(windows.PROCESS_SET_QUOTA|windows.PROCESS_TERMINATE, false, uint32(cmd.Process.Pid))
		require.NoError(t, err)

		err = windows.AssignProcessToJobObject(job, process)
		_ = windows.CloseHandle(process)

		require.NoError(t, err)

		pids = append(pids, uint32(cmd.Process.Pid))
	}

	return job, pids
}

// More processes than the initial buffer holds must trigger ERROR_MORE_DATA and a retry.
func TestQueryJobObjectProcessIDs(t *testing.T) {
	t.Parallel()

	job, pids := startJobProcesses(t, 20)

	got, err := kernel32.QueryJobObjectProcessIDs(job)
	require.NoError(t, err)
	require.ElementsMatch(t, pids, got)
}

// IsProcessInJob wrote the BOOL into the pointer variable instead of the result,
// so it always reported false.
func TestIsProcessInJob(t *testing.T) {
	t.Parallel()

	job, pids := startJobProcesses(t, 1)

	process, err := windows.OpenProcess(windows.PROCESS_QUERY_LIMITED_INFORMATION, false, pids[0])
	require.NoError(t, err)

	t.Cleanup(func() {
		_ = windows.CloseHandle(process)
	})

	isInJob, err := kernel32.IsProcessInJob(process, job)
	require.NoError(t, err)
	require.True(t, isInJob)

	isInJob, err = kernel32.IsProcessInJob(windows.CurrentProcess(), job)
	require.NoError(t, err)
	require.False(t, isInJob)
}
