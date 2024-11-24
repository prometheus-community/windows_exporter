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

package iis

import (
	"fmt"

	"github.com/prometheus-community/windows_exporter/internal/perfdata"
	"github.com/prometheus-community/windows_exporter/internal/types"
	"github.com/prometheus/client_golang/prometheus"
)

type collectorWebService struct {
	perfDataCollectorWebService *perfdata.Collector

	webServiceCurrentAnonymousUsers               *prometheus.Desc
	webServiceCurrentBlockedAsyncIORequests       *prometheus.Desc
	webServiceCurrentCGIRequests                  *prometheus.Desc
	webServiceCurrentConnections                  *prometheus.Desc
	webServiceCurrentISAPIExtensionRequests       *prometheus.Desc
	webServiceCurrentNonAnonymousUsers            *prometheus.Desc
	webServiceServiceUptime                       *prometheus.Desc
	webServiceTotalBytesReceived                  *prometheus.Desc
	webServiceTotalBytesSent                      *prometheus.Desc
	webServiceTotalAnonymousUsers                 *prometheus.Desc
	webServiceTotalBlockedAsyncIORequests         *prometheus.Desc
	webServiceTotalCGIRequests                    *prometheus.Desc
	webServiceTotalConnectionAttemptsAllInstances *prometheus.Desc
	webServiceTotalRequests                       *prometheus.Desc
	webServiceTotalFilesReceived                  *prometheus.Desc
	webServiceTotalFilesSent                      *prometheus.Desc
	webServiceTotalISAPIExtensionRequests         *prometheus.Desc
	webServiceTotalLockedErrors                   *prometheus.Desc
	webServiceTotalLogonAttempts                  *prometheus.Desc
	webServiceTotalNonAnonymousUsers              *prometheus.Desc
	webServiceTotalNotFoundErrors                 *prometheus.Desc
	webServiceTotalRejectedAsyncIORequests        *prometheus.Desc
}

const (
	webServiceCurrentAnonymousUsers               = "Current Anonymous Users"
	webServiceCurrentBlockedAsyncIORequests       = "Current Blocked Async I/O Requests"
	webServiceCurrentCGIRequests                  = "Current CGI Requests"
	webServiceCurrentConnections                  = "Current Connections"
	webServiceCurrentISAPIExtensionRequests       = "Current ISAPI Extension Requests"
	webServiceCurrentNonAnonymousUsers            = "Current NonAnonymous Users"
	webServiceServiceUptime                       = "Service Uptime"
	webServiceTotalBytesReceived                  = "Total Bytes Received"
	webServiceTotalBytesSent                      = "Total Bytes Sent"
	webServiceTotalAnonymousUsers                 = "Total Anonymous Users"
	webServiceTotalBlockedAsyncIORequests         = "Total Blocked Async I/O Requests"
	webServiceTotalCGIRequests                    = "Total CGI Requests"
	webServiceTotalConnectionAttemptsAllInstances = "Total Connection Attempts (all instances)"
	webServiceTotalFilesReceived                  = "Total Files Received"
	webServiceTotalFilesSent                      = "Total Files Sent"
	webServiceTotalISAPIExtensionRequests         = "Total ISAPI Extension Requests"
	webServiceTotalLockedErrors                   = "Total Locked Errors"
	webServiceTotalLogonAttempts                  = "Total Logon Attempts"
	webServiceTotalNonAnonymousUsers              = "Total NonAnonymous Users"
	webServiceTotalNotFoundErrors                 = "Total Not Found Errors"
	webServiceTotalRejectedAsyncIORequests        = "Total Rejected Async I/O Requests"
	webServiceTotalCopyRequests                   = "Total Copy Requests"
	webServiceTotalDeleteRequests                 = "Total Delete Requests"
	webServiceTotalGetRequests                    = "Total Get Requests"
	webServiceTotalHeadRequests                   = "Total Head Requests"
	webServiceTotalLockRequests                   = "Total Lock Requests"
	webServiceTotalMkcolRequests                  = "Total Mkcol Requests"
	webServiceTotalMoveRequests                   = "Total Move Requests"
	webServiceTotalOptionsRequests                = "Total Options Requests"
	webServiceTotalOtherRequests                  = "Total Other Request Methods"
	webServiceTotalPostRequests                   = "Total Post Requests"
	webServiceTotalPropfindRequests               = "Total Propfind Requests"
	webServiceTotalProppatchRequests              = "Total Proppatch Requests"
	webServiceTotalPutRequests                    = "Total Put Requests"
	webServiceTotalSearchRequests                 = "Total Search Requests"
	webServiceTotalTraceRequests                  = "Total Trace Requests"
	webServiceTotalUnlockRequests                 = "Total Unlock Requests"
)

func (c *Collector) buildWebService() error {
	var err error

	c.perfDataCollectorWebService, err = perfdata.NewCollector("Web Service", perfdata.InstancesAll, []string{
		webServiceCurrentAnonymousUsers,
		webServiceCurrentBlockedAsyncIORequests,
		webServiceCurrentCGIRequests,
		webServiceCurrentConnections,
		webServiceCurrentISAPIExtensionRequests,
		webServiceCurrentNonAnonymousUsers,
		webServiceServiceUptime,
		webServiceTotalBytesReceived,
		webServiceTotalBytesSent,
		webServiceTotalAnonymousUsers,
		webServiceTotalBlockedAsyncIORequests,
		webServiceTotalCGIRequests,
		webServiceTotalConnectionAttemptsAllInstances,
		webServiceTotalFilesReceived,
		webServiceTotalFilesSent,
		webServiceTotalISAPIExtensionRequests,
		webServiceTotalLockedErrors,
		webServiceTotalLogonAttempts,
		webServiceTotalNonAnonymousUsers,
		webServiceTotalNotFoundErrors,
		webServiceTotalRejectedAsyncIORequests,
		webServiceTotalCopyRequests,
		webServiceTotalDeleteRequests,
		webServiceTotalGetRequests,
		webServiceTotalHeadRequests,
		webServiceTotalLockRequests,
		webServiceTotalMkcolRequests,
		webServiceTotalMoveRequests,
		webServiceTotalOptionsRequests,
		webServiceTotalOtherRequests,
		webServiceTotalPostRequests,
		webServiceTotalPropfindRequests,
		webServiceTotalProppatchRequests,
		webServiceTotalPutRequests,
		webServiceTotalSearchRequests,
		webServiceTotalTraceRequests,
		webServiceTotalUnlockRequests,
	})
	if err != nil {
		return fmt.Errorf("failed to create Web Service collector: %w", err)
	}

	c.webServiceCurrentAnonymousUsers = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "current_anonymous_users"),
		"Number of users who currently have an anonymous connection using the Web service (WebService.CurrentAnonymousUsers)",
		[]string{"site"},
		nil,
	)
	c.webServiceCurrentBlockedAsyncIORequests = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "current_blocked_async_io_requests"),
		"Current requests temporarily blocked due to bandwidth throttling settings (WebService.CurrentBlockedAsyncIORequests)",
		[]string{"site"},
		nil,
	)
	c.webServiceCurrentCGIRequests = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "current_cgi_requests"),
		"Current number of CGI requests being simultaneously processed by the Web service (WebService.CurrentCGIRequests)",
		[]string{"site"},
		nil,
	)
	c.webServiceCurrentConnections = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "current_connections"),
		"Current number of connections established with the Web service (WebService.CurrentConnections)",
		[]string{"site"},
		nil,
	)
	c.webServiceCurrentISAPIExtensionRequests = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "current_isapi_extension_requests"),
		"Current number of ISAPI requests being simultaneously processed by the Web service (WebService.CurrentISAPIExtensionRequests)",
		[]string{"site"},
		nil,
	)
	c.webServiceCurrentNonAnonymousUsers = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "current_non_anonymous_users"),
		"Number of users who currently have a non-anonymous connection using the Web service (WebService.CurrentNonAnonymousUsers)",
		[]string{"site"},
		nil,
	)
	c.webServiceServiceUptime = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "service_uptime"),
		"Number of seconds the WebService is up (WebService.ServiceUptime)",
		[]string{"site"},
		nil,
	)
	c.webServiceTotalBytesReceived = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "received_bytes_total"),
		"Number of data bytes that have been received by the Web service (WebService.TotalBytesReceived)",
		[]string{"site"},
		nil,
	)
	c.webServiceTotalBytesSent = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "sent_bytes_total"),
		"Number of data bytes that have been sent by the Web service (WebService.TotalBytesSent)",
		[]string{"site"},
		nil,
	)
	c.webServiceTotalAnonymousUsers = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "anonymous_users_total"),
		"Total number of users who established an anonymous connection with the Web service (WebService.TotalAnonymousUsers)",
		[]string{"site"},
		nil,
	)
	c.webServiceTotalBlockedAsyncIORequests = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "blocked_async_io_requests_total"),
		"Total requests temporarily blocked due to bandwidth throttling settings (WebService.TotalBlockedAsyncIORequests)",
		[]string{"site"},
		nil,
	)
	c.webServiceTotalCGIRequests = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "cgi_requests_total"),
		"Total CGI requests is the total number of CGI requests (WebService.TotalCGIRequests)",
		[]string{"site"},
		nil,
	)
	c.webServiceTotalConnectionAttemptsAllInstances = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "connection_attempts_all_instances_total"),
		"Number of connections that have been attempted using the Web service (WebService.TotalConnectionAttemptsAllInstances)",
		[]string{"site"},
		nil,
	)
	c.webServiceTotalRequests = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "requests_total"),
		"Number of HTTP requests (WebService.TotalRequests)",
		[]string{"site", "method"},
		nil,
	)
	c.webServiceTotalFilesReceived = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "files_received_total"),
		"Number of files received by the Web service (WebService.TotalFilesReceived)",
		[]string{"site"},
		nil,
	)
	c.webServiceTotalFilesSent = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "files_sent_total"),
		"Number of files sent by the Web service (WebService.TotalFilesSent)",
		[]string{"site"},
		nil,
	)
	c.webServiceTotalISAPIExtensionRequests = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "ipapi_extension_requests_total"),
		"ISAPI Extension Requests received (WebService.TotalISAPIExtensionRequests)",
		[]string{"site"},
		nil,
	)
	c.webServiceTotalLockedErrors = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "locked_errors_total"),
		"Number of requests that couldn't be satisfied by the server because the requested resource was locked (WebService.TotalLockedErrors)",
		[]string{"site"},
		nil,
	)
	c.webServiceTotalLogonAttempts = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "logon_attempts_total"),
		"Number of logons attempts to the Web Service (WebService.TotalLogonAttempts)",
		[]string{"site"},
		nil,
	)
	c.webServiceTotalNonAnonymousUsers = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "non_anonymous_users_total"),
		"Number of users who established a non-anonymous connection with the Web service (WebService.TotalNonAnonymousUsers)",
		[]string{"site"},
		nil,
	)
	c.webServiceTotalNotFoundErrors = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "not_found_errors_total"),
		"Number of requests that couldn't be satisfied by the server because the requested document could not be found (WebService.TotalNotFoundErrors)",
		[]string{"site"},
		nil,
	)
	c.webServiceTotalRejectedAsyncIORequests = prometheus.NewDesc(
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
			c.webServiceCurrentAnonymousUsers,
			prometheus.GaugeValue,
			app[webServiceCurrentAnonymousUsers].FirstValue,
			name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.webServiceCurrentBlockedAsyncIORequests,
			prometheus.GaugeValue,
			app[webServiceCurrentBlockedAsyncIORequests].FirstValue,
			name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.webServiceCurrentCGIRequests,
			prometheus.GaugeValue,
			app[webServiceCurrentCGIRequests].FirstValue,
			name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.webServiceCurrentConnections,
			prometheus.GaugeValue,
			app[webServiceCurrentConnections].FirstValue,
			name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.webServiceCurrentISAPIExtensionRequests,
			prometheus.GaugeValue,
			app[webServiceCurrentISAPIExtensionRequests].FirstValue,
			name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.webServiceCurrentNonAnonymousUsers,
			prometheus.GaugeValue,
			app[webServiceCurrentNonAnonymousUsers].FirstValue,
			name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.webServiceServiceUptime,
			prometheus.GaugeValue,
			app[webServiceServiceUptime].FirstValue,
			name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.webServiceTotalBytesReceived,
			prometheus.CounterValue,
			app[webServiceTotalBytesReceived].FirstValue,
			name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.webServiceTotalBytesSent,
			prometheus.CounterValue,
			app[webServiceTotalBytesSent].FirstValue,
			name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.webServiceTotalAnonymousUsers,
			prometheus.CounterValue,
			app[webServiceTotalAnonymousUsers].FirstValue,
			name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.webServiceTotalBlockedAsyncIORequests,
			prometheus.CounterValue,
			app[webServiceTotalBlockedAsyncIORequests].FirstValue,
			name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.webServiceTotalCGIRequests,
			prometheus.CounterValue,
			app[webServiceTotalCGIRequests].FirstValue,
			name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.webServiceTotalConnectionAttemptsAllInstances,
			prometheus.CounterValue,
			app[webServiceTotalConnectionAttemptsAllInstances].FirstValue,
			name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.webServiceTotalFilesReceived,
			prometheus.CounterValue,
			app[webServiceTotalFilesReceived].FirstValue,
			name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.webServiceTotalFilesSent,
			prometheus.CounterValue,
			app[webServiceTotalFilesSent].FirstValue,
			name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.webServiceTotalISAPIExtensionRequests,
			prometheus.CounterValue,
			app[webServiceTotalISAPIExtensionRequests].FirstValue,
			name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.webServiceTotalLockedErrors,
			prometheus.CounterValue,
			app[webServiceTotalLockedErrors].FirstValue,
			name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.webServiceTotalLogonAttempts,
			prometheus.CounterValue,
			app[webServiceTotalLogonAttempts].FirstValue,
			name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.webServiceTotalNonAnonymousUsers,
			prometheus.CounterValue,
			app[webServiceTotalNonAnonymousUsers].FirstValue,
			name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.webServiceTotalNotFoundErrors,
			prometheus.CounterValue,
			app[webServiceTotalNotFoundErrors].FirstValue,
			name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.webServiceTotalRejectedAsyncIORequests,
			prometheus.CounterValue,
			app[webServiceTotalRejectedAsyncIORequests].FirstValue,
			name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.webServiceTotalRequests,
			prometheus.CounterValue,
			app[webServiceTotalOtherRequests].FirstValue,
			name,
			"other",
		)
		ch <- prometheus.MustNewConstMetric(
			c.webServiceTotalRequests,
			prometheus.CounterValue,
			app[webServiceTotalCopyRequests].FirstValue,
			name,
			"COPY",
		)
		ch <- prometheus.MustNewConstMetric(
			c.webServiceTotalRequests,
			prometheus.CounterValue,
			app[webServiceTotalDeleteRequests].FirstValue,
			name,
			"DELETE",
		)
		ch <- prometheus.MustNewConstMetric(
			c.webServiceTotalRequests,
			prometheus.CounterValue,
			app[webServiceTotalGetRequests].FirstValue,
			name,
			"GET",
		)
		ch <- prometheus.MustNewConstMetric(
			c.webServiceTotalRequests,
			prometheus.CounterValue,
			app[webServiceTotalHeadRequests].FirstValue,
			name,
			"HEAD",
		)
		ch <- prometheus.MustNewConstMetric(
			c.webServiceTotalRequests,
			prometheus.CounterValue,
			app[webServiceTotalLockRequests].FirstValue,
			name,
			"LOCK",
		)
		ch <- prometheus.MustNewConstMetric(
			c.webServiceTotalRequests,
			prometheus.CounterValue,
			app[webServiceTotalMkcolRequests].FirstValue,
			name,
			"MKCOL",
		)
		ch <- prometheus.MustNewConstMetric(
			c.webServiceTotalRequests,
			prometheus.CounterValue,
			app[webServiceTotalMoveRequests].FirstValue,
			name,
			"MOVE",
		)
		ch <- prometheus.MustNewConstMetric(
			c.webServiceTotalRequests,
			prometheus.CounterValue,
			app[webServiceTotalOptionsRequests].FirstValue,
			name,
			"OPTIONS",
		)
		ch <- prometheus.MustNewConstMetric(
			c.webServiceTotalRequests,
			prometheus.CounterValue,
			app[webServiceTotalPostRequests].FirstValue,
			name,
			"POST",
		)
		ch <- prometheus.MustNewConstMetric(
			c.webServiceTotalRequests,
			prometheus.CounterValue,
			app[webServiceTotalPropfindRequests].FirstValue,
			name,
			"PROPFIND",
		)
		ch <- prometheus.MustNewConstMetric(
			c.webServiceTotalRequests,
			prometheus.CounterValue,
			app[webServiceTotalProppatchRequests].FirstValue,
			name,
			"PROPPATCH",
		)
		ch <- prometheus.MustNewConstMetric(
			c.webServiceTotalRequests,
			prometheus.CounterValue,
			app[webServiceTotalPutRequests].FirstValue,
			name,
			"PUT",
		)
		ch <- prometheus.MustNewConstMetric(
			c.webServiceTotalRequests,
			prometheus.CounterValue,
			app[webServiceTotalSearchRequests].FirstValue,
			name,
			"SEARCH",
		)
		ch <- prometheus.MustNewConstMetric(
			c.webServiceTotalRequests,
			prometheus.CounterValue,
			app[webServiceTotalTraceRequests].FirstValue,
			name,
			"TRACE",
		)
		ch <- prometheus.MustNewConstMetric(
			c.webServiceTotalRequests,
			prometheus.CounterValue,
			app[webServiceTotalUnlockRequests].FirstValue,
			name,
			"UNLOCK",
		)
	}

	return nil
}
