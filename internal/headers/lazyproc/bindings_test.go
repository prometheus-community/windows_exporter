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

package lazyproc_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/prometheus-community/windows_exporter/internal/headers/cfgmgr32"
	"github.com/prometheus-community/windows_exporter/internal/headers/dhcpsapi"
	"github.com/prometheus-community/windows_exporter/internal/headers/gdi32"
	"github.com/prometheus-community/windows_exporter/internal/headers/hcn"
	"github.com/prometheus-community/windows_exporter/internal/headers/hcs"
	"github.com/prometheus-community/windows_exporter/internal/headers/iphlpapi"
	"github.com/prometheus-community/windows_exporter/internal/headers/kernel32"
	"github.com/prometheus-community/windows_exporter/internal/headers/ntdll"
	"github.com/prometheus-community/windows_exporter/internal/headers/psapi"
	"github.com/prometheus-community/windows_exporter/internal/headers/slc"
	"github.com/prometheus-community/windows_exporter/internal/headers/sysinfoapi"
	"github.com/prometheus-community/windows_exporter/internal/headers/winspool"
	"github.com/prometheus-community/windows_exporter/internal/headers/wtsapi32"
	"github.com/prometheus-community/windows_exporter/internal/pdh"
	"github.com/stretchr/testify/require"
	"golang.org/x/sys/windows"
)

// TestBindingsLoad checks the export names of every binding package against
// the DLLs of this host. A DLL that belongs to an optional Windows feature
// skips its binding when the DLL is not installed.
func TestBindingsLoad(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name string
		load func() error
		// optionalDLL is set for DLLs of optional Windows features.
		optionalDLL string
	}{
		{name: "cfgmgr32", load: cfgmgr32.Load},
		{name: "dhcpsapi", load: dhcpsapi.Load, optionalDLL: "dhcpsapi.dll"},
		{name: "gdi32", load: gdi32.Load},
		{name: "hcn", load: hcn.Load, optionalDLL: "vmcompute.dll"},
		{name: "hcs", load: hcs.Load, optionalDLL: "computecore.dll"},
		{name: "iphlpapi", load: iphlpapi.Load},
		{name: "kernel32", load: kernel32.Load},
		{name: "ntdll", load: ntdll.Load},
		{name: "psapi", load: psapi.Load},
		{name: "slc", load: slc.Load},
		{name: "sysinfoapi", load: sysinfoapi.Load},
		{name: "winspool", load: winspool.Load},
		{name: "wtsapi32", load: wtsapi32.Load},
		{name: "pdh", load: pdh.Load},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if tc.optionalDLL != "" {
				systemDirectory, err := windows.GetSystemDirectory()
				require.NoError(t, err)

				if _, err := os.Stat(filepath.Join(systemDirectory, tc.optionalDLL)); err != nil {
					t.Skipf("%s is not installed: %v", tc.optionalDLL, err)
				}
			}

			require.NoError(t, tc.load())
		})
	}
}
