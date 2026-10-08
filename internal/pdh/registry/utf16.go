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

package registry

import (
	"encoding/binary"
	"io"

	"golang.org/x/sys/windows"
)

// readUTF16StringAtPos Read an unterminated UTF16 string at a given position, specifying its length.
func readUTF16StringAtPos(r io.ReadSeeker, absPos int64, length uint32) (string, error) {
	value := make([]uint16, length/2)

	_, err := r.Seek(absPos, io.SeekStart)
	if err != nil {
		return "", err
	}

	err = binary.Read(r, bo, value)
	if err != nil {
		return "", err
	}

	return windows.UTF16ToString(value), nil
}

// readUTF16String Reads a null-terminated UTF16 string at the current offset.
func readUTF16String(r io.Reader) (string, error) {
	b := make([]byte, 2)
	out := make([]uint16, 0, 100)

	for {
		// A short read at the end of the buffer is an error, not a code unit.
		if _, err := io.ReadFull(r, b); err != nil {
			return "", err
		}

		if b[0] == 0 && b[1] == 0 {
			return windows.UTF16ToString(out), nil
		}

		out = append(out, bo.Uint16(b))
	}
}
