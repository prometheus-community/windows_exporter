package ad

/*
DRAInboundFullSyncObjectsRemaining
DRAInboundLinkValueUpdatesRemaininginPacket
DRAInboundObjectsAppliedPerSec
DRAInboundObjectsFilteredPerSec
DRAInboundObjectsPerSec
DRAInboundObjectUpdatesRemaininginPacket
DRAInboundPropertiesAppliedPerSec
DRAInboundPropertiesFilteredPerSec
DRAInboundPropertiesTotalPerSec
DRAInboundTotalUpdatesRemaininginPacket
DRAInboundValuesDNsonlyPerSec
DRAInboundValuesTotalPerSec
DRAOutboundBytesCompressedBetweenSitesAfterCompressionPerSec
DRAOutboundBytesCompressedBetweenSitesAfterCompressionSinceBoot
DRAOutboundBytesCompressedBetweenSitesBeforeCompressionPerSec
DRAOutboundBytesCompressedBetweenSitesBeforeCompressionSinceBoot
DRAOutboundBytesNotCompressedWithinSitePerSec
DRAOutboundBytesNotCompressedWithinSiteSinceBoot
DRAOutboundBytesTotalPerSec
DRAOutboundBytesTotalSinceBoot
DRAOutboundObjectsFilteredPerSec
DRAOutboundObjectsPerSec
DRAOutboundPropertiesPerSec
DRAOutboundValuesDNsonlyPerSec
DRAOutboundValuesTotalPerSec
DRAPendingReplicationOperations
DRAPendingReplicationSynchronizations
DRASyncFailuresonSchemaMismatch
DRASyncRequestsMade
DRASyncRequestsSuccessful
DRAThreadsGettingNCChanges
DRAThreadsGettingNCChangesHoldingSemaphore
DSClientBindsPerSec
DSClientNameTranslationsPerSec
DSDirectoryReadsPerSec
DSDirectorySearchesPerSec
DSDirectoryWritesPerSec
DSMonitorListSize
DSNameCachehitrate
DSNameCachehitrate_Base
DSNotifyQueueSize
DSPercentReadsfromDRA
DSPercentReadsfromKCC
DSPercentReadsfromLSA
DSPercentReadsfromNSPI
DSPercentReadsfromNTDSAPI
DSPercentReadsfromSAM
DSPercentReadsOther
DSPercentSearchesfromDRA
DSPercentSearchesfromKCC
DSPercentSearchesfromLDAP
DSPercentSearchesfromLSA
DSPercentSearchesfromNSPI
DSPercentSearchesfromNTDSAPI
DSPercentSearchesfromSAM
DSPercentSearchesOther
DSPercentWritesfromDRA
DSPercentWritesfromKCC
DSPercentWritesfromLDAP
DSPercentWritesfromLSA
DSPercentWritesfromNSPI
DSPercentWritesfromNTDSAPI
DSPercentWritesfromSAM
DSPercentWritesOther
DSSearchsuboperationsPerSec
DSSecurityDescriptorPropagationsEvents
DSSecurityDescriptorPropagatorAverageExclusionTime
DSSecurityDescriptorPropagatorRuntimeQueue
DSSecurityDescriptorsuboperationsPerSec
DSServerBindsPerSec
DSServerNameTranslationsPerSec
DSThreadsinUse
ExternalBindsPerSec
FastBindsPerSec
LDAPActiveThreads
LDAPBindTime
LDAPClientSessions
LDAPClosedConnectionsPerSec
LDAPNewConnectionsPerSec
LDAPNewSSLConnectionsPerSec
LDAPSearchesPerSec
LDAPSuccessfulBindsPerSec
LDAPUDPoperationsPerSec
LDAPWritesPerSec
LinkValuesCleanedPerSec
NegotiatedBindsPerSec
NTLMBindsPerSec
OnelevelsearchesPerSec
PhantomsCleanedPerSec
PhantomsVisitedPerSec
SAMAccountGroupEvaluationLatency
SAMDisplayInformationQueriesPerSec
SAMDomainLocalGroupMembershipEvaluationsPerSec
SAMEnumerationsPerSec
SAMGCEvaluationsPerSec
SAMGlobalGroupMembershipEvaluationsPerSec
SAMMachineCreationAttemptsPerSec
SAMMembershipChangesPerSec
SAMNonTransitiveMembershipEvaluationsPerSec
SAMPasswordChangesPerSec
SAMResourceGroupEvaluationLatency
SAMSuccessfulComputerCreationsPerSecIncludesallrequests
SAMSuccessfulUserCreationsPerSec
SAMTransitiveMembershipEvaluationsPerSec
SAMUniversalGroupMembershipEvaluationsPerSec
SAMUserCreationAttemptsPerSec
SimpleBindsPerSec
SubtreesearchesPerSec
TombstonesGarbageCollectedPerSec
TombstonesVisitedPerSec
Transitiveoperationsmillisecondsrun
TransitiveoperationsPerSec
TransitivesuboperationsPerSec
*/
const (
	ABANRPerSec                                                     = "AB ANR/sec"
	ABBrowsesPerSec                                                 = "AB Browses/sec"
	ABClientSessions                                                = "AB Client Sessions"
	ABMatchesPerSec                                                 = "AB Matches/sec"
	ABPropertyReadsPerSec                                           = "AB Property Reads/sec"
	ABProxyLookupsPerSec                                            = "AB Proxy Lookups/sec"
	ABSearchesPerSec                                                = "AB Searches/sec"
	ApproximateHighestDNT                                           = "Approximate highest DNT"
	ATQEstimatedQueueDelay                                          = "ATQ Estimated Queue Delay"
	ATQOutstandingQueuedRequests                                    = "ATQ Outstanding Queued Requests"
	_                                                               = "ATQ Queue Latency"
	ATQRequestLatency                                               = "ATQ Request Latency"
	ATQThreadsLDAP                                                  = "ATQ Threads LDAP"
	ATQThreadsOther                                                 = "ATQ Threads Other"
	ATQThreadsTotal                                                 = "ATQ Threads Total"
	BaseSearchesPerSec                                              = "Base searches/sec"
	DatabaseAddsPerSec                                              = "Database adds/sec"
	DatabaseDeletesPerSec                                           = "Database deletes/sec"
	DatabaseModifysPerSec                                           = "Database modifys/sec"
	DatabaseRecyclesPerSec                                          = "Database recycles/sec"
	DigestBindsPerSec                                               = "Digest Binds/sec"
	_                                                               = "DirSync session throttling rate"
	_                                                               = "DirSync sessions in progress"
	DRAHighestUSNCommittedHighPart                                  = "DRA Highest USN Committed (High part)"
	DRAHighestUSNCommittedLowPart                                   = "DRA Highest USN Committed (Low part)"
	DRAHighestUSNIssuedHighPart                                     = "DRA Highest USN Issued (High part)"
	DRAHighestUSNIssuedLowPart                                      = "DRA Highest USN Issued (Low part)"
	DRAInboundBytesCompressedBetweenSitesAfterCompressionSinceBoot  = "DRA Inbound Bytes Compressed (Between Sites, After Compression) Since Boot"
	DRAInboundBytesCompressedBetweenSitesAfterCompressionPerSec     = "DRA Inbound Bytes Compressed (Between Sites, After Compression)/sec"
	DRAInboundBytesCompressedBetweenSitesBeforeCompressionSinceBoot = "DRA Inbound Bytes Compressed (Between Sites, Before Compression) Since Boot"
	DRAInboundBytesCompressedBetweenSitesBeforeCompressionPerSec    = "DRA Inbound Bytes Compressed (Between Sites, Before Compression)/sec"
	DRAInboundBytesNotCompressedWithinSiteSinceBoot                 = "DRA Inbound Bytes Not Compressed (Within Site) Since Boot"
	DRAInboundBytesNotCompressedWithinSitePerSec                    = "DRA Inbound Bytes Not Compressed (Within Site)/sec"
	DRAInboundBytesTotalSinceBoot                                   = "DRA Inbound Bytes Total Since Boot"
	DRAInboundBytesTotalPerSec                                      = "DRA Inbound Bytes Total/sec"
	_                                                               = "DRA Inbound Full Sync Objects Remaining"
	_                                                               = "DRA Inbound Link Value Updates Remaining in Packet"
	_                                                               = "DRA Inbound Link Values/sec"
	_                                                               = "DRA Inbound Object Updates Remaining in Packet"
	_                                                               = "DRA Inbound Objects Applied/sec"
	_                                                               = "DRA Inbound Objects Filtered/sec"
	_                                                               = "DRA Inbound Objects/sec"
	_                                                               = "DRA Inbound Properties Applied/sec"
	_                                                               = "DRA Inbound Properties Filtered/sec"
	_                                                               = "DRA Inbound Properties Total/sec"
	_                                                               = "DRA Inbound Sync Link Deletion/sec"
	_                                                               = "DRA Inbound Total Updates Remaining in Packet"
	_                                                               = "DRA Inbound Values (DNs only)/sec"
	_                                                               = "DRA Inbound Values Total/sec"
	_                                                               = "DRA number of NC replication calls since boot"
	_                                                               = "DRA number of successful NC replication calls since boot"
	_                                                               = "DRA Outbound Bytes Compressed (Between Sites, After Compression) Since Boot"
	_                                                               = "DRA Outbound Bytes Compressed (Between Sites, After Compression)/sec"
	_                                                               = "DRA Outbound Bytes Compressed (Between Sites, Before Compression) Since Boot"
	_                                                               = "DRA Outbound Bytes Compressed (Between Sites, Before Compression)/sec"
	_                                                               = "DRA Outbound Bytes Not Compressed (Within Site) Since Boot"
	_                                                               = "DRA Outbound Bytes Not Compressed (Within Site)/sec"
	_                                                               = "DRA Outbound Bytes Total Since Boot"
	_                                                               = "DRA Outbound Bytes Total/sec"
	_                                                               = "DRA Outbound Objects Filtered/sec"
	_                                                               = "DRA Outbound Objects/sec"
	_                                                               = "DRA Outbound Properties/sec"
	_                                                               = "DRA Outbound Values (DNs only)/sec"
	_                                                               = "DRA Outbound Values Total/sec"
	_                                                               = "DRA Pending Replication Operations"
	_                                                               = "DRA Pending Replication Synchronizations"
	_                                                               = "DRA Sync Failures on Schema Mismatch"
	_                                                               = "DRA Sync Requests Made"
	_                                                               = "DRA Sync Requests Successful"
	_                                                               = "DRA Threads Getting NC Changes"
	_                                                               = "DRA Threads Getting NC Changes Holding Semaphore"
	_                                                               = "DRA total number of Busy failures since boot"
	_                                                               = "DRA total number of MissingParent failures since boot"
	_                                                               = "DRA total number of NotEnoughAttrs/MissingObject failures since boot"
	_                                                               = "DRA total number of Preempted failures since boot"
	_                                                               = "DRA total time of applying replication package since boot"
	_                                                               = "DRA total time of NC replication calls since boot"
	_                                                               = "DRA total time of successful NC replication calls since boot"
	_                                                               = "DRA total time of successfully applying replication package since boot"
	_                                                               = "DRA total time on waiting async replication packages since boot"
	_                                                               = "DRA total time on waiting sync replication packages since boot"
	_                                                               = "DS % Reads from DRA"
	_                                                               = "DS % Reads from KCC"
	_                                                               = "DS % Reads from LSA"
	_                                                               = "DS % Reads from NSPI"
	_                                                               = "DS % Reads from NTDSAPI"
	_                                                               = "DS % Reads from SAM"
	_                                                               = "DS % Reads Other"
	_                                                               = "DS % Searches from DRA"
	_                                                               = "DS % Searches from KCC"
	_                                                               = "DS % Searches from LDAP"
	_                                                               = "DS % Searches from LSA"
	_                                                               = "DS % Searches from NSPI"
	_                                                               = "DS % Searches from NTDSAPI"
	_                                                               = "DS % Searches from SAM"
	_                                                               = "DS % Searches Other"
	_                                                               = "DS % Writes from DRA"
	_                                                               = "DS % Writes from KCC"
	_                                                               = "DS % Writes from LDAP"
	_                                                               = "DS % Writes from LSA"
	_                                                               = "DS % Writes from NSPI"
	_                                                               = "DS % Writes from NTDSAPI"
	_                                                               = "DS % Writes from SAM"
	_                                                               = "DS % Writes Other"
	_                                                               = "DS Client Binds/sec"
	_                                                               = "DS Client Name Translations/sec"
	_                                                               = "DS Directory Reads/sec"
	_                                                               = "DS Directory Searches/sec"
	_                                                               = "DS Directory Writes/sec"
	_                                                               = "DS Monitor List Size"
	_                                                               = "DS Name Cache hit rate"
	_                                                               = "DS Notify Queue Size"
	_                                                               = "DS Search sub-operations/sec"
	_                                                               = "DS Security Descriptor Propagations Events"
	_                                                               = "DS Security Descriptor Propagator Average Exclusion Time"
	_                                                               = "DS Security Descriptor Propagator Runtime Queue"
	_                                                               = "DS Security Descriptor sub-operations/sec"
	_                                                               = "DS Server Binds/sec"
	_                                                               = "DS Server Name Translations/sec"
	_                                                               = "DS Threads in Use"
	_                                                               = "Error eventlogs since boot"
	_                                                               = "Error events since boot"
	_                                                               = "External Binds/sec"
	_                                                               = "Fast Binds/sec"
	_                                                               = "Fatal events since boot"
	_                                                               = "Info eventlogs since boot"
	_                                                               = "LDAP Active Threads"
	_                                                               = "LDAP Add Operations"
	_                                                               = "LDAP Add Operations/sec"
	_                                                               = "LDAP batch slots available"
	_                                                               = "LDAP Bind Time"
	_                                                               = "LDAP busy retries"
	_                                                               = "LDAP busy retries/sec"
	_                                                               = "LDAP Client Sessions"
	_                                                               = "LDAP Closed Connections/sec"
	_                                                               = "LDAP Delete Operations"
	_                                                               = "LDAP Delete Operations/sec"
	_                                                               = "LDAP Modify DN Operations"
	_                                                               = "LDAP Modify DN Operations/sec"
	_                                                               = "LDAP Modify Operations"
	_                                                               = "LDAP Modify Operations/sec"
	_                                                               = "LDAP New Connections/sec"
	_                                                               = "LDAP New SSL Connections/sec"
	_                                                               = "LDAP Outbound Bytes"
	_                                                               = "LDAP Outbound Bytes/sec"
	_                                                               = "LDAP Page Search Cache entries count"
	_                                                               = "LDAP Page Search Cache size"
	_                                                               = "LDAP Searches/sec"
	_                                                               = "LDAP Successful Binds/sec"
	_                                                               = "LDAP Threads Sleeping on BUSY"
	_                                                               = "LDAP UDP operations/sec"
	_                                                               = "LDAP Writes/sec"
	_                                                               = "Link Values Cleaned/sec"
	_                                                               = "Links added"
	_                                                               = "Links added/sec"
	_                                                               = "Links visited"
	_                                                               = "Links visited/sec"
	_                                                               = "Logical link deletes"
	_                                                               = "Logical link deletes/sec"
	_                                                               = "Negotiated Binds/sec"
	_                                                               = "NTLM Binds/sec"
	_                                                               = "Objects returned"
	_                                                               = "Objects returned/sec"
	_                                                               = "Objects visited"
	_                                                               = "Objects visited/sec"
	_                                                               = "Onelevel searches/sec"
	_                                                               = "PDC failed password update notifications"
	_                                                               = "PDC password update notifications/sec"
	_                                                               = "PDC successful password update notifications"
	_                                                               = "Phantoms Cleaned/sec"
	_                                                               = "Phantoms Visited/sec"
	_                                                               = "Physical link deletes"
	_                                                               = "Physical link deletes/sec"
	_                                                               = "Replicate Single Object operations"
	_                                                               = "Replicate Single Object operations/sec"
	_                                                               = "RID Pool invalidations since boot"
	_                                                               = "RID Pool request failures since boot"
	_                                                               = "RID Pool request successes since boot"
	_                                                               = "SAM Account Group Evaluation Latency"
	_                                                               = "SAM Display Information Queries/sec"
	_                                                               = "SAM Domain Local Group Membership Evaluations/sec"
	_                                                               = "SAM Enumerations/sec"
	_                                                               = "SAM GC Evaluations/sec"
	_                                                               = "SAM Global Group Membership Evaluations/sec"
	_                                                               = "SAM Machine Creation Attempts/sec"
	_                                                               = "SAM Membership Changes/sec"
	_                                                               = "SAM Non-Transitive Membership Evaluations/sec"
	_                                                               = "SAM Password Changes/sec"
	_                                                               = "SAM Resource Group Evaluation Latency"
	_                                                               = "SAM Successful Computer Creations/sec: Includes all requests"
	_                                                               = "SAM Successful User Creations/sec"
	_                                                               = "SAM Transitive Membership Evaluations/sec"
	_                                                               = "SAM Universal Group Membership Evaluations/sec"
	_                                                               = "SAM User Creation Attempts/sec"
	_                                                               = "Simple Binds/sec"
	_                                                               = "Subtree searches/sec"
	_                                                               = "Tombstones Garbage Collected/sec"
	_                                                               = "Tombstones Visited/sec"
	_                                                               = "Transitive operations milliseconds run"
	_                                                               = "Transitive operations/sec"
	_                                                               = "Transitive suboperations/sec"
	_                                                               = "Warning eventlogs since boot"
	_                                                               = "Warning events since boot"
)
