//go:build windows

package wtsapi32

import (
	"errors"
	"fmt"
	"log/slog"
	"unsafe"

	"golang.org/x/sys/windows"
)

type WTSTypeClass int

// The valid values for the WTSTypeClass enumeration.
const (
	WTSTypeProcessInfoLevel0 WTSTypeClass = iota
	WTSTypeProcessInfoLevel1
	WTSTypeSessionInfoLevel1
)

type WTSConnectState uint32

const (
	// wtsActive A user is logged on to the WinStation. This state occurs when a user is signed in and actively connected to the device.
	wtsActive WTSConnectState = iota
	// wtsConnected The WinStation is connected to the client.
	wtsConnected
	// wtsConnectQuery The WinStation is in the process of connecting to the client.
	wtsConnectQuery
	// wtsShadow The WinStation is shadowing another WinStation.
	wtsShadow
	// wtsDisconnected The WinStation is active but the client is disconnected.
	// This state occurs when a user is signed in but not actively connected to the device, such as when the user has chosen to exit to the lock screen.
	wtsDisconnected
	// wtsIdle The WinStation is waiting for a client to connect.
	wtsIdle
	// wtsListen The WinStation is listening for a connection. A listener session waits for requests for new client connections.
	// No user is logged on a listener session. A listener session cannot be reset, shadowed, or changed to a regular client session.
	wtsListen
	// wtsReset The WinStation is being reset.
	wtsReset
	// wtsDown The WinStation is down due to an error.
	wtsDown
	// wtsInit The WinStation is initializing.
	wtsInit
)

// WTSSessionInfo1w contains information about a session on a Remote Desktop Session Host (RD Session Host) server.
// docs: https://docs.microsoft.com/en-us/windows/win32/api/wtsapi32/ns-wtsapi32-wts_session_info_1w
type wtsSessionInfo1 struct {
	// ExecEnvID An identifier that uniquely identifies the session within the list of sessions returned by the WTSEnumerateSessionsEx function.
	ExecEnvID uint32
	// State A value of the WTSConnectState enumeration type that specifies the connection state of a Remote Desktop Services session.
	State uint32
	// SessionID A session identifier assigned by the RD Session Host server, RD Virtualization Host server, or virtual machine.
	SessionID uint32
	// pSessionName A pointer to a null-terminated string that contains the name of this session. For example, "services", "console", or "RDP-Tcp#0".
	pSessionName *uint16
	// pHostName A pointer to a null-terminated string that contains the name of the computer that the session is running on.
	// If the session is running directly on an RD Session Host server or RD Virtualization Host server, the string contains NULL.
	// If the session is running on a virtual machine, the string contains the name of the virtual machine.
	pHostName *uint16
	// pUserName A pointer to a null-terminated string that contains the name of the user who is logged on to the session.
	// If no user is logged on to the session, the string contains NULL.
	pUserName *uint16
	// pDomainName A pointer to a null-terminated string that contains the domain name of the user who is logged on to the session.
	// If no user is logged on to the session, the string contains NULL.
	pDomainName *uint16
	// pFarmName A pointer to a null-terminated string that contains the name of the farm that the virtual machine is joined to.
	// If the session is not running on a virtual machine that is joined to a farm, the string contains NULL.
	pFarmName *uint16
}

type WTSSession struct {
	ExecEnvID   uint32
	State       WTSConnectState
	SessionID   uint32
	SessionName string
	HostName    string
	UserName    string
	DomainName  string
	FarmName    string
}

var (
	wtsapi32 = windows.NewLazySystemDLL("wtsapi32.dll")

	procWTSOpenServerEx        = wtsapi32.NewProc("WTSOpenServerExW")
	procWTSEnumerateSessionsEx = wtsapi32.NewProc("WTSEnumerateSessionsExW")
	procWTSFreeMemoryEx        = wtsapi32.NewProc("WTSFreeMemoryExW")
	procWTSCloseServer         = wtsapi32.NewProc("WTSCloseServer")

	WTSSessionStates = map[WTSConnectState]string{
		wtsActive:       "active",
		wtsConnected:    "connected",
		wtsConnectQuery: "connect_query",
		wtsShadow:       "shadow",
		wtsDisconnected: "disconnected",
		wtsIdle:         "idle",
		wtsListen:       "listen",
		wtsReset:        "reset",
		wtsDown:         "down",
		wtsInit:         "init",
	}
)

func WTSOpenServer(server string) (windows.Handle, error) {
	var (
		err        error
		serverName *uint16
	)

	if server != "" {
		serverName, err = windows.UTF16PtrFromString(server)
		if err != nil {
			return windows.InvalidHandle, err
		}
	}

	r1, _, err := procWTSOpenServerEx.Call(uintptr(unsafe.Pointer(serverName)))
	serverHandle := windows.Handle(r1)

	if serverHandle == windows.InvalidHandle {
		return windows.InvalidHandle, err
	}

	return serverHandle, nil
}

func WTSCloseServer(server windows.Handle) error {
	r1, _, err := procWTSCloseServer.Call(uintptr(server))

	if r1 != 1 && !errors.Is(err, windows.ERROR_SUCCESS) {
		return fmt.Errorf("failed to close server: %w", err)
	}

	return nil
}

func WTSFreeMemoryEx(class WTSTypeClass, pMemory uintptr, numberOfEntries uint32) error {
	r1, _, err := procWTSFreeMemoryEx.Call(
		uintptr(class),
		pMemory,
		uintptr(numberOfEntries),
	)

	if r1 != 1 {
		return fmt.Errorf("failed to free memory: %w", err)
	}

	return nil
}

func WTSEnumerateSessionsEx(server windows.Handle, logger *slog.Logger) ([]WTSSession, error) {
	var sessionInfoPointer uintptr

	var count uint32

	pLevel := uint32(1)
	r1, _, err := procWTSEnumerateSessionsEx.Call(
		uintptr(server),
		uintptr(unsafe.Pointer(&pLevel)),
		uintptr(0),
		uintptr(unsafe.Pointer(&sessionInfoPointer)),
		uintptr(unsafe.Pointer(&count)),
	)

	if r1 != 1 {
		return nil, err
	}

	if sessionInfoPointer != 0 {
		defer func(class WTSTypeClass, pMemory uintptr, NumberOfEntries uint32) {
			if err := WTSFreeMemoryEx(class, pMemory, NumberOfEntries); err != nil {
				logger.Warn("failed to free memory", "err", fmt.Errorf("WTSEnumerateSessionsEx: %w", err))
			}
		}(WTSTypeSessionInfoLevel1, sessionInfoPointer, count)
	}

	var sizeTest wtsSessionInfo1
	sessionSize := unsafe.Sizeof(sizeTest)

	sessions := make([]WTSSession, 0, count)

	for i := range count {
		curPtr := unsafe.Pointer(sessionInfoPointer + (uintptr(i) * sessionSize))
		data := (*wtsSessionInfo1)(curPtr)

		sessionInfo := WTSSession{
			ExecEnvID:   data.ExecEnvID,
			State:       WTSConnectState(data.State),
			SessionID:   data.SessionID,
			SessionName: windows.UTF16PtrToString(data.pSessionName),
			HostName:    windows.UTF16PtrToString(data.pHostName),
			UserName:    windows.UTF16PtrToString(data.pUserName),
			DomainName:  windows.UTF16PtrToString(data.pDomainName),
			FarmName:    windows.UTF16PtrToString(data.pFarmName),
		}
		sessions = append(sessions, sessionInfo)
	}

	return sessions, nil
}
