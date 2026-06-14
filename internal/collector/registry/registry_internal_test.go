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

import "testing"

func TestParseKeyPathNormalization(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name      string
		input     string
		wantLabel string
	}{
		{"short hive lowercased", `HKLM\SOFTWARE\Foo`, `hklm\software\foo`},
		{"long hive normalized to short", `HKEY_LOCAL_MACHINE\SOFTWARE\Foo`, `hklm\software\foo`},
		{"forward slashes converted", `HKLM/SOFTWARE/Foo`, `hklm\software\foo`},
		{"mixed case sub path lowercased", `hklm\SoFtWaRe\BAR`, `hklm\software\bar`},
		{"surrounding slashes trimmed", `\HKLM\SOFTWARE\Foo\`, `hklm\software\foo`},
		{"hive only", `HKCU`, `hkcu`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			_, _, label, err := parseKeyPath(tc.input)
			if err != nil {
				t.Fatalf("parseKeyPath(%q) returned error: %v", tc.input, err)
			}

			if label != tc.wantLabel {
				t.Errorf("parseKeyPath(%q) label = %q, want %q", tc.input, label, tc.wantLabel)
			}
		})
	}
}
