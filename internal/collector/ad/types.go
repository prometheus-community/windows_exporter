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

package ad

type perfDataCounterValues struct {
	AbANRPerSec                                                      float64 `perfdata:"AB ANR/sec"`
	AbBrowsesPerSec                                                  float64 `perfdata:"AB Browses/sec"`
	AbClientSessions                                                 float64 `perfdata:"AB Client Sessions"`
	AbMatchesPerSec                                                  float64 `perfdata:"AB Matches/sec"`
	AbPropertyReadsPerSec                                            float64 `perfdata:"AB Property Reads/sec"`
	AbProxyLookupsPerSec                                             float64 `perfdata:"AB Proxy Lookups/sec"`
	AbSearchesPerSec                                                 float64 `perfdata:"AB Searches/sec"`
	ApproximateHighestDNT                                            float64 `perfdata:"Approximate highest DNT"`
	AtqEstimatedQueueDelay                                           float64 `perfdata:"ATQ Estimated Queue Delay"`
	AtqOutstandingQueuedRequests                                     float64 `perfdata:"ATQ Outstanding Queued Requests"`
	_                                                                float64 `perfdata:"ATQ Queue Latency"`
	AtqRequestLatency                                                float64 `perfdata:"ATQ Request Latency"`
	AtqThreadsLDAP                                                   float64 `perfdata:"ATQ Threads LDAP"`
	AtqThreadsOther                                                  float64 `perfdata:"ATQ Threads Other"`
	AtqThreadsTotal                                                  float64 `perfdata:"ATQ Threads Total"`
	BaseSearchesPerSec                                               float64 `perfdata:"Base searches/sec"`
	DatabaseAddsPerSec                                               float64 `perfdata:"Database adds/sec"`
	DatabaseDeletesPerSec                                            float64 `perfdata:"Database deletes/sec"`
	DatabaseModifiesPerSec                                           float64 `perfdata:"Database modifys/sec"`
	DatabaseRecyclesPerSec                                           float64 `perfdata:"Database recycles/sec"`
	DigestBindsPerSec                                                float64 `perfdata:"Digest Binds/sec"`
	_                                                                float64 `perfdata:"DirSync session throttling rate"`
	_                                                                float64 `perfdata:"DirSync sessions in progress"`
	DRAHighestUSNCommittedHighPart                                   float64 `perfdata:"DRA Highest USN Committed (High part)"`
	DRAHighestUSNCommittedLowPart                                    float64 `perfdata:"DRA Highest USN Committed (Low part)"`
	DRAHighestUSNIssuedHighPart                                      float64 `perfdata:"DRA Highest USN Issued (High part)"`
	DRAHighestUSNIssuedLowPart                                       float64 `perfdata:"DRA Highest USN Issued (Low part)"`
	DRAInboundBytesCompressedBetweenSitesAfterCompressionSinceBoot   float64 `perfdata:"DRA Inbound Bytes Compressed (Between Sites, After Compression) Since Boot"`
	DRAInboundBytesCompressedBetweenSitesAfterCompressionPerSec      float64 `perfdata:"DRA Inbound Bytes Compressed (Between Sites, After Compression)/sec"`
	DRAInboundBytesCompressedBetweenSitesBeforeCompressionSinceBoot  float64 `perfdata:"DRA Inbound Bytes Compressed (Between Sites, Before Compression) Since Boot"`
	DRAInboundBytesCompressedBetweenSitesBeforeCompressionPerSec     float64 `perfdata:"DRA Inbound Bytes Compressed (Between Sites, Before Compression)/sec"`
	DRAInboundBytesNotCompressedWithinSiteSinceBoot                  float64 `perfdata:"DRA Inbound Bytes Not Compressed (Within Site) Since Boot"`
	DRAInboundBytesNotCompressedWithinSitePerSec                     float64 `perfdata:"DRA Inbound Bytes Not Compressed (Within Site)/sec"`
	DRAInboundBytesTotalSinceBoot                                    float64 `perfdata:"DRA Inbound Bytes Total Since Boot"`
	DRAInboundBytesTotalPerSec                                       float64 `perfdata:"DRA Inbound Bytes Total/sec"`
	DRAInboundFullSyncObjectsRemaining                               float64 `perfdata:"DRA Inbound Full Sync Objects Remaining"`
	DRAInboundLinkValueUpdatesRemainingInPacket                      float64 `perfdata:"DRA Inbound Link Value Updates Remaining in Packet"`
	_                                                                float64 `perfdata:"DRA Inbound Link Values/sec"`
	DRAInboundObjectUpdatesRemainingInPacket                         float64 `perfdata:"DRA Inbound Object Updates Remaining in Packet"`
	DRAInboundObjectsAppliedPerSec                                   float64 `perfdata:"DRA Inbound Objects Applied/sec"`
	DRAInboundObjectsFilteredPerSec                                  float64 `perfdata:"DRA Inbound Objects Filtered/sec"`
	DRAInboundObjectsPerSec                                          float64 `perfdata:"DRA Inbound Objects/sec"`
	DRAInboundPropertiesAppliedPerSec                                float64 `perfdata:"DRA Inbound Properties Applied/sec"`
	DRAInboundPropertiesFilteredPerSec                               float64 `perfdata:"DRA Inbound Properties Filtered/sec"`
	DRAInboundPropertiesTotalPerSec                                  float64 `perfdata:"DRA Inbound Properties Total/sec"`
	_                                                                float64 `perfdata:"DRA Inbound Sync Link Deletion/sec"`
	DRAInboundTotalUpdatesRemainingInPacket                          float64 `perfdata:"DRA Inbound Total Updates Remaining in Packet"`
	DRAInboundValuesDNsOnlyPerSec                                    float64 `perfdata:"DRA Inbound Values (DNs only)/sec"`
	DRAInboundValuesTotalPerSec                                      float64 `perfdata:"DRA Inbound Values Total/sec"`
	_                                                                float64 `perfdata:"DRA number of NC replication calls since boot"`
	_                                                                float64 `perfdata:"DRA number of successful NC replication calls since boot"`
	DRAOutboundBytesCompressedBetweenSitesAfterCompressionSinceBoot  float64 `perfdata:"DRA Outbound Bytes Compressed (Between Sites, After Compression) Since Boot"`
	DRAOutboundBytesCompressedBetweenSitesAfterCompressionPerSec     float64 `perfdata:"DRA Outbound Bytes Compressed (Between Sites, After Compression)/sec"`
	DRAOutboundBytesCompressedBetweenSitesBeforeCompressionSinceBoot float64 `perfdata:"DRA Outbound Bytes Compressed (Between Sites, Before Compression) Since Boot"`
	DRAOutboundBytesCompressedBetweenSitesBeforeCompressionPerSec    float64 `perfdata:"DRA Outbound Bytes Compressed (Between Sites, Before Compression)/sec"`
	DRAOutboundBytesNotCompressedWithinSiteSinceBoot                 float64 `perfdata:"DRA Outbound Bytes Not Compressed (Within Site) Since Boot"`
	DRAOutboundBytesNotCompressedWithinSitePerSec                    float64 `perfdata:"DRA Outbound Bytes Not Compressed (Within Site)/sec"`
	DRAOutboundBytesTotalSinceBoot                                   float64 `perfdata:"DRA Outbound Bytes Total Since Boot"`
	DRAOutboundBytesTotalPerSec                                      float64 `perfdata:"DRA Outbound Bytes Total/sec"`
	DRAOutboundObjectsFilteredPerSec                                 float64 `perfdata:"DRA Outbound Objects Filtered/sec"`
	DRAOutboundObjectsPerSec                                         float64 `perfdata:"DRA Outbound Objects/sec"`
	DRAOutboundPropertiesPerSec                                      float64 `perfdata:"DRA Outbound Properties/sec"`
	DRAOutboundValuesDNsOnlyPerSec                                   float64 `perfdata:"DRA Outbound Values (DNs only)/sec"`
	DRAOutboundValuesTotalPerSec                                     float64 `perfdata:"DRA Outbound Values Total/sec"`
	DRAPendingReplicationOperations                                  float64 `perfdata:"DRA Pending Replication Operations"`
	DRAPendingReplicationSynchronizations                            float64 `perfdata:"DRA Pending Replication Synchronizations"`
	DRASyncFailuresOnSchemaMismatch                                  float64 `perfdata:"DRA Sync Failures on Schema Mismatch"`
	DRASyncRequestsMade                                              float64 `perfdata:"DRA Sync Requests Made"`
	DRASyncRequestsSuccessful                                        float64 `perfdata:"DRA Sync Requests Successful"`
	DRAThreadsGettingNCChanges                                       float64 `perfdata:"DRA Threads Getting NC Changes"`
	DRAThreadsGettingNCChangesHoldingSemaphore                       float64 `perfdata:"DRA Threads Getting NC Changes Holding Semaphore"`
	_                                                                float64 `perfdata:"DRA total number of Busy failures since boot"`
	_                                                                float64 `perfdata:"DRA total number of MissingParent failures since boot"`
	_                                                                float64 `perfdata:"DRA total number of NotEnoughAttrs/MissingObject failures since boot"`
	_                                                                float64 `perfdata:"DRA total number of Preempted failures since boot"`
	_                                                                float64 `perfdata:"DRA total time of applying replication package since boot"`
	_                                                                float64 `perfdata:"DRA total time of NC replication calls since boot"`
	_                                                                float64 `perfdata:"DRA total time of successful NC replication calls since boot"`
	_                                                                float64 `perfdata:"DRA total time of successfully applying replication package since boot"`
	_                                                                float64 `perfdata:"DRA total time on waiting async replication packages since boot"`
	_                                                                float64 `perfdata:"DRA total time on waiting sync replication packages since boot"`
	DsPercentReadsFromDRA                                            float64 `perfdata:"DS % Reads from DRA"`
	DsPercentReadsFromKCC                                            float64 `perfdata:"DS % Reads from KCC"`
	DsPercentReadsFromLSA                                            float64 `perfdata:"DS % Reads from LSA"`
	DsPercentReadsFromNSPI                                           float64 `perfdata:"DS % Reads from NSPI"`
	DsPercentReadsFromNTDSAPI                                        float64 `perfdata:"DS % Reads from NTDSAPI"`
	DsPercentReadsFromSAM                                            float64 `perfdata:"DS % Reads from SAM"`
	DsPercentReadsOther                                              float64 `perfdata:"DS % Reads Other"`
	DsPercentSearchesFromDRA                                         float64 `perfdata:"DS % Searches from DRA"`
	DsPercentSearchesFromKCC                                         float64 `perfdata:"DS % Searches from KCC"`
	DsPercentSearchesFromLDAP                                        float64 `perfdata:"DS % Searches from LDAP"`
	DsPercentSearchesFromLSA                                         float64 `perfdata:"DS % Searches from LSA"`
	DsPercentSearchesFromNSPI                                        float64 `perfdata:"DS % Searches from NSPI"`
	DsPercentSearchesFromNTDSAPI                                     float64 `perfdata:"DS % Searches from NTDSAPI"`
	DsPercentSearchesFromSAM                                         float64 `perfdata:"DS % Searches from SAM"`
	DsPercentSearchesOther                                           float64 `perfdata:"DS % Searches Other"`
	DsPercentWritesFromDRA                                           float64 `perfdata:"DS % Writes from DRA"`
	DsPercentWritesFromKCC                                           float64 `perfdata:"DS % Writes from KCC"`
	DsPercentWritesFromLDAP                                          float64 `perfdata:"DS % Writes from LDAP"`
	DsPercentWritesFromLSA                                           float64 `perfdata:"DS % Writes from LSA"`
	DsPercentWritesFromNSPI                                          float64 `perfdata:"DS % Writes from NSPI"`
	DsPercentWritesFromNTDSAPI                                       float64 `perfdata:"DS % Writes from NTDSAPI"`
	DsPercentWritesFromSAM                                           float64 `perfdata:"DS % Writes from SAM"`
	DsPercentWritesOther                                             float64 `perfdata:"DS % Writes Other"`
	DsClientBindsPerSec                                              float64 `perfdata:"DS Client Binds/sec"`
	DsClientNameTranslationsPerSec                                   float64 `perfdata:"DS Client Name Translations/sec"`
	DsDirectoryReadsPerSec                                           float64 `perfdata:"DS Directory Reads/sec"`
	DsDirectorySearchesPerSec                                        float64 `perfdata:"DS Directory Searches/sec"`
	DsDirectoryWritesPerSec                                          float64 `perfdata:"DS Directory Writes/sec"`
	DsMonitorListSize                                                float64 `perfdata:"DS Monitor List Size"`
	DsNameCacheHitRate                                               float64 `perfdata:"DS Name Cache hit rate"`
	DsNameCacheHitRateSecondValue                                    float64 `perfdata:"DS Name Cache hit rate,secondvalue"`
	DsNotifyQueueSize                                                float64 `perfdata:"DS Notify Queue Size"`
	DsSearchSubOperationsPerSec                                      float64 `perfdata:"DS Search sub-operations/sec"`
	DsSecurityDescriptorPropagationsEvents                           float64 `perfdata:"DS Security Descriptor Propagations Events"`
	DsSecurityDescriptorPropagatorAverageExclusionTime               float64 `perfdata:"DS Security Descriptor Propagator Average Exclusion Time"`
	DsSecurityDescriptorPropagatorRuntimeQueue                       float64 `perfdata:"DS Security Descriptor Propagator Runtime Queue"`
	DsSecurityDescriptorSubOperationsPerSec                          float64 `perfdata:"DS Security Descriptor sub-operations/sec"`
	DsServerBindsPerSec                                              float64 `perfdata:"DS Server Binds/sec"`
	DsServerNameTranslationsPerSec                                   float64 `perfdata:"DS Server Name Translations/sec"`
	DsThreadsInUse                                                   float64 `perfdata:"DS Threads in Use"`
	_                                                                float64 `perfdata:"Error eventlogs since boot"`
	_                                                                float64 `perfdata:"Error events since boot"`
	ExternalBindsPerSec                                              float64 `perfdata:"External Binds/sec"`
	FastBindsPerSec                                                  float64 `perfdata:"Fast Binds/sec"`
	_                                                                float64 `perfdata:"Fatal events since boot"`
	_                                                                float64 `perfdata:"Info eventlogs since boot"`
	LdapActiveThreads                                                float64 `perfdata:"LDAP Active Threads"`
	_                                                                float64 `perfdata:"LDAP Add Operations"`
	_                                                                float64 `perfdata:"LDAP Add Operations/sec"`
	_                                                                float64 `perfdata:"LDAP batch slots available"`
	LdapBindTime                                                     float64 `perfdata:"LDAP Bind Time"`
	_                                                                float64 `perfdata:"LDAP busy retries"`
	_                                                                float64 `perfdata:"LDAP busy retries/sec"`
	LdapClientSessions                                               float64 `perfdata:"LDAP Client Sessions"`
	LdapClosedConnectionsPerSec                                      float64 `perfdata:"LDAP Closed Connections/sec"`
	_                                                                float64 `perfdata:"LDAP Delete Operations"`
	_                                                                float64 `perfdata:"LDAP Delete Operations/sec"`
	_                                                                float64 `perfdata:"LDAP Modify DN Operations"`
	_                                                                float64 `perfdata:"LDAP Modify DN Operations/sec"`
	_                                                                float64 `perfdata:"LDAP Modify Operations"`
	_                                                                float64 `perfdata:"LDAP Modify Operations/sec"`
	LdapNewConnectionsPerSec                                         float64 `perfdata:"LDAP New Connections/sec"`
	LdapNewSSLConnectionsPerSec                                      float64 `perfdata:"LDAP New SSL Connections/sec"`
	_                                                                float64 `perfdata:"LDAP Outbound Bytes"`
	_                                                                float64 `perfdata:"LDAP Outbound Bytes/sec"`
	_                                                                float64 `perfdata:"LDAP Page Search Cache entries count"`
	_                                                                float64 `perfdata:"LDAP Page Search Cache size"`
	LdapSearchesPerSec                                               float64 `perfdata:"LDAP Searches/sec"`
	LdapSuccessfulBindsPerSec                                        float64 `perfdata:"LDAP Successful Binds/sec"`
	_                                                                float64 `perfdata:"LDAP Threads Sleeping on BUSY"`
	LdapUDPOperationsPerSec                                          float64 `perfdata:"LDAP UDP operations/sec"`
	LdapWritesPerSec                                                 float64 `perfdata:"LDAP Writes/sec"`
	LinkValuesCleanedPerSec                                          float64 `perfdata:"Link Values Cleaned/sec"`
	_                                                                float64 `perfdata:"Links added"`
	_                                                                float64 `perfdata:"Links added/sec"`
	_                                                                float64 `perfdata:"Links visited"`
	_                                                                float64 `perfdata:"Links visited/sec"`
	_                                                                float64 `perfdata:"Logical link deletes"`
	_                                                                float64 `perfdata:"Logical link deletes/sec"`
	NegotiatedBindsPerSec                                            float64 `perfdata:"Negotiated Binds/sec"`
	NTLMBindsPerSec                                                  float64 `perfdata:"NTLM Binds/sec"`
	_                                                                float64 `perfdata:"Objects returned"`
	_                                                                float64 `perfdata:"Objects returned/sec"`
	_                                                                float64 `perfdata:"Objects visited"`
	_                                                                float64 `perfdata:"Objects visited/sec"`
	OneLevelSearchesPerSec                                           float64 `perfdata:"Onelevel searches/sec"`
	_                                                                float64 `perfdata:"PDC failed password update notifications"`
	_                                                                float64 `perfdata:"PDC password update notifications/sec"`
	_                                                                float64 `perfdata:"PDC successful password update notifications"`
	PhantomsCleanedPerSec                                            float64 `perfdata:"Phantoms Cleaned/sec"`
	PhantomsVisitedPerSec                                            float64 `perfdata:"Phantoms Visited/sec"`
	_                                                                float64 `perfdata:"Physical link deletes"`
	_                                                                float64 `perfdata:"Physical link deletes/sec"`
	_                                                                float64 `perfdata:"Replicate Single Object operations"`
	_                                                                float64 `perfdata:"Replicate Single Object operations/sec"`
	_                                                                float64 `perfdata:"RID Pool invalidations since boot"`
	_                                                                float64 `perfdata:"RID Pool request failures since boot"`
	_                                                                float64 `perfdata:"RID Pool request successes since boot"`
	SamAccountGroupEvaluationLatency                                 float64 `perfdata:"SAM Account Group Evaluation Latency"`
	SamDisplayInformationQueriesPerSec                               float64 `perfdata:"SAM Display Information Queries/sec"`
	SamDomainLocalGroupMembershipEvaluationsPerSec                   float64 `perfdata:"SAM Domain Local Group Membership Evaluations/sec"`
	SamEnumerationsPerSec                                            float64 `perfdata:"SAM Enumerations/sec"`
	SamGCEvaluationsPerSec                                           float64 `perfdata:"SAM GC Evaluations/sec"`
	SamGlobalGroupMembershipEvaluationsPerSec                        float64 `perfdata:"SAM Global Group Membership Evaluations/sec"`
	SamMachineCreationAttemptsPerSec                                 float64 `perfdata:"SAM Machine Creation Attempts/sec"`
	SamMembershipChangesPerSec                                       float64 `perfdata:"SAM Membership Changes/sec"`
	SamNonTransitiveMembershipEvaluationsPerSec                      float64 `perfdata:"SAM Non-Transitive Membership Evaluations/sec"`
	SamPasswordChangesPerSec                                         float64 `perfdata:"SAM Password Changes/sec"`
	SamResourceGroupEvaluationLatency                                float64 `perfdata:"SAM Resource Group Evaluation Latency"`
	SamSuccessfulComputerCreationsPerSecIncludesAllRequests          float64 `perfdata:"SAM Successful Computer Creations/sec: Includes all requests"`
	SamSuccessfulUserCreationsPerSec                                 float64 `perfdata:"SAM Successful User Creations/sec"`
	SamTransitiveMembershipEvaluationsPerSec                         float64 `perfdata:"SAM Transitive Membership Evaluations/sec"`
	SamUniversalGroupMembershipEvaluationsPerSec                     float64 `perfdata:"SAM Universal Group Membership Evaluations/sec"`
	SamUserCreationAttemptsPerSec                                    float64 `perfdata:"SAM User Creation Attempts/sec"`
	SimpleBindsPerSec                                                float64 `perfdata:"Simple Binds/sec"`
	SubtreeSearchesPerSec                                            float64 `perfdata:"Subtree searches/sec"`
	TombstonesGarbageCollectedPerSec                                 float64 `perfdata:"Tombstones Garbage Collected/sec"`
	TombstonesVisitedPerSec                                          float64 `perfdata:"Tombstones Visited/sec"`
	TransitiveOperationsMillisecondsRun                              float64 `perfdata:"Transitive operations milliseconds run"`
	TransitiveOperationsPerSec                                       float64 `perfdata:"Transitive operations/sec"`
	TransitiveSubOperationsPerSec                                    float64 `perfdata:"Transitive suboperations/sec"`
	_                                                                float64 `perfdata:"Warning eventlogs since boot"`
	_                                                                float64 `perfdata:"Warning events since boot"`
}
