//go:build windows

package vmware_blast

import (
	"github.com/alecthomas/kingpin/v2"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/prometheus-community/windows_exporter/pkg/types"
	"github.com/prometheus-community/windows_exporter/pkg/wmi"
	"github.com/prometheus/client_golang/prometheus"
)

const Name = "vmware_blast"

type Config struct{}

var ConfigDefaults = Config{}

// A collector is a Prometheus collector for WMI metrics:
// win32_PerfRawData_Counters_VMwareBlastAudioCounters
// win32_PerfRawData_Counters_VMwareBlastCDRCounters
// win32_PerfRawData_Counters_VMwareBlastClipboardCounters
// win32_PerfRawData_Counters_VMwareBlastHTML5MMRCounters
// win32_PerfRawData_Counters_VMwareBlastImagingCounters
// win32_PerfRawData_Counters_VMwareBlastRTAVCounters
// win32_PerfRawData_Counters_VMwareBlastSerialPortandScannerCounters
// win32_PerfRawData_Counters_VMwareBlastSessionCounters
// win32_PerfRawData_Counters_VMwareBlastSkypeforBusinessControlCounters
// win32_PerfRawData_Counters_VMwareBlastThinPrintCounters
// win32_PerfRawData_Counters_VMwareBlastUSBCounters
// win32_PerfRawData_Counters_VMwareBlastWindowsMediaMMRCounters

type Collector struct {
	config Config
	logger log.Logger

	audioReceivedBytes      *prometheus.Desc
	audioReceivedPackets    *prometheus.Desc
	audioTransmittedBytes   *prometheus.Desc
	audioTransmittedPackets *prometheus.Desc

	cdrReceivedBytes      *prometheus.Desc
	cdrReceivedPackets    *prometheus.Desc
	cdrTransmittedBytes   *prometheus.Desc
	cdrTransmittedPackets *prometheus.Desc

	clipboardReceivedBytes      *prometheus.Desc
	clipboardReceivedPackets    *prometheus.Desc
	clipboardTransmittedBytes   *prometheus.Desc
	clipboardTransmittedPackets *prometheus.Desc

	html5MMRReceivedBytes      *prometheus.Desc
	html5MMRReceivedPackets    *prometheus.Desc
	html5MMRTransmittedBytes   *prometheus.Desc
	html5MMRTransmittedPackets *prometheus.Desc

	imagingDirtyFramesPerSecond *prometheus.Desc
	imagingFBCRate              *prometheus.Desc
	imagingFramesPerSecond      *prometheus.Desc
	imagingPollRate             *prometheus.Desc
	imagingReceivedBytes        *prometheus.Desc
	imagingReceivedPackets      *prometheus.Desc
	imagingTotalDirtyFrames     *prometheus.Desc
	imagingTotalFBC             *prometheus.Desc
	imagingTotalFrames          *prometheus.Desc
	imagingTotalPoll            *prometheus.Desc
	imagingTransmittedBytes     *prometheus.Desc
	imagingTransmittedPackets   *prometheus.Desc

	rtAVReceivedBytes      *prometheus.Desc
	rtAVReceivedPackets    *prometheus.Desc
	rtAVTransmittedBytes   *prometheus.Desc
	rtAVTransmittedPackets *prometheus.Desc

	serialPortAndScannerReceivedBytes      *prometheus.Desc
	serialPortAndScannerReceivedPackets    *prometheus.Desc
	serialPortAndScannerTransmittedBytes   *prometheus.Desc
	serialPortAndScannerTransmittedPackets *prometheus.Desc

	sessionAutomaticReconnectCount              *prometheus.Desc
	sessionCumulativeReceivedBytesOverTCP       *prometheus.Desc
	sessionCumulativeReceivedBytesOverUDP       *prometheus.Desc
	sessionCumulativeTransmittedBytesOverTCP    *prometheus.Desc
	sessionCumulativeTransmittedBytesOverUDP    *prometheus.Desc
	sessionEstimatedBandwidthUplink             *prometheus.Desc
	sessionInstantaneousReceivedBytesOverTCP    *prometheus.Desc
	sessionInstantaneousReceivedBytesOverUDP    *prometheus.Desc
	sessionInstantaneousTransmittedBytesOverTCP *prometheus.Desc
	sessionInstantaneousTransmittedBytesOverUDP *prometheus.Desc
	sessionJitterUplink                         *prometheus.Desc
	sessionPacketLossUplink                     *prometheus.Desc
	sessionReceivedBytes                        *prometheus.Desc
	sessionReceivedPackets                      *prometheus.Desc
	sessionRTT                                  *prometheus.Desc
	sessionTransmittedBytes                     *prometheus.Desc
	sessionTransmittedPackets                   *prometheus.Desc

	skypeForBusinessControlReceivedBytes      *prometheus.Desc
	skypeForBusinessControlReceivedPackets    *prometheus.Desc
	skypeForBusinessControlTransmittedBytes   *prometheus.Desc
	skypeForBusinessControlTransmittedPackets *prometheus.Desc

	thinPrintReceivedBytes      *prometheus.Desc
	thinPrintReceivedPackets    *prometheus.Desc
	thinPrintTransmittedBytes   *prometheus.Desc
	thinPrintTransmittedPackets *prometheus.Desc

	usbReceivedBytes      *prometheus.Desc
	usbReceivedPackets    *prometheus.Desc
	usbTransmittedBytes   *prometheus.Desc
	usbTransmittedPackets *prometheus.Desc

	windowsMediaMMRReceivedBytes      *prometheus.Desc
	windowsMediaMMRReceivedPackets    *prometheus.Desc
	windowsMediaMMRTransmittedBytes   *prometheus.Desc
	windowsMediaMMRTransmittedPackets *prometheus.Desc
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
	return []string{}, nil
}

func (c *Collector) Close() error {
	return nil
}

func (c *Collector) Build() error {
	_ = level.Warn(c.logger).Log("msg", "vmware_blast collector is deprecated and will be removed in the future.")

	c.audioReceivedBytes = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "audio_received_bytes_total"),
		"(AudioReceivedBytes)",
		nil,
		nil,
	)
	c.audioReceivedPackets = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "audio_received_packets_total"),
		"(AudioReceivedPackets)",
		nil,
		nil,
	)
	c.audioTransmittedBytes = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "audio_transmitted_bytes_total"),
		"(AudioTransmittedBytes)",
		nil,
		nil,
	)
	c.audioTransmittedPackets = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "audio_transmitted_packets_total"),
		"(AudioTransmittedPackets)",
		nil,
		nil,
	)

	c.cdrReceivedBytes = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "cdr_received_bytes_total"),
		"(CDRReceivedBytes)",
		nil,
		nil,
	)
	c.cdrReceivedPackets = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "cdr_received_packets_total"),
		"(CDRReceivedPackets)",
		nil,
		nil,
	)
	c.cdrTransmittedBytes = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "cdr_transmitted_bytes_total"),
		"(CDRTransmittedBytes)",
		nil,
		nil,
	)
	c.cdrTransmittedPackets = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "cdr_transmitted_packets_total"),
		"(CDRTransmittedPackets)",
		nil,
		nil,
	)

	c.clipboardReceivedBytes = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "clipboard_received_bytes_total"),
		"(ClipboardReceivedBytes)",
		nil,
		nil,
	)
	c.clipboardReceivedPackets = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "clipboard_received_packets_total"),
		"(ClipboardReceivedPackets)",
		nil,
		nil,
	)
	c.clipboardTransmittedBytes = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "clipboard_transmitted_bytes_total"),
		"(ClipboardTransmittedBytes)",
		nil,
		nil,
	)
	c.clipboardTransmittedPackets = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "clipboard_transmitted_packets_total"),
		"(ClipboardTransmittedPackets)",
		nil,
		nil,
	)

	c.html5MMRReceivedBytes = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "html5_mmr_received_bytes_total"),
		"(HTML5MMRReceivedBytes)",
		nil,
		nil,
	)
	c.html5MMRReceivedPackets = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "html5_mmr_received_packets_total"),
		"(HTML5MMRReceivedPackets)",
		nil,
		nil,
	)
	c.html5MMRTransmittedBytes = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "html5_mmr_transmitted_bytes_total"),
		"(HTML5MMRTransmittedBytes)",
		nil,
		nil,
	)
	c.html5MMRTransmittedPackets = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "html5_mmr_transmitted_packets_total"),
		"(HTML5MMRTransmittedPackets)",
		nil,
		nil,
	)

	c.imagingDirtyFramesPerSecond = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "imaging_dirty_frames_per_second"),
		"(ImagingDirtyFramesPerSecond)",
		nil,
		nil,
	)
	c.imagingFBCRate = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "imaging_fbc_rate"),
		"(ImagingFBCRate)",
		nil,
		nil,
	)
	c.imagingFramesPerSecond = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "imaging_frames_per_second"),
		"(ImagingFramesPerSecond)",
		nil,
		nil,
	)
	c.imagingPollRate = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "imaging_poll_rate"),
		"(ImagingPollRate)",
		nil,
		nil,
	)
	c.imagingReceivedBytes = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "imaging_received_bytes_total"),
		"(ImagingReceivedBytes)",
		nil,
		nil,
	)
	c.imagingReceivedPackets = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "imaging_received_packets_total"),
		"(ImagingReceivedPackets)",
		nil,
		nil,
	)
	c.imagingTotalDirtyFrames = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "imaging_total_dirty_frames_total"),
		"(ImagingTotalDirtyFrames)",
		nil,
		nil,
	)
	c.imagingTotalFBC = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "imaging_fbc_total"),
		"(ImagingTotalFBC)",
		nil,
		nil,
	)
	c.imagingTotalFrames = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "imaging_frames_total"),
		"(ImagingTotalFrames)",
		nil,
		nil,
	)
	c.imagingTotalPoll = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "imaging_poll_total"),
		"(ImagingTotalPoll)",
		nil,
		nil,
	)
	c.imagingTransmittedBytes = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "imaging_transmitted_bytes_total"),
		"(ImagingTransmittedBytes)",
		nil,
		nil,
	)
	c.imagingTransmittedPackets = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "imaging_transmitted_packets_total"),
		"(ImagingTransmittedPackets)",
		nil,
		nil,
	)

	c.rtAVReceivedBytes = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "rtav_received_bytes_total"),
		"(RTAVReceivedBytes)",
		nil,
		nil,
	)
	c.rtAVReceivedPackets = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "rtav_received_packets_total"),
		"(RTAVReceivedPackets)",
		nil,
		nil,
	)
	c.rtAVTransmittedBytes = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "rtav_transmitted_bytes_total"),
		"(RTAVTransmittedBytes)",
		nil,
		nil,
	)
	c.rtAVTransmittedPackets = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "rtav_transmitted_packets_total"),
		"(RTAVTransmittedPackets)",
		nil,
		nil,
	)

	c.serialPortAndScannerReceivedBytes = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "serial_port_and_scanner_received_bytes_total"),
		"(SerialPortandScannerReceivedBytes)",
		nil,
		nil,
	)
	c.serialPortAndScannerReceivedPackets = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "serial_port_and_scanner_received_packets_total"),
		"(SerialPortandScannerReceivedPackets)",
		nil,
		nil,
	)
	c.serialPortAndScannerTransmittedBytes = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "serial_port_and_scanner_transmitted_bytes_total"),
		"(SerialPortandScannerTransmittedBytes)",
		nil,
		nil,
	)
	c.serialPortAndScannerTransmittedPackets = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "serial_port_and_scanner_transmitted_packets_total"),
		"(SerialPortandScannerTransmittedPackets)",
		nil,
		nil,
	)

	c.sessionAutomaticReconnectCount = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "session_automatic_reconnect_count_total"),
		"(SessionAutomaticReconnectCount)",
		nil,
		nil,
	)
	c.sessionCumulativeReceivedBytesOverTCP = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "session_cumlative_received_bytes_over_tcp_total"),
		"(SessionCumulativeReceivedBytesOverTCP)",
		nil,
		nil,
	)
	c.sessionCumulativeReceivedBytesOverUDP = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "session_cumlative_received_bytes_over_udp_total"),
		"(SessionCumulativeReceivedBytesOverUDP)",
		nil,
		nil,
	)
	c.sessionCumulativeTransmittedBytesOverTCP = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "session_cumlative_transmitted_bytes_over_tcp_total"),
		"(SessionCumulativeTransmittedBytesOverTCP)",
		nil,
		nil,
	)
	c.sessionCumulativeTransmittedBytesOverUDP = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "session_cumlative_transmitted_bytes_over_udp_total"),
		"(SessionCumulativeTransmittedBytesOverUDP)",
		nil,
		nil,
	)
	c.sessionEstimatedBandwidthUplink = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "session_estimated_bandwidth_uplink"),
		"(SessionEstimatedBandwidthUplink)",
		nil,
		nil,
	)
	c.sessionInstantaneousReceivedBytesOverTCP = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "session_instantaneous_received_bytes_over_tcp_total"),
		"(SessionInstantaneousReceivedBytesOverTCP)",
		nil,
		nil,
	)
	c.sessionInstantaneousReceivedBytesOverUDP = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "session_instantaneous_received_bytes_over_udp_total"),
		"(SessionInstantaneousReceivedBytesOverUDP)",
		nil,
		nil,
	)
	c.sessionInstantaneousTransmittedBytesOverTCP = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "session_instantaneous_transmitted_bytes_over_tcp_total"),
		"(SessionInstantaneousTransmittedBytesOverTCP)",
		nil,
		nil,
	)
	c.sessionInstantaneousTransmittedBytesOverUDP = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "session_instantaneous_transmitted_bytes_over_udp_total"),
		"(SessionInstantaneousTransmittedBytesOverUDP)",
		nil,
		nil,
	)
	c.sessionJitterUplink = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "session_jitter_uplink"),
		"(SessionJitterUplink)",
		nil,
		nil,
	)
	c.sessionPacketLossUplink = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "session_packet_loss_uplink"),
		"(SessionPacketLossUplink)",
		nil,
		nil,
	)
	c.sessionReceivedBytes = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "session_received_bytes_total"),
		"(SessionReceivedBytes)",
		nil,
		nil,
	)
	c.sessionReceivedPackets = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "session_received_packets_total"),
		"(SessionReceivedPackets)",
		nil,
		nil,
	)
	c.sessionRTT = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "session_rtt"),
		"(SessionRTT)",
		nil,
		nil,
	)
	c.sessionTransmittedBytes = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "session_transmitted_bytes_total"),
		"(SessionTransmittedBytes)",
		nil,
		nil,
	)
	c.sessionTransmittedPackets = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "session_transmitted_packets_total"),
		"(SessionTransmittedPackets)",
		nil,
		nil,
	)

	c.skypeForBusinessControlReceivedBytes = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "skype_for_business_control_received_bytes_total"),
		"(SkypeforBusinessControlReceivedBytes)",
		nil,
		nil,
	)
	c.skypeForBusinessControlReceivedPackets = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "skype_for_business_control_received_packets_total"),
		"(SkypeforBusinessControlReceivedPackets)",
		nil,
		nil,
	)
	c.skypeForBusinessControlTransmittedBytes = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "skype_for_business_control_transmitted_bytes_total"),
		"(SkypeforBusinessControlTransmittedBytes)",
		nil,
		nil,
	)
	c.skypeForBusinessControlTransmittedPackets = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "skype_for_business_control_transmitted_packets_total"),
		"(SkypeforBusinessControlTransmittedPackets)",
		nil,
		nil,
	)

	c.thinPrintReceivedBytes = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "thinprint_received_bytes_total"),
		"(ThinPrintReceivedBytes)",
		nil,
		nil,
	)
	c.thinPrintReceivedPackets = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "thinprint_received_packets_total"),
		"(ThinPrintReceivedPackets)",
		nil,
		nil,
	)
	c.thinPrintTransmittedBytes = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "thinprint_transmitted_bytes_total"),
		"(ThinPrintTransmittedBytes)",
		nil,
		nil,
	)
	c.thinPrintTransmittedPackets = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "thinprint_transmitted_packets_total"),
		"(ThinPrintTransmittedPackets)",
		nil,
		nil,
	)

	c.usbReceivedBytes = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "usb_received_bytes_total"),
		"(USBReceivedBytes)",
		nil,
		nil,
	)
	c.usbReceivedPackets = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "usb_received_packets_total"),
		"(USBReceivedPackets)",
		nil,
		nil,
	)
	c.usbTransmittedBytes = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "usb_transmitted_bytes_total"),
		"(USBTransmittedBytes)",
		nil,
		nil,
	)
	c.usbTransmittedPackets = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "usb_transmitted_packets_total"),
		"(USBTransmittedPackets)",
		nil,
		nil,
	)

	c.windowsMediaMMRReceivedBytes = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "windows_media_mmr_received_bytes_total"),
		"(WindowsMediaMMRReceivedBytes)",
		nil,
		nil,
	)
	c.windowsMediaMMRReceivedPackets = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "windows_media_mmr_received_packets_total"),
		"(WindowsMediaMMRReceivedPackets)",
		nil,
		nil,
	)
	c.windowsMediaMMRTransmittedBytes = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "windows_media_mmr_transmitted_bytes_total"),
		"(WindowsMediaMMRTransmittedBytes)",
		nil,
		nil,
	)
	c.windowsMediaMMRTransmittedPackets = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "windows_media_mmr_transmitted_packets_total"),
		"(WindowsMediaMMRTransmittedPackets)",
		nil,
		nil,
	)
	return nil
}

// Collect sends the metric values for each metric
// to the provided prometheus Metric channel.
func (c *Collector) Collect(_ *types.ScrapeContext, ch chan<- prometheus.Metric) error {
	if err := c.collectAudio(ch); err != nil {
		_ = level.Error(c.logger).Log("msg", "failed collecting vmware blast audio metrics", "err", err)
		return err
	}
	if err := c.collectCdr(ch); err != nil {
		_ = level.Error(c.logger).Log("msg", "failed collecting vmware blast CDR metrics", "err", err)
		return err
	}
	if err := c.collectClipboard(ch); err != nil {
		_ = level.Error(c.logger).Log("msg", "failed collecting vmware blast clipboard metrics", "err", err)
		return err
	}
	if err := c.collectHtml5Mmr(ch); err != nil {
		_ = level.Error(c.logger).Log("msg", "failed collecting vmware blast HTML5 MMR metrics", "err", err)
		return err
	}
	if err := c.collectImaging(ch); err != nil {
		_ = level.Error(c.logger).Log("msg", "failed collecting vmware blast imaging metrics", "err", err)
		return err
	}
	if err := c.collectRtav(ch); err != nil {
		_ = level.Error(c.logger).Log("msg", "failed collecting vmware blast RTAV metrics", "err", err)
		return err
	}
	if err := c.collectSerialPortandScanner(ch); err != nil {
		_ = level.Error(c.logger).Log("msg", "failed collecting vmware blast serial port and scanner metrics", "err", err)
		return err
	}
	if err := c.collectSession(ch); err != nil {
		_ = level.Error(c.logger).Log("msg", "failed collecting vmware blast metrics", "err", err)
		return err
	}
	if err := c.collectSkypeforBusinessControl(ch); err != nil {
		_ = level.Error(c.logger).Log("msg", "failed collecting vmware blast skype for business control metrics", "err", err)
		return err
	}
	if err := c.collectThinPrint(ch); err != nil {
		_ = level.Error(c.logger).Log("msg", "failed collecting vmware blast thin print metrics", "err", err)
		return err
	}
	if err := c.collectUsb(ch); err != nil {
		_ = level.Error(c.logger).Log("msg", "failed collecting vmware blast USB metrics", "err", err)
		return err
	}
	if err := c.collectWindowsMediaMmr(ch); err != nil {
		_ = level.Error(c.logger).Log("msg", "failed collecting vmware blast windows media MMR metrics", "err", err)
		return err
	}
	return nil
}

type win32_PerfRawData_Counters_VMwareBlastAudioCounters struct {
	ReceivedBytes      uint32
	ReceivedPackets    uint32
	TransmittedBytes   uint32
	TransmittedPackets uint32
}

type win32_PerfRawData_Counters_VMwareBlastCDRCounters struct {
	ReceivedBytes      uint32
	ReceivedPackets    uint32
	TransmittedBytes   uint32
	TransmittedPackets uint32
}

type win32_PerfRawData_Counters_VMwareBlastClipboardCounters struct {
	ReceivedBytes      uint32
	ReceivedPackets    uint32
	TransmittedBytes   uint32
	TransmittedPackets uint32
}

type win32_PerfRawData_Counters_VMwareBlastHTML5MMRcounters struct {
	ReceivedBytes      uint32
	ReceivedPackets    uint32
	TransmittedBytes   uint32
	TransmittedPackets uint32
}

type win32_PerfRawData_Counters_VMwareBlastImagingCounters struct {
	Dirtyframespersecond uint32
	FBCRate              uint32
	Framespersecond      uint32
	PollRate             uint32
	ReceivedBytes        uint32
	ReceivedPackets      uint32
	Totaldirtyframes     uint32
	TotalFBC             uint32
	Totalframes          uint32
	Totalpoll            uint32
	TransmittedBytes     uint32
	TransmittedPackets   uint32
}

type win32_PerfRawData_Counters_VMwareBlastRTAVCounters struct {
	ReceivedBytes      uint32
	ReceivedPackets    uint32
	TransmittedBytes   uint32
	TransmittedPackets uint32
}

type win32_PerfRawData_Counters_VMwareBlastSerialPortandScannerCounters struct {
	ReceivedBytes      uint32
	ReceivedPackets    uint32
	TransmittedBytes   uint32
	TransmittedPackets uint32
}

type win32_PerfRawData_Counters_VMwareBlastSessionCounters struct {
	AutomaticReconnectCount              uint32
	CumulativeReceivedBytesoverTCP       uint32
	CumulativeReceivedBytesoverUDP       uint32
	CumulativeTransmittedBytesoverTCP    uint32
	CumulativeTransmittedBytesoverUDP    uint32
	EstimatedBandwidthUplink             uint32
	InstantaneousReceivedBytesoverTCP    uint32
	InstantaneousReceivedBytesoverUDP    uint32
	InstantaneousTransmittedBytesoverTCP uint32
	InstantaneousTransmittedBytesoverUDP uint32
	JitterUplink                         uint32
	PacketLossUplink                     uint32
	ReceivedBytes                        uint32
	ReceivedPackets                      uint32
	RTT                                  uint32
	TransmittedBytes                     uint32
	TransmittedPackets                   uint32
}

type win32_PerfRawData_Counters_VMwareBlastSkypeforBusinessControlCounters struct {
	ReceivedBytes      uint32
	ReceivedPackets    uint32
	TransmittedBytes   uint32
	TransmittedPackets uint32
}

type win32_PerfRawData_Counters_VMwareBlastThinPrintCounters struct {
	ReceivedBytes      uint32
	ReceivedPackets    uint32
	TransmittedBytes   uint32
	TransmittedPackets uint32
}

type win32_PerfRawData_Counters_VMwareBlastUSBCounters struct {
	ReceivedBytes      uint32
	ReceivedPackets    uint32
	TransmittedBytes   uint32
	TransmittedPackets uint32
}

type win32_PerfRawData_Counters_VMwareBlastWindowsMediaMMRCounters struct {
	ReceivedBytes      uint32
	ReceivedPackets    uint32
	TransmittedBytes   uint32
	TransmittedPackets uint32
}

func (c *Collector) collectAudio(ch chan<- prometheus.Metric) error {
	var dst []win32_PerfRawData_Counters_VMwareBlastAudioCounters
	q := wmi.QueryAll(&dst, c.logger)
	if err := wmi.Query(q, &dst); err != nil {
		return err
	}

	if len(dst) == 0 {
		// It's possible for these classes to legitimately return null when queried
		return nil
	}

	ch <- prometheus.MustNewConstMetric(
		c.audioReceivedBytes,
		prometheus.CounterValue,
		float64(dst[0].ReceivedBytes),
	)

	ch <- prometheus.MustNewConstMetric(
		c.audioReceivedPackets,
		prometheus.CounterValue,
		float64(dst[0].ReceivedPackets),
	)

	ch <- prometheus.MustNewConstMetric(
		c.audioTransmittedBytes,
		prometheus.CounterValue,
		float64(dst[0].TransmittedBytes),
	)

	ch <- prometheus.MustNewConstMetric(
		c.audioTransmittedPackets,
		prometheus.CounterValue,
		float64(dst[0].TransmittedPackets),
	)

	return nil
}

func (c *Collector) collectCdr(ch chan<- prometheus.Metric) error {
	var dst []win32_PerfRawData_Counters_VMwareBlastCDRCounters
	q := wmi.QueryAll(&dst, c.logger)
	if err := wmi.Query(q, &dst); err != nil {
		return err
	}

	if len(dst) == 0 {
		// It's possible for these classes to legitimately return null when queried
		return nil
	}

	ch <- prometheus.MustNewConstMetric(
		c.cdrReceivedBytes,
		prometheus.CounterValue,
		float64(dst[0].ReceivedBytes),
	)

	ch <- prometheus.MustNewConstMetric(
		c.cdrReceivedPackets,
		prometheus.CounterValue,
		float64(dst[0].ReceivedPackets),
	)

	ch <- prometheus.MustNewConstMetric(
		c.cdrTransmittedBytes,
		prometheus.CounterValue,
		float64(dst[0].TransmittedBytes),
	)

	ch <- prometheus.MustNewConstMetric(
		c.cdrTransmittedPackets,
		prometheus.CounterValue,
		float64(dst[0].TransmittedPackets),
	)

	return nil
}

func (c *Collector) collectClipboard(ch chan<- prometheus.Metric) error {
	var dst []win32_PerfRawData_Counters_VMwareBlastClipboardCounters
	q := wmi.QueryAll(&dst, c.logger)
	if err := wmi.Query(q, &dst); err != nil {
		return err
	}

	if len(dst) == 0 {
		// It's possible for these classes to legitimately return null when queried
		return nil
	}

	ch <- prometheus.MustNewConstMetric(
		c.clipboardReceivedBytes,
		prometheus.CounterValue,
		float64(dst[0].ReceivedBytes),
	)

	ch <- prometheus.MustNewConstMetric(
		c.clipboardReceivedPackets,
		prometheus.CounterValue,
		float64(dst[0].ReceivedPackets),
	)

	ch <- prometheus.MustNewConstMetric(
		c.clipboardTransmittedBytes,
		prometheus.CounterValue,
		float64(dst[0].TransmittedBytes),
	)

	ch <- prometheus.MustNewConstMetric(
		c.clipboardTransmittedPackets,
		prometheus.CounterValue,
		float64(dst[0].TransmittedPackets),
	)

	return nil
}

func (c *Collector) collectHtml5Mmr(ch chan<- prometheus.Metric) error {
	var dst []win32_PerfRawData_Counters_VMwareBlastHTML5MMRcounters
	q := wmi.QueryAll(&dst, c.logger)
	if err := wmi.Query(q, &dst); err != nil {
		return err
	}

	if len(dst) == 0 {
		// It's possible for these classes to legitimately return null when queried
		return nil
	}

	ch <- prometheus.MustNewConstMetric(
		c.html5MMRReceivedBytes,
		prometheus.CounterValue,
		float64(dst[0].ReceivedBytes),
	)

	ch <- prometheus.MustNewConstMetric(
		c.html5MMRReceivedPackets,
		prometheus.CounterValue,
		float64(dst[0].ReceivedPackets),
	)

	ch <- prometheus.MustNewConstMetric(
		c.html5MMRTransmittedBytes,
		prometheus.CounterValue,
		float64(dst[0].TransmittedBytes),
	)

	ch <- prometheus.MustNewConstMetric(
		c.html5MMRTransmittedPackets,
		prometheus.CounterValue,
		float64(dst[0].TransmittedPackets),
	)

	return nil
}

func (c *Collector) collectImaging(ch chan<- prometheus.Metric) error {
	var dst []win32_PerfRawData_Counters_VMwareBlastImagingCounters
	q := wmi.QueryAll(&dst, c.logger)
	if err := wmi.Query(q, &dst); err != nil {
		return err
	}

	if len(dst) == 0 {
		// It's possible for these classes to legitimately return null when queried
		return nil
	}

	ch <- prometheus.MustNewConstMetric(
		c.imagingDirtyFramesPerSecond,
		prometheus.GaugeValue,
		float64(dst[0].Dirtyframespersecond),
	)

	ch <- prometheus.MustNewConstMetric(
		c.imagingFBCRate,
		prometheus.GaugeValue,
		float64(dst[0].FBCRate),
	)

	ch <- prometheus.MustNewConstMetric(
		c.imagingFramesPerSecond,
		prometheus.GaugeValue,
		float64(dst[0].Framespersecond),
	)

	ch <- prometheus.MustNewConstMetric(
		c.imagingPollRate,
		prometheus.GaugeValue,
		float64(dst[0].PollRate),
	)

	ch <- prometheus.MustNewConstMetric(
		c.imagingReceivedBytes,
		prometheus.CounterValue,
		float64(dst[0].ReceivedBytes),
	)

	ch <- prometheus.MustNewConstMetric(
		c.imagingReceivedPackets,
		prometheus.CounterValue,
		float64(dst[0].ReceivedPackets),
	)

	ch <- prometheus.MustNewConstMetric(
		c.imagingTotalDirtyFrames,
		prometheus.CounterValue,
		float64(dst[0].Totaldirtyframes),
	)

	ch <- prometheus.MustNewConstMetric(
		c.imagingTotalFBC,
		prometheus.CounterValue,
		float64(dst[0].TotalFBC),
	)

	ch <- prometheus.MustNewConstMetric(
		c.imagingTotalFrames,
		prometheus.CounterValue,
		float64(dst[0].Totalframes),
	)

	ch <- prometheus.MustNewConstMetric(
		c.imagingTotalPoll,
		prometheus.CounterValue,
		float64(dst[0].Totalpoll),
	)

	ch <- prometheus.MustNewConstMetric(
		c.imagingTransmittedBytes,
		prometheus.CounterValue,
		float64(dst[0].TransmittedBytes),
	)

	ch <- prometheus.MustNewConstMetric(
		c.imagingTransmittedPackets,
		prometheus.CounterValue,
		float64(dst[0].TransmittedPackets),
	)

	return nil
}

func (c *Collector) collectRtav(ch chan<- prometheus.Metric) error {
	var dst []win32_PerfRawData_Counters_VMwareBlastRTAVCounters
	q := wmi.QueryAll(&dst, c.logger)
	if err := wmi.Query(q, &dst); err != nil {
		return err
	}

	if len(dst) == 0 {
		// It's possible for these classes to legitimately return null when queried
		return nil
	}

	ch <- prometheus.MustNewConstMetric(
		c.rtAVReceivedBytes,
		prometheus.CounterValue,
		float64(dst[0].ReceivedBytes),
	)

	ch <- prometheus.MustNewConstMetric(
		c.rtAVReceivedPackets,
		prometheus.CounterValue,
		float64(dst[0].ReceivedPackets),
	)

	ch <- prometheus.MustNewConstMetric(
		c.rtAVTransmittedBytes,
		prometheus.CounterValue,
		float64(dst[0].TransmittedBytes),
	)

	ch <- prometheus.MustNewConstMetric(
		c.rtAVTransmittedPackets,
		prometheus.CounterValue,
		float64(dst[0].TransmittedPackets),
	)

	return nil
}

func (c *Collector) collectSerialPortandScanner(ch chan<- prometheus.Metric) error {
	var dst []win32_PerfRawData_Counters_VMwareBlastSerialPortandScannerCounters
	q := wmi.QueryAll(&dst, c.logger)
	if err := wmi.Query(q, &dst); err != nil {
		return err
	}

	if len(dst) == 0 {
		// It's possible for these classes to legitimately return null when queried
		return nil
	}

	ch <- prometheus.MustNewConstMetric(
		c.serialPortAndScannerReceivedBytes,
		prometheus.CounterValue,
		float64(dst[0].ReceivedBytes),
	)

	ch <- prometheus.MustNewConstMetric(
		c.serialPortAndScannerReceivedPackets,
		prometheus.CounterValue,
		float64(dst[0].ReceivedPackets),
	)

	ch <- prometheus.MustNewConstMetric(
		c.serialPortAndScannerTransmittedBytes,
		prometheus.CounterValue,
		float64(dst[0].TransmittedBytes),
	)

	ch <- prometheus.MustNewConstMetric(
		c.serialPortAndScannerTransmittedPackets,
		prometheus.CounterValue,
		float64(dst[0].TransmittedPackets),
	)

	return nil
}

func (c *Collector) collectSession(ch chan<- prometheus.Metric) error {
	var dst []win32_PerfRawData_Counters_VMwareBlastSessionCounters
	q := wmi.QueryAll(&dst, c.logger)
	if err := wmi.Query(q, &dst); err != nil {
		return err
	}

	if len(dst) == 0 {
		// It's possible for these classes to legitimately return null when queried
		return nil
	}

	ch <- prometheus.MustNewConstMetric(
		c.sessionAutomaticReconnectCount,
		prometheus.CounterValue,
		float64(dst[0].AutomaticReconnectCount),
	)

	ch <- prometheus.MustNewConstMetric(
		c.sessionCumulativeReceivedBytesOverTCP,
		prometheus.CounterValue,
		float64(dst[0].CumulativeReceivedBytesoverTCP),
	)

	ch <- prometheus.MustNewConstMetric(
		c.sessionCumulativeReceivedBytesOverUDP,
		prometheus.CounterValue,
		float64(dst[0].CumulativeReceivedBytesoverUDP),
	)

	ch <- prometheus.MustNewConstMetric(
		c.sessionCumulativeTransmittedBytesOverTCP,
		prometheus.CounterValue,
		float64(dst[0].CumulativeTransmittedBytesoverTCP),
	)

	ch <- prometheus.MustNewConstMetric(
		c.sessionCumulativeTransmittedBytesOverUDP,
		prometheus.CounterValue,
		float64(dst[0].CumulativeTransmittedBytesoverUDP),
	)

	ch <- prometheus.MustNewConstMetric(
		c.sessionEstimatedBandwidthUplink,
		prometheus.GaugeValue,
		float64(dst[0].EstimatedBandwidthUplink),
	)

	ch <- prometheus.MustNewConstMetric(
		c.sessionInstantaneousReceivedBytesOverTCP,
		prometheus.CounterValue,
		float64(dst[0].InstantaneousReceivedBytesoverTCP),
	)

	ch <- prometheus.MustNewConstMetric(
		c.sessionInstantaneousReceivedBytesOverUDP,
		prometheus.CounterValue,
		float64(dst[0].InstantaneousReceivedBytesoverUDP),
	)

	ch <- prometheus.MustNewConstMetric(
		c.sessionInstantaneousTransmittedBytesOverTCP,
		prometheus.CounterValue,
		float64(dst[0].InstantaneousTransmittedBytesoverTCP),
	)

	ch <- prometheus.MustNewConstMetric(
		c.sessionInstantaneousTransmittedBytesOverUDP,
		prometheus.CounterValue,
		float64(dst[0].InstantaneousTransmittedBytesoverUDP),
	)

	ch <- prometheus.MustNewConstMetric(
		c.sessionJitterUplink,
		prometheus.GaugeValue,
		float64(dst[0].JitterUplink),
	)

	ch <- prometheus.MustNewConstMetric(
		c.sessionPacketLossUplink,
		prometheus.GaugeValue,
		float64(dst[0].PacketLossUplink),
	)

	ch <- prometheus.MustNewConstMetric(
		c.sessionReceivedBytes,
		prometheus.CounterValue,
		float64(dst[0].ReceivedBytes),
	)

	ch <- prometheus.MustNewConstMetric(
		c.sessionReceivedPackets,
		prometheus.CounterValue,
		float64(dst[0].ReceivedPackets),
	)

	ch <- prometheus.MustNewConstMetric(
		c.sessionRTT,
		prometheus.GaugeValue,
		float64(dst[0].RTT),
	)

	ch <- prometheus.MustNewConstMetric(
		c.sessionTransmittedBytes,
		prometheus.CounterValue,
		float64(dst[0].TransmittedBytes),
	)

	ch <- prometheus.MustNewConstMetric(
		c.sessionTransmittedPackets,
		prometheus.CounterValue,
		float64(dst[0].TransmittedPackets),
	)

	return nil
}

func (c *Collector) collectSkypeforBusinessControl(ch chan<- prometheus.Metric) error {
	var dst []win32_PerfRawData_Counters_VMwareBlastSkypeforBusinessControlCounters
	q := wmi.QueryAll(&dst, c.logger)
	if err := wmi.Query(q, &dst); err != nil {
		return err
	}

	if len(dst) == 0 {
		// It's possible for these classes to legitimately return null when queried
		return nil
	}

	ch <- prometheus.MustNewConstMetric(
		c.skypeForBusinessControlReceivedBytes,
		prometheus.CounterValue,
		float64(dst[0].ReceivedBytes),
	)

	ch <- prometheus.MustNewConstMetric(
		c.skypeForBusinessControlReceivedPackets,
		prometheus.CounterValue,
		float64(dst[0].ReceivedPackets),
	)

	ch <- prometheus.MustNewConstMetric(
		c.skypeForBusinessControlTransmittedBytes,
		prometheus.CounterValue,
		float64(dst[0].TransmittedBytes),
	)

	ch <- prometheus.MustNewConstMetric(
		c.skypeForBusinessControlTransmittedPackets,
		prometheus.CounterValue,
		float64(dst[0].TransmittedPackets),
	)

	return nil
}

func (c *Collector) collectThinPrint(ch chan<- prometheus.Metric) error {
	var dst []win32_PerfRawData_Counters_VMwareBlastThinPrintCounters
	q := wmi.QueryAll(&dst, c.logger)
	if err := wmi.Query(q, &dst); err != nil {
		return err
	}

	if len(dst) == 0 {
		// It's possible for these classes to legitimately return null when queried
		return nil
	}

	ch <- prometheus.MustNewConstMetric(
		c.thinPrintReceivedBytes,
		prometheus.CounterValue,
		float64(dst[0].ReceivedBytes),
	)

	ch <- prometheus.MustNewConstMetric(
		c.thinPrintReceivedPackets,
		prometheus.CounterValue,
		float64(dst[0].ReceivedPackets),
	)

	ch <- prometheus.MustNewConstMetric(
		c.thinPrintTransmittedBytes,
		prometheus.CounterValue,
		float64(dst[0].TransmittedBytes),
	)

	ch <- prometheus.MustNewConstMetric(
		c.thinPrintTransmittedPackets,
		prometheus.CounterValue,
		float64(dst[0].TransmittedPackets),
	)

	return nil
}

func (c *Collector) collectUsb(ch chan<- prometheus.Metric) error {
	var dst []win32_PerfRawData_Counters_VMwareBlastUSBCounters
	q := wmi.QueryAll(&dst, c.logger)
	if err := wmi.Query(q, &dst); err != nil {
		return err
	}

	if len(dst) == 0 {
		// It's possible for these classes to legitimately return null when queried
		return nil
	}

	ch <- prometheus.MustNewConstMetric(
		c.usbReceivedBytes,
		prometheus.CounterValue,
		float64(dst[0].ReceivedBytes),
	)

	ch <- prometheus.MustNewConstMetric(
		c.usbReceivedPackets,
		prometheus.CounterValue,
		float64(dst[0].ReceivedPackets),
	)

	ch <- prometheus.MustNewConstMetric(
		c.usbTransmittedBytes,
		prometheus.CounterValue,
		float64(dst[0].TransmittedBytes),
	)

	ch <- prometheus.MustNewConstMetric(
		c.usbTransmittedPackets,
		prometheus.CounterValue,
		float64(dst[0].TransmittedPackets),
	)

	return nil
}

func (c *Collector) collectWindowsMediaMmr(ch chan<- prometheus.Metric) error {
	var dst []win32_PerfRawData_Counters_VMwareBlastWindowsMediaMMRCounters
	q := wmi.QueryAll(&dst, c.logger)
	if err := wmi.Query(q, &dst); err != nil {
		return err
	}

	if len(dst) == 0 {
		// It's possible for these classes to legitimately return null when queried
		return nil
	}

	ch <- prometheus.MustNewConstMetric(
		c.windowsMediaMMRReceivedBytes,
		prometheus.CounterValue,
		float64(dst[0].ReceivedBytes),
	)

	ch <- prometheus.MustNewConstMetric(
		c.windowsMediaMMRReceivedPackets,
		prometheus.CounterValue,
		float64(dst[0].ReceivedPackets),
	)

	ch <- prometheus.MustNewConstMetric(
		c.windowsMediaMMRTransmittedBytes,
		prometheus.CounterValue,
		float64(dst[0].TransmittedBytes),
	)

	ch <- prometheus.MustNewConstMetric(
		c.windowsMediaMMRTransmittedPackets,
		prometheus.CounterValue,
		float64(dst[0].TransmittedPackets),
	)

	return nil
}
