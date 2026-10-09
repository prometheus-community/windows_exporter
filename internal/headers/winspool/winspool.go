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

// Package winspool reads printers and print jobs through the print spooler API.
//
// EnumPrintersW and EnumJobsW only read the spooler state. Unlike the CIMWin32
// Win32_Printer provider, they never call DeviceCapabilities, which can block
// for about a minute on an unreachable network printer.
package winspool

import (
	"errors"
	"fmt"
	"unsafe"

	"golang.org/x/sys/windows"
)

//nolint:gochecknoglobals
var (
	modWinspool = windows.NewLazySystemDLL("winspool.drv")

	procEnumPrintersW = modWinspool.NewProc("EnumPrintersW")
	procOpenPrinterW  = modWinspool.NewProc("OpenPrinterW")
	procClosePrinter  = modWinspool.NewProc("ClosePrinter")
	procEnumJobsW     = modWinspool.NewProc("EnumJobsW")
)

// EnumPrinters returns the printers selected by flags, a combination of PRINTER_ENUM_* values.
func EnumPrinters(flags uint32) ([]Printer, error) {
	buf, count, err := enum(func(buf *byte, size uint32, needed, returned *uint32) (uintptr, error) {
		r1, _, err := procEnumPrintersW.Call(
			uintptr(flags),
			0, // local machine
			2, // PRINTER_INFO_2
			uintptr(unsafe.Pointer(buf)),
			uintptr(size),
			uintptr(unsafe.Pointer(needed)),
			uintptr(unsafe.Pointer(returned)),
		)

		return r1, err
	})
	if err != nil {
		return nil, fmt.Errorf("EnumPrintersW: %w", err)
	}

	if count == 0 {
		return []Printer{}, nil
	}

	infos := unsafe.Slice((*printerInfo2)(unsafe.Pointer(&buf[0])), count)
	printers := make([]Printer, 0, count)

	for _, info := range infos {
		printers = append(printers, Printer{
			Name:   windows.UTF16PtrToString(info.pPrinterName),
			Status: info.Status,
			Jobs:   info.cJobs,
		})
	}

	return printers, nil
}

// OpenPrinter opens printerName with the default access rights of the caller.
// The handle must be closed with ClosePrinter.
func OpenPrinter(printerName string) (windows.Handle, error) {
	printerNamePtr, err := windows.UTF16PtrFromString(printerName)
	if err != nil {
		return 0, fmt.Errorf("invalid printer name %q: %w", printerName, err)
	}

	var handle windows.Handle

	r1, _, err := procOpenPrinterW.Call(
		uintptr(unsafe.Pointer(printerNamePtr)),
		uintptr(unsafe.Pointer(&handle)),
		0, // PRINTER_DEFAULTS
	)
	if r1 == 0 {
		return 0, fmt.Errorf("OpenPrinterW(%s): %w", printerName, err)
	}

	return handle, nil
}

func ClosePrinter(handle windows.Handle) error {
	r1, _, err := procClosePrinter.Call(uintptr(handle))
	if r1 == 0 {
		return fmt.Errorf("ClosePrinter: %w", err)
	}

	return nil
}

// EnumJobs returns all jobs of the printer in queue order.
func EnumJobs(handle windows.Handle) ([]Job, error) {
	buf, count, err := enum(func(buf *byte, size uint32, needed, returned *uint32) (uintptr, error) {
		r1, _, err := procEnumJobsW.Call(
			uintptr(handle),
			0,          // FirstJob
			0xFFFFFFFF, // NoJobs
			1,          // JOB_INFO_1
			uintptr(unsafe.Pointer(buf)),
			uintptr(size),
			uintptr(unsafe.Pointer(needed)),
			uintptr(unsafe.Pointer(returned)),
		)

		return r1, err
	})
	if err != nil {
		return nil, fmt.Errorf("EnumJobsW: %w", err)
	}

	if count == 0 {
		return []Job{}, nil
	}

	infos := unsafe.Slice((*jobInfo1)(unsafe.Pointer(&buf[0])), count)
	jobs := make([]Job, 0, count)

	for _, info := range infos {
		jobs = append(jobs, Job{
			ID:          info.JobId,
			PrinterName: windows.UTF16PtrToString(info.pPrinterName),
			Status:      info.Status,
			StatusText:  windows.UTF16PtrToString(info.pStatus),
		})
	}

	return jobs, nil
}

// enum calls a spooler enumeration function until the buffer is large enough.
// The returned buffer holds count structures followed by the strings they point to.
func enum(call func(buf *byte, size uint32, needed, returned *uint32) (uintptr, error)) ([]byte, uint32, error) {
	var (
		buf      []byte
		needed   uint32
		returned uint32
	)

	// The required size can grow between two calls if printers or jobs are added.
	for range 5 {
		var bufPtr *byte
		if len(buf) > 0 {
			bufPtr = &buf[0]
		}

		r1, err := call(bufPtr, uint32(len(buf)), &needed, &returned)
		if r1 != 0 {
			return buf, returned, nil
		}

		if !errors.Is(err, windows.ERROR_INSUFFICIENT_BUFFER) {
			return nil, 0, err
		}

		buf = make([]byte, needed)
	}

	return nil, 0, windows.ERROR_INSUFFICIENT_BUFFER
}
