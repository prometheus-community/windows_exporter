package adcs

const (
	RequestsPerSecond                            = "Requests/sec"
	RequestProcessingTime                        = "Request processing time (ms)"
	RetrievalsPerSecond                          = "Retrievals/sec"
	RetrievalProcessingTime                      = "Retrieval processing time (ms)"
	FailedRequestsPerSecond                      = "Failed Requests/sec"
	IssuedRequestsPerSecond                      = "Issued Requests/sec"
	PendingRequestsPerSecond                     = "Pending Requests/sec"
	RequestCryptographicSigningTime              = "Request cryptographic signing time (ms)"
	RequestPolicyModuleProcessingTime            = "Request policy module processing time (ms)"
	ChallengeResponsesPerSecond                  = "Challenge Responses/sec"
	ChallengeResponseProcessingTime              = "Challenge Response processing time (ms)"
	SignedCertificateTimestampListsPerSecond     = "Signed Certificate Timestamp Lists/sec"
	SignedCertificateTimestampListProcessingTime = "Signed Certificate Timestamp List processing time (ms)"
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
