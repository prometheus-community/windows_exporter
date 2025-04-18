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

package dns

type perfDataCounterValues struct {
	_                              float64 `perfdata:"% User Time"`
	_                              float64 `perfdata:"176"`
	_                              float64 `perfdata:"Async Fast Reads/sec"`
	AxfrRequestReceived            float64 `perfdata:"AXFR Request Received"`
	AxfrRequestSent                float64 `perfdata:"AXFR Request Sent"`
	AxfrResponseReceived           float64 `perfdata:"AXFR Response Received"`
	AxfrSuccessReceived            float64 `perfdata:"AXFR Success Received"`
	AxfrSuccessSent                float64 `perfdata:"AXFR Success Sent"`
	CachingMemory                  float64 `perfdata:"Caching Memory"`
	_                              float64 `perfdata:"Data Flush Pages/sec"`
	_                              float64 `perfdata:"Data Flushes/sec"`
	DatabaseNodeMemory             float64 `perfdata:"Database Node Memory"`
	DynamicUpdateNoOperation       float64 `perfdata:"Dynamic Update NoOperation"`
	_                              float64 `perfdata:"Dynamic Update NoOperation/sec"`
	DynamicUpdateQueued            float64 `perfdata:"Dynamic Update Queued"`
	_                              float64 `perfdata:"Dynamic Update Received"`
	_                              float64 `perfdata:"Dynamic Update Received/sec"`
	DynamicUpdateRejected          float64 `perfdata:"Dynamic Update Rejected"`
	DynamicUpdateTimeOuts          float64 `perfdata:"Dynamic Update TimeOuts"`
	DynamicUpdateWrittenToDatabase float64 `perfdata:"Dynamic Update Written to Database"`
	_                              float64 `perfdata:"Dynamic Update Written to Database/sec"`
	_                              float64 `perfdata:"Enumerations Server/sec"`
	_                              float64 `perfdata:"Fast Read Not Possibles/sec"`
	_                              float64 `perfdata:"Fast Read Resource Misses/sec"`
	IxfrRequestReceived            float64 `perfdata:"IXFR Request Received"`
	IxfrRequestSent                float64 `perfdata:"IXFR Request Sent"`
	IxfrResponseReceived           float64 `perfdata:"IXFR Response Received"`
	_                              float64 `perfdata:"IXFR Success Received"`
	IxfrSuccessSent                float64 `perfdata:"IXFR Success Sent"`
	IxfrTCPSuccessReceived         float64 `perfdata:"IXFR TCP Success Received"`
	IxfrUDPSuccessReceived         float64 `perfdata:"IXFR UDP Success Received"`
	_                              float64 `perfdata:"Lazy Write Flushes/sec"`
	_                              float64 `perfdata:"Lazy Write Pages/sec"`
	_                              float64 `perfdata:"Level 2 TLB Fills/sec"`
	NbStatMemory                   float64 `perfdata:"Nbstat Memory"`
	NotifyReceived                 float64 `perfdata:"Notify Received"`
	NotifySent                     float64 `perfdata:"Notify Sent"`
	_                              float64 `perfdata:"Query Dropped Bad Socket"`
	_                              float64 `perfdata:"Query Dropped Bad Socket/sec"`
	_                              float64 `perfdata:"Query Dropped By Policy"`
	_                              float64 `perfdata:"Query Dropped By Policy/sec"`
	_                              float64 `perfdata:"Query Dropped By Response Rate Limiting"`
	_                              float64 `perfdata:"Query Dropped By Response Rate Limiting/sec"`
	_                              float64 `perfdata:"Query Dropped Send"`
	_                              float64 `perfdata:"Query Dropped Send/sec"`
	_                              float64 `perfdata:"Query Dropped Total"`
	_                              float64 `perfdata:"Query Dropped Total/sec"`
	RecordFlowMemory               float64 `perfdata:"Record Flow Memory"`
	RecursiveQueries               float64 `perfdata:"Recursive Queries"`
	_                              float64 `perfdata:"Recursive Queries/sec"`
	RecursiveQueryFailure          float64 `perfdata:"Recursive Query Failure"`
	_                              float64 `perfdata:"Recursive Query Failure/sec"`
	_                              float64 `perfdata:"Recursive Send TimeOuts"`
	RecursiveSendTimeOuts          float64 `perfdata:"Recursive TimeOut/sec"`
	_                              float64 `perfdata:"Responses Suppressed"`
	_                              float64 `perfdata:"Responses Suppressed/sec"`
	SecureUpdateFailure            float64 `perfdata:"Secure Update Failure"`
	SecureUpdateReceived           float64 `perfdata:"Secure Update Received"`
	_                              float64 `perfdata:"Secure Update Received/sec"`
	TcpMessageMemory               float64 `perfdata:"TCP Message Memory"`
	TcpQueryReceived               float64 `perfdata:"TCP Query Received"`
	_                              float64 `perfdata:"TCP Query Received/sec"`
	TcpResponseSent                float64 `perfdata:"TCP Response Sent"`
	_                              float64 `perfdata:"TCP Response Sent/sec"`
	_                              float64 `perfdata:"Total Query Received"`
	_                              float64 `perfdata:"Total Query Received/sec"`
	_                              float64 `perfdata:"Total Remote Inflight Queries"`
	_                              float64 `perfdata:"Total Response Sent"`
	_                              float64 `perfdata:"Total Response Sent/sec"`
	UdpMessageMemory               float64 `perfdata:"UDP Message Memory"`
	UdpQueryReceived               float64 `perfdata:"UDP Query Received"`
	_                              float64 `perfdata:"UDP Query Received/sec"`
	UdpResponseSent                float64 `perfdata:"UDP Response Sent"`
	_                              float64 `perfdata:"UDP Response Sent/sec"`
	UnmatchedResponsesReceived     float64 `perfdata:"Unmatched Responses Received"`
	_                              float64 `perfdata:"Virtual Bytes"`
	WinsLookupReceived             float64 `perfdata:"WINS Lookup Received"`
	_                              float64 `perfdata:"WINS Lookup Received/sec"`
	WinsResponseSent               float64 `perfdata:"WINS Response Sent"`
	_                              float64 `perfdata:"WINS Response Sent/sec"`
	WinsReverseLookupReceived      float64 `perfdata:"WINS Reverse Lookup Received"`
	_                              float64 `perfdata:"WINS Reverse Lookup Received/sec"`
	WinsReverseResponseSent        float64 `perfdata:"WINS Reverse Response Sent"`
	_                              float64 `perfdata:"WINS Reverse Response Sent/sec"`
	ZoneTransferFailure            float64 `perfdata:"Zone Transfer Failure"`
	ZoneTransferSOARequestSent     float64 `perfdata:"Zone Transfer Request Received"`
	_                              float64 `perfdata:"Zone Transfer SOA Request Sent"`
	_                              float64 `perfdata:"Zone Transfer Success"`
}

// Statistic represents the structure for DNS error statistics
type Statistic struct {
	Name           string `mi:"Name"`
	CollectionName string `mi:"CollectionName"`
	Value          uint64 `mi:"Value"`
	DnsServerName  string `mi:"DnsServerName"`
}
