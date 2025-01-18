package win32api

import "golang.org/x/sys/windows"

type DATE_TIME = windows.Filetime
type DWORD = uint32
type LPWSTR struct {
	*uint16
}

func (s LPWSTR) String() string {
	return windows.UTF16PtrToString(s.uint16)
}
