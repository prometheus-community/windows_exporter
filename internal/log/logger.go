// Copyright 2024 The Prometheus Authors
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
	"os"

	"github.com/prometheus-community/windows_exporter/internal/log/eventlog"
	"github.com/prometheus/common/promslog"
	wineventlog "golang.org/x/sys/windows/svc/eventlog"
)

// AllowedFile is a settable identifier for the output file that the logger can have.
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
		file, err := os.OpenFile(s, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0o200)
		if err != nil {
			return fmt.Errorf("failed to open log file: %w", err)
		}

		f.w = file
	}

	return nil
}

// Config is a struct containing configurable settings for the logger.
type Config struct {
	*promslog.Config

	File *AllowedFile
}

func New(config *Config) (*slog.Logger, error) {
	if config.File == nil {
		return nil, errors.New("log file undefined")
	}

	config.Config.Writer = config.File.w
	config.Config.Style = promslog.SlogStyle

	return promslog.New(config.Config), nil
}
