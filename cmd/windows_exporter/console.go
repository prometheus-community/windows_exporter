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

import (
	"golang.org/x/sys/windows"
)

// disableConsoleQuickEdit turns off the QuickEdit mode of the console the
// process is attached to. It returns whether the mode was changed and a
// function that restores the previous input mode. Without an attached console,
// for example as a service, it changes nothing and the restore is a no-op.
//
// With QuickEdit on, a click into a conhost window starts a text selection and
// the console stops accepting output until the selection ends. A log write
// then blocks in WriteConsole while holding the slog handler lock, so every
// scrape that logs waits behind it and /metrics stops answering. Ctrl+C copies
// the selection instead of interrupting the process, which is why such a hang
// looks like a crash without a log line (#2024).
//
// The console input buffer is opened as CONIN$, because the hang comes from
// writing to the console and happens with a redirected stdin as well.
func disableConsoleQuickEdit() (bool, func()) {
	noop := func() {}

	conin, err := windows.CreateFile(
		windows.StringToUTF16Ptr("CONIN$"),
		windows.GENERIC_READ|windows.GENERIC_WRITE,
		windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE,
		nil,
		windows.OPEN_EXISTING,
		0,
		0,
	)
	if err != nil {
		return false, noop
	}

	var mode uint32
	if err := windows.GetConsoleMode(conin, &mode); err != nil {
		_ = windows.CloseHandle(conin)

		return false, noop
	}

	// ENABLE_EXTENDED_FLAGS is required to change ENABLE_QUICK_EDIT_MODE.
	newMode := (mode | windows.ENABLE_EXTENDED_FLAGS) &^ windows.ENABLE_QUICK_EDIT_MODE

	// Until a process has set ENABLE_EXTENDED_FLAGS on a console, GetConsoleMode
	// reports the extended flags as clear although QuickEdit and insert mode are
	// on by default. Keep insert mode on in that case instead of switching the
	// shell to overwrite mode.
	if mode&windows.ENABLE_EXTENDED_FLAGS == 0 {
		newMode |= windows.ENABLE_INSERT_MODE
	}

	if err := windows.SetConsoleMode(conin, newMode); err != nil {
		_ = windows.CloseHandle(conin)

		return false, noop
	}

	return true, func() {
		// The mode as read restores the previous state: without the extended
		// flags, the console returns to its own defaults.
		_ = windows.SetConsoleMode(conin, mode)
		_ = windows.CloseHandle(conin)
	}
}
