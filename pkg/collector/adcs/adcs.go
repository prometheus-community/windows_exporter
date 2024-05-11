//go:build windows

package adcs

import (
	"errors"
	"strings"

	"github.com/alecthomas/kingpin/v2"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/prometheus-community/windows_exporter/pkg/perflib"
	"github.com/prometheus-community/windows_exporter/pkg/types"
	"github.com/prometheus-community/windows_exporter/pkg/utils"
	"github.com/prometheus/client_golang/prometheus"
)

const Name = "adcs"

type Config struct{}

var ConfigDefaults = Config{}

type collector struct {
	logger log.Logger

	RequestsPerSecond                            *prometheus.Desc
	RequestProcessingTime                        *prometheus.Desc
	RetrievalsPerSecond                          *prometheus.Desc
	RetrievalProcessingTime                      *prometheus.Desc
	FailedRequestsPerSecond                      *prometheus.Desc
	IssuedRequestsPerSecond                      *prometheus.Desc
	PendingRequestsPerSecond                     *prometheus.Desc
	RequestCryptographicSigningTime              *prometheus.Desc
	RequestPolicyModuleProcessingTime            *prometheus.Desc
	ChallengeResponsesPerSecond                  *prometheus.Desc
	ChallengeResponseProcessingTime              *prometheus.Desc
	SignedCertificateTimestampListsPerSecond     *prometheus.Desc
	SignedCertificateTimestampListProcessingTime *prometheus.Desc
}

func New(logger log.Logger, _ *Config) types.Collector {
	c := &collector{}
	c.SetLogger(logger)
	return c
}

func NewWithFlags(_ *kingpin.Application) types.Collector {
	return &collector{}
}

func (c *collector) GetName() string {
	return Name
}

func (c *collector) SetLogger(logger log.Logger) {
	c.logger = log.With(logger, "collector", Name)
}

func (c *collector) GetPerfCounter() ([]string, error) {
	return []string{"Certification Authority"}, nil
}

func (c *collector) Build() error {
	c.RequestsPerSecond = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "requests_total"),
		"Total certificate requests processed",
		[]string{"cert_template"},
		nil,
	)
	c.RequestProcessingTime = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "request_processing_time_seconds"),
		"Last time elapsed for certificate requests",
		[]string{"cert_template"},
		nil,
	)
	c.RetrievalsPerSecond = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "retrievals_total"),
		"Total certificate retrieval requests processed",
		[]string{"cert_template"},
		nil,
	)
	c.RetrievalProcessingTime = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "retrievals_processing_time_seconds"),
		"Last time elapsed for certificate retrieval request",
		[]string{"cert_template"},
		nil,
	)
	c.FailedRequestsPerSecond = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "failed_requests_total"),
		"Total failed certificate requests processed",
		[]string{"cert_template"},
		nil,
	)
	c.IssuedRequestsPerSecond = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "issued_requests_total"),
		"Total issued certificate requests processed",
		[]string{"cert_template"},
		nil,
	)
	c.PendingRequestsPerSecond = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "pending_requests_total"),
		"Total pending certificate requests processed",
		[]string{"cert_template"},
		nil,
	)
	c.RequestCryptographicSigningTime = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "request_cryptographic_signing_time_seconds"),
		"Last time elapsed for signing operation request",
		[]string{"cert_template"},
		nil,
	)
	c.RequestPolicyModuleProcessingTime = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "request_policy_module_processing_time_seconds"),
		"Last time elapsed for policy module processing request",
		[]string{"cert_template"},
		nil,
	)
	c.ChallengeResponsesPerSecond = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "challenge_responses_total"),
		"Total certificate challenge responses processed",
		[]string{"cert_template"},
		nil,
	)
	c.ChallengeResponseProcessingTime = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "challenge_response_processing_time_seconds"),
		"Last time elapsed for challenge response",
		[]string{"cert_template"},
		nil,
	)
	c.SignedCertificateTimestampListsPerSecond = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "signed_certificate_timestamp_lists_total"),
		"Total Signed Certificate Timestamp Lists processed",
		[]string{"cert_template"},
		nil,
	)
	c.SignedCertificateTimestampListProcessingTime = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "signed_certificate_timestamp_list_processing_time_seconds"),
		"Last time elapsed for Signed Certificate Timestamp List",
		[]string{"cert_template"},
		nil,
	)

	return nil
}

func (c *collector) Collect(ctx *types.ScrapeContext, ch chan<- prometheus.Metric) error {
	if err := c.collectADCSCounters(ctx, ch); err != nil {
		_ = level.Error(c.logger).Log("msg", "failed collecting ADCS metrics", "err", err)
		return err
	}
	return nil
}

type perflibADCS struct {
	Name                                         string
	RequestsPerSecond                            float64 `perflib:"Requests/sec"`
	RequestProcessingTime                        float64 `perflib:"Request processing time (ms)"`
	RetrievalsPerSecond                          float64 `perflib:"Retrievals/sec"`
	RetrievalProcessingTime                      float64 `perflib:"Retrieval processing time (ms)"`
	FailedRequestsPerSecond                      float64 `perflib:"Failed Requests/sec"`
	IssuedRequestsPerSecond                      float64 `perflib:"Issued Requests/sec"`
	PendingRequestsPerSecond                     float64 `perflib:"Pending Requests/sec"`
	RequestCryptographicSigningTime              float64 `perflib:"Request cryptographic signing time (ms)"`
	RequestPolicyModuleProcessingTime            float64 `perflib:"Request policy module processing time (ms)"`
	ChallengeResponsesPerSecond                  float64 `perflib:"Challenge Responses/sec"`
	ChallengeResponseProcessingTime              float64 `perflib:"Challenge Response processing time (ms)"`
	SignedCertificateTimestampListsPerSecond     float64 `perflib:"Signed Certificate Timestamp Lists/sec"`
	SignedCertificateTimestampListProcessingTime float64 `perflib:"Signed Certificate Timestamp List processing time (ms)"`
}

func (c *collector) collectADCSCounters(ctx *types.ScrapeContext, ch chan<- prometheus.Metric) error {
	dst := make([]perflibADCS, 0)
	if _, ok := ctx.PerfObjects["Certification Authority"]; !ok {
		return errors.New("perflib did not contain an entry for Certification Authority")
	}
	err := perflib.UnmarshalObject(ctx.PerfObjects["Certification Authority"], &dst, c.logger)
	if err != nil {
		return err
	}
	if len(dst) == 0 {
		return errors.New("perflib query for Certification Authority (ADCS) returned empty result set")
	}

	for _, d := range dst {
		n := strings.ToLower(d.Name)
		if n == "" {
			continue
		}
		ch <- prometheus.MustNewConstMetric(
			c.RequestsPerSecond,
			prometheus.CounterValue,
			d.RequestsPerSecond,
			d.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.RequestProcessingTime,
			prometheus.GaugeValue,
			utils.MilliSecToSec(d.RequestProcessingTime),
			d.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.RetrievalsPerSecond,
			prometheus.CounterValue,
			d.RetrievalsPerSecond,
			d.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.RetrievalProcessingTime,
			prometheus.GaugeValue,
			utils.MilliSecToSec(d.RetrievalProcessingTime),
			d.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.FailedRequestsPerSecond,
			prometheus.CounterValue,
			d.FailedRequestsPerSecond,
			d.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.IssuedRequestsPerSecond,
			prometheus.CounterValue,
			d.IssuedRequestsPerSecond,
			d.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.PendingRequestsPerSecond,
			prometheus.CounterValue,
			d.PendingRequestsPerSecond,
			d.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.RequestCryptographicSigningTime,
			prometheus.GaugeValue,
			utils.MilliSecToSec(d.RequestCryptographicSigningTime),
			d.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.RequestPolicyModuleProcessingTime,
			prometheus.GaugeValue,
			utils.MilliSecToSec(d.RequestPolicyModuleProcessingTime),
			d.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.ChallengeResponsesPerSecond,
			prometheus.CounterValue,
			d.ChallengeResponsesPerSecond,
			d.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.ChallengeResponseProcessingTime,
			prometheus.GaugeValue,
			utils.MilliSecToSec(d.ChallengeResponseProcessingTime),
			d.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.SignedCertificateTimestampListsPerSecond,
			prometheus.CounterValue,
			d.SignedCertificateTimestampListsPerSecond,
			d.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.SignedCertificateTimestampListProcessingTime,
			prometheus.GaugeValue,
			utils.MilliSecToSec(d.SignedCertificateTimestampListProcessingTime),
			d.Name,
		)
	}

	return nil
}
