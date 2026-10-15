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
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/prometheus/common/promslog"
	"github.com/stretchr/testify/require"
)

func newConfig(t *testing.T, file string, rotation Rotation) *Config {
	t.Helper()

	allowedFile := &AllowedFile{}
	require.NoError(t, allowedFile.Set(file))

	return &Config{Config: &promslog.Config{}, File: allowedFile, Rotation: rotation}
}

func TestNewFileWithoutRotation(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "windows_exporter.log")

	logger, closer, err := New(newConfig(t, path, Rotation{}))
	require.NoError(t, err)
	require.NotNil(t, closer)

	logger.Info("hello")
	require.NoError(t, closer.Close())

	data, err := os.ReadFile(path)
	require.NoError(t, err)
	require.Contains(t, string(data), "hello")
}

func writeMegabytes(logger *slog.Logger, megabytes int) {
	message := strings.Repeat("x", 1024)
	for range megabytes * 1100 {
		logger.Info(message)
	}
}

func TestNewFileWithSizeRotation(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "windows_exporter.log")

	logger, closer, err := New(newConfig(t, path, Rotation{MaxSize: 1}))
	require.NoError(t, err)
	require.NotNil(t, closer)

	writeMegabytes(logger, 2)
	require.NoError(t, closer.Close())

	entries, err := os.ReadDir(dir)
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(entries), 2, "expected the active log file and at least one backup")

	info, err := os.Stat(path)
	require.NoError(t, err)
	require.LessOrEqual(t, info.Size(), int64(1024*1024))
}

func TestNewFileMaxBackups(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "windows_exporter.log")

	logger, closer, err := New(newConfig(t, path, Rotation{MaxSize: 1, MaxBackups: 2}))
	require.NoError(t, err)

	// Rotates at least four times; only the two newest backups are kept.
	writeMegabytes(logger, 5)
	require.NoError(t, closer.Close())

	entries, err := os.ReadDir(dir)
	require.NoError(t, err)
	require.Len(t, entries, 3, "expected the active log file and two backups")
}

func TestNewFileUnwritable(t *testing.T) {
	t.Parallel()

	// The parent of the log file is a file, so the log file can't be created.
	parent := filepath.Join(t.TempDir(), "not-a-directory")
	require.NoError(t, os.WriteFile(parent, nil, 0o600))

	path := filepath.Join(parent, "windows_exporter.log")

	for name, rotation := range map[string]Rotation{
		"without rotation": {},
		"with rotation":    {MaxSize: 1},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			_, _, err := New(newConfig(t, path, rotation))
			require.ErrorContains(t, err, "failed to open log file")
		})
	}
}

func TestNewRotationValidation(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "windows_exporter.log")

	for name, tc := range map[string]struct {
		file     string
		rotation Rotation
		err      string
	}{
		"stream":           {"stdout", Rotation{MaxSize: 10}, "requires log.file to be a file path"},
		"no max size":      {path, Rotation{MaxBackups: 3}, "requires log.rotation.max-size"},
		"negative size":    {path, Rotation{MaxSize: -1}, "max-size must not be negative"},
		"negative backups": {path, Rotation{MaxSize: 1, MaxBackups: -1}, "max-backups must not be negative"},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			_, _, err := New(newConfig(t, tc.file, tc.rotation))
			require.ErrorContains(t, err, tc.err)
		})
	}
}

func TestNewStreamHasNoCloser(t *testing.T) {
	t.Parallel()

	_, closer, err := New(newConfig(t, "stderr", Rotation{}))
	require.NoError(t, err)
	require.Nil(t, closer)
}
