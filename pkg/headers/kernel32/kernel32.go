package kernel32

import (
	"syscall"
	"unsafe"
)

var (
	kernel32 = syscall.NewLazyDLL("kernel32.dll")

	procGetDynamicTimeZoneInformationSys = kernel32.NewProc("GetDynamicTimeZoneInformation")
)

// SYSTEMTIME contains a date and time.
// 📑 https://docs.microsoft.com/en-us/windows/win32/api/minwinbase/ns-minwinbase-systemtime
type SYSTEMTIME struct {
	WYear         uint16
	WMonth        uint16
	WDayOfWeek    uint16
	WDay          uint16
	WHour         uint16
	WMinute       uint16
	WSecond       uint16
	WMilliseconds uint16
}

// DynamicTimezoneInformation contains the current dynamic daylight time settings.
// 📑 https://docs.microsoft.com/en-us/windows/win32/api/timezoneapi/ns-timezoneapi-dynamic_time_zone_information
type DynamicTimezoneInformation struct {
	Bias                        int32
	standardName                [32]uint16
	StandardDate                SYSTEMTIME
	StandardBias                int32
	DaylightName                [32]uint16
	DaylightDate                SYSTEMTIME
	DaylightBias                int32
	TimeZoneKeyName             [128]uint16
	DynamicDaylightTimeDisabled uint8 // BOOLEAN
}

// GetDynamicTimeZoneInformation retrieves the current dynamic daylight time settings.
// 📑 https://docs.microsoft.com/en-us/windows/win32/api/timezoneapi/nf-timezoneapi-getdynamictimezoneinformation
func GetDynamicTimeZoneInformation() (DynamicTimezoneInformation, error) {
	var tzi DynamicTimezoneInformation

	r0, _, err := syscall.SyscallN(procGetDynamicTimeZoneInformationSys.Addr(), uintptr(unsafe.Pointer(&tzi)))
	if uint32(r0) == 0xffffffff {
		return tzi, err
	}

	return tzi, nil
}
