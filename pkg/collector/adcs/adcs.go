//go:build windows

package adcs

import (
	"errors"
	"fmt"
	"log/slog"

	"github.com/alecthomas/kingpin/v2"
	"github.com/prometheus-community/windows_exporter/pkg/perfdata"
	"github.com/prometheus-community/windows_exporter/pkg/perflib"
	"github.com/prometheus-community/windows_exporter/pkg/types"
	"github.com/prometheus-community/windows_exporter/pkg/utils"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/yusufpapurcu/wmi"
)

const Name = "adcs"

type Config struct{}

var ConfigDefaults = Config{}

type Collector struct {
	config Config

	perfDataCollector *perfdata.Collector

	challengeResponseProcessingTime              *prometheus.Desc
	challengeResponsesPerSecond                  *prometheus.Desc
	failedRequestsPerSecond                      *prometheus.Desc
	issuedRequestsPerSecond                      *prometheus.Desc
	pendingRequestsPerSecond                     *prometheus.Desc
	requestCryptographicSigningTime              *prometheus.Desc
	requestPolicyModuleProcessingTime            *prometheus.Desc
	requestProcessingTime                        *prometheus.Desc
	requestsPerSecond                            *prometheus.Desc
	retrievalProcessingTime                      *prometheus.Desc
	retrievalsPerSecond                          *prometheus.Desc
	signedCertificateTimestampListProcessingTime *prometheus.Desc
	signedCertificateTimestampListsPerSecond     *prometheus.Desc
}

func New(config *Config) *Collector {
	if config == nil {
		config = &ConfigDefaults
	}

	c := &Collector{
		config: *config,
	}

	return c
}

func NewWithFlags(_ *kingpin.Application) *Collector {
	return &Collector{}
}

func (c *Collector) GetName() string {
	return Name
}

func (c *Collector) GetPerfCounter(_ *slog.Logger) ([]string, error) {
	if utils.PDHEnabled() {
		return []string{}, nil
	}

	return []string{"Certification Authority"}, nil
}

func (c *Collector) Close(_ *slog.Logger) error {
	return nil
}

func (c *Collector) Build(_ *slog.Logger, _ *wmi.Client) error {
	if utils.PDHEnabled() {
		counters := []string{
			RequestsPerSecond,
			RequestProcessingTime,
			RetrievalsPerSecond,
			RetrievalProcessingTime,
			FailedRequestsPerSecond,
			IssuedRequestsPerSecond,
			PendingRequestsPerSecond,
			RequestCryptographicSigningTime,
			RequestPolicyModuleProcessingTime,
			ChallengeResponsesPerSecond,
			ChallengeResponseProcessingTime,
			SignedCertificateTimestampListsPerSecond,
			SignedCertificateTimestampListProcessingTime,
		}

		var err error

		c.perfDataCollector, err = perfdata.NewCollector("Processor Information", []string{"*"}, counters)
		if err != nil {
			return fmt.Errorf("failed to create Processor Information collector: %w", err)
		}
	}

	c.requestsPerSecond = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "requests_total"),
		"Total certificate requests processed",
		[]string{"cert_template"},
		nil,
	)
	c.requestProcessingTime = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "request_processing_time_seconds"),
		"Last time elapsed for certificate requests",
		[]string{"cert_template"},
		nil,
	)
	c.retrievalsPerSecond = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "retrievals_total"),
		"Total certificate retrieval requests processed",
		[]string{"cert_template"},
		nil,
	)
	c.retrievalProcessingTime = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "retrievals_processing_time_seconds"),
		"Last time elapsed for certificate retrieval request",
		[]string{"cert_template"},
		nil,
	)
	c.failedRequestsPerSecond = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "failed_requests_total"),
		"Total failed certificate requests processed",
		[]string{"cert_template"},
		nil,
	)
	c.issuedRequestsPerSecond = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "issued_requests_total"),
		"Total issued certificate requests processed",
		[]string{"cert_template"},
		nil,
	)
	c.pendingRequestsPerSecond = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "pending_requests_total"),
		"Total pending certificate requests processed",
		[]string{"cert_template"},
		nil,
	)
	c.requestCryptographicSigningTime = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "request_cryptographic_signing_time_seconds"),
		"Last time elapsed for signing operation request",
		[]string{"cert_template"},
		nil,
	)
	c.requestPolicyModuleProcessingTime = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "request_policy_module_processing_time_seconds"),
		"Last time elapsed for policy module processing request",
		[]string{"cert_template"},
		nil,
	)
	c.challengeResponsesPerSecond = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "challenge_responses_total"),
		"Total certificate challenge responses processed",
		[]string{"cert_template"},
		nil,
	)
	c.challengeResponseProcessingTime = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "challenge_response_processing_time_seconds"),
		"Last time elapsed for challenge response",
		[]string{"cert_template"},
		nil,
	)
	c.signedCertificateTimestampListsPerSecond = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "signed_certificate_timestamp_lists_total"),
		"Total Signed Certificate Timestamp Lists processed",
		[]string{"cert_template"},
		nil,
	)
	c.signedCertificateTimestampListProcessingTime = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "signed_certificate_timestamp_list_processing_time_seconds"),
		"Last time elapsed for Signed Certificate Timestamp List",
		[]string{"cert_template"},
		nil,
	)

	return nil
}

func (c *Collector) Collect(ctx *types.ScrapeContext, logger *slog.Logger, ch chan<- prometheus.Metric) error {
	if utils.PDHEnabled() {
		return c.collectPDH(ch)
	}

	logger = logger.With(slog.String("collector", Name))
	if err := c.collectADCSCounters(ctx, logger, ch); err != nil {
		logger.Error("failed collecting ADCS metrics",
			slog.Any("err", err),
		)

		return err
	}

	return nil
}

func (c *Collector) collectADCSCounters(ctx *types.ScrapeContext, logger *slog.Logger, ch chan<- prometheus.Metric) error {
	dst := make([]perflibADCS, 0)

	if _, ok := ctx.PerfObjects["Certification Authority"]; !ok {
		return errors.New("perflib did not contain an entry for Certification Authority")
	}

	err := perflib.UnmarshalObject(ctx.PerfObjects["Certification Authority"], &dst, logger)
	if err != nil {
		return err
	}

	if len(dst) == 0 {
		return errors.New("perflib query for Certification Authority (ADCS) returned empty result set")
	}

	for _, d := range dst {
		if d.Name == "" {
			continue
		}

		ch <- prometheus.MustNewConstMetric(
			c.requestsPerSecond,
			prometheus.CounterValue,
			d.RequestsPerSecond,
			d.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.requestProcessingTime,
			prometheus.GaugeValue,
			utils.MilliSecToSec(d.RequestProcessingTime),
			d.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.retrievalsPerSecond,
			prometheus.CounterValue,
			d.RetrievalsPerSecond,
			d.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.retrievalProcessingTime,
			prometheus.GaugeValue,
			utils.MilliSecToSec(d.RetrievalProcessingTime),
			d.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.failedRequestsPerSecond,
			prometheus.CounterValue,
			d.FailedRequestsPerSecond,
			d.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.issuedRequestsPerSecond,
			prometheus.CounterValue,
			d.IssuedRequestsPerSecond,
			d.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.pendingRequestsPerSecond,
			prometheus.CounterValue,
			d.PendingRequestsPerSecond,
			d.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.requestCryptographicSigningTime,
			prometheus.GaugeValue,
			utils.MilliSecToSec(d.RequestCryptographicSigningTime),
			d.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.requestPolicyModuleProcessingTime,
			prometheus.GaugeValue,
			utils.MilliSecToSec(d.RequestPolicyModuleProcessingTime),
			d.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.challengeResponsesPerSecond,
			prometheus.CounterValue,
			d.ChallengeResponsesPerSecond,
			d.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.challengeResponseProcessingTime,
			prometheus.GaugeValue,
			utils.MilliSecToSec(d.ChallengeResponseProcessingTime),
			d.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.signedCertificateTimestampListsPerSecond,
			prometheus.CounterValue,
			d.SignedCertificateTimestampListsPerSecond,
			d.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.signedCertificateTimestampListProcessingTime,
			prometheus.GaugeValue,
			utils.MilliSecToSec(d.SignedCertificateTimestampListProcessingTime),
			d.Name,
		)
	}

	return nil
}

func (c *Collector) collectPDH(ch chan<- prometheus.Metric) error {
	data, err := c.perfDataCollector.Collect()
	if err != nil {
		return fmt.Errorf("failed to collect Certification Authority (ADCS) metrics: %w", err)
	}

	if len(data) == 0 {
		return errors.New("perflib query for Certification Authority (ADCS) returned empty result set")
	}

	for name, adcsData := range data {
		ch <- prometheus.MustNewConstMetric(
			c.requestsPerSecond,
			prometheus.CounterValue,
			adcsData[RequestsPerSecond].FirstValue,
			name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.requestProcessingTime,
			prometheus.GaugeValue,
			utils.MilliSecToSec(adcsData[RequestProcessingTime].FirstValue),
			name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.retrievalsPerSecond,
			prometheus.CounterValue,
			adcsData[RetrievalsPerSecond].FirstValue,
			name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.retrievalProcessingTime,
			prometheus.GaugeValue,
			utils.MilliSecToSec(adcsData[RetrievalProcessingTime].FirstValue),
			name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.failedRequestsPerSecond,
			prometheus.CounterValue,
			adcsData[FailedRequestsPerSecond].FirstValue,
			name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.issuedRequestsPerSecond,
			prometheus.CounterValue,
			adcsData[IssuedRequestsPerSecond].FirstValue,
			name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.pendingRequestsPerSecond,
			prometheus.CounterValue,
			adcsData[PendingRequestsPerSecond].FirstValue,
			name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.requestCryptographicSigningTime,
			prometheus.GaugeValue,
			utils.MilliSecToSec(adcsData[RequestCryptographicSigningTime].FirstValue),
			name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.requestPolicyModuleProcessingTime,
			prometheus.GaugeValue,
			utils.MilliSecToSec(adcsData[RequestPolicyModuleProcessingTime].FirstValue),
			name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.challengeResponsesPerSecond,
			prometheus.CounterValue,
			adcsData[ChallengeResponsesPerSecond].FirstValue,
			name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.challengeResponseProcessingTime,
			prometheus.GaugeValue,
			utils.MilliSecToSec(adcsData[ChallengeResponseProcessingTime].FirstValue),
			name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.signedCertificateTimestampListsPerSecond,
			prometheus.CounterValue,
			adcsData[SignedCertificateTimestampListsPerSecond].FirstValue,
			name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.signedCertificateTimestampListProcessingTime,
			prometheus.GaugeValue,
			utils.MilliSecToSec(adcsData[SignedCertificateTimestampListProcessingTime].FirstValue),
			name,
		)
	}

	return nil
}
