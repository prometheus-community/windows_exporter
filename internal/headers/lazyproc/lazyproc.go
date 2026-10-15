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

// Package lazyproc checks that lazily loaded native functions exist.
//
// windows.LazyProc.Call panics when its DLL or export is missing, e.g. on a
// Windows version or edition without the subsystem. A binding package lists
// the functions it calls without a check of its own in a Load function built
// with [Once], and collectors call it in Build, so a missing export fails
// Build instead of panicking on the first scrape.
package lazyproc

import (
	"errors"
	"fmt"
	"strings"
	"sync"

	"golang.org/x/sys/windows"
)

// Error reports a missing DLL or missing exports. It matches
// [errors.ErrUnsupported] and has no other cause, so a collector that fails
// Build with it is skipped as an unavailable subsystem: the loader errors,
// e.g. ERROR_MOD_NOT_FOUND, are only kept as text and in Missing.
type Error struct {
	// DLL is the name of the DLL.
	DLL string
	// Missing holds the error of the DLL, if it could not be loaded, or of
	// each missing export.
	Missing []error
}

func (e *Error) Error() string {
	var message strings.Builder

	fmt.Fprintf(&message, "%s: %s", errors.ErrUnsupported, e.DLL)

	for _, err := range e.Missing {
		message.WriteString(": ")
		message.WriteString(err.Error())
	}

	return message.String()
}

// Is reports whether target is [errors.ErrUnsupported].
func (e *Error) Is(target error) bool {
	return target == errors.ErrUnsupported
}

// Find loads dll and every proc, which must belong to dll. It returns nil if
// all of them exist, and otherwise an [*Error] that names the missing DLL or
// each missing export.
func Find(dll *windows.LazyDLL, procs ...*windows.LazyProc) error {
	if err := dll.Load(); err != nil {
		return &Error{DLL: dll.Name, Missing: []error{err}}
	}

	var missing []error

	for _, proc := range procs {
		if err := proc.Find(); err != nil {
			missing = append(missing, err)
		}
	}

	if len(missing) == 0 {
		return nil
	}

	return &Error{DLL: dll.Name, Missing: missing}
}

// Once returns a function that runs [Find] once and then returns its result.
// A DLL or export that is missing stays missing for the life of the process.
func Once(dll *windows.LazyDLL, procs ...*windows.LazyProc) func() error {
	return sync.OnceValue(func() error {
		return Find(dll, procs...)
	})
}
