//go:build windows

package remote_fx

import (
	"strings"

	"github.com/alecthomas/kingpin/v2"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/prometheus-community/windows_exporter/pkg/perflib"
	"github.com/prometheus-community/windows_exporter/pkg/types"
	"github.com/prometheus-community/windows_exporter/pkg/utils"
	"github.com/prometheus/client_golang/prometheus"
)

const Name = "remote_fx"

type Config struct{}

var ConfigDefaults = Config{}

// collector
// A RemoteFxNetworkCollector is a Prometheus collector for
// WMI Win32_PerfRawData_Counters_RemoteFXNetwork & Win32_PerfRawData_Counters_RemoteFXGraphics metrics
// https://wutils.com/wmi/root/cimv2/win32_perfrawdata_counters_remotefxnetwork/
// https://wutils.com/wmi/root/cimv2/win32_perfrawdata_counters_remotefxgraphics/
type collector struct {
	logger log.Logger

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
	FECRate                  *prometheus.Desc
	LossRate                 *prometheus.Desc
	RetransmissionRate       *prometheus.Desc

	// gfx
	AverageEncodingTime                         *prometheus.Desc
	FrameQuality                                *prometheus.Desc
	FramesSkippedPerSecondInsufficientResources *prometheus.Desc
	GraphicsCompressionratio                    *prometheus.Desc
	InputFramesPerSecond                        *prometheus.Desc
	OutputFramesPerSecond                       *prometheus.Desc
	SourceFramesPerSecond                       *prometheus.Desc
}

func New(logger log.Logger, _ *Config) types.Collector {
	c := &collector{}
	c.SetLogger(logger)
	return c
}

func NewWithFlags(_ *kingpin.Application) types.Collector {
	return &collector{}
}

func (c *collector) GetName() string {
	return Name
}

func (c *collector) SetLogger(logger log.Logger) {
	c.logger = log.With(logger, "collector", Name)
}

func (c *collector) GetPerfCounter() ([]string, error) {
	return []string{"RemoteFX Network", "RemoteFX Graphics"}, nil
}

func (c *collector) Build() error {
	// net
	c.BaseTCPRTT = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "net_base_tcp_rtt_seconds"),
		"Base TCP round-trip time (RTT) detected in seconds",
		[]string{"session_name"},
		nil,
	)
	c.BaseUDPRTT = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "net_base_udp_rtt_seconds"),
		"Base UDP round-trip time (RTT) detected in seconds.",
		[]string{"session_name"},
		nil,
	)
	c.CurrentTCPBandwidth = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "net_current_tcp_bandwidth"),
		"TCP Bandwidth detected in bytes per second.",
		[]string{"session_name"},
		nil,
	)
	c.CurrentTCPRTT = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "net_current_tcp_rtt_seconds"),
		"Average TCP round-trip time (RTT) detected in seconds.",
		[]string{"session_name"},
		nil,
	)
	c.CurrentUDPBandwidth = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "net_current_udp_bandwidth"),
		"UDP Bandwidth detected in bytes per second.",
		[]string{"session_name"},
		nil,
	)
	c.CurrentUDPRTT = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "net_current_udp_rtt_seconds"),
		"Average UDP round-trip time (RTT) detected in seconds.",
		[]string{"session_name"},
		nil,
	)
	c.TotalReceivedBytes = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "net_received_bytes_total"),
		"(TotalReceivedBytes)",
		[]string{"session_name"},
		nil,
	)
	c.TotalSentBytes = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "net_sent_bytes_total"),
		"(TotalSentBytes)",
		[]string{"session_name"},
		nil,
	)
	c.UDPPacketsReceivedPersec = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "net_udp_packets_received_total"),
		"Rate in packets per second at which packets are received over UDP.",
		[]string{"session_name"},
		nil,
	)
	c.UDPPacketsSentPersec = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "net_udp_packets_sent_total"),
		"Rate in packets per second at which packets are sent over UDP.",
		[]string{"session_name"},
		nil,
	)
	c.FECRate = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "net_fec_rate"),
		"Forward Error Correction (FEC) percentage",
		[]string{"session_name"},
		nil,
	)
	c.LossRate = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "net_loss_rate"),
		"Loss percentage",
		[]string{"session_name"},
		nil,
	)
	c.RetransmissionRate = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "net_retransmission_rate"),
		"Percentage of packets that have been retransmitted",
		[]string{"session_name"},
		nil,
	)

	// gfx
	c.AverageEncodingTime = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "gfx_average_encoding_time_seconds"),
		"Average frame encoding time in seconds",
		[]string{"session_name"},
		nil,
	)
	c.FrameQuality = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "gfx_frame_quality"),
		"Quality of the output frame expressed as a percentage of the quality of the source frame.",
		[]string{"session_name"},
		nil,
	)
	c.FramesSkippedPerSecondInsufficientResources = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "gfx_frames_skipped_insufficient_resource_total"),
		"Number of frames skipped per second due to insufficient client resources.",
		[]string{"session_name", "resource"},
		nil,
	)
	c.GraphicsCompressionratio = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "gfx_graphics_compression_ratio"),
		"Ratio of the number of bytes encoded to the number of bytes input.",
		[]string{"session_name"},
		nil,
	)
	c.InputFramesPerSecond = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "gfx_input_frames_total"),
		"Number of sources frames provided as input to RemoteFX graphics per second.",
		[]string{"session_name"},
		nil,
	)
	c.OutputFramesPerSecond = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "gfx_output_frames_total"),
		"Number of frames sent to the client per second.",
		[]string{"session_name"},
		nil,
	)
	c.SourceFramesPerSecond = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "gfx_source_frames_total"),
		"Number of frames composed by the source (DWM) per second.",
		[]string{"session_name"},
		nil,
	)
	return nil
}

// Collect sends the metric values for each metric
// to the provided prometheus Metric channel.
func (c *collector) Collect(ctx *types.ScrapeContext, ch chan<- prometheus.Metric) error {
	if err := c.collectRemoteFXNetworkCount(ctx, ch); err != nil {
		_ = level.Error(c.logger).Log("msg", "failed collecting terminal services session count metrics", "err", err)
		return err
	}
	if err := c.collectRemoteFXGraphicsCounters(ctx, ch); err != nil {
		_ = level.Error(c.logger).Log("msg", "failed collecting terminal services session count metrics", "err", err)
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
	FECRate                  float64 `perflib:"Forward Error Correction (FEC) percentage"`
	LossRate                 float64 `perflib:"Loss percentage"`
	RetransmissionRate       float64 `perflib:"Percentage of packets that have been retransmitted"`
}

func (c *collector) collectRemoteFXNetworkCount(ctx *types.ScrapeContext, ch chan<- prometheus.Metric) error {
	dst := make([]perflibRemoteFxNetwork, 0)
	err := perflib.UnmarshalObject(ctx.PerfObjects["RemoteFX Network"], &dst, c.logger)
	if err != nil {
		return err
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
			utils.MilliSecToSec(d.BaseTCPRTT),
			d.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.BaseUDPRTT,
			prometheus.GaugeValue,
			utils.MilliSecToSec(d.BaseUDPRTT),
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
			utils.MilliSecToSec(d.CurrentTCPRTT),
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
			utils.MilliSecToSec(d.CurrentUDPRTT),
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
		ch <- prometheus.MustNewConstMetric(
			c.FECRate,
			prometheus.GaugeValue,
			d.FECRate,
			d.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.LossRate,
			prometheus.GaugeValue,
			d.LossRate,
			d.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.RetransmissionRate,
			prometheus.GaugeValue,
			d.RetransmissionRate,
			d.Name,
		)
	}
	return nil
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

func (c *collector) collectRemoteFXGraphicsCounters(ctx *types.ScrapeContext, ch chan<- prometheus.Metric) error {
	dst := make([]perflibRemoteFxGraphics, 0)
	err := perflib.UnmarshalObject(ctx.PerfObjects["RemoteFX Graphics"], &dst, c.logger)
	if err != nil {
		return err
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
			utils.MilliSecToSec(d.AverageEncodingTime),
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

	return nil
}
