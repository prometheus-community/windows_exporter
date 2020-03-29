// +build windows

package collector

import (
	"github.com/prometheus/client_golang/prometheus"
)

func init() {
	registerCollector("remote_fx", NewRemoteFx)
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
	FECRate                  *prometheus.Desc
	FECRate_Base             *prometheus.Desc
	LossRate                 *prometheus.Desc
	LossRate_Base            *prometheus.Desc
	RetransmissionRate       *prometheus.Desc
	RetransmissionRate_Base  *prometheus.Desc
	TCPReceivedRate          *prometheus.Desc
	TCPSentRate              *prometheus.Desc
	TotalReceivedRate        *prometheus.Desc
	TotalSentRate            *prometheus.Desc
	UDPPacketsReceivedPersec *prometheus.Desc
	UDPPacketsSentPersec     *prometheus.Desc
	UDPReceivedRate          *prometheus.Desc
	UDPSentRate              *prometheus.Desc

	//gfx
	AverageEncodingTime                                *prometheus.Desc
	FrameQuality                                       *prometheus.Desc
	FramesSkippedPerSecondInsufficientClientResources  *prometheus.Desc
	FramesSkippedPerSecondInsufficientNetworkResources *prometheus.Desc
	FramesSkippedPerSecondInsufficientServerResources  *prometheus.Desc
	GraphicsCompressionratio                           *prometheus.Desc
	InputFramesPerSecond                               *prometheus.Desc
	OutputFramesPerSecond                              *prometheus.Desc
	SourceFramesPerSecond                              *prometheus.Desc
}

// NewRemoteFx ...
func NewRemoteFx() (Collector, error) {
	const subsystem = "remote_fx"
	return &RemoteFxCollector{
		// net
		BaseTCPRTT: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "net_base_tcp_rrt"),
			"Base TCP round-trip time (RTT) detected in milliseconds",
			[]string{"session"},
			nil,
		),
		BaseUDPRTT: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "net_base_udp_rrt"),
			"Base UDP round-trip time (RTT) detected in milliseconds.",
			[]string{"session"},
			nil,
		),
		CurrentTCPBandwidth: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "net_current_tcp_bandwidth"),
			"TCP Bandwidth detected in thousands of bits per second (1000 bps).",
			[]string{"session"},
			nil,
		),
		CurrentTCPRTT: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "net_current_tcp_rtt"),
			"Average TCP round-trip time (RTT) detected in milliseconds.",
			[]string{"session"},
			nil,
		),
		CurrentUDPBandwidth: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "net_current_udp_bandwidth"),
			"UDP Bandwidth detected in thousands of bits per second (1000 bps).",
			[]string{"session"},
			nil,
		),
		CurrentUDPRTT: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "net_current_udp_rtt"),
			"Average UDP round-trip time (RTT) detected in milliseconds.",
			[]string{"session"},
			nil,
		),
		FECRate: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "net_fec_rate"),
			"Forward Error Correction (FEC) percentage",
			[]string{"session"},
			nil,
		),
		FECRate_Base: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "net_fec_rate_base"),
			"Forward Error Correction (FEC) percentage _Base value",
			[]string{"session"},
			nil,
		),
		LossRate: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "net_loss_rate"),
			"Loss percentage",
			[]string{"session"},
			nil,
		),
		LossRate_Base: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "net_loss_rate_base"),
			"Loss percentage _Base value.",
			[]string{"session"},
			nil,
		),
		RetransmissionRate: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "net_retransmission_rate"),
			"Percentage of packets that have been retransmitted",
			[]string{"session"},
			nil,
		),
		RetransmissionRate_Base: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "net_retransmission_rate_Base"),
			"Percentage of packets that have been retransmitted _base value",
			[]string{"session"},
			nil,
		),
		TCPReceivedRate: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "net_tcp_received_rate"),
			"Rate in bits per second (bps) at which data is received over TCP.",
			[]string{"session"},
			nil,
		),
		TCPSentRate: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "net_tcp_sent_rate"),
			"Rate in bits per second (bps) at which data is sent over TCP.",
			[]string{"session"},
			nil,
		),
		TotalReceivedRate: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "net_total_received_rate"),
			"Rate in bits per second (bps) at which data is received.",
			[]string{"session"},
			nil,
		),
		TotalSentRate: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "net_total_sent_rate"),
			"Rate in bits per second (bps) at which data is sent.",
			[]string{"session"},
			nil,
		),
		UDPPacketsReceivedPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "net_udp_packets_received_persec"),
			"Rate in packets per second at which packets are received over UDP.",
			[]string{"session"},
			nil,
		),
		UDPPacketsSentPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "net_udp_packets_sent_persec"),
			"Rate in packets per second at which packets are sent over UDP.",
			[]string{"session"},
			nil,
		),
		UDPReceivedRate: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "net_udp_received_rate"),
			"Rate in bits per second (bps) at which data is received over UDP.",
			[]string{"session"},
			nil,
		),
		UDPSentRate: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "net_udp_sent_rate"),
			"Rate in bits per second (bps) at which data is sent over UDP.",
			[]string{"session"},
			nil,
		),

		//gfx
		AverageEncodingTime: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "gfx_average_encoding_time"),
			"Average frame encoding time.",
			[]string{"session"},
			nil,
		),
		FrameQuality: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "gfx_frame_quality"),
			"Quality of the output frame expressed as a percentage of the quality of the source frame.",
			[]string{"session"},
			nil,
		),
		FramesSkippedPerSecondInsufficientClientResources: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "gfx_frames_skipped_persec_insufficient_clt_res"),
			"Number of frames skipped per second due to insufficient client resources.",
			[]string{"session"},
			nil,
		),
		FramesSkippedPerSecondInsufficientNetworkResources: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "gfx_frames_skipped_persec_insufficient_net_res"),
			"Number of frames skipped per second due to insufficient network resources.",
			[]string{"session"},
			nil,
		),
		FramesSkippedPerSecondInsufficientServerResources: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "gfx_frames_skipped_persec_insufficient_srv_res"),
			"Number of frames skipped per second due to insufficient server resources.",
			[]string{"session"},
			nil,
		),
		GraphicsCompressionratio: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "gfx_graphics_compression_ratio"),
			"Ratio of the number of bytes encoded to the number of bytes input.",
			[]string{"session"},
			nil,
		),
		InputFramesPerSecond: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "gfx_input_frames_persec"),
			"Number of sources frames provided as input to RemoteFX graphics per second.",
			[]string{"session"},
			nil,
		),
		OutputFramesPerSecond: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "gfx_output_frames_persec"),
			"Number of frames sent to the client per second.",
			[]string{"session"},
			nil,
		),
		SourceFramesPerSecond: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "gfx_source_frames_persec"),
			"Number of frames composed by the source (DWM) per second.",
			[]string{"session"},
			nil,
		),
	}, nil
}

// Collect sends the metric values for each metric
// to the provided prometheus Metric channel.
func (c *RemoteFxCollector) Collect(ctx *ScrapeContext, ch chan<- prometheus.Metric) error {
	if desc, err := c.collectRemoteFXNetworkCount(ch); err != nil {
		log.Error("failed collecting terminal services session count metrics:", desc, err)
		return err
	}
	if desc, err := c.collectRemoteFXGraphicsCounters(ch); err != nil {
		log.Error("failed collecting terminal services session count metrics:", desc, err)
		return err
	}
	return nil
}

type Win32_PerfRawData_Counters_RemoteFXNetwork struct {
	Name                     string
	BaseTCPRTT               uint32
	BaseUDPRTT               uint32
	CurrentTCPBandwidth      uint32
	CurrentTCPRTT            uint32
	CurrentUDPBandwidth      uint32
	CurrentUDPRTT            uint32
	FECRate                  uint32
	FECRate_Base             uint32
	LossRate                 uint32
	LossRate_Base            uint32
	RetransmissionRate       uint32
	RetransmissionRate_Base  uint32
	TCPReceivedRate          uint32
	TCPSentRate              uint32
	TotalReceivedRate        uint32
	TotalSentRate            uint32
	UDPPacketsReceivedPersec uint32
	UDPPacketsSentPersec     uint32
	UDPReceivedRate          uint32
	UDPSentRate              uint32
}

func (c *RemoteFxCollector) collectRemoteFXNetworkCount(ch chan<- prometheus.Metric) (*prometheus.Desc, error) {
	var dst []Win32_PerfRawData_Counters_RemoteFXNetwork
	q := queryAll(&dst)
	if err := wmi.Query(q, &dst); err != nil {
		return nil, err
	}
	if len(dst) == 0 {
		return nil, errors.New("WMI query returned empty result set")
	}

	for _, d := range dst {
		ch <- prometheus.MustNewConstMetric(
			c.BaseTCPRTT,
			prometheus.GaugeValue,
			float64(d.BaseTCPRTT),
			d.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.BaseUDPRTT,
			prometheus.GaugeValue,
			float64(d.BaseUDPRTT),
			d.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.CurrentTCPBandwidth,
			prometheus.GaugeValue,
			float64(d.CurrentTCPBandwidth),
			d.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.CurrentTCPRTT,
			prometheus.GaugeValue,
			float64(d.CurrentTCPRTT),
			d.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.CurrentUDPBandwidth,
			prometheus.GaugeValue,
			float64(d.CurrentUDPBandwidth),
			d.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.CurrentUDPRTT,
			prometheus.GaugeValue,
			float64(d.CurrentUDPRTT),
			d.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.FECRate,
			prometheus.GaugeValue,
			float64(d.FECRate),
			d.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.FECRate_Base,
			prometheus.GaugeValue,
			float64(d.FECRate_Base),
			d.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.LossRate,
			prometheus.GaugeValue,
			float64(d.LossRate),
			d.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.LossRate_Base,
			prometheus.GaugeValue,
			float64(d.LossRate_Base),
			d.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.RetransmissionRate,
			prometheus.GaugeValue,
			float64(d.RetransmissionRate),
			d.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.RetransmissionRate_Base,
			prometheus.GaugeValue,
			float64(d.RetransmissionRate_Base),
			d.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.TCPReceivedRate,
			prometheus.GaugeValue,
			float64(d.TCPReceivedRate),
			d.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.TCPSentRate,
			prometheus.GaugeValue,
			float64(d.TCPSentRate),
			d.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.TotalReceivedRate,
			prometheus.GaugeValue,
			float64(d.TotalReceivedRate),
			d.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.TotalSentRate,
			prometheus.GaugeValue,
			float64(d.TotalSentRate),
			d.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.UDPPacketsReceivedPersec,
			prometheus.GaugeValue,
			float64(d.UDPPacketsReceivedPersec),
			d.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.UDPPacketsSentPersec,
			prometheus.GaugeValue,
			float64(d.UDPPacketsSentPersec),
			d.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.UDPReceivedRate,
			prometheus.GaugeValue,
			float64(d.UDPReceivedRate),
			d.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.UDPSentRate,
			prometheus.GaugeValue,
			float64(d.UDPSentRate),
			d.Name,
		)
	}
	return nil, nil
}

type Win32_PerfRawData_Counters_RemoteFXGraphics struct {
	Name                                               string
	AverageEncodingTime                                uint32
	FrameQuality                                       uint32
	FramesSkippedPerSecondInsufficientClientResources  uint32
	FramesSkippedPerSecondInsufficientNetworkResources uint32
	FramesSkippedPerSecondInsufficientServerResources  uint32
	GraphicsCompressionratio                           uint32
	InputFramesPerSecond                               uint32
	OutputFramesPerSecond                              uint32
	SourceFramesPerSecond                              uint32
}

func (c *RemoteFxCollector) collectRemoteFXGraphicsCounters(ch chan<- prometheus.Metric) (*prometheus.Desc, error) {
	var dst []Win32_PerfRawData_Counters_RemoteFXGraphics
	q := queryAll(&dst)
	if err := wmi.Query(q, &dst); err != nil {
		return nil, err
	}
	if len(dst) == 0 {
		return nil, errors.New("WMI query returned empty result set")
	}

	for _, d := range dst {
		ch <- prometheus.MustNewConstMetric(
			c.AverageEncodingTime,
			prometheus.GaugeValue,
			float64(d.AverageEncodingTime),
			d.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.FrameQuality,
			prometheus.GaugeValue,
			float64(d.FrameQuality),
			d.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.FramesSkippedPerSecondInsufficientClientResources,
			prometheus.GaugeValue,
			float64(d.FramesSkippedPerSecondInsufficientClientResources),
			d.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.FramesSkippedPerSecondInsufficientNetworkResources,
			prometheus.GaugeValue,
			float64(d.FramesSkippedPerSecondInsufficientNetworkResources),
			d.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.FramesSkippedPerSecondInsufficientServerResources,
			prometheus.GaugeValue,
			float64(d.FramesSkippedPerSecondInsufficientServerResources),
			d.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.GraphicsCompressionratio,
			prometheus.GaugeValue,
			float64(d.GraphicsCompressionratio),
			d.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.InputFramesPerSecond,
			prometheus.GaugeValue,
			float64(d.InputFramesPerSecond),
			d.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.OutputFramesPerSecond,
			prometheus.GaugeValue,
			float64(d.OutputFramesPerSecond),
			d.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.SourceFramesPerSecond,
			prometheus.GaugeValue,
			float64(d.SourceFramesPerSecond),
			d.Name,
		)
	}

	return nil, nil
}
