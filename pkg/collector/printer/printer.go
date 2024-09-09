//go:build windows

package printer

import (
	"errors"
	"fmt"
	"log/slog"
	"regexp"
	"strings"

	"github.com/alecthomas/kingpin/v2"
	"github.com/prometheus-community/windows_exporter/pkg/types"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/yusufpapurcu/wmi"
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
	config    Config
	wmiClient *wmi.Client

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

func (c *Collector) Close(_ *slog.Logger) error {
	return nil
}

func (c *Collector) Build(_ *slog.Logger, wmiClient *wmi.Client) error {
	if wmiClient == nil || wmiClient.SWbemServicesClient == nil {
		return errors.New("wmiClient or SWbemServicesClient is nil")
	}

	c.wmiClient = wmiClient

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

func (c *Collector) GetPerfCounter(_ *slog.Logger) ([]string, error) {
	return []string{"Printer"}, nil
}

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

func (c *Collector) Collect(_ *types.ScrapeContext, logger *slog.Logger, ch chan<- prometheus.Metric) error {
	logger = logger.With(slog.String("collector", Name))
	if err := c.collectPrinterStatus(ch); err != nil {
		logger.Error("failed to collect printer status metrics",
			slog.Any("err", err),
		)

		return err
	}

	if err := c.collectPrinterJobStatus(ch); err != nil {
		logger.Error("failed to collect printer job status metrics",
			slog.Any("err", err),
		)

		return err
	}

	return nil
}

func (c *Collector) collectPrinterStatus(ch chan<- prometheus.Metric) error {
	var printers []wmiPrinter
	if err := c.wmiClient.Query("SELECT * FROM win32_Printer", &printers); err != nil {
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
	if err := c.wmiClient.Query("SELECT * FROM win32_PrintJob", &printJobs); err != nil {
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
