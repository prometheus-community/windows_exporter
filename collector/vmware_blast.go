//go:build windows
// +build windows

package collector

import (
	"github.com/StackExchange/wmi"
	"github.com/prometheus-community/windows_exporter/log"
	"github.com/prometheus/client_golang/prometheus"
)

func init() {
	registerCollector("vmware_blast", NewVmwareBlastCollector)
}

// A VmwareBlastCollector is a Prometheus collector for WMI metrics:
// Win32_PerfRawData_Counters_VMwareBlastAudioCounters
// Win32_PerfRawData_Counters_VMwareBlastCDRCounters
// Win32_PerfRawData_Counters_VMwareBlastClipboardCounters
// Win32_PerfRawData_Counters_VMwareBlastHTML5MMRCounters
// Win32_PerfRawData_Counters_VMwareBlastImagingCounters
// Win32_PerfRawData_Counters_VMwareBlastRTAVCounters
// Win32_PerfRawData_Counters_VMwareBlastSerialPortandScannerCounters
// Win32_PerfRawData_Counters_VMwareBlastSessionCounters
// Win32_PerfRawData_Counters_VMwareBlastSkypeforBusinessControlCounters
// Win32_PerfRawData_Counters_VMwareBlastThinPrintCounters
// Win32_PerfRawData_Counters_VMwareBlastUSBCounters
// Win32_PerfRawData_Counters_VMwareBlastWindowsMediaMMRCounters

type VmwareBlastCollector struct {
	AudioReceivedBytes      *prometheus.Desc
	AudioReceivedPackets    *prometheus.Desc
	AudioTransmittedBytes   *prometheus.Desc
	AudioTransmittedPackets *prometheus.Desc

	CDRReceivedBytes      *prometheus.Desc
	CDRReceivedPackets    *prometheus.Desc
	CDRTransmittedBytes   *prometheus.Desc
	CDRTransmittedPackets *prometheus.Desc

	ClipboardReceivedBytes      *prometheus.Desc
	ClipboardReceivedPackets    *prometheus.Desc
	ClipboardTransmittedBytes   *prometheus.Desc
	ClipboardTransmittedPackets *prometheus.Desc

	HTML5MMRReceivedBytes      *prometheus.Desc
	HTML5MMRReceivedPackets    *prometheus.Desc
	HTML5MMRTransmittedBytes   *prometheus.Desc
	HTML5MMRTransmittedPackets *prometheus.Desc

	ImagingDirtyFramesPerSecond *prometheus.Desc
	ImagingFBCRate              *prometheus.Desc
	ImagingFramesPerSecond      *prometheus.Desc
	ImagingPollRate             *prometheus.Desc
	ImagingReceivedBytes        *prometheus.Desc
	ImagingReceivedPackets      *prometheus.Desc
	ImagingTotalDirtyFrames     *prometheus.Desc
	ImagingTotalFBC             *prometheus.Desc
	ImagingTotalFrames          *prometheus.Desc
	ImagingTotalPoll            *prometheus.Desc
	ImagingTransmittedBytes     *prometheus.Desc
	ImagingTransmittedPackets   *prometheus.Desc

	RTAVReceivedBytes      *prometheus.Desc
	RTAVReceivedPackets    *prometheus.Desc
	RTAVTransmittedBytes   *prometheus.Desc
	RTAVTransmittedPackets *prometheus.Desc

	SerialPortandScannerReceivedBytes      *prometheus.Desc
	SerialPortandScannerReceivedPackets    *prometheus.Desc
	SerialPortandScannerTransmittedBytes   *prometheus.Desc
	SerialPortandScannerTransmittedPackets *prometheus.Desc

	SessionAutomaticReconnectCount              *prometheus.Desc
	SessionCumulativeReceivedBytesOverTCP       *prometheus.Desc
	SessionCumulativeReceivedBytesOverUDP       *prometheus.Desc
	SessionCumulativeTransmittedBytesOverTCP    *prometheus.Desc
	SessionCumulativeTransmittedBytesOverUDP    *prometheus.Desc
	SessionEstimatedBandwidthUplink             *prometheus.Desc
	SessionInstantaneousReceivedBytesOverTCP    *prometheus.Desc
	SessionInstantaneousReceivedBytesOverUDP    *prometheus.Desc
	SessionInstantaneousTransmittedBytesOverTCP *prometheus.Desc
	SessionInstantaneousTransmittedBytesOverUDP *prometheus.Desc
	SessionJitterUplink                         *prometheus.Desc
	SessionPacketLossUplink                     *prometheus.Desc
	SessionReceivedBytes                        *prometheus.Desc
	SessionReceivedPackets                      *prometheus.Desc
	SessionRTT                                  *prometheus.Desc
	SessionTransmittedBytes                     *prometheus.Desc
	SessionTransmittedPackets                   *prometheus.Desc

	SkypeforBusinessControlReceivedBytes      *prometheus.Desc
	SkypeforBusinessControlReceivedPackets    *prometheus.Desc
	SkypeforBusinessControlTransmittedBytes   *prometheus.Desc
	SkypeforBusinessControlTransmittedPackets *prometheus.Desc

	ThinPrintReceivedBytes      *prometheus.Desc
	ThinPrintReceivedPackets    *prometheus.Desc
	ThinPrintTransmittedBytes   *prometheus.Desc
	ThinPrintTransmittedPackets *prometheus.Desc

	USBReceivedBytes      *prometheus.Desc
	USBReceivedPackets    *prometheus.Desc
	USBTransmittedBytes   *prometheus.Desc
	USBTransmittedPackets *prometheus.Desc

	WindowsMediaMMRReceivedBytes      *prometheus.Desc
	WindowsMediaMMRReceivedPackets    *prometheus.Desc
	WindowsMediaMMRTransmittedBytes   *prometheus.Desc
	WindowsMediaMMRTransmittedPackets *prometheus.Desc
}

// NewVmwareBlastCollector constructs a new VmwareBlastCollector
func NewVmwareBlastCollector() (Collector, error) {
	const subsystem = "vmware_blast"
	return &VmwareBlastCollector{
		AudioReceivedBytes: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "audio_received_bytes"),
			"(AudioReceivedBytes)",
			nil,
			nil,
		),
		AudioReceivedPackets: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "audio_received_packets"),
			"(AudioReceivedPackets)",
			nil,
			nil,
		),
		AudioTransmittedBytes: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "audio_transmitted_bytes"),
			"(AudioTransmittedBytes)",
			nil,
			nil,
		),
		AudioTransmittedPackets: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "audio_transmitted_packets"),
			"(AudioTransmittedPackets)",
			nil,
			nil,
		),

		CDRReceivedBytes: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "cdr_received_bytes"),
			"(CDRReceivedBytes)",
			nil,
			nil,
		),
		CDRReceivedPackets: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "cdr_received_packets"),
			"(CDRReceivedPackets)",
			nil,
			nil,
		),
		CDRTransmittedBytes: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "cdr_transmitted_bytes"),
			"(CDRTransmittedBytes)",
			nil,
			nil,
		),
		CDRTransmittedPackets: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "cdr_transmitted_packets"),
			"(CDRTransmittedPackets)",
			nil,
			nil,
		),

		ClipboardReceivedBytes: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "clipboard_received_bytes"),
			"(ClipboardReceivedBytes)",
			nil,
			nil,
		),
		ClipboardReceivedPackets: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "clipboard_received_packets"),
			"(ClipboardReceivedPackets)",
			nil,
			nil,
		),
		ClipboardTransmittedBytes: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "clipboard_transmitted_bytes"),
			"(ClipboardTransmittedBytes)",
			nil,
			nil,
		),
		ClipboardTransmittedPackets: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "clipboard_transmitted_packets"),
			"(ClipboardTransmittedPackets)",
			nil,
			nil,
		),

		HTML5MMRReceivedBytes: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "html5_mmr_received_bytes"),
			"(HTML5MMRReceivedBytes)",
			nil,
			nil,
		),
		HTML5MMRReceivedPackets: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "html5_mmr_received_packets"),
			"(HTML5MMRReceivedPackets)",
			nil,
			nil,
		),
		HTML5MMRTransmittedBytes: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "html5_mmr_transmitted_bytes"),
			"(HTML5MMRTransmittedBytes)",
			nil,
			nil,
		),
		HTML5MMRTransmittedPackets: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "html5_mmr_transmitted_packets"),
			"(HTML5MMRTransmittedPackets)",
			nil,
			nil,
		),

		ImagingDirtyFramesPerSecond: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "imaging_dirty_frames_per_second"),
			"(ImagingDirtyFramesPerSecond)",
			nil,
			nil,
		),
		ImagingFBCRate: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "imaging_fbc_rate"),
			"(ImagingFBCRate)",
			nil,
			nil,
		),
		ImagingFramesPerSecond: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "imaging_frames_per_second"),
			"(ImagingFramesPerSecond)",
			nil,
			nil,
		),
		ImagingPollRate: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "imaging_poll_rate"),
			"(ImagingPollRate)",
			nil,
			nil,
		),
		ImagingReceivedBytes: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "imaging_received_bytes"),
			"(ImagingReceivedBytes)",
			nil,
			nil,
		),
		ImagingReceivedPackets: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "imaging_received_packets"),
			"(ImagingReceivedPackets)",
			nil,
			nil,
		),
		ImagingTotalDirtyFrames: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "imaging_total_dirty_frames"),
			"(ImagingTotalDirtyFrames)",
			nil,
			nil,
		),
		ImagingTotalFBC: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "imaging_total_fbc"),
			"(ImagingTotalFBC)",
			nil,
			nil,
		),
		ImagingTotalFrames: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "imaging_total_frames"),
			"(ImagingTotalFrames)",
			nil,
			nil,
		),
		ImagingTotalPoll: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "imaging_total_poll"),
			"(ImagingTotalPoll)",
			nil,
			nil,
		),
		ImagingTransmittedBytes: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "imaging_transmitted_bytes"),
			"(ImagingTransmittedBytes)",
			nil,
			nil,
		),
		ImagingTransmittedPackets: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "imaging_transmitted_packets"),
			"(ImagingTransmittedPackets)",
			nil,
			nil,
		),

		RTAVReceivedBytes: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "rtav_received_bytes"),
			"(RTAVReceivedBytes)",
			nil,
			nil,
		),
		RTAVReceivedPackets: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "rtav_received_packets"),
			"(RTAVReceivedPackets)",
			nil,
			nil,
		),
		RTAVTransmittedBytes: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "rtav_transmitted_bytes"),
			"(RTAVTransmittedBytes)",
			nil,
			nil,
		),
		RTAVTransmittedPackets: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "rtav_transmitted_packets"),
			"(RTAVTransmittedPackets)",
			nil,
			nil,
		),

		SerialPortandScannerReceivedBytes: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "serial_port_and_scanner_received_bytes"),
			"(SerialPortandScannerReceivedBytes)",
			nil,
			nil,
		),
		SerialPortandScannerReceivedPackets: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "serial_port_and_scanner_received_packets"),
			"(SerialPortandScannerReceivedPackets)",
			nil,
			nil,
		),
		SerialPortandScannerTransmittedBytes: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "serial_port_and_scanner_transmitted_bytes"),
			"(SerialPortandScannerTransmittedBytes)",
			nil,
			nil,
		),
		SerialPortandScannerTransmittedPackets: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "serial_port_and_scanner_transmitted_packets"),
			"(SerialPortandScannerTransmittedPackets)",
			nil,
			nil,
		),

		SessionAutomaticReconnectCount: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "session_automatic_reconnect_count"),
			"(SessionAutomaticReconnectCount)",
			nil,
			nil,
		),
		SessionCumulativeReceivedBytesOverTCP: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "session_cumlative_received_bytes_over_tcp"),
			"(SessionCumulativeReceivedBytesOverTCP)",
			nil,
			nil,
		),
		SessionCumulativeReceivedBytesOverUDP: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "session_cumlative_received_bytes_over_udp"),
			"(SessionCumulativeReceivedBytesOverUDP)",
			nil,
			nil,
		),
		SessionCumulativeTransmittedBytesOverTCP: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "session_cumlative_transmitted_bytes_over_tcp"),
			"(SessionCumulativeTransmittedBytesOverTCP)",
			nil,
			nil,
		),
		SessionCumulativeTransmittedBytesOverUDP: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "session_cumlative_transmitted_bytes_over_udp"),
			"(SessionCumulativeTransmittedBytesOverUDP)",
			nil,
			nil,
		),
		SessionEstimatedBandwidthUplink: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "session_estimated_bandwidth_uplink"),
			"(SessionEstimatedBandwidthUplink)",
			nil,
			nil,
		),
		SessionInstantaneousReceivedBytesOverTCP: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "session_instantaneous_received_bytes_over_tcp"),
			"(SessionInstantaneousReceivedBytesOverTCP)",
			nil,
			nil,
		),
		SessionInstantaneousReceivedBytesOverUDP: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "session_instantaneous_received_bytes_over_udp"),
			"(SessionInstantaneousReceivedBytesOverUDP)",
			nil,
			nil,
		),
		SessionInstantaneousTransmittedBytesOverTCP: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "session_instantaneous_transmitted_bytes_over_tcp"),
			"(SessionInstantaneousTransmittedBytesOverTCP)",
			nil,
			nil,
		),
		SessionInstantaneousTransmittedBytesOverUDP: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "session_instantaneous_transmitted_bytes_over_udp"),
			"(SessionInstantaneousTransmittedBytesOverUDP)",
			nil,
			nil,
		),
		SessionJitterUplink: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "session_jitter_uplink"),
			"(SessionJitterUplink)",
			nil,
			nil,
		),
		SessionPacketLossUplink: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "session_packet_loss_uplink"),
			"(SessionPacketLossUplink)",
			nil,
			nil,
		),
		SessionReceivedBytes: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "session_received_bytes"),
			"(SessionReceivedBytes)",
			nil,
			nil,
		),
		SessionReceivedPackets: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "session_received_packets"),
			"(SessionReceivedPackets)",
			nil,
			nil,
		),
		SessionRTT: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "session_rtt"),
			"(SessionRTT)",
			nil,
			nil,
		),
		SessionTransmittedBytes: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "session_transmitted_bytes"),
			"(SessionTransmittedBytes)",
			nil,
			nil,
		),
		SessionTransmittedPackets: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "session_transmitted_packets"),
			"(SessionTransmittedPackets)",
			nil,
			nil,
		),

		SkypeforBusinessControlReceivedBytes: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "skype_for_business_control_received_bytes"),
			"(SkypeforBusinessControlReceivedBytes)",
			nil,
			nil,
		),
		SkypeforBusinessControlReceivedPackets: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "skype_for_business_control_received_packets"),
			"(SkypeforBusinessControlReceivedPackets)",
			nil,
			nil,
		),
		SkypeforBusinessControlTransmittedBytes: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "skype_for_business_control_transmitted_bytes"),
			"(SkypeforBusinessControlTransmittedBytes)",
			nil,
			nil,
		),
		SkypeforBusinessControlTransmittedPackets: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "skype_for_business_control_transmitted_packets"),
			"(SkypeforBusinessControlTransmittedPackets)",
			nil,
			nil,
		),

		ThinPrintReceivedBytes: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "thinprint_received_bytes"),
			"(ThinPrintReceivedBytes)",
			nil,
			nil,
		),
		ThinPrintReceivedPackets: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "thinprint_received_packets"),
			"(ThinPrintReceivedPackets)",
			nil,
			nil,
		),
		ThinPrintTransmittedBytes: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "thinprint_transmitted_bytes"),
			"(ThinPrintTransmittedBytes)",
			nil,
			nil,
		),
		ThinPrintTransmittedPackets: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "thinprint_transmitted_packets"),
			"(ThinPrintTransmittedPackets)",
			nil,
			nil,
		),

		USBReceivedBytes: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "usb_received_bytes"),
			"(USBReceivedBytes)",
			nil,
			nil,
		),
		USBReceivedPackets: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "usb_received_packets"),
			"(USBReceivedPackets)",
			nil,
			nil,
		),
		USBTransmittedBytes: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "usb_transmitted_bytes"),
			"(USBTransmittedBytes)",
			nil,
			nil,
		),
		USBTransmittedPackets: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "usb_transmitted_packets"),
			"(USBTransmittedPackets)",
			nil,
			nil,
		),

		WindowsMediaMMRReceivedBytes: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "windows_media_mmr_received_bytes"),
			"(WindowsMediaMMRReceivedBytes)",
			nil,
			nil,
		),
		WindowsMediaMMRReceivedPackets: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "windows_media_mmr_received_packets"),
			"(WindowsMediaMMRReceivedPackets)",
			nil,
			nil,
		),
		WindowsMediaMMRTransmittedBytes: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "windows_media_mmr_transmitted_bytes"),
			"(WindowsMediaMMRTransmittedBytes)",
			nil,
			nil,
		),
		WindowsMediaMMRTransmittedPackets: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "windows_media_mmr_transmitted_packets"),
			"(WindowsMediaMMRTransmittedPackets)",
			nil,
			nil,
		),
	}, nil
}

// Collect sends the metric values for each metric
// to the provided prometheus Metric channel.
func (c *VmwareBlastCollector) Collect(ctx *ScrapeContext, ch chan<- prometheus.Metric) error {
	if desc, err := c.collectAudio(ch); err != nil {
		log.Error("failed collecting vmware blast audio metrics:", desc, err)
		return err
	}
	if desc, err := c.collectCdr(ch); err != nil {
		log.Error("failed collecting vmware blast CDR metrics:", desc, err)
		return err
	}
	if desc, err := c.collectClipboard(ch); err != nil {
		log.Error("failed collecting vmware blast clipboard metrics:", desc, err)
		return err
	}
	if desc, err := c.collectHtml5Mmr(ch); err != nil {
		log.Error("failed collecting vmware blast HTML5 MMR metrics:", desc, err)
		return err
	}
	if desc, err := c.collectImaging(ch); err != nil {
		log.Error("failed collecting vmware blast imaging metrics:", desc, err)
		return err
	}
	if desc, err := c.collectRtav(ch); err != nil {
		log.Error("failed collecting vmware blast RTAV metrics:", desc, err)
		return err
	}
	if desc, err := c.collectSerialPortandScanner(ch); err != nil {
		log.Error("failed collecting vmware blast serial port and scanner metrics:", desc, err)
		return err
	}
	if desc, err := c.collectSession(ch); err != nil {
		log.Error("failed collecting vmware blast metrics:", desc, err)
		return err
	}
	if desc, err := c.collectSkypeforBusinessControl(ch); err != nil {
		log.Error("failed collecting vmware blast skype for business control metrics:", desc, err)
		return err
	}
	if desc, err := c.collectThinPrint(ch); err != nil {
		log.Error("failed collecting vmware blast thin print metrics:", desc, err)
		return err
	}
	if desc, err := c.collectUsb(ch); err != nil {
		log.Error("failed collecting vmware blast USB metrics:", desc, err)
		return err
	}
	if desc, err := c.collectWindowsMediaMmr(ch); err != nil {
		log.Error("failed collecting vmware blast windows media MMR metrics:", desc, err)
		return err
	}
	return nil
}

type Win32_PerfRawData_Counters_VMwareBlastAudioCounters struct {
	ReceivedBytes      uint32
	ReceivedPackets    uint32
	TransmittedBytes   uint32
	TransmittedPackets uint32
}

type Win32_PerfRawData_Counters_VMwareBlastCDRCounters struct {
	ReceivedBytes      uint32
	ReceivedPackets    uint32
	TransmittedBytes   uint32
	TransmittedPackets uint32
}

type Win32_PerfRawData_Counters_VMwareBlastClipboardCounters struct {
	ReceivedBytes      uint32
	ReceivedPackets    uint32
	TransmittedBytes   uint32
	TransmittedPackets uint32
}

type Win32_PerfRawData_Counters_VMwareBlastHTML5MMRcounters struct {
	ReceivedBytes      uint32
	ReceivedPackets    uint32
	TransmittedBytes   uint32
	TransmittedPackets uint32
}

type Win32_PerfRawData_Counters_VMwareBlastImagingCounters struct {
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

type Win32_PerfRawData_Counters_VMwareBlastRTAVCounters struct {
	ReceivedBytes      uint32
	ReceivedPackets    uint32
	TransmittedBytes   uint32
	TransmittedPackets uint32
}

type Win32_PerfRawData_Counters_VMwareBlastSerialPortandScannerCounters struct {
	ReceivedBytes      uint32
	ReceivedPackets    uint32
	TransmittedBytes   uint32
	TransmittedPackets uint32
}

type Win32_PerfRawData_Counters_VMwareBlastSessionCounters struct {
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

type Win32_PerfRawData_Counters_VMwareBlastSkypeforBusinessControlCounters struct {
	ReceivedBytes      uint32
	ReceivedPackets    uint32
	TransmittedBytes   uint32
	TransmittedPackets uint32
}

type Win32_PerfRawData_Counters_VMwareBlastThinPrintCounters struct {
	ReceivedBytes      uint32
	ReceivedPackets    uint32
	TransmittedBytes   uint32
	TransmittedPackets uint32
}

type Win32_PerfRawData_Counters_VMwareBlastUSBCounters struct {
	ReceivedBytes      uint32
	ReceivedPackets    uint32
	TransmittedBytes   uint32
	TransmittedPackets uint32
}

type Win32_PerfRawData_Counters_VMwareBlastWindowsMediaMMRCounters struct {
	ReceivedBytes      uint32
	ReceivedPackets    uint32
	TransmittedBytes   uint32
	TransmittedPackets uint32
}

func (c *VmwareBlastCollector) collectAudio(ch chan<- prometheus.Metric) (*prometheus.Desc, error) {
	var dst []Win32_PerfRawData_Counters_VMwareBlastAudioCounters
	q := queryAll(&dst)
	if err := wmi.Query(q, &dst); err != nil {
		return nil, err
	}

	if len(dst) == 0 {
		// It's possible for these classes to legitimately return null when queried
		return nil, nil
	}

	ch <- prometheus.MustNewConstMetric(
		c.AudioReceivedBytes,
		prometheus.CounterValue,
		float64(dst[0].ReceivedBytes),
	)

	ch <- prometheus.MustNewConstMetric(
		c.AudioReceivedPackets,
		prometheus.CounterValue,
		float64(dst[0].ReceivedPackets),
	)

	ch <- prometheus.MustNewConstMetric(
		c.AudioTransmittedBytes,
		prometheus.CounterValue,
		float64(dst[0].TransmittedBytes),
	)

	ch <- prometheus.MustNewConstMetric(
		c.AudioTransmittedPackets,
		prometheus.CounterValue,
		float64(dst[0].TransmittedPackets),
	)

	return nil, nil
}

func (c *VmwareBlastCollector) collectCdr(ch chan<- prometheus.Metric) (*prometheus.Desc, error) {
	var dst []Win32_PerfRawData_Counters_VMwareBlastCDRCounters
	q := queryAll(&dst)
	if err := wmi.Query(q, &dst); err != nil {
		return nil, err
	}

	if len(dst) == 0 {
		// It's possible for these classes to legitimately return null when queried
		return nil, nil
	}

	ch <- prometheus.MustNewConstMetric(
		c.CDRReceivedBytes,
		prometheus.CounterValue,
		float64(dst[0].ReceivedBytes),
	)

	ch <- prometheus.MustNewConstMetric(
		c.CDRReceivedPackets,
		prometheus.CounterValue,
		float64(dst[0].ReceivedPackets),
	)

	ch <- prometheus.MustNewConstMetric(
		c.CDRTransmittedBytes,
		prometheus.CounterValue,
		float64(dst[0].TransmittedBytes),
	)

	ch <- prometheus.MustNewConstMetric(
		c.CDRTransmittedPackets,
		prometheus.CounterValue,
		float64(dst[0].TransmittedPackets),
	)

	return nil, nil
}

func (c *VmwareBlastCollector) collectClipboard(ch chan<- prometheus.Metric) (*prometheus.Desc, error) {
	var dst []Win32_PerfRawData_Counters_VMwareBlastClipboardCounters
	q := queryAll(&dst)
	if err := wmi.Query(q, &dst); err != nil {
		return nil, err
	}

	if len(dst) == 0 {
		// It's possible for these classes to legitimately return null when queried
		return nil, nil
	}

	ch <- prometheus.MustNewConstMetric(
		c.ClipboardReceivedBytes,
		prometheus.CounterValue,
		float64(dst[0].ReceivedBytes),
	)

	ch <- prometheus.MustNewConstMetric(
		c.ClipboardReceivedPackets,
		prometheus.CounterValue,
		float64(dst[0].ReceivedPackets),
	)

	ch <- prometheus.MustNewConstMetric(
		c.ClipboardTransmittedBytes,
		prometheus.CounterValue,
		float64(dst[0].TransmittedBytes),
	)

	ch <- prometheus.MustNewConstMetric(
		c.ClipboardTransmittedPackets,
		prometheus.CounterValue,
		float64(dst[0].TransmittedPackets),
	)

	return nil, nil
}

func (c *VmwareBlastCollector) collectHtml5Mmr(ch chan<- prometheus.Metric) (*prometheus.Desc, error) {
	var dst []Win32_PerfRawData_Counters_VMwareBlastHTML5MMRcounters
	q := queryAll(&dst)
	if err := wmi.Query(q, &dst); err != nil {
		return nil, err
	}

	if len(dst) == 0 {
		// It's possible for these classes to legitimately return null when queried
		return nil, nil
	}

	ch <- prometheus.MustNewConstMetric(
		c.HTML5MMRReceivedBytes,
		prometheus.CounterValue,
		float64(dst[0].ReceivedBytes),
	)

	ch <- prometheus.MustNewConstMetric(
		c.HTML5MMRReceivedPackets,
		prometheus.CounterValue,
		float64(dst[0].ReceivedPackets),
	)

	ch <- prometheus.MustNewConstMetric(
		c.HTML5MMRTransmittedBytes,
		prometheus.CounterValue,
		float64(dst[0].TransmittedBytes),
	)

	ch <- prometheus.MustNewConstMetric(
		c.HTML5MMRTransmittedPackets,
		prometheus.CounterValue,
		float64(dst[0].TransmittedPackets),
	)

	return nil, nil
}

func (c *VmwareBlastCollector) collectImaging(ch chan<- prometheus.Metric) (*prometheus.Desc, error) {
	var dst []Win32_PerfRawData_Counters_VMwareBlastImagingCounters
	q := queryAll(&dst)
	if err := wmi.Query(q, &dst); err != nil {
		return nil, err
	}

	if len(dst) == 0 {
		// It's possible for these classes to legitimately return null when queried
		return nil, nil
	}

	ch <- prometheus.MustNewConstMetric(
		c.ImagingDirtyFramesPerSecond,
		prometheus.GaugeValue,
		float64(dst[0].Dirtyframespersecond),
	)

	ch <- prometheus.MustNewConstMetric(
		c.ImagingFBCRate,
		prometheus.GaugeValue,
		float64(dst[0].FBCRate),
	)

	ch <- prometheus.MustNewConstMetric(
		c.ImagingFramesPerSecond,
		prometheus.GaugeValue,
		float64(dst[0].Framespersecond),
	)

	ch <- prometheus.MustNewConstMetric(
		c.ImagingPollRate,
		prometheus.GaugeValue,
		float64(dst[0].PollRate),
	)

	ch <- prometheus.MustNewConstMetric(
		c.ImagingReceivedBytes,
		prometheus.CounterValue,
		float64(dst[0].ReceivedBytes),
	)

	ch <- prometheus.MustNewConstMetric(
		c.ImagingReceivedPackets,
		prometheus.CounterValue,
		float64(dst[0].ReceivedPackets),
	)

	ch <- prometheus.MustNewConstMetric(
		c.ImagingTotalDirtyFrames,
		prometheus.CounterValue,
		float64(dst[0].Totaldirtyframes),
	)

	ch <- prometheus.MustNewConstMetric(
		c.ImagingTotalFBC,
		prometheus.CounterValue,
		float64(dst[0].TotalFBC),
	)

	ch <- prometheus.MustNewConstMetric(
		c.ImagingTotalFrames,
		prometheus.CounterValue,
		float64(dst[0].Totalframes),
	)

	ch <- prometheus.MustNewConstMetric(
		c.ImagingTotalPoll,
		prometheus.CounterValue,
		float64(dst[0].Totalpoll),
	)

	ch <- prometheus.MustNewConstMetric(
		c.ImagingTransmittedBytes,
		prometheus.CounterValue,
		float64(dst[0].TransmittedBytes),
	)

	ch <- prometheus.MustNewConstMetric(
		c.ImagingTransmittedPackets,
		prometheus.CounterValue,
		float64(dst[0].TransmittedPackets),
	)

	return nil, nil
}

func (c *VmwareBlastCollector) collectRtav(ch chan<- prometheus.Metric) (*prometheus.Desc, error) {
	var dst []Win32_PerfRawData_Counters_VMwareBlastRTAVCounters
	q := queryAll(&dst)
	if err := wmi.Query(q, &dst); err != nil {
		return nil, err
	}

	if len(dst) == 0 {
		// It's possible for these classes to legitimately return null when queried
		return nil, nil
	}

	ch <- prometheus.MustNewConstMetric(
		c.RTAVReceivedBytes,
		prometheus.CounterValue,
		float64(dst[0].ReceivedBytes),
	)

	ch <- prometheus.MustNewConstMetric(
		c.RTAVReceivedPackets,
		prometheus.CounterValue,
		float64(dst[0].ReceivedPackets),
	)

	ch <- prometheus.MustNewConstMetric(
		c.RTAVTransmittedBytes,
		prometheus.CounterValue,
		float64(dst[0].TransmittedBytes),
	)

	ch <- prometheus.MustNewConstMetric(
		c.RTAVTransmittedPackets,
		prometheus.CounterValue,
		float64(dst[0].TransmittedPackets),
	)

	return nil, nil
}

func (c *VmwareBlastCollector) collectSerialPortandScanner(ch chan<- prometheus.Metric) (*prometheus.Desc, error) {
	var dst []Win32_PerfRawData_Counters_VMwareBlastSerialPortandScannerCounters
	q := queryAll(&dst)
	if err := wmi.Query(q, &dst); err != nil {
		return nil, err
	}

	if len(dst) == 0 {
		// It's possible for these classes to legitimately return null when queried
		return nil, nil
	}

	ch <- prometheus.MustNewConstMetric(
		c.SerialPortandScannerReceivedBytes,
		prometheus.CounterValue,
		float64(dst[0].ReceivedBytes),
	)

	ch <- prometheus.MustNewConstMetric(
		c.SerialPortandScannerReceivedPackets,
		prometheus.CounterValue,
		float64(dst[0].ReceivedPackets),
	)

	ch <- prometheus.MustNewConstMetric(
		c.SerialPortandScannerTransmittedBytes,
		prometheus.CounterValue,
		float64(dst[0].TransmittedBytes),
	)

	ch <- prometheus.MustNewConstMetric(
		c.SerialPortandScannerTransmittedPackets,
		prometheus.CounterValue,
		float64(dst[0].TransmittedPackets),
	)

	return nil, nil
}

func (c *VmwareBlastCollector) collectSession(ch chan<- prometheus.Metric) (*prometheus.Desc, error) {
	var dst []Win32_PerfRawData_Counters_VMwareBlastSessionCounters
	q := queryAll(&dst)
	if err := wmi.Query(q, &dst); err != nil {
		return nil, err
	}

	if len(dst) == 0 {
		// It's possible for these classes to legitimately return null when queried
		return nil, nil
	}

	ch <- prometheus.MustNewConstMetric(
		c.SessionAutomaticReconnectCount,
		prometheus.CounterValue,
		float64(dst[0].AutomaticReconnectCount),
	)

	ch <- prometheus.MustNewConstMetric(
		c.SessionCumulativeReceivedBytesOverTCP,
		prometheus.CounterValue,
		float64(dst[0].CumulativeReceivedBytesoverTCP),
	)

	ch <- prometheus.MustNewConstMetric(
		c.SessionCumulativeReceivedBytesOverUDP,
		prometheus.CounterValue,
		float64(dst[0].CumulativeReceivedBytesoverUDP),
	)

	ch <- prometheus.MustNewConstMetric(
		c.SessionCumulativeTransmittedBytesOverTCP,
		prometheus.CounterValue,
		float64(dst[0].CumulativeTransmittedBytesoverTCP),
	)

	ch <- prometheus.MustNewConstMetric(
		c.SessionCumulativeTransmittedBytesOverUDP,
		prometheus.CounterValue,
		float64(dst[0].CumulativeTransmittedBytesoverUDP),
	)

	ch <- prometheus.MustNewConstMetric(
		c.SessionEstimatedBandwidthUplink,
		prometheus.GaugeValue,
		float64(dst[0].EstimatedBandwidthUplink),
	)

	ch <- prometheus.MustNewConstMetric(
		c.SessionInstantaneousReceivedBytesOverTCP,
		prometheus.CounterValue,
		float64(dst[0].InstantaneousReceivedBytesoverTCP),
	)

	ch <- prometheus.MustNewConstMetric(
		c.SessionInstantaneousReceivedBytesOverUDP,
		prometheus.CounterValue,
		float64(dst[0].InstantaneousReceivedBytesoverUDP),
	)

	ch <- prometheus.MustNewConstMetric(
		c.SessionInstantaneousTransmittedBytesOverTCP,
		prometheus.CounterValue,
		float64(dst[0].InstantaneousTransmittedBytesoverTCP),
	)

	ch <- prometheus.MustNewConstMetric(
		c.SessionInstantaneousTransmittedBytesOverUDP,
		prometheus.CounterValue,
		float64(dst[0].InstantaneousTransmittedBytesoverUDP),
	)

	ch <- prometheus.MustNewConstMetric(
		c.SessionJitterUplink,
		prometheus.GaugeValue,
		float64(dst[0].JitterUplink),
	)

	ch <- prometheus.MustNewConstMetric(
		c.SessionPacketLossUplink,
		prometheus.GaugeValue,
		float64(dst[0].PacketLossUplink),
	)

	ch <- prometheus.MustNewConstMetric(
		c.SessionReceivedBytes,
		prometheus.CounterValue,
		float64(dst[0].ReceivedBytes),
	)

	ch <- prometheus.MustNewConstMetric(
		c.SessionReceivedPackets,
		prometheus.CounterValue,
		float64(dst[0].ReceivedPackets),
	)

	ch <- prometheus.MustNewConstMetric(
		c.SessionRTT,
		prometheus.GaugeValue,
		float64(dst[0].RTT),
	)

	ch <- prometheus.MustNewConstMetric(
		c.SessionTransmittedBytes,
		prometheus.CounterValue,
		float64(dst[0].TransmittedBytes),
	)

	ch <- prometheus.MustNewConstMetric(
		c.SessionTransmittedPackets,
		prometheus.CounterValue,
		float64(dst[0].TransmittedPackets),
	)

	return nil, nil
}

func (c *VmwareBlastCollector) collectSkypeforBusinessControl(ch chan<- prometheus.Metric) (*prometheus.Desc, error) {
	var dst []Win32_PerfRawData_Counters_VMwareBlastSkypeforBusinessControlCounters
	q := queryAll(&dst)
	if err := wmi.Query(q, &dst); err != nil {
		return nil, err
	}

	if len(dst) == 0 {
		// It's possible for these classes to legitimately return null when queried
		return nil, nil
	}

	ch <- prometheus.MustNewConstMetric(
		c.SkypeforBusinessControlReceivedBytes,
		prometheus.CounterValue,
		float64(dst[0].ReceivedBytes),
	)

	ch <- prometheus.MustNewConstMetric(
		c.SkypeforBusinessControlReceivedPackets,
		prometheus.CounterValue,
		float64(dst[0].ReceivedPackets),
	)

	ch <- prometheus.MustNewConstMetric(
		c.SkypeforBusinessControlTransmittedBytes,
		prometheus.CounterValue,
		float64(dst[0].TransmittedBytes),
	)

	ch <- prometheus.MustNewConstMetric(
		c.SkypeforBusinessControlTransmittedPackets,
		prometheus.CounterValue,
		float64(dst[0].TransmittedPackets),
	)

	return nil, nil
}

func (c *VmwareBlastCollector) collectThinPrint(ch chan<- prometheus.Metric) (*prometheus.Desc, error) {
	var dst []Win32_PerfRawData_Counters_VMwareBlastThinPrintCounters
	q := queryAll(&dst)
	if err := wmi.Query(q, &dst); err != nil {
		return nil, err
	}

	if len(dst) == 0 {
		// It's possible for these classes to legitimately return null when queried
		return nil, nil
	}

	ch <- prometheus.MustNewConstMetric(
		c.ThinPrintReceivedBytes,
		prometheus.CounterValue,
		float64(dst[0].ReceivedBytes),
	)

	ch <- prometheus.MustNewConstMetric(
		c.ThinPrintReceivedPackets,
		prometheus.CounterValue,
		float64(dst[0].ReceivedPackets),
	)

	ch <- prometheus.MustNewConstMetric(
		c.ThinPrintTransmittedBytes,
		prometheus.CounterValue,
		float64(dst[0].TransmittedBytes),
	)

	ch <- prometheus.MustNewConstMetric(
		c.ThinPrintTransmittedPackets,
		prometheus.CounterValue,
		float64(dst[0].TransmittedPackets),
	)

	return nil, nil
}

func (c *VmwareBlastCollector) collectUsb(ch chan<- prometheus.Metric) (*prometheus.Desc, error) {
	var dst []Win32_PerfRawData_Counters_VMwareBlastUSBCounters
	q := queryAll(&dst)
	if err := wmi.Query(q, &dst); err != nil {
		return nil, err
	}

	if len(dst) == 0 {
		// It's possible for these classes to legitimately return null when queried
		return nil, nil
	}

	ch <- prometheus.MustNewConstMetric(
		c.USBReceivedBytes,
		prometheus.CounterValue,
		float64(dst[0].ReceivedBytes),
	)

	ch <- prometheus.MustNewConstMetric(
		c.USBReceivedPackets,
		prometheus.CounterValue,
		float64(dst[0].ReceivedPackets),
	)

	ch <- prometheus.MustNewConstMetric(
		c.USBTransmittedBytes,
		prometheus.CounterValue,
		float64(dst[0].TransmittedBytes),
	)

	ch <- prometheus.MustNewConstMetric(
		c.USBTransmittedPackets,
		prometheus.CounterValue,
		float64(dst[0].TransmittedPackets),
	)

	return nil, nil
}

func (c *VmwareBlastCollector) collectWindowsMediaMmr(ch chan<- prometheus.Metric) (*prometheus.Desc, error) {
	var dst []Win32_PerfRawData_Counters_VMwareBlastWindowsMediaMMRCounters
	q := queryAll(&dst)
	if err := wmi.Query(q, &dst); err != nil {
		return nil, err
	}

	if len(dst) == 0 {
		// It's possible for these classes to legitimately return null when queried
		return nil, nil
	}

	ch <- prometheus.MustNewConstMetric(
		c.WindowsMediaMMRReceivedBytes,
		prometheus.CounterValue,
		float64(dst[0].ReceivedBytes),
	)

	ch <- prometheus.MustNewConstMetric(
		c.WindowsMediaMMRReceivedPackets,
		prometheus.CounterValue,
		float64(dst[0].ReceivedPackets),
	)

	ch <- prometheus.MustNewConstMetric(
		c.WindowsMediaMMRTransmittedBytes,
		prometheus.CounterValue,
		float64(dst[0].TransmittedBytes),
	)

	ch <- prometheus.MustNewConstMetric(
		c.WindowsMediaMMRTransmittedPackets,
		prometheus.CounterValue,
		float64(dst[0].TransmittedPackets),
	)

	return nil, nil
}
