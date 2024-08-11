//go:build windows

package printer

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/alecthomas/kingpin/v2"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/prometheus-community/windows_exporter/pkg/types"
	"github.com/prometheus-community/windows_exporter/pkg/wmi"
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
	config Config
	logger log.Logger

	printerStatus    *prometheus.Desc
	printerJobStatus *prometheus.Desc
	printerJobCount  *prometheus.Desc
}

func New(logger log.Logger, config *Config) *Collector {
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

	c.SetLogger(logger)

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
	).Default(c.config.PrinterInclude.String()).StringVar(&printerInclude)

	app.Flag(
		"collector.printer.exclude",
		"Regular expression to match printers to exclude",
	).Default(c.config.PrinterExclude.String()).StringVar(&printerExclude)

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

func (c *Collector) SetLogger(logger log.Logger) {
	c.logger = log.With(logger, "collector", Name)
}

func (c *Collector) Close() error {
	return nil
}

func (c *Collector) Build() error {
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

func (c *Collector) GetPerfCounter() ([]string, error) { return []string{"Printer"}, nil }

type wmiPrinter struct {
	Name                   string
	Default                bool
	PrinterStatus          uint16
	JobCountSinceLastReset uint32
}

type wmiPrintJob struct {
	Name   string
	Status string
}

func (c *Collector) Collect(_ *types.ScrapeContext, ch chan<- prometheus.Metric) error {
	if err := c.collectPrinterStatus(ch); err != nil {
		_ = level.Error(c.logger).Log("msg", "failed to collect printer status metrics", "err", err)
		return err
	}

	if err := c.collectPrinterJobStatus(ch); err != nil {
		_ = level.Error(c.logger).Log("msg", "failed to collect printer job status metrics", "err", err)
		return err
	}

	return nil
}

func (c *Collector) collectPrinterStatus(ch chan<- prometheus.Metric) error {
	var printers []wmiPrinter

	q := wmi.QueryAllForClass(&printers, "win32_Printer", c.logger)
	if err := wmi.Query(q, &printers); err != nil {
		return err
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

	q := wmi.QueryAllForClass(&printJobs, "win32_PrintJob", c.logger)
	if err := wmi.Query(q, &printJobs); err != nil {
		return err
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
