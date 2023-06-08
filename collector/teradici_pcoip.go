//go:build windows
// +build windows

package collector

import (
	"errors"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/yusufpapurcu/wmi"
)

// A teradiciPcoipCollector is a Prometheus collector for WMI metrics:
// win32_PerfRawData_TeradiciPerf_PCoIPSessionAudioStatistics
// win32_PerfRawData_TeradiciPerf_PCoIPSessionGeneralStatistics
// win32_PerfRawData_TeradiciPerf_PCoIPSessionImagingStatistics
// win32_PerfRawData_TeradiciPerf_PCoIPSessionNetworkStatistics
// win32_PerfRawData_TeradiciPerf_PCoIPSessionUsbStatistics

type teradiciPcoipCollector struct {
	logger log.Logger

	AudioBytesReceived       *prometheus.Desc
	AudioBytesSent           *prometheus.Desc
	AudioRXBWkbitPersec      *prometheus.Desc
	AudioTXBWkbitPersec      *prometheus.Desc
	AudioTXBWLimitkbitPersec *prometheus.Desc

	BytesReceived          *prometheus.Desc
	BytesSent              *prometheus.Desc
	PacketsReceived        *prometheus.Desc
	PacketsSent            *prometheus.Desc
	RXPacketsLost          *prometheus.Desc
	SessionDurationSeconds *prometheus.Desc
	TXPacketsLost          *prometheus.Desc

	ImagingActiveMinimumQuality        *prometheus.Desc
	ImagingApex2800Offload             *prometheus.Desc
	ImagingBytesReceived               *prometheus.Desc
	ImagingBytesSent                   *prometheus.Desc
	ImagingDecoderCapabilitykbitPersec *prometheus.Desc
	ImagingEncodedFramesPersec         *prometheus.Desc
	ImagingMegapixelPersec             *prometheus.Desc
	ImagingNegativeAcknowledgements    *prometheus.Desc
	ImagingRXBWkbitPersec              *prometheus.Desc
	ImagingSVGAdevTapframesPersec      *prometheus.Desc
	ImagingTXBWkbitPersec              *prometheus.Desc

	RoundTripLatencyms        *prometheus.Desc
	RXBWkbitPersec            *prometheus.Desc
	RXBWPeakkbitPersec        *prometheus.Desc
	RXPacketLossPercent       *prometheus.Desc
	RXPacketLossPercent_Base  *prometheus.Desc
	TXBWActiveLimitkbitPersec *prometheus.Desc
	TXBWkbitPersec            *prometheus.Desc
	TXBWLimitkbitPersec       *prometheus.Desc
	TXPacketLossPercent       *prometheus.Desc
	TXPacketLossPercent_Base  *prometheus.Desc

	USBBytesReceived  *prometheus.Desc
	USBBytesSent      *prometheus.Desc
	USBRXBWkbitPersec *prometheus.Desc
	USBTXBWkbitPersec *prometheus.Desc
}

// newTeradiciPcoipCollector constructs a new teradiciPcoipCollector
func newTeradiciPcoipCollector(logger log.Logger) (Collector, error) {
	const subsystem = "teradici_pcoip"
	return &teradiciPcoipCollector{
		logger: log.With(logger, "collector", subsystem),
		AudioBytesReceived: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "audio_bytes_received_total"),
			"(AudioBytesReceived)",
			nil,
			nil,
		),
		AudioBytesSent: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "audio_bytes_sent_total"),
			"(AudioBytesSent)",
			nil,
			nil,
		),
		AudioRXBWkbitPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "audio_rx_bw_kbit_persec"),
			"(AudioRXBWkbitPersec)",
			nil,
			nil,
		),
		AudioTXBWkbitPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "audio_tx_bw_kbit_persec"),
			"(AudioTXBWkbitPersec)",
			nil,
			nil,
		),
		AudioTXBWLimitkbitPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "audio_tx_bw_limit_kbit_persec"),
			"(AudioTXBWLimitkbitPersec)",
			nil,
			nil,
		),

		BytesReceived: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "bytes_received_total"),
			"(BytesReceived)",
			nil,
			nil,
		),
		BytesSent: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "bytes_sent_total"),
			"(BytesSent)",
			nil,
			nil,
		),
		PacketsReceived: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "packets_received_total"),
			"(PacketsReceived)",
			nil,
			nil,
		),
		PacketsSent: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "packets_sent_total"),
			"(PacketsSent)",
			nil,
			nil,
		),
		RXPacketsLost: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "rx_packets_lost_total"),
			"(RXPacketsLost)",
			nil,
			nil,
		),
		SessionDurationSeconds: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "session_duration_seconds_total"),
			"(SessionDurationSeconds)",
			nil,
			nil,
		),
		TXPacketsLost: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "tx_packets_lost_total"),
			"(TXPacketsLost)",
			nil,
			nil,
		),

		ImagingActiveMinimumQuality: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "imaging_active_min_quality"),
			"(ImagingActiveMinimumQuality)",
			nil,
			nil,
		),
		ImagingApex2800Offload: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "imaging_apex2800_offload"),
			"(ImagingApex2800Offload)",
			nil,
			nil,
		),
		ImagingBytesReceived: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "imaging_bytes_received_total"),
			"(ImagingBytesReceived)",
			nil,
			nil,
		),
		ImagingBytesSent: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "imaging_bytes_sent_total"),
			"(ImagingBytesSent)",
			nil,
			nil,
		),
		ImagingDecoderCapabilitykbitPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "imaging_decoder_capability_kbit_persec"),
			"(ImagingDecoderCapabilitykbitPersec)",
			nil,
			nil,
		),
		ImagingEncodedFramesPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "imaging_encoded_frames_persec"),
			"(ImagingEncodedFramesPersec)",
			nil,
			nil,
		),
		ImagingMegapixelPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "imaging_megapixel_persec"),
			"(ImagingMegapixelPersec)",
			nil,
			nil,
		),
		ImagingNegativeAcknowledgements: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "imaging_negative_acks_total"),
			"(ImagingNegativeAcknowledgements)",
			nil,
			nil,
		),
		ImagingRXBWkbitPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "imaging_rx_bw_kbit_persec"),
			"(ImagingRXBWkbitPersec)",
			nil,
			nil,
		),
		ImagingSVGAdevTapframesPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "imaging_svga_devtap_frames_persec"),
			"(ImagingSVGAdevTapframesPersec)",
			nil,
			nil,
		),
		ImagingTXBWkbitPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "imaging_tx_bw_kbit_persec"),
			"(ImagingTXBWkbitPersec)",
			nil,
			nil,
		),

		RoundTripLatencyms: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "round_trip_latency_ms"),
			"(RoundTripLatencyms)",
			nil,
			nil,
		),
		RXBWkbitPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "rx_bw_kbit_persec"),
			"(RXBWkbitPersec)",
			nil,
			nil,
		),
		RXBWPeakkbitPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "rx_bw_peak_kbit_persec"),
			"(RXBWPeakkbitPersec)",
			nil,
			nil,
		),
		RXPacketLossPercent: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "rx_packet_loss_percent"),
			"(RXPacketLossPercent)",
			nil,
			nil,
		),
		RXPacketLossPercent_Base: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "rx_packet_loss_percent_base"),
			"(RXPacketLossPercent_Base)",
			nil,
			nil,
		),
		TXBWActiveLimitkbitPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "tx_bw_active_limit_kbit_persec"),
			"(TXBWActiveLimitkbitPersec)",
			nil,
			nil,
		),
		TXBWkbitPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "tx_bw_kbit_persec"),
			"(TXBWkbitPersec)",
			nil,
			nil,
		),
		TXBWLimitkbitPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "tx_bw_limit_kbit_persec"),
			"(TXBWLimitkbitPersec)",
			nil,
			nil,
		),
		TXPacketLossPercent: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "tx_packet_loss_percent"),
			"(TXPacketLossPercent)",
			nil,
			nil,
		),
		TXPacketLossPercent_Base: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "tx_packet_loss_percent_base"),
			"(TXPacketLossPercent_Base)",
			nil,
			nil,
		),

		USBBytesReceived: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "usb_bytes_received_total"),
			"(USBBytesReceived)",
			nil,
			nil,
		),
		USBBytesSent: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "usb_bytes_sent_total"),
			"(USBBytesSent)",
			nil,
			nil,
		),
		USBRXBWkbitPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "usb_rx_bw_kbit_persec"),
			"(USBRXBWkbitPersec)",
			nil,
			nil,
		),
		USBTXBWkbitPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "usb_tx_bw_kbit_persec"),
			"(USBTXBWkbitPersec)",
			nil,
			nil,
		),
	}, nil
}

// Collect sends the metric values for each metric
// to the provided prometheus Metric channel.
func (c *teradiciPcoipCollector) Collect(ctx *ScrapeContext, ch chan<- prometheus.Metric) error {
	if desc, err := c.collectAudio(ch); err != nil {
		_ = level.Error(c.logger).Log("failed collecting teradici session audio metrics", "desc", desc, "err", err)
		return err
	}
	if desc, err := c.collectGeneral(ch); err != nil {
		_ = level.Error(c.logger).Log("failed collecting teradici session general metrics", "desc", desc, "err", err)
		return err
	}
	if desc, err := c.collectImaging(ch); err != nil {
		_ = level.Error(c.logger).Log("failed collecting teradici session imaging metrics", "desc", desc, "err", err)
		return err
	}
	if desc, err := c.collectNetwork(ch); err != nil {
		_ = level.Error(c.logger).Log("failed collecting teradici session network metrics", "desc", desc, "err", err)
		return err
	}
	if desc, err := c.collectUsb(ch); err != nil {
		_ = level.Error(c.logger).Log("failed collecting teradici session USB metrics", "desc", desc, "err", err)
		return err
	}
	return nil
}

type win32_PerfRawData_TeradiciPerf_PCoIPSessionAudioStatistics struct {
	AudioBytesReceived       uint64
	AudioBytesSent           uint64
	AudioRXBWkbitPersec      uint64
	AudioTXBWkbitPersec      uint64
	AudioTXBWLimitkbitPersec uint64
}

type win32_PerfRawData_TeradiciPerf_PCoIPSessionGeneralStatistics struct {
	BytesReceived          uint64
	BytesSent              uint64
	PacketsReceived        uint64
	PacketsSent            uint64
	RXPacketsLost          uint64
	SessionDurationSeconds uint64
	TXPacketsLost          uint64
}

type win32_PerfRawData_TeradiciPerf_PCoIPSessionImagingStatistics struct {
	ImagingActiveMinimumQuality        uint32
	ImagingApex2800Offload             uint32
	ImagingBytesReceived               uint64
	ImagingBytesSent                   uint64
	ImagingDecoderCapabilitykbitPersec uint32
	ImagingEncodedFramesPersec         uint32
	ImagingMegapixelPersec             uint32
	ImagingNegativeAcknowledgements    uint32
	ImagingRXBWkbitPersec              uint64
	ImagingSVGAdevTapframesPersec      uint32
	ImagingTXBWkbitPersec              uint64
}

type win32_PerfRawData_TeradiciPerf_PCoIPSessionNetworkStatistics struct {
	RoundTripLatencyms        uint32
	RXBWkbitPersec            uint64
	RXBWPeakkbitPersec        uint32
	RXPacketLossPercent       uint32
	RXPacketLossPercent_Base  uint32
	TXBWActiveLimitkbitPersec uint32
	TXBWkbitPersec            uint64
	TXBWLimitkbitPersec       uint32
	TXPacketLossPercent       uint32
	TXPacketLossPercent_Base  uint32
}

type win32_PerfRawData_TeradiciPerf_PCoIPSessionUsbStatistics struct {
	USBBytesReceived  uint64
	USBBytesSent      uint64
	USBRXBWkbitPersec uint64
	USBTXBWkbitPersec uint64
}

func (c *teradiciPcoipCollector) collectAudio(ch chan<- prometheus.Metric) (*prometheus.Desc, error) {
	var dst []win32_PerfRawData_TeradiciPerf_PCoIPSessionAudioStatistics
	q := queryAll(&dst, c.logger)
	if err := wmi.Query(q, &dst); err != nil {
		return nil, err
	}
	if len(dst) == 0 {
		return nil, errors.New("WMI query returned empty result set")
	}

	ch <- prometheus.MustNewConstMetric(
		c.AudioBytesReceived,
		prometheus.CounterValue,
		float64(dst[0].AudioBytesReceived),
	)

	ch <- prometheus.MustNewConstMetric(
		c.AudioBytesSent,
		prometheus.CounterValue,
		float64(dst[0].AudioBytesSent),
	)

	ch <- prometheus.MustNewConstMetric(
		c.AudioRXBWkbitPersec,
		prometheus.GaugeValue,
		float64(dst[0].AudioRXBWkbitPersec),
	)

	ch <- prometheus.MustNewConstMetric(
		c.AudioTXBWkbitPersec,
		prometheus.GaugeValue,
		float64(dst[0].AudioTXBWkbitPersec),
	)

	ch <- prometheus.MustNewConstMetric(
		c.AudioTXBWLimitkbitPersec,
		prometheus.GaugeValue,
		float64(dst[0].AudioTXBWLimitkbitPersec),
	)

	return nil, nil
}

func (c *teradiciPcoipCollector) collectGeneral(ch chan<- prometheus.Metric) (*prometheus.Desc, error) {
	var dst []win32_PerfRawData_TeradiciPerf_PCoIPSessionGeneralStatistics
	q := queryAll(&dst, c.logger)
	if err := wmi.Query(q, &dst); err != nil {
		return nil, err
	}
	if len(dst) == 0 {
		return nil, errors.New("WMI query returned empty result set")
	}

	ch <- prometheus.MustNewConstMetric(
		c.BytesReceived,
		prometheus.CounterValue,
		float64(dst[0].BytesReceived),
	)

	ch <- prometheus.MustNewConstMetric(
		c.BytesSent,
		prometheus.CounterValue,
		float64(dst[0].BytesSent),
	)

	ch <- prometheus.MustNewConstMetric(
		c.PacketsReceived,
		prometheus.CounterValue,
		float64(dst[0].PacketsReceived),
	)

	ch <- prometheus.MustNewConstMetric(
		c.PacketsSent,
		prometheus.CounterValue,
		float64(dst[0].PacketsSent),
	)

	ch <- prometheus.MustNewConstMetric(
		c.RXPacketsLost,
		prometheus.CounterValue,
		float64(dst[0].RXPacketsLost),
	)

	ch <- prometheus.MustNewConstMetric(
		c.SessionDurationSeconds,
		prometheus.CounterValue,
		float64(dst[0].SessionDurationSeconds),
	)

	ch <- prometheus.MustNewConstMetric(
		c.TXPacketsLost,
		prometheus.CounterValue,
		float64(dst[0].TXPacketsLost),
	)

	return nil, nil
}

func (c *teradiciPcoipCollector) collectImaging(ch chan<- prometheus.Metric) (*prometheus.Desc, error) {
	var dst []win32_PerfRawData_TeradiciPerf_PCoIPSessionImagingStatistics
	q := queryAll(&dst, c.logger)
	if err := wmi.Query(q, &dst); err != nil {
		return nil, err
	}
	if len(dst) == 0 {
		return nil, errors.New("WMI query returned empty result set")
	}

	ch <- prometheus.MustNewConstMetric(
		c.ImagingActiveMinimumQuality,
		prometheus.GaugeValue,
		float64(dst[0].ImagingActiveMinimumQuality),
	)

	ch <- prometheus.MustNewConstMetric(
		c.ImagingApex2800Offload,
		prometheus.GaugeValue,
		float64(dst[0].ImagingApex2800Offload),
	)

	ch <- prometheus.MustNewConstMetric(
		c.ImagingBytesReceived,
		prometheus.CounterValue,
		float64(dst[0].ImagingBytesReceived),
	)

	ch <- prometheus.MustNewConstMetric(
		c.ImagingBytesSent,
		prometheus.CounterValue,
		float64(dst[0].ImagingBytesSent),
	)

	ch <- prometheus.MustNewConstMetric(
		c.ImagingDecoderCapabilitykbitPersec,
		prometheus.GaugeValue,
		float64(dst[0].ImagingDecoderCapabilitykbitPersec),
	)

	ch <- prometheus.MustNewConstMetric(
		c.ImagingEncodedFramesPersec,
		prometheus.GaugeValue,
		float64(dst[0].ImagingEncodedFramesPersec),
	)

	ch <- prometheus.MustNewConstMetric(
		c.ImagingMegapixelPersec,
		prometheus.GaugeValue,
		float64(dst[0].ImagingMegapixelPersec),
	)

	ch <- prometheus.MustNewConstMetric(
		c.ImagingNegativeAcknowledgements,
		prometheus.CounterValue,
		float64(dst[0].ImagingNegativeAcknowledgements),
	)

	ch <- prometheus.MustNewConstMetric(
		c.ImagingRXBWkbitPersec,
		prometheus.GaugeValue,
		float64(dst[0].ImagingRXBWkbitPersec),
	)

	ch <- prometheus.MustNewConstMetric(
		c.ImagingSVGAdevTapframesPersec,
		prometheus.GaugeValue,
		float64(dst[0].ImagingSVGAdevTapframesPersec),
	)

	ch <- prometheus.MustNewConstMetric(
		c.ImagingTXBWkbitPersec,
		prometheus.GaugeValue,
		float64(dst[0].ImagingTXBWkbitPersec),
	)

	return nil, nil
}

func (c *teradiciPcoipCollector) collectNetwork(ch chan<- prometheus.Metric) (*prometheus.Desc, error) {
	var dst []win32_PerfRawData_TeradiciPerf_PCoIPSessionNetworkStatistics
	q := queryAll(&dst, c.logger)
	if err := wmi.Query(q, &dst); err != nil {
		return nil, err
	}
	if len(dst) == 0 {
		return nil, errors.New("WMI query returned empty result set")
	}

	ch <- prometheus.MustNewConstMetric(
		c.RoundTripLatencyms,
		prometheus.GaugeValue,
		float64(dst[0].RoundTripLatencyms),
	)

	ch <- prometheus.MustNewConstMetric(
		c.RXBWkbitPersec,
		prometheus.GaugeValue,
		float64(dst[0].RXBWkbitPersec),
	)

	ch <- prometheus.MustNewConstMetric(
		c.RXBWPeakkbitPersec,
		prometheus.GaugeValue,
		float64(dst[0].RXBWPeakkbitPersec),
	)

	ch <- prometheus.MustNewConstMetric(
		c.RXPacketLossPercent,
		prometheus.GaugeValue,
		float64(dst[0].RXPacketLossPercent),
	)

	ch <- prometheus.MustNewConstMetric(
		c.RXPacketLossPercent_Base,
		prometheus.GaugeValue,
		float64(dst[0].RXPacketLossPercent_Base),
	)

	ch <- prometheus.MustNewConstMetric(
		c.TXBWActiveLimitkbitPersec,
		prometheus.GaugeValue,
		float64(dst[0].TXBWActiveLimitkbitPersec),
	)

	ch <- prometheus.MustNewConstMetric(
		c.TXBWkbitPersec,
		prometheus.GaugeValue,
		float64(dst[0].TXBWkbitPersec),
	)

	ch <- prometheus.MustNewConstMetric(
		c.TXBWLimitkbitPersec,
		prometheus.GaugeValue,
		float64(dst[0].TXBWLimitkbitPersec),
	)

	ch <- prometheus.MustNewConstMetric(
		c.TXPacketLossPercent,
		prometheus.GaugeValue,
		float64(dst[0].TXPacketLossPercent),
	)

	ch <- prometheus.MustNewConstMetric(
		c.TXPacketLossPercent_Base,
		prometheus.GaugeValue,
		float64(dst[0].TXPacketLossPercent_Base),
	)

	return nil, nil
}

func (c *teradiciPcoipCollector) collectUsb(ch chan<- prometheus.Metric) (*prometheus.Desc, error) {
	var dst []win32_PerfRawData_TeradiciPerf_PCoIPSessionUsbStatistics
	q := queryAll(&dst, c.logger)
	if err := wmi.Query(q, &dst); err != nil {
		return nil, err
	}
	if len(dst) == 0 {
		return nil, errors.New("WMI query returned empty result set")
	}

	ch <- prometheus.MustNewConstMetric(
		c.USBBytesReceived,
		prometheus.CounterValue,
		float64(dst[0].USBBytesReceived),
	)

	ch <- prometheus.MustNewConstMetric(
		c.USBBytesSent,
		prometheus.CounterValue,
		float64(dst[0].USBBytesSent),
	)

	ch <- prometheus.MustNewConstMetric(
		c.USBRXBWkbitPersec,
		prometheus.GaugeValue,
		float64(dst[0].USBRXBWkbitPersec),
	)

	ch <- prometheus.MustNewConstMetric(
		c.USBTXBWkbitPersec,
		prometheus.GaugeValue,
		float64(dst[0].USBTXBWkbitPersec),
	)

	return nil, nil
}
