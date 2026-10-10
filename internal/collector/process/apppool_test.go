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

package process

import (
	"fmt"
	"log/slog"
	"os"
	"regexp"
	"strings"
	"testing"
	"unicode/utf8"
	"unsafe"

	"github.com/prometheus-community/windows_exporter/internal/mi"
	"github.com/prometheus-community/windows_exporter/internal/types"
	"github.com/stretchr/testify/require"
	"golang.org/x/sys/windows"
)

// A command line as WAS starts IIS 10 worker processes.
const wasCommandLine = `c:\windows\system32\inetsrv\w3wp.exe -ap "DefaultAppPool" -v "v4.0" -l "webengine4.dll" -a \\.\pipe\iisipm5f5e2b1c-8d6a-4b2e-9a0c-3d1b2f4e6a7c -h "C:\inetpub\temp\apppools\DefaultAppPool\DefaultAppPool.config" -w "" -m 0 -t 20 -ta 0`

//nolint:gochecknoglobals
var commandLineTests = []struct {
	name    string
	cmdLine string
	args    []string
}{
	{"empty", ``, nil},
	{"program only", `w3wp.exe`, []string{`w3wp.exe`}},
	{"leading space", `  a b`, []string{``, `a`, `b`}},
	{"whitespace only", "  \t ", []string{``}},
	{"trailing space", `a b  `, []string{`a`, `b`}},
	{"tabs", "a\tb\t\tc", []string{`a`, `b`, `c`}},
	{"quoted program", `"C:\Program Files\w3wp.exe" -ap x`, []string{`C:\Program Files\w3wp.exe`, `-ap`, `x`}},
	{"quoted program backslash", `"C:\dir\" x`, []string{`C:\dir\`, `x`}},
	{"quoted program unterminated", `"C:\dir x`, []string{`C:\dir x`}},
	{"quoted program followed by text", `"a"b c`, []string{`a`, `b`, `c`}},
	{"program backslash quote", `a\"b c`, []string{`a\"b`, `c`}},
	{"empty quoted argument", `a "" b`, []string{`a`, ``, `b`}},
	{"trailing empty quoted argument", `a ""`, []string{`a`, ``}},
	{"quoted spaces", `a "b c" d`, []string{`a`, `b c`, `d`}},
	{"embedded quotes", `a b"c d"e`, []string{`a`, `bc de`}},
	{"unterminated quote", `a "b c`, []string{`a`, `b c`}},
	{"escaped quote", `a b\"c`, []string{`a`, `b"c`}},
	{"escaped backslash before quote", `a "b\\" c`, []string{`a`, `b\`, `c`}},
	{"three backslashes before quote", `a b\\\"c`, []string{`a`, `b\"c`}},
	{"backslashes without quote", `a \\server\share\ b`, []string{`a`, `\\server\share\`, `b`}},
	{"double quote in quoted part", `a "b""c" d`, []string{`a`, `b"c d`}},
	{"three quotes in quoted part", `a "b"""c d`, []string{`a`, `b"c d`}},
	{"control character ends program", "a\vb c", []string{`a`, `b`, `c`}},
	{"control character after program", "a \vb", []string{`a`, "\vb"}},
	{"control character after quoted program", "\"a\"\vb", []string{`a`, "\vb"}},
	{"control character in argument", "a b\vc\nd", []string{`a`, "b\vc\nd"}},
	{"unicode spaces", "a b c　d", []string{"a b", "c　d"}},
	{"three quotes", `a """b c`, []string{`a`, `"b`, `c`}},
	{"unicode", `a "Café Пул ✓ Ωμέγα" b`, []string{`a`, `Café Пул ✓ Ωμέγα`, `b`}},
	{"was command line", wasCommandLine, []string{
		`c:\windows\system32\inetsrv\w3wp.exe`, `-ap`, `DefaultAppPool`, `-v`, `v4.0`, `-l`, `webengine4.dll`,
		`-a`, `\\.\pipe\iisipm5f5e2b1c-8d6a-4b2e-9a0c-3d1b2f4e6a7c`, `-h`, `C:\inetpub\temp\apppools\DefaultAppPool\DefaultAppPool.config`,
		`-w`, ``, `-m`, `0`, `-t`, `20`, `-ta`, `0`,
	}},
}

func TestSplitCommandLine(t *testing.T) {
	t.Parallel()

	for _, tc := range commandLineTests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			require.Equal(t, tc.args, splitCommandLine(tc.cmdLine))
		})
	}
}

// splitCommandLine must split like CommandLineToArgvW, which is the reference implementation.
func TestSplitCommandLineMatchesCommandLineToArgvW(t *testing.T) {
	t.Parallel()

	for _, tc := range commandLineTests {
		if tc.cmdLine == "" {
			// CommandLineToArgvW returns the path of the current executable for an empty command line.
			continue
		}

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			requireCommandLineToArgvW(t, tc.cmdLine)
		})
	}
}

func TestAppPoolFromCommandLine(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name    string
		cmdLine string
		pool    string
		ok      bool
	}{
		{"was command line", wasCommandLine, "DefaultAppPool", true},
		{"spaces", `w3wp.exe -ap "Classic .NET AppPool" -v "v2.0"`, "Classic .NET AppPool", true},
		{"unicode", `w3wp.exe -ap "Café Пул ✓ Ωμέγα 🚀" -v "v4.0"`, "Café Пул ✓ Ωμέγα 🚀", true},
		{"unquoted", `w3wp.exe -ap DefaultAppPool`, "DefaultAppPool", true},
		{"tab separated", "w3wp.exe\t-ap\t\"Pool\"", "Pool", true},
		{"upper case flag", `w3wp.exe -AP "Pool"`, "Pool", true},
		{"not first argument", `w3wp.exe -debug -ap "Pool" -v "v4.0"`, "Pool", true},
		{"first of duplicates", `w3wp.exe -ap "First" -ap "Second"`, "First", true},
		{"quoted program with spaces", `"C:\Program Files\w3wp.exe" -ap "Pool"`, "Pool", true},
		{"embedded quotes", `w3wp.exe -ap Default"App Pool"`, "DefaultApp Pool", true},
		{"escaped quote", `w3wp.exe -ap "a\"b"`, `a"b`, true},
		{"trailing backslash", `w3wp.exe -ap "a\\"`, `a\`, true},
		{"literal backslash", `w3wp.exe -ap "a\b"`, `a\b`, true},
		{"missing", `w3wp.exe -v "v4.0"`, "", false},
		{"no arguments", `c:\windows\system32\inetsrv\w3wp.exe`, "", false},
		{"empty command line", ``, "", false},
		{"missing value", `w3wp.exe -v "v4.0" -ap`, "", false},
		{"empty value", `w3wp.exe -ap "" -v "v4.0"`, "", false},
		{"flag inside quotes", `w3wp.exe "-ap Pool"`, "", false},
		{"flag prefix", `w3wp.exe -apx "Pool"`, "", false},
		{"slash flag", `w3wp.exe /ap "Pool"`, "", false},
		{"program named -ap", `-ap Pool`, "", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			pool, ok := appPoolFromCommandLine(tc.cmdLine)
			require.Equal(t, tc.ok, ok)
			require.Equal(t, tc.pool, pool)
		})
	}
}

func FuzzSplitCommandLine(f *testing.F) {
	for _, tc := range commandLineTests {
		f.Add(tc.cmdLine)
	}

	f.Add(`w3wp.exe -ap "a\\\"b" -ap`)
	f.Add(`"" "" ""`)
	f.Add(`a """"""" b`)
	f.Add("a\v\r\nb\u00a0c\u3000d")

	f.Fuzz(func(t *testing.T, cmdLine string) {
		// CommandLineToArgvW takes NUL-terminated UTF-16 and returns the executable path for an empty command line.
		if cmdLine == "" || strings.ContainsRune(cmdLine, 0) || !utf8.ValidString(cmdLine) {
			t.Skip()
		}

		args := requireCommandLineToArgvW(t, cmdLine)

		var (
			want   string
			wantOK bool
		)

		for i := 1; i < len(args)-1; i++ {
			if strings.EqualFold(args[i], "-ap") {
				want, wantOK = args[i+1], args[i+1] != ""

				break
			}
		}

		pool, ok := appPoolFromCommandLine(cmdLine)
		require.Equal(t, wantOK, ok)
		require.Equal(t, want, pool)
	})
}

func requireCommandLineToArgvW(t *testing.T, cmdLine string) []string {
	t.Helper()

	want, err := windows.DecomposeCommandLine(cmdLine)
	require.NoError(t, err)

	got := splitCommandLine(cmdLine)
	if len(want) == 0 {
		require.Empty(t, got, "command line %q", cmdLine)
	} else {
		require.Equal(t, want, got, "command line %q", cmdLine)
	}

	return got
}

// The command line is read from a process that never ran, with an argument
// list as WAS passes it, so no IIS installation is needed.
func TestWorkerProcessAppPool(t *testing.T) {
	t.Parallel()

	executable, err := os.Executable()
	require.NoError(t, err)

	const pool = "Café Пул ✓ Ωμέγα 🚀"

	cmdLine := `"` + executable + `" -ap "` + pool + `" -v "v4.0" -a \\.\pipe\iisipm -w "" -m 0`

	appName, err := windows.UTF16PtrFromString(executable)
	require.NoError(t, err)

	cmdLineUTF16, err := windows.UTF16FromString(cmdLine)
	require.NoError(t, err)

	startupInfo := windows.StartupInfo{Cb: uint32(unsafe.Sizeof(windows.StartupInfo{}))}

	var processInfo windows.ProcessInformation

	// CREATE_SUSPENDED: the process parameters exist, but the test binary never runs.
	require.NoError(t, windows.CreateProcess(appName, &cmdLineUTF16[0], nil, nil, false,
		windows.CREATE_SUSPENDED|windows.CREATE_NO_WINDOW, nil, nil, &startupInfo, &processInfo))

	t.Cleanup(func() {
		require.NoError(t, windows.TerminateProcess(processInfo.Process, 1))
		require.NoError(t, windows.CloseHandle(processInfo.Thread))
		require.NoError(t, windows.CloseHandle(processInfo.Process))
	})

	pid := processInfo.ProcessId

	got, err := queryProcessCommandLine(pid)
	require.NoError(t, err)
	require.Equal(t, cmdLine, got)

	gotPool, err := workerProcessAppPool(pid)
	require.NoError(t, err)
	require.Equal(t, pool, gotPool)
}

func TestWorkerProcessAppPoolWithoutArgument(t *testing.T) {
	t.Parallel()

	_, err := workerProcessAppPool(windows.GetCurrentProcessId())
	require.ErrorIs(t, err, errNoAppPoolArgument)
}

func TestQueryProcessCommandLine(t *testing.T) {
	t.Parallel()

	got, err := queryProcessCommandLine(windows.GetCurrentProcessId())
	require.NoError(t, err)
	require.Equal(t, windows.UTF16PtrToString(windows.GetCommandLine()), got)
}

func TestResolveAppPools(t *testing.T) {
	t.Parallel()

	logger := slog.New(slog.DiscardHandler)

	commandLines := func(results map[uint32]error) func(uint32) (string, error) {
		return func(pid uint32) (string, error) {
			if err := results[pid]; err != nil {
				return "", err
			}

			return fmt.Sprintf("native-%d", pid), nil
		}
	}

	wmiResult := []WorkerProcess{
		{ProcessId: 4, AppPoolName: "wmi-4"},
		{ProcessId: 8, AppPoolName: "wmi-8"},
		{ProcessId: 12, AppPoolName: ""},
	}

	t.Run("command line only", func(t *testing.T) {
		t.Parallel()

		pools, err := resolveAppPools(logger, []uint32{4, 8}, commandLines(nil), func() ([]WorkerProcess, error) {
			t.Fatal("WMI must not be queried when every command line has the pool")

			return nil, nil
		})
		require.NoError(t, err)
		require.Equal(t, map[uint32]string{4: "native-4", 8: "native-8"}, pools)
	})

	t.Run("WMI fallback", func(t *testing.T) {
		t.Parallel()

		calls := 0
		pools, err := resolveAppPools(logger, []uint32{4, 8, 12, 16}, commandLines(map[uint32]error{
			8:  windows.ERROR_ACCESS_DENIED,
			12: errNoAppPoolArgument,
			16: windows.STATUS_INVALID_INFO_CLASS,
		}), func() ([]WorkerProcess, error) {
			calls++

			return wmiResult, nil
		})
		require.NoError(t, err)
		require.Equal(t, 1, calls)
		// An empty WMI pool name and a process WMI doesn't know get no label.
		require.Equal(t, map[uint32]string{4: "native-4", 8: "wmi-8"}, pools)
	})

	t.Run("exited process", func(t *testing.T) {
		t.Parallel()

		pools, err := resolveAppPools(logger, []uint32{4, 8}, commandLines(map[uint32]error{
			8: fmt.Errorf("OpenProcess: %w", windows.ERROR_INVALID_PARAMETER),
		}), func() ([]WorkerProcess, error) {
			t.Fatal("WMI must not be queried for exited processes")

			return nil, nil
		})
		require.NoError(t, err)
		require.Equal(t, map[uint32]string{4: "native-4"}, pools)
	})

	t.Run("without WMI", func(t *testing.T) {
		t.Parallel()

		pools, err := resolveAppPools(logger, []uint32{4, 8}, commandLines(map[uint32]error{
			8: windows.ERROR_ACCESS_DENIED,
		}), nil)
		require.NoError(t, err)
		require.Equal(t, map[uint32]string{4: "native-4"}, pools)
	})

	t.Run("WMI error", func(t *testing.T) {
		t.Parallel()

		wmiErr := windows.ERROR_TIMEOUT
		pools, err := resolveAppPools(logger, []uint32{4, 8}, commandLines(map[uint32]error{
			8: windows.ERROR_ACCESS_DENIED,
		}), func() ([]WorkerProcess, error) {
			return nil, wmiErr
		})
		require.ErrorIs(t, err, wmiErr)
		require.Equal(t, map[uint32]string{4: "native-4"}, pools)
	})
}

// Grafana Alloy builds and closes collectors repeatedly. Without
// root\WebAdministration, Build must succeed and use the command line only.
func TestBuildCloseWorkerProcess(t *testing.T) {
	t.Parallel()

	miApp, err := mi.ApplicationInitialize()
	require.NoError(t, err)

	t.Cleanup(func() { require.NoError(t, miApp.Close()) })

	miSession, err := miApp.NewSession(nil)
	require.NoError(t, err)

	t.Cleanup(func() { require.NoError(t, miSession.Close()) })

	c := New(&Config{
		ProcessInclude:      regexp.MustCompile(`^(?:w3wp)$`),
		ProcessExclude:      types.RegExpEmpty,
		EnableWorkerProcess: true,
	})

	logger := slog.New(slog.DiscardHandler)

	for _, session := range []*mi.Session{nil, miSession, nil} {
		require.NoError(t, c.Build(logger, session))

		if session == nil {
			require.Nil(t, c.miSession)
		}

		require.NoError(t, c.Close())
		require.NoError(t, c.Close())
		require.Nil(t, c.miSession)
	}
}

// The command line source costs one OpenProcess and one NtQueryInformationProcess per worker process.
func BenchmarkWorkerProcessAppPool(b *testing.B) {
	pid := windows.GetCurrentProcessId()

	for b.Loop() {
		cmdLine, err := queryProcessCommandLine(pid)
		require.NoError(b, err)

		appPoolFromCommandLine(cmdLine)
	}
}

func BenchmarkAppPoolFromCommandLine(b *testing.B) {
	for b.Loop() {
		appPoolFromCommandLine(wasCommandLine)
	}
}

// Errors must be the native Windows error, so callers can tell exited processes from denied access.
func TestQueryProcessCommandLineMissingProcess(t *testing.T) {
	t.Parallel()

	// Process IDs are multiples of four and far below this value.
	_, err := queryProcessCommandLine(0xFFFFFFFC)
	require.ErrorIs(t, err, windows.ERROR_INVALID_PARAMETER)
}
