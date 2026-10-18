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

package main

import "golang.org/x/sys/windows"

// disableConsoleQuickEdit turns off the QuickEdit mode of the console the
// exporter is attached to and returns a function that restores the previous
// input mode. It returns a no-op when stdin is not a console, for example when
// the exporter runs as a service or with redirected input.
//
// With QuickEdit on, a click into the console window starts a text selection
// and the console stops accepting output until the selection ends. A log
// write then blocks in WriteConsole while holding the slog handler lock, so
// every scrape that logs waits behind it and /metrics stops answering. Ctrl+C
// copies the selection instead of interrupting the process, which is why the
// hang looks like a crash without any log line (#2024).
func disableConsoleQuickEdit() func() {
	noop := func() {}

	var mode uint32
	if err := windows.GetConsoleMode(windows.Stdin, &mode); err != nil {
		return noop
	}

	// ENABLE_EXTENDED_FLAGS is required to change ENABLE_QUICK_EDIT_MODE.
	if err := windows.SetConsoleMode(windows.Stdin, (mode|windows.ENABLE_EXTENDED_FLAGS)&^windows.ENABLE_QUICK_EDIT_MODE); err != nil {
		return noop
	}

	return func() {
		_ = windows.SetConsoleMode(windows.Stdin, mode|windows.ENABLE_EXTENDED_FLAGS)
	}
}
