//go:build windows

package remote_fx

import (
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/alecthomas/kingpin/v2"
	"github.com/prometheus-community/windows_exporter/internal/mi"
	"github.com/prometheus-community/windows_exporter/internal/perfdata"
	"github.com/prometheus-community/windows_exporter/internal/types"
	"github.com/prometheus-community/windows_exporter/internal/utils"
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

	perfDataCollectorNetwork  *perfdata.Collector
	perfDataCollectorGraphics *perfdata.Collector

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

func New(config *Config) *Collector {
	if config == nil {
		config = &ConfigDefaults
	}

	c := &Collector{
		config: *config,
	}

	return c
}

func NewWithFlags(_ *kingpin.Application) *Collector {
	return &Collector{}
}

func (c *Collector) GetName() string {
	return Name
}

func (c *Collector) Close() error {
	c.perfDataCollectorNetwork.Close()
	c.perfDataCollectorGraphics.Close()

	return nil
}

func (c *Collector) Build(*slog.Logger, *mi.Session) error {
	var err error

	c.perfDataCollectorNetwork, err = perfdata.NewCollector("RemoteFX Network", perfdata.InstanceAll, []string{
		BaseTCPRTT,
		BaseUDPRTT,
		CurrentTCPBandwidth,
		CurrentTCPRTT,
		CurrentUDPBandwidth,
		CurrentUDPRTT,
		TotalReceivedBytes,
		TotalSentBytes,
		UDPPacketsReceivedPersec,
		UDPPacketsSentPersec,
		FECRate,
		LossRate,
		RetransmissionRate,
	})
	if err != nil {
		return fmt.Errorf("failed to create RemoteFX Network collector: %w", err)
	}

	c.perfDataCollectorGraphics, err = perfdata.NewCollector("RemoteFX Graphics", perfdata.InstanceAll, []string{
		AverageEncodingTime,
		FrameQuality,
		FramesSkippedPerSecondInsufficientClientResources,
		FramesSkippedPerSecondInsufficientNetworkResources,
		FramesSkippedPerSecondInsufficientServerResources,
		GraphicsCompressionratio,
		InputFramesPerSecond,
		OutputFramesPerSecond,
		SourceFramesPerSecond,
	})
	if err != nil {
		return fmt.Errorf("failed to create RemoteFX Graphics collector: %w", err)
	}

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
func (c *Collector) Collect(ch chan<- prometheus.Metric) error {
	errs := make([]error, 0, 2)

	if err := c.collectRemoteFXNetworkCount(ch); err != nil {
		errs = append(errs, fmt.Errorf("failed collecting RemoteFX Network metrics: %w", err))
	}

	if err := c.collectRemoteFXGraphicsCounters(ch); err != nil {
		errs = append(errs, fmt.Errorf("failed collecting RemoteFX Graphics metrics: %w", err))
	}

	return errors.Join(errs...)
}

func (c *Collector) collectRemoteFXNetworkCount(ch chan<- prometheus.Metric) error {
	perfData, err := c.perfDataCollectorNetwork.Collect()
	if err != nil {
		return fmt.Errorf("failed to collect RemoteFX Network metrics: %w", err)
	}

	for name, data := range perfData {
		// only connect metrics for remote named sessions
		sessionName := normalizeSessionName(name)
		if n := strings.ToLower(sessionName); n == "" || n == "services" || n == "console" {
			continue
		}

		ch <- prometheus.MustNewConstMetric(
			c.baseTCPRTT,
			prometheus.GaugeValue,
			utils.MilliSecToSec(data[BaseTCPRTT].FirstValue),
			sessionName,
		)
		ch <- prometheus.MustNewConstMetric(
			c.baseUDPRTT,
			prometheus.GaugeValue,
			utils.MilliSecToSec(data[BaseUDPRTT].FirstValue),
			sessionName,
		)
		ch <- prometheus.MustNewConstMetric(
			c.currentTCPBandwidth,
			prometheus.GaugeValue,
			(data[CurrentTCPBandwidth].FirstValue*1000)/8,
			sessionName,
		)
		ch <- prometheus.MustNewConstMetric(
			c.currentTCPRTT,
			prometheus.GaugeValue,
			utils.MilliSecToSec(data[CurrentTCPRTT].FirstValue),
			sessionName,
		)
		ch <- prometheus.MustNewConstMetric(
			c.currentUDPBandwidth,
			prometheus.GaugeValue,
			(data[CurrentUDPBandwidth].FirstValue*1000)/8,
			sessionName,
		)
		ch <- prometheus.MustNewConstMetric(
			c.currentUDPRTT,
			prometheus.GaugeValue,
			utils.MilliSecToSec(data[CurrentUDPRTT].FirstValue),
			sessionName,
		)
		ch <- prometheus.MustNewConstMetric(
			c.totalReceivedBytes,
			prometheus.CounterValue,
			data[TotalReceivedBytes].FirstValue,
			sessionName,
		)
		ch <- prometheus.MustNewConstMetric(
			c.totalSentBytes,
			prometheus.CounterValue,
			data[TotalSentBytes].FirstValue,
			sessionName,
		)
		ch <- prometheus.MustNewConstMetric(
			c.udpPacketsReceivedPerSec,
			prometheus.CounterValue,
			data[UDPPacketsReceivedPersec].FirstValue,
			sessionName,
		)
		ch <- prometheus.MustNewConstMetric(
			c.udpPacketsSentPerSec,
			prometheus.CounterValue,
			data[UDPPacketsSentPersec].FirstValue,
			sessionName,
		)
		ch <- prometheus.MustNewConstMetric(
			c.fecRate,
			prometheus.GaugeValue,
			data[FECRate].FirstValue,
			sessionName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.lossRate,
			prometheus.GaugeValue,
			data[LossRate].FirstValue,
			sessionName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.retransmissionRate,
			prometheus.GaugeValue,
			data[RetransmissionRate].FirstValue,
			sessionName,
		)
	}

	return nil
}

func (c *Collector) collectRemoteFXGraphicsCounters(ch chan<- prometheus.Metric) error {
	perfData, err := c.perfDataCollectorNetwork.Collect()
	if err != nil {
		return fmt.Errorf("failed to collect RemoteFX Graphics metrics: %w", err)
	}

	for name, data := range perfData {
		// only connect metrics for remote named sessions
		sessionName := normalizeSessionName(name)
		if n := strings.ToLower(sessionName); n == "" || n == "services" || n == "console" {
			continue
		}

		ch <- prometheus.MustNewConstMetric(
			c.averageEncodingTime,
			prometheus.GaugeValue,
			utils.MilliSecToSec(data[AverageEncodingTime].FirstValue),
			sessionName,
		)
		ch <- prometheus.MustNewConstMetric(
			c.frameQuality,
			prometheus.GaugeValue,
			data[FrameQuality].FirstValue,
			sessionName,
		)
		ch <- prometheus.MustNewConstMetric(
			c.framesSkippedPerSecondInsufficientResources,
			prometheus.CounterValue,
			data[FramesSkippedPerSecondInsufficientClientResources].FirstValue,
			sessionName,
			"client",
		)
		ch <- prometheus.MustNewConstMetric(
			c.framesSkippedPerSecondInsufficientResources,
			prometheus.CounterValue,
			data[FramesSkippedPerSecondInsufficientNetworkResources].FirstValue,
			sessionName,
			"network",
		)
		ch <- prometheus.MustNewConstMetric(
			c.framesSkippedPerSecondInsufficientResources,
			prometheus.CounterValue,
			data[FramesSkippedPerSecondInsufficientServerResources].FirstValue,
			sessionName,
			"server",
		)
		ch <- prometheus.MustNewConstMetric(
			c.graphicsCompressionRatio,
			prometheus.GaugeValue,
			data[GraphicsCompressionratio].FirstValue,
			sessionName,
		)
		ch <- prometheus.MustNewConstMetric(
			c.inputFramesPerSecond,
			prometheus.CounterValue,
			data[InputFramesPerSecond].FirstValue,
			sessionName,
		)
		ch <- prometheus.MustNewConstMetric(
			c.outputFramesPerSecond,
			prometheus.CounterValue,
			data[OutputFramesPerSecond].FirstValue,
			sessionName,
		)
		ch <- prometheus.MustNewConstMetric(
			c.sourceFramesPerSecond,
			prometheus.CounterValue,
			data[SourceFramesPerSecond].FirstValue,
			sessionName,
		)
	}

	return nil
}

// normalizeSessionName ensure that the session is the same between WTS API and performance counters.
func normalizeSessionName(sessionName string) string {
	return strings.Replace(sessionName, "RDP-tcp", "RDP-Tcp", 1)
}
