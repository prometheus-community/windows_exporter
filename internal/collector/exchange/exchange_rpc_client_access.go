package exchange

import (
	"errors"
	"fmt"
	"log/slog"

	"github.com/prometheus-community/windows_exporter/internal/perfdata"
	v1 "github.com/prometheus-community/windows_exporter/internal/perfdata/v1"
	"github.com/prometheus-community/windows_exporter/internal/types"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	rpcAveragedLatency = "RPC Averaged Latency"
	rpcRequests        = "RPC Requests"
	// activeUserCount    = "Active User Count"
	connectionCount     = "Connection Count"
	rpcOperationsPerSec = "RPC Operations/sec"
	userCount           = "User Count"
)

// Perflib: [29366] MSExchange RpcClientAccess.
type perflibRPCClientAccess struct {
	RPCAveragedLatency  float64 `perflib:"RPC Averaged Latency"`
	RPCRequests         float64 `perflib:"RPC Requests"`
	ActiveUserCount     float64 `perflib:"Active User Count"`
	ConnectionCount     float64 `perflib:"Connection Count"`
	RPCOperationsPerSec float64 `perflib:"RPC Operations/sec"`
	UserCount           float64 `perflib:"User Count"`
}

func (c *Collector) buildRPC() error {
	counters := []string{
		rpcAveragedLatency,
		rpcRequests,
		activeUserCount,
		connectionCount,
		rpcOperationsPerSec,
		userCount,
	}

	var err error

	c.perfDataCollectorRpcClientAccess, err = perfdata.NewCollector(perfdata.V2, "MSExchange RpcClientAccess", perfdata.AllInstances, counters)
	if err != nil {
		return fmt.Errorf("failed to create MSExchange RpcClientAccess collector: %w", err)
	}

	return nil
}

func (c *Collector) collectRPC(ctx *types.ScrapeContext, logger *slog.Logger, ch chan<- prometheus.Metric) error {
	var data []perflibRPCClientAccess

	if err := v1.UnmarshalObject(ctx.PerfObjects["MSExchange RpcClientAccess"], &data, logger); err != nil {
		return err
	}

	for _, rpc := range data {
		ch <- prometheus.MustNewConstMetric(
			c.rpcAveragedLatency,
			prometheus.GaugeValue,
			c.msToSec(rpc.RPCAveragedLatency),
		)
		ch <- prometheus.MustNewConstMetric(
			c.rpcRequests,
			prometheus.GaugeValue,
			rpc.RPCRequests,
		)
		ch <- prometheus.MustNewConstMetric(
			c.activeUserCount,
			prometheus.GaugeValue,
			rpc.ActiveUserCount,
		)
		ch <- prometheus.MustNewConstMetric(
			c.connectionCount,
			prometheus.GaugeValue,
			rpc.ConnectionCount,
		)
		ch <- prometheus.MustNewConstMetric(
			c.rpcOperationsPerSec,
			prometheus.CounterValue,
			rpc.RPCOperationsPerSec,
		)
		ch <- prometheus.MustNewConstMetric(
			c.userCount,
			prometheus.GaugeValue,
			rpc.UserCount,
		)
	}

	return nil
}

func (c *Collector) collectPDHRPC(ch chan<- prometheus.Metric) error {
	perfData, err := c.perfDataCollectorRpcClientAccess.Collect()
	if err != nil {
		return fmt.Errorf("failed to collect MSExchange RpcClientAccess: %w", err)
	}

	if len(perfData) == 0 {
		return errors.New("perflib query for MSExchange RpcClientAccess returned empty result set")
	}

	for _, data := range perfData {
		ch <- prometheus.MustNewConstMetric(
			c.rpcAveragedLatency,
			prometheus.GaugeValue,
			c.msToSec(data[rpcAveragedLatency].FirstValue),
		)
		ch <- prometheus.MustNewConstMetric(
			c.rpcRequests,
			prometheus.GaugeValue,
			data[rpcRequests].FirstValue,
		)
		ch <- prometheus.MustNewConstMetric(
			c.activeUserCount,
			prometheus.GaugeValue,
			data[activeUserCount].FirstValue,
		)
		ch <- prometheus.MustNewConstMetric(
			c.connectionCount,
			prometheus.GaugeValue,
			data[connectionCount].FirstValue,
		)
		ch <- prometheus.MustNewConstMetric(
			c.rpcOperationsPerSec,
			prometheus.CounterValue,
			data[rpcOperationsPerSec].FirstValue,
		)
		ch <- prometheus.MustNewConstMetric(
			c.userCount,
			prometheus.GaugeValue,
			data[userCount].FirstValue,
		)
	}

	return nil
}
