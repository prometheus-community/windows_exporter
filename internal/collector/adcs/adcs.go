// Copyright 2024 The Prometheus Authors
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

//go:build windows

package adcs

import (
	"fmt"
	"log/slog"

	"github.com/alecthomas/kingpin/v2"
	"github.com/prometheus-community/windows_exporter/internal/mi"
	"github.com/prometheus-community/windows_exporter/internal/pdh"
	"github.com/prometheus-community/windows_exporter/internal/types"
	"github.com/prometheus-community/windows_exporter/internal/utils"
	"github.com/prometheus/client_golang/prometheus"
)

const Name = "adcs"

type Config struct{}

//nolint:gochecknoglobals
var ConfigDefaults = Config{}

type Collector struct {
	config Config

	perfDataCollector *pdh.Collector
	perfDataObject    []perfDataCounterValues

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

func (c *Collector) Close() error {
	c.perfDataCollector.Close()

	return nil
}

func (c *Collector) Build(_ *slog.Logger, _ *mi.Session) error {
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

	var err error

	c.perfDataCollector, err = pdh.NewCollector[perfDataCounterValues](pdh.CounterTypeRaw, "Certification Authority", pdh.InstancesAll)
	if err != nil {
		return fmt.Errorf("failed to create Certification Authority collector: %w", err)
	}

	return nil
}

func (c *Collector) Collect(ch chan<- prometheus.Metric) error {
	err := c.perfDataCollector.Collect(&c.perfDataObject)
	if err != nil {
		return fmt.Errorf("failed to collect Certification Authority (ADCS) metrics: %w", err)
	}

	for _, data := range c.perfDataObject {
		ch <- prometheus.MustNewConstMetric(
			c.requestsPerSecond,
			prometheus.CounterValue,
			data.RequestsPerSecond,
			data.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.requestProcessingTime,
			prometheus.GaugeValue,
			utils.MilliSecToSec(data.RequestProcessingTime),
			data.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.retrievalsPerSecond,
			prometheus.CounterValue,
			data.RetrievalsPerSecond,
			data.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.retrievalProcessingTime,
			prometheus.GaugeValue,
			utils.MilliSecToSec(data.RetrievalProcessingTime),
			data.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.failedRequestsPerSecond,
			prometheus.CounterValue,
			data.FailedRequestsPerSecond,
			data.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.issuedRequestsPerSecond,
			prometheus.CounterValue,
			data.IssuedRequestsPerSecond,
			data.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.pendingRequestsPerSecond,
			prometheus.CounterValue,
			data.PendingRequestsPerSecond,
			data.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.requestCryptographicSigningTime,
			prometheus.GaugeValue,
			utils.MilliSecToSec(data.RequestCryptographicSigningTime),
			data.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.requestPolicyModuleProcessingTime,
			prometheus.GaugeValue,
			utils.MilliSecToSec(data.RequestPolicyModuleProcessingTime),
			data.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.challengeResponsesPerSecond,
			prometheus.CounterValue,
			data.ChallengeResponsesPerSecond,
			data.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.challengeResponseProcessingTime,
			prometheus.GaugeValue,
			utils.MilliSecToSec(data.ChallengeResponseProcessingTime),
			data.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.signedCertificateTimestampListsPerSecond,
			prometheus.CounterValue,
			data.SignedCertificateTimestampListsPerSecond,
			data.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.signedCertificateTimestampListProcessingTime,
			prometheus.GaugeValue,
			utils.MilliSecToSec(data.SignedCertificateTimestampListProcessingTime),
			data.Name,
		)
	}

	return nil
}
