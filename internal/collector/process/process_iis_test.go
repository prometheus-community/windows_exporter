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

package process_test

import (
	"context"
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"testing"
	"time"
	"unsafe"

	"github.com/prometheus-community/windows_exporter/internal/collector/process"
	"github.com/prometheus-community/windows_exporter/internal/mi"
	"github.com/prometheus-community/windows_exporter/internal/types"
	"github.com/prometheus-community/windows_exporter/internal/utils/testutils"
	"github.com/stretchr/testify/require"
	"golang.org/x/sys/windows"
)

// appPoolIdentityPrefix is the domain of ApplicationPoolIdentity virtual accounts,
// whose names are the application pool names.
const appPoolIdentityPrefix = `IIS APPPOOL\`

// The application pool read from each w3wp command line must match
// WorkerProcess.AppPoolName and the ApplicationPoolIdentity of the process.
func TestWorkerProcessAppPoolIIS(t *testing.T) {
	pids := requireWorkerProcesses(t)

	appPools := make(map[uint32]string, len(pids))

	for _, pid := range pids {
		appPool, err := process.WorkerProcessAppPool(pid)
		if errors.Is(err, windows.ERROR_INVALID_PARAMETER) {
			// The worker process exited, for example after its idle timeout.
			continue
		}

		if errors.Is(err, windows.ERROR_ACCESS_DENIED) && !iisRequired() {
			t.Skipf("worker process %d: %v; run the test elevated", pid, err)
		}

		require.NoError(t, err, "worker process %d", pid)
		require.NotEmpty(t, appPool)

		appPools[pid] = appPool

		if owner := processOwner(t, pid); strings.HasPrefix(owner, appPoolIdentityPrefix) {
			require.Equal(t, strings.TrimPrefix(owner, appPoolIdentityPrefix), appPool, "worker process %d", pid)
		}
	}

	workerProcesses, err := queryWorkerProcesses(t)
	if err != nil {
		t.Logf(`root\WebAdministration isn't available, skipping the WMI comparison: %v`, err)

		return
	}

	for _, wp := range workerProcesses {
		if appPool, ok := appPools[uint32(wp.ProcessId)]; ok {
			require.Equal(t, wp.AppPoolName, appPool, "worker process %d", wp.ProcessId)
		}
	}
}

// The collector must publish the application pool in the process label.
func TestCollectorIISAppPool(t *testing.T) {
	requireWorkerProcesses(t)

	families := testutils.TestCollector(t, process.New, &process.Config{
		ProcessInclude:      regexp.MustCompile(`^(?:w3wp)$`),
		ProcessExclude:      types.RegExpEmpty,
		EnableWorkerProcess: true,
		CounterVersion:      1,
	})

	require.Contains(t, families, "windows_process_info")

	var labeled int

	for _, metric := range families["windows_process_info"].GetMetric() {
		labels := make(map[string]string)
		for _, label := range metric.GetLabel() {
			labels[label.GetName()] = label.GetValue()
		}

		owner := labels["owner"]
		if !strings.HasPrefix(owner, appPoolIdentityPrefix) {
			continue
		}

		require.Equal(t, "w3wp_"+strings.TrimPrefix(owner, appPoolIdentityPrefix), labels["process"], "process_id %s", labels["process_id"])

		labeled++
	}

	if iisRequired() {
		require.Positive(t, labeled, "no worker process runs as an ApplicationPoolIdentity")
	}
}

// A suspended copy of the test binary named w3wp.exe, started with a WAS
// command line, must get the pool in its process label without IIS.
func TestCollectorWorkerProcessCommandLine(t *testing.T) {
	executable, err := os.Executable()
	require.NoError(t, err)

	content, err := os.ReadFile(executable)
	require.NoError(t, err)

	workerProcess := filepath.Join(t.TempDir(), "w3wp.exe")
	//nolint:gosec // The destination is a fixed filename in a test-owned temporary directory.
	require.NoError(t, os.WriteFile(workerProcess, content, 0o600))

	const appPool = "Test Pool Ωμέγα 🚀"

	pid := startSuspended(t, workerProcess, `"`+workerProcess+`" -ap "`+appPool+`" -v "v4.0" -w "" -m 0`)

	families := testutils.TestCollector(t, process.New, &process.Config{
		ProcessInclude:      regexp.MustCompile(`^(?:w3wp)$`),
		ProcessExclude:      types.RegExpEmpty,
		EnableWorkerProcess: true,
		CounterVersion:      1,
	})

	// The process snapshot includes suspended processes.
	require.Contains(t, families, "windows_process_info", "no w3wp process in the process snapshot")

	for _, metric := range families["windows_process_info"].GetMetric() {
		labels := make(map[string]string)
		for _, label := range metric.GetLabel() {
			labels[label.GetName()] = label.GetValue()
		}

		if labels["process_id"] == strconv.FormatUint(uint64(pid), 10) {
			require.Equal(t, "w3wp_"+appPool, labels["process"])

			return
		}
	}

	t.Fatalf("windows_process_info has no series for process %d", pid)
}

// startSuspended starts a process that never runs and returns its ID.
func startSuspended(t *testing.T, path, cmdLine string) uint32 {
	t.Helper()

	appName, err := windows.UTF16PtrFromString(path)
	require.NoError(t, err)

	cmdLineUTF16, err := windows.UTF16FromString(cmdLine)
	require.NoError(t, err)

	startupInfo := windows.StartupInfo{Cb: uint32(unsafe.Sizeof(windows.StartupInfo{}))}

	var processInfo windows.ProcessInformation

	require.NoError(t, windows.CreateProcess(appName, &cmdLineUTF16[0], nil, nil, false,
		windows.CREATE_SUSPENDED|windows.CREATE_NO_WINDOW, nil, nil, &startupInfo, &processInfo))

	t.Cleanup(func() {
		require.NoError(t, windows.TerminateProcess(processInfo.Process, 1))
		_, _ = windows.WaitForSingleObject(processInfo.Process, windows.INFINITE)
		require.NoError(t, windows.CloseHandle(processInfo.Thread))
		require.NoError(t, windows.CloseHandle(processInfo.Process))
	})

	return processInfo.ProcessId
}

// iisRequired reports whether CI provisioned IIS, so a missing worker process is a failure.
func iisRequired() bool {
	return slices.Contains(strings.Split(os.Getenv("WINDOWS_EXPORTER_TEST_COLLECTORS"), ","), "iis")
}

// requireWorkerProcesses returns the IIS worker processes. If W3SVC runs
// without one, it requests the local default site to start one.
func requireWorkerProcesses(t *testing.T) []uint32 {
	t.Helper()

	if pids := workerProcessIDs(t); len(pids) > 0 {
		return pids
	}

	if !serviceRunning(t, "W3SVC") {
		if iisRequired() {
			t.Fatal("W3SVC isn't running")
		}

		t.Skip("IIS isn't running")
	}

	ctx, cancel := context.WithTimeout(t.Context(), 30*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://127.0.0.1/", nil)
	require.NoError(t, err)

	resp, err := http.DefaultClient.Do(req)
	if err == nil {
		_ = resp.Body.Close()
	} else {
		t.Logf("request to start a worker process failed: %v", err)
	}

	pids := workerProcessIDs(t)
	if len(pids) == 0 {
		if iisRequired() {
			t.Fatal("no IIS worker process is running")
		}

		t.Skip("no IIS worker process is running")
	}

	return pids
}

func workerProcessIDs(t *testing.T) []uint32 {
	t.Helper()

	snapshot, err := windows.CreateToolhelp32Snapshot(windows.TH32CS_SNAPPROCESS, 0)
	require.NoError(t, err)

	defer func() { _ = windows.CloseHandle(snapshot) }()

	var (
		pids  []uint32
		entry = windows.ProcessEntry32{Size: uint32(unsafe.Sizeof(windows.ProcessEntry32{}))}
	)

	for err = windows.Process32First(snapshot, &entry); err == nil; err = windows.Process32Next(snapshot, &entry) {
		if strings.EqualFold(windows.UTF16ToString(entry.ExeFile[:]), "w3wp.exe") {
			pids = append(pids, entry.ProcessID)
		}
	}

	require.ErrorIs(t, err, windows.ERROR_NO_MORE_FILES)

	return pids
}

func serviceRunning(t *testing.T, name string) bool {
	t.Helper()

	scm, err := windows.OpenSCManager(nil, nil, windows.SC_MANAGER_CONNECT)
	require.NoError(t, err)

	defer func() { _ = windows.CloseServiceHandle(scm) }()

	serviceName, err := windows.UTF16PtrFromString(name)
	require.NoError(t, err)

	h, err := windows.OpenService(scm, serviceName, windows.SERVICE_QUERY_STATUS)
	if errors.Is(err, windows.ERROR_SERVICE_DOES_NOT_EXIST) {
		return false
	}

	require.NoError(t, err)

	defer func() { _ = windows.CloseServiceHandle(h) }()

	var status windows.SERVICE_STATUS

	require.NoError(t, windows.QueryServiceStatus(h, &status))

	return status.CurrentState == windows.SERVICE_RUNNING
}

// processOwner returns the DOMAIN\user owner of pid, or "" if it can't be read.
func processOwner(t *testing.T, pid uint32) string {
	t.Helper()

	hProcess, err := windows.OpenProcess(windows.PROCESS_QUERY_LIMITED_INFORMATION, false, pid)
	if err != nil {
		return ""
	}

	defer func() { _ = windows.CloseHandle(hProcess) }()

	var token windows.Token
	if err := windows.OpenProcessToken(hProcess, windows.TOKEN_QUERY, &token); err != nil {
		return ""
	}

	defer func() { _ = token.Close() }()

	user, err := token.GetTokenUser()
	require.NoError(t, err)

	account, domain, _, err := user.User.Sid.LookupAccount("")
	if err != nil {
		t.Logf("worker process %d: LookupAccountSid: %v", pid, err)

		return ""
	}

	return domain + `\` + account
}

func queryWorkerProcesses(t *testing.T) ([]process.WorkerProcess, error) {
	t.Helper()

	miApp, err := mi.ApplicationInitialize()
	require.NoError(t, err)

	t.Cleanup(func() { require.NoError(t, miApp.Close()) })

	miSession, err := miApp.NewSession(nil)
	require.NoError(t, err)

	t.Cleanup(func() { require.NoError(t, miSession.Close()) })

	query, err := mi.NewQuery("SELECT AppPoolName, ProcessId FROM WorkerProcess")
	require.NoError(t, err)

	var workerProcesses []process.WorkerProcess

	if err := miSession.Query(&workerProcesses, mi.NamespaceRootWebAdministration, query, 0); err != nil {
		return nil, err
	}

	for _, wp := range workerProcesses {
		t.Logf("WorkerProcess %s: %q", strconv.FormatUint(wp.ProcessId, 10), wp.AppPoolName)
	}

	return workerProcesses, nil
}
