// SPDX-License-Identifier: Apache-2.0
//
// Copyright The Prometheus Authors
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

package iphlpapi

import (
	"encoding/binary"
	"fmt"

	"github.com/go-ole/go-ole"
)

// MIB_TCPROW_OWNER_PID structure for IPv4.
// https://learn.microsoft.com/en-us/windows/win32/api/tcpmib/ns-tcpmib-mib_tcprow_owner_pid
//
//nolint:unused
type MIB_TCPROW_OWNER_PID struct {
	dwState      MIB_TCP_STATE
	dwLocalAddr  BigEndianUint32
	dwLocalPort  BigEndianUint32
	dwRemoteAddr BigEndianUint32
	dwRemotePort BigEndianUint32
	dwOwningPid  uint32
}

// MIB_TCP6ROW_OWNER_PID structure for IPv6.
// https://learn.microsoft.com/en-us/windows/win32/api/tcpmib/ns-tcpmib-mib_tcp6row_owner_pid
//
//nolint:unused
type MIB_TCP6ROW_OWNER_PID struct {
	ucLocalAddr     [16]byte
	dwLocalScopeId  uint32
	dwLocalPort     BigEndianUint32
	ucRemoteAddr    [16]byte
	dwRemoteScopeId uint32
	dwRemotePort    BigEndianUint32
	dwState         MIB_TCP_STATE
	dwOwningPid     uint32
}

type MIB_TCP_STATE uint32

const (
	_ MIB_TCP_STATE = iota
	TCPStateClosed
	TCPStateListening
	TCPStateSynSent
	TCPStateSynRcvd
	TCPStateEstablished
	TCPStateFinWait1
	TCPStateFinWait2
	TCPStateCloseWait
	TCPStateClosing
	TCPStateLastAck
	TCPStateTimeWait
	TCPStateDeleteTcb
)

func (state MIB_TCP_STATE) String() string {
	switch state {
	case TCPStateClosed:
		return "CLOSED"
	case TCPStateListening:
		return "LISTENING"
	case TCPStateSynSent:
		return "SYN_SENT"
	case TCPStateSynRcvd:
		return "SYN_RECEIVED"
	case TCPStateEstablished:
		return "ESTABLISHED"
	case TCPStateFinWait1:
		return "FIN_WAIT1"
	case TCPStateFinWait2:
		return "FIN_WAIT2"
	case TCPStateCloseWait:
		return "CLOSE_WAIT"
	case TCPStateClosing:
		return "CLOSING"
	case TCPStateLastAck:
		return "LAST_ACK"
	case TCPStateTimeWait:
		return "TIME_WAIT"
	case TCPStateDeleteTcb:
		return "DELETE_TCB"
	default:
		return fmt.Sprintf("UNKNOWN_%d", state)
	}
}

type BigEndianUint32 uint32

func (b BigEndianUint32) uint16() uint16 {
	data := make([]byte, 2)
	binary.BigEndian.PutUint16(data, uint16(b))

	return binary.LittleEndian.Uint16(data)
}

// Constants from Windows headers
const (
	IF_MAX_STRING_SIZE         = 256
	IF_MAX_PHYS_ADDRESS_LENGTH = 32
)

// MIB_IF_ROW2 represents network interface statistics
type MIB_IF_ROW2 struct {
	InterfaceLuid               uint64
	InterfaceIndex              uint32
	InterfaceGuid               ole.GUID
	Alias                       [IF_MAX_STRING_SIZE + 1]uint16
	Description                 [IF_MAX_STRING_SIZE + 1]uint16
	PhysicalAddressLength       uint32
	PhysicalAddress             [IF_MAX_PHYS_ADDRESS_LENGTH]byte
	PermanentPhysicalAddress    [IF_MAX_PHYS_ADDRESS_LENGTH]byte
	Mtu                         uint32
	Type                        uint32
	TunnelType                  uint32
	MediaType                   uint32
	PhysicalMediumType          uint32
	AccessType                  uint32
	DirectionType               uint32
	InterfaceAndOperStatusFlags uint8
	OperStatus                  uint32
	AdminStatus                 uint32
	MediaConnectState           uint32
	NetworkGuid                 [16]byte
	ConnectionType              uint32
	TransmitLinkSpeed           uint64
	ReceiveLinkSpeed            uint64
	InOctets                    uint64
	InUcastPkts                 uint64
	InNUcastPkts                uint64
	InDiscards                  uint64
	InErrors                    uint64
	InUnknownProtos             uint64
	InUcastOctets               uint64
	InMulticastOctets           uint64
	InBroadcastOctets           uint64
	OutOctets                   uint64
	OutUcastPkts                uint64
	OutNUcastPkts               uint64
	OutDiscards                 uint64
	OutErrors                   uint64
	OutUcastOctets              uint64
	OutMulticastOctets          uint64
	OutBroadcastOctets          uint64
	OutQLen                     uint64
}
