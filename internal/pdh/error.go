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

	var e *Error
	if errors.As(err, &e) {
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
