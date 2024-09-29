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
