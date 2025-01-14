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

	"github.com/prometheus-community/windows_exporter/internal/pdh"
	"github.com/prometheus-community/windows_exporter/internal/types"
	"github.com/prometheus/client_golang/prometheus"
)

type collectorWebService struct {
	perfDataCollectorWebService *pdh.Collector
	perfDataObjectWebService    []perfDataCounterValuesWebService

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

type perfDataCounterValuesWebService struct {
	Name string

	WebServiceCurrentAnonymousUsers               float64 `perfdata:"Current Anonymous Users"`
	WebServiceCurrentBlockedAsyncIORequests       float64 `perfdata:"Current Blocked Async I/O Requests"`
	WebServiceCurrentCGIRequests                  float64 `perfdata:"Current CGI Requests"`
	WebServiceCurrentConnections                  float64 `perfdata:"Current Connections"`
	WebServiceCurrentISAPIExtensionRequests       float64 `perfdata:"Current ISAPI Extension Requests"`
	WebServiceCurrentNonAnonymousUsers            float64 `perfdata:"Current NonAnonymous Users"`
	WebServiceServiceUptime                       float64 `perfdata:"Service Uptime"`
	WebServiceTotalBytesReceived                  float64 `perfdata:"Total Bytes Received"`
	WebServiceTotalBytesSent                      float64 `perfdata:"Total Bytes Sent"`
	WebServiceTotalAnonymousUsers                 float64 `perfdata:"Total Anonymous Users"`
	WebServiceTotalBlockedAsyncIORequests         float64 `perfdata:"Total Blocked Async I/O Requests"`
	WebServiceTotalCGIRequests                    float64 `perfdata:"Total CGI Requests"`
	WebServiceTotalConnectionAttemptsAllInstances float64 `perfdata:"Total Connection Attempts (all instances)"`
	WebServiceTotalFilesReceived                  float64 `perfdata:"Total Files Received"`
	WebServiceTotalFilesSent                      float64 `perfdata:"Total Files Sent"`
	WebServiceTotalISAPIExtensionRequests         float64 `perfdata:"Total ISAPI Extension Requests"`
	WebServiceTotalLockedErrors                   float64 `perfdata:"Total Locked Errors"`
	WebServiceTotalLogonAttempts                  float64 `perfdata:"Total Logon Attempts"`
	WebServiceTotalNonAnonymousUsers              float64 `perfdata:"Total NonAnonymous Users"`
	WebServiceTotalNotFoundErrors                 float64 `perfdata:"Total Not Found Errors"`
	WebServiceTotalRejectedAsyncIORequests        float64 `perfdata:"Total Rejected Async I/O Requests"`
	WebServiceTotalCopyRequests                   float64 `perfdata:"Total Copy Requests"`
	WebServiceTotalDeleteRequests                 float64 `perfdata:"Total Delete Requests"`
	WebServiceTotalGetRequests                    float64 `perfdata:"Total Get Requests"`
	WebServiceTotalHeadRequests                   float64 `perfdata:"Total Head Requests"`
	WebServiceTotalLockRequests                   float64 `perfdata:"Total Lock Requests"`
	WebServiceTotalMkcolRequests                  float64 `perfdata:"Total Mkcol Requests"`
	WebServiceTotalMoveRequests                   float64 `perfdata:"Total Move Requests"`
	WebServiceTotalOptionsRequests                float64 `perfdata:"Total Options Requests"`
	WebServiceTotalOtherRequests                  float64 `perfdata:"Total Other Request Methods"`
	WebServiceTotalPostRequests                   float64 `perfdata:"Total Post Requests"`
	WebServiceTotalPropfindRequests               float64 `perfdata:"Total Propfind Requests"`
	WebServiceTotalProppatchRequests              float64 `perfdata:"Total Proppatch Requests"`
	WebServiceTotalPutRequests                    float64 `perfdata:"Total Put Requests"`
	WebServiceTotalSearchRequests                 float64 `perfdata:"Total Search Requests"`
	WebServiceTotalTraceRequests                  float64 `perfdata:"Total Trace Requests"`
	WebServiceTotalUnlockRequests                 float64 `perfdata:"Total Unlock Requests"`
}

func (p perfDataCounterValuesWebService) GetName() string {
	return p.Name
}

func (c *Collector) buildWebService() error {
	var err error

	c.perfDataCollectorWebService, err = pdh.NewCollector[perfDataCounterValuesWebService](pdh.CounterTypeRaw, "Web Service", pdh.InstancesAll)
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
	err := c.perfDataCollectorWebService.Collect(&c.perfDataObjectWebService)
	if err != nil {
		return fmt.Errorf("failed to collect Web Service metrics: %w", err)
	}

	deduplicateIISNames(c.perfDataObjectWebService)

	for _, data := range c.perfDataObjectWebService {
		if c.config.SiteExclude.MatchString(data.Name) || !c.config.SiteInclude.MatchString(data.Name) {
			continue
		}

		ch <- prometheus.MustNewConstMetric(
			c.webServiceCurrentAnonymousUsers,
			prometheus.GaugeValue,
			data.WebServiceCurrentAnonymousUsers,
			data.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.webServiceCurrentBlockedAsyncIORequests,
			prometheus.GaugeValue,
			data.WebServiceCurrentBlockedAsyncIORequests,
			data.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.webServiceCurrentCGIRequests,
			prometheus.GaugeValue,
			data.WebServiceCurrentCGIRequests,
			data.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.webServiceCurrentConnections,
			prometheus.GaugeValue,
			data.WebServiceCurrentConnections,
			data.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.webServiceCurrentISAPIExtensionRequests,
			prometheus.GaugeValue,
			data.WebServiceCurrentISAPIExtensionRequests,
			data.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.webServiceCurrentNonAnonymousUsers,
			prometheus.GaugeValue,
			data.WebServiceCurrentNonAnonymousUsers,
			data.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.webServiceServiceUptime,
			prometheus.GaugeValue,
			data.WebServiceServiceUptime,
			data.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.webServiceTotalBytesReceived,
			prometheus.CounterValue,
			data.WebServiceTotalBytesReceived,
			data.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.webServiceTotalBytesSent,
			prometheus.CounterValue,
			data.WebServiceTotalBytesSent,
			data.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.webServiceTotalAnonymousUsers,
			prometheus.CounterValue,
			data.WebServiceTotalAnonymousUsers,
			data.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.webServiceTotalBlockedAsyncIORequests,
			prometheus.CounterValue,
			data.WebServiceTotalBlockedAsyncIORequests,
			data.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.webServiceTotalCGIRequests,
			prometheus.CounterValue,
			data.WebServiceTotalCGIRequests,
			data.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.webServiceTotalConnectionAttemptsAllInstances,
			prometheus.CounterValue,
			data.WebServiceTotalConnectionAttemptsAllInstances,
			data.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.webServiceTotalFilesReceived,
			prometheus.CounterValue,
			data.WebServiceTotalFilesReceived,
			data.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.webServiceTotalFilesSent,
			prometheus.CounterValue,
			data.WebServiceTotalFilesSent,
			data.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.webServiceTotalISAPIExtensionRequests,
			prometheus.CounterValue,
			data.WebServiceTotalISAPIExtensionRequests,
			data.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.webServiceTotalLockedErrors,
			prometheus.CounterValue,
			data.WebServiceTotalLockedErrors,
			data.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.webServiceTotalLogonAttempts,
			prometheus.CounterValue,
			data.WebServiceTotalLogonAttempts,
			data.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.webServiceTotalNonAnonymousUsers,
			prometheus.CounterValue,
			data.WebServiceTotalNonAnonymousUsers,
			data.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.webServiceTotalNotFoundErrors,
			prometheus.CounterValue,
			data.WebServiceTotalNotFoundErrors,
			data.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.webServiceTotalRejectedAsyncIORequests,
			prometheus.CounterValue,
			data.WebServiceTotalRejectedAsyncIORequests,
			data.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.webServiceTotalRequests,
			prometheus.CounterValue,
			data.WebServiceTotalOtherRequests,
			data.Name,
			"other",
		)
		ch <- prometheus.MustNewConstMetric(
			c.webServiceTotalRequests,
			prometheus.CounterValue,
			data.WebServiceTotalCopyRequests,
			data.Name,
			"COPY",
		)
		ch <- prometheus.MustNewConstMetric(
			c.webServiceTotalRequests,
			prometheus.CounterValue,
			data.WebServiceTotalDeleteRequests,
			data.Name,
			"DELETE",
		)
		ch <- prometheus.MustNewConstMetric(
			c.webServiceTotalRequests,
			prometheus.CounterValue,
			data.WebServiceTotalGetRequests,
			data.Name,
			"GET",
		)
		ch <- prometheus.MustNewConstMetric(
			c.webServiceTotalRequests,
			prometheus.CounterValue,
			data.WebServiceTotalHeadRequests,
			data.Name,
			"HEAD",
		)
		ch <- prometheus.MustNewConstMetric(
			c.webServiceTotalRequests,
			prometheus.CounterValue,
			data.WebServiceTotalLockRequests,
			data.Name,
			"LOCK",
		)
		ch <- prometheus.MustNewConstMetric(
			c.webServiceTotalRequests,
			prometheus.CounterValue,
			data.WebServiceTotalMkcolRequests,
			data.Name,
			"MKCOL",
		)
		ch <- prometheus.MustNewConstMetric(
			c.webServiceTotalRequests,
			prometheus.CounterValue,
			data.WebServiceTotalMoveRequests,
			data.Name,
			"MOVE",
		)
		ch <- prometheus.MustNewConstMetric(
			c.webServiceTotalRequests,
			prometheus.CounterValue,
			data.WebServiceTotalOptionsRequests,
			data.Name,
			"OPTIONS",
		)
		ch <- prometheus.MustNewConstMetric(
			c.webServiceTotalRequests,
			prometheus.CounterValue,
			data.WebServiceTotalPostRequests,
			data.Name,
			"POST",
		)
		ch <- prometheus.MustNewConstMetric(
			c.webServiceTotalRequests,
			prometheus.CounterValue,
			data.WebServiceTotalPropfindRequests,
			data.Name,
			"PROPFIND",
		)
		ch <- prometheus.MustNewConstMetric(
			c.webServiceTotalRequests,
			prometheus.CounterValue,
			data.WebServiceTotalProppatchRequests,
			data.Name,
			"PROPPATCH",
		)
		ch <- prometheus.MustNewConstMetric(
			c.webServiceTotalRequests,
			prometheus.CounterValue,
			data.WebServiceTotalPutRequests,
			data.Name,
			"PUT",
		)
		ch <- prometheus.MustNewConstMetric(
			c.webServiceTotalRequests,
			prometheus.CounterValue,
			data.WebServiceTotalSearchRequests,
			data.Name,
			"SEARCH",
		)
		ch <- prometheus.MustNewConstMetric(
			c.webServiceTotalRequests,
			prometheus.CounterValue,
			data.WebServiceTotalTraceRequests,
			data.Name,
			"TRACE",
		)
		ch <- prometheus.MustNewConstMetric(
			c.webServiceTotalRequests,
			prometheus.CounterValue,
			data.WebServiceTotalUnlockRequests,
			data.Name,
			"UNLOCK",
		)
	}

	return nil
}
