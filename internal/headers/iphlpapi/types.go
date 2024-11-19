//go:build windows

package iphlpapi

import (
	"encoding/binary"
	"fmt"
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
