package main

import (
	"fmt"

	"github.com/Microsoft/go-winio/pkg/process"
	"github.com/prometheus-community/windows_exporter/internal/headers/iphlpapi"
	"golang.org/x/sys/windows"
)

func main() {
	pid, err := iphlpapi.GetOwnerPIDOfTCPPort(windows.AF_INET, 135)
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
}
