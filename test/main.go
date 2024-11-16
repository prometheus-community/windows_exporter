package main

import (
	"fmt"
	"syscall"
	"unsafe"

	"github.com/Microsoft/go-winio/pkg/process"
	"github.com/prometheus-community/windows_exporter/internal/headers/iphlpapi"
	"golang.org/x/sys/windows"
)

func main() {
	pid, err := iphlpapi.GetOwnerPIDOfTCPPort(windows.AF_INET, 1433)
	if err != nil {
		panic(err)
	}

	hProcess, err := windows.OpenProcess(windows.PROCESS_QUERY_LIMITED_INFORMATION, false, pid)
	if err != nil {
		panic(err)
	}

	cmdLine, err := process.QueryFullProcessImageName(hProcess, process.ImageNameFormatWin32Path)
	if err != nil {
		panic(err)
	}
	fmt.Println(cmdLine)

	// Load the file version information
	size, err := windows.GetFileVersionInfoSize(cmdLine, nil)
	if err != nil {
		panic(err)
	}

	data := make([]byte, size)
	err = windows.GetFileVersionInfo(cmdLine, 0, uint32(len(data)), unsafe.Pointer(&data[0]))
	if err != nil {
		panic(err)
	}

	var verSize uint32
	var verData *byte

	err = windows.VerQueryValue(unsafe.Pointer(&data[0]), `\StringFileInfo\040904b0\ProductVersion`, unsafe.Pointer(&verData), &verSize)
	if err != nil {
		panic(err)
	}

	version := syscall.UTF16ToString((*[1 << 16]uint16)(unsafe.Pointer(verData))[:verSize])
	fmt.Println(version)
}
