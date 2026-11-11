// SPDX-License-Identifier: MIT
//
// Copyright The Prometheus Authors

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

func TestSanitizeMetricName(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name  string
		input string
		want  string
	}{
		{"already valid", "windows_registry_foo", "windows_registry_foo"},
		{"spaces and case", "Windows NT Version", "windows_nt_version"},
		{"backslashes from key path", `windows_registry_hklm\software\foo`, "windows_registry_hklm_software_foo"},
		{"surrounding non-alphanumerics trimmed", "%value%", "value"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if got := sanitizeMetricName(tc.input); got != tc.want {
				t.Errorf("sanitizeMetricName(%q) = %q, want %q", tc.input, got, tc.want)
			}
		})
	}
}
