package iphlpapi

import "fmt"

// MIB_TCPROW_OWNER_PID structure for IPv4.
// https://learn.microsoft.com/en-us/windows/win32/api/tcpmib/ns-tcpmib-mib_tcprow_owner_pid
type MIB_TCPROW_OWNER_PID struct {
	dwState      MIB_TCP_STATE
	dwLocalAddr  uint32
	dwLocalPort  uint32
	dwRemoteAddr uint32
	dwRemotePort uint32
	dwOwningPid  uint32
}

// MIB_TCP6ROW_OWNER_PID structure for IPv6.
// https://learn.microsoft.com/en-us/windows/win32/api/tcpmib/ns-tcpmib-mib_tcp6row_owner_pid
type MIB_TCP6ROW_OWNER_PID struct {
	ucLocalAddr     [16]byte
	dwLocalScopeId  uint32
	dwLocalPort     uint32
	ucRemoteAddr    [16]byte
	dwRemoteScopeId uint32
	dwRemotePort    uint32
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
