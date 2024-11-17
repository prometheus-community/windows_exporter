//go:build windows

package iis

import (
	"fmt"

	"github.com/prometheus-community/windows_exporter/internal/perfdata"
	"github.com/prometheus-community/windows_exporter/internal/types"
	"github.com/prometheus/client_golang/prometheus"
)

type collectorWebService struct {
	perfDataCollectorWebService *perfdata.Collector

	currentAnonymousUsers               *prometheus.Desc
	currentBlockedAsyncIORequests       *prometheus.Desc
	currentCGIRequests                  *prometheus.Desc
	currentConnections                  *prometheus.Desc
	currentISAPIExtensionRequests       *prometheus.Desc
	currentNonAnonymousUsers            *prometheus.Desc
	serviceUptime                       *prometheus.Desc
	totalBytesReceived                  *prometheus.Desc
	totalBytesSent                      *prometheus.Desc
	totalAnonymousUsers                 *prometheus.Desc
	totalBlockedAsyncIORequests         *prometheus.Desc
	totalCGIRequests                    *prometheus.Desc
	totalConnectionAttemptsAllInstances *prometheus.Desc
	totalRequests                       *prometheus.Desc
	totalFilesReceived                  *prometheus.Desc
	totalFilesSent                      *prometheus.Desc
	totalISAPIExtensionRequests         *prometheus.Desc
	totalLockedErrors                   *prometheus.Desc
	totalLogonAttempts                  *prometheus.Desc
	totalNonAnonymousUsers              *prometheus.Desc
	totalNotFoundErrors                 *prometheus.Desc
	totalRejectedAsyncIORequests        *prometheus.Desc
}

const (
	CurrentAnonymousUsers               = "Current Anonymous Users"
	CurrentBlockedAsyncIORequests       = "Current Blocked Async I/O Requests"
	CurrentCGIRequests                  = "Current CGI Requests"
	CurrentConnections                  = "Current Connections"
	CurrentISAPIExtensionRequests       = "Current ISAPI Extension Requests"
	CurrentNonAnonymousUsers            = "Current NonAnonymous Users"
	ServiceUptime                       = "Service Uptime"
	TotalBytesReceived                  = "Total Bytes Received"
	TotalBytesSent                      = "Total Bytes Sent"
	TotalAnonymousUsers                 = "Total Anonymous Users"
	TotalBlockedAsyncIORequests         = "Total Blocked Async I/O Requests"
	TotalCGIRequests                    = "Total CGI Requests"
	TotalConnectionAttemptsAllInstances = "Total Connection Attempts (all instances)"
	TotalFilesReceived                  = "Total Files Received"
	TotalFilesSent                      = "Total Files Sent"
	TotalISAPIExtensionRequests         = "Total ISAPI Extension Requests"
	TotalLockedErrors                   = "Total Locked Errors"
	TotalLogonAttempts                  = "Total Logon Attempts"
	TotalNonAnonymousUsers              = "Total NonAnonymous Users"
	TotalNotFoundErrors                 = "Total Not Found Errors"
	TotalRejectedAsyncIORequests        = "Total Rejected Async I/O Requests"
	TotalCopyRequests                   = "Total Copy Requests"
	TotalDeleteRequests                 = "Total Delete Requests"
	TotalGetRequests                    = "Total Get Requests"
	TotalHeadRequests                   = "Total Head Requests"
	TotalLockRequests                   = "Total Lock Requests"
	TotalMkcolRequests                  = "Total Mkcol Requests"
	TotalMoveRequests                   = "Total Move Requests"
	TotalOptionsRequests                = "Total Options Requests"
	TotalOtherRequests                  = "Total Other Request Methods"
	TotalPostRequests                   = "Total Post Requests"
	TotalPropfindRequests               = "Total Propfind Requests"
	TotalProppatchRequests              = "Total Proppatch Requests"
	TotalPutRequests                    = "Total Put Requests"
	TotalSearchRequests                 = "Total Search Requests"
	TotalTraceRequests                  = "Total Trace Requests"
	TotalUnlockRequests                 = "Total Unlock Requests"
)

func (c *Collector) buildWebService() error {
	var err error

	c.perfDataCollectorWebService, err = perfdata.NewCollector("Web Service", perfdata.InstanceAll, []string{
		CurrentAnonymousUsers,
		CurrentBlockedAsyncIORequests,
		CurrentCGIRequests,
		CurrentConnections,
		CurrentISAPIExtensionRequests,
		CurrentNonAnonymousUsers,
		ServiceUptime,
		TotalBytesReceived,
		TotalBytesSent,
		TotalAnonymousUsers,
		TotalBlockedAsyncIORequests,
		TotalCGIRequests,
		TotalConnectionAttemptsAllInstances,
		TotalFilesReceived,
		TotalFilesSent,
		TotalISAPIExtensionRequests,
		TotalLockedErrors,
		TotalLogonAttempts,
		TotalNonAnonymousUsers,
		TotalNotFoundErrors,
		TotalRejectedAsyncIORequests,
		TotalCopyRequests,
		TotalDeleteRequests,
		TotalGetRequests,
		TotalHeadRequests,
		TotalLockRequests,
		TotalMkcolRequests,
		TotalMoveRequests,
		TotalOptionsRequests,
		TotalOtherRequests,
		TotalPostRequests,
		TotalPropfindRequests,
		TotalProppatchRequests,
		TotalPutRequests,
		TotalSearchRequests,
		TotalTraceRequests,
		TotalUnlockRequests,
	})
	if err != nil {
		return fmt.Errorf("failed to create Web Service collector: %w", err)
	}

	c.currentAnonymousUsers = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "current_anonymous_users"),
		"Number of users who currently have an anonymous connection using the Web service (WebService.CurrentAnonymousUsers)",
		[]string{"site"},
		nil,
	)
	c.currentBlockedAsyncIORequests = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "current_blocked_async_io_requests"),
		"Current requests temporarily blocked due to bandwidth throttling settings (WebService.CurrentBlockedAsyncIORequests)",
		[]string{"site"},
		nil,
	)
	c.currentCGIRequests = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "current_cgi_requests"),
		"Current number of CGI requests being simultaneously processed by the Web service (WebService.CurrentCGIRequests)",
		[]string{"site"},
		nil,
	)
	c.currentConnections = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "current_connections"),
		"Current number of connections established with the Web service (WebService.CurrentConnections)",
		[]string{"site"},
		nil,
	)
	c.currentISAPIExtensionRequests = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "current_isapi_extension_requests"),
		"Current number of ISAPI requests being simultaneously processed by the Web service (WebService.CurrentISAPIExtensionRequests)",
		[]string{"site"},
		nil,
	)
	c.currentNonAnonymousUsers = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "current_non_anonymous_users"),
		"Number of users who currently have a non-anonymous connection using the Web service (WebService.CurrentNonAnonymousUsers)",
		[]string{"site"},
		nil,
	)
	c.serviceUptime = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "service_uptime"),
		"Number of seconds the WebService is up (WebService.ServiceUptime)",
		[]string{"site"},
		nil,
	)
	c.totalBytesReceived = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "received_bytes_total"),
		"Number of data bytes that have been received by the Web service (WebService.TotalBytesReceived)",
		[]string{"site"},
		nil,
	)
	c.totalBytesSent = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "sent_bytes_total"),
		"Number of data bytes that have been sent by the Web service (WebService.TotalBytesSent)",
		[]string{"site"},
		nil,
	)
	c.totalAnonymousUsers = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "anonymous_users_total"),
		"Total number of users who established an anonymous connection with the Web service (WebService.TotalAnonymousUsers)",
		[]string{"site"},
		nil,
	)
	c.totalBlockedAsyncIORequests = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "blocked_async_io_requests_total"),
		"Total requests temporarily blocked due to bandwidth throttling settings (WebService.TotalBlockedAsyncIORequests)",
		[]string{"site"},
		nil,
	)
	c.totalCGIRequests = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "cgi_requests_total"),
		"Total CGI requests is the total number of CGI requests (WebService.TotalCGIRequests)",
		[]string{"site"},
		nil,
	)
	c.totalConnectionAttemptsAllInstances = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "connection_attempts_all_instances_total"),
		"Number of connections that have been attempted using the Web service (WebService.TotalConnectionAttemptsAllInstances)",
		[]string{"site"},
		nil,
	)
	c.totalRequests = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "requests_total"),
		"Number of HTTP requests (WebService.TotalRequests)",
		[]string{"site", "method"},
		nil,
	)
	c.totalFilesReceived = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "files_received_total"),
		"Number of files received by the Web service (WebService.TotalFilesReceived)",
		[]string{"site"},
		nil,
	)
	c.totalFilesSent = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "files_sent_total"),
		"Number of files sent by the Web service (WebService.TotalFilesSent)",
		[]string{"site"},
		nil,
	)
	c.totalISAPIExtensionRequests = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "ipapi_extension_requests_total"),
		"ISAPI Extension Requests received (WebService.TotalISAPIExtensionRequests)",
		[]string{"site"},
		nil,
	)
	c.totalLockedErrors = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "locked_errors_total"),
		"Number of requests that couldn't be satisfied by the server because the requested resource was locked (WebService.TotalLockedErrors)",
		[]string{"site"},
		nil,
	)
	c.totalLogonAttempts = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "logon_attempts_total"),
		"Number of logons attempts to the Web Service (WebService.TotalLogonAttempts)",
		[]string{"site"},
		nil,
	)
	c.totalNonAnonymousUsers = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "non_anonymous_users_total"),
		"Number of users who established a non-anonymous connection with the Web service (WebService.TotalNonAnonymousUsers)",
		[]string{"site"},
		nil,
	)
	c.totalNotFoundErrors = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "not_found_errors_total"),
		"Number of requests that couldn't be satisfied by the server because the requested document could not be found (WebService.TotalNotFoundErrors)",
		[]string{"site"},
		nil,
	)
	c.totalRejectedAsyncIORequests = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "rejected_async_io_requests_total"),
		"Requests rejected due to bandwidth throttling settings (WebService.TotalRejectedAsyncIORequests)",
		[]string{"site"},
		nil,
	)

	return nil
}

func (c *Collector) collectWebService(ch chan<- prometheus.Metric) error {
	perfData, err := c.perfDataCollectorWebService.Collect()
	if err != nil {
		return fmt.Errorf("failed to collect Web Service metrics: %w", err)
	}

	deduplicateIISNames(perfData)

	for name, app := range perfData {
		if c.config.SiteExclude.MatchString(name) || !c.config.SiteInclude.MatchString(name) {
			continue
		}

		ch <- prometheus.MustNewConstMetric(
			c.currentAnonymousUsers,
			prometheus.GaugeValue,
			app[CurrentAnonymousUsers].FirstValue,
			name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.currentBlockedAsyncIORequests,
			prometheus.GaugeValue,
			app[CurrentBlockedAsyncIORequests].FirstValue,
			name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.currentCGIRequests,
			prometheus.GaugeValue,
			app[CurrentCGIRequests].FirstValue,
			name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.currentConnections,
			prometheus.GaugeValue,
			app[CurrentConnections].FirstValue,
			name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.currentISAPIExtensionRequests,
			prometheus.GaugeValue,
			app[CurrentISAPIExtensionRequests].FirstValue,
			name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.currentNonAnonymousUsers,
			prometheus.GaugeValue,
			app[CurrentNonAnonymousUsers].FirstValue,
			name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.serviceUptime,
			prometheus.GaugeValue,
			app[ServiceUptime].FirstValue,
			name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.totalBytesReceived,
			prometheus.CounterValue,
			app[TotalBytesReceived].FirstValue,
			name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.totalBytesSent,
			prometheus.CounterValue,
			app[TotalBytesSent].FirstValue,
			name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.totalAnonymousUsers,
			prometheus.CounterValue,
			app[TotalAnonymousUsers].FirstValue,
			name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.totalBlockedAsyncIORequests,
			prometheus.CounterValue,
			app[TotalBlockedAsyncIORequests].FirstValue,
			name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.totalCGIRequests,
			prometheus.CounterValue,
			app[TotalCGIRequests].FirstValue,
			name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.totalConnectionAttemptsAllInstances,
			prometheus.CounterValue,
			app[TotalConnectionAttemptsAllInstances].FirstValue,
			name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.totalFilesReceived,
			prometheus.CounterValue,
			app[TotalFilesReceived].FirstValue,
			name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.totalFilesSent,
			prometheus.CounterValue,
			app[TotalFilesSent].FirstValue,
			name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.totalISAPIExtensionRequests,
			prometheus.CounterValue,
			app[TotalISAPIExtensionRequests].FirstValue,
			name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.totalLockedErrors,
			prometheus.CounterValue,
			app[TotalLockedErrors].FirstValue,
			name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.totalLogonAttempts,
			prometheus.CounterValue,
			app[TotalLogonAttempts].FirstValue,
			name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.totalNonAnonymousUsers,
			prometheus.CounterValue,
			app[TotalNonAnonymousUsers].FirstValue,
			name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.totalNotFoundErrors,
			prometheus.CounterValue,
			app[TotalNotFoundErrors].FirstValue,
			name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.totalRejectedAsyncIORequests,
			prometheus.CounterValue,
			app[TotalRejectedAsyncIORequests].FirstValue,
			name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.totalRequests,
			prometheus.CounterValue,
			app[TotalOtherRequests].FirstValue,
			name,
			"other",
		)
		ch <- prometheus.MustNewConstMetric(
			c.totalRequests,
			prometheus.CounterValue,
			app[TotalCopyRequests].FirstValue,
			name,
			"COPY",
		)
		ch <- prometheus.MustNewConstMetric(
			c.totalRequests,
			prometheus.CounterValue,
			app[TotalDeleteRequests].FirstValue,
			name,
			"DELETE",
		)
		ch <- prometheus.MustNewConstMetric(
			c.totalRequests,
			prometheus.CounterValue,
			app[TotalGetRequests].FirstValue,
			name,
			"GET",
		)
		ch <- prometheus.MustNewConstMetric(
			c.totalRequests,
			prometheus.CounterValue,
			app[TotalHeadRequests].FirstValue,
			name,
			"HEAD",
		)
		ch <- prometheus.MustNewConstMetric(
			c.totalRequests,
			prometheus.CounterValue,
			app[TotalLockRequests].FirstValue,
			name,
			"LOCK",
		)
		ch <- prometheus.MustNewConstMetric(
			c.totalRequests,
			prometheus.CounterValue,
			app[TotalMkcolRequests].FirstValue,
			name,
			"MKCOL",
		)
		ch <- prometheus.MustNewConstMetric(
			c.totalRequests,
			prometheus.CounterValue,
			app[TotalMoveRequests].FirstValue,
			name,
			"MOVE",
		)
		ch <- prometheus.MustNewConstMetric(
			c.totalRequests,
			prometheus.CounterValue,
			app[TotalOptionsRequests].FirstValue,
			name,
			"OPTIONS",
		)
		ch <- prometheus.MustNewConstMetric(
			c.totalRequests,
			prometheus.CounterValue,
			app[TotalPostRequests].FirstValue,
			name,
			"POST",
		)
		ch <- prometheus.MustNewConstMetric(
			c.totalRequests,
			prometheus.CounterValue,
			app[TotalPropfindRequests].FirstValue,
			name,
			"PROPFIND",
		)
		ch <- prometheus.MustNewConstMetric(
			c.totalRequests,
			prometheus.CounterValue,
			app[TotalProppatchRequests].FirstValue,
			name,
			"PROPPATCH",
		)
		ch <- prometheus.MustNewConstMetric(
			c.totalRequests,
			prometheus.CounterValue,
			app[TotalPutRequests].FirstValue,
			name,
			"PUT",
		)
		ch <- prometheus.MustNewConstMetric(
			c.totalRequests,
			prometheus.CounterValue,
			app[TotalSearchRequests].FirstValue,
			name,
			"SEARCH",
		)
		ch <- prometheus.MustNewConstMetric(
			c.totalRequests,
			prometheus.CounterValue,
			app[TotalTraceRequests].FirstValue,
			name,
			"TRACE",
		)
		ch <- prometheus.MustNewConstMetric(
			c.totalRequests,
			prometheus.CounterValue,
			app[TotalUnlockRequests].FirstValue,
			name,
			"UNLOCK",
		)
	}

	return nil
}
