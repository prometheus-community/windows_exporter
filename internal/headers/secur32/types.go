//go:build windows

package secur32

import (
	"fmt"
	"time"

	"golang.org/x/sys/windows"
)

type LogonType uint32

type LSA_LAST_INTER_LOGON_INFO struct {
	LastSuccessfulLogon                        windows.Filetime
	LastFailedLogon                            windows.Filetime
	FailedAttemptCountSinceLastSuccessfulLogon uint32
}

type SECURITY_LOGON_SESSION_DATA struct {
	Size                  uint32
	LogonId               LUID
	UserName              windows.NTUnicodeString
	LogonDomain           windows.NTUnicodeString
	AuthenticationPackage windows.NTUnicodeString
	LogonType             LogonType
	Session               uint32
	Sid                   *windows.SID
	LogonTime             windows.Filetime
	LogonServer           windows.NTUnicodeString
	DnsDomainName         windows.NTUnicodeString
	Upn                   windows.NTUnicodeString
	UserFlags             uint32
	LastLogonInfo         LSA_LAST_INTER_LOGON_INFO
	LogonScript           windows.NTUnicodeString
	ProfilePath           windows.NTUnicodeString
	HomeDirectory         windows.NTUnicodeString
	HomeDirectoryDrive    windows.NTUnicodeString
	LogoffTime            windows.Filetime
	KickOffTime           windows.Filetime
	PasswordLastSet       windows.Filetime
	PasswordCanChange     windows.Filetime
	PasswordMustChange    windows.Filetime
}

const (
	// LogonTypeSystem Not explicitly defined in LSA, but according to
	// https://docs.microsoft.com/en-us/windows/win32/cimwin32prov/win32-logonsession,
	// LogonType=0 is "Used only by the System account."
	LogonTypeSystem LogonType = iota
	_                         // LogonType=1 is not used
	LogonTypeInteractive
	LogonTypeNetwork
	LogonTypeBatch
	LogonTypeService
	LogonTypeProxy
	LogonTypeUnlock
	LogonTypeNetworkCleartext
	LogonTypeNewCredentials
	LogonTypeRemoteInteractive
	LogonTypeCachedInteractive
	LogonTypeCachedRemoteInteractive
	LogonTypeCachedUnlock
)

func (lt LogonType) String() string {
	switch lt {
	case LogonTypeSystem:
		return "System"
	case LogonTypeInteractive:
		return "Interactive"
	case LogonTypeNetwork:
		return "Network"
	case LogonTypeBatch:
		return "Batch"
	case LogonTypeService:
		return "Service"
	case LogonTypeProxy:
		return "Proxy"
	case LogonTypeUnlock:
		return "Unlock"
	case LogonTypeNetworkCleartext:
		return "NetworkCleartext"
	case LogonTypeNewCredentials:
		return "NewCredentials"
	case LogonTypeRemoteInteractive:
		return "RemoteInteractive"
	case LogonTypeCachedInteractive:
		return "CachedInteractive"
	case LogonTypeCachedRemoteInteractive:
		return "CachedRemoteInteractive"
	case LogonTypeCachedUnlock:
		return "CachedUnlock"
	default:
		return fmt.Sprintf("Undefined LogonType(%d)", lt)
	}
}

type LogonSessionData struct {
	LogonId               LUID
	UserName              string
	LogonDomain           string
	AuthenticationPackage string
	LogonType             LogonType
	Session               uint32
	Sid                   *windows.SID
	LogonTime             time.Time
}

type LUID windows.LUID

func (l LUID) String() string {
	return fmt.Sprintf("0x%x:0x%x", l.HighPart, l.LowPart)
}
