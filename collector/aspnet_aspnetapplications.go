// returns data points from Win32_PerfFormattedData_ASPNET_ASPNETApplications
// <add link to documentation here> - Win32_PerfFormattedData_ASPNET_ASPNETApplications class
package collector

import (
	"github.com/StackExchange/wmi"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
)

func init() {
	Factories["aspnet_aspnetapplications"] = Newaspnet_aspnetapplicationsCollector
}

// A aspnet_aspnetapplicationsCollector is a Prometheus collector for WMI Win32_PerfFormattedData_ASPNET_ASPNETApplications metrics
type aspnet_aspnetapplicationsCollector struct {
	AnonymousRequests                     *prometheus.Desc
	AnonymousRequestsPerSec               *prometheus.Desc
	ApplicationLifetimeEvents             *prometheus.Desc
	ApplicationLifetimeEventsPerSec       *prometheus.Desc
	AuditFailureEventsRaised              *prometheus.Desc
	AuditSuccessEventsRaised              *prometheus.Desc
	CacheAPIEntries                       *prometheus.Desc
	CacheAPIHitRatio                      *prometheus.Desc
	CacheAPIHits                          *prometheus.Desc
	CacheAPIMisses                        *prometheus.Desc
	CacheAPITrims                         *prometheus.Desc
	CacheAPITurnoverRate                  *prometheus.Desc
	CachePercentMachineMemoryLimitUsed    *prometheus.Desc
	CachePercentProcessMemoryLimitUsed    *prometheus.Desc
	CacheTotalEntries                     *prometheus.Desc
	CacheTotalHitRatio                    *prometheus.Desc
	CacheTotalHits                        *prometheus.Desc
	CacheTotalMisses                      *prometheus.Desc
	CacheTotalTrims                       *prometheus.Desc
	CacheTotalTurnoverRate                *prometheus.Desc
	CompilationsTotal                     *prometheus.Desc
	DebuggingRequests                     *prometheus.Desc
	ErrorEventsRaised                     *prometheus.Desc
	ErrorEventsRaisedPerSec               *prometheus.Desc
	ErrorsDuringCompilation               *prometheus.Desc
	ErrorsDuringExecution                 *prometheus.Desc
	ErrorsDuringPreprocessing             *prometheus.Desc
	ErrorsTotal                           *prometheus.Desc
	ErrorsTotalPerSec                     *prometheus.Desc
	ErrorsUnhandledDuringExecution        *prometheus.Desc
	ErrorsUnhandledDuringExecutionPerSec  *prometheus.Desc
	EventsRaised                          *prometheus.Desc
	EventsRaisedPerSec                    *prometheus.Desc
	FormsAuthenticationFailure            *prometheus.Desc
	FormsAuthenticationSuccess            *prometheus.Desc
	InfrastructureErrorEventsRaised       *prometheus.Desc
	InfrastructureErrorEventsRaisedPerSec *prometheus.Desc
	ManagedMemoryUsedestimated            *prometheus.Desc
	MembershipAuthenticationFailure       *prometheus.Desc
	MembershipAuthenticationSuccess       *prometheus.Desc
	OutputCacheEntries                    *prometheus.Desc
	OutputCacheHitRatio                   *prometheus.Desc
	OutputCacheHits                       *prometheus.Desc
	OutputCacheMisses                     *prometheus.Desc
	OutputCacheTrims                      *prometheus.Desc
	OutputCacheTurnoverRate               *prometheus.Desc
	PercentManagedProcessorTimeestimated  *prometheus.Desc
	PipelineInstanceCount                 *prometheus.Desc
	RequestBytesInTotal                   *prometheus.Desc
	RequestBytesInTotalWebSockets         *prometheus.Desc
	RequestBytesOutTotal                  *prometheus.Desc
	RequestBytesOutTotalWebSockets        *prometheus.Desc
	RequestErrorEventsRaised              *prometheus.Desc
	RequestErrorEventsRaisedPerSec        *prometheus.Desc
	RequestEventsRaised                   *prometheus.Desc
	RequestEventsRaisedPerSec             *prometheus.Desc
	RequestExecutionTime                  *prometheus.Desc
	RequestsDisconnected                  *prometheus.Desc
	RequestsExecuting                     *prometheus.Desc
	RequestsExecutingWebSockets           *prometheus.Desc
	RequestsFailed                        *prometheus.Desc
	RequestsFailedWebSockets              *prometheus.Desc
	RequestsInApplicationQueue            *prometheus.Desc
	RequestsNotAuthorized                 *prometheus.Desc
	RequestsNotFound                      *prometheus.Desc
	RequestsPerSec                        *prometheus.Desc
	RequestsRejected                      *prometheus.Desc
	RequestsSucceeded                     *prometheus.Desc
	RequestsSucceededWebSockets           *prometheus.Desc
	RequestsTimedOut                      *prometheus.Desc
	RequestsTotal                         *prometheus.Desc
	RequestsTotalWebSockets               *prometheus.Desc
	RequestWaitTime                       *prometheus.Desc
	SessionsAbandoned                     *prometheus.Desc
	SessionsActive                        *prometheus.Desc
	SessionSQLServerconnectionstotal      *prometheus.Desc
	SessionStateServerconnectionstotal    *prometheus.Desc
	SessionsTimedOut                      *prometheus.Desc
	SessionsTotal                         *prometheus.Desc
	TransactionsAborted                   *prometheus.Desc
	TransactionsCommitted                 *prometheus.Desc
	TransactionsPending                   *prometheus.Desc
	TransactionsPerSec                    *prometheus.Desc
	TransactionsTotal                     *prometheus.Desc
	ViewstateMACValidationFailure         *prometheus.Desc
}

// Newaspnet_aspnetapplicationsCollector ...
func Newaspnet_aspnetapplicationsCollector() (Collector, error) {
	const subsystem = "aspnet_aspnetapplications"
	return &aspnet_aspnetapplicationsCollector{
		AnonymousRequests: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "anonymous_requests"),
			"(AnonymousRequests)",
			nil,
			nil,
		),
		AnonymousRequestsPerSec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "anonymous_requests_per_sec"),
			"(AnonymousRequestsPerSec)",
			nil,
			nil,
		),
		ApplicationLifetimeEvents: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "application_lifetime_events"),
			"(ApplicationLifetimeEvents)",
			nil,
			nil,
		),
		ApplicationLifetimeEventsPerSec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "application_lifetime_events_per_sec"),
			"(ApplicationLifetimeEventsPerSec)",
			nil,
			nil,
		),
		AuditFailureEventsRaised: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "audit_failure_events_raised"),
			"(AuditFailureEventsRaised)",
			nil,
			nil,
		),
		AuditSuccessEventsRaised: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "audit_success_events_raised"),
			"(AuditSuccessEventsRaised)",
			nil,
			nil,
		),
		CacheAPIEntries: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "cache_api_entries"),
			"(CacheAPIEntries)",
			nil,
			nil,
		),
		CacheAPIHitRatio: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "cache_api_hit_ratio"),
			"(CacheAPIHitRatio)",
			nil,
			nil,
		),
		CacheAPIHits: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "cache_api_hits"),
			"(CacheAPIHits)",
			nil,
			nil,
		),
		CacheAPIMisses: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "cache_api_misses"),
			"(CacheAPIMisses)",
			nil,
			nil,
		),
		CacheAPITrims: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "cache_api_trims"),
			"(CacheAPITrims)",
			nil,
			nil,
		),
		CacheAPITurnoverRate: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "cache_api_turnover_rate"),
			"(CacheAPITurnoverRate)",
			nil,
			nil,
		),
		CachePercentMachineMemoryLimitUsed: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "cache_percent_machine_memory_limit_used"),
			"(CachePercentMachineMemoryLimitUsed)",
			nil,
			nil,
		),
		CachePercentProcessMemoryLimitUsed: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "cache_percent_process_memory_limit_used"),
			"(CachePercentProcessMemoryLimitUsed)",
			nil,
			nil,
		),
		CacheTotalEntries: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "cache_total_entries"),
			"(CacheTotalEntries)",
			nil,
			nil,
		),
		CacheTotalHitRatio: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "cache_total_hit_ratio"),
			"(CacheTotalHitRatio)",
			nil,
			nil,
		),
		CacheTotalHits: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "cache_total_hits"),
			"(CacheTotalHits)",
			nil,
			nil,
		),
		CacheTotalMisses: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "cache_total_misses"),
			"(CacheTotalMisses)",
			nil,
			nil,
		),
		CacheTotalTrims: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "cache_total_trims"),
			"(CacheTotalTrims)",
			nil,
			nil,
		),
		CacheTotalTurnoverRate: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "cache_total_turnover_rate"),
			"(CacheTotalTurnoverRate)",
			nil,
			nil,
		),
		CompilationsTotal: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "compilations_total"),
			"(CompilationsTotal)",
			nil,
			nil,
		),
		DebuggingRequests: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "debugging_requests"),
			"(DebuggingRequests)",
			nil,
			nil,
		),
		ErrorEventsRaised: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "error_events_raised"),
			"(ErrorEventsRaised)",
			nil,
			nil,
		),
		ErrorEventsRaisedPerSec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "error_events_raised_per_sec"),
			"(ErrorEventsRaisedPerSec)",
			nil,
			nil,
		),
		ErrorsDuringCompilation: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "errors_during_compilation"),
			"(ErrorsDuringCompilation)",
			nil,
			nil,
		),
		ErrorsDuringExecution: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "errors_during_execution"),
			"(ErrorsDuringExecution)",
			nil,
			nil,
		),
		ErrorsDuringPreprocessing: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "errors_during_preprocessing"),
			"(ErrorsDuringPreprocessing)",
			nil,
			nil,
		),
		ErrorsTotal: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "errors_total"),
			"(ErrorsTotal)",
			nil,
			nil,
		),
		ErrorsTotalPerSec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "errors_total_per_sec"),
			"(ErrorsTotalPerSec)",
			nil,
			nil,
		),
		ErrorsUnhandledDuringExecution: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "errors_unhandled_during_execution"),
			"(ErrorsUnhandledDuringExecution)",
			nil,
			nil,
		),
		ErrorsUnhandledDuringExecutionPerSec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "errors_unhandled_during_execution_per_sec"),
			"(ErrorsUnhandledDuringExecutionPerSec)",
			nil,
			nil,
		),
		EventsRaised: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "events_raised"),
			"(EventsRaised)",
			nil,
			nil,
		),
		EventsRaisedPerSec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "events_raised_per_sec"),
			"(EventsRaisedPerSec)",
			nil,
			nil,
		),
		FormsAuthenticationFailure: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "forms_authentication_failure"),
			"(FormsAuthenticationFailure)",
			nil,
			nil,
		),
		FormsAuthenticationSuccess: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "forms_authentication_success"),
			"(FormsAuthenticationSuccess)",
			nil,
			nil,
		),
		InfrastructureErrorEventsRaised: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "infrastructure_error_events_raised"),
			"(InfrastructureErrorEventsRaised)",
			nil,
			nil,
		),
		InfrastructureErrorEventsRaisedPerSec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "infrastructure_error_events_raised_per_sec"),
			"(InfrastructureErrorEventsRaisedPerSec)",
			nil,
			nil,
		),
		ManagedMemoryUsedestimated: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "managed_memory_usedestimated"),
			"(ManagedMemoryUsedestimated)",
			nil,
			nil,
		),
		MembershipAuthenticationFailure: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "membership_authentication_failure"),
			"(MembershipAuthenticationFailure)",
			nil,
			nil,
		),
		MembershipAuthenticationSuccess: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "membership_authentication_success"),
			"(MembershipAuthenticationSuccess)",
			nil,
			nil,
		),
		OutputCacheEntries: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "output_cache_entries"),
			"(OutputCacheEntries)",
			nil,
			nil,
		),
		OutputCacheHitRatio: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "output_cache_hit_ratio"),
			"(OutputCacheHitRatio)",
			nil,
			nil,
		),
		OutputCacheHits: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "output_cache_hits"),
			"(OutputCacheHits)",
			nil,
			nil,
		),
		OutputCacheMisses: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "output_cache_misses"),
			"(OutputCacheMisses)",
			nil,
			nil,
		),
		OutputCacheTrims: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "output_cache_trims"),
			"(OutputCacheTrims)",
			nil,
			nil,
		),
		OutputCacheTurnoverRate: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "output_cache_turnover_rate"),
			"(OutputCacheTurnoverRate)",
			nil,
			nil,
		),
		PercentManagedProcessorTimeestimated: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "percent_managed_processor_timeestimated"),
			"(PercentManagedProcessorTimeestimated)",
			nil,
			nil,
		),
		PipelineInstanceCount: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "pipeline_instance_count"),
			"(PipelineInstanceCount)",
			nil,
			nil,
		),
		RequestBytesInTotal: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "request_bytes_in_total"),
			"(RequestBytesInTotal)",
			nil,
			nil,
		),
		RequestBytesInTotalWebSockets: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "request_bytes_in_total_web_sockets"),
			"(RequestBytesInTotalWebSockets)",
			nil,
			nil,
		),
		RequestBytesOutTotal: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "request_bytes_out_total"),
			"(RequestBytesOutTotal)",
			nil,
			nil,
		),
		RequestBytesOutTotalWebSockets: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "request_bytes_out_total_web_sockets"),
			"(RequestBytesOutTotalWebSockets)",
			nil,
			nil,
		),
		RequestErrorEventsRaised: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "request_error_events_raised"),
			"(RequestErrorEventsRaised)",
			nil,
			nil,
		),
		RequestErrorEventsRaisedPerSec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "request_error_events_raised_per_sec"),
			"(RequestErrorEventsRaisedPerSec)",
			nil,
			nil,
		),
		RequestEventsRaised: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "request_events_raised"),
			"(RequestEventsRaised)",
			nil,
			nil,
		),
		RequestEventsRaisedPerSec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "request_events_raised_per_sec"),
			"(RequestEventsRaisedPerSec)",
			nil,
			nil,
		),
		RequestExecutionTime: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "request_execution_time"),
			"(RequestExecutionTime)",
			nil,
			nil,
		),
		RequestsDisconnected: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "requests_disconnected"),
			"(RequestsDisconnected)",
			nil,
			nil,
		),
		RequestsExecuting: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "requests_executing"),
			"(RequestsExecuting)",
			nil,
			nil,
		),
		RequestsExecutingWebSockets: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "requests_executing_web_sockets"),
			"(RequestsExecutingWebSockets)",
			nil,
			nil,
		),
		RequestsFailed: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "requests_failed"),
			"(RequestsFailed)",
			nil,
			nil,
		),
		RequestsFailedWebSockets: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "requests_failed_web_sockets"),
			"(RequestsFailedWebSockets)",
			nil,
			nil,
		),
		RequestsInApplicationQueue: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "requests_in_application_queue"),
			"(RequestsInApplicationQueue)",
			nil,
			nil,
		),
		RequestsNotAuthorized: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "requests_not_authorized"),
			"(RequestsNotAuthorized)",
			nil,
			nil,
		),
		RequestsNotFound: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "requests_not_found"),
			"(RequestsNotFound)",
			nil,
			nil,
		),
		RequestsPerSec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "requests_per_sec"),
			"(RequestsPerSec)",
			nil,
			nil,
		),
		RequestsRejected: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "requests_rejected"),
			"(RequestsRejected)",
			nil,
			nil,
		),
		RequestsSucceeded: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "requests_succeeded"),
			"(RequestsSucceeded)",
			nil,
			nil,
		),
		RequestsSucceededWebSockets: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "requests_succeeded_web_sockets"),
			"(RequestsSucceededWebSockets)",
			nil,
			nil,
		),
		RequestsTimedOut: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "requests_timed_out"),
			"(RequestsTimedOut)",
			nil,
			nil,
		),
		RequestsTotal: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "requests_total"),
			"(RequestsTotal)",
			nil,
			nil,
		),
		RequestsTotalWebSockets: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "requests_total_web_sockets"),
			"(RequestsTotalWebSockets)",
			nil,
			nil,
		),
		RequestWaitTime: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "request_wait_time"),
			"(RequestWaitTime)",
			nil,
			nil,
		),
		SessionsAbandoned: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "sessions_abandoned"),
			"(SessionsAbandoned)",
			nil,
			nil,
		),
		SessionsActive: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "sessions_active"),
			"(SessionsActive)",
			nil,
			nil,
		),
		SessionSQLServerconnectionstotal: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "session_sql_serverconnectionstotal"),
			"(SessionSQLServerconnectionstotal)",
			nil,
			nil,
		),
		SessionStateServerconnectionstotal: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "session_state_serverconnectionstotal"),
			"(SessionStateServerconnectionstotal)",
			nil,
			nil,
		),
		SessionsTimedOut: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "sessions_timed_out"),
			"(SessionsTimedOut)",
			nil,
			nil,
		),
		SessionsTotal: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "sessions_total"),
			"(SessionsTotal)",
			nil,
			nil,
		),
		TransactionsAborted: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "transactions_aborted"),
			"(TransactionsAborted)",
			nil,
			nil,
		),
		TransactionsCommitted: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "transactions_committed"),
			"(TransactionsCommitted)",
			nil,
			nil,
		),
		TransactionsPending: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "transactions_pending"),
			"(TransactionsPending)",
			nil,
			nil,
		),
		TransactionsPerSec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "transactions_per_sec"),
			"(TransactionsPerSec)",
			nil,
			nil,
		),
		TransactionsTotal: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "transactions_total"),
			"(TransactionsTotal)",
			nil,
			nil,
		),
		ViewstateMACValidationFailure: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "viewstate_mac_validation_failure"),
			"(ViewstateMACValidationFailure)",
			nil,
			nil,
		),
	}, nil
}

// Collect sends the metric values for each metric
// to the provided prometheus Metric channel.
func (c *aspnet_aspnetapplicationsCollector) Collect(ch chan<- prometheus.Metric) error {
	if desc, err := c.collect(ch); err != nil {
		log.Error("failed collecting aspnet_aspnetapplications metrics:", desc, err)
		return err
	}
	return nil
}

type Win32_PerfFormattedData_ASPNET_ASPNETApplications struct {
	Name string

	AnonymousRequests                     uint32
	AnonymousRequestsPerSec               uint32
	ApplicationLifetimeEvents             uint32
	ApplicationLifetimeEventsPerSec       uint32
	AuditFailureEventsRaised              uint32
	AuditSuccessEventsRaised              uint32
	CacheAPIEntries                       uint32
	CacheAPIHitRatio                      uint32
	CacheAPIHits                          uint32
	CacheAPIMisses                        uint32
	CacheAPITrims                         uint32
	CacheAPITurnoverRate                  uint32
	CachePercentMachineMemoryLimitUsed    uint32
	CachePercentProcessMemoryLimitUsed    uint32
	CacheTotalEntries                     uint32
	CacheTotalHitRatio                    uint32
	CacheTotalHits                        uint32
	CacheTotalMisses                      uint32
	CacheTotalTrims                       uint32
	CacheTotalTurnoverRate                uint32
	CompilationsTotal                     uint32
	DebuggingRequests                     uint32
	ErrorEventsRaised                     uint32
	ErrorEventsRaisedPerSec               uint32
	ErrorsDuringCompilation               uint32
	ErrorsDuringExecution                 uint32
	ErrorsDuringPreprocessing             uint32
	ErrorsTotal                           uint32
	ErrorsTotalPerSec                     uint32
	ErrorsUnhandledDuringExecution        uint32
	ErrorsUnhandledDuringExecutionPerSec  uint32
	EventsRaised                          uint32
	EventsRaisedPerSec                    uint32
	FormsAuthenticationFailure            uint32
	FormsAuthenticationSuccess            uint32
	InfrastructureErrorEventsRaised       uint32
	InfrastructureErrorEventsRaisedPerSec uint32
	ManagedMemoryUsedestimated            uint32
	MembershipAuthenticationFailure       uint32
	MembershipAuthenticationSuccess       uint32
	OutputCacheEntries                    uint32
	OutputCacheHitRatio                   uint32
	OutputCacheHits                       uint32
	OutputCacheMisses                     uint32
	OutputCacheTrims                      uint32
	OutputCacheTurnoverRate               uint32
	PercentManagedProcessorTimeestimated  uint32
	PipelineInstanceCount                 uint32
	RequestBytesInTotal                   uint32
	RequestBytesInTotalWebSockets         uint32
	RequestBytesOutTotal                  uint32
	RequestBytesOutTotalWebSockets        uint32
	RequestErrorEventsRaised              uint32
	RequestErrorEventsRaisedPerSec        uint32
	RequestEventsRaised                   uint32
	RequestEventsRaisedPerSec             uint32
	RequestExecutionTime                  uint32
	RequestsDisconnected                  uint32
	RequestsExecuting                     uint32
	RequestsExecutingWebSockets           uint32
	RequestsFailed                        uint32
	RequestsFailedWebSockets              uint32
	RequestsInApplicationQueue            uint32
	RequestsNotAuthorized                 uint32
	RequestsNotFound                      uint32
	RequestsPerSec                        uint32
	RequestsRejected                      uint32
	RequestsSucceeded                     uint32
	RequestsSucceededWebSockets           uint32
	RequestsTimedOut                      uint32
	RequestsTotal                         uint32
	RequestsTotalWebSockets               uint32
	RequestWaitTime                       uint32
	SessionsAbandoned                     uint32
	SessionsActive                        uint32
	SessionSQLServerconnectionstotal      uint32
	SessionStateServerconnectionstotal    uint32
	SessionsTimedOut                      uint32
	SessionsTotal                         uint32
	TransactionsAborted                   uint32
	TransactionsCommitted                 uint32
	TransactionsPending                   uint32
	TransactionsPerSec                    uint32
	TransactionsTotal                     uint32
	ViewstateMACValidationFailure         uint32
}

func (c *aspnet_aspnetapplicationsCollector) collect(ch chan<- prometheus.Metric) (*prometheus.Desc, error) {
	var dst []Win32_PerfFormattedData_ASPNET_ASPNETApplications
	q := queryAll(&dst)
	if err := wmi.Query(q, &dst); err != nil {
		return nil, err
	}

	ch <- prometheus.MustNewConstMetric(
		c.AnonymousRequests,
		prometheus.GaugeValue,
		float64(dst[0].AnonymousRequests),
	)

	ch <- prometheus.MustNewConstMetric(
		c.AnonymousRequestsPerSec,
		prometheus.GaugeValue,
		float64(dst[0].AnonymousRequestsPerSec),
	)

	ch <- prometheus.MustNewConstMetric(
		c.ApplicationLifetimeEvents,
		prometheus.GaugeValue,
		float64(dst[0].ApplicationLifetimeEvents),
	)

	ch <- prometheus.MustNewConstMetric(
		c.ApplicationLifetimeEventsPerSec,
		prometheus.GaugeValue,
		float64(dst[0].ApplicationLifetimeEventsPerSec),
	)

	ch <- prometheus.MustNewConstMetric(
		c.AuditFailureEventsRaised,
		prometheus.GaugeValue,
		float64(dst[0].AuditFailureEventsRaised),
	)

	ch <- prometheus.MustNewConstMetric(
		c.AuditSuccessEventsRaised,
		prometheus.GaugeValue,
		float64(dst[0].AuditSuccessEventsRaised),
	)

	ch <- prometheus.MustNewConstMetric(
		c.CacheAPIEntries,
		prometheus.GaugeValue,
		float64(dst[0].CacheAPIEntries),
	)

	ch <- prometheus.MustNewConstMetric(
		c.CacheAPIHitRatio,
		prometheus.GaugeValue,
		float64(dst[0].CacheAPIHitRatio),
	)

	ch <- prometheus.MustNewConstMetric(
		c.CacheAPIHits,
		prometheus.GaugeValue,
		float64(dst[0].CacheAPIHits),
	)

	ch <- prometheus.MustNewConstMetric(
		c.CacheAPIMisses,
		prometheus.GaugeValue,
		float64(dst[0].CacheAPIMisses),
	)

	ch <- prometheus.MustNewConstMetric(
		c.CacheAPITrims,
		prometheus.GaugeValue,
		float64(dst[0].CacheAPITrims),
	)

	ch <- prometheus.MustNewConstMetric(
		c.CacheAPITurnoverRate,
		prometheus.GaugeValue,
		float64(dst[0].CacheAPITurnoverRate),
	)

	ch <- prometheus.MustNewConstMetric(
		c.CachePercentMachineMemoryLimitUsed,
		prometheus.GaugeValue,
		float64(dst[0].CachePercentMachineMemoryLimitUsed),
	)

	ch <- prometheus.MustNewConstMetric(
		c.CachePercentProcessMemoryLimitUsed,
		prometheus.GaugeValue,
		float64(dst[0].CachePercentProcessMemoryLimitUsed),
	)

	ch <- prometheus.MustNewConstMetric(
		c.CacheTotalEntries,
		prometheus.GaugeValue,
		float64(dst[0].CacheTotalEntries),
	)

	ch <- prometheus.MustNewConstMetric(
		c.CacheTotalHitRatio,
		prometheus.GaugeValue,
		float64(dst[0].CacheTotalHitRatio),
	)

	ch <- prometheus.MustNewConstMetric(
		c.CacheTotalHits,
		prometheus.GaugeValue,
		float64(dst[0].CacheTotalHits),
	)

	ch <- prometheus.MustNewConstMetric(
		c.CacheTotalMisses,
		prometheus.GaugeValue,
		float64(dst[0].CacheTotalMisses),
	)

	ch <- prometheus.MustNewConstMetric(
		c.CacheTotalTrims,
		prometheus.GaugeValue,
		float64(dst[0].CacheTotalTrims),
	)

	ch <- prometheus.MustNewConstMetric(
		c.CacheTotalTurnoverRate,
		prometheus.GaugeValue,
		float64(dst[0].CacheTotalTurnoverRate),
	)

	ch <- prometheus.MustNewConstMetric(
		c.CompilationsTotal,
		prometheus.GaugeValue,
		float64(dst[0].CompilationsTotal),
	)

	ch <- prometheus.MustNewConstMetric(
		c.DebuggingRequests,
		prometheus.GaugeValue,
		float64(dst[0].DebuggingRequests),
	)

	ch <- prometheus.MustNewConstMetric(
		c.ErrorEventsRaised,
		prometheus.GaugeValue,
		float64(dst[0].ErrorEventsRaised),
	)

	ch <- prometheus.MustNewConstMetric(
		c.ErrorEventsRaisedPerSec,
		prometheus.GaugeValue,
		float64(dst[0].ErrorEventsRaisedPerSec),
	)

	ch <- prometheus.MustNewConstMetric(
		c.ErrorsDuringCompilation,
		prometheus.GaugeValue,
		float64(dst[0].ErrorsDuringCompilation),
	)

	ch <- prometheus.MustNewConstMetric(
		c.ErrorsDuringExecution,
		prometheus.GaugeValue,
		float64(dst[0].ErrorsDuringExecution),
	)

	ch <- prometheus.MustNewConstMetric(
		c.ErrorsDuringPreprocessing,
		prometheus.GaugeValue,
		float64(dst[0].ErrorsDuringPreprocessing),
	)

	ch <- prometheus.MustNewConstMetric(
		c.ErrorsTotal,
		prometheus.GaugeValue,
		float64(dst[0].ErrorsTotal),
	)

	ch <- prometheus.MustNewConstMetric(
		c.ErrorsTotalPerSec,
		prometheus.GaugeValue,
		float64(dst[0].ErrorsTotalPerSec),
	)

	ch <- prometheus.MustNewConstMetric(
		c.ErrorsUnhandledDuringExecution,
		prometheus.GaugeValue,
		float64(dst[0].ErrorsUnhandledDuringExecution),
	)

	ch <- prometheus.MustNewConstMetric(
		c.ErrorsUnhandledDuringExecutionPerSec,
		prometheus.GaugeValue,
		float64(dst[0].ErrorsUnhandledDuringExecutionPerSec),
	)

	ch <- prometheus.MustNewConstMetric(
		c.EventsRaised,
		prometheus.GaugeValue,
		float64(dst[0].EventsRaised),
	)

	ch <- prometheus.MustNewConstMetric(
		c.EventsRaisedPerSec,
		prometheus.GaugeValue,
		float64(dst[0].EventsRaisedPerSec),
	)

	ch <- prometheus.MustNewConstMetric(
		c.FormsAuthenticationFailure,
		prometheus.GaugeValue,
		float64(dst[0].FormsAuthenticationFailure),
	)

	ch <- prometheus.MustNewConstMetric(
		c.FormsAuthenticationSuccess,
		prometheus.GaugeValue,
		float64(dst[0].FormsAuthenticationSuccess),
	)

	ch <- prometheus.MustNewConstMetric(
		c.InfrastructureErrorEventsRaised,
		prometheus.GaugeValue,
		float64(dst[0].InfrastructureErrorEventsRaised),
	)

	ch <- prometheus.MustNewConstMetric(
		c.InfrastructureErrorEventsRaisedPerSec,
		prometheus.GaugeValue,
		float64(dst[0].InfrastructureErrorEventsRaisedPerSec),
	)

	ch <- prometheus.MustNewConstMetric(
		c.ManagedMemoryUsedestimated,
		prometheus.GaugeValue,
		float64(dst[0].ManagedMemoryUsedestimated),
	)

	ch <- prometheus.MustNewConstMetric(
		c.MembershipAuthenticationFailure,
		prometheus.GaugeValue,
		float64(dst[0].MembershipAuthenticationFailure),
	)

	ch <- prometheus.MustNewConstMetric(
		c.MembershipAuthenticationSuccess,
		prometheus.GaugeValue,
		float64(dst[0].MembershipAuthenticationSuccess),
	)

	ch <- prometheus.MustNewConstMetric(
		c.OutputCacheEntries,
		prometheus.GaugeValue,
		float64(dst[0].OutputCacheEntries),
	)

	ch <- prometheus.MustNewConstMetric(
		c.OutputCacheHitRatio,
		prometheus.GaugeValue,
		float64(dst[0].OutputCacheHitRatio),
	)

	ch <- prometheus.MustNewConstMetric(
		c.OutputCacheHits,
		prometheus.GaugeValue,
		float64(dst[0].OutputCacheHits),
	)

	ch <- prometheus.MustNewConstMetric(
		c.OutputCacheMisses,
		prometheus.GaugeValue,
		float64(dst[0].OutputCacheMisses),
	)

	ch <- prometheus.MustNewConstMetric(
		c.OutputCacheTrims,
		prometheus.GaugeValue,
		float64(dst[0].OutputCacheTrims),
	)

	ch <- prometheus.MustNewConstMetric(
		c.OutputCacheTurnoverRate,
		prometheus.GaugeValue,
		float64(dst[0].OutputCacheTurnoverRate),
	)

	ch <- prometheus.MustNewConstMetric(
		c.PercentManagedProcessorTimeestimated,
		prometheus.GaugeValue,
		float64(dst[0].PercentManagedProcessorTimeestimated),
	)

	ch <- prometheus.MustNewConstMetric(
		c.PipelineInstanceCount,
		prometheus.GaugeValue,
		float64(dst[0].PipelineInstanceCount),
	)

	ch <- prometheus.MustNewConstMetric(
		c.RequestBytesInTotal,
		prometheus.GaugeValue,
		float64(dst[0].RequestBytesInTotal),
	)

	ch <- prometheus.MustNewConstMetric(
		c.RequestBytesInTotalWebSockets,
		prometheus.GaugeValue,
		float64(dst[0].RequestBytesInTotalWebSockets),
	)

	ch <- prometheus.MustNewConstMetric(
		c.RequestBytesOutTotal,
		prometheus.GaugeValue,
		float64(dst[0].RequestBytesOutTotal),
	)

	ch <- prometheus.MustNewConstMetric(
		c.RequestBytesOutTotalWebSockets,
		prometheus.GaugeValue,
		float64(dst[0].RequestBytesOutTotalWebSockets),
	)

	ch <- prometheus.MustNewConstMetric(
		c.RequestErrorEventsRaised,
		prometheus.GaugeValue,
		float64(dst[0].RequestErrorEventsRaised),
	)

	ch <- prometheus.MustNewConstMetric(
		c.RequestErrorEventsRaisedPerSec,
		prometheus.GaugeValue,
		float64(dst[0].RequestErrorEventsRaisedPerSec),
	)

	ch <- prometheus.MustNewConstMetric(
		c.RequestEventsRaised,
		prometheus.GaugeValue,
		float64(dst[0].RequestEventsRaised),
	)

	ch <- prometheus.MustNewConstMetric(
		c.RequestEventsRaisedPerSec,
		prometheus.GaugeValue,
		float64(dst[0].RequestEventsRaisedPerSec),
	)

	ch <- prometheus.MustNewConstMetric(
		c.RequestExecutionTime,
		prometheus.GaugeValue,
		float64(dst[0].RequestExecutionTime),
	)

	ch <- prometheus.MustNewConstMetric(
		c.RequestsDisconnected,
		prometheus.GaugeValue,
		float64(dst[0].RequestsDisconnected),
	)

	ch <- prometheus.MustNewConstMetric(
		c.RequestsExecuting,
		prometheus.GaugeValue,
		float64(dst[0].RequestsExecuting),
	)

	ch <- prometheus.MustNewConstMetric(
		c.RequestsExecutingWebSockets,
		prometheus.GaugeValue,
		float64(dst[0].RequestsExecutingWebSockets),
	)

	ch <- prometheus.MustNewConstMetric(
		c.RequestsFailed,
		prometheus.GaugeValue,
		float64(dst[0].RequestsFailed),
	)

	ch <- prometheus.MustNewConstMetric(
		c.RequestsFailedWebSockets,
		prometheus.GaugeValue,
		float64(dst[0].RequestsFailedWebSockets),
	)

	ch <- prometheus.MustNewConstMetric(
		c.RequestsInApplicationQueue,
		prometheus.GaugeValue,
		float64(dst[0].RequestsInApplicationQueue),
	)

	ch <- prometheus.MustNewConstMetric(
		c.RequestsNotAuthorized,
		prometheus.GaugeValue,
		float64(dst[0].RequestsNotAuthorized),
	)

	ch <- prometheus.MustNewConstMetric(
		c.RequestsNotFound,
		prometheus.GaugeValue,
		float64(dst[0].RequestsNotFound),
	)

	ch <- prometheus.MustNewConstMetric(
		c.RequestsPerSec,
		prometheus.GaugeValue,
		float64(dst[0].RequestsPerSec),
	)

	ch <- prometheus.MustNewConstMetric(
		c.RequestsRejected,
		prometheus.GaugeValue,
		float64(dst[0].RequestsRejected),
	)

	ch <- prometheus.MustNewConstMetric(
		c.RequestsSucceeded,
		prometheus.GaugeValue,
		float64(dst[0].RequestsSucceeded),
	)

	ch <- prometheus.MustNewConstMetric(
		c.RequestsSucceededWebSockets,
		prometheus.GaugeValue,
		float64(dst[0].RequestsSucceededWebSockets),
	)

	ch <- prometheus.MustNewConstMetric(
		c.RequestsTimedOut,
		prometheus.GaugeValue,
		float64(dst[0].RequestsTimedOut),
	)

	ch <- prometheus.MustNewConstMetric(
		c.RequestsTotal,
		prometheus.GaugeValue,
		float64(dst[0].RequestsTotal),
	)

	ch <- prometheus.MustNewConstMetric(
		c.RequestsTotalWebSockets,
		prometheus.GaugeValue,
		float64(dst[0].RequestsTotalWebSockets),
	)

	ch <- prometheus.MustNewConstMetric(
		c.RequestWaitTime,
		prometheus.GaugeValue,
		float64(dst[0].RequestWaitTime),
	)

	ch <- prometheus.MustNewConstMetric(
		c.SessionsAbandoned,
		prometheus.GaugeValue,
		float64(dst[0].SessionsAbandoned),
	)

	ch <- prometheus.MustNewConstMetric(
		c.SessionsActive,
		prometheus.GaugeValue,
		float64(dst[0].SessionsActive),
	)

	ch <- prometheus.MustNewConstMetric(
		c.SessionSQLServerconnectionstotal,
		prometheus.GaugeValue,
		float64(dst[0].SessionSQLServerconnectionstotal),
	)

	ch <- prometheus.MustNewConstMetric(
		c.SessionStateServerconnectionstotal,
		prometheus.GaugeValue,
		float64(dst[0].SessionStateServerconnectionstotal),
	)

	ch <- prometheus.MustNewConstMetric(
		c.SessionsTimedOut,
		prometheus.GaugeValue,
		float64(dst[0].SessionsTimedOut),
	)

	ch <- prometheus.MustNewConstMetric(
		c.SessionsTotal,
		prometheus.GaugeValue,
		float64(dst[0].SessionsTotal),
	)

	ch <- prometheus.MustNewConstMetric(
		c.TransactionsAborted,
		prometheus.GaugeValue,
		float64(dst[0].TransactionsAborted),
	)

	ch <- prometheus.MustNewConstMetric(
		c.TransactionsCommitted,
		prometheus.GaugeValue,
		float64(dst[0].TransactionsCommitted),
	)

	ch <- prometheus.MustNewConstMetric(
		c.TransactionsPending,
		prometheus.GaugeValue,
		float64(dst[0].TransactionsPending),
	)

	ch <- prometheus.MustNewConstMetric(
		c.TransactionsPerSec,
		prometheus.GaugeValue,
		float64(dst[0].TransactionsPerSec),
	)

	ch <- prometheus.MustNewConstMetric(
		c.TransactionsTotal,
		prometheus.GaugeValue,
		float64(dst[0].TransactionsTotal),
	)

	ch <- prometheus.MustNewConstMetric(
		c.ViewstateMACValidationFailure,
		prometheus.GaugeValue,
		float64(dst[0].ViewstateMACValidationFailure),
	)

	return nil, nil
}
