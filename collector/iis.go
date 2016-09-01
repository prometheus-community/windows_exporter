// returns data points from Win32_PerfRawData_W3SVC_WebService
// https://msdn.microsoft.com/en-us/library/aa394345.aspx - Win32_OperatingSystem class

package collector

import (
	"flag"
	"fmt"
	"log"
	"regexp"

	"github.com/StackExchange/wmi"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	siteWhitelist = flag.String("collector.iis.site-whitelist", ".+", "Regexp of sites to whitelist. Site name must both match whitelist and not match blacklist to be included.")
	siteBlacklist = flag.String("collector.iis.site-blacklist", "", "Regexp of sites to blacklist. Site name must both match whitelist and not match blacklist to be included.")
)

// A IISCollector is a Prometheus collector for WMI Win32_PerfRawData_W3SVC_WebService metrics
type IISCollector struct {
	CurrentAnonymousUsers         *prometheus.Desc
	CurrentBlockedAsyncIORequests *prometheus.Desc
	CurrentCGIRequests            *prometheus.Desc
	CurrentConnections            *prometheus.Desc
	CurrentISAPIExtensionRequests *prometheus.Desc
	CurrentNonAnonymousUsers      *prometheus.Desc

	TotalBytesReceived                  *prometheus.Desc
	TotalBytesSent                      *prometheus.Desc
	TotalAnonymousUsers                 *prometheus.Desc
	TotalBlockedAsyncIORequests         *prometheus.Desc
	TotalCGIRequests                    *prometheus.Desc
	TotalConnectionAttemptsAllInstances *prometheus.Desc
	TotalRequests                       *prometheus.Desc
	TotalFilesReceived                  *prometheus.Desc
	TotalFilesSent                      *prometheus.Desc
	TotalISAPIExtensionRequests         *prometheus.Desc
	TotalLockedErrors                   *prometheus.Desc
	TotalLogonAttempts                  *prometheus.Desc
	TotalNonAnonymousUsers              *prometheus.Desc
	TotalNotFoundErrors                 *prometheus.Desc
	TotalRejectedAsyncIORequests        *prometheus.Desc

	siteWhitelistPattern *regexp.Regexp
	siteBlacklistPattern *regexp.Regexp
}

// NewIISCollector ...
func NewIISCollector() *IISCollector {
	const subsystem = "iis"

	return &IISCollector{
		// Gauges
		CurrentAnonymousUsers: prometheus.NewDesc(
			prometheus.BuildFQName(wmiNamespace, subsystem, "current_anonymous_users"),
			"Number of users who currently have an anonymous connection using the Web service (WebService.CurrentAnonymousUsers)",
			[]string{"site"},
			nil,
		),
		CurrentBlockedAsyncIORequests: prometheus.NewDesc(
			prometheus.BuildFQName(wmiNamespace, subsystem, "current_blocked_async_io_requests"),
			"Current requests temporarily blocked due to bandwidth throttling settings (WebService.CurrentBlockedAsyncIORequests)",
			[]string{"site"},
			nil,
		),
		CurrentCGIRequests: prometheus.NewDesc(
			prometheus.BuildFQName(wmiNamespace, subsystem, "current_cgi_requests"),
			"Current number of CGI requests being simultaneously processed by the Web service (WebService.CurrentCGIRequests)",
			[]string{"site"},
			nil,
		),
		CurrentConnections: prometheus.NewDesc(
			prometheus.BuildFQName(wmiNamespace, subsystem, "current_connections"),
			"Current number of connections established with the Web service (WebService.CurrentConnections)",
			[]string{"site"},
			nil,
		),
		CurrentISAPIExtensionRequests: prometheus.NewDesc(
			prometheus.BuildFQName(wmiNamespace, subsystem, "current_isapi_extension_requests"),
			"Current number of ISAPI requests being simultaneously processed by the Web service (WebService.CurrentISAPIExtensionRequests)",
			[]string{"site"},
			nil,
		),
		CurrentNonAnonymousUsers: prometheus.NewDesc(
			prometheus.BuildFQName(wmiNamespace, subsystem, "current_non_anonymous_users"),
			"Number of users who currently have a non-anonymous connection using the Web service (WebService.CurrentNonAnonymousUsers)",
			[]string{"site"},
			nil,
		),

		// Counters
		TotalBytesReceived: prometheus.NewDesc(
			prometheus.BuildFQName(wmiNamespace, subsystem, "received_bytes_total"),
			"Number of data bytes that have been received by the Web service (WebService.TotalBytesReceived)",
			[]string{"site"},
			nil,
		),
		TotalBytesSent: prometheus.NewDesc(
			prometheus.BuildFQName(wmiNamespace, subsystem, "sent_bytes_total"),
			"Number of data bytes that have been sent by the Web service (WebService.TotalBytesSent)",
			[]string{"site"},
			nil,
		),
		TotalAnonymousUsers: prometheus.NewDesc(
			prometheus.BuildFQName(wmiNamespace, subsystem, "anonymous_users_total"),
			"Total number of users who established an anonymous connection with the Web service (WebService.TotalAnonymousUsers)",
			[]string{"site"},
			nil,
		),
		TotalBlockedAsyncIORequests: prometheus.NewDesc(
			prometheus.BuildFQName(wmiNamespace, subsystem, "blocked_async_io_requests_total"),
			"Total requests temporarily blocked due to bandwidth throttling settings (WebService.TotalBlockedAsyncIORequests)",
			[]string{"site"},
			nil,
		),
		TotalCGIRequests: prometheus.NewDesc(
			prometheus.BuildFQName(wmiNamespace, subsystem, "cgi_requests_total"),
			"Total CGI requests is the total number of CGI requests (WebService.TotalCGIRequests)",
			[]string{"site"},
			nil,
		),
		TotalConnectionAttemptsAllInstances: prometheus.NewDesc(
			prometheus.BuildFQName(wmiNamespace, subsystem, "connection_attempts_all_instances_total"),
			"Number of connections that have been attempted using the Web service (WebService.TotalConnectionAttemptsAllInstances)",
			[]string{"site"},
			nil,
		),
		TotalRequests: prometheus.NewDesc(
			prometheus.BuildFQName(wmiNamespace, subsystem, "requests_total"),
			"Number of HTTP requests (WebService.TotalRequests)",
			[]string{"site", "method"},
			nil,
		),
		TotalFilesReceived: prometheus.NewDesc(
			prometheus.BuildFQName(wmiNamespace, subsystem, "files_received_total"),
			"Number of files received by the Web service (WebService.TotalFilesReceived)",
			[]string{"site"},
			nil,
		),
		TotalFilesSent: prometheus.NewDesc(
			prometheus.BuildFQName(wmiNamespace, subsystem, "files_sent_total"),
			"Number of files sent by the Web service (WebService.TotalFilesSent)",
			[]string{"site"},
			nil,
		),
		TotalISAPIExtensionRequests: prometheus.NewDesc(
			prometheus.BuildFQName(wmiNamespace, subsystem, "ipapi_extension_requests_total"),
			"ISAPI Extension Requests received (WebService.TotalISAPIExtensionRequests)",
			[]string{"site"},
			nil,
		),
		TotalLockedErrors: prometheus.NewDesc(
			prometheus.BuildFQName(wmiNamespace, subsystem, "locked_errors_total"),
			"Number of requests that couldn't be satisfied by the server because the requested resource was locked (WebService.TotalLockedErrors)",
			[]string{"site"},
			nil,
		),
		TotalLogonAttempts: prometheus.NewDesc(
			prometheus.BuildFQName(wmiNamespace, subsystem, "logon_attempts_total"),
			"Number of logons attempts to the Web Service (WebService.TotalLogonAttempts)",
			[]string{"site"},
			nil,
		),
		TotalNonAnonymousUsers: prometheus.NewDesc(
			prometheus.BuildFQName(wmiNamespace, subsystem, "non_anonymous_users_total"),
			"Number of users who established a non-anonymous connection with the Web service (WebService.TotalNonAnonymousUsers)",
			[]string{"site"},
			nil,
		),
		TotalNotFoundErrors: prometheus.NewDesc(
			prometheus.BuildFQName(wmiNamespace, subsystem, "not_found_errors_total"),
			"Number of requests that couldn't be satisfied by the server because the requested document could not be found (WebService.TotalNotFoundErrors)",
			[]string{"site"},
			nil,
		),
		TotalRejectedAsyncIORequests: prometheus.NewDesc(
			prometheus.BuildFQName(wmiNamespace, subsystem, "rejected_async_io_requests_total"),
			"Requests rejected due to bandwidth throttling settings (WebService.TotalRejectedAsyncIORequests)",
			[]string{"site"},
			nil,
		),

		siteWhitelistPattern: regexp.MustCompile(fmt.Sprintf("^(?:%s)$", *siteWhitelist)),
		siteBlacklistPattern: regexp.MustCompile(fmt.Sprintf("^(?:%s)$", *siteBlacklist)),
	}
}

// Collect sends the metric values for each metric
// to the provided prometheus Metric channel.
func (c *IISCollector) Collect(ch chan<- prometheus.Metric) {
	if desc, err := c.collect(ch); err != nil {
		log.Println("[ERROR] failed collecting iis metrics:", desc, err)
		return
	}
}

// Describe sends the descriptors of each metric over to the provided channel.
// The corresponding metric values are sent separately.
func (c *IISCollector) Describe(ch chan<- *prometheus.Desc) {

}

type Win32_PerfRawData_W3SVC_WebService struct {
	Name string

	CurrentAnonymousUsers         uint32
	CurrentBlockedAsyncIORequests uint32
	CurrentCGIRequests            uint32
	CurrentConnections            uint32
	CurrentISAPIExtensionRequests uint32
	CurrentNonAnonymousUsers      uint32

	TotalBytesSent                      uint64
	TotalBytesReceived                  uint64
	TotalAnonymousUsers                 uint32
	TotalBlockedAsyncIORequests         uint32
	TotalCGIRequests                    uint32
	TotalConnectionAttemptsAllInstances uint32
	TotalCopyRequests                   uint32
	TotalDeleteRequests                 uint32
	TotalFilesReceived                  uint32
	TotalFilesSent                      uint32
	TotalGetRequests                    uint32
	TotalHeadRequests                   uint32
	TotalISAPIExtensionRequests         uint32
	TotalLockedErrors                   uint32
	TotalLockRequests                   uint32
	TotalLogonAttempts                  uint32
	TotalMethodRequests                 uint32
	TotalMethodRequestsPerSec           uint32
	TotalMkcolRequests                  uint32
	TotalMoveRequests                   uint32
	TotalNonAnonymousUsers              uint32
	TotalNotFoundErrors                 uint32
	TotalOptionsRequests                uint32
	TotalOtherRequestMethods            uint32
	TotalPostRequests                   uint32
	TotalPropfindRequests               uint32
	TotalProppatchRequests              uint32
	TotalPutRequests                    uint32
	TotalRejectedAsyncIORequests        uint32
	TotalSearchRequests                 uint32
	TotalTraceRequests                  uint32
	TotalUnlockRequests                 uint32
}

func (c *IISCollector) collect(ch chan<- prometheus.Metric) (*prometheus.Desc, error) {
	var dst []Win32_PerfRawData_W3SVC_WebService
	q := wmi.CreateQuery(&dst, "")
	if err := wmi.Query(q, &dst); err != nil {
		return nil, err
	}

	for _, site := range dst {
		if site.Name == "_Total" ||
			c.siteBlacklistPattern.MatchString(site.Name) ||
			!c.siteWhitelistPattern.MatchString(site.Name) {
			continue
		}

		// Gauges
		ch <- prometheus.MustNewConstMetric(
			c.CurrentAnonymousUsers,
			prometheus.GaugeValue,
			float64(site.CurrentAnonymousUsers),
			site.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.CurrentBlockedAsyncIORequests,
			prometheus.GaugeValue,
			float64(site.CurrentBlockedAsyncIORequests),
			site.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.CurrentCGIRequests,
			prometheus.GaugeValue,
			float64(site.CurrentCGIRequests),
			site.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.CurrentConnections,
			prometheus.GaugeValue,
			float64(site.CurrentConnections),
			site.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.CurrentISAPIExtensionRequests,
			prometheus.GaugeValue,
			float64(site.CurrentISAPIExtensionRequests),
			site.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.CurrentNonAnonymousUsers,
			prometheus.GaugeValue,
			float64(site.CurrentNonAnonymousUsers),
			site.Name,
		)

		// Counters
		ch <- prometheus.MustNewConstMetric(
			c.TotalBytesReceived,
			prometheus.CounterValue,
			float64(site.TotalBytesReceived),
			site.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.TotalBytesSent,
			prometheus.CounterValue,
			float64(site.TotalBytesSent),
			site.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.TotalAnonymousUsers,
			prometheus.CounterValue,
			float64(site.TotalAnonymousUsers),
			site.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.TotalBlockedAsyncIORequests,
			prometheus.CounterValue,
			float64(site.TotalBlockedAsyncIORequests),
			site.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.TotalRejectedAsyncIORequests,
			prometheus.CounterValue,
			float64(site.TotalRejectedAsyncIORequests),
			site.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.TotalCGIRequests,
			prometheus.CounterValue,
			float64(site.TotalCGIRequests),
			site.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.TotalConnectionAttemptsAllInstances,
			prometheus.CounterValue,
			float64(site.TotalConnectionAttemptsAllInstances),
			site.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.TotalFilesReceived,
			prometheus.CounterValue,
			float64(site.TotalFilesReceived),
			site.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.TotalFilesSent,
			prometheus.CounterValue,
			float64(site.TotalFilesSent),
			site.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.TotalLockedErrors,
			prometheus.CounterValue,
			float64(site.TotalLockedErrors),
			site.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.TotalLogonAttempts,
			prometheus.CounterValue,
			float64(site.TotalLogonAttempts),
			site.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.TotalNonAnonymousUsers,
			prometheus.CounterValue,
			float64(site.TotalNonAnonymousUsers),
			site.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.TotalNotFoundErrors,
			prometheus.CounterValue,
			float64(site.TotalNotFoundErrors),
			site.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.TotalISAPIExtensionRequests,
			prometheus.CounterValue,
			float64(site.TotalISAPIExtensionRequests),
			site.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.TotalRequests,
			prometheus.CounterValue,
			float64(site.TotalOtherRequestMethods),
			site.Name,
			"other",
		)
		ch <- prometheus.MustNewConstMetric(
			c.TotalRequests,
			prometheus.CounterValue,
			float64(site.TotalCopyRequests),
			site.Name,
			"COPY",
		)
		ch <- prometheus.MustNewConstMetric(
			c.TotalRequests,
			prometheus.CounterValue,
			float64(site.TotalDeleteRequests),
			site.Name,
			"DELETE",
		)
		ch <- prometheus.MustNewConstMetric(
			c.TotalRequests,
			prometheus.CounterValue,
			float64(site.TotalGetRequests),
			site.Name,
			"GET",
		)
		ch <- prometheus.MustNewConstMetric(
			c.TotalRequests,
			prometheus.CounterValue,
			float64(site.TotalHeadRequests),
			site.Name,
			"HEAD",
		)
		ch <- prometheus.MustNewConstMetric(
			c.TotalRequests,
			prometheus.CounterValue,
			float64(site.TotalLockRequests),
			site.Name,
			"LOCK",
		)
		ch <- prometheus.MustNewConstMetric(
			c.TotalRequests,
			prometheus.CounterValue,
			float64(site.TotalMkcolRequests),
			site.Name,
			"MKCOL",
		)
		ch <- prometheus.MustNewConstMetric(
			c.TotalRequests,
			prometheus.CounterValue,
			float64(site.TotalMoveRequests),
			site.Name,
			"MOVE",
		)
		ch <- prometheus.MustNewConstMetric(
			c.TotalRequests,
			prometheus.CounterValue,
			float64(site.TotalOptionsRequests),
			site.Name,
			"OPTIONS",
		)
		ch <- prometheus.MustNewConstMetric(
			c.TotalRequests,
			prometheus.CounterValue,
			float64(site.TotalPostRequests),
			site.Name,
			"POST",
		)
		ch <- prometheus.MustNewConstMetric(
			c.TotalRequests,
			prometheus.CounterValue,
			float64(site.TotalPropfindRequests),
			site.Name,
			"PROPFIND",
		)
		ch <- prometheus.MustNewConstMetric(
			c.TotalRequests,
			prometheus.CounterValue,
			float64(site.TotalProppatchRequests),
			site.Name,
			"PROPPATCH",
		)
		ch <- prometheus.MustNewConstMetric(
			c.TotalRequests,
			prometheus.CounterValue,
			float64(site.TotalPutRequests),
			site.Name,
			"PUT",
		)
		ch <- prometheus.MustNewConstMetric(
			c.TotalRequests,
			prometheus.CounterValue,
			float64(site.TotalSearchRequests),
			site.Name,
			"SEARCH",
		)
		ch <- prometheus.MustNewConstMetric(
			c.TotalRequests,
			prometheus.CounterValue,
			float64(site.TotalTraceRequests),
			site.Name,
			"TRACE",
		)
		ch <- prometheus.MustNewConstMetric(
			c.TotalRequests,
			prometheus.CounterValue,
			float64(site.TotalUnlockRequests),
			site.Name,
			"UNLOCK",
		)

	}

	return nil, nil
}
