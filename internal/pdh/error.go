// SPDX-License-Identifier: MIT
//
// Copyright The Prometheus Authors

//go:build windows

package pdh

import "errors"

var (
	ErrNoData                           = NewPdhError(NoData)
	ErrPerformanceCounterNotInitialized = errors.New("performance counter not initialized")
)

// Error represents error returned from Performance Counters API.
type Error struct {
	ErrorCode uint32
	errorText string
}

func (m *Error) Is(err error) bool {
	if err == nil {
		return false
	}

	if e, ok := errors.AsType[*Error](err); ok {
		return m.ErrorCode == e.ErrorCode
	}

	return false
}

func (m *Error) Error() string {
	return m.errorText
}

func NewPdhError(code uint32) error {
	return &Error{
		ErrorCode: code,
		errorText: FormatError(code),
	}
}
