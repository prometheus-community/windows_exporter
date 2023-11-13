//go:build windows

package winversion

import (
	"fmt"
	"strconv"

	"golang.org/x/sys/windows/registry"
)

var WindowsVersion string
var WindowsVersionFloat float64

func init() {
	var err error
	WindowsVersion, WindowsVersionFloat, err = GetWindowsVersion()
	if err != nil {
		panic(err)
	}
}

// GetWindowsVersion reads the version number of the OS from the Registry
// See https://docs.microsoft.com/en-us/windows/desktop/sysinfo/operating-system-version
func GetWindowsVersion() (string, float64, error) {
	reg, err := registry.OpenKey(registry.LOCAL_MACHINE, `SOFTWARE\Microsoft\Windows NT\CurrentVersion`, registry.QUERY_VALUE)
	if err != nil {
		return "", 0, fmt.Errorf("couldn't open registry: %w", err)
	}

	defer reg.Close()

	windowsVersion, _, err := reg.GetStringValue("CurrentVersion")
	if err != nil {
		return "", 0, fmt.Errorf("couldn't open registry: %w", err)
	}

	windowsVersionFloat, err := strconv.ParseFloat(windowsVersion, 64)
	if err != nil {
		return "", 0, fmt.Errorf("couldn't open registry: %w", err)
	}

	return windowsVersion, windowsVersionFloat, nil
}
