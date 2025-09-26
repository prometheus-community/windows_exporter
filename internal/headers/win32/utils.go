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

package win32

// ParseMultiSz splits a UTF-16 encoded MULTI_SZ buffer (Windows style) into
// individual UTF-16 string slices.
//
// A MULTI_SZ buffer is a sequence of UTF-16 strings separated by single null
// terminators (0x0000) and terminated by an extra null (i.e., two consecutive
// nulls) to mark the end of the list.
//
// Example layout in memory (UTF-16):
//
//	"foo\0bar\0baz\0\0"
//
// Given such a []uint16, this function returns a [][]uint16 where each inner
// slice is one null-terminated string segment without the trailing null.
//
// The returned slices reference the original buffer (no copying).
func ParseMultiSz(buf []uint16) [][]uint16 {
	var (
		result [][]uint16
		start  int
	)

	for i := range buf {
		if buf[i] == 0 {
			// Found a null terminator.
			if i == start {
				// Two consecutive nulls → end of list.
				break
			}

			// Append current string slice (excluding null).
			result = append(result, buf[start:i])
			// Move start to next character after null.
			start = i + 1
		}
	}

	return result
}
