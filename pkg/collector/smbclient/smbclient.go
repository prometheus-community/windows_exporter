//go:build windows

package smbclient

import (
	"fmt"
	"os"
	"slices"
	"strings"

	"github.com/alecthomas/kingpin/v2"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/prometheus-community/windows_exporter/pkg/perflib"
	"github.com/prometheus-community/windows_exporter/pkg/types"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	Name = "smbclient"
)

type Config struct {
	CollectorsEnabled string `yaml:"collectors_enabled"`
}

var ConfigDefaults = Config{
	CollectorsEnabled: "",
}

type Collector struct {
	logger log.Logger

	enabledCollectors []string

	smbClientListAllCollectors *bool
	smbClientCollectorsEnabled *string

	readBytesTotal                            *prometheus.Desc
	readBytesTransmittedViaSMBDirectTotal     *prometheus.Desc
	readRequestQueueSecsTotal                 *prometheus.Desc
	readRequestsTransmittedViaSMBDirectTotal  *prometheus.Desc
	readSecsTotal                             *prometheus.Desc
	readsTotal                                *prometheus.Desc
	turboIOReadsTotal                         *prometheus.Desc
	TurboIOWritesTotal                        *prometheus.Desc
	writeBytesTotal                           *prometheus.Desc
	writeBytesTransmittedViaSMBDirectTotal    *prometheus.Desc
	writeRequestQueueSecsTotal                *prometheus.Desc
	writeRequestsTransmittedViaSMBDirectTotal *prometheus.Desc
	writeSecsTotal                            *prometheus.Desc
	writesTotal                               *prometheus.Desc

	creditStallsTotal     *prometheus.Desc
	currentDataQueued     *prometheus.Desc
	dataBytesTotal        *prometheus.Desc
	dataRequestsTotal     *prometheus.Desc
	metadataRequestsTotal *prometheus.Desc
	requestQueueSecsTotal *prometheus.Desc
	requestSecs           *prometheus.Desc
}

// All available collector functions
var smbclientAllCollectorNames = []string{
	"ClientShares",
}

func New(logger log.Logger, config *Config) *Collector {
	if config == nil {
		config = &ConfigDefaults
	}

	smbclientListAllCollectors := false
	c := &Collector{
		smbClientCollectorsEnabled: &config.CollectorsEnabled,
		smbClientListAllCollectors: &smbclientListAllCollectors,
	}
	c.SetLogger(logger)

	return c
}

func NewWithFlags(app *kingpin.Application) *Collector {
	return &Collector{
		smbClientListAllCollectors: app.Flag(
			"collectors.smbclient.list",
			"List the collectors along with their perflib object name/ids",
		).Bool(),

		smbClientCollectorsEnabled: app.Flag(
			"collectors.smbclient.enabled",
			"Comma-separated list of collectors to use. Defaults to all, if not specified.",
		).Default(ConfigDefaults.CollectorsEnabled).String(),
	}
}

func (c *Collector) GetName() string {
	return Name
}

func (c *Collector) SetLogger(logger log.Logger) {
	c.logger = log.With(logger, "collector", Name)
}

func (c *Collector) GetPerfCounter() ([]string, error) {
	return []string{
		"SMB Client Shares",
	}, nil
}

func (c *Collector) Close() error {
	return nil
}

func (c *Collector) Build() error {
	// desc creates a new prometheus description
	desc := func(metricName string, description string, labels []string) *prometheus.Desc {
		return prometheus.NewDesc(
			prometheus.BuildFQName(types.Namespace, "smbclient", metricName),
			description,
			labels,
			nil,
		)
	}

	c.requestQueueSecsTotal = desc("data_queue_seconds_total",
		"Seconds requests waited on queue on this share",
		[]string{"server", "share"},
	)
	c.readRequestQueueSecsTotal = desc("read_queue_seconds_total",
		"Seconds read requests waited on queue on this share",
		[]string{"server", "share"},
	)
	c.writeRequestQueueSecsTotal = desc("write_queue_seconds_total",
		"Seconds write requests waited on queue on this share",
		[]string{"server", "share"},
	)
	c.requestSecs = desc("request_seconds_total",
		"Seconds waiting for requests on this share",
		[]string{"server", "share"},
	)
	c.creditStallsTotal = desc("stalls_total",
		"The number of requests delayed based on insufficient credits on this share",
		[]string{"server", "share"},
	)
	c.currentDataQueued = desc("requests_queued",
		"The point in time number of requests outstanding on this share",
		[]string{"server", "share"},
	)
	c.dataBytesTotal = desc("data_bytes_total",
		"The bytes read or written on this share",
		[]string{"server", "share"},
	)
	c.dataRequestsTotal = desc("requests_total",
		"The requests on this share",
		[]string{"server", "share"},
	)
	c.metadataRequestsTotal = desc("metadata_requests_total",
		"The metadata requests on this share",
		[]string{"server", "share"},
	)
	c.readBytesTransmittedViaSMBDirectTotal = desc("read_bytes_via_smbdirect_total",
		"The bytes read from this share via RDMA direct placement",
		[]string{"server", "share"},
	)
	c.readBytesTotal = desc("read_bytes_total",
		"The bytes read on this share",
		[]string{"server", "share"},
	)
	c.readRequestsTransmittedViaSMBDirectTotal = desc("read_requests_via_smbdirect_total",
		"The read requests on this share via RDMA direct placement",
		[]string{"server", "share"},
	)
	c.readsTotal = desc("read_requests_total",
		"The read requests on this share",
		[]string{"server", "share"},
	)
	c.turboIOReadsTotal = desc("turbo_io_reads_total",
		"The read requests that go through Turbo I/O",
		[]string{"server", "share"},
	)
	c.TurboIOWritesTotal = desc("turbo_io_writes_total",
		"The write requests that go through Turbo I/O",
		[]string{"server", "share"},
	)
	c.writeBytesTransmittedViaSMBDirectTotal = desc("write_bytes_via_smbdirect_total",
		"The written bytes to this share via RDMA direct placement",
		[]string{"server", "share"},
	)
	c.writeBytesTotal = desc("write_bytes_total",
		"The bytes written on this share",
		[]string{"server", "share"},
	)
	c.writeRequestsTransmittedViaSMBDirectTotal = desc("write_requests_via_smbdirect_total",
		"The write requests to this share via RDMA direct placement",
		[]string{"server", "share"},
	)
	c.writesTotal = desc("write_requests_total",
		"The write requests on this share",
		[]string{"server", "share"},
	)
	c.readSecsTotal = desc("read_seconds_total",
		"Seconds waiting for read requests on this share",
		[]string{"server", "share"},
	)
	c.writeSecsTotal = desc("write_seconds_total",
		"Seconds waiting for write requests on this share",
		[]string{"server", "share"},
	)

	c.enabledCollectors = make([]string, 0, len(smbclientAllCollectorNames))

	collectorDesc := map[string]string{
		"ClientShares": "SMB Client Shares",
	}

	if *c.smbClientListAllCollectors {
		fmt.Printf("%-32s %-32s\n", "Collector Name", "Perflib Object") //nolint:forbidigo
		for _, cname := range smbclientAllCollectorNames {
			fmt.Printf("%-32s %-32s\n", cname, collectorDesc[cname]) //nolint:forbidigo
		}

		os.Exit(0)
	}

	if *c.smbClientCollectorsEnabled == "" {
		for _, collectorName := range smbclientAllCollectorNames {
			c.enabledCollectors = append(c.enabledCollectors, collectorName)
		}
	} else {
		for _, collectorName := range strings.Split(*c.smbClientCollectorsEnabled, ",") {
			if slices.Contains(smbclientAllCollectorNames, collectorName) {
				c.enabledCollectors = append(c.enabledCollectors, collectorName)
			} else {
				return fmt.Errorf("unknown smbclient Collector: %s", collectorName)
			}
		}
	}

	return nil
}

// Collect collects smb client metrics and sends them to prometheus
func (c *Collector) Collect(ctx *types.ScrapeContext, ch chan<- prometheus.Metric) error {
	collectorFuncs := map[string]func(ctx *types.ScrapeContext, ch chan<- prometheus.Metric) error{
		"ClientShares": c.collectClientShares,
	}

	for _, collectorName := range c.enabledCollectors {
		if err := collectorFuncs[collectorName](ctx, ch); err != nil {
			_ = level.Error(c.logger).Log("msg", "Error in "+collectorName, "err", err)
			return err
		}
	}
	return nil
}

// Perflib: SMB Client Shares
type perflibClientShares struct {
	Name string

	AvgDataQueueLength                         float64 `perflib:"Avg. Data Queue Length"`
	AvgReadQueueLength                         float64 `perflib:"Avg. Read Queue Length"`
	AvgSecPerRead                              float64 `perflib:"Avg. sec/Read"`
	AvgSecPerWrite                             float64 `perflib:"Avg. sec/Write"`
	AvgSecPerDataRequest                       float64 `perflib:"Avg. sec/Data Request"`
	AvgWriteQueueLength                        float64 `perflib:"Avg. Write Queue Length"`
	CreditStallsPerSec                         float64 `perflib:"Credit Stalls/sec"`
	CurrentDataQueueLength                     float64 `perflib:"Current Data Queue Length"`
	DataBytesPerSec                            float64 `perflib:"Data Bytes/sec"`
	DataRequestsPerSec                         float64 `perflib:"Data Requests/sec"`
	MetadataRequestsPerSec                     float64 `perflib:"Metadata Requests/sec"`
	ReadBytesTransmittedViaSMBDirectPerSec     float64 `perflib:"Read Bytes transmitted via SMB Direct/sec"`
	ReadBytesPerSec                            float64 `perflib:"Read Bytes/sec"`
	ReadRequestsTransmittedViaSMBDirectPerSec  float64 `perflib:"Read Requests transmitted via SMB Direct/sec"`
	ReadRequestsPerSec                         float64 `perflib:"Read Requests/sec"`
	TurboIOReadsPerSec                         float64 `perflib:"Turbo I/O Reads/sec"`
	TurboIOWritesPerSec                        float64 `perflib:"Turbo I/O Writes/sec"`
	WriteBytesTransmittedViaSMBDirectPerSec    float64 `perflib:"Write Bytes transmitted via SMB Direct/sec"`
	WriteBytesPerSec                           float64 `perflib:"Write Bytes/sec"`
	WriteRequestsTransmittedViaSMBDirectPerSec float64 `perflib:"Write Requests transmitted via SMB Direct/sec"`
	WriteRequestsPerSec                        float64 `perflib:"Write Requests/sec"`
}

func (c *Collector) collectClientShares(ctx *types.ScrapeContext, ch chan<- prometheus.Metric) error {
	var data []perflibClientShares
	if err := perflib.UnmarshalObject(ctx.PerfObjects["SMB Client Shares"], &data, c.logger); err != nil {
		return err
	}
	for _, instance := range data {
		if instance.Name == "_Total" {
			continue
		}

		parsed := strings.FieldsFunc(instance.Name, func(r rune) bool { return r == '\\' })
		serverValue := parsed[0]
		shareValue := parsed[1]
		// Request time spent on queue. Convert from ticks to seconds.
		ch <- prometheus.MustNewConstMetric(
			c.requestQueueSecsTotal,
			prometheus.CounterValue,
			instance.AvgDataQueueLength*perflib.TicksToSecondScaleFactor,
			serverValue, shareValue,
		)

		// Read time spent on queue. Convert from ticks to seconds.
		ch <- prometheus.MustNewConstMetric(
			c.readRequestQueueSecsTotal,
			prometheus.CounterValue,
			instance.AvgReadQueueLength*perflib.TicksToSecondScaleFactor,
			serverValue, shareValue,
		)

		ch <- prometheus.MustNewConstMetric(
			c.readSecsTotal,
			prometheus.CounterValue,
			instance.AvgSecPerRead*perflib.TicksToSecondScaleFactor,
			serverValue, shareValue,
		)

		ch <- prometheus.MustNewConstMetric(
			c.writeSecsTotal,
			prometheus.CounterValue,
			instance.AvgSecPerWrite*perflib.TicksToSecondScaleFactor,
			serverValue, shareValue,
		)

		ch <- prometheus.MustNewConstMetric(
			c.requestSecs,
			prometheus.CounterValue,
			instance.AvgSecPerDataRequest*perflib.TicksToSecondScaleFactor,
			serverValue, shareValue,
		)

		// Write time spent on queue. Convert from  ticks to seconds.
		ch <- prometheus.MustNewConstMetric(
			c.writeRequestQueueSecsTotal,
			prometheus.CounterValue,
			instance.AvgWriteQueueLength*perflib.TicksToSecondScaleFactor,
			serverValue, shareValue,
		)

		ch <- prometheus.MustNewConstMetric(
			c.creditStallsTotal,
			prometheus.CounterValue,
			instance.CreditStallsPerSec,
			serverValue, shareValue,
		)

		ch <- prometheus.MustNewConstMetric(
			c.currentDataQueued,
			prometheus.GaugeValue,
			instance.CurrentDataQueueLength,
			serverValue, shareValue,
		)

		ch <- prometheus.MustNewConstMetric(
			c.dataBytesTotal,
			prometheus.CounterValue,
			instance.DataBytesPerSec,
			serverValue, shareValue,
		)

		ch <- prometheus.MustNewConstMetric(
			c.dataRequestsTotal,
			prometheus.CounterValue,
			instance.DataRequestsPerSec,
			serverValue, shareValue,
		)

		ch <- prometheus.MustNewConstMetric(
			c.metadataRequestsTotal,
			prometheus.CounterValue,
			instance.MetadataRequestsPerSec,
			serverValue, shareValue,
		)

		ch <- prometheus.MustNewConstMetric(
			c.readBytesTransmittedViaSMBDirectTotal,
			prometheus.CounterValue,
			instance.ReadBytesTransmittedViaSMBDirectPerSec,
			serverValue, shareValue,
		)

		ch <- prometheus.MustNewConstMetric(
			c.readBytesTotal,
			prometheus.CounterValue,
			instance.ReadBytesPerSec,
			serverValue, shareValue,
		)

		ch <- prometheus.MustNewConstMetric(
			c.readRequestsTransmittedViaSMBDirectTotal,
			prometheus.CounterValue,
			instance.ReadRequestsTransmittedViaSMBDirectPerSec,
			serverValue, shareValue,
		)

		ch <- prometheus.MustNewConstMetric(
			c.readsTotal,
			prometheus.CounterValue,
			instance.ReadRequestsPerSec,
			serverValue, shareValue,
		)

		ch <- prometheus.MustNewConstMetric(
			c.turboIOReadsTotal,
			prometheus.CounterValue,
			instance.TurboIOReadsPerSec,
			serverValue, shareValue,
		)

		ch <- prometheus.MustNewConstMetric(
			c.TurboIOWritesTotal,
			prometheus.CounterValue,
			instance.TurboIOWritesPerSec,
			serverValue, shareValue,
		)

		ch <- prometheus.MustNewConstMetric(
			c.writeBytesTransmittedViaSMBDirectTotal,
			prometheus.CounterValue,
			instance.WriteBytesTransmittedViaSMBDirectPerSec,
			serverValue, shareValue,
		)

		ch <- prometheus.MustNewConstMetric(
			c.writeBytesTotal,
			prometheus.CounterValue,
			instance.WriteBytesPerSec,
			serverValue, shareValue,
		)

		ch <- prometheus.MustNewConstMetric(
			c.writeRequestsTransmittedViaSMBDirectTotal,
			prometheus.CounterValue,
			instance.WriteRequestsTransmittedViaSMBDirectPerSec,
			serverValue, shareValue,
		)

		ch <- prometheus.MustNewConstMetric(
			c.writesTotal,
			prometheus.CounterValue,
			instance.WriteRequestsPerSec,
			serverValue, shareValue,
		)
	}
	return nil
}
