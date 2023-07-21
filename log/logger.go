package log

import (
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/go-kit/log"
	"github.com/prometheus-community/windows_exporter/log/eventlog"
	"github.com/prometheus/common/promlog"
	goeventlog "golang.org/x/sys/windows/svc/eventlog"
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
		f.w = nil
	default:
		file, err := os.OpenFile(s, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0200)
		if err != nil {
			return err
		}
		f.w = file
	}
	return nil
}

// Config is a struct containing configurable settings for the logger
type Config struct {
	promlog.Config

	File *AllowedFile
}

func New(config *Config) (log.Logger, error) {
	if config.File == nil {
		return nil, errors.New("log file undefined")
	}

	if config.Format == nil {
		return nil, errors.New("log format undefined")
	}

	var (
		l          log.Logger
		loggerFunc func(io.Writer) log.Logger
	)

	switch config.Format.String() {
	case "json":
		loggerFunc = log.NewJSONLogger
	case "logfmt":
		loggerFunc = log.NewLogfmtLogger
	default:
		return nil, fmt.Errorf("unsupported log.format %q", config.Format.String())
	}

	if config.File.s == "eventlog" {
		w, err := goeventlog.Open("windows_exporter")
		if err != nil {
			return nil, err
		}
		l = eventlog.NewEventLogLogger(w, loggerFunc)
	} else if config.File.w == nil {
		panic("logger: file writer is nil")
	} else {
		l = loggerFunc(log.NewSyncWriter(config.File.w))
	}

	promlogConfig := promlog.Config{
		Format: config.Format,
		Level:  config.Level,
	}

	return promlog.NewWithLogger(l, &promlogConfig), nil
}
