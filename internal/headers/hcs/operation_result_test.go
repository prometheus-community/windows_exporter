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

package hcs

import (
	"testing"
	"unsafe"

	"github.com/stretchr/testify/require"
	"golang.org/x/sys/windows"
)

func TestOperationResultDocumentCleanup(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name     string
		result   uintptr
		document bool
	}{
		{"successful document", 0, true},
		{"successful empty", 0, false},
		{"failure diagnostic", 0x80070005, true},
		{"failure empty", 0x80070005, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var document *uint16
			if tc.document {
				encoded, err := windows.UTF16FromString("diagnostic")
				require.NoError(t, err)

				document = &encoded[0]
			}

			releases := 0
			release := func(pointer unsafe.Pointer) {
				require.Equal(t, unsafe.Pointer(document), pointer)

				releases++
			}

			result, err := consumeOperationResult(tc.result, document, release)
			if tc.result != 0 {
				require.ErrorIs(t, err, windows.ERROR_ACCESS_DENIED)
				require.Empty(t, result)
			} else {
				require.NoError(t, err)

				if tc.document {
					require.Equal(t, "diagnostic", result)
				} else {
					require.Empty(t, result)
				}
			}

			if tc.document {
				require.Equal(t, 1, releases)
			} else {
				require.Zero(t, releases)
			}
		})
	}
}
