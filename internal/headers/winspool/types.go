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
	"golang.org/x/sys/windows"
)

// EnumPrinters flags.
// https://learn.microsoft.com/en-us/windows/win32/printdocs/enumprinters
const (
	PRINTER_ENUM_LOCAL       uint32 = 0x00000002
	PRINTER_ENUM_CONNECTIONS uint32 = 0x00000004
)

// PRINTER_STATUS_* values of PRINTER_INFO_2.Status.
// https://learn.microsoft.com/en-us/windows/win32/printdocs/printer-info-2
const (
	PRINTER_STATUS_PAUSED               uint32 = 0x00000001
	PRINTER_STATUS_ERROR                uint32 = 0x00000002
	PRINTER_STATUS_PENDING_DELETION     uint32 = 0x00000004
	PRINTER_STATUS_PAPER_JAM            uint32 = 0x00000008
	PRINTER_STATUS_PAPER_OUT            uint32 = 0x00000010
	PRINTER_STATUS_MANUAL_FEED          uint32 = 0x00000020
	PRINTER_STATUS_PAPER_PROBLEM        uint32 = 0x00000040
	PRINTER_STATUS_OFFLINE              uint32 = 0x00000080
	PRINTER_STATUS_IO_ACTIVE            uint32 = 0x00000100
	PRINTER_STATUS_BUSY                 uint32 = 0x00000200
	PRINTER_STATUS_PRINTING             uint32 = 0x00000400
	PRINTER_STATUS_OUTPUT_BIN_FULL      uint32 = 0x00000800
	PRINTER_STATUS_NOT_AVAILABLE        uint32 = 0x00001000
	PRINTER_STATUS_WAITING              uint32 = 0x00002000
	PRINTER_STATUS_PROCESSING           uint32 = 0x00004000
	PRINTER_STATUS_INITIALIZING         uint32 = 0x00008000
	PRINTER_STATUS_WARMING_UP           uint32 = 0x00010000
	PRINTER_STATUS_TONER_LOW            uint32 = 0x00020000
	PRINTER_STATUS_NO_TONER             uint32 = 0x00040000
	PRINTER_STATUS_PAGE_PUNT            uint32 = 0x00080000
	PRINTER_STATUS_USER_INTERVENTION    uint32 = 0x00100000
	PRINTER_STATUS_OUT_OF_MEMORY        uint32 = 0x00200000
	PRINTER_STATUS_DOOR_OPEN            uint32 = 0x00400000
	PRINTER_STATUS_SERVER_UNKNOWN       uint32 = 0x00800000
	PRINTER_STATUS_POWER_SAVE           uint32 = 0x01000000
	PRINTER_STATUS_SERVER_OFFLINE       uint32 = 0x02000000
	PRINTER_STATUS_DRIVER_UPDATE_NEEDED uint32 = 0x04000000
)

// JOB_STATUS_* values of JOB_INFO_1.Status.
// https://learn.microsoft.com/en-us/windows/win32/printdocs/job-info-1
const (
	JOB_STATUS_PAUSED            uint32 = 0x00000001
	JOB_STATUS_ERROR             uint32 = 0x00000002
	JOB_STATUS_DELETING          uint32 = 0x00000004
	JOB_STATUS_SPOOLING          uint32 = 0x00000008
	JOB_STATUS_PRINTING          uint32 = 0x00000010
	JOB_STATUS_OFFLINE           uint32 = 0x00000020
	JOB_STATUS_PAPEROUT          uint32 = 0x00000040
	JOB_STATUS_PRINTED           uint32 = 0x00000080
	JOB_STATUS_DELETED           uint32 = 0x00000100
	JOB_STATUS_BLOCKED_DEVQ      uint32 = 0x00000200
	JOB_STATUS_USER_INTERVENTION uint32 = 0x00000400
	JOB_STATUS_RESTART           uint32 = 0x00000800
	JOB_STATUS_COMPLETE          uint32 = 0x00001000
	JOB_STATUS_RETAINED          uint32 = 0x00002000
)

// printerInfo2 is PRINTER_INFO_2.
// https://learn.microsoft.com/en-us/windows/win32/printdocs/printer-info-2
type printerInfo2 struct {
	pServerName         *uint16
	pPrinterName        *uint16
	pShareName          *uint16
	pPortName           *uint16
	pDriverName         *uint16
	pComment            *uint16
	pLocation           *uint16
	pDevMode            uintptr
	pSepFile            *uint16
	pPrintProcessor     *uint16
	pDatatype           *uint16
	pParameters         *uint16
	pSecurityDescriptor uintptr
	Attributes          uint32
	Priority            uint32
	DefaultPriority     uint32
	StartTime           uint32
	UntilTime           uint32
	Status              uint32
	cJobs               uint32
	AveragePPM          uint32
}

// jobInfo1 is JOB_INFO_1.
// https://learn.microsoft.com/en-us/windows/win32/printdocs/job-info-1
type jobInfo1 struct {
	JobId        uint32
	pPrinterName *uint16
	pMachineName *uint16
	pUserName    *uint16
	pDocument    *uint16
	pDatatype    *uint16
	pStatus      *uint16
	Status       uint32
	Priority     uint32
	Position     uint32
	TotalPages   uint32
	PagesPrinted uint32
	Submitted    windows.Systemtime
}

// Printer contains the PRINTER_INFO_2 fields used by windows_exporter.
type Printer struct {
	Name string
	// Status is a combination of PRINTER_STATUS_* values.
	Status uint32
	// Jobs is the number of jobs currently queued for the printer (cJobs).
	Jobs uint32
}

// Job contains the JOB_INFO_1 fields used by windows_exporter.
type Job struct {
	ID          uint32
	PrinterName string
	// Status is a combination of JOB_STATUS_* values.
	Status uint32
	// StatusText is the driver or port monitor status string (pStatus).
	// It is empty if the spooler reports no custom status.
	StatusText string
}
