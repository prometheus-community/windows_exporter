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

package printer

import (
	"fmt"
	"testing"

	"github.com/prometheus-community/windows_exporter/internal/headers/winspool"
	"github.com/stretchr/testify/require"
)

// The expected values follow the PrinterStatus switch in cimwin32.dll (10.0.26100).
func TestPrinterStatus(t *testing.T) {
	t.Parallel()

	for status, want := range map[uint32]uint16{
		winspool.PRINTER_STATUS_PAUSED:            printerStatusOther,
		winspool.PRINTER_STATUS_ERROR:             printerStatusOther,
		winspool.PRINTER_STATUS_PENDING_DELETION:  printerStatusOther,
		winspool.PRINTER_STATUS_PAPER_JAM:         printerStatusOther,
		winspool.PRINTER_STATUS_PAPER_OUT:         printerStatusOther,
		winspool.PRINTER_STATUS_MANUAL_FEED:       printerStatusOther,
		winspool.PRINTER_STATUS_PAPER_PROBLEM:     printerStatusOther,
		winspool.PRINTER_STATUS_OFFLINE:           printerStatusOther,
		winspool.PRINTER_STATUS_IO_ACTIVE:         printerStatusPrinting,
		winspool.PRINTER_STATUS_BUSY:              printerStatusPrinting,
		winspool.PRINTER_STATUS_PRINTING:          printerStatusPrinting,
		winspool.PRINTER_STATUS_OUTPUT_BIN_FULL:   printerStatusOther,
		winspool.PRINTER_STATUS_NOT_AVAILABLE:     printerStatusUnknown,
		winspool.PRINTER_STATUS_WAITING:           printerStatusIdle,
		winspool.PRINTER_STATUS_PROCESSING:        printerStatusPrinting,
		winspool.PRINTER_STATUS_INITIALIZING:      printerStatusWarmup,
		winspool.PRINTER_STATUS_WARMING_UP:        printerStatusWarmup,
		winspool.PRINTER_STATUS_TONER_LOW:         printerStatusOther,
		winspool.PRINTER_STATUS_NO_TONER:          printerStatusOther,
		winspool.PRINTER_STATUS_PAGE_PUNT:         printerStatusOther,
		winspool.PRINTER_STATUS_USER_INTERVENTION: printerStatusOther,
		winspool.PRINTER_STATUS_OUT_OF_MEMORY:     printerStatusOther,
		winspool.PRINTER_STATUS_DOOR_OPEN:         printerStatusOther,
		winspool.PRINTER_STATUS_SERVER_UNKNOWN:    printerStatusIdle,
		winspool.PRINTER_STATUS_POWER_SAVE:        printerStatusIdle,
	} {
		t.Run(fmt.Sprintf("0x%08X", status), func(t *testing.T) {
			t.Parallel()

			// A single flag ignores the jobs.
			require.Equal(t, want, printerStatus(status, nil))
			require.Equal(t, want, printerStatus(status, []winspool.Job{{Status: winspool.JOB_STATUS_ERROR}}))
		})
	}
}

func TestPrinterStatusFromJobs(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name   string
		status uint32
		jobs   []winspool.Job
		want   uint16
	}{
		{name: "empty queue", jobs: nil, want: printerStatusIdle},
		{name: "spooling", jobs: []winspool.Job{{Status: winspool.JOB_STATUS_SPOOLING}}, want: printerStatusPrinting},
		{name: "printing", jobs: []winspool.Job{{Status: winspool.JOB_STATUS_PRINTING}}, want: printerStatusPrinting},
		{name: "paused", jobs: []winspool.Job{{Status: winspool.JOB_STATUS_PAUSED | winspool.JOB_STATUS_SPOOLING}}, want: printerStatusOther},
		{name: "error", jobs: []winspool.Job{{Status: winspool.JOB_STATUS_ERROR | winspool.JOB_STATUS_PRINTING}}, want: printerStatusOther},
		{name: "printed", jobs: []winspool.Job{{Status: winspool.JOB_STATUS_PRINTED}}, want: printerStatusOther},
		{name: "no job flags", jobs: []winspool.Job{{Status: 0}}, want: printerStatusUnknown},
		{name: "blocked only", jobs: []winspool.Job{{Status: winspool.JOB_STATUS_BLOCKED_DEVQ}}, want: printerStatusUnknown},
		{name: "status text", jobs: []winspool.Job{{Status: winspool.JOB_STATUS_PRINTING, StatusText: "Toner low"}}, want: printerStatusUnknown},
		{
			name: "only the first job counts",
			jobs: []winspool.Job{{Status: winspool.JOB_STATUS_SPOOLING}, {Status: winspool.JOB_STATUS_ERROR}},
			want: printerStatusPrinting,
		},
		{
			name:   "combined printer flags use the jobs",
			status: winspool.PRINTER_STATUS_PAUSED | winspool.PRINTER_STATUS_OFFLINE,
			jobs:   nil,
			want:   printerStatusIdle,
		},
		{
			name:   "server offline uses the jobs",
			status: winspool.PRINTER_STATUS_SERVER_OFFLINE,
			jobs:   []winspool.Job{{Status: winspool.JOB_STATUS_PRINTING}},
			want:   printerStatusPrinting,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			require.Equal(t, tc.want, printerStatus(tc.status, tc.jobs))
		})
	}
}

// The expected values follow the Win32_PrintJob Status logic in cimwin32.dll (10.0.26100).
func TestJobStatus(t *testing.T) {
	t.Parallel()

	for status, want := range map[uint32]string{
		0:                                "UNKNOWN",
		winspool.JOB_STATUS_BLOCKED_DEVQ: "UNKNOWN",
		winspool.JOB_STATUS_SPOOLING:     "OK",
		winspool.JOB_STATUS_PRINTING:     "OK",
		winspool.JOB_STATUS_DELETING:     "OK",
		winspool.JOB_STATUS_PRINTED:      "OK",
		winspool.JOB_STATUS_PAUSED:       "Degraded",
		winspool.JOB_STATUS_OFFLINE:      "Degraded",
		winspool.JOB_STATUS_PAPEROUT:     "Degraded",
		winspool.JOB_STATUS_ERROR:        "Error",
		winspool.JOB_STATUS_PAUSED | winspool.JOB_STATUS_SPOOLING: "Degraded",
		winspool.JOB_STATUS_ERROR | winspool.JOB_STATUS_PAUSED:    "Error",
	} {
		require.Equal(t, want, jobStatus(status), "status 0x%08X", status)
	}
}

func TestGroupJobsByStatus(t *testing.T) {
	t.Parallel()

	require.Equal(t, map[string]int{"OK": 2, "Error": 1}, groupJobsByStatus([]winspool.Job{
		{Status: winspool.JOB_STATUS_SPOOLING},
		{Status: winspool.JOB_STATUS_PRINTING},
		{Status: winspool.JOB_STATUS_ERROR | winspool.JOB_STATUS_PRINTING},
	}))
	require.Empty(t, groupJobsByStatus(nil))
}
