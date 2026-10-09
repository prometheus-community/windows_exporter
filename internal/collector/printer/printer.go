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
	"errors"
	"fmt"
	"log/slog"
	"regexp"
	"time"

	"github.com/alecthomas/kingpin/v2"
	"github.com/prometheus-community/windows_exporter/internal/headers/winspool"
	"github.com/prometheus-community/windows_exporter/internal/mi"
	"github.com/prometheus-community/windows_exporter/internal/types"
	"github.com/prometheus/client_golang/prometheus"
	"golang.org/x/sys/windows"
)

const Name = "printer"

// Win32_Printer PrinterStatus values.
// https://learn.microsoft.com/en-us/windows/win32/cimwin32prov/win32-printer
const (
	printerStatusOther    uint16 = 1
	printerStatusUnknown  uint16 = 2
	printerStatusIdle     uint16 = 3
	printerStatusPrinting uint16 = 4
	printerStatusWarmup   uint16 = 5
)

// printerStatusMap source: https://learn.microsoft.com/en-us/windows/win32/cimwin32prov/win32-printer#:~:text=Power%20Save-,PrinterStatus,Offline%20(7),-PrintJobDataType
//
//nolint:gochecknoglobals
var printerStatusMap = map[uint16]string{
	1: "Other",
	2: "Unknown",
	3: "Idle",
	4: "Printing",
	5: "Warmup",
	6: "Stopped Printing",
	7: "Offline",
}

type Config struct {
	PrinterInclude *regexp.Regexp `yaml:"include"`
	PrinterExclude *regexp.Regexp `yaml:"exclude"`
}

//nolint:gochecknoglobals
var ConfigDefaults = Config{
	PrinterInclude: types.RegExpAny,
	PrinterExclude: types.RegExpEmpty,
}

type Collector struct {
	config Config

	printerStatus    *prometheus.Desc
	printerJobStatus *prometheus.Desc
	printerJobCount  *prometheus.Desc
}

func New(config *Config) *Collector {
	if config == nil {
		config = &ConfigDefaults
	}

	if config.PrinterExclude == nil {
		config.PrinterExclude = ConfigDefaults.PrinterExclude
	}

	if config.PrinterInclude == nil {
		config.PrinterInclude = ConfigDefaults.PrinterInclude
	}

	c := &Collector{
		config: *config,
	}

	return c
}

func NewWithFlags(app *kingpin.Application) *Collector {
	c := &Collector{
		config: ConfigDefaults,
	}

	var printerInclude, printerExclude string

	app.Flag(
		"collector.printer.include",
		"Regular expression to match printers to collect metrics for",
	).Default(".+").StringVar(&printerInclude)

	app.Flag(
		"collector.printer.exclude",
		"Regular expression to match printers to exclude",
	).Default("").StringVar(&printerExclude)

	app.Action(func(*kingpin.ParseContext) error {
		var err error

		c.config.PrinterInclude, err = regexp.Compile(fmt.Sprintf("^(?:%s)$", printerInclude))
		if err != nil {
			return fmt.Errorf("collector.printer.include: %w", err)
		}

		c.config.PrinterExclude, err = regexp.Compile(fmt.Sprintf("^(?:%s)$", printerExclude))
		if err != nil {
			return fmt.Errorf("collector.printer.exclude: %w", err)
		}

		return nil
	})

	return c
}

func (c *Collector) Close() error {
	return nil
}

func (c *Collector) Build(_ *slog.Logger, _ *mi.Session) error {
	c.printerJobStatus = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "job_status"),
		"A counter of printer jobs by status",
		[]string{"printer", "status"},
		nil,
	)
	c.printerStatus = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "status"),
		"Printer status",
		[]string{"printer", "status"},
		nil,
	)
	c.printerJobCount = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "job_count"),
		"Number of jobs processed by the printer since the last reset",
		[]string{"printer"},
		nil,
	)

	return nil
}

func (c *Collector) GetName() string { return Name }

// Collect reads printers and jobs from the print spooler.
// Win32_Printer is not used because the CIMWin32 provider calls DeviceCapabilities
// for every printer, which blocks for about a minute on an unreachable network printer.
func (c *Collector) Collect(ch chan<- prometheus.Metric, _ time.Duration) error {
	// CIMWin32 enumerates the same printers for Win32_Printer.
	printers, err := winspool.EnumPrinters(winspool.PRINTER_ENUM_LOCAL | winspool.PRINTER_ENUM_CONNECTIONS)
	if err != nil {
		return fmt.Errorf("failed to enumerate printers: %w", err)
	}

	var errs []error

	for _, printer := range printers {
		if c.config.PrinterExclude.MatchString(printer.Name) ||
			!c.config.PrinterInclude.MatchString(printer.Name) {
			continue
		}

		// Win32_Printer JobCountSinceLastReset is PRINTER_INFO_2.cJobs, the number of queued jobs.
		ch <- prometheus.MustNewConstMetric(
			c.printerJobCount,
			prometheus.CounterValue,
			float64(printer.Jobs),
			printer.Name,
		)

		jobs, err := printerJobs(printer)
		if err != nil {
			errs = append(errs, fmt.Errorf("failed to enumerate jobs of printer %s: %w", printer.Name, err))

			continue
		}

		currentStatus := printerStatus(printer.Status, jobs)

		for status, statusName := range printerStatusMap {
			isCurrentStatus := 0.0
			if status == currentStatus {
				isCurrentStatus = 1.0
			}

			ch <- prometheus.MustNewConstMetric(
				c.printerStatus,
				prometheus.GaugeValue,
				isCurrentStatus,
				printer.Name,
				statusName,
			)
		}

		for status, count := range groupJobsByStatus(jobs) {
			ch <- prometheus.MustNewConstMetric(
				c.printerJobStatus,
				prometheus.GaugeValue,
				float64(count),
				printer.Name,
				status,
			)
		}
	}

	return errors.Join(errs...)
}

// printerJobs returns the jobs of the printer in queue order.
func printerJobs(printer winspool.Printer) (_ []winspool.Job, err error) {
	// Skip OpenPrinter for an empty queue. For a printer connection, it can contact the print server.
	if printer.Jobs == 0 {
		return nil, nil
	}

	handle, err := winspool.OpenPrinter(printer.Name)
	if err != nil {
		// The printer was deleted after EnumPrinters.
		if errors.Is(err, windows.ERROR_INVALID_PRINTER_NAME) {
			return nil, nil
		}

		return nil, err
	}

	defer func() {
		err = errors.Join(err, winspool.ClosePrinter(handle))
	}()

	return winspool.EnumJobs(handle)
}

// printerStatus returns the Win32_Printer PrinterStatus value for PRINTER_INFO_2.Status,
// mapped the same way as the CIMWin32 provider (cimwin32.dll).
//
// CIMWin32 compares the whole Status value with single PRINTER_STATUS_* flags.
// For 0 and for combinations of flags, it uses the first job in the queue instead.
// It never reports Stopped Printing (6) or Offline (7); an offline printer is Other (1).
func printerStatus(status uint32, jobs []winspool.Job) uint16 {
	switch status {
	case winspool.PRINTER_STATUS_PAUSED,
		winspool.PRINTER_STATUS_ERROR,
		winspool.PRINTER_STATUS_PENDING_DELETION,
		winspool.PRINTER_STATUS_PAPER_JAM,
		winspool.PRINTER_STATUS_PAPER_OUT,
		winspool.PRINTER_STATUS_MANUAL_FEED,
		winspool.PRINTER_STATUS_PAPER_PROBLEM,
		winspool.PRINTER_STATUS_OFFLINE,
		winspool.PRINTER_STATUS_OUTPUT_BIN_FULL,
		winspool.PRINTER_STATUS_TONER_LOW,
		winspool.PRINTER_STATUS_NO_TONER,
		winspool.PRINTER_STATUS_PAGE_PUNT,
		winspool.PRINTER_STATUS_USER_INTERVENTION,
		winspool.PRINTER_STATUS_OUT_OF_MEMORY,
		winspool.PRINTER_STATUS_DOOR_OPEN:
		return printerStatusOther
	case winspool.PRINTER_STATUS_IO_ACTIVE,
		winspool.PRINTER_STATUS_BUSY,
		winspool.PRINTER_STATUS_PRINTING,
		winspool.PRINTER_STATUS_PROCESSING:
		return printerStatusPrinting
	case winspool.PRINTER_STATUS_NOT_AVAILABLE:
		return printerStatusUnknown
	case winspool.PRINTER_STATUS_WAITING,
		winspool.PRINTER_STATUS_SERVER_UNKNOWN,
		winspool.PRINTER_STATUS_POWER_SAVE:
		return printerStatusIdle
	case winspool.PRINTER_STATUS_INITIALIZING,
		winspool.PRINTER_STATUS_WARMING_UP:
		return printerStatusWarmup
	}

	if len(jobs) == 0 {
		return printerStatusIdle
	}

	job := jobs[0]

	switch {
	case job.StatusText != "":
		return printerStatusUnknown
	case job.Status&(winspool.JOB_STATUS_PAUSED|winspool.JOB_STATUS_ERROR|winspool.JOB_STATUS_DELETING|
		winspool.JOB_STATUS_OFFLINE|winspool.JOB_STATUS_PAPEROUT|winspool.JOB_STATUS_PRINTED) != 0:
		return printerStatusOther
	case job.Status&(winspool.JOB_STATUS_SPOOLING|winspool.JOB_STATUS_PRINTING) != 0:
		return printerStatusPrinting
	default:
		return printerStatusUnknown
	}
}

// jobStatus returns the Win32_PrintJob Status value for JOB_INFO_1.Status,
// mapped the same way as the CIMWin32 provider (cimwin32.dll).
func jobStatus(status uint32) string {
	switch {
	case status&winspool.JOB_STATUS_ERROR != 0:
		return "Error"
	case status&(winspool.JOB_STATUS_PAUSED|winspool.JOB_STATUS_OFFLINE|winspool.JOB_STATUS_PAPEROUT) != 0:
		return "Degraded"
	case status&(winspool.JOB_STATUS_DELETING|winspool.JOB_STATUS_SPOOLING|winspool.JOB_STATUS_PRINTING|winspool.JOB_STATUS_PRINTED) != 0:
		return "OK"
	default:
		return "UNKNOWN"
	}
}

func groupJobsByStatus(jobs []winspool.Job) map[string]int {
	groupedJobs := make(map[string]int)

	for _, job := range jobs {
		groupedJobs[jobStatus(job.Status)]++
	}

	return groupedJobs
}
