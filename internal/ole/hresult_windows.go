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
	"fmt"
	"strings"

	"golang.org/x/sys/windows"
)

func formatHRESULT(h HRESULT) string {
	numeric := fmt.Sprintf("COM HRESULT 0x%08X", uint32(h))

	code := uint32(h)
	if (code>>16)&0x1fff == 7 { // FACILITY_WIN32 stores the system error in the low word.
		code &= 0xffff
	}

	var buffer [1024]uint16

	n, err := windows.FormatMessage(
		windows.FORMAT_MESSAGE_FROM_SYSTEM|windows.FORMAT_MESSAGE_IGNORE_INSERTS,
		0,
		code,
		0,
		buffer[:],
		nil,
	)
	if err != nil {
		return numeric
	}

	message := strings.TrimSpace(windows.UTF16ToString(buffer[:n]))
	if message == "" {
		return numeric
	}

	return numeric + ": " + message
}
