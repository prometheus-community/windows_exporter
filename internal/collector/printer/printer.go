//go:build windows

package printer

import (
	"errors"
	"fmt"
	"log/slog"
	"regexp"
	"strings"

	"github.com/alecthomas/kingpin/v2"
	"github.com/prometheus-community/windows_exporter/internal/mi"
	"github.com/prometheus-community/windows_exporter/internal/types"
	"github.com/prometheus/client_golang/prometheus"
)

const Name = "printer"

// printerStatusMap source: https://learn.microsoft.com/en-us/windows/win32/cimwin32prov/win32-printer#:~:text=Power%20Save-,PrinterStatus,Offline%20(7),-PrintJobDataType
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
	PrinterInclude *regexp.Regexp `yaml:"printer_include"`
	PrinterExclude *regexp.Regexp `yaml:"printer_exclude"`
}

var ConfigDefaults = Config{
	PrinterInclude: types.RegExpAny,
	PrinterExclude: types.RegExpEmpty,
}

type Collector struct {
	config             Config
	miSession          *mi.Session
	miQueryPrinterJobs mi.Query
	miQueryPrinter     mi.Query

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

func (c *Collector) Build(_ *slog.Logger, miSession *mi.Session) error {
	if miSession == nil {
		return errors.New("miSession is nil")
	}

	miQuery, err := mi.NewQuery("SELECT Name, Default, PrinterStatus, JobCountSinceLastReset FROM win32_Printer")
	if err != nil {
		return fmt.Errorf("failed to create WMI query: %w", err)
	}

	c.miQueryPrinter = miQuery

	miQuery, err = mi.NewQuery("SELECT Name, Status FROM win32_PrintJob")
	if err != nil {
		return fmt.Errorf("failed to create WMI query: %w", err)
	}

	c.miQueryPrinterJobs = miQuery
	c.miSession = miSession

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

type wmiPrinter struct {
	Name                   string `mi:"Name"`
	Default                bool   `mi:"Default"`
	PrinterStatus          uint16 `mi:"PrinterStatus"`
	JobCountSinceLastReset uint32 `mi:"JobCountSinceLastReset"`
}

type wmiPrintJob struct {
	Name   string `mi:"Name"`
	Status string `mi:"Status"`
}

func (c *Collector) Collect(ch chan<- prometheus.Metric) error {
	var errs []error

	if err := c.collectPrinterStatus(ch); err != nil {
		errs = append(errs, fmt.Errorf("failed to collect printer status metrics: %w", err))
	}

	if err := c.collectPrinterJobStatus(ch); err != nil {
		errs = append(errs, fmt.Errorf("failed to collect printer job status metrics: %w", err))
	}

	return errors.Join(errs...)
}

func (c *Collector) collectPrinterStatus(ch chan<- prometheus.Metric) error {
	var printers []wmiPrinter
	if err := c.miSession.Query(&printers, mi.NamespaceRootCIMv2, c.miQueryPrinter); err != nil {
		return fmt.Errorf("WMI query failed: %w", err)
	}

	for _, printer := range printers {
		if c.config.PrinterExclude.MatchString(printer.Name) ||
			!c.config.PrinterInclude.MatchString(printer.Name) {
			continue
		}

		for printerStatus, printerStatusName := range printerStatusMap {
			isCurrentStatus := 0.0
			if printerStatus == printer.PrinterStatus {
				isCurrentStatus = 1.0
			}

			ch <- prometheus.MustNewConstMetric(
				c.printerStatus,
				prometheus.GaugeValue,
				isCurrentStatus,
				printer.Name,
				printerStatusName,
			)
		}

		ch <- prometheus.MustNewConstMetric(
			c.printerJobCount,
			prometheus.CounterValue,
			float64(printer.JobCountSinceLastReset),
			printer.Name,
		)
	}

	return nil
}

func (c *Collector) collectPrinterJobStatus(ch chan<- prometheus.Metric) error {
	var printJobs []wmiPrintJob
	if err := c.miSession.Query(&printJobs, mi.NamespaceRootCIMv2, c.miQueryPrinterJobs); err != nil {
		return fmt.Errorf("WMI query failed: %w", err)
	}

	groupedPrintJobs := c.groupPrintJobs(printJobs)
	for group, count := range groupedPrintJobs {
		ch <- prometheus.MustNewConstMetric(
			c.printerJobStatus,
			prometheus.GaugeValue,
			float64(count),
			group.printerName,
			group.status,
		)
	}

	return nil
}

type PrintJobStatusGroup struct {
	printerName string
	status      string
}

func (c *Collector) groupPrintJobs(printJobs []wmiPrintJob) map[PrintJobStatusGroup]int {
	groupedPrintJobs := make(map[PrintJobStatusGroup]int)

	for _, printJob := range printJobs {
		printerName := strings.Split(printJob.Name, ",")[0]

		if c.config.PrinterExclude.MatchString(printerName) ||
			!c.config.PrinterInclude.MatchString(printerName) {
			continue
		}

		groupedPrintJobs[PrintJobStatusGroup{
			printerName: printerName,
			status:      printJob.Status,
		}]++
	}

	return groupedPrintJobs
}
