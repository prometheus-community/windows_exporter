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

package log

import (
	"compress/gzip"
	"errors"
	"io"
	"math"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// testRotatingFile returns a rotating file with a small size limit and a
// clock that only moves when the test moves it.
func testRotatingFile(t *testing.T, maxSize int64, maxBackups int, compress bool) (*rotatingFile, string, *time.Time) {
	t.Helper()

	path := filepath.Join(t.TempDir(), "windows_exporter.log")
	now := time.Date(2026, 10, 15, 14, 3, 5, 0, time.UTC)

	file, err := newRotatingFile(path, maxSize, maxBackups, compress, func() time.Time { return now })
	require.NoError(t, err)

	t.Cleanup(func() { _ = file.Close() })

	return file, path, &now
}

func write(t *testing.T, w io.Writer, line string) {
	t.Helper()

	n, err := w.Write([]byte(line))
	require.NoError(t, err)
	require.Equal(t, len(line), n)
}

func dirEntries(t *testing.T, dir string) []string {
	t.Helper()

	entries, err := os.ReadDir(dir)
	require.NoError(t, err)

	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		names = append(names, entry.Name())
	}

	slices.Sort(names)

	return names
}

func readFile(t *testing.T, path string) string {
	t.Helper()

	data, err := os.ReadFile(path)
	require.NoError(t, err)

	return string(data)
}

func TestRotatingFileRotatesBeforeExceedingMaxSize(t *testing.T) {
	t.Parallel()

	file, path, now := testRotatingFile(t, 10, 0, false)

	write(t, file, "12345\n")
	write(t, file, "1234\n") // 11 bytes would exceed 10: rotate first

	*now = now.Add(time.Second)

	write(t, file, "abcdefghijklmn\n") // longer than max-size on its own: rotate, then write it whole

	require.NoError(t, file.Close())

	require.Equal(t, []string{
		"windows_exporter-2026-10-15T14-03-05.000.log",
		"windows_exporter-2026-10-15T14-03-06.000.log",
		"windows_exporter.log",
	}, dirEntries(t, filepath.Dir(path)))

	require.Equal(t, "12345\n", readFile(t, filepath.Join(filepath.Dir(path), "windows_exporter-2026-10-15T14-03-05.000.log")))
	require.Equal(t, "1234\n", readFile(t, filepath.Join(filepath.Dir(path), "windows_exporter-2026-10-15T14-03-06.000.log")))
	require.Equal(t, "abcdefghijklmn\n", readFile(t, path))
}

func TestRotatingFileKeepsWritingWhenRenameFails(t *testing.T) {
	t.Parallel()

	file, path, now := testRotatingFile(t, 10, 0, false)

	errBlocked := errors.New("file is in use")
	renames := 0
	file.rename = func(oldPath, newPath string) error {
		renames++

		if renames == 1 {
			return errBlocked
		}

		return os.Rename(oldPath, newPath)
	}

	write(t, file, "1234567890")
	write(t, file, "after a failed rotation\n") // the rename fails: the line goes to the current file
	require.Equal(t, 1, renames)

	write(t, file, "still no retry\n") // within the retry interval: no new attempt
	require.Equal(t, 1, renames)

	*now = now.Add(file.retryInterval)
	write(t, file, "rotated\n") // the retry succeeds
	require.Equal(t, 2, renames)

	require.NoError(t, file.Close())

	backup := filepath.Join(filepath.Dir(path), "windows_exporter-2026-10-15T14-04-05.000.log")
	require.Equal(t, "1234567890after a failed rotation\nstill no retry\n", readFile(t, backup), "no line may be lost")
	require.Equal(t, "rotated\n", readFile(t, path))
}

func TestRotatingFileStartsWhenInitialRotationFails(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "windows_exporter.log")
	require.NoError(t, os.WriteFile(path, []byte("an old log that is already too large\n"), 0o600))

	now := time.Date(2026, 10, 15, 14, 3, 5, 0, time.UTC)

	// Rotating the oversized file on startup works.
	file, err := newRotatingFile(path, 10, 0, false, func() time.Time { return now })
	require.NoError(t, err)
	write(t, file, "new\n")
	require.NoError(t, file.Close())
	require.Equal(t, "new\n", readFile(t, path))

	// If it fails, the logger still starts and appends.
	require.NoError(t, os.WriteFile(path, []byte("an old log that is already too large\n"), 0o600))

	file, err = newRotatingFileWithRename(path, 10, 0, false, func() time.Time { return now }, func(string, string) error {
		return errors.New("file is in use")
	})
	require.NoError(t, err)
	write(t, file, "appended\n")
	require.NoError(t, file.Close())
	require.Equal(t, "an old log that is already too large\nappended\n", readFile(t, path))
}

func TestRotatingFileMaxBackups(t *testing.T) {
	t.Parallel()

	file, path, now := testRotatingFile(t, 4, 2, false)

	// A file that only looks similar is never removed.
	other := filepath.Join(filepath.Dir(path), "windows_exporter-notes.log")
	require.NoError(t, os.WriteFile(other, []byte("keep"), 0o600))

	for i := range 5 {
		*now = now.Add(time.Second)

		write(t, file, strings.Repeat(string(rune('a'+i)), 4))
	}

	require.NoError(t, file.Close())

	require.Equal(t, []string{
		"windows_exporter-2026-10-15T14-03-09.000.log",
		"windows_exporter-2026-10-15T14-03-10.000.log",
		"windows_exporter-notes.log",
		"windows_exporter.log",
	}, dirEntries(t, filepath.Dir(path)))
}

func TestRotatingFileCompress(t *testing.T) {
	t.Parallel()

	file, path, now := testRotatingFile(t, 4, 0, true)

	write(t, file, "abcd")

	*now = now.Add(time.Second)

	write(t, file, "efgh")

	require.NoError(t, file.Close())

	require.Equal(t, []string{
		"windows_exporter-2026-10-15T14-03-06.000.log.gz",
		"windows_exporter.log",
	}, dirEntries(t, filepath.Dir(path)))

	compressed, err := os.Open(filepath.Join(filepath.Dir(path), "windows_exporter-2026-10-15T14-03-06.000.log.gz"))
	require.NoError(t, err)

	defer compressed.Close()

	reader, err := gzip.NewReader(compressed)
	require.NoError(t, err)

	data, err := io.ReadAll(reader)
	require.NoError(t, err)
	require.Equal(t, "abcd", string(data))
}

func TestRotatingFileCloseBoundsCleanup(t *testing.T) {
	t.Parallel()

	file, _, _ := testRotatingFile(t, 10, 0, false)
	file.closeTimeout = 50 * time.Millisecond

	// A cleanup that only returns once Close aborts it, like a compression
	// that is still reading.
	file.cleanup.Go(func() {
		<-file.stop
	})

	start := time.Now()
	err := file.Close()
	elapsed := time.Since(start)

	require.ErrorContains(t, err, "aborted the cleanup of rotated log files")
	require.GreaterOrEqual(t, elapsed, file.closeTimeout)
	require.Less(t, elapsed, file.closeTimeout+cleanupAbortTimeout, "Close did not return when the aborted cleanup did")

	// A cleanup that ignores the abort delays Close by at most cleanupAbortTimeout.
	block := make(chan struct{})
	defer close(block)

	file.cleanup.Go(func() {
		<-block
	})

	start = time.Now()

	require.Error(t, file.Close())
	require.Less(t, time.Since(start), file.closeTimeout+cleanupAbortTimeout+time.Second)
}

func TestCompressFileAborts(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "windows_exporter-2026-10-15T14-03-05.000.log")
	require.NoError(t, os.WriteFile(path, []byte("abcd"), 0o600))

	stop := make(chan struct{})
	close(stop)

	require.ErrorIs(t, compressFile(path, stop), errCompressionAborted)
	require.Equal(t, []string{filepath.Base(path)}, dirEntries(t, filepath.Dir(path)), "the rotated file must stay uncompressed, without a temporary file")
	require.Equal(t, "abcd", readFile(t, path))
}

func TestRotatingFileBackupNamesAreUnique(t *testing.T) {
	t.Parallel()

	// The clock doesn't move, so every rotation has the same timestamp.
	file, path, _ := testRotatingFile(t, 4, 0, false)

	write(t, file, "aaaa")
	write(t, file, "bbbb")
	write(t, file, "cccc")

	require.NoError(t, file.Close())

	require.Equal(t, []string{
		"windows_exporter-2026-10-15T14-03-05.000-1.log",
		"windows_exporter-2026-10-15T14-03-05.000.log",
		"windows_exporter.log",
	}, dirEntries(t, filepath.Dir(path)))
	require.Equal(t, "aaaa", readFile(t, filepath.Join(filepath.Dir(path), "windows_exporter-2026-10-15T14-03-05.000.log")))
	require.Equal(t, "bbbb", readFile(t, filepath.Join(filepath.Dir(path), "windows_exporter-2026-10-15T14-03-05.000-1.log")))
}

func TestRotatingFileWriteAfterClose(t *testing.T) {
	t.Parallel()

	file, _, _ := testRotatingFile(t, 10, 0, false)
	require.NoError(t, file.Close())

	_, err := file.Write([]byte("late\n"))
	require.ErrorIs(t, err, os.ErrClosed)
	require.NoError(t, file.Close(), "a second Close is a no-op")
}

func TestRotatingFileRecreatesMovedFileOnRotation(t *testing.T) {
	t.Parallel()

	file, path, _ := testRotatingFile(t, 100, 0, false)

	write(t, file, strings.Repeat("a", 59)+"\n")

	moved := filepath.Join(filepath.Dir(path), "moved.log")
	require.NoError(t, os.Rename(path, moved))

	// The rotation finds no file to rename and starts a new one at the path.
	write(t, file, strings.Repeat("b", 59)+"\n")
	require.NoError(t, file.Close())

	require.Equal(t, strings.Repeat("a", 59)+"\n", readFile(t, moved))
	require.Equal(t, strings.Repeat("b", 59)+"\n", readFile(t, path))
}

func TestRotatingFileRecreatesMovedOrDeletedFile(t *testing.T) {
	t.Parallel()

	for name, remove := range map[string]func(path string) error{
		"moved":   func(path string) error { return os.Rename(path, path+".moved") },
		"deleted": os.Remove,
		"replaced": func(path string) error {
			return errors.Join(os.Rename(path, path+".old"), os.WriteFile(path, []byte("other\n"), 0o600))
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// Far below max-size: no rotation is involved.
			file, path, now := testRotatingFile(t, 1<<20, 0, false)

			write(t, file, "before\n")
			require.NoError(t, remove(path))

			// Until the next check, lines go to the old file.
			write(t, file, "until the check\n")

			*now = now.Add(file.retryInterval)

			write(t, file, "after\n")
			require.NoError(t, file.Close())

			require.True(t, strings.HasSuffix(readFile(t, path), "after\n"), "the path has a new log file")
			require.NotContains(t, readFile(t, path), "before", "the new file doesn't continue the old one")
		})
	}
}

func TestRotatingFileKeepsFileAtPath(t *testing.T) {
	t.Parallel()

	file, path, now := testRotatingFile(t, 1<<20, 0, false)

	write(t, file, "one\n")

	*now = now.Add(file.retryInterval)

	// The check finds the same file: nothing is reopened.
	write(t, file, "two\n")
	require.NoError(t, file.Close())

	require.Equal(t, "one\ntwo\n", readFile(t, path))
	require.Equal(t, []string{"windows_exporter.log"}, dirEntries(t, filepath.Dir(path)))
}

func TestRotatingFileWithoutSizeLimit(t *testing.T) {
	t.Parallel()

	// A plain log file: no size limit, never rotated.
	file, path, now := testRotatingFile(t, math.MaxInt64, 0, false)

	write(t, file, strings.Repeat("a", 1000))
	*now = now.Add(file.retryInterval)

	write(t, file, strings.Repeat("b", 1000))
	require.NoError(t, file.Close())

	require.Equal(t, []string{"windows_exporter.log"}, dirEntries(t, filepath.Dir(path)))
	require.Len(t, readFile(t, path), 2000)
}
