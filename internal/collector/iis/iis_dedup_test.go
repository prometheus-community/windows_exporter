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

package iis

import (
	"testing"
)

// testEntry is a minimal collectorName implementation used in tests.
type testEntry struct {
	Name string
}

func (e testEntry) GetName() string { return e.Name }

func names(entries []testEntry) []string {
	out := make([]string, len(entries))
	for i, e := range entries {
		out[i] = e.Name
	}

	return out
}

func mkEntries(ss ...string) []testEntry {
	out := make([]testEntry, len(ss))
	for i, s := range ss {
		out[i] = testEntry{Name: s}
	}

	return out
}

func TestIISCounterBaseName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input string
		want  string
	}{
		// Standard IIS counter suffixes – numeric only.
		{"Site_B#2", "Site_B"},
		{"Site_B#10", "Site_B"},
		// Pool name legitimately contains '#' – not a numeric suffix.
		{"App#Pool", "App#Pool"},
		{"Pool#Name#2", "Pool#Name"},
		// No '#' at all.
		{"DefaultAppPool", "DefaultAppPool"},
		// Edge-cases.
		{"#2", ""},
		{"Site#", "Site#"},   // empty suffix → not treated as counter
		{"Site##2", "Site#"}, // double '#', last segment is numeric
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			t.Parallel()

			got := iisCounterBaseName(tt.input)
			if got != tt.want {
				t.Errorf("iisCounterBaseName(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestDeduplicateIISNames(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input []testEntry
		// wantNames is the set of Name fields present in the result (order-independent).
		wantNames []string
	}{
		{
			name:      "no duplicates",
			input:     mkEntries("Site_A", "Site_B", "Site_C"),
			wantNames: []string{"Site_A", "Site_B", "Site_C"},
		},
		{
			name:      "one recycled instance",
			input:     mkEntries("Site_A", "Site_B", "Site_C", "Site_B#2"),
			wantNames: []string{"Site_A", "Site_B#2", "Site_C"},
		},
		{
			name:      "multiple recycled suffixes – keep highest",
			input:     mkEntries("Site_B", "Site_B#3", "Site_B#2"),
			wantNames: []string{"Site_B#3"},
		},
		{
			name:      "multiple recycled suffixes with high numbers – keep highest",
			input:     mkEntries("Site_B", "Site_B#9", "Site_B#10"),
			wantNames: []string{"Site_B#10"},
		},
		{
			name:      "pool name with hash character (not a suffix)",
			input:     mkEntries("App#Pool", "Other Pool"),
			wantNames: []string{"App#Pool", "Other Pool"},
		},
		{
			name:      "pool name with hash does not clash with recycled entry",
			input:     mkEntries("App#Pool", "App#Pool#2"),
			wantNames: []string{"App#Pool#2"},
		},
		{
			name:      "space in pool name",
			input:     mkEntries("My App Pool", "Default App Pool"),
			wantNames: []string{"My App Pool", "Default App Pool"},
		},
		{
			name:      "empty input",
			input:     mkEntries(),
			wantNames: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := deduplicateIISNames(tt.input)
			gotNames := names(got)

			if len(gotNames) != len(tt.wantNames) {
				t.Fatalf("deduplicateIISNames(%v) returned %v, want %v",
					names(tt.input), gotNames, tt.wantNames)
			}

			wantSet := make(map[string]bool, len(tt.wantNames))
			for _, n := range tt.wantNames {
				wantSet[n] = true
			}

			for _, n := range gotNames {
				if !wantSet[n] {
					t.Errorf("unexpected name %q in result %v (want %v)", n, gotNames, tt.wantNames)
				}
			}
		})
	}
}
