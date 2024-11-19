package hyperv

import (
	"errors"
	"fmt"

	"github.com/prometheus-community/windows_exporter/internal/perfdata"
	"github.com/prometheus-community/windows_exporter/internal/types"
	"github.com/prometheus/client_golang/prometheus"
)

// collectorVirtualNetworkAdapterDropReasons Hyper-V Virtual Network Adapter Drop Reasons metrics
type collectorVirtualNetworkAdapterDropReasons struct {
	perfDataCollectorVirtualNetworkAdapterDropReasons *perfdata.Collector

	// \Hyper-V Virtual Network Adapter Drop Reasons(*)\Outgoing LowPowerPacketFilter
	// \Hyper-V Virtual Network Adapter Drop Reasons(*)\Incoming LowPowerPacketFilter
	// \Hyper-V Virtual Network Adapter Drop Reasons(*)\Outgoing InvalidPDQueue
	// \Hyper-V Virtual Network Adapter Drop Reasons(*)\Incoming InvalidPDQueue
	// \Hyper-V Virtual Network Adapter Drop Reasons(*)\Outgoing FilteredIsolationUntagged
	// \Hyper-V Virtual Network Adapter Drop Reasons(*)\Incoming FilteredIsolationUntagged
	// \Hyper-V Virtual Network Adapter Drop Reasons(*)\Outgoing SwitchDataFlowDisabled
	// \Hyper-V Virtual Network Adapter Drop Reasons(*)\Incoming SwitchDataFlowDisabled
	// \Hyper-V Virtual Network Adapter Drop Reasons(*)\Outgoing FailedPacketFilter
	// \Hyper-V Virtual Network Adapter Drop Reasons(*)\Incoming FailedPacketFilter
	// \Hyper-V Virtual Network Adapter Drop Reasons(*)\Outgoing NicDisabled
	// \Hyper-V Virtual Network Adapter Drop Reasons(*)\Incoming NicDisabled
	// \Hyper-V Virtual Network Adapter Drop Reasons(*)\Outgoing FailedDestinationListUpdate
	// \Hyper-V Virtual Network Adapter Drop Reasons(*)\Incoming FailedDestinationListUpdate
	// \Hyper-V Virtual Network Adapter Drop Reasons(*)\Outgoing InjectedIcmp
	// \Hyper-V Virtual Network Adapter Drop Reasons(*)\Incoming InjectedIcmp
	// \Hyper-V Virtual Network Adapter Drop Reasons(*)\Outgoing StormLimit
	// \Hyper-V Virtual Network Adapter Drop Reasons(*)\Incoming StormLimit
	// \Hyper-V Virtual Network Adapter Drop Reasons(*)\Outgoing InvalidFirstNBTooSmall
	// \Hyper-V Virtual Network Adapter Drop Reasons(*)\Incoming InvalidFirstNBTooSmall
	// \Hyper-V Virtual Network Adapter Drop Reasons(*)\Outgoing InvalidSourceMac
	// \Hyper-V Virtual Network Adapter Drop Reasons(*)\Incoming InvalidSourceMac
	// \Hyper-V Virtual Network Adapter Drop Reasons(*)\Outgoing InvalidDestMac
	// \Hyper-V Virtual Network Adapter Drop Reasons(*)\Incoming InvalidDestMac
	// \Hyper-V Virtual Network Adapter Drop Reasons(*)\Outgoing InvalidVlanFormat
	// \Hyper-V Virtual Network Adapter Drop Reasons(*)\Incoming InvalidVlanFormat
	// \Hyper-V Virtual Network Adapter Drop Reasons(*)\Outgoing NativeFwdingReq
	// \Hyper-V Virtual Network Adapter Drop Reasons(*)\Incoming NativeFwdingReq
	// \Hyper-V Virtual Network Adapter Drop Reasons(*)\Outgoing MTUMismatch
	// \Hyper-V Virtual Network Adapter Drop Reasons(*)\Incoming MTUMismatch
	// \Hyper-V Virtual Network Adapter Drop Reasons(*)\Outgoing InvalidConfig
	// \Hyper-V Virtual Network Adapter Drop Reasons(*)\Incoming InvalidConfig
	// \Hyper-V Virtual Network Adapter Drop Reasons(*)\Outgoing RequiredExtensionMissing
	// \Hyper-V Virtual Network Adapter Drop Reasons(*)\Incoming RequiredExtensionMissing
	// \Hyper-V Virtual Network Adapter Drop Reasons(*)\Outgoing VirtualSubnetId
	// \Hyper-V Virtual Network Adapter Drop Reasons(*)\Incoming VirtualSubnetId
	// \Hyper-V Virtual Network Adapter Drop Reasons(*)\Outgoing BridgeReserved
	// \Hyper-V Virtual Network Adapter Drop Reasons(*)\Incoming BridgeReserved
	// \Hyper-V Virtual Network Adapter Drop Reasons(*)\Outgoing RouterGuard
	// \Hyper-V Virtual Network Adapter Drop Reasons(*)\Incoming RouterGuard
	// \Hyper-V Virtual Network Adapter Drop Reasons(*)\Outgoing DhcpGuard
	// \Hyper-V Virtual Network Adapter Drop Reasons(*)\Incoming DhcpGuard
	// \Hyper-V Virtual Network Adapter Drop Reasons(*)\Outgoing MacSpoofing
	// \Hyper-V Virtual Network Adapter Drop Reasons(*)\Incoming MacSpoofing
	// \Hyper-V Virtual Network Adapter Drop Reasons(*)\Outgoing Ipsec
	// \Hyper-V Virtual Network Adapter Drop Reasons(*)\Incoming Ipsec
	// \Hyper-V Virtual Network Adapter Drop Reasons(*)\Outgoing Qos
	// \Hyper-V Virtual Network Adapter Drop Reasons(*)\Incoming Qos
	// \Hyper-V Virtual Network Adapter Drop Reasons(*)\Outgoing FailedPvlanSetting
	// \Hyper-V Virtual Network Adapter Drop Reasons(*)\Incoming FailedPvlanSetting
	// \Hyper-V Virtual Network Adapter Drop Reasons(*)\Outgoing FailedSecurityPolicy
	// \Hyper-V Virtual Network Adapter Drop Reasons(*)\Incoming FailedSecurityPolicy
	// \Hyper-V Virtual Network Adapter Drop Reasons(*)\Outgoing UnauthorizedMAC
	// \Hyper-V Virtual Network Adapter Drop Reasons(*)\Incoming UnauthorizedMAC
	// \Hyper-V Virtual Network Adapter Drop Reasons(*)\Outgoing UnauthorizedVLAN
	// \Hyper-V Virtual Network Adapter Drop Reasons(*)\Incoming UnauthorizedVLAN
	// \Hyper-V Virtual Network Adapter Drop Reasons(*)\Outgoing FilteredVLAN
	// \Hyper-V Virtual Network Adapter Drop Reasons(*)\Incoming FilteredVLAN
	// \Hyper-V Virtual Network Adapter Drop Reasons(*)\Outgoing Filtered
	// \Hyper-V Virtual Network Adapter Drop Reasons(*)\Incoming Filtered
	// \Hyper-V Virtual Network Adapter Drop Reasons(*)\Outgoing Busy
	// \Hyper-V Virtual Network Adapter Drop Reasons(*)\Incoming Busy
	// \Hyper-V Virtual Network Adapter Drop Reasons(*)\Outgoing NotAccepted
	// \Hyper-V Virtual Network Adapter Drop Reasons(*)\Incoming NotAccepted
	// \Hyper-V Virtual Network Adapter Drop Reasons(*)\Outgoing Disconnected
	// \Hyper-V Virtual Network Adapter Drop Reasons(*)\Incoming Disconnected
	// \Hyper-V Virtual Network Adapter Drop Reasons(*)\Outgoing NotReady
	// \Hyper-V Virtual Network Adapter Drop Reasons(*)\Incoming NotReady
	// \Hyper-V Virtual Network Adapter Drop Reasons(*)\Outgoing Resources
	// \Hyper-V Virtual Network Adapter Drop Reasons(*)\Incoming Resources
	// \Hyper-V Virtual Network Adapter Drop Reasons(*)\Outgoing InvalidPacket
	// \Hyper-V Virtual Network Adapter Drop Reasons(*)\Incoming InvalidPacket
	// \Hyper-V Virtual Network Adapter Drop Reasons(*)\Outgoing InvalidData
	// \Hyper-V Virtual Network Adapter Drop Reasons(*)\Incoming InvalidData
	// \Hyper-V Virtual Network Adapter Drop Reasons(*)\Outgoing Unknown
	// \Hyper-V Virtual Network Adapter Drop Reasons(*)\Incoming Unknown
	virtualNetworkAdapterDropReasons *prometheus.Desc
}

const (
	virtualNetworkAdapterDropReasonsOutgoingNativeFwdingReq          = "Outgoing NativeFwdingReq"
	virtualNetworkAdapterDropReasonsIncomingNativeFwdingReq          = "Incoming NativeFwdingReq"
	virtualNetworkAdapterDropReasonsOutgoingMTUMismatch              = "Outgoing MTUMismatch"
	virtualNetworkAdapterDropReasonsIncomingMTUMismatch              = "Incoming MTUMismatch"
	virtualNetworkAdapterDropReasonsOutgoingInvalidConfig            = "Outgoing InvalidConfig"
	virtualNetworkAdapterDropReasonsIncomingInvalidConfig            = "Incoming InvalidConfig"
	virtualNetworkAdapterDropReasonsOutgoingRequiredExtensionMissing = "Outgoing RequiredExtensionMissing"
	virtualNetworkAdapterDropReasonsIncomingRequiredExtensionMissing = "Incoming RequiredExtensionMissing"
	virtualNetworkAdapterDropReasonsOutgoingVirtualSubnetId          = "Outgoing VirtualSubnetId"
	virtualNetworkAdapterDropReasonsIncomingVirtualSubnetId          = "Incoming VirtualSubnetId"
	virtualNetworkAdapterDropReasonsOutgoingBridgeReserved           = "Outgoing BridgeReserved"
	virtualNetworkAdapterDropReasonsIncomingBridgeReserved           = "Incoming BridgeReserved"
	virtualNetworkAdapterDropReasonsOutgoingRouterGuard              = "Outgoing RouterGuard"
	virtualNetworkAdapterDropReasonsIncomingRouterGuard              = "Incoming RouterGuard"
	virtualNetworkAdapterDropReasonsOutgoingDhcpGuard                = "Outgoing DhcpGuard"
	virtualNetworkAdapterDropReasonsIncomingDhcpGuard                = "Incoming DhcpGuard"
	virtualNetworkAdapterDropReasonsOutgoingMacSpoofing              = "Outgoing MacSpoofing"
	virtualNetworkAdapterDropReasonsIncomingMacSpoofing              = "Incoming MacSpoofing"
	virtualNetworkAdapterDropReasonsOutgoingIpsec                    = "Outgoing Ipsec"
	virtualNetworkAdapterDropReasonsIncomingIpsec                    = "Incoming Ipsec"
	virtualNetworkAdapterDropReasonsOutgoingQos                      = "Outgoing Qos"
	virtualNetworkAdapterDropReasonsIncomingQos                      = "Incoming Qos"
	virtualNetworkAdapterDropReasonsOutgoingFailedPvlanSetting       = "Outgoing FailedPvlanSetting"
	virtualNetworkAdapterDropReasonsIncomingFailedPvlanSetting       = "Incoming FailedPvlanSetting"
	virtualNetworkAdapterDropReasonsOutgoingFailedSecurityPolicy     = "Outgoing FailedSecurityPolicy"
	virtualNetworkAdapterDropReasonsIncomingFailedSecurityPolicy     = "Incoming FailedSecurityPolicy"
	virtualNetworkAdapterDropReasonsOutgoingUnauthorizedMAC          = "Outgoing UnauthorizedMAC"
	virtualNetworkAdapterDropReasonsIncomingUnauthorizedMAC          = "Incoming UnauthorizedMAC"
	virtualNetworkAdapterDropReasonsOutgoingUnauthorizedVLAN         = "Outgoing UnauthorizedVLAN"
	virtualNetworkAdapterDropReasonsIncomingUnauthorizedVLAN         = "Incoming UnauthorizedVLAN"
	virtualNetworkAdapterDropReasonsOutgoingFilteredVLAN             = "Outgoing FilteredVLAN"
	virtualNetworkAdapterDropReasonsIncomingFilteredVLAN             = "Incoming FilteredVLAN"
	virtualNetworkAdapterDropReasonsOutgoingFiltered                 = "Outgoing Filtered"
	virtualNetworkAdapterDropReasonsIncomingFiltered                 = "Incoming Filtered"
	virtualNetworkAdapterDropReasonsOutgoingBusy                     = "Outgoing Busy"
	virtualNetworkAdapterDropReasonsIncomingBusy                     = "Incoming Busy"
	virtualNetworkAdapterDropReasonsOutgoingNotAccepted              = "Outgoing NotAccepted"
	virtualNetworkAdapterDropReasonsIncomingNotAccepted              = "Incoming NotAccepted"
	virtualNetworkAdapterDropReasonsOutgoingDisconnected             = "Outgoing Disconnected"
	virtualNetworkAdapterDropReasonsIncomingDisconnected             = "Incoming Disconnected"
	virtualNetworkAdapterDropReasonsOutgoingNotReady                 = "Outgoing NotReady"
	virtualNetworkAdapterDropReasonsIncomingNotReady                 = "Incoming NotReady"
	virtualNetworkAdapterDropReasonsOutgoingResources                = "Outgoing Resources"
	virtualNetworkAdapterDropReasonsIncomingResources                = "Incoming Resources"
	virtualNetworkAdapterDropReasonsOutgoingInvalidPacket            = "Outgoing InvalidPacket"
	virtualNetworkAdapterDropReasonsIncomingInvalidPacket            = "Incoming InvalidPacket"
	virtualNetworkAdapterDropReasonsOutgoingInvalidData              = "Outgoing InvalidData"
	virtualNetworkAdapterDropReasonsIncomingInvalidData              = "Incoming InvalidData"
	virtualNetworkAdapterDropReasonsOutgoingUnknown                  = "Outgoing Unknown"
	virtualNetworkAdapterDropReasonsIncomingUnknown                  = "Incoming Unknown"
)

func (c *Collector) buildVirtualNetworkAdapterDropReasons() error {
	var err error

	c.perfDataCollectorVirtualNetworkAdapterDropReasons, err = perfdata.NewCollector("Hyper-V Virtual Network Adapter Drop Reasons", perfdata.InstanceAll, []string{
		virtualNetworkAdapterDropReasonsOutgoingNativeFwdingReq,
		virtualNetworkAdapterDropReasonsIncomingNativeFwdingReq,
		virtualNetworkAdapterDropReasonsOutgoingMTUMismatch,
		virtualNetworkAdapterDropReasonsIncomingMTUMismatch,
		virtualNetworkAdapterDropReasonsOutgoingInvalidConfig,
		virtualNetworkAdapterDropReasonsIncomingInvalidConfig,
		virtualNetworkAdapterDropReasonsOutgoingRequiredExtensionMissing,
		virtualNetworkAdapterDropReasonsIncomingRequiredExtensionMissing,
		virtualNetworkAdapterDropReasonsOutgoingVirtualSubnetId,
		virtualNetworkAdapterDropReasonsIncomingVirtualSubnetId,
		virtualNetworkAdapterDropReasonsOutgoingBridgeReserved,
		virtualNetworkAdapterDropReasonsIncomingBridgeReserved,
		virtualNetworkAdapterDropReasonsOutgoingRouterGuard,
		virtualNetworkAdapterDropReasonsIncomingRouterGuard,
		virtualNetworkAdapterDropReasonsOutgoingDhcpGuard,
		virtualNetworkAdapterDropReasonsIncomingDhcpGuard,
		virtualNetworkAdapterDropReasonsOutgoingMacSpoofing,
		virtualNetworkAdapterDropReasonsIncomingMacSpoofing,
		virtualNetworkAdapterDropReasonsOutgoingIpsec,
		virtualNetworkAdapterDropReasonsIncomingIpsec,
		virtualNetworkAdapterDropReasonsOutgoingQos,
		virtualNetworkAdapterDropReasonsIncomingQos,
		virtualNetworkAdapterDropReasonsOutgoingFailedPvlanSetting,
		virtualNetworkAdapterDropReasonsIncomingFailedPvlanSetting,
		virtualNetworkAdapterDropReasonsOutgoingFailedSecurityPolicy,
		virtualNetworkAdapterDropReasonsIncomingFailedSecurityPolicy,
		virtualNetworkAdapterDropReasonsOutgoingUnauthorizedMAC,
		virtualNetworkAdapterDropReasonsIncomingUnauthorizedMAC,
		virtualNetworkAdapterDropReasonsOutgoingUnauthorizedVLAN,
		virtualNetworkAdapterDropReasonsIncomingUnauthorizedVLAN,
		virtualNetworkAdapterDropReasonsOutgoingFilteredVLAN,
		virtualNetworkAdapterDropReasonsIncomingFilteredVLAN,
		virtualNetworkAdapterDropReasonsOutgoingFiltered,
		virtualNetworkAdapterDropReasonsIncomingFiltered,
		virtualNetworkAdapterDropReasonsOutgoingBusy,
		virtualNetworkAdapterDropReasonsIncomingBusy,
		virtualNetworkAdapterDropReasonsOutgoingNotAccepted,
		virtualNetworkAdapterDropReasonsIncomingNotAccepted,
		virtualNetworkAdapterDropReasonsOutgoingDisconnected,
		virtualNetworkAdapterDropReasonsIncomingDisconnected,
		virtualNetworkAdapterDropReasonsOutgoingNotReady,
		virtualNetworkAdapterDropReasonsIncomingNotReady,
		virtualNetworkAdapterDropReasonsOutgoingResources,
		virtualNetworkAdapterDropReasonsIncomingResources,
		virtualNetworkAdapterDropReasonsOutgoingInvalidPacket,
		virtualNetworkAdapterDropReasonsIncomingInvalidPacket,
		virtualNetworkAdapterDropReasonsOutgoingInvalidData,
		virtualNetworkAdapterDropReasonsIncomingInvalidData,
		virtualNetworkAdapterDropReasonsOutgoingUnknown,
		virtualNetworkAdapterDropReasonsIncomingUnknown,
	})
	if err != nil && !errors.Is(err, perfdata.ErrNoData) {
		return fmt.Errorf("failed to create Hyper-V Virtual Network Adapter Drop Reasons collector: %w", err)
	}

	c.virtualNetworkAdapterDropReasons = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "virtual_network_adapter_drop_reasons"),
		"Hyper-V Virtual Network Adapter Drop Reasons",
		[]string{"adapter", "reason", "direction"},
		nil,
	)

	return nil
}

func (c *Collector) collectVirtualNetworkAdapterDropReasons(ch chan<- prometheus.Metric) error {
	data, err := c.perfDataCollectorVirtualNetworkAdapterDropReasons.Collect()
	if err != nil && !errors.Is(err, perfdata.ErrNoData) {
		return fmt.Errorf("failed to collect Hyper-V Virtual Network Adapter Drop Reasons metrics: %w", err)
	}

	for name, adapterData := range data {
		ch <- prometheus.MustNewConstMetric(
			c.virtualNetworkAdapterDropReasons,
			prometheus.CounterValue,
			adapterData[virtualNetworkAdapterDropReasonsOutgoingNativeFwdingReq].FirstValue,
			name, "NativeFwdingReq", "outgoing",
		)
		ch <- prometheus.MustNewConstMetric(
			c.virtualNetworkAdapterDropReasons,
			prometheus.CounterValue,
			adapterData[virtualNetworkAdapterDropReasonsIncomingNativeFwdingReq].FirstValue,
			name, "NativeFwdingReq", "incoming",
		)
		ch <- prometheus.MustNewConstMetric(
			c.virtualNetworkAdapterDropReasons,
			prometheus.CounterValue,
			adapterData[virtualNetworkAdapterDropReasonsOutgoingMTUMismatch].FirstValue,
			name, "MTUMismatch", "outgoing",
		)
		ch <- prometheus.MustNewConstMetric(
			c.virtualNetworkAdapterDropReasons,
			prometheus.CounterValue,
			adapterData[virtualNetworkAdapterDropReasonsIncomingMTUMismatch].FirstValue,
			name, "MTUMismatch", "incoming",
		)
		ch <- prometheus.MustNewConstMetric(
			c.virtualNetworkAdapterDropReasons,
			prometheus.CounterValue,
			adapterData[virtualNetworkAdapterDropReasonsOutgoingInvalidConfig].FirstValue,
			name, "InvalidConfig", "outgoing",
		)
		ch <- prometheus.MustNewConstMetric(
			c.virtualNetworkAdapterDropReasons,
			prometheus.CounterValue,
			adapterData[virtualNetworkAdapterDropReasonsIncomingInvalidConfig].FirstValue,
			name, "InvalidConfig", "incoming",
		)
		ch <- prometheus.MustNewConstMetric(
			c.virtualNetworkAdapterDropReasons,
			prometheus.CounterValue,
			adapterData[virtualNetworkAdapterDropReasonsOutgoingRequiredExtensionMissing].FirstValue,
			name, "RequiredExtensionMissing", "outgoing",
		)
		ch <- prometheus.MustNewConstMetric(
			c.virtualNetworkAdapterDropReasons,
			prometheus.CounterValue,
			adapterData[virtualNetworkAdapterDropReasonsIncomingRequiredExtensionMissing].FirstValue,
			name, "RequiredExtensionMissing", "incoming",
		)
		ch <- prometheus.MustNewConstMetric(
			c.virtualNetworkAdapterDropReasons,
			prometheus.CounterValue,
			adapterData[virtualNetworkAdapterDropReasonsOutgoingVirtualSubnetId].FirstValue,
			name, "VirtualSubnetId", "outgoing",
		)
		ch <- prometheus.MustNewConstMetric(
			c.virtualNetworkAdapterDropReasons,
			prometheus.CounterValue,
			adapterData[virtualNetworkAdapterDropReasonsIncomingVirtualSubnetId].FirstValue,
			name, "VirtualSubnetId", "incoming",
		)
		ch <- prometheus.MustNewConstMetric(
			c.virtualNetworkAdapterDropReasons,
			prometheus.CounterValue,
			adapterData[virtualNetworkAdapterDropReasonsOutgoingBridgeReserved].FirstValue,
			name, "BridgeReserved", "outgoing",
		)
		ch <- prometheus.MustNewConstMetric(
			c.virtualNetworkAdapterDropReasons,
			prometheus.CounterValue,
			adapterData[virtualNetworkAdapterDropReasonsIncomingBridgeReserved].FirstValue,
			name, "BridgeReserved", "incoming",
		)
		ch <- prometheus.MustNewConstMetric(
			c.virtualNetworkAdapterDropReasons,
			prometheus.CounterValue,
			adapterData[virtualNetworkAdapterDropReasonsOutgoingRouterGuard].FirstValue,
			name, "RouterGuard", "outgoing",
		)
		ch <- prometheus.MustNewConstMetric(
			c.virtualNetworkAdapterDropReasons,
			prometheus.CounterValue,
			adapterData[virtualNetworkAdapterDropReasonsIncomingRouterGuard].FirstValue,
			name, "RouterGuard", "incoming",
		)
		ch <- prometheus.MustNewConstMetric(
			c.virtualNetworkAdapterDropReasons,
			prometheus.CounterValue,
			adapterData[virtualNetworkAdapterDropReasonsOutgoingDhcpGuard].FirstValue,
			name, "DhcpGuard", "outgoing",
		)
		ch <- prometheus.MustNewConstMetric(
			c.virtualNetworkAdapterDropReasons,
			prometheus.CounterValue,
			adapterData[virtualNetworkAdapterDropReasonsIncomingDhcpGuard].FirstValue,
			name, "DhcpGuard", "incoming",
		)
		ch <- prometheus.MustNewConstMetric(
			c.virtualNetworkAdapterDropReasons,
			prometheus.CounterValue,
			adapterData[virtualNetworkAdapterDropReasonsOutgoingMacSpoofing].FirstValue,
			name, "MacSpoofing", "outgoing",
		)
		ch <- prometheus.MustNewConstMetric(
			c.virtualNetworkAdapterDropReasons,
			prometheus.CounterValue,
			adapterData[virtualNetworkAdapterDropReasonsIncomingMacSpoofing].FirstValue,
			name, "MacSpoofing", "incoming",
		)
		ch <- prometheus.MustNewConstMetric(
			c.virtualNetworkAdapterDropReasons,
			prometheus.CounterValue,
			adapterData[virtualNetworkAdapterDropReasonsOutgoingIpsec].FirstValue,
			name, "Ipsec", "outgoing",
		)
		ch <- prometheus.MustNewConstMetric(
			c.virtualNetworkAdapterDropReasons,
			prometheus.CounterValue,
			adapterData[virtualNetworkAdapterDropReasonsIncomingIpsec].FirstValue,
			name, "Ipsec", "incoming",
		)
		ch <- prometheus.MustNewConstMetric(
			c.virtualNetworkAdapterDropReasons,
			prometheus.CounterValue,
			adapterData[virtualNetworkAdapterDropReasonsOutgoingQos].FirstValue,
			name, "Qos", "outgoing",
		)
		ch <- prometheus.MustNewConstMetric(
			c.virtualNetworkAdapterDropReasons,
			prometheus.CounterValue,
			adapterData[virtualNetworkAdapterDropReasonsIncomingQos].FirstValue,
			name, "Qos", "incoming",
		)
		ch <- prometheus.MustNewConstMetric(
			c.virtualNetworkAdapterDropReasons,
			prometheus.CounterValue,
			adapterData[virtualNetworkAdapterDropReasonsOutgoingFailedPvlanSetting].FirstValue,
			name, "FailedPvlanSetting", "outgoing",
		)
		ch <- prometheus.MustNewConstMetric(
			c.virtualNetworkAdapterDropReasons,
			prometheus.CounterValue,
			adapterData[virtualNetworkAdapterDropReasonsIncomingFailedPvlanSetting].FirstValue,
			name, "FailedPvlanSetting", "incoming",
		)
		ch <- prometheus.MustNewConstMetric(
			c.virtualNetworkAdapterDropReasons,
			prometheus.CounterValue,
			adapterData[virtualNetworkAdapterDropReasonsOutgoingFailedSecurityPolicy].FirstValue,
			name, "FailedSecurityPolicy", "outgoing",
		)
		ch <- prometheus.MustNewConstMetric(
			c.virtualNetworkAdapterDropReasons,
			prometheus.CounterValue,
			adapterData[virtualNetworkAdapterDropReasonsIncomingFailedSecurityPolicy].FirstValue,
			name, "FailedSecurityPolicy", "incoming",
		)
		ch <- prometheus.MustNewConstMetric(
			c.virtualNetworkAdapterDropReasons,
			prometheus.CounterValue,
			adapterData[virtualNetworkAdapterDropReasonsOutgoingUnauthorizedMAC].FirstValue,
			name, "UnauthorizedMAC", "outgoing",
		)
		ch <- prometheus.MustNewConstMetric(
			c.virtualNetworkAdapterDropReasons,
			prometheus.CounterValue,
			adapterData[virtualNetworkAdapterDropReasonsIncomingUnauthorizedMAC].FirstValue,
			name, "UnauthorizedMAC", "incoming",
		)
		ch <- prometheus.MustNewConstMetric(
			c.virtualNetworkAdapterDropReasons,
			prometheus.CounterValue,
			adapterData[virtualNetworkAdapterDropReasonsOutgoingUnauthorizedVLAN].FirstValue,
			name, "UnauthorizedVLAN", "outgoing",
		)
		ch <- prometheus.MustNewConstMetric(
			c.virtualNetworkAdapterDropReasons,
			prometheus.CounterValue,
			adapterData[virtualNetworkAdapterDropReasonsIncomingUnauthorizedVLAN].FirstValue,
			name, "UnauthorizedVLAN", "incoming",
		)
		ch <- prometheus.MustNewConstMetric(
			c.virtualNetworkAdapterDropReasons,
			prometheus.CounterValue,
			adapterData[virtualNetworkAdapterDropReasonsOutgoingFilteredVLAN].FirstValue,
			name, "FilteredVLAN", "outgoing",
		)
		ch <- prometheus.MustNewConstMetric(
			c.virtualNetworkAdapterDropReasons,
			prometheus.CounterValue,
			adapterData[virtualNetworkAdapterDropReasonsIncomingFilteredVLAN].FirstValue,
			name, "FilteredVLAN", "incoming",
		)
		ch <- prometheus.MustNewConstMetric(
			c.virtualNetworkAdapterDropReasons,
			prometheus.CounterValue,
			adapterData[virtualNetworkAdapterDropReasonsOutgoingFiltered].FirstValue,
			name, "Filtered", "outgoing",
		)
		ch <- prometheus.MustNewConstMetric(
			c.virtualNetworkAdapterDropReasons,
			prometheus.CounterValue,
			adapterData[virtualNetworkAdapterDropReasonsIncomingFiltered].FirstValue,
			name, "Filtered", "incoming",
		)
		ch <- prometheus.MustNewConstMetric(
			c.virtualNetworkAdapterDropReasons,
			prometheus.CounterValue,
			adapterData[virtualNetworkAdapterDropReasonsOutgoingBusy].FirstValue,
			name, "Busy", "outgoing",
		)
		ch <- prometheus.MustNewConstMetric(
			c.virtualNetworkAdapterDropReasons,
			prometheus.CounterValue,
			adapterData[virtualNetworkAdapterDropReasonsIncomingBusy].FirstValue,
			name, "Busy", "incoming",
		)
		ch <- prometheus.MustNewConstMetric(
			c.virtualNetworkAdapterDropReasons,
			prometheus.CounterValue,
			adapterData[virtualNetworkAdapterDropReasonsOutgoingNotAccepted].FirstValue,
			name, "NotAccepted", "outgoing",
		)
		ch <- prometheus.MustNewConstMetric(
			c.virtualNetworkAdapterDropReasons,
			prometheus.CounterValue,
			adapterData[virtualNetworkAdapterDropReasonsIncomingNotAccepted].FirstValue,
			name, "NotAccepted", "incoming",
		)
		ch <- prometheus.MustNewConstMetric(
			c.virtualNetworkAdapterDropReasons,
			prometheus.CounterValue,
			adapterData[virtualNetworkAdapterDropReasonsOutgoingDisconnected].FirstValue,
			name, "Disconnected", "outgoing",
		)
		ch <- prometheus.MustNewConstMetric(
			c.virtualNetworkAdapterDropReasons,
			prometheus.CounterValue,
			adapterData[virtualNetworkAdapterDropReasonsIncomingDisconnected].FirstValue,
			name, "Disconnected", "incoming",
		)
		ch <- prometheus.MustNewConstMetric(
			c.virtualNetworkAdapterDropReasons,
			prometheus.CounterValue,
			adapterData[virtualNetworkAdapterDropReasonsOutgoingNotReady].FirstValue,
			name, "NotReady", "outgoing",
		)
		ch <- prometheus.MustNewConstMetric(
			c.virtualNetworkAdapterDropReasons,
			prometheus.CounterValue,
			adapterData[virtualNetworkAdapterDropReasonsIncomingNotReady].FirstValue,
			name, "NotReady", "incoming",
		)
		ch <- prometheus.MustNewConstMetric(
			c.virtualNetworkAdapterDropReasons,
			prometheus.CounterValue,
			adapterData[virtualNetworkAdapterDropReasonsOutgoingResources].FirstValue,
			name, "Resources", "outgoing",
		)
		ch <- prometheus.MustNewConstMetric(
			c.virtualNetworkAdapterDropReasons,
			prometheus.CounterValue,
			adapterData[virtualNetworkAdapterDropReasonsIncomingResources].FirstValue,
			name, "Resources", "incoming",
		)
		ch <- prometheus.MustNewConstMetric(
			c.virtualNetworkAdapterDropReasons,
			prometheus.CounterValue,
			adapterData[virtualNetworkAdapterDropReasonsOutgoingInvalidPacket].FirstValue,
			name, "InvalidPacket", "outgoing",
		)
		ch <- prometheus.MustNewConstMetric(
			c.virtualNetworkAdapterDropReasons,
			prometheus.CounterValue,
			adapterData[virtualNetworkAdapterDropReasonsIncomingInvalidPacket].FirstValue,
			name, "InvalidPacket", "incoming",
		)
		ch <- prometheus.MustNewConstMetric(
			c.virtualNetworkAdapterDropReasons,
			prometheus.CounterValue,
			adapterData[virtualNetworkAdapterDropReasonsOutgoingInvalidData].FirstValue,
			name, "InvalidData", "outgoing",
		)
		ch <- prometheus.MustNewConstMetric(
			c.virtualNetworkAdapterDropReasons,
			prometheus.CounterValue,
			adapterData[virtualNetworkAdapterDropReasonsIncomingInvalidData].FirstValue,
			name, "InvalidData", "incoming",
		)
		ch <- prometheus.MustNewConstMetric(
			c.virtualNetworkAdapterDropReasons,
			prometheus.CounterValue,
			adapterData[virtualNetworkAdapterDropReasonsOutgoingUnknown].FirstValue,
			name, "Unknown", "outgoing",
		)
		ch <- prometheus.MustNewConstMetric(
			c.virtualNetworkAdapterDropReasons,
			prometheus.CounterValue,
			adapterData[virtualNetworkAdapterDropReasonsIncomingUnknown].FirstValue,
			name, "Unknown", "incoming",
		)
	}

	return nil
}
