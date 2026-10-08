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

package ole

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"golang.org/x/sys/windows"
)

func TestHRESULTFormatting(t *testing.T) {
	t.Run("Win32 facility", func(t *testing.T) {
		message := formatHRESULT(HRESULT(0x80070005))
		require.Contains(t, message, "COM HRESULT 0x80070005")
		require.Contains(t, message, strings.TrimSpace(windows.ERROR_ACCESS_DENIED.Error()))
		require.Equal(t, strings.TrimSpace(message), message)
	})

	t.Run("COM system message", func(t *testing.T) {
		message := formatHRESULT(HRESULT(0x80004005))
		require.Contains(t, message, "COM HRESULT 0x80004005: ")
	})

	t.Run("unknown status", func(t *testing.T) {
		require.Equal(t, "COM HRESULT 0xDEADBEEF", formatHRESULT(HRESULT(0xdeadbeef)))
	})
}
