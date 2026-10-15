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
	"errors"
	"fmt"
	"io"
	"log/slog"
	"math"
	"os"
	"time"

	"github.com/prometheus-community/windows_exporter/internal/log/eventlog"
	"github.com/prometheus/common/promslog"
	wineventlog "golang.org/x/sys/windows/svc/eventlog"
)

// AllowedFile is a settable identifier for the output file that the logger can have.
// A path is only stored by Set; New opens it, since the rotation settings are
// parsed after the file flag.
type AllowedFile struct {
	s string
	w io.Writer
}

func (f *AllowedFile) String() string {
	if f == nil {
		return ""
	}

	return f.s
}

// Set updates the value of the allowed format.
func (f *AllowedFile) Set(s string) error {
	f.s = s

	switch s {
	case "stdout":
		f.w = os.Stdout
	case "stderr":
		f.w = os.Stderr
	case "eventlog":
		eventLog, err := wineventlog.Open("windows_exporter")
		if err != nil {
			return fmt.Errorf("failed to open event log: %w", err)
		}

		f.w = eventlog.NewEventLogWriter(eventLog)
	default:
		f.w = nil
	}

	return nil
}

// isPath reports whether the output is a log file path rather than a stream.
func (f *AllowedFile) isPath() bool {
	return f.s != "" && f.w == nil
}

// Rotation configures the size-based rotation of a log file. Rotation is
// disabled unless MaxSize is set.
type Rotation struct {
	// MaxSize is the size in megabytes at which the log file is rotated. 0 disables rotation.
	MaxSize int
	// MaxBackups is the number of rotated files to keep. 0 keeps all.
	MaxBackups int
	// Compress compresses rotated files with gzip.
	Compress bool
}

func (r Rotation) configured() bool {
	return r != Rotation{}
}

func (r Rotation) validate() error {
	var errs []error

	if r.MaxSize < 0 {
		errs = append(errs, errors.New("log.rotation.max-size must not be negative"))
	}

	if r.MaxBackups < 0 {
		errs = append(errs, errors.New("log.rotation.max-backups must not be negative"))
	}

	return errors.Join(errs...)
}

// Config is a struct containing configurable settings for the logger.
type Config struct {
	*promslog.Config

	File     *AllowedFile
	Rotation Rotation
}

// New creates the logger. The returned io.Closer releases the log file; it is
// nil for stdout, stderr and the event log.
func New(config *Config) (*slog.Logger, io.Closer, error) {
	if config.File == nil {
		return nil, nil, errors.New("log file undefined")
	}

	if err := config.Rotation.validate(); err != nil {
		return nil, nil, err
	}

	writer := config.File.w

	var closer io.Closer

	switch {
	case !config.File.isPath():
		if config.Rotation.configured() {
			return nil, nil, fmt.Errorf("log rotation requires log.file to be a file path, not %s", config.File.s)
		}
	case config.Rotation.MaxSize > 0:
		file, err := newRotatingFile(config.File.s, int64(config.Rotation.MaxSize)<<20, config.Rotation.MaxBackups, config.Rotation.Compress, time.Now)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to open log file: %w", err)
		}

		writer, closer = file, file
	case config.Rotation.configured():
		return nil, nil, errors.New("log rotation requires log.rotation.max-size")
	default:
		// Without a size limit the file is never rotated, but a moved or
		// deleted file is still replaced by a new one at the path.
		file, err := newRotatingFile(config.File.s, math.MaxInt64, 0, false, time.Now)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to open log file: %w", err)
		}

		writer, closer = file, file
	}

	config.Writer = writer
	config.Style = promslog.SlogStyle

	return promslog.New(config.Config), closer, nil
}
