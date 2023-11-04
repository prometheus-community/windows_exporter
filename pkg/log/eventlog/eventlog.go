//go:build windows
// +build windows

// Package eventlog provides a Logger that writes to Windows Event Log.
package eventlog

import (
	"bytes"
	"fmt"
	"io"
	"sync"
	"syscall"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"golang.org/x/sys/windows"
	goeventlog "golang.org/x/sys/windows/svc/eventlog"
)

const (
	// NeLogOemCode is a generic error log entry for OEMs to use to
	// elog errors from OEM value added services.
	// See: https://github.com/microsoft/win32metadata/blob/2f3c5282ce1024a712aeccd90d3aa50bf7a49e27/generation/WinSDK/RecompiledIdlHeaders/um/LMErrlog.h#L824-L845
	neLogOemCode = uint32(3299)
)

type Priority struct {
	etype int
}

// NewEventLogLogger returns a new Logger which writes to Windows EventLog in event log format.
// The body of the log message is the formatted output from the Logger returned
// by newLogger.
func NewEventLogLogger(w *goeventlog.Log, newLogger func(io.Writer) log.Logger) log.Logger {
	l := &eventlogLogger{
		w:                w,
		newLogger:        newLogger,
		prioritySelector: defaultPrioritySelector,
		bufPool: sync.Pool{New: func() interface{} {
			return &loggerBuf{}
		}},
	}

	return l
}

type eventlogLogger struct {
	w                *goeventlog.Log
	newLogger        func(io.Writer) log.Logger
	prioritySelector PrioritySelector
	bufPool          sync.Pool
}

func (l *eventlogLogger) Log(keyvals ...interface{}) error {
	priority := l.prioritySelector(keyvals...)

	lb := l.getLoggerBuf()
	defer l.putLoggerBuf(lb)
	if err := lb.logger.Log(keyvals...); err != nil {
		return err
	}

	// golang.org/x/sys/windows/svc/eventlog does not provide func which allows to send more than one string.
	// See: https://github.com/golang/go/issues/59780

	msg, err := syscall.UTF16PtrFromString(lb.buf.String())
	if err != nil {
		return fmt.Errorf("error convert string to UTF-16: %v", err)
	}

	ss := []*uint16{msg, nil, nil, nil, nil, nil, nil, nil, nil}
	return windows.ReportEvent(l.w.Handle, uint16(priority.etype), 0, neLogOemCode, 0, 9, 0, &ss[0], nil)
}

type loggerBuf struct {
	buf    *bytes.Buffer
	logger log.Logger
}

func (l *eventlogLogger) getLoggerBuf() *loggerBuf {
	lb := l.bufPool.Get().(*loggerBuf)
	if lb.buf == nil {
		lb.buf = &bytes.Buffer{}
		lb.logger = l.newLogger(lb.buf)
	} else {
		lb.buf.Reset()
	}
	return lb
}

func (l *eventlogLogger) putLoggerBuf(lb *loggerBuf) {
	l.bufPool.Put(lb)
}

// PrioritySelector inspects the list of keyvals and selects an eventlog priority.
type PrioritySelector func(keyvals ...interface{}) Priority

// defaultPrioritySelector convert a kit/log level into a Windows Eventlog level
func defaultPrioritySelector(keyvals ...interface{}) Priority {
	l := len(keyvals)

	eType := windows.EVENTLOG_SUCCESS

	for i := 0; i < l; i += 2 {
		if keyvals[i] == level.Key() {
			var val interface{}
			if i+1 < l {
				val = keyvals[i+1]
			}
			if v, ok := val.(level.Value); ok {
				switch v {
				case level.ErrorValue():
					eType = windows.EVENTLOG_ERROR_TYPE
				case level.WarnValue():
					eType = windows.EVENTLOG_WARNING_TYPE
				case level.InfoValue():
					eType = windows.EVENTLOG_INFORMATION_TYPE
				}
			}
		}
	}

	return Priority{etype: eType}
}
