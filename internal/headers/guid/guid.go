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

package guid

import (
	"fmt"
	"strconv"
	"strings"

	"golang.org/x/sys/windows"
)

type GUID windows.GUID

// FromString parses a string containing a GUID and returns the GUID. The only
// format currently supported is the `xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx`
// format.
func FromString(s string) (GUID, error) {
	if len(s) != 36 {
		return GUID{}, fmt.Errorf("invalid GUID %q", s)
	}

	if s[8] != '-' || s[13] != '-' || s[18] != '-' || s[23] != '-' {
		return GUID{}, fmt.Errorf("invalid GUID %q", s)
	}

	var g GUID

	data1, err := strconv.ParseUint(s[0:8], 16, 32)
	if err != nil {
		return GUID{}, fmt.Errorf("invalid GUID %q", s)
	}

	g.Data1 = uint32(data1)

	data2, err := strconv.ParseUint(s[9:13], 16, 16)
	if err != nil {
		return GUID{}, fmt.Errorf("invalid GUID %q", s)
	}

	g.Data2 = uint16(data2)

	data3, err := strconv.ParseUint(s[14:18], 16, 16)
	if err != nil {
		return GUID{}, fmt.Errorf("invalid GUID %q", s)
	}

	g.Data3 = uint16(data3)

	for i, x := range []int{19, 21, 24, 26, 28, 30, 32, 34} {
		v, err := strconv.ParseUint(s[x:x+2], 16, 8)
		if err != nil {
			return GUID{}, fmt.Errorf("invalid GUID %q", s)
		}

		g.Data4[i] = uint8(v)
	}

	return g, nil
}

func (g *GUID) UnmarshalJSON(b []byte) error {
	guid, err := FromString(strings.Trim(strings.Trim(string(b), `"`), `{}`))
	if err != nil {
		return err
	}

	*g = guid

	return nil
}

func (g *GUID) String() string {
	return fmt.Sprintf(
		"%08x-%04x-%04x-%04x-%012x",
		g.Data1,
		g.Data2,
		g.Data3,
		g.Data4[:2],
		g.Data4[2:])
}
