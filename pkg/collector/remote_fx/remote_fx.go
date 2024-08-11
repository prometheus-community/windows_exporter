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

// Collector
// A RemoteFxNetworkCollector is a Prometheus Collector for
// WMI Win32_PerfRawData_Counters_RemoteFXNetwork & Win32_PerfRawData_Counters_RemoteFXGraphics metrics
// https://wutils.com/wmi/root/cimv2/win32_perfrawdata_counters_remotefxnetwork/
// https://wutils.com/wmi/root/cimv2/win32_perfrawdata_counters_remotefxgraphics/
type Collector struct {
	config Config
	logger log.Logger

	// net
	baseTCPRTT               *prometheus.Desc
	baseUDPRTT               *prometheus.Desc
	currentTCPBandwidth      *prometheus.Desc
	currentTCPRTT            *prometheus.Desc
	currentUDPBandwidth      *prometheus.Desc
	currentUDPRTT            *prometheus.Desc
	fecRate                  *prometheus.Desc
	lossRate                 *prometheus.Desc
	retransmissionRate       *prometheus.Desc
	totalReceivedBytes       *prometheus.Desc
	totalSentBytes           *prometheus.Desc
	udpPacketsReceivedPerSec *prometheus.Desc
	udpPacketsSentPerSec     *prometheus.Desc

	// gfx
	averageEncodingTime                         *prometheus.Desc
	frameQuality                                *prometheus.Desc
	framesSkippedPerSecondInsufficientResources *prometheus.Desc
	graphicsCompressionRatio                    *prometheus.Desc
	inputFramesPerSecond                        *prometheus.Desc
	outputFramesPerSecond                       *prometheus.Desc
	sourceFramesPerSecond                       *prometheus.Desc
}

func New(logger log.Logger, config *Config) *Collector {
	if config == nil {
		config = &ConfigDefaults
	}

	c := &Collector{
		config: *config,
	}

	c.SetLogger(logger)

	return c
}

func NewWithFlags(_ *kingpin.Application) *Collector {
	return &Collector{}
}

func (c *Collector) GetName() string {
	return Name
}

func (c *Collector) SetLogger(logger log.Logger) {
	c.logger = log.With(logger, "collector", Name)
}

func (c *Collector) GetPerfCounter() ([]string, error) {
	return []string{"RemoteFX Network", "RemoteFX Graphics"}, nil
}

func (c *Collector) Close() error {
	return nil
}

func (c *Collector) Build() error {
	// net
	c.baseTCPRTT = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "net_base_tcp_rtt_seconds"),
		"Base TCP round-trip time (RTT) detected in seconds",
		[]string{"session_name"},
		nil,
	)
	c.baseUDPRTT = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "net_base_udp_rtt_seconds"),
		"Base UDP round-trip time (RTT) detected in seconds.",
		[]string{"session_name"},
		nil,
	)
	c.currentTCPBandwidth = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "net_current_tcp_bandwidth"),
		"TCP Bandwidth detected in bytes per second.",
		[]string{"session_name"},
		nil,
	)
	c.currentTCPRTT = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "net_current_tcp_rtt_seconds"),
		"Average TCP round-trip time (RTT) detected in seconds.",
		[]string{"session_name"},
		nil,
	)
	c.currentUDPBandwidth = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "net_current_udp_bandwidth"),
		"UDP Bandwidth detected in bytes per second.",
		[]string{"session_name"},
		nil,
	)
	c.currentUDPRTT = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "net_current_udp_rtt_seconds"),
		"Average UDP round-trip time (RTT) detected in seconds.",
		[]string{"session_name"},
		nil,
	)
	c.totalReceivedBytes = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "net_received_bytes_total"),
		"(TotalReceivedBytes)",
		[]string{"session_name"},
		nil,
	)
	c.totalSentBytes = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "net_sent_bytes_total"),
		"(TotalSentBytes)",
		[]string{"session_name"},
		nil,
	)
	c.udpPacketsReceivedPerSec = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "net_udp_packets_received_total"),
		"Rate in packets per second at which packets are received over UDP.",
		[]string{"session_name"},
		nil,
	)
	c.udpPacketsSentPerSec = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "net_udp_packets_sent_total"),
		"Rate in packets per second at which packets are sent over UDP.",
		[]string{"session_name"},
		nil,
	)
	c.fecRate = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "net_fec_rate"),
		"Forward Error Correction (FEC) percentage",
		[]string{"session_name"},
		nil,
	)
	c.lossRate = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "net_loss_rate"),
		"Loss percentage",
		[]string{"session_name"},
		nil,
	)
	c.retransmissionRate = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "net_retransmission_rate"),
		"Percentage of packets that have been retransmitted",
		[]string{"session_name"},
		nil,
	)

	// gfx
	c.averageEncodingTime = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "gfx_average_encoding_time_seconds"),
		"Average frame encoding time in seconds",
		[]string{"session_name"},
		nil,
	)
	c.frameQuality = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "gfx_frame_quality"),
		"Quality of the output frame expressed as a percentage of the quality of the source frame.",
		[]string{"session_name"},
		nil,
	)
	c.framesSkippedPerSecondInsufficientResources = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "gfx_frames_skipped_insufficient_resource_total"),
		"Number of frames skipped per second due to insufficient client resources.",
		[]string{"session_name", "resource"},
		nil,
	)
	c.graphicsCompressionRatio = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "gfx_graphics_compression_ratio"),
		"Ratio of the number of bytes encoded to the number of bytes input.",
		[]string{"session_name"},
		nil,
	)
	c.inputFramesPerSecond = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "gfx_input_frames_total"),
		"Number of sources frames provided as input to RemoteFX graphics per second.",
		[]string{"session_name"},
		nil,
	)
	c.outputFramesPerSecond = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "gfx_output_frames_total"),
		"Number of frames sent to the client per second.",
		[]string{"session_name"},
		nil,
	)
	c.sourceFramesPerSecond = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "gfx_source_frames_total"),
		"Number of frames composed by the source (DWM) per second.",
		[]string{"session_name"},
		nil,
	)
	return nil
}

// Collect sends the metric values for each metric
// to the provided prometheus Metric channel.
func (c *Collector) Collect(ctx *types.ScrapeContext, ch chan<- prometheus.Metric) error {
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

func (c *Collector) collectRemoteFXNetworkCount(ctx *types.ScrapeContext, ch chan<- prometheus.Metric) error {
	dst := make([]perflibRemoteFxNetwork, 0)
	err := perflib.UnmarshalObject(ctx.PerfObjects["RemoteFX Network"], &dst, c.logger)
	if err != nil {
		return err
	}

	for _, d := range dst {
		// only connect metrics for remote named sessions
		n := strings.ToLower(normalizeSessionName(d.Name))
		if n == "" || n == "services" || n == "console" {
			continue
		}
		ch <- prometheus.MustNewConstMetric(
			c.baseTCPRTT,
			prometheus.GaugeValue,
			utils.MilliSecToSec(d.BaseTCPRTT),
			normalizeSessionName(d.Name),
		)
		ch <- prometheus.MustNewConstMetric(
			c.baseUDPRTT,
			prometheus.GaugeValue,
			utils.MilliSecToSec(d.BaseUDPRTT),
			normalizeSessionName(d.Name),
		)
		ch <- prometheus.MustNewConstMetric(
			c.currentTCPBandwidth,
			prometheus.GaugeValue,
			(d.CurrentTCPBandwidth*1000)/8,
			normalizeSessionName(d.Name),
		)
		ch <- prometheus.MustNewConstMetric(
			c.currentTCPRTT,
			prometheus.GaugeValue,
			utils.MilliSecToSec(d.CurrentTCPRTT),
			normalizeSessionName(d.Name),
		)
		ch <- prometheus.MustNewConstMetric(
			c.currentUDPBandwidth,
			prometheus.GaugeValue,
			(d.CurrentUDPBandwidth*1000)/8,
			normalizeSessionName(d.Name),
		)
		ch <- prometheus.MustNewConstMetric(
			c.currentUDPRTT,
			prometheus.GaugeValue,
			utils.MilliSecToSec(d.CurrentUDPRTT),
			normalizeSessionName(d.Name),
		)
		ch <- prometheus.MustNewConstMetric(
			c.totalReceivedBytes,
			prometheus.CounterValue,
			d.TotalReceivedBytes,
			normalizeSessionName(d.Name),
		)
		ch <- prometheus.MustNewConstMetric(
			c.totalSentBytes,
			prometheus.CounterValue,
			d.TotalSentBytes,
			normalizeSessionName(d.Name),
		)
		ch <- prometheus.MustNewConstMetric(
			c.udpPacketsReceivedPerSec,
			prometheus.CounterValue,
			d.UDPPacketsReceivedPersec,
			normalizeSessionName(d.Name),
		)
		ch <- prometheus.MustNewConstMetric(
			c.udpPacketsSentPerSec,
			prometheus.CounterValue,
			d.UDPPacketsSentPersec,
			normalizeSessionName(d.Name),
		)
		ch <- prometheus.MustNewConstMetric(
			c.fecRate,
			prometheus.GaugeValue,
			d.FECRate,
			normalizeSessionName(d.Name),
		)

		ch <- prometheus.MustNewConstMetric(
			c.lossRate,
			prometheus.GaugeValue,
			d.LossRate,
			normalizeSessionName(d.Name),
		)

		ch <- prometheus.MustNewConstMetric(
			c.retransmissionRate,
			prometheus.GaugeValue,
			d.RetransmissionRate,
			normalizeSessionName(d.Name),
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

func (c *Collector) collectRemoteFXGraphicsCounters(ctx *types.ScrapeContext, ch chan<- prometheus.Metric) error {
	dst := make([]perflibRemoteFxGraphics, 0)
	err := perflib.UnmarshalObject(ctx.PerfObjects["RemoteFX Graphics"], &dst, c.logger)
	if err != nil {
		return err
	}

	for _, d := range dst {
		// only connect metrics for remote named sessions
		n := strings.ToLower(normalizeSessionName(d.Name))
		if n == "" || n == "services" || n == "console" {
			continue
		}
		ch <- prometheus.MustNewConstMetric(
			c.averageEncodingTime,
			prometheus.GaugeValue,
			utils.MilliSecToSec(d.AverageEncodingTime),
			normalizeSessionName(d.Name),
		)
		ch <- prometheus.MustNewConstMetric(
			c.frameQuality,
			prometheus.GaugeValue,
			d.FrameQuality,
			normalizeSessionName(d.Name),
		)
		ch <- prometheus.MustNewConstMetric(
			c.framesSkippedPerSecondInsufficientResources,
			prometheus.CounterValue,
			d.FramesSkippedPerSecondInsufficientClientResources,
			normalizeSessionName(d.Name),
			"client",
		)
		ch <- prometheus.MustNewConstMetric(
			c.framesSkippedPerSecondInsufficientResources,
			prometheus.CounterValue,
			d.FramesSkippedPerSecondInsufficientNetworkResources,
			normalizeSessionName(d.Name),
			"network",
		)
		ch <- prometheus.MustNewConstMetric(
			c.framesSkippedPerSecondInsufficientResources,
			prometheus.CounterValue,
			d.FramesSkippedPerSecondInsufficientServerResources,
			normalizeSessionName(d.Name),
			"server",
		)
		ch <- prometheus.MustNewConstMetric(
			c.graphicsCompressionRatio,
			prometheus.GaugeValue,
			d.GraphicsCompressionratio,
			normalizeSessionName(d.Name),
		)
		ch <- prometheus.MustNewConstMetric(
			c.inputFramesPerSecond,
			prometheus.CounterValue,
			d.InputFramesPerSecond,
			normalizeSessionName(d.Name),
		)
		ch <- prometheus.MustNewConstMetric(
			c.outputFramesPerSecond,
			prometheus.CounterValue,
			d.OutputFramesPerSecond,
			normalizeSessionName(d.Name),
		)
		ch <- prometheus.MustNewConstMetric(
			c.sourceFramesPerSecond,
			prometheus.CounterValue,
			d.SourceFramesPerSecond,
			normalizeSessionName(d.Name),
		)
	}

	return nil
}

// normalizeSessionName ensure that the session is the same between WTS API and performance counters.
func normalizeSessionName(sessionName string) string {
	return strings.Replace(sessionName, "RDP-tcp", "RDP-Tcp", 1)
}
