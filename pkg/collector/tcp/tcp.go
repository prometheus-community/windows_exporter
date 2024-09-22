//go:build windows

package tcp

import (
	"fmt"
	"log/slog"

	"github.com/alecthomas/kingpin/v2"
	"github.com/pkg/errors"
	"github.com/prometheus-community/windows_exporter/pkg/perfdata"
	"github.com/prometheus-community/windows_exporter/pkg/types"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/yusufpapurcu/wmi"
	"syscall"
	"unsafe"
)

const Name = "tcp"

type Config struct{}

var (
	modiphlpapi             = syscall.NewLazyDLL("iphlpapi.dll")
	procGetExtendedTcpTable = modiphlpapi.NewProc("GetExtendedTcpTable")
)

// MIB_TCPROW_OWNER_PID represents a row in the TCP connection table
type MIB_TCPROW_OWNER_PID struct {
	dwState      uint32
	dwLocalAddr  uint32
	dwLocalPort  uint32
	dwRemoteAddr uint32
	dwRemotePort uint32
	dwOwningPid  uint32
}

var ConfigDefaults = Config{}

// A Collector is a Prometheus Collector for WMI Win32_PerfRawData_Tcpip_TCPv{4,6} metrics.
type Collector struct {
	config Config

	perfDataCollector4 *perfdata.Collector
	perfDataCollector6 *perfdata.Collector

	connectionFailures         *prometheus.Desc
	connectionsActive          *prometheus.Desc
	connectionsEstablished     *prometheus.Desc
	connectionsPassive         *prometheus.Desc
	connectionsReset           *prometheus.Desc
	segmentsTotal              *prometheus.Desc
	segmentsReceivedTotal      *prometheus.Desc
	segmentsRetransmittedTotal *prometheus.Desc
	segmentsSentTotal          *prometheus.Desc

	connectionsCLOSED    *prometheus.Desc
	connectionsLISTENING *prometheus.Desc
	connectionsSYNSENT   *prometheus.Desc
	connectionsSYNRCVD   *prometheus.Desc
	connectionsFINWAIT1  *prometheus.Desc
	connectionsFINWAIT2  *prometheus.Desc
	connectionsCLOSEWAIT *prometheus.Desc
	connectionsCLOSING   *prometheus.Desc
	connectionsLASTACK   *prometheus.Desc
	connectionsTIMEWAIT  *prometheus.Desc
	connectionsDELETETCB *prometheus.Desc
}

func getExtendedTcpTable(pTCPTable uintptr, pdwSize *uint32, bOrder bool, ulAf uint32, tableClass uint32, reserved uint32) uintptr {
	ret, _, _ := procGetExtendedTcpTable.Call(
		pTCPTable,
		uintptr(unsafe.Pointer(pdwSize)),
		uintptr(boolToInt(bOrder)),
		uintptr(ulAf),
		uintptr(tableClass),
		uintptr(reserved),
	)
	return ret
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
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

func (c *Collector) GetPerfCounter(_ *slog.Logger) ([]string, error) {
	return []string{}, nil
}

func (c *Collector) Close(_ *slog.Logger) error {
	return nil
}

func (c *Collector) Build(_ *slog.Logger, _ *wmi.Client) error {
	counters := []string{
		ConnectionFailures,
		ConnectionsActive,
		ConnectionsEstablished,
		ConnectionsPassive,
		ConnectionsReset,
		SegmentsPersec,
		SegmentsReceivedPersec,
		SegmentsRetransmittedPersec,
		SegmentsSentPersec,
	}

	var err error

	c.perfDataCollector4, err = perfdata.NewCollector("TCPv4", nil, counters)
	if err != nil {
		return fmt.Errorf("failed to create TCPv4 collector: %w", err)
	}

	c.perfDataCollector6, err = perfdata.NewCollector("TCPv6", nil, counters)
	if err != nil {
		return fmt.Errorf("failed to create TCPv6 collector: %w", err)
	}

	c.connectionFailures = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "connection_failures_total"),
		"(TCP.ConnectionFailures)",
		[]string{"af"},
		nil,
	)
	c.connectionsActive = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "connections_active_total"),
		"(TCP.ConnectionsActive)",
		[]string{"af"},
		nil,
	)
	c.connectionsEstablished = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "connections_established"),
		"(TCP.ConnectionsEstablished)",
		[]string{"af"},
		nil,
	)
	c.connectionsPassive = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "connections_passive_total"),
		"(TCP.ConnectionsPassive)",
		[]string{"af"},
		nil,
	)
	c.connectionsReset = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "connections_reset_total"),
		"(TCP.ConnectionsReset)",
		[]string{"af"},
		nil,
	)
	c.segmentsTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "segments_total"),
		"(TCP.SegmentsTotal)",
		[]string{"af"},
		nil,
	)
	c.segmentsReceivedTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "segments_received_total"),
		"(TCP.SegmentsReceivedTotal)",
		[]string{"af"},
		nil,
	)
	c.segmentsRetransmittedTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "segments_retransmitted_total"),
		"(TCP.SegmentsRetransmittedTotal)",
		[]string{"af"},
		nil,
	)
	c.segmentsSentTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "segments_sent_total"),
		"(TCP.SegmentsSentTotal)",
		[]string{"af"},
		nil,
	)
	c.connectionsCLOSED = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "connections_closed"),
		"Number of TCP connections in CLOSED state",
		[]string{"af"},
		nil,
	)
	c.connectionsLISTENING = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "connections_listening"),
		"Number of TCP connections in LISTENING state",
		[]string{"af"},
		nil,
	)
	c.connectionsSYNSENT = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "connections_syn_sent"),
		"Number of TCP connections in SYN_SENT state",
		[]string{"af"},
		nil,
	)
	c.connectionsSYNRCVD = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "connections_syn_rcvd"),
		"Number of TCP connections in SYN_RCVD state",
		[]string{"af"},
		nil,
	)
	c.connectionsFINWAIT1 = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "connections_fin_wait_1"),
		"Number of TCP connections in FIN_WAIT_1 state",
		[]string{"af"},
		nil,
	)
	c.connectionsFINWAIT2 = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "connections_fin_wait_2"),
		"Number of TCP connections in FIN_WAIT_2 state",
		[]string{"af"},
		nil,
	)
	c.connectionsCLOSEWAIT = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "connections_close_wait"),
		"Number of TCP connections in CLOSE_WAIT state",
		[]string{"af"},
		nil,
	)
	c.connectionsCLOSING = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "connections_closing"),
		"Number of TCP connections in CLOSING state",
		[]string{"af"},
		nil,
	)
	c.connectionsLASTACK = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "connections_last_ack"),
		"Number of TCP connections in LAST_ACK state",
		[]string{"af"},
		nil,
	)
	c.connectionsTIMEWAIT = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "connections_time_wait"),
		"Number of TCP connections in TIME_WAIT state",
		[]string{"af"},
		nil,
	)
	c.connectionsDELETETCB = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "connections_delete_tcb"),
		"Number of TCP connections in DELETE_TCB state",
		[]string{"af"},
		nil,
	)
	return nil
}

// Collect sends the metric values for each metric
// to the provided prometheus Metric channel.
func (c *Collector) Collect(_ *types.ScrapeContext, logger *slog.Logger, ch chan<- prometheus.Metric) error {
	logger = logger.With(slog.String("collector", Name))
	if err := c.collect(ch); err != nil {
		logger.Error("failed collecting tcp metrics",
			slog.Any("err", err),
		)

		return err
	}

	return nil
}

func writeTCPCounters(metrics map[string]perfdata.CounterValues, labels []string, c *Collector, ch chan<- prometheus.Metric) {
	ch <- prometheus.MustNewConstMetric(
		c.connectionFailures,
		prometheus.CounterValue,
		metrics[ConnectionFailures].FirstValue,
		labels...,
	)
	ch <- prometheus.MustNewConstMetric(
		c.connectionsActive,
		prometheus.CounterValue,
		metrics[ConnectionsActive].FirstValue,
		labels...,
	)
	ch <- prometheus.MustNewConstMetric(
		c.connectionsEstablished,
		prometheus.GaugeValue,
		metrics[ConnectionsEstablished].FirstValue,
		labels...,
	)
	ch <- prometheus.MustNewConstMetric(
		c.connectionsPassive,
		prometheus.CounterValue,
		metrics[ConnectionsPassive].FirstValue,
		labels...,
	)
	ch <- prometheus.MustNewConstMetric(
		c.connectionsReset,
		prometheus.CounterValue,
		metrics[ConnectionsReset].FirstValue,
		labels...,
	)
	ch <- prometheus.MustNewConstMetric(
		c.segmentsTotal,
		prometheus.CounterValue,
		metrics[SegmentsPersec].FirstValue,
		labels...,
	)
	ch <- prometheus.MustNewConstMetric(
		c.segmentsReceivedTotal,
		prometheus.CounterValue,
		metrics[SegmentsReceivedPersec].FirstValue,
		labels...,
	)
	ch <- prometheus.MustNewConstMetric(
		c.segmentsRetransmittedTotal,
		prometheus.CounterValue,
		metrics[SegmentsRetransmittedPersec].FirstValue,
		labels...,
	)
	ch <- prometheus.MustNewConstMetric(
		c.segmentsSentTotal,
		prometheus.CounterValue,
		metrics[SegmentsSentPersec].FirstValue,
		labels...,
	)
}

func (c *Collector) collect(ch chan<- prometheus.Metric) error {
	data, err := c.perfDataCollector4.Collect()
	if err != nil {
		return fmt.Errorf("failed to collect TCPv4 metrics: %w", err)
	}

	writeTCPCounters(data[perfdata.EmptyInstance], []string{"ipv4"}, c, ch)

	data, err = c.perfDataCollector6.Collect()
	if err != nil {
		return fmt.Errorf("failed to collect TCPv6 metrics: %w", err)
	}

	writeTCPCounters(data[perfdata.EmptyInstance], []string{"ipv6"}, c, ch)

	stateCountsIPv4, err := getTCPConnectionStates(syscall.AF_INET)
	if err != nil {
		return fmt.Errorf("failed to collect TCP connection states for IPv4: %w", err)
	}

	stateCountsIPv6, err := getTCPConnectionStates(syscall.AF_INET6)
	if err != nil {
		return fmt.Errorf("failed to collect TCP connection states for IPv6: %w", err)
	}

	c.sendTCPStateMetrics(ch, stateCountsIPv4, "ipv4")
	c.sendTCPStateMetrics(ch, stateCountsIPv6, "ipv6")
	return nil
}

func (c *Collector) sendTCPStateMetrics(ch chan<- prometheus.Metric, stateCounts map[uint32]uint32, af string) {
	ch <- prometheus.MustNewConstMetric(c.connectionsCLOSED, prometheus.GaugeValue, float64(stateCounts[TCPStateClosed]), af)
	ch <- prometheus.MustNewConstMetric(c.connectionsLISTENING, prometheus.GaugeValue, float64(stateCounts[TCPStateListening]), af)
	ch <- prometheus.MustNewConstMetric(c.connectionsSYNSENT, prometheus.GaugeValue, float64(stateCounts[TCPStateSynSent]), af)
	ch <- prometheus.MustNewConstMetric(c.connectionsSYNRCVD, prometheus.GaugeValue, float64(stateCounts[TCPStateSynRcvd]), af)
	ch <- prometheus.MustNewConstMetric(c.connectionsFINWAIT1, prometheus.GaugeValue, float64(stateCounts[TCPStateFinWait1]), af)
	ch <- prometheus.MustNewConstMetric(c.connectionsFINWAIT2, prometheus.GaugeValue, float64(stateCounts[TCPStateFinWait2]), af)
	ch <- prometheus.MustNewConstMetric(c.connectionsCLOSEWAIT, prometheus.GaugeValue, float64(stateCounts[TCPStateCloseWait]), af)
	ch <- prometheus.MustNewConstMetric(c.connectionsCLOSING, prometheus.GaugeValue, float64(stateCounts[TCPStateClosing]), af)
	ch <- prometheus.MustNewConstMetric(c.connectionsLASTACK, prometheus.GaugeValue, float64(stateCounts[TCPStateLastAck]), af)
	ch <- prometheus.MustNewConstMetric(c.connectionsTIMEWAIT, prometheus.GaugeValue, float64(stateCounts[TCPStateTimeWait]), af)
	ch <- prometheus.MustNewConstMetric(c.connectionsDELETETCB, prometheus.GaugeValue, float64(stateCounts[TCPStateDeleteTcb]), af)
}

func getTCPConnectionStates(family uint32) (map[uint32]uint32, error) {
	var size uint32
	stateCounts := make(map[uint32]uint32)

	ret := getExtendedTcpTable(0, &size, true, family, TCPTableClass, 0)
	if ret != 0 && ret != uintptr(syscall.ERROR_INSUFFICIENT_BUFFER) {
		return nil, errors.Errorf("getExtendedTcpTable failed with code %d", ret)
	}

	buf := make([]byte, size)
	ret = getExtendedTcpTable(uintptr(unsafe.Pointer(&buf[0])), &size, true, family, TCPTableClass, 0)
	if ret != 0 {
		return nil, errors.Errorf("getExtendedTcpTable failed with code %d", ret)
	}

	numEntries := *(*uint32)(unsafe.Pointer(&buf[0]))
	for i := uint32(0); i < numEntries; i++ {
		row := (*MIB_TCPROW_OWNER_PID)(unsafe.Pointer(&buf[4+i*uint32(unsafe.Sizeof(MIB_TCPROW_OWNER_PID{}))]))
		stateCounts[row.dwState]++
	}

	return stateCounts, nil
}
