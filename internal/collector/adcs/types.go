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
