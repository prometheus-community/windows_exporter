//go:build windows

package msmq

import (
	"errors"
	"log/slog"
	"strings"

	"github.com/alecthomas/kingpin/v2"
	"github.com/prometheus-community/windows_exporter/pkg/types"
	"github.com/prometheus-community/windows_exporter/pkg/utils"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/yusufpapurcu/wmi"
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
	config    Config
	wmiClient *wmi.Client

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

func (c *Collector) GetPerfCounter(_ *slog.Logger) ([]string, error) {
	return []string{}, nil
}

func (c *Collector) Close(_ *slog.Logger) error {
	return nil
}

func (c *Collector) Build(logger *slog.Logger, wmiClient *wmi.Client) error {
	logger = logger.With(slog.String("collector", Name))

	if wmiClient == nil || wmiClient.SWbemServicesClient == nil {
		return errors.New("wmiClient or SWbemServicesClient is nil")
	}

	c.wmiClient = wmiClient

	if *c.config.QueryWhereClause == "" {
		logger.Warn("No where-clause specified for msmq collector. This will generate a very large number of metrics!")
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
func (c *Collector) Collect(_ *types.ScrapeContext, logger *slog.Logger, ch chan<- prometheus.Metric) error {
	logger = logger.With(slog.String("collector", Name))
	if err := c.collect(ch); err != nil {
		logger.Error("failed collecting msmq metrics",
			slog.Any("err", err),
		)

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

func (c *Collector) collect(ch chan<- prometheus.Metric) error {
	var dst []msmqQueue

	query := "SELECT * FROM Win32_PerfRawData_MSMQ_MSMQQueue"
	if *c.config.QueryWhereClause != "" {
		query += " WHERE " + *c.config.QueryWhereClause
	}

	if err := c.wmiClient.Query(query, &dst); err != nil {
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
