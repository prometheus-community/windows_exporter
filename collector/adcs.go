//go:build windows
// +build windows

package collector

import (
	"errors"
	"github.com/prometheus-community/windows_exporter/log"
	"github.com/prometheus/client_golang/prometheus"
	"strings"
)

func init() {
	registerCollector("adcs", adcsCollectorMethod, "Certification Authority")
}

type adcsCollector struct {
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

// ADCSCollectorMethod ...
func adcsCollectorMethod() (Collector, error) {
	const subsystem = "adcs"
	return &adcsCollector{
		RequestsPerSecond: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "requests_total"),
			"Total certificate requests processed",
			[]string{"cert_template"},
			nil,
		),
		RequestProcessingTime: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "request_processing_time_seconds"),
			"Last time elapsed for certificate requests",
			[]string{"cert_template"},
			nil,
		),
		RetrievalsPerSecond: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "retrievals_total"),
			"Total certificate retrieval requests processed",
			[]string{"cert_template"},
			nil,
		),
		RetrievalProcessingTime: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "retrievals_processing_time_seconds"),
			"Last time elapsed for certificate retrieval request",
			[]string{"cert_template"},
			nil,
		),
		FailedRequestsPerSecond: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "failed_requests_total"),
			"Total failed certificate requests processed",
			[]string{"cert_template"},
			nil,
		),
		IssuedRequestsPerSecond: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "issued_requests_total"),
			"Total issued certificate requests processed",
			[]string{"cert_template"},
			nil,
		),
		PendingRequestsPerSecond: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "pending_requests_total"),
			"Total pending certificate requests processed",
			[]string{"cert_template"},
			nil,
		),
		RequestCryptographicSigningTime: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "request_cryptographic_signing_time_seconds"),
			"Last time elapsed for signing operation request",
			[]string{"cert_template"},
			nil,
		),
		RequestPolicyModuleProcessingTime: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "request_policy_module_processing_time_seconds"),
			"Last time elapsed for policy module processing request",
			[]string{"cert_template"},
			nil,
		),
		ChallengeResponsesPerSecond: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "challenge_responses_total"),
			"Total certificate challenge responses processed",
			[]string{"cert_template"},
			nil,
		),
		ChallengeResponseProcessingTime: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "challenge_response_processing_time_seconds"),
			"Last time elapsed for challenge response",
			[]string{"cert_template"},
			nil,
		),
		SignedCertificateTimestampListsPerSecond: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "signed_certificate_timestamp_lists_total"),
			"Total Signed Certificate Timestamp Lists processed",
			[]string{"cert_template"},
			nil,
		),
		SignedCertificateTimestampListProcessingTime: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "signed_certificate_timestamp_list_processing_time_seconds"),
			"Last time elapsed for Signed Certificate Timestamp List",
			[]string{"cert_template"},
			nil,
		),
	}, nil
}

func (c *adcsCollector) Collect(ctx *ScrapeContext, ch chan<- prometheus.Metric) error {
	if desc, err := c.collectADCSCounters(ctx, ch); err != nil {
		log.Error("Failed collecting ADCS Metrics:", desc, err)
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

func (c *adcsCollector) collectADCSCounters(ctx *ScrapeContext, ch chan<- prometheus.Metric) (*prometheus.Desc, error) {
	dst := make([]perflibADCS, 0)
	if _, ok := ctx.perfObjects["Certification Authority"]; !ok {
		return nil, errors.New("Perflib did not contain an entry for Certification Authority")
	}
	err := unmarshalObject(ctx.perfObjects["Certification Authority"], &dst)
	if err != nil {
		return nil, err
	}
	if len(dst) == 0 {
		return nil, errors.New("Perflib query for Certification Authority (ADCS) returned empty result set")
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
			milliSecToSec(d.RequestProcessingTime),
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
			milliSecToSec(d.RetrievalProcessingTime),
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
			milliSecToSec(d.RequestCryptographicSigningTime),
			d.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.RequestPolicyModuleProcessingTime,
			prometheus.GaugeValue,
			milliSecToSec(d.RequestPolicyModuleProcessingTime),
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
			milliSecToSec(d.ChallengeResponseProcessingTime),
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
			milliSecToSec(d.SignedCertificateTimestampListProcessingTime),
			d.Name,
		)
	}

	return nil, nil
}
