//go:build windows

package teradici_pcoip

import (
	"errors"

	"github.com/alecthomas/kingpin/v2"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/prometheus-community/windows_exporter/pkg/types"
	"github.com/prometheus-community/windows_exporter/pkg/wmi"
	"github.com/prometheus/client_golang/prometheus"
)

const Name = "teradici_pcoip"

type Config struct{}

var ConfigDefaults = Config{}

// Collector is a Prometheus Collector for WMI metrics:
// win32_PerfRawData_TeradiciPerf_PCoIPSessionAudioStatistics
// win32_PerfRawData_TeradiciPerf_PCoIPSessionGeneralStatistics
// win32_PerfRawData_TeradiciPerf_PCoIPSessionImagingStatistics
// win32_PerfRawData_TeradiciPerf_PCoIPSessionNetworkStatistics
// win32_PerfRawData_TeradiciPerf_PCoIPSessionUsbStatistics
type Collector struct {
	logger log.Logger

	audioBytesReceived       *prometheus.Desc
	audioBytesSent           *prometheus.Desc
	audioRXBWKBitPerSec      *prometheus.Desc
	audioTXBWKBitPerSec      *prometheus.Desc
	audioTXBWLimitKBitPerSec *prometheus.Desc

	bytesReceived          *prometheus.Desc
	bytesSent              *prometheus.Desc
	packetsReceived        *prometheus.Desc
	packetsSent            *prometheus.Desc
	rxPacketsLost          *prometheus.Desc
	sessionDurationSeconds *prometheus.Desc
	txPacketsLost          *prometheus.Desc

	imagingActiveMinimumQuality        *prometheus.Desc
	imagingApex2800Offload             *prometheus.Desc
	imagingBytesReceived               *prometheus.Desc
	imagingBytesSent                   *prometheus.Desc
	imagingDecoderCapabilityKBitPerSec *prometheus.Desc
	imagingEncodedFramesPerSec         *prometheus.Desc
	imagingMegapixelPerSec             *prometheus.Desc
	imagingNegativeAcknowledgements    *prometheus.Desc
	imagingRXBWKBitPerSec              *prometheus.Desc
	imagingSVGAdevTapframesPerSec      *prometheus.Desc
	imagingTXBWKBitPerSec              *prometheus.Desc

	RoundTripLatencyms        *prometheus.Desc
	rxBWKBitPerSec            *prometheus.Desc
	rxBWPeakKBitPerSec        *prometheus.Desc
	rxPacketLossPercent       *prometheus.Desc
	rxPacketLossPercentBase   *prometheus.Desc
	txBWActiveLimitKBitPerSec *prometheus.Desc
	txBWKBitPerSec            *prometheus.Desc
	txBWLimitKBitPerSec       *prometheus.Desc
	txPacketLossPercent       *prometheus.Desc
	txPacketLossPercentBase   *prometheus.Desc

	usbBytesReceived  *prometheus.Desc
	usbBytesSent      *prometheus.Desc
	usbRXBWKBitPerSec *prometheus.Desc
	usbTXBWKBitPerSec *prometheus.Desc
}

func New(logger log.Logger, _ *Config) *Collector {
	c := &Collector{}
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
	return []string{}, nil
}

func (c *Collector) Close() error {
	return nil
}

func (c *Collector) Build() error {
	// _ = level.Warn(c.logger).Log("msg", "teradici_pcoip collector is deprecated and will be removed in the future.")

	c.audioBytesReceived = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "audio_bytes_received_total"),
		"(AudioBytesReceived)",
		nil,
		nil,
	)
	c.audioBytesSent = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "audio_bytes_sent_total"),
		"(AudioBytesSent)",
		nil,
		nil,
	)
	c.audioRXBWKBitPerSec = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "audio_rx_bw_KBit_persec"),
		"(AudioRXBWKBitPerSec)",
		nil,
		nil,
	)
	c.audioTXBWKBitPerSec = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "audio_tx_bw_KBit_persec"),
		"(AudioTXBWKBitPerSec)",
		nil,
		nil,
	)
	c.audioTXBWLimitKBitPerSec = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "audio_tx_bw_limit_KBit_persec"),
		"(AudioTXBWLimitKBitPerSec)",
		nil,
		nil,
	)

	c.bytesReceived = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "bytes_received_total"),
		"(BytesReceived)",
		nil,
		nil,
	)
	c.bytesSent = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "bytes_sent_total"),
		"(BytesSent)",
		nil,
		nil,
	)
	c.packetsReceived = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "packets_received_total"),
		"(PacketsReceived)",
		nil,
		nil,
	)
	c.packetsSent = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "packets_sent_total"),
		"(PacketsSent)",
		nil,
		nil,
	)
	c.rxPacketsLost = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "rx_packets_lost_total"),
		"(RXPacketsLost)",
		nil,
		nil,
	)
	c.sessionDurationSeconds = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "session_duration_seconds_total"),
		"(SessionDurationSeconds)",
		nil,
		nil,
	)
	c.txPacketsLost = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "tx_packets_lost_total"),
		"(TXPacketsLost)",
		nil,
		nil,
	)

	c.imagingActiveMinimumQuality = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "imaging_active_min_quality"),
		"(ImagingActiveMinimumQuality)",
		nil,
		nil,
	)
	c.imagingApex2800Offload = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "imaging_apex2800_offload"),
		"(ImagingApex2800Offload)",
		nil,
		nil,
	)
	c.imagingBytesReceived = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "imaging_bytes_received_total"),
		"(ImagingBytesReceived)",
		nil,
		nil,
	)
	c.imagingBytesSent = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "imaging_bytes_sent_total"),
		"(ImagingBytesSent)",
		nil,
		nil,
	)
	c.imagingDecoderCapabilityKBitPerSec = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "imaging_decoder_capability_KBit_persec"),
		"(ImagingDecoderCapabilityKBitPerSec)",
		nil,
		nil,
	)
	c.imagingEncodedFramesPerSec = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "imaging_encoded_frames_persec"),
		"(ImagingEncodedFramesPerSec)",
		nil,
		nil,
	)
	c.imagingMegapixelPerSec = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "imaging_megapixel_persec"),
		"(ImagingMegapixelPerSec)",
		nil,
		nil,
	)
	c.imagingNegativeAcknowledgements = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "imaging_negative_acks_total"),
		"(ImagingNegativeAcknowledgements)",
		nil,
		nil,
	)
	c.imagingRXBWKBitPerSec = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "imaging_rx_bw_KBit_persec"),
		"(ImagingRXBWKBitPerSec)",
		nil,
		nil,
	)
	c.imagingSVGAdevTapframesPerSec = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "imaging_svga_devtap_frames_persec"),
		"(ImagingSVGAdevTapframesPerSec)",
		nil,
		nil,
	)
	c.imagingTXBWKBitPerSec = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "imaging_tx_bw_KBit_persec"),
		"(ImagingTXBWKBitPerSec)",
		nil,
		nil,
	)

	c.RoundTripLatencyms = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "round_trip_latency_ms"),
		"(RoundTripLatencyms)",
		nil,
		nil,
	)
	c.rxBWKBitPerSec = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "rx_bw_KBit_persec"),
		"(RXBWKBitPerSec)",
		nil,
		nil,
	)
	c.rxBWPeakKBitPerSec = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "rx_bw_peak_KBit_persec"),
		"(RXBWPeakKBitPerSec)",
		nil,
		nil,
	)
	c.rxPacketLossPercent = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "rx_packet_loss_percent"),
		"(RXPacketLossPercent)",
		nil,
		nil,
	)
	c.rxPacketLossPercentBase = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "rx_packet_loss_percent_base"),
		"(RXPacketLossPercent_Base)",
		nil,
		nil,
	)
	c.txBWActiveLimitKBitPerSec = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "tx_bw_active_limit_KBit_persec"),
		"(TXBWActiveLimitKBitPerSec)",
		nil,
		nil,
	)
	c.txBWKBitPerSec = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "tx_bw_KBit_persec"),
		"(TXBWKBitPerSec)",
		nil,
		nil,
	)
	c.txBWLimitKBitPerSec = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "tx_bw_limit_KBit_persec"),
		"(TXBWLimitKBitPerSec)",
		nil,
		nil,
	)
	c.txPacketLossPercent = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "tx_packet_loss_percent"),
		"(TXPacketLossPercent)",
		nil,
		nil,
	)
	c.txPacketLossPercentBase = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "tx_packet_loss_percent_base"),
		"(TXPacketLossPercent_Base)",
		nil,
		nil,
	)

	c.usbBytesReceived = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "usb_bytes_received_total"),
		"(USBBytesReceived)",
		nil,
		nil,
	)
	c.usbBytesSent = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "usb_bytes_sent_total"),
		"(USBBytesSent)",
		nil,
		nil,
	)
	c.usbRXBWKBitPerSec = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "usb_rx_bw_KBit_persec"),
		"(USBRXBWKBitPerSec)",
		nil,
		nil,
	)
	c.usbTXBWKBitPerSec = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "usb_tx_bw_KBit_persec"),
		"(USBTXBWKBitPerSec)",
		nil,
		nil,
	)
	return nil
}

// Collect sends the metric values for each metric
// to the provided prometheus Metric channel.
func (c *Collector) Collect(_ *types.ScrapeContext, ch chan<- prometheus.Metric) error {
	if err := c.collectAudio(ch); err != nil {
		_ = level.Error(c.logger).Log("msg", "failed collecting teradici session audio metrics", "err", err)
		return err
	}
	if err := c.collectGeneral(ch); err != nil {
		_ = level.Error(c.logger).Log("msg", "failed collecting teradici session general metrics", "err", err)
		return err
	}
	if err := c.collectImaging(ch); err != nil {
		_ = level.Error(c.logger).Log("msg", "failed collecting teradici session imaging metrics", "err", err)
		return err
	}
	if err := c.collectNetwork(ch); err != nil {
		_ = level.Error(c.logger).Log("msg", "failed collecting teradici session network metrics", "err", err)
		return err
	}
	if err := c.collectUsb(ch); err != nil {
		_ = level.Error(c.logger).Log("msg", "failed collecting teradici session USB metrics", "err", err)
		return err
	}
	return nil
}

type win32_PerfRawData_TeradiciPerf_PCoIPSessionAudioStatistics struct {
	AudioBytesReceived       uint64
	AudioBytesSent           uint64
	AudioRXBWKBitPerSec      uint64
	AudioTXBWKBitPerSec      uint64
	AudioTXBWLimitKBitPerSec uint64
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
	ImagingDecoderCapabilityKBitPerSec uint32
	ImagingEncodedFramesPerSec         uint32
	ImagingMegapixelPerSec             uint32
	ImagingNegativeAcknowledgements    uint32
	ImagingRXBWKBitPerSec              uint64
	ImagingSVGAdevTapframesPerSec      uint32
	ImagingTXBWKBitPerSec              uint64
}

type win32_PerfRawData_TeradiciPerf_PCoIPSessionNetworkStatistics struct {
	RoundTripLatencyms        uint32
	RXBWKBitPerSec            uint64
	RXBWPeakKBitPerSec        uint32
	RXPacketLossPercent       uint32
	RXPacketLossPercentBase   uint32
	TXBWActiveLimitKBitPerSec uint32
	TXBWKBitPerSec            uint64
	TXBWLimitKBitPerSec       uint32
	TXPacketLossPercent       uint32
	TXPacketLossPercentBase   uint32
}

type win32_PerfRawData_TeradiciPerf_PCoIPSessionUsbStatistics struct {
	USBBytesReceived  uint64
	USBBytesSent      uint64
	USBRXBWKBitPerSec uint64
	USBTXBWKBitPerSec uint64
}

func (c *Collector) collectAudio(ch chan<- prometheus.Metric) error {
	var dst []win32_PerfRawData_TeradiciPerf_PCoIPSessionAudioStatistics
	q := wmi.QueryAll(&dst, c.logger)
	if err := wmi.Query(q, &dst); err != nil {
		return err
	}
	if len(dst) == 0 {
		return errors.New("WMI query returned empty result set")
	}

	ch <- prometheus.MustNewConstMetric(
		c.audioBytesReceived,
		prometheus.CounterValue,
		float64(dst[0].AudioBytesReceived),
	)

	ch <- prometheus.MustNewConstMetric(
		c.audioBytesSent,
		prometheus.CounterValue,
		float64(dst[0].AudioBytesSent),
	)

	ch <- prometheus.MustNewConstMetric(
		c.audioRXBWKBitPerSec,
		prometheus.GaugeValue,
		float64(dst[0].AudioRXBWKBitPerSec),
	)

	ch <- prometheus.MustNewConstMetric(
		c.audioTXBWKBitPerSec,
		prometheus.GaugeValue,
		float64(dst[0].AudioTXBWKBitPerSec),
	)

	ch <- prometheus.MustNewConstMetric(
		c.audioTXBWLimitKBitPerSec,
		prometheus.GaugeValue,
		float64(dst[0].AudioTXBWLimitKBitPerSec),
	)

	return nil
}

func (c *Collector) collectGeneral(ch chan<- prometheus.Metric) error {
	var dst []win32_PerfRawData_TeradiciPerf_PCoIPSessionGeneralStatistics
	q := wmi.QueryAll(&dst, c.logger)
	if err := wmi.Query(q, &dst); err != nil {
		return err
	}
	if len(dst) == 0 {
		return errors.New("WMI query returned empty result set")
	}

	ch <- prometheus.MustNewConstMetric(
		c.bytesReceived,
		prometheus.CounterValue,
		float64(dst[0].BytesReceived),
	)

	ch <- prometheus.MustNewConstMetric(
		c.bytesSent,
		prometheus.CounterValue,
		float64(dst[0].BytesSent),
	)

	ch <- prometheus.MustNewConstMetric(
		c.packetsReceived,
		prometheus.CounterValue,
		float64(dst[0].PacketsReceived),
	)

	ch <- prometheus.MustNewConstMetric(
		c.packetsSent,
		prometheus.CounterValue,
		float64(dst[0].PacketsSent),
	)

	ch <- prometheus.MustNewConstMetric(
		c.rxPacketsLost,
		prometheus.CounterValue,
		float64(dst[0].RXPacketsLost),
	)

	ch <- prometheus.MustNewConstMetric(
		c.sessionDurationSeconds,
		prometheus.CounterValue,
		float64(dst[0].SessionDurationSeconds),
	)

	ch <- prometheus.MustNewConstMetric(
		c.txPacketsLost,
		prometheus.CounterValue,
		float64(dst[0].TXPacketsLost),
	)

	return nil
}

func (c *Collector) collectImaging(ch chan<- prometheus.Metric) error {
	var dst []win32_PerfRawData_TeradiciPerf_PCoIPSessionImagingStatistics
	q := wmi.QueryAll(&dst, c.logger)
	if err := wmi.Query(q, &dst); err != nil {
		return err
	}
	if len(dst) == 0 {
		return errors.New("WMI query returned empty result set")
	}

	ch <- prometheus.MustNewConstMetric(
		c.imagingActiveMinimumQuality,
		prometheus.GaugeValue,
		float64(dst[0].ImagingActiveMinimumQuality),
	)

	ch <- prometheus.MustNewConstMetric(
		c.imagingApex2800Offload,
		prometheus.GaugeValue,
		float64(dst[0].ImagingApex2800Offload),
	)

	ch <- prometheus.MustNewConstMetric(
		c.imagingBytesReceived,
		prometheus.CounterValue,
		float64(dst[0].ImagingBytesReceived),
	)

	ch <- prometheus.MustNewConstMetric(
		c.imagingBytesSent,
		prometheus.CounterValue,
		float64(dst[0].ImagingBytesSent),
	)

	ch <- prometheus.MustNewConstMetric(
		c.imagingDecoderCapabilityKBitPerSec,
		prometheus.GaugeValue,
		float64(dst[0].ImagingDecoderCapabilityKBitPerSec),
	)

	ch <- prometheus.MustNewConstMetric(
		c.imagingEncodedFramesPerSec,
		prometheus.GaugeValue,
		float64(dst[0].ImagingEncodedFramesPerSec),
	)

	ch <- prometheus.MustNewConstMetric(
		c.imagingMegapixelPerSec,
		prometheus.GaugeValue,
		float64(dst[0].ImagingMegapixelPerSec),
	)

	ch <- prometheus.MustNewConstMetric(
		c.imagingNegativeAcknowledgements,
		prometheus.CounterValue,
		float64(dst[0].ImagingNegativeAcknowledgements),
	)

	ch <- prometheus.MustNewConstMetric(
		c.imagingRXBWKBitPerSec,
		prometheus.GaugeValue,
		float64(dst[0].ImagingRXBWKBitPerSec),
	)

	ch <- prometheus.MustNewConstMetric(
		c.imagingSVGAdevTapframesPerSec,
		prometheus.GaugeValue,
		float64(dst[0].ImagingSVGAdevTapframesPerSec),
	)

	ch <- prometheus.MustNewConstMetric(
		c.imagingTXBWKBitPerSec,
		prometheus.GaugeValue,
		float64(dst[0].ImagingTXBWKBitPerSec),
	)

	return nil
}

func (c *Collector) collectNetwork(ch chan<- prometheus.Metric) error {
	var dst []win32_PerfRawData_TeradiciPerf_PCoIPSessionNetworkStatistics
	q := wmi.QueryAll(&dst, c.logger)
	if err := wmi.Query(q, &dst); err != nil {
		return err
	}
	if len(dst) == 0 {
		return errors.New("WMI query returned empty result set")
	}

	ch <- prometheus.MustNewConstMetric(
		c.RoundTripLatencyms,
		prometheus.GaugeValue,
		float64(dst[0].RoundTripLatencyms),
	)

	ch <- prometheus.MustNewConstMetric(
		c.rxBWKBitPerSec,
		prometheus.GaugeValue,
		float64(dst[0].RXBWKBitPerSec),
	)

	ch <- prometheus.MustNewConstMetric(
		c.rxBWPeakKBitPerSec,
		prometheus.GaugeValue,
		float64(dst[0].RXBWPeakKBitPerSec),
	)

	ch <- prometheus.MustNewConstMetric(
		c.rxPacketLossPercent,
		prometheus.GaugeValue,
		float64(dst[0].RXPacketLossPercent),
	)

	ch <- prometheus.MustNewConstMetric(
		c.rxPacketLossPercentBase,
		prometheus.GaugeValue,
		float64(dst[0].RXPacketLossPercentBase),
	)

	ch <- prometheus.MustNewConstMetric(
		c.txBWActiveLimitKBitPerSec,
		prometheus.GaugeValue,
		float64(dst[0].TXBWActiveLimitKBitPerSec),
	)

	ch <- prometheus.MustNewConstMetric(
		c.txBWKBitPerSec,
		prometheus.GaugeValue,
		float64(dst[0].TXBWKBitPerSec),
	)

	ch <- prometheus.MustNewConstMetric(
		c.txBWLimitKBitPerSec,
		prometheus.GaugeValue,
		float64(dst[0].TXBWLimitKBitPerSec),
	)

	ch <- prometheus.MustNewConstMetric(
		c.txPacketLossPercent,
		prometheus.GaugeValue,
		float64(dst[0].TXPacketLossPercent),
	)

	ch <- prometheus.MustNewConstMetric(
		c.txPacketLossPercentBase,
		prometheus.GaugeValue,
		float64(dst[0].TXPacketLossPercentBase),
	)

	return nil
}

func (c *Collector) collectUsb(ch chan<- prometheus.Metric) error {
	var dst []win32_PerfRawData_TeradiciPerf_PCoIPSessionUsbStatistics
	q := wmi.QueryAll(&dst, c.logger)
	if err := wmi.Query(q, &dst); err != nil {
		return err
	}
	if len(dst) == 0 {
		return errors.New("WMI query returned empty result set")
	}

	ch <- prometheus.MustNewConstMetric(
		c.usbBytesReceived,
		prometheus.CounterValue,
		float64(dst[0].USBBytesReceived),
	)

	ch <- prometheus.MustNewConstMetric(
		c.usbBytesSent,
		prometheus.CounterValue,
		float64(dst[0].USBBytesSent),
	)

	ch <- prometheus.MustNewConstMetric(
		c.usbRXBWKBitPerSec,
		prometheus.GaugeValue,
		float64(dst[0].USBRXBWKBitPerSec),
	)

	ch <- prometheus.MustNewConstMetric(
		c.usbTXBWKBitPerSec,
		prometheus.GaugeValue,
		float64(dst[0].USBTXBWKBitPerSec),
	)

	return nil
}
