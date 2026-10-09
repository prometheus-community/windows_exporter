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

package winspool

import (
	"path/filepath"
	"slices"
	"testing"
	"unsafe"

	"github.com/stretchr/testify/require"
	"golang.org/x/sys/windows"
)

//nolint:gochecknoglobals
var (
	procStartDocPrinterW = modWinspool.NewProc("StartDocPrinterW")
	procAbortPrinter     = modWinspool.NewProc("AbortPrinter")
)

// docInfo1 is DOC_INFO_1.
type docInfo1 struct {
	pDocName    *uint16
	pOutputFile *uint16
	pDatatype   *uint16
}

func TestLayout(t *testing.T) {
	t.Parallel()

	if unsafe.Sizeof(uintptr(0)) != 8 {
		t.Skip("offsets are checked for 64-bit Windows only")
	}

	// Offsets of the 64-bit SDK structures.
	require.Equal(t, uintptr(136), unsafe.Sizeof(printerInfo2{}))
	require.Equal(t, uintptr(0x08), unsafe.Offsetof(printerInfo2{}.pPrinterName))
	require.Equal(t, uintptr(0x7c), unsafe.Offsetof(printerInfo2{}.Status))
	require.Equal(t, uintptr(0x80), unsafe.Offsetof(printerInfo2{}.cJobs))

	require.Equal(t, uintptr(96), unsafe.Sizeof(jobInfo1{}))
	require.Equal(t, uintptr(0x08), unsafe.Offsetof(jobInfo1{}.pPrinterName))
	require.Equal(t, uintptr(0x30), unsafe.Offsetof(jobInfo1{}.pStatus))
	require.Equal(t, uintptr(0x38), unsafe.Offsetof(jobInfo1{}.Status))
	require.Equal(t, uintptr(0x4c), unsafe.Offsetof(jobInfo1{}.Submitted))
}

func TestEnumPrinters(t *testing.T) {
	t.Parallel()

	printers, err := EnumPrinters(PRINTER_ENUM_LOCAL | PRINTER_ENUM_CONNECTIONS)
	require.NoError(t, err)
	require.NotNil(t, printers)

	for _, printer := range printers {
		require.NotEmpty(t, printer.Name)
	}
}

func TestOpenPrinterInvalidName(t *testing.T) {
	t.Parallel()

	_, err := OpenPrinter("windows_exporter test printer that does not exist")
	require.ErrorIs(t, err, windows.ERROR_INVALID_PRINTER_NAME)
}

// TestEnumJobs starts a document without writing data, so the job stays in the
// spooling state and never reaches the port. AbortPrinter deletes it again.
func TestEnumJobs(t *testing.T) {
	t.Parallel()

	printers, err := EnumPrinters(PRINTER_ENUM_LOCAL)
	require.NoError(t, err)

	// CIPrinter is created by the CI workflow. Microsoft Print to PDF exists on most desktops.
	idx := slices.IndexFunc(printers, func(p Printer) bool {
		return p.Name == "CIPrinter" || p.Name == "Microsoft Print to PDF"
	})
	if idx == -1 {
		t.Skip("no test printer available")
	}

	printerName := printers[idx].Name

	handle, err := OpenPrinter(printerName)
	require.NoError(t, err)

	t.Cleanup(func() {
		require.NoError(t, ClosePrinter(handle))
	})

	docName, err := windows.UTF16PtrFromString("windows_exporter test job")
	require.NoError(t, err)

	// An output file keeps Microsoft Print to PDF from prompting for a file name.
	outputFile, err := windows.UTF16PtrFromString(filepath.Join(t.TempDir(), "job.out"))
	require.NoError(t, err)

	docInfo := docInfo1{
		pDocName:    docName,
		pOutputFile: outputFile,
	}

	jobID, _, err := procStartDocPrinterW.Call(uintptr(handle), 1, uintptr(unsafe.Pointer(&docInfo)))
	require.NotZero(t, jobID, "StartDocPrinterW: %v", err)

	t.Cleanup(func() {
		r1, _, err := procAbortPrinter.Call(uintptr(handle))
		require.NotZero(t, r1, "AbortPrinter: %v", err)
	})

	jobs, err := EnumJobs(handle)
	require.NoError(t, err)

	idx = slices.IndexFunc(jobs, func(j Job) bool { return j.ID == uint32(jobID) })
	require.NotEqual(t, -1, idx, "job %d not found in %+v", jobID, jobs)

	job := jobs[idx]
	require.Equal(t, printerName, job.PrinterName)
	require.NotZero(t, job.Status&JOB_STATUS_SPOOLING, "status 0x%08X", job.Status)

	printers, err = EnumPrinters(PRINTER_ENUM_LOCAL)
	require.NoError(t, err)

	idx = slices.IndexFunc(printers, func(p Printer) bool { return p.Name == printerName })
	require.NotEqual(t, -1, idx)
	require.NotZero(t, printers[idx].Jobs)
}
