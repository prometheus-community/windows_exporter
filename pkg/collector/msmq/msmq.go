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
	QueryWhereClause string `yaml:"query_where_clause"`
}

var ConfigDefaults = Config{
	QueryWhereClause: "",
}

// A Collector is a Prometheus Collector for WMI Win32_PerfRawData_MSMQ_MSMQQueue metrics
type Collector struct {
	logger log.Logger

	queryWhereClause *string

	BytesinJournalQueue    *prometheus.Desc
	BytesinQueue           *prometheus.Desc
	MessagesinJournalQueue *prometheus.Desc
	MessagesinQueue        *prometheus.Desc
}

func New(logger log.Logger, config *Config) *Collector {
	if config == nil {
		config = &ConfigDefaults
	}

	c := &Collector{
		queryWhereClause: &config.QueryWhereClause,
	}
	c.SetLogger(logger)

	return c
}

func NewWithFlags(app *kingpin.Application) *Collector {
	return &Collector{
		queryWhereClause: app.
			Flag("collector.msmq.msmq-where", "WQL 'where' clause to use in WMI metrics query. Limits the response to the msmqs you specify and reduces the size of the response.").
			Default(ConfigDefaults.QueryWhereClause).String(),
	}
}

func (c *Collector) GetName() string {
	return Name
}

func (c *Collector) SetLogger(logger log.Logger) {
	c.logger = log.With(logger, "collector", Name)
}

func (c *Collector) GetPerfCounter() ([]string, error) {
	return []string{}, nil
}

func (c *Collector) Close() error {
	return nil
}

func (c *Collector) Build() error {
	if utils.IsEmpty(c.queryWhereClause) {
		_ = level.Warn(c.logger).Log("msg", "No where-clause specified for msmq collector. This will generate a very large number of metrics!")
	}

	c.BytesinJournalQueue = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "bytes_in_journal_queue"),
		"Size of queue journal in bytes",
		[]string{"name"},
		nil,
	)
	c.BytesinQueue = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "bytes_in_queue"),
		"Size of queue in bytes",
		[]string{"name"},
		nil,
	)
	c.MessagesinJournalQueue = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "messages_in_journal_queue"),
		"Count messages in queue journal",
		[]string{"name"},
		nil,
	)
	c.MessagesinQueue = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "messages_in_queue"),
		"Count messages in queue",
		[]string{"name"},
		nil,
	)
	return nil
}

// Collect sends the metric values for each metric
// to the provided prometheus Metric channel.
func (c *Collector) Collect(_ *types.ScrapeContext, ch chan<- prometheus.Metric) error {
	if err := c.collect(ch); err != nil {
		_ = level.Error(c.logger).Log("msg", "failed collecting msmq metrics", "err", err)
		return err
	}
	return nil
}

type Win32_PerfRawData_MSMQ_MSMQQueue struct {
	Name string

	BytesinJournalQueue    uint64
	BytesinQueue           uint64
	MessagesinJournalQueue uint64
	MessagesinQueue        uint64
}

func (c *Collector) collect(ch chan<- prometheus.Metric) error {
	var dst []Win32_PerfRawData_MSMQ_MSMQQueue
	q := wmi.QueryAllWhere(&dst, *c.queryWhereClause, c.logger)
	if err := wmi.Query(q, &dst); err != nil {
		return err
	}

	for _, msmq := range dst {
		ch <- prometheus.MustNewConstMetric(
			c.BytesinJournalQueue,
			prometheus.GaugeValue,
			float64(msmq.BytesinJournalQueue),
			strings.ToLower(msmq.Name),
		)

		ch <- prometheus.MustNewConstMetric(
			c.BytesinQueue,
			prometheus.GaugeValue,
			float64(msmq.BytesinQueue),
			strings.ToLower(msmq.Name),
		)

		ch <- prometheus.MustNewConstMetric(
			c.MessagesinJournalQueue,
			prometheus.GaugeValue,
			float64(msmq.MessagesinJournalQueue),
			strings.ToLower(msmq.Name),
		)

		ch <- prometheus.MustNewConstMetric(
			c.MessagesinQueue,
			prometheus.GaugeValue,
			float64(msmq.MessagesinQueue),
			strings.ToLower(msmq.Name),
		)
	}
	return nil
}
