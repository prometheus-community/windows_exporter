//go:build windows
// +build windows

package collector

import (
	"strings"

	"github.com/prometheus-community/windows_exporter/log"
	"github.com/prometheus/client_golang/prometheus"
)

func init() {
	registerCollector("remote_fx", NewRemoteFx, "RemoteFX Network", "RemoteFX Graphics")
}

// A RemoteFxNetworkCollector is a Prometheus collector for
// WMI Win32_PerfRawData_Counters_RemoteFXNetwork & Win32_PerfRawData_Counters_RemoteFXGraphics metrics
// https://wutils.com/wmi/root/cimv2/win32_perfrawdata_counters_remotefxnetwork/
// https://wutils.com/wmi/root/cimv2/win32_perfrawdata_counters_remotefxgraphics/

type RemoteFxCollector struct {
	// net
	BaseTCPRTT               *prometheus.Desc
	BaseUDPRTT               *prometheus.Desc
	CurrentTCPBandwidth      *prometheus.Desc
	CurrentTCPRTT            *prometheus.Desc
	CurrentUDPBandwidth      *prometheus.Desc
	CurrentUDPRTT            *prometheus.Desc
	TotalReceivedBytes       *prometheus.Desc
	TotalSentBytes           *prometheus.Desc
	UDPPacketsReceivedPersec *prometheus.Desc
	UDPPacketsSentPersec     *prometheus.Desc

	//gfx
	AverageEncodingTime                         *prometheus.Desc
	FrameQuality                                *prometheus.Desc
	FramesSkippedPerSecondInsufficientResources *prometheus.Desc
	GraphicsCompressionratio                    *prometheus.Desc
	InputFramesPerSecond                        *prometheus.Desc
	OutputFramesPerSecond                       *prometheus.Desc
	SourceFramesPerSecond                       *prometheus.Desc
}

// NewRemoteFx ...
func NewRemoteFx() (Collector, error) {
	const subsystem = "remote_fx"
	return &RemoteFxCollector{
		// net
		BaseTCPRTT: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "net_base_tcp_rtt_seconds"),
			"Base TCP round-trip time (RTT) detected in seconds",
			[]string{"session_name"},
			nil,
		),
		BaseUDPRTT: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "net_base_udp_rtt_seconds"),
			"Base UDP round-trip time (RTT) detected in seconds.",
			[]string{"session_name"},
			nil,
		),
		CurrentTCPBandwidth: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "net_current_tcp_bandwidth"),
			"TCP Bandwidth detected in bytes per second.",
			[]string{"session_name"},
			nil,
		),
		CurrentTCPRTT: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "net_current_tcp_rtt_seconds"),
			"Average TCP round-trip time (RTT) detected in seconds.",
			[]string{"session_name"},
			nil,
		),
		CurrentUDPBandwidth: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "net_current_udp_bandwidth"),
			"UDP Bandwidth detected in bytes per second.",
			[]string{"session_name"},
			nil,
		),
		CurrentUDPRTT: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "net_current_udp_rtt_seconds"),
			"Average UDP round-trip time (RTT) detected in seconds.",
			[]string{"session_name"},
			nil,
		),
		TotalReceivedBytes: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "net_received_bytes_total"),
			"(TotalReceivedBytes)",
			[]string{"session_name"},
			nil,
		),
		TotalSentBytes: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "net_sent_bytes_total"),
			"(TotalSentBytes)",
			[]string{"session_name"},
			nil,
		),
		UDPPacketsReceivedPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "net_udp_packets_received_total"),
			"Rate in packets per second at which packets are received over UDP.",
			[]string{"session_name"},
			nil,
		),
		UDPPacketsSentPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "net_udp_packets_sent_total"),
			"Rate in packets per second at which packets are sent over UDP.",
			[]string{"session_name"},
			nil,
		),

		//gfx
		AverageEncodingTime: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "gfx_average_encoding_time_seconds"),
			"Average frame encoding time in seconds",
			[]string{"session_name"},
			nil,
		),
		FrameQuality: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "gfx_frame_quality"),
			"Quality of the output frame expressed as a percentage of the quality of the source frame.",
			[]string{"session_name"},
			nil,
		),
		FramesSkippedPerSecondInsufficientResources: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "gfx_frames_skipped_insufficient_resource_total"),
			"Number of frames skipped per second due to insufficient client resources.",
			[]string{"session_name", "resource"},
			nil,
		),
		GraphicsCompressionratio: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "gfx_graphics_compression_ratio"),
			"Ratio of the number of bytes encoded to the number of bytes input.",
			[]string{"session_name"},
			nil,
		),
		InputFramesPerSecond: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "gfx_input_frames_total"),
			"Number of sources frames provided as input to RemoteFX graphics per second.",
			[]string{"session_name"},
			nil,
		),
		OutputFramesPerSecond: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "gfx_output_frames_total"),
			"Number of frames sent to the client per second.",
			[]string{"session_name"},
			nil,
		),
		SourceFramesPerSecond: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "gfx_source_frames_total"),
			"Number of frames composed by the source (DWM) per second.",
			[]string{"session_name"},
			nil,
		),
	}, nil
}

// Collect sends the metric values for each metric
// to the provided prometheus Metric channel.
func (c *RemoteFxCollector) Collect(ctx *ScrapeContext, ch chan<- prometheus.Metric) error {
	if desc, err := c.collectRemoteFXNetworkCount(ctx, ch); err != nil {
		log.Error("failed collecting terminal services session count metrics:", desc, err)
		return err
	}
	if desc, err := c.collectRemoteFXGraphicsCounters(ctx, ch); err != nil {
		log.Error("failed collecting terminal services session count metrics:", desc, err)
		return err
	}
	return nil
}

type perflibRemoteFxNetwork struct {
	Name                     string
	BaseTCPRTT               float64 `perflib:"Base TCP RTT"`
	BaseUDPRTT               float64 `perflib:"Base UDP RTT"`
	CurrentTCPBandwidth      float64 `perflib:"Current TCP Bandwidth"`
	CurrentTCPRTT            float64 `perflib:"Current TCP RTT"`
	CurrentUDPBandwidth      float64 `perflib:"Current UDP Bandwidth"`
	CurrentUDPRTT            float64 `perflib:"Current UDP RTT"`
	TotalReceivedBytes       float64 `perflib:"Total Received Bytes"`
	TotalSentBytes           float64 `perflib:"Total Sent Bytes"`
	UDPPacketsReceivedPersec float64 `perflib:"UDP Packets Received/sec"`
	UDPPacketsSentPersec     float64 `perflib:"UDP Packets Sent/sec"`
}

func (c *RemoteFxCollector) collectRemoteFXNetworkCount(ctx *ScrapeContext, ch chan<- prometheus.Metric) (*prometheus.Desc, error) {
	dst := make([]perflibRemoteFxNetwork, 0)
	err := unmarshalObject(ctx.perfObjects["RemoteFX Network"], &dst)
	if err != nil {
		return nil, err
	}

	for _, d := range dst {
		// only connect metrics for remote named sessions
		n := strings.ToLower(d.Name)
		if n == "" || n == "services" || n == "console" {
			continue
		}
		ch <- prometheus.MustNewConstMetric(
			c.BaseTCPRTT,
			prometheus.GaugeValue,
			milliSecToSec(d.BaseTCPRTT),
			d.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.BaseUDPRTT,
			prometheus.GaugeValue,
			milliSecToSec(d.BaseUDPRTT),
			d.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.CurrentTCPBandwidth,
			prometheus.GaugeValue,
			(d.CurrentTCPBandwidth*1000)/8,
			d.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.CurrentTCPRTT,
			prometheus.GaugeValue,
			milliSecToSec(d.CurrentTCPRTT),
			d.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.CurrentUDPBandwidth,
			prometheus.GaugeValue,
			(d.CurrentUDPBandwidth*1000)/8,
			d.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.CurrentUDPRTT,
			prometheus.GaugeValue,
			milliSecToSec(d.CurrentUDPRTT),
			d.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.TotalReceivedBytes,
			prometheus.CounterValue,
			d.TotalReceivedBytes,
			d.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.TotalSentBytes,
			prometheus.CounterValue,
			d.TotalSentBytes,
			d.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.UDPPacketsReceivedPersec,
			prometheus.CounterValue,
			d.UDPPacketsReceivedPersec,
			d.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.UDPPacketsSentPersec,
			prometheus.CounterValue,
			d.UDPPacketsSentPersec,
			d.Name,
		)
	}
	return nil, nil
}

type perflibRemoteFxGraphics struct {
	Name                                               string
	AverageEncodingTime                                float64 `perflib:"Average Encoding Time"`
	FrameQuality                                       float64 `perflib:"Frame Quality"`
	FramesSkippedPerSecondInsufficientClientResources  float64 `perflib:"Frames Skipped/Second - Insufficient Server Resources"`
	FramesSkippedPerSecondInsufficientNetworkResources float64 `perflib:"Frames Skipped/Second - Insufficient Network Resources"`
	FramesSkippedPerSecondInsufficientServerResources  float64 `perflib:"Frames Skipped/Second - Insufficient Client Resources"`
	GraphicsCompressionratio                           float64 `perflib:"Graphics Compression ratio"`
	InputFramesPerSecond                               float64 `perflib:"Input Frames/Second"`
	OutputFramesPerSecond                              float64 `perflib:"Output Frames/Second"`
	SourceFramesPerSecond                              float64 `perflib:"Source Frames/Second"`
}

func (c *RemoteFxCollector) collectRemoteFXGraphicsCounters(ctx *ScrapeContext, ch chan<- prometheus.Metric) (*prometheus.Desc, error) {
	dst := make([]perflibRemoteFxGraphics, 0)
	err := unmarshalObject(ctx.perfObjects["RemoteFX Graphics"], &dst)
	if err != nil {
		return nil, err
	}

	for _, d := range dst {
		// only connect metrics for remote named sessions
		n := strings.ToLower(d.Name)
		if n == "" || n == "services" || n == "console" {
			continue
		}
		ch <- prometheus.MustNewConstMetric(
			c.AverageEncodingTime,
			prometheus.GaugeValue,
			milliSecToSec(d.AverageEncodingTime),
			d.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.FrameQuality,
			prometheus.GaugeValue,
			d.FrameQuality,
			d.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.FramesSkippedPerSecondInsufficientResources,
			prometheus.CounterValue,
			d.FramesSkippedPerSecondInsufficientClientResources,
			d.Name,
			"client",
		)
		ch <- prometheus.MustNewConstMetric(
			c.FramesSkippedPerSecondInsufficientResources,
			prometheus.CounterValue,
			d.FramesSkippedPerSecondInsufficientNetworkResources,
			d.Name,
			"network",
		)
		ch <- prometheus.MustNewConstMetric(
			c.FramesSkippedPerSecondInsufficientResources,
			prometheus.CounterValue,
			d.FramesSkippedPerSecondInsufficientServerResources,
			d.Name,
			"server",
		)
		ch <- prometheus.MustNewConstMetric(
			c.GraphicsCompressionratio,
			prometheus.GaugeValue,
			d.GraphicsCompressionratio,
			d.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.InputFramesPerSecond,
			prometheus.CounterValue,
			d.InputFramesPerSecond,
			d.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.OutputFramesPerSecond,
			prometheus.CounterValue,
			d.OutputFramesPerSecond,
			d.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.SourceFramesPerSecond,
			prometheus.CounterValue,
			d.SourceFramesPerSecond,
			d.Name,
		)
	}

	return nil, nil
}
