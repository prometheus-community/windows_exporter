package printer

import (
	"fmt"
	"github.com/alecthomas/kingpin/v2"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/prometheus-community/windows_exporter/pkg/types"
	"github.com/prometheus-community/windows_exporter/pkg/wmi"
	"github.com/prometheus/client_golang/prometheus"
	"regexp"
	"strings"
)

const (
	Name = "printer"

	FlagPrinterInclude = "collector.printer.include"
	FlagPrinterExclude = "collector.printer.exclude"
)

type Config struct {
	printerInclude string `yaml:"printer_include"`
	printerExclude string `yaml:"printer_exclude"`
}

var ConfigDefaults = Config{
	printerInclude: ".+",
	printerExclude: "",
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
		printerInclude: &config.printerInclude,
		printerExclude: &config.printerExclude,
	}
	c.SetLogger(logger)
	return c
}

func NewWithFlags(app *kingpin.Application) types.Collector {
	c := &collector{
		printerInclude: app.Flag(
			FlagPrinterInclude,
			"Regular expression to match printers to collectPrinterStatus",
		).Default(ConfigDefaults.printerInclude).String(),
		printerExclude: app.Flag(
			FlagPrinterExclude,
			"Regular expression to match printers to exclude",
		).Default(ConfigDefaults.printerExclude).String(),
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

func (c *collector) GetPerfCounter() ([]string, error) { return []string{}, nil }

type win32_Printer struct {
	Name                   string
	Default                bool
	Status                 string
	JobCountSinceLastReset uint32
}

type win32_PrintJob struct {
	Name   string
	Status string
}

func (c *collector) Collect(ctx *types.ScrapeContext, ch chan<- prometheus.Metric) error {
	if desc, err := c.collectPrinterStatus(ch); err != nil {
		_ = level.Error(c.logger).Log("msg", "failed to collect printer status metrics", "desc", desc, "err", err)
		return err
	}
	if desc, err := c.collectPrinterJobStatus(ch); err != nil {
		_ = level.Error(c.logger).Log("msg", "failed to collect printer job status metrics", "desc", desc, "err", err)
		return err
	}
	return nil
}

func (c *collector) collectPrinterStatus(ch chan<- prometheus.Metric) (*prometheus.Desc, error) {
	var printers []win32_Printer
	q := wmi.QueryAll(&printers, c.logger)
	if err := wmi.Query(q, &printers); err != nil {
		return nil, err
	}

	for _, printer := range printers {
		if c.printerExcludePattern.MatchString(printer.Name) ||
			!c.printerIncludePattern.MatchString(printer.Name) {
			continue
		}
		ch <- prometheus.MustNewConstMetric(
			c.printerStatus,
			prometheus.GaugeValue,
			1,
			printer.Name,
			printer.Status,
		)
		ch <- prometheus.MustNewConstMetric(
			c.printerJobCount,
			prometheus.GaugeValue,
			float64(printer.JobCountSinceLastReset),
			printer.Name,
		)
	}

	return nil, nil
}

func (c *collector) collectPrinterJobStatus(ch chan<- prometheus.Metric) (*prometheus.Desc, error) {
	var printJobs []win32_PrintJob
	q := wmi.QueryAll(&printJobs, c.logger)
	if err := wmi.Query(q, &printJobs); err != nil {
		return nil, err
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
	return nil, nil
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
