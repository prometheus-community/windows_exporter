// SPDX-License-Identifier: MIT
//
// Copyright The Prometheus Authors

//go:build windows

package exchange

import (
	"fmt"

	"github.com/prometheus-community/windows_exporter/internal/pdh"
	"github.com/prometheus-community/windows_exporter/internal/types"
	"github.com/prometheus-community/windows_exporter/internal/utils"
	"github.com/prometheus/client_golang/prometheus"
)

type collectorRpcClientAccess struct {
	perfDataCollectorRpcClientAccess *pdh.Collector[perfDataCounterValuesRpcClientAccess]
	perfDataObjectRpcClientAccess    []perfDataCounterValuesRpcClientAccess

	activeUserCount     *prometheus.Desc
	connectionCount     *prometheus.Desc
	rpcAveragedLatency  *prometheus.Desc
	rpcOperationsPerSec *prometheus.Desc
	rpcRequests         *prometheus.Desc
	userCount           *prometheus.Desc
}

type perfDataCounterValuesRpcClientAccess struct {
	RpcAveragedLatency  float64 `perfdata:"RPC Averaged Latency"`
	RpcRequests         float64 `perfdata:"RPC Requests"`
	ActiveUserCount     float64 `perfdata:"Active User Count"`
	ConnectionCount     float64 `perfdata:"Connection Count"`
	RpcOperationsPerSec float64 `perfdata:"RPC Operations/sec"`
	UserCount           float64 `perfdata:"User Count"`
}

func (c *Collector) buildRpcClientAccess() error {
	var err error

	c.perfDataCollectorRpcClientAccess, err = pdh.NewCollector[perfDataCounterValuesRpcClientAccess](c.logger, pdh.CounterTypeRaw, "MSExchange RpcClientAccess", pdh.InstancesAll)
	if err != nil {
		return fmt.Errorf("failed to create MSExchange RpcClientAccess collector: %w", err)
	}

	c.rpcAveragedLatency = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "rpc_avg_latency_seconds"),
		"The latency in seconds averaged for the past 1024 packets",
		nil,
		nil,
	)
	c.rpcRequests = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "rpc_requests"),
		"Number of client requests currently being processed by the RPC Client Access service",
		nil,
		nil,
	)
	c.activeUserCount = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "rpc_active_users"),
		"Number of unique users that have shown some kind of activity in the last 2 minutes",
		nil,
		nil,
	)
	c.connectionCount = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "rpc_connections"),
		"Total number of client connections maintained",
		nil,
		nil,
	)
	c.rpcOperationsPerSec = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "rpc_operations_total"),
		"The rate at which RPC operations occur",
		nil,
		nil,
	)
	c.userCount = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "rpc_users"),
		"Number of users",
		nil,
		nil,
	)

	return nil
}

func (c *Collector) collectRpcClientAccess(ch chan<- prometheus.Metric) error {
	err := c.perfDataCollectorRpcClientAccess.Collect(&c.perfDataObjectRpcClientAccess)
	if err != nil {
		return fmt.Errorf("failed to collect MSExchange RpcClientAccess: %w", err)
	}

	for _, data := range c.perfDataObjectRpcClientAccess {
		ch <- prometheus.MustNewConstMetric(
			c.rpcAveragedLatency,
			prometheus.GaugeValue,
			utils.MilliSecToSec(data.RpcAveragedLatency),
		)

		ch <- prometheus.MustNewConstMetric(
			c.rpcRequests,
			prometheus.GaugeValue,
			data.RpcRequests,
		)

		ch <- prometheus.MustNewConstMetric(
			c.activeUserCount,
			prometheus.GaugeValue,
			data.ActiveUserCount,
		)

		ch <- prometheus.MustNewConstMetric(
			c.connectionCount,
			prometheus.GaugeValue,
			data.ConnectionCount,
		)

		ch <- prometheus.MustNewConstMetric(
			c.rpcOperationsPerSec,
			prometheus.CounterValue,
			data.RpcOperationsPerSec,
		)

		ch <- prometheus.MustNewConstMetric(
			c.userCount,
			prometheus.GaugeValue,
			data.UserCount,
		)
	}

	return nil
}
