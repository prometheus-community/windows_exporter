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

package log

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"golang.org/x/sys/windows"
)

// openWithoutDeleteSharing opens path like a log shipper that doesn't allow
// delete sharing, which blocks renaming the file.
func openWithoutDeleteSharing(t *testing.T, path string) windows.Handle {
	t.Helper()

	pathPtr, err := windows.UTF16PtrFromString(path)
	require.NoError(t, err)

	handle, err := windows.CreateFile(pathPtr, windows.GENERIC_READ, windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE, nil, windows.OPEN_EXISTING, windows.FILE_ATTRIBUTE_NORMAL, 0)
	require.NoError(t, err)

	return handle
}

func TestRotatingFileBlockedByOtherProcess(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "windows_exporter.log")
	now := time.Date(2026, 10, 15, 14, 3, 5, 0, time.UTC)

	file, err := newRotatingFile(path, 10, 0, false, func() time.Time { return now })
	require.NoError(t, err)

	t.Cleanup(func() { _ = file.Close() })

	write(t, file, "1234567890")

	blocker := openWithoutDeleteSharing(t, path)

	// The rename fails while the file is held: the line is appended instead.
	write(t, file, "while blocked\n")
	require.Equal(t, []string{"windows_exporter.log"}, dirEntries(t, filepath.Dir(path)))

	require.NoError(t, windows.CloseHandle(blocker))

	now = now.Add(file.retryInterval)
	write(t, file, "after release\n")
	require.NoError(t, file.Close())

	require.Equal(t, "1234567890while blocked\n", readFile(t, filepath.Join(filepath.Dir(path), "windows_exporter-2026-10-15T14-04-05.000.log")))
	require.Equal(t, "after release\n", readFile(t, path))
}

func TestOpenLogFileAllowsRenameAndDelete(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "windows_exporter.log")

	file, err := openLogFile(path)
	require.NoError(t, err)

	t.Cleanup(func() { _ = file.Close() })

	_, err = file.WriteString("first\n")
	require.NoError(t, err)

	// An admin can move the file while the exporter writes to it.
	moved := filepath.Join(dir, "moved.log")
	require.NoError(t, os.Rename(path, moved))

	_, err = file.WriteString("second\n")
	require.NoError(t, err)
	require.Equal(t, "first\nsecond\n", readFile(t, moved))

	// And delete it.
	require.NoError(t, os.Remove(moved))
}

func TestOpenLogFileAppends(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "windows_exporter.log")
	require.NoError(t, os.WriteFile(path, []byte("existing\n"), 0o600))

	file, err := openLogFile(path)
	require.NoError(t, err)

	_, err = file.WriteString("appended\n")
	require.NoError(t, err)
	require.NoError(t, file.Close())

	require.Equal(t, "existing\nappended\n", readFile(t, path))
}
