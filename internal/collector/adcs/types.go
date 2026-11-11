// SPDX-License-Identifier: MIT
//
// Copyright The Prometheus Authors

//go:build windows

package adcs

type perfDataCounterValues struct {
	Name string

	ChallengeResponseProcessingTime              float64 `perfdata:"Challenge Response processing time (ms)"`
	ChallengeResponsesPerSecond                  float64 `perfdata:"Challenge Responses/sec"`
	FailedRequestsPerSecond                      float64 `perfdata:"Failed Requests/sec"`
	IssuedRequestsPerSecond                      float64 `perfdata:"Issued Requests/sec"`
	PendingRequestsPerSecond                     float64 `perfdata:"Pending Requests/sec"`
	RequestCryptographicSigningTime              float64 `perfdata:"Request cryptographic signing time (ms)"`
	RequestPolicyModuleProcessingTime            float64 `perfdata:"Request policy module processing time (ms)"`
	RequestProcessingTime                        float64 `perfdata:"Request processing time (ms)"`
	RequestsPerSecond                            float64 `perfdata:"Requests/sec"`
	RetrievalProcessingTime                      float64 `perfdata:"Retrieval processing time (ms)"`
	RetrievalsPerSecond                          float64 `perfdata:"Retrievals/sec"`
	SignedCertificateTimestampListProcessingTime float64 `perfdata:"Signed Certificate Timestamp List processing time (ms)"`
	SignedCertificateTimestampListsPerSecond     float64 `perfdata:"Signed Certificate Timestamp Lists/sec"`
}
