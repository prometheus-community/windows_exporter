package netapi32

import (
	"errors"
	"unsafe"

	"golang.org/x/sys/windows"
)

// WKSTAInfo102 is a wrapper of WKSTA_Info_102
//https://docs.microsoft.com/en-us/windows/win32/api/lmwksta/ns-lmwksta-wksta_info_102
type wKSTAInfo102 struct {
	wki102_platform_id     uint32
	wki102_computername    *uint16
	wki102_langroup        *uint16
	wki102_ver_major       uint32
	wki102_ver_minor       uint32
	wki102_lanroot         *uint16
	wki102_logged_on_users uint32
}

// WorkstationInfo is an idiomatic wrapper of WKSTAInfo102
type WorkstationInfo struct {
	PlatformId    uint32
	ComputerName  string
	LanGroup      string
	VersionMajor  uint32
	VersionMinor  uint32
	LanRoot       string
	LoggedOnUsers uint32
}

var (
	netapi32             = windows.NewLazySystemDLL("netapi32")
	procNetWkstaGetInfo  = netapi32.NewProc("NetWkstaGetInfo")
	procNetApiBufferFree = netapi32.NewProc("NetApiBufferFree")
)

// NetApiStatus is a map of Network Management Error Codes.
// https://docs.microsoft.com/en-gb/windows/win32/netmgmt/network-management-error-codes?redirectedfrom=MSDN
var NetApiStatus = map[uint32]string{
	// Success
	0: "NERR_Success",
	// This computer name is invalid.
	2351: "NERR_InvalidComputer",
	// This operation is only allowed on the primary domain controller of the domain.
	2226: "NERR_NotPrimary",
	/// This operation is not allowed on this special group.
	2234: "NERR_SpeGroupOp",
	/// This operation is not allowed on the last administrative account.
	2452: "NERR_LastAdmin",
	/// The password parameter is invalid.
	2203: "NERR_BadPassword",
	/// The password does not meet the password policy requirements.
	/// Check the minimum password length, password complexity and password history requirements.
	2245: "NERR_PasswordTooShort",
	/// The user name could not be found.
	2221: "NERR_UserNotFound",
	// Errors
	5:    "ERROR_ACCESS_DENIED",
	8:    "ERROR_NOT_ENOUGH_MEMORY",
	87:   "ERROR_INVALID_PARAMETER",
	123:  "ERROR_INVALID_NAME",
	124:  "ERROR_INVALID_LEVEL",
	234:  "ERROR_MORE_DATA",
	1219: "ERROR_SESSION_CREDENTIAL_CONFLICT",
}

// NetApiBufferFree frees the memory other network management functions use internally to return information.
// https://docs.microsoft.com/en-us/windows/win32/api/lmapibuf/nf-lmapibuf-netapibufferfree
func netApiBufferFree(buffer *wKSTAInfo102) {
	procNetApiBufferFree.Call(uintptr(unsafe.Pointer(buffer))) //nolint:errcheck
}

// NetWkstaGetInfo returns information about the configuration of a workstation.
// https://docs.microsoft.com/en-us/windows/win32/api/lmwksta/nf-lmwksta-netwkstagetinfo
func netWkstaGetInfo() (wKSTAInfo102, uint32, error) {
	var lpwi *wKSTAInfo102
	pLevel := uintptr(102)

	r1, _, _ := procNetWkstaGetInfo.Call(0, pLevel, uintptr(unsafe.Pointer(&lpwi)))
	defer netApiBufferFree(lpwi)

	if ret := *(*uint32)(unsafe.Pointer(&r1)); ret != 0 {
		return wKSTAInfo102{}, ret, errors.New(NetApiStatus[ret])
	}

	deref := *lpwi
	return deref, 0, nil
}

// GetWorkstationInfo is an idiomatic wrapper for netWkstaGetInfo
func GetWorkstationInfo() (WorkstationInfo, error) {
	info, _, err := netWkstaGetInfo()
	if err != nil {
		return WorkstationInfo{}, err
	}
	workstationInfo := WorkstationInfo{
		PlatformId:    info.wki102_platform_id,
		ComputerName:  windows.UTF16PtrToString(info.wki102_computername),
		LanGroup:      windows.UTF16PtrToString(info.wki102_langroup),
		VersionMajor:  info.wki102_ver_major,
		VersionMinor:  info.wki102_ver_minor,
		LanRoot:       windows.UTF16PtrToString(info.wki102_lanroot),
		LoggedOnUsers: info.wki102_logged_on_users,
	}
	return workstationInfo, nil
}
