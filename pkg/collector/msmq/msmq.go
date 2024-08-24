//go:build windows

package msmq

import (
	"strings"

	"github.com/alecthomas/kingpin/v2"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/prometheus-community/windows_exporter/pkg/types"
	"github.com/prometheus-community/windows_exporter/pkg/utils"
	"github.com/prometheus-community/windows_exporter/pkg/wmi"
	"github.com/prometheus/client_golang/prometheus"
)

const Name = "msmq"

type Config struct {
	QueryWhereClause *string `yaml:"query_where_clause"`
}

var ConfigDefaults = Config{
	QueryWhereClause: utils.ToPTR(""),
}

// A Collector is a Prometheus Collector for WMI Win32_PerfRawData_MSMQ_MSMQQueue metrics.
type Collector struct {
	config Config

	bytesInJournalQueue    *prometheus.Desc
	bytesInQueue           *prometheus.Desc
	messagesInJournalQueue *prometheus.Desc
	messagesInQueue        *prometheus.Desc
}

func New(config *Config) *Collector {
	if config == nil {
		config = &ConfigDefaults
	}

	if config.QueryWhereClause == nil {
		config.QueryWhereClause = ConfigDefaults.QueryWhereClause
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

	app.Flag("collector.msmq.msmq-where", "WQL 'where' clause to use in WMI metrics query. "+
		"Limits the response to the msmqs you specify and reduces the size of the response.").
		Default(*c.config.QueryWhereClause).StringVar(c.config.QueryWhereClause)

	return c
}

func (c *Collector) GetName() string {
	return Name
}

func (c *Collector) GetPerfCounter(_ log.Logger) ([]string, error) {
	return []string{}, nil
}

func (c *Collector) Close() error {
	return nil
}

func (c *Collector) Build(logger log.Logger) error {
	logger = log.With(logger, "collector", Name)

	if *c.config.QueryWhereClause == "" {
		_ = level.Warn(logger).Log("msg", "No where-clause specified for msmq collector. This will generate a very large number of metrics!")
	}

	c.bytesInJournalQueue = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "bytes_in_journal_queue"),
		"Size of queue journal in bytes",
		[]string{"name"},
		nil,
	)
	c.bytesInQueue = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "bytes_in_queue"),
		"Size of queue in bytes",
		[]string{"name"},
		nil,
	)
	c.messagesInJournalQueue = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "messages_in_journal_queue"),
		"Count messages in queue journal",
		[]string{"name"},
		nil,
	)
	c.messagesInQueue = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "messages_in_queue"),
		"Count messages in queue",
		[]string{"name"},
		nil,
	)
	return nil
}

// Collect sends the metric values for each metric
// to the provided prometheus Metric channel.
func (c *Collector) Collect(_ *types.ScrapeContext, logger log.Logger, ch chan<- prometheus.Metric) error {
	logger = log.With(logger, "collector", Name)
	if err := c.collect(logger, ch); err != nil {
		_ = level.Error(logger).Log("msg", "failed collecting msmq metrics", "err", err)
		return err
	}
	return nil
}

type msmqQueue struct {
	Name string

	BytesInJournalQueue    uint64
	BytesInQueue           uint64
	MessagesInJournalQueue uint64
	MessagesInQueue        uint64
}

func (c *Collector) collect(logger log.Logger, ch chan<- prometheus.Metric) error {
	var dst []msmqQueue

	q := wmi.QueryAllForClassWhere(&dst, "Win32_PerfRawData_MSMQ_MSMQQueue", *c.config.QueryWhereClause, logger)
	if err := wmi.Query(q, &dst); err != nil {
		return err
	}

	for _, msmq := range dst {
		ch <- prometheus.MustNewConstMetric(
			c.bytesInJournalQueue,
			prometheus.GaugeValue,
			float64(msmq.BytesInJournalQueue),
			strings.ToLower(msmq.Name),
		)

		ch <- prometheus.MustNewConstMetric(
			c.bytesInQueue,
			prometheus.GaugeValue,
			float64(msmq.BytesInQueue),
			strings.ToLower(msmq.Name),
		)

		ch <- prometheus.MustNewConstMetric(
			c.messagesInJournalQueue,
			prometheus.GaugeValue,
			float64(msmq.MessagesInJournalQueue),
			strings.ToLower(msmq.Name),
		)

		ch <- prometheus.MustNewConstMetric(
			c.messagesInQueue,
			prometheus.GaugeValue,
			float64(msmq.MessagesInQueue),
			strings.ToLower(msmq.Name),
		)
	}

	return nil
}
