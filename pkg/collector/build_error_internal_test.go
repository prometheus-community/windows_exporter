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

package collector

import (
	"errors"
	"fmt"
	"testing"

	"github.com/prometheus-community/windows_exporter/internal/headers/lazyproc"
	"github.com/stretchr/testify/require"
	"golang.org/x/sys/windows"
)

// TestExpectedBuildError_MissingNativeAPI checks that a collector whose
// binding Load fails because of a missing DLL or export is skipped, also
// inside the joined errors of collectors with sub-collectors.
func TestExpectedBuildError_MissingNativeAPI(t *testing.T) {
	t.Parallel()

	kernel32 := windows.NewLazySystemDLL("kernel32.dll")
	missingDLL := windows.NewLazySystemDLL("windows_exporter_missing.dll")

	missingExport := lazyproc.Find(kernel32, kernel32.NewProc("GetTickCount64"), kernel32.NewProc("WindowsExporterMissingExport"))
	require.Error(t, missingExport)

	missingLibrary := lazyproc.Find(missingDLL, missingDLL.NewProc("A"), missingDLL.NewProc("B"))
	require.Error(t, missingLibrary)

	for name, tc := range map[string]struct {
		err      error
		expected bool
	}{
		"missing export": {
			err:      fmt.Errorf("error build collector printer: %w", missingExport),
			expected: true,
		},
		"missing DLL": {
			err:      fmt.Errorf("error build collector dhcp: DHCP server API not available: %w", missingLibrary),
			expected: true,
		},
		"joined bindings": {
			err:      fmt.Errorf("error build collector gpu: GPU APIs not available: %w", errors.Join(missingExport, missingLibrary)),
			expected: true,
		},
		"hyperv sub-collectors": {
			err: fmt.Errorf("error build collector hyperv: %w", errors.Join(
				fmt.Errorf("failed to build host collector: %w", missingLibrary),
				fmt.Errorf("failed to build wmi_health collector: %w", missingExport),
			)),
			expected: true,
		},
		"hyperv sub-collectors with an unexpected failure": {
			err: fmt.Errorf("error build collector hyperv: %w", errors.Join(
				fmt.Errorf("failed to build host collector: %w", missingLibrary),
				fmt.Errorf("failed to build datastore collector: %w", windows.ERROR_ACCESS_DENIED),
			)),
			expected: false,
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			require.Equal(t, tc.expected, expectedBuildError(tc.err), tc.err)
		})
	}
}
