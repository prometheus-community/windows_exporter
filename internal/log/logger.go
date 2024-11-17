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
	"golang.org/x/sys/windows"
)

// AllowedFile is a settable identifier for the output file that the logger can have.
type AllowedFile struct {
	s string
	w io.Writer
}

func (f *AllowedFile) String() string {
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
		handle, err := windows.RegisterEventSource(nil, windows.StringToUTF16Ptr("windows_exporter"))
		if err != nil {
			return fmt.Errorf("failed to open event log: %w", err)
		}

		f.w = eventlog.NewEventLogWriter(handle)
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
