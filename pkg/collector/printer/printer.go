//go:build windows

package printer

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/alecthomas/kingpin/v2"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/prometheus-community/windows_exporter/pkg/types"
	"github.com/prometheus-community/windows_exporter/pkg/wmi"
)

const (
	Name = "printer"

	FlagPrinterInclude = "collector.printer.include"
	FlagPrinterExclude = "collector.printer.exclude"
)

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
	Include string `yaml:"printer_include"`
	Exclude string `yaml:"printer_exclude"`
}

var ConfigDefaults = Config{
	Include: ".+",
	Exclude: "",
}

type collector struct {
	logger log.Logger

	printerInclude *string
	printerExclude *string

	printerStatus    *prometheus.Desc
	printerJobStatus *prometheus.Desc
	printerJobCount  *prometheus.Desc

	printerIncludePattern *regexp.Regexp
	printerExcludePattern *regexp.Regexp
}

func New(logger log.Logger, config *Config) types.Collector {
	if config == nil {
		config = &ConfigDefaults
	}
	c := &collector{
		printerInclude: &config.Include,
		printerExclude: &config.Exclude,
	}
	c.SetLogger(logger)
	return c
}

func NewWithFlags(app *kingpin.Application) types.Collector {
	c := &collector{
		printerInclude: app.Flag(
			FlagPrinterInclude,
			"Regular expression to match printers to collect metrics for",
		).Default(ConfigDefaults.Include).String(),
		printerExclude: app.Flag(
			FlagPrinterExclude,
			"Regular expression to match printers to exclude",
		).Default(ConfigDefaults.Exclude).String(),
	}
	return c
}

func (c *collector) SetLogger(logger log.Logger) {
	c.logger = log.With(logger, "collector", Name)
}

func (c *collector) Build() error {
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

	var err error
	c.printerIncludePattern, err = regexp.Compile(fmt.Sprintf("^(?:%s)$", *c.printerInclude))
	if err != nil {
		return err
	}
	c.printerExcludePattern, err = regexp.Compile(fmt.Sprintf("^(?:%s)$", *c.printerExclude))
	return err
}

func (c *collector) GetName() string { return Name }

func (c *collector) GetPerfCounter() ([]string, error) { return []string{"Printer"}, nil }

type win32_Printer struct {
	Name                   string
	Default                bool
	PrinterStatus          uint16
	JobCountSinceLastReset uint32
}

type win32_PrintJob struct {
	Name   string
	Status string
}

func (c *collector) Collect(_ *types.ScrapeContext, ch chan<- prometheus.Metric) error {
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

func (c *collector) collectPrinterStatus(ch chan<- prometheus.Metric) error {
	var printers []win32_Printer
	q := wmi.QueryAll(&printers, c.logger)
	if err := wmi.Query(q, &printers); err != nil {
		return err
	}

	for _, printer := range printers {
		if c.printerExcludePattern.MatchString(printer.Name) ||
			!c.printerIncludePattern.MatchString(printer.Name) {
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

func (c *collector) collectPrinterJobStatus(ch chan<- prometheus.Metric) error {
	var printJobs []win32_PrintJob
	q := wmi.QueryAll(&printJobs, c.logger)
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

func (c *collector) groupPrintJobs(printJobs []win32_PrintJob) map[PrintJobStatusGroup]int {
	groupedPrintJobs := make(map[PrintJobStatusGroup]int)
	for _, printJob := range printJobs {
		printerName := strings.Split(printJob.Name, ",")[0]

		if c.printerExcludePattern.MatchString(printerName) ||
			!c.printerIncludePattern.MatchString(printerName) {
			continue
		}
		groupedPrintJobs[PrintJobStatusGroup{
			printerName: printerName,
			status:      printJob.Status,
		}]++
	}
	return groupedPrintJobs
}
