//go:build windows

package secur32

import (
	"errors"
	"fmt"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"
)

// based on https://github.com/carlpett/winlsa/blob/master/winlsa.go

var (
	secur32  = windows.NewLazySystemDLL("Secur32.dll")
	advapi32 = windows.NewLazySystemDLL("advapi32.dll")

	procLsaEnumerateLogonSessions = secur32.NewProc("LsaEnumerateLogonSessions")
	procLsaGetLogonSessionData    = secur32.NewProc("LsaGetLogonSessionData")
	procLsaFreeReturnBuffer       = secur32.NewProc("LsaFreeReturnBuffer")
	procLsaNtStatusToWinError     = advapi32.NewProc("LsaNtStatusToWinError")
)

func GetLogonSessions() ([]*LogonSessionData, error) {
	var (
		buffer       uintptr
		sessionCount uint32
	)

	err := LsaEnumerateLogonSessions(&sessionCount, &buffer)
	if err != nil {
		return nil, err
	}

	if buffer != 0 {
		defer func(buffer uintptr) {
			_ = LsaFreeReturnBuffer(buffer)
		}(buffer)
	}

	sizeLUID := unsafe.Sizeof(windows.LUID{})

	sessionDataSlice := make([]*LogonSessionData, 0, sessionCount)

	for i := range sessionCount {
		curPtr := unsafe.Pointer(buffer + (uintptr(i) * sizeLUID))
		luid := (*windows.LUID)(curPtr)

		sessionData, err := GetLogonSessionData(luid)
		if err != nil {
			if errors.Is(err, windows.ERROR_ACCESS_DENIED) {
				// Skip logon sessions that we don't have access to
				continue
			}

			return nil, err
		}

		sessionDataSlice = append(sessionDataSlice, sessionData)
	}

	return sessionDataSlice, nil
}

func GetLogonSessionData(luid *windows.LUID) (*LogonSessionData, error) {
	var dataBuffer *SECURITY_LOGON_SESSION_DATA
	if err := LsaGetLogonSessionData(luid, &dataBuffer); err != nil {
		return nil, fmt.Errorf("failed to get logon session data: %w", err)
	}

	defer func(buffer uintptr) {
		_ = LsaFreeReturnBuffer(buffer)
	}(uintptr(unsafe.Pointer(dataBuffer)))

	return newLogonSessionData(dataBuffer), nil
}

func LsaEnumerateLogonSessions(sessionCount *uint32, sessions *uintptr) error {
	r0, _, _ := procLsaEnumerateLogonSessions.Call(uintptr(unsafe.Pointer(sessionCount)), uintptr(unsafe.Pointer(sessions)))

	return LsaNtStatusToWinError(r0)
}

func LsaGetLogonSessionData(luid *windows.LUID, ppLogonSessionData **SECURITY_LOGON_SESSION_DATA) error {
	r0, _, _ := procLsaGetLogonSessionData.Call(uintptr(unsafe.Pointer(luid)), uintptr(unsafe.Pointer(ppLogonSessionData)))

	return LsaNtStatusToWinError(r0)
}

func LsaFreeReturnBuffer(buffer uintptr) error {
	r0, _, _ := procLsaFreeReturnBuffer.Call(buffer)

	return LsaNtStatusToWinError(r0)
}

func LsaNtStatusToWinError(ntstatus uintptr) error {
	r0, _, err := procLsaNtStatusToWinError.Call(ntstatus)

	switch {
	case errors.Is(err, windows.ERROR_SUCCESS):
		if r0 == 0 {
			return nil
		}
	case errors.Is(err, windows.ERROR_MR_MID_NOT_FOUND):
		return fmt.Errorf("unknown LSA NTSTATUS code %x", ntstatus)
	}

	return windows.Errno(r0)
}

func newLogonSessionData(data *SECURITY_LOGON_SESSION_DATA) *LogonSessionData {
	return &LogonSessionData{
		LogonId:     data.LogonId,
		UserName:    data.UserName.String(),
		LogonDomain: data.LogonDomain.String(),
		LogonType:   data.LogonType,
		LogonTime:   time.Unix(0, data.LogonTime.Nanoseconds()),
	}
}
