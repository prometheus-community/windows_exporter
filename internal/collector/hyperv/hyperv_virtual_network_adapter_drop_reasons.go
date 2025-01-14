// Copyright 2024 The Prometheus Authors
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

//go:build windows

package hyperv

import (
	"fmt"

	"github.com/prometheus-community/windows_exporter/internal/pdh"
	"github.com/prometheus-community/windows_exporter/internal/types"
	"github.com/prometheus/client_golang/prometheus"
)

// collectorVirtualNetworkAdapterDropReasons Hyper-V Virtual Network Adapter Drop Reasons metrics
type collectorVirtualNetworkAdapterDropReasons struct {
	perfDataCollectorVirtualNetworkAdapterDropReasons *pdh.Collector
	perfDataObjectVirtualNetworkAdapterDropReasons    []perfDataCounterValuesVirtualNetworkAdapterDropReasons

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

type perfDataCounterValuesVirtualNetworkAdapterDropReasons struct {
	Name string

	VirtualNetworkAdapterDropReasonsOutgoingNativeFwdingReq          float64 `perfdata:"Outgoing NativeFwdingReq"`
	VirtualNetworkAdapterDropReasonsIncomingNativeFwdingReq          float64 `perfdata:"Incoming NativeFwdingReq"`
	VirtualNetworkAdapterDropReasonsOutgoingMTUMismatch              float64 `perfdata:"Outgoing MTUMismatch"`
	VirtualNetworkAdapterDropReasonsIncomingMTUMismatch              float64 `perfdata:"Incoming MTUMismatch"`
	VirtualNetworkAdapterDropReasonsOutgoingInvalidConfig            float64 `perfdata:"Outgoing InvalidConfig"`
	VirtualNetworkAdapterDropReasonsIncomingInvalidConfig            float64 `perfdata:"Incoming InvalidConfig"`
	VirtualNetworkAdapterDropReasonsOutgoingRequiredExtensionMissing float64 `perfdata:"Outgoing RequiredExtensionMissing"`
	VirtualNetworkAdapterDropReasonsIncomingRequiredExtensionMissing float64 `perfdata:"Incoming RequiredExtensionMissing"`
	VirtualNetworkAdapterDropReasonsOutgoingVirtualSubnetId          float64 `perfdata:"Outgoing VirtualSubnetId"`
	VirtualNetworkAdapterDropReasonsIncomingVirtualSubnetId          float64 `perfdata:"Incoming VirtualSubnetId"`
	VirtualNetworkAdapterDropReasonsOutgoingBridgeReserved           float64 `perfdata:"Outgoing BridgeReserved"`
	VirtualNetworkAdapterDropReasonsIncomingBridgeReserved           float64 `perfdata:"Incoming BridgeReserved"`
	VirtualNetworkAdapterDropReasonsOutgoingRouterGuard              float64 `perfdata:"Outgoing RouterGuard"`
	VirtualNetworkAdapterDropReasonsIncomingRouterGuard              float64 `perfdata:"Incoming RouterGuard"`
	VirtualNetworkAdapterDropReasonsOutgoingDhcpGuard                float64 `perfdata:"Outgoing DhcpGuard"`
	VirtualNetworkAdapterDropReasonsIncomingDhcpGuard                float64 `perfdata:"Incoming DhcpGuard"`
	VirtualNetworkAdapterDropReasonsOutgoingMacSpoofing              float64 `perfdata:"Outgoing MacSpoofing"`
	VirtualNetworkAdapterDropReasonsIncomingMacSpoofing              float64 `perfdata:"Incoming MacSpoofing"`
	VirtualNetworkAdapterDropReasonsOutgoingIpsec                    float64 `perfdata:"Outgoing Ipsec"`
	VirtualNetworkAdapterDropReasonsIncomingIpsec                    float64 `perfdata:"Incoming Ipsec"`
	VirtualNetworkAdapterDropReasonsOutgoingQos                      float64 `perfdata:"Outgoing Qos"`
	VirtualNetworkAdapterDropReasonsIncomingQos                      float64 `perfdata:"Incoming Qos"`
	VirtualNetworkAdapterDropReasonsOutgoingFailedPvlanSetting       float64 `perfdata:"Outgoing FailedPvlanSetting"`
	VirtualNetworkAdapterDropReasonsIncomingFailedPvlanSetting       float64 `perfdata:"Incoming FailedPvlanSetting"`
	VirtualNetworkAdapterDropReasonsOutgoingFailedSecurityPolicy     float64 `perfdata:"Outgoing FailedSecurityPolicy"`
	VirtualNetworkAdapterDropReasonsIncomingFailedSecurityPolicy     float64 `perfdata:"Incoming FailedSecurityPolicy"`
	VirtualNetworkAdapterDropReasonsOutgoingUnauthorizedMAC          float64 `perfdata:"Outgoing UnauthorizedMAC"`
	VirtualNetworkAdapterDropReasonsIncomingUnauthorizedMAC          float64 `perfdata:"Incoming UnauthorizedMAC"`
	VirtualNetworkAdapterDropReasonsOutgoingUnauthorizedVLAN         float64 `perfdata:"Outgoing UnauthorizedVLAN"`
	VirtualNetworkAdapterDropReasonsIncomingUnauthorizedVLAN         float64 `perfdata:"Incoming UnauthorizedVLAN"`
	VirtualNetworkAdapterDropReasonsOutgoingFilteredVLAN             float64 `perfdata:"Outgoing FilteredVLAN"`
	VirtualNetworkAdapterDropReasonsIncomingFilteredVLAN             float64 `perfdata:"Incoming FilteredVLAN"`
	VirtualNetworkAdapterDropReasonsOutgoingFiltered                 float64 `perfdata:"Outgoing Filtered"`
	VirtualNetworkAdapterDropReasonsIncomingFiltered                 float64 `perfdata:"Incoming Filtered"`
	VirtualNetworkAdapterDropReasonsOutgoingBusy                     float64 `perfdata:"Outgoing Busy"`
	VirtualNetworkAdapterDropReasonsIncomingBusy                     float64 `perfdata:"Incoming Busy"`
	VirtualNetworkAdapterDropReasonsOutgoingNotAccepted              float64 `perfdata:"Outgoing NotAccepted"`
	VirtualNetworkAdapterDropReasonsIncomingNotAccepted              float64 `perfdata:"Incoming NotAccepted"`
	VirtualNetworkAdapterDropReasonsOutgoingDisconnected             float64 `perfdata:"Outgoing Disconnected"`
	VirtualNetworkAdapterDropReasonsIncomingDisconnected             float64 `perfdata:"Incoming Disconnected"`
	VirtualNetworkAdapterDropReasonsOutgoingNotReady                 float64 `perfdata:"Outgoing NotReady"`
	VirtualNetworkAdapterDropReasonsIncomingNotReady                 float64 `perfdata:"Incoming NotReady"`
	VirtualNetworkAdapterDropReasonsOutgoingResources                float64 `perfdata:"Outgoing Resources"`
	VirtualNetworkAdapterDropReasonsIncomingResources                float64 `perfdata:"Incoming Resources"`
	VirtualNetworkAdapterDropReasonsOutgoingInvalidPacket            float64 `perfdata:"Outgoing InvalidPacket"`
	VirtualNetworkAdapterDropReasonsIncomingInvalidPacket            float64 `perfdata:"Incoming InvalidPacket"`
	VirtualNetworkAdapterDropReasonsOutgoingInvalidData              float64 `perfdata:"Outgoing InvalidData"`
	VirtualNetworkAdapterDropReasonsIncomingInvalidData              float64 `perfdata:"Incoming InvalidData"`
	VirtualNetworkAdapterDropReasonsOutgoingUnknown                  float64 `perfdata:"Outgoing Unknown"`
	VirtualNetworkAdapterDropReasonsIncomingUnknown                  float64 `perfdata:"Incoming Unknown"`
}

func (c *Collector) buildVirtualNetworkAdapterDropReasons() error {
	var err error

	c.perfDataCollectorVirtualNetworkAdapterDropReasons, err = pdh.NewCollector[perfDataCounterValuesVirtualNetworkAdapterDropReasons](pdh.CounterTypeRaw, "Hyper-V Virtual Network Adapter Drop Reasons", pdh.InstancesAll)
	if err != nil {
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
	err := c.perfDataCollectorVirtualNetworkAdapterDropReasons.Collect(&c.perfDataObjectVirtualNetworkAdapterDropReasons)
	if err != nil {
		return fmt.Errorf("failed to collect Hyper-V Virtual Network Adapter Drop Reasons metrics: %w", err)
	}

	for _, data := range c.perfDataObjectVirtualNetworkAdapterDropReasons {
		ch <- prometheus.MustNewConstMetric(
			c.virtualNetworkAdapterDropReasons,
			prometheus.CounterValue,
			data.VirtualNetworkAdapterDropReasonsOutgoingNativeFwdingReq,
			data.Name, "NativeFwdingReq", "outgoing",
		)
		ch <- prometheus.MustNewConstMetric(
			c.virtualNetworkAdapterDropReasons,
			prometheus.CounterValue,
			data.VirtualNetworkAdapterDropReasonsIncomingNativeFwdingReq,
			data.Name, "NativeFwdingReq", "incoming",
		)
		ch <- prometheus.MustNewConstMetric(
			c.virtualNetworkAdapterDropReasons,
			prometheus.CounterValue,
			data.VirtualNetworkAdapterDropReasonsOutgoingMTUMismatch,
			data.Name, "MTUMismatch", "outgoing",
		)
		ch <- prometheus.MustNewConstMetric(
			c.virtualNetworkAdapterDropReasons,
			prometheus.CounterValue,
			data.VirtualNetworkAdapterDropReasonsIncomingMTUMismatch,
			data.Name, "MTUMismatch", "incoming",
		)
		ch <- prometheus.MustNewConstMetric(
			c.virtualNetworkAdapterDropReasons,
			prometheus.CounterValue,
			data.VirtualNetworkAdapterDropReasonsOutgoingInvalidConfig,
			data.Name, "InvalidConfig", "outgoing",
		)
		ch <- prometheus.MustNewConstMetric(
			c.virtualNetworkAdapterDropReasons,
			prometheus.CounterValue,
			data.VirtualNetworkAdapterDropReasonsIncomingInvalidConfig,
			data.Name, "InvalidConfig", "incoming",
		)
		ch <- prometheus.MustNewConstMetric(
			c.virtualNetworkAdapterDropReasons,
			prometheus.CounterValue,
			data.VirtualNetworkAdapterDropReasonsOutgoingRequiredExtensionMissing,
			data.Name, "RequiredExtensionMissing", "outgoing",
		)
		ch <- prometheus.MustNewConstMetric(
			c.virtualNetworkAdapterDropReasons,
			prometheus.CounterValue,
			data.VirtualNetworkAdapterDropReasonsIncomingRequiredExtensionMissing,
			data.Name, "RequiredExtensionMissing", "incoming",
		)
		ch <- prometheus.MustNewConstMetric(
			c.virtualNetworkAdapterDropReasons,
			prometheus.CounterValue,
			data.VirtualNetworkAdapterDropReasonsOutgoingVirtualSubnetId,
			data.Name, "VirtualSubnetId", "outgoing",
		)
		ch <- prometheus.MustNewConstMetric(
			c.virtualNetworkAdapterDropReasons,
			prometheus.CounterValue,
			data.VirtualNetworkAdapterDropReasonsIncomingVirtualSubnetId,
			data.Name, "VirtualSubnetId", "incoming",
		)
		ch <- prometheus.MustNewConstMetric(
			c.virtualNetworkAdapterDropReasons,
			prometheus.CounterValue,
			data.VirtualNetworkAdapterDropReasonsOutgoingBridgeReserved,
			data.Name, "BridgeReserved", "outgoing",
		)
		ch <- prometheus.MustNewConstMetric(
			c.virtualNetworkAdapterDropReasons,
			prometheus.CounterValue,
			data.VirtualNetworkAdapterDropReasonsIncomingBridgeReserved,
			data.Name, "BridgeReserved", "incoming",
		)
		ch <- prometheus.MustNewConstMetric(
			c.virtualNetworkAdapterDropReasons,
			prometheus.CounterValue,
			data.VirtualNetworkAdapterDropReasonsOutgoingRouterGuard,
			data.Name, "RouterGuard", "outgoing",
		)
		ch <- prometheus.MustNewConstMetric(
			c.virtualNetworkAdapterDropReasons,
			prometheus.CounterValue,
			data.VirtualNetworkAdapterDropReasonsIncomingRouterGuard,
			data.Name, "RouterGuard", "incoming",
		)
		ch <- prometheus.MustNewConstMetric(
			c.virtualNetworkAdapterDropReasons,
			prometheus.CounterValue,
			data.VirtualNetworkAdapterDropReasonsOutgoingDhcpGuard,
			data.Name, "DhcpGuard", "outgoing",
		)
		ch <- prometheus.MustNewConstMetric(
			c.virtualNetworkAdapterDropReasons,
			prometheus.CounterValue,
			data.VirtualNetworkAdapterDropReasonsIncomingDhcpGuard,
			data.Name, "DhcpGuard", "incoming",
		)
		ch <- prometheus.MustNewConstMetric(
			c.virtualNetworkAdapterDropReasons,
			prometheus.CounterValue,
			data.VirtualNetworkAdapterDropReasonsOutgoingMacSpoofing,
			data.Name, "MacSpoofing", "outgoing",
		)
		ch <- prometheus.MustNewConstMetric(
			c.virtualNetworkAdapterDropReasons,
			prometheus.CounterValue,
			data.VirtualNetworkAdapterDropReasonsIncomingMacSpoofing,
			data.Name, "MacSpoofing", "incoming",
		)
		ch <- prometheus.MustNewConstMetric(
			c.virtualNetworkAdapterDropReasons,
			prometheus.CounterValue,
			data.VirtualNetworkAdapterDropReasonsOutgoingIpsec,
			data.Name, "Ipsec", "outgoing",
		)
		ch <- prometheus.MustNewConstMetric(
			c.virtualNetworkAdapterDropReasons,
			prometheus.CounterValue,
			data.VirtualNetworkAdapterDropReasonsIncomingIpsec,
			data.Name, "Ipsec", "incoming",
		)
		ch <- prometheus.MustNewConstMetric(
			c.virtualNetworkAdapterDropReasons,
			prometheus.CounterValue,
			data.VirtualNetworkAdapterDropReasonsOutgoingQos,
			data.Name, "Qos", "outgoing",
		)
		ch <- prometheus.MustNewConstMetric(
			c.virtualNetworkAdapterDropReasons,
			prometheus.CounterValue,
			data.VirtualNetworkAdapterDropReasonsIncomingQos,
			data.Name, "Qos", "incoming",
		)
		ch <- prometheus.MustNewConstMetric(
			c.virtualNetworkAdapterDropReasons,
			prometheus.CounterValue,
			data.VirtualNetworkAdapterDropReasonsOutgoingFailedPvlanSetting,
			data.Name, "FailedPvlanSetting", "outgoing",
		)
		ch <- prometheus.MustNewConstMetric(
			c.virtualNetworkAdapterDropReasons,
			prometheus.CounterValue,
			data.VirtualNetworkAdapterDropReasonsIncomingFailedPvlanSetting,
			data.Name, "FailedPvlanSetting", "incoming",
		)
		ch <- prometheus.MustNewConstMetric(
			c.virtualNetworkAdapterDropReasons,
			prometheus.CounterValue,
			data.VirtualNetworkAdapterDropReasonsOutgoingFailedSecurityPolicy,
			data.Name, "FailedSecurityPolicy", "outgoing",
		)
		ch <- prometheus.MustNewConstMetric(
			c.virtualNetworkAdapterDropReasons,
			prometheus.CounterValue,
			data.VirtualNetworkAdapterDropReasonsIncomingFailedSecurityPolicy,
			data.Name, "FailedSecurityPolicy", "incoming",
		)
		ch <- prometheus.MustNewConstMetric(
			c.virtualNetworkAdapterDropReasons,
			prometheus.CounterValue,
			data.VirtualNetworkAdapterDropReasonsOutgoingUnauthorizedMAC,
			data.Name, "UnauthorizedMAC", "outgoing",
		)
		ch <- prometheus.MustNewConstMetric(
			c.virtualNetworkAdapterDropReasons,
			prometheus.CounterValue,
			data.VirtualNetworkAdapterDropReasonsIncomingUnauthorizedMAC,
			data.Name, "UnauthorizedMAC", "incoming",
		)
		ch <- prometheus.MustNewConstMetric(
			c.virtualNetworkAdapterDropReasons,
			prometheus.CounterValue,
			data.VirtualNetworkAdapterDropReasonsOutgoingUnauthorizedVLAN,
			data.Name, "UnauthorizedVLAN", "outgoing",
		)
		ch <- prometheus.MustNewConstMetric(
			c.virtualNetworkAdapterDropReasons,
			prometheus.CounterValue,
			data.VirtualNetworkAdapterDropReasonsIncomingUnauthorizedVLAN,
			data.Name, "UnauthorizedVLAN", "incoming",
		)
		ch <- prometheus.MustNewConstMetric(
			c.virtualNetworkAdapterDropReasons,
			prometheus.CounterValue,
			data.VirtualNetworkAdapterDropReasonsOutgoingFilteredVLAN,
			data.Name, "FilteredVLAN", "outgoing",
		)
		ch <- prometheus.MustNewConstMetric(
			c.virtualNetworkAdapterDropReasons,
			prometheus.CounterValue,
			data.VirtualNetworkAdapterDropReasonsIncomingFilteredVLAN,
			data.Name, "FilteredVLAN", "incoming",
		)
		ch <- prometheus.MustNewConstMetric(
			c.virtualNetworkAdapterDropReasons,
			prometheus.CounterValue,
			data.VirtualNetworkAdapterDropReasonsOutgoingFiltered,
			data.Name, "Filtered", "outgoing",
		)
		ch <- prometheus.MustNewConstMetric(
			c.virtualNetworkAdapterDropReasons,
			prometheus.CounterValue,
			data.VirtualNetworkAdapterDropReasonsIncomingFiltered,
			data.Name, "Filtered", "incoming",
		)
		ch <- prometheus.MustNewConstMetric(
			c.virtualNetworkAdapterDropReasons,
			prometheus.CounterValue,
			data.VirtualNetworkAdapterDropReasonsOutgoingBusy,
			data.Name, "Busy", "outgoing",
		)
		ch <- prometheus.MustNewConstMetric(
			c.virtualNetworkAdapterDropReasons,
			prometheus.CounterValue,
			data.VirtualNetworkAdapterDropReasonsIncomingBusy,
			data.Name, "Busy", "incoming",
		)
		ch <- prometheus.MustNewConstMetric(
			c.virtualNetworkAdapterDropReasons,
			prometheus.CounterValue,
			data.VirtualNetworkAdapterDropReasonsOutgoingNotAccepted,
			data.Name, "NotAccepted", "outgoing",
		)
		ch <- prometheus.MustNewConstMetric(
			c.virtualNetworkAdapterDropReasons,
			prometheus.CounterValue,
			data.VirtualNetworkAdapterDropReasonsIncomingNotAccepted,
			data.Name, "NotAccepted", "incoming",
		)
		ch <- prometheus.MustNewConstMetric(
			c.virtualNetworkAdapterDropReasons,
			prometheus.CounterValue,
			data.VirtualNetworkAdapterDropReasonsOutgoingDisconnected,
			data.Name, "Disconnected", "outgoing",
		)
		ch <- prometheus.MustNewConstMetric(
			c.virtualNetworkAdapterDropReasons,
			prometheus.CounterValue,
			data.VirtualNetworkAdapterDropReasonsIncomingDisconnected,
			data.Name, "Disconnected", "incoming",
		)
		ch <- prometheus.MustNewConstMetric(
			c.virtualNetworkAdapterDropReasons,
			prometheus.CounterValue,
			data.VirtualNetworkAdapterDropReasonsOutgoingNotReady,
			data.Name, "NotReady", "outgoing",
		)
		ch <- prometheus.MustNewConstMetric(
			c.virtualNetworkAdapterDropReasons,
			prometheus.CounterValue,
			data.VirtualNetworkAdapterDropReasonsIncomingNotReady,
			data.Name, "NotReady", "incoming",
		)
		ch <- prometheus.MustNewConstMetric(
			c.virtualNetworkAdapterDropReasons,
			prometheus.CounterValue,
			data.VirtualNetworkAdapterDropReasonsOutgoingResources,
			data.Name, "Resources", "outgoing",
		)
		ch <- prometheus.MustNewConstMetric(
			c.virtualNetworkAdapterDropReasons,
			prometheus.CounterValue,
			data.VirtualNetworkAdapterDropReasonsIncomingResources,
			data.Name, "Resources", "incoming",
		)
		ch <- prometheus.MustNewConstMetric(
			c.virtualNetworkAdapterDropReasons,
			prometheus.CounterValue,
			data.VirtualNetworkAdapterDropReasonsOutgoingInvalidPacket,
			data.Name, "InvalidPacket", "outgoing",
		)
		ch <- prometheus.MustNewConstMetric(
			c.virtualNetworkAdapterDropReasons,
			prometheus.CounterValue,
			data.VirtualNetworkAdapterDropReasonsIncomingInvalidPacket,
			data.Name, "InvalidPacket", "incoming",
		)
		ch <- prometheus.MustNewConstMetric(
			c.virtualNetworkAdapterDropReasons,
			prometheus.CounterValue,
			data.VirtualNetworkAdapterDropReasonsOutgoingInvalidData,
			data.Name, "InvalidData", "outgoing",
		)
		ch <- prometheus.MustNewConstMetric(
			c.virtualNetworkAdapterDropReasons,
			prometheus.CounterValue,
			data.VirtualNetworkAdapterDropReasonsIncomingInvalidData,
			data.Name, "InvalidData", "incoming",
		)
		ch <- prometheus.MustNewConstMetric(
			c.virtualNetworkAdapterDropReasons,
			prometheus.CounterValue,
			data.VirtualNetworkAdapterDropReasonsOutgoingUnknown,
			data.Name, "Unknown", "outgoing",
		)
		ch <- prometheus.MustNewConstMetric(
			c.virtualNetworkAdapterDropReasons,
			prometheus.CounterValue,
			data.VirtualNetworkAdapterDropReasonsIncomingUnknown,
			data.Name, "Unknown", "incoming",
		)
	}

	return nil
}
