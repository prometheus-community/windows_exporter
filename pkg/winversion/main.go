//go:build windows

package winversion

import (
	"fmt"
	"strconv"
	"sync"

	"golang.org/x/sys/windows/registry"
)

var WindowsVersionFloat = sync.OnceValue[float64](func() float64 {
	version, err := getWindowsVersion()
	if err != nil {
		panic(err)
	}

	return version
})

// GetWindowsVersion reads the version number of the OS from the Registry
// See https://docs.microsoft.com/en-us/windows/desktop/sysinfo/operating-system-version
func getWindowsVersion() (float64, error) {
	reg, err := registry.OpenKey(registry.LOCAL_MACHINE, `SOFTWARE\Microsoft\Windows NT\CurrentVersion`, registry.QUERY_VALUE)
	if err != nil {
		return 0, fmt.Errorf("couldn't open registry: %w", err)
	}

	defer reg.Close()

	windowsVersion, _, err := reg.GetStringValue("CurrentVersion")
	if err != nil {
		return 0, fmt.Errorf("couldn't open registry: %w", err)
	}

	windowsVersionFloat, err := strconv.ParseFloat(windowsVersion, 64)
	if err != nil {
		return 0, fmt.Errorf("couldn't open registry: %w", err)
	}

	return windowsVersionFloat, nil
}
