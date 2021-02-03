package collector

import (
	"github.com/StackExchange/wmi"
	"github.com/prometheus/client_golang/prometheus"
)

func init() {
	registerCollector("terminal_service_gateway", newTerminalServiceGatewayCollector) // TODO: Add any perflib dependencies here
}

// A TerminalServiceGatewayCollector is a Prometheus collector for WMI Win32_PerfRawData_TSGateway_TerminalServiceGateway metrics
// see https://wutils.com/wmi/root/cimv2/win32_perfrawdata_tsgateway_terminalservicegateway/
type TerminalServiceGatewayCollector struct {
	ConnectionRequestAuthorizationTime *prometheus.Desc
	CurrentConnections                 *prometheus.Desc
	FailedConnectionAuthorization      *prometheus.Desc
	FailedConnections                  *prometheus.Desc
	FailedResourceAuthorization        *prometheus.Desc
	SuccessfulConnections              *prometheus.Desc
}

func newTerminalServiceGatewayCollector() (Collector, error) {
	const subsystem = "terminal_service_gateway"
	return &TerminalServiceGatewayCollector{
		ConnectionRequestAuthorizationTime: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "connection_request_authorization_time_seconds"),
			"Shows the average connection request authentication and authorization times in seconds",
			nil,
			nil,
		),
		CurrentConnections: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "current_connections_count"),
			"Shows the total number of active/inactive connections to the RDG server at any given moment",
			nil,
			nil,
		),
		FailedConnectionAuthorization: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "failed_connection_authorization_total"),
			"Shows the total number of requests that failed due to insufficient connection authorization privilege",
			nil,
			nil,
		),
		FailedConnections: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "failed_connections_total"),
			"Shows the number of connection requests that are all failed due to errors and authorization failure",
			nil,
			nil,
		),
		FailedResourceAuthorization: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "failed_resource_authorization_total"),
			"Shows the total number of requests that failed due to insufficient resource authorization privilege",
			nil,
			nil,
		),
		SuccessfulConnections: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "successful_connections_total"),
			"Shows the number of requests that were successfully processed and connected",
			nil,
			nil,
		),
	}, nil
}

// Win32_PerfRawData_TSGateway_TerminalServiceGateway docs:
// - <add link to documentation here>
type Win32_PerfRawData_TSGateway_TerminalServiceGateway struct {
	Name string

	ConnectionRequestAuthorizationTime uint32
	CurrentConnections                 uint32
	FailedConnectionAuthorization      uint32
	FailedConnections                  uint32
	FailedResourceAuthorization        uint32
	SuccessfulConnections              uint32
}

// Collect sends the metric values for each metric
// to the provided prometheus Metric channel.
func (c *TerminalServiceGatewayCollector) Collect(ctx *ScrapeContext, ch chan<- prometheus.Metric) error {
	var dst []Win32_PerfRawData_TSGateway_TerminalServiceGateway
	q := queryAll(&dst)
	if err := wmi.Query(q, &dst); err != nil {
		return err
	}

	ch <- prometheus.MustNewConstMetric(
		c.ConnectionRequestAuthorizationTime,
		prometheus.GaugeValue,
		float64(dst[0].ConnectionRequestAuthorizationTime),
	)

	ch <- prometheus.MustNewConstMetric(
		c.CurrentConnections,
		prometheus.GaugeValue,
		float64(dst[0].CurrentConnections),
	)

	ch <- prometheus.MustNewConstMetric(
		c.FailedConnectionAuthorization,
		prometheus.GaugeValue,
		float64(dst[0].FailedConnectionAuthorization),
	)

	ch <- prometheus.MustNewConstMetric(
		c.FailedConnections,
		prometheus.GaugeValue,
		float64(dst[0].FailedConnections),
	)

	ch <- prometheus.MustNewConstMetric(
		c.FailedResourceAuthorization,
		prometheus.GaugeValue,
		float64(dst[0].FailedResourceAuthorization),
	)

	ch <- prometheus.MustNewConstMetric(
		c.SuccessfulConnections,
		prometheus.GaugeValue,
		float64(dst[0].SuccessfulConnections),
	)

	return nil
}
