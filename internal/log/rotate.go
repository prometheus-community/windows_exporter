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
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	// rotationRetryInterval is how long a log file whose rotation failed is
	// appended to before the rotation is tried again.
	rotationRetryInterval = time.Minute

	backupTimeFormat = "2006-01-02T15-04-05.000"
	gzipSuffix       = ".gz"
)

// rotatingFile appends to a log file and rotates it before a write would make
// it larger than maxSize. A rotation that fails, for example because another
// process holds the file open without delete sharing, never drops a line: the
// file is appended to and the rotation is retried after rotationRetryInterval.
type rotatingFile struct {
	path       string
	maxSize    int64
	maxBackups int
	compress   bool

	now           func() time.Time
	rename        func(oldPath, newPath string) error
	retryInterval time.Duration

	mu         sync.Mutex
	file       *os.File
	size       int64
	retryAfter time.Time
	// checkAfter is when the next write checks that the path still points
	// to the open file; an admin may have moved or deleted it.
	checkAfter time.Time

	// cleanup compresses and prunes rotated files in the background; Close
	// waits for it. cleanupMu runs one cleanup at a time.
	cleanup   sync.WaitGroup
	cleanupMu sync.Mutex
}

func newRotatingFile(path string, maxSize int64, maxBackups int, compress bool, now func() time.Time) (*rotatingFile, error) {
	return newRotatingFileWithRename(path, maxSize, maxBackups, compress, now, os.Rename)
}

func newRotatingFileWithRename(
	path string,
	maxSize int64,
	maxBackups int,
	compress bool,
	now func() time.Time,
	rename func(oldPath, newPath string) error,
) (*rotatingFile, error) {
	r := &rotatingFile{
		path:          path,
		maxSize:       maxSize,
		maxBackups:    maxBackups,
		compress:      compress,
		now:           now,
		rename:        rename,
		retryInterval: rotationRetryInterval,
	}

	file, size, err := openLogFileWithSize(path)
	if err != nil {
		return nil, err
	}

	r.file, r.size = file, size
	r.checkAfter = now().Add(r.retryInterval)

	// A file that is already too large is rotated right away. If that fails,
	// the logger still starts and the rotation is retried on a later write.
	if r.size >= r.maxSize {
		r.mu.Lock()
		_ = r.rotate()
		r.mu.Unlock()
	}

	return r, nil
}

func openLogFileWithSize(path string) (*os.File, int64, error) {
	file, err := openLogFile(path)
	if err != nil {
		return nil, 0, err
	}

	info, err := file.Stat()
	if err != nil {
		_ = file.Close()

		return nil, 0, fmt.Errorf("failed to read log file size: %w", err)
	}

	return file, info.Size(), nil
}

// Write appends p to the log file. p is never split across files.
func (r *rotatingFile) Write(p []byte) (int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.file == nil {
		return 0, os.ErrClosed
	}

	if now := r.now(); !now.Before(r.checkAfter) {
		r.checkAfter = now.Add(r.retryInterval)

		// The file was moved, deleted or replaced: start a new one at the
		// path. If that fails, keep writing to the current file.
		if same, err := sameFile(r.file, r.path); err != nil || !same {
			_ = r.reopen()
		}
	}

	if r.size > 0 && int64(len(p)) > r.maxSize-r.size && !r.now().Before(r.retryAfter) {
		// On failure, keep appending to the current file.
		_ = r.rotate()
	}

	n, err := r.file.Write(p)
	r.size += int64(n)

	return n, err
}

// Close closes the log file and waits for compression and pruning of rotated files.
func (r *rotatingFile) Close() error {
	r.mu.Lock()

	var err error

	if r.file != nil {
		err = r.file.Close()
		r.file = nil
	}

	r.mu.Unlock()

	r.cleanup.Wait()

	return err
}

// rotate renames the log file to a backup and opens a new one. r.mu must be held.
func (r *rotatingFile) rotate() error {
	backup := r.backupName()

	if err := r.rename(r.path, backup); err != nil {
		// The file was moved or deleted: there is nothing to rename, so
		// start a new file at the path.
		if errors.Is(err, os.ErrNotExist) {
			err = r.reopen()
			if err == nil {
				return nil
			}
		}

		r.retryAfter = r.now().Add(r.retryInterval)

		return fmt.Errorf("failed to rename log file: %w", err)
	}

	file, size, err := openLogFileWithSize(r.path)
	if err != nil {
		// Keep the current file under its name and keep writing to it.
		_ = r.rename(backup, r.path)
		r.retryAfter = r.now().Add(r.retryInterval)

		return fmt.Errorf("failed to open new log file: %w", err)
	}

	_ = r.file.Close()
	r.file, r.size = file, size
	r.retryAfter = time.Time{}

	r.cleanup.Go(func() {
		r.cleanupMu.Lock()
		defer r.cleanupMu.Unlock()

		r.cleanupBackups(backup)
	})

	return nil
}

// reopen opens the file at the path and writes to it from now on. r.mu must be held.
func (r *rotatingFile) reopen() error {
	file, size, err := openLogFileWithSize(r.path)
	if err != nil {
		return err
	}

	_ = r.file.Close()
	r.file, r.size = file, size
	r.retryAfter = time.Time{}

	return nil
}

// backupName returns <name>-<UTC time><ext> next to the log file, with a
// counter before the extension if that name, or its compressed version, exists.
func (r *rotatingFile) backupName() string {
	dir, prefix, ext := r.nameParts()
	stamp := r.now().UTC().Format(backupTimeFormat)

	for i := 0; ; i++ {
		name := prefix + stamp + ext
		if i > 0 {
			name = prefix + stamp + "-" + strconv.Itoa(i) + ext
		}

		path := filepath.Join(dir, name)

		if !exists(path) && !exists(path+gzipSuffix) {
			return path
		}
	}
}

// nameParts returns the directory of the log file, the prefix of its backups
// (name without extension and "-") and its extension.
func (r *rotatingFile) nameParts() (string, string, string) {
	base := filepath.Base(r.path)
	ext := filepath.Ext(base)

	return filepath.Dir(r.path), strings.TrimSuffix(base, ext) + "-", ext
}

func exists(path string) bool {
	_, err := os.Lstat(path)

	return err == nil
}

// cleanupBackups compresses the new backup and removes the oldest backups
// beyond maxBackups. Errors only leave files behind, so they are ignored.
func (r *rotatingFile) cleanupBackups(backup string) {
	if r.compress {
		_ = compressFile(backup)
	}

	if r.maxBackups <= 0 {
		return
	}

	backups := r.backups()
	if len(backups) <= r.maxBackups {
		return
	}

	for _, old := range backups[:len(backups)-r.maxBackups] {
		_ = os.Remove(old.path)
	}
}

type backupFile struct {
	path    string
	time    time.Time
	counter int
}

// backups returns the rotated files of the log file, oldest first. Files that
// don't match the backup name exactly are ignored.
func (r *rotatingFile) backups() []backupFile {
	dir, prefix, ext := r.nameParts()

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}

	var backups []backupFile

	for _, entry := range entries {
		name, ok := strings.CutPrefix(entry.Name(), prefix)
		if !ok || entry.IsDir() {
			continue
		}

		name = strings.TrimSuffix(name, gzipSuffix)

		name, ok = strings.CutSuffix(name, ext)
		if !ok || len(name) < len(backupTimeFormat) {
			continue
		}

		stamp, counter := name[:len(backupTimeFormat)], 0

		if rest := name[len(backupTimeFormat):]; rest != "" {
			number, ok := strings.CutPrefix(rest, "-")
			if !ok {
				continue
			}

			if counter, err = strconv.Atoi(number); err != nil {
				continue
			}
		}

		t, err := time.Parse(backupTimeFormat, stamp)
		if err != nil {
			continue
		}

		backups = append(backups, backupFile{path: filepath.Join(dir, entry.Name()), time: t, counter: counter})
	}

	slices.SortFunc(backups, func(a, b backupFile) int {
		if c := a.time.Compare(b.time); c != 0 {
			return c
		}

		return a.counter - b.counter
	})

	return backups
}

// compressFile replaces path with path.gz.
func compressFile(path string) (err error) {
	src, err := os.Open(path)
	if err != nil {
		return err
	}

	defer func() { _ = src.Close() }()

	tmp := path + gzipSuffix + ".tmp"

	dst, err := os.OpenFile(tmp, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o640)
	if err != nil {
		return err
	}

	defer func() {
		if err != nil {
			_ = dst.Close()
			_ = os.Remove(tmp)
		}
	}()

	writer := gzip.NewWriter(dst)

	if _, err = io.Copy(writer, src); err != nil {
		return err
	}

	if err = writer.Close(); err != nil {
		return err
	}

	if err = dst.Close(); err != nil {
		return err
	}

	_ = src.Close()

	if err = os.Rename(tmp, path+gzipSuffix); err != nil {
		return err
	}

	return os.Remove(path)
}
