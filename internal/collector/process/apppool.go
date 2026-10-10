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

package process

import (
	"errors"
	"fmt"
	"strings"
	"unicode/utf16"
	"unsafe"

	"golang.org/x/sys/windows"
)

// workerProcessName is the process name of IIS worker processes (w3wp.exe).
const workerProcessName = "w3wp"

// errNoAppPoolArgument reports a w3wp command line without a non-empty -ap argument.
var errNoAppPoolArgument = errors.New("command line has no -ap argument")

// isWorkerProcess reports whether the process name is an IIS worker process.
func isWorkerProcess(name string) bool {
	return strings.EqualFold(name, workerProcessName)
}

// workerProcessAppPool returns the application pool of the IIS worker process pid.
//
// The Windows Process Activation Service (WAS) starts every worker process as
// `w3wp.exe -ap "<pool>" ...`. The pool name is read from the command line of
// the process. It opens the process with PROCESS_QUERY_LIMITED_INFORMATION only.
func workerProcessAppPool(pid uint32) (string, error) {
	cmdLine, err := queryProcessCommandLine(pid)
	if err != nil {
		return "", err
	}

	pool, ok := appPoolFromCommandLine(cmdLine)
	if !ok {
		return "", errNoAppPoolArgument
	}

	return pool, nil
}

// queryProcessCommandLine returns the command line of process pid.
//
// It uses NtQueryInformationProcess(ProcessCommandLineInformation), which is
// available since Windows 8.1 and Windows Server 2012 R2. Earlier versions
// return STATUS_INVALID_INFO_CLASS. Unlike reading the PEB, it needs only
// PROCESS_QUERY_LIMITED_INFORMATION, not PROCESS_VM_READ.
func queryProcessCommandLine(pid uint32) (string, error) {
	hProcess, err := windows.OpenProcess(windows.PROCESS_QUERY_LIMITED_INFORMATION, false, pid)
	if err != nil {
		return "", fmt.Errorf("OpenProcess: %w", err)
	}

	defer func() {
		_ = windows.CloseHandle(hProcess)
	}()

	return processCommandLine(hProcess)
}

// unicodeStringHeader mirrors UNICODE_STRING. Buffer is kept as an address
// because it points into the buffer returned by the kernel.
type unicodeStringHeader struct {
	Length        uint16
	MaximumLength uint16
	Buffer        uintptr
}

func processCommandLine(hProcess windows.Handle) (string, error) {
	// UNICODE_STRING header plus a typical w3wp command line.
	size := uint32(1024)

	var err error

	for range 4 {
		// []uint64 keeps the UNICODE_STRING header pointer-aligned.
		buf := make([]uint64, (size+7)/8)
		bufSize := uint32(len(buf) * 8)
		retLen := uint32(0)

		err = windows.NtQueryInformationProcess(hProcess, windows.ProcessCommandLineInformation, unsafe.Pointer(&buf[0]), bufSize, &retLen)
		if err == nil {
			return decodeUnicodeString(buf)
		}

		if !errors.Is(err, windows.STATUS_INFO_LENGTH_MISMATCH) &&
			!errors.Is(err, windows.STATUS_BUFFER_TOO_SMALL) &&
			!errors.Is(err, windows.STATUS_BUFFER_OVERFLOW) {
			break
		}

		if retLen <= bufSize {
			break
		}

		size = retLen
	}

	return "", fmt.Errorf("NtQueryInformationProcess(ProcessCommandLineInformation): %w", err)
}

// decodeUnicodeString decodes the UNICODE_STRING that the kernel placed at the
// start of buf. Its Buffer must point into buf.
func decodeUnicodeString(buf []uint64) (string, error) {
	base := unsafe.Pointer(&buf[0])
	header := (*unicodeStringHeader)(base)

	if header.Length == 0 {
		return "", nil
	}

	bufSize := uintptr(len(buf)) * 8
	offset := header.Buffer - uintptr(base)

	if header.Buffer < uintptr(base) || offset < unsafe.Sizeof(*header) || offset > bufSize ||
		uintptr(header.Length) > bufSize-offset || header.Length%2 != 0 {
		return "", fmt.Errorf("UNICODE_STRING with length %d at offset %d does not fit the %d byte buffer", header.Length, offset, bufSize)
	}

	data := unsafe.Slice((*uint16)(unsafe.Add(base, offset)), header.Length/2)

	return string(utf16.Decode(data)), nil
}

// appPoolFromCommandLine returns the value of the first -ap argument of a w3wp
// command line. Arguments are split with the CommandLineToArgvW rules.
//
// Application pool names can't contain '"' or '\', so the quoting WAS applies
// is unambiguous for every valid pool name.
func appPoolFromCommandLine(cmdLine string) (string, bool) {
	args := splitCommandLine(cmdLine)

	// args[0] is the program name.
	for i := 1; i < len(args)-1; i++ {
		if strings.EqualFold(args[i], "-ap") {
			return args[i+1], args[i+1] != ""
		}
	}

	return "", false
}

func isCommandLineSpace(c byte) bool {
	return c == ' ' || c == '\t'
}

// splitCommandLine splits a command line like CommandLineToArgvW does.
//
// The program name ends at the next space or control character, which is
// consumed, or, if it starts with a double quote, at the next double quote;
// backslashes are literal in it. In the other arguments:
//   - spaces and tabs outside double quotes separate arguments,
//   - 2n backslashes followed by a double quote produce n backslashes, and the
//     double quote starts or ends a quoted part,
//   - 2n+1 backslashes followed by a double quote produce n backslashes and a
//     literal double quote,
//   - backslashes not followed by a double quote are literal,
//   - inside a quoted part, every third consecutive double quote is literal.
//
// The rules only treat ASCII bytes specially, so the input is processed as
// UTF-8 bytes without decoding runes.
func splitCommandLine(cmdLine string) []string {
	if cmdLine == "" {
		return nil
	}

	var (
		args []string
		i    int
	)

	if cmdLine[0] == '"' {
		end := strings.IndexByte(cmdLine[1:], '"')
		if end < 0 {
			return []string{cmdLine[1:]}
		}

		args = append(args, cmdLine[1:end+1])
		i = end + 2
	} else {
		// Any control character or space ends an unquoted program name, and is consumed.
		for i < len(cmdLine) && cmdLine[i] > ' ' {
			i++
		}

		args = append(args, cmdLine[:i])
		i++
	}

	for i < len(cmdLine) && isCommandLineSpace(cmdLine[i]) {
		i++
	}

	var (
		arg         []byte
		inArg       bool
		quotes      int
		backslashes int
	)

	for i < len(cmdLine) {
		c := cmdLine[i]

		switch {
		case isCommandLineSpace(c) && quotes == 0:
			args = append(args, string(arg))
			arg = arg[:0]
			inArg = false
			backslashes = 0

			for i < len(cmdLine) && isCommandLineSpace(cmdLine[i]) {
				i++
			}

			continue
		case c == '\\':
			arg = append(arg, c)
			backslashes++
		case c == '"':
			if backslashes%2 == 0 {
				// 2n backslashes and a quote: n backslashes, the quote toggles quoting.
				arg = arg[:len(arg)-backslashes/2]
				quotes++
			} else {
				// 2n+1 backslashes and a quote: n backslashes and a literal quote.
				arg = append(arg[:len(arg)-backslashes/2-1], '"')
			}

			backslashes = 0

			for i+1 < len(cmdLine) && cmdLine[i+1] == '"' {
				i++
				quotes++

				if quotes == 3 {
					arg = append(arg, '"')
					quotes = 0
				}
			}

			if quotes == 2 {
				quotes = 0
			}
		default:
			arg = append(arg, c)
			backslashes = 0
		}

		inArg = true
		i++
	}

	if inArg {
		args = append(args, string(arg))
	}

	return args
}
