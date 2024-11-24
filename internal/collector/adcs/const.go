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

const (
	challengeResponseProcessingTime              = "Challenge Response processing time (ms)"
	challengeResponsesPerSecond                  = "Challenge Responses/sec"
	failedRequestsPerSecond                      = "Failed Requests/sec"
	issuedRequestsPerSecond                      = "Issued Requests/sec"
	pendingRequestsPerSecond                     = "Pending Requests/sec"
	requestCryptographicSigningTime              = "Request cryptographic signing time (ms)"
	requestPolicyModuleProcessingTime            = "Request policy module processing time (ms)"
	requestProcessingTime                        = "Request processing time (ms)"
	requestsPerSecond                            = "Requests/sec"
	retrievalProcessingTime                      = "Retrieval processing time (ms)"
	retrievalsPerSecond                          = "Retrievals/sec"
	signedCertificateTimestampListProcessingTime = "Signed Certificate Timestamp List processing time (ms)"
	signedCertificateTimestampListsPerSecond     = "Signed Certificate Timestamp Lists/sec"
)
