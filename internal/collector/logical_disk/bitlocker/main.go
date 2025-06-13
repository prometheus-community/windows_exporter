package main

import (
	"fmt"
	"runtime"

	"github.com/go-ole/go-ole"
	"github.com/prometheus-community/windows_exporter/internal/headers/propsys"
	"github.com/prometheus-community/windows_exporter/internal/headers/shell32"
)

func main() {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	if err := ole.CoInitializeEx(0, ole.COINIT_APARTMENTTHREADED); err != nil {
		panic(err)
	}
	defer ole.CoUninitialize()

	var pkey propsys.PROPERTYKEY
	if err := propsys.PSGetPropertyKeyFromName("System.Volume.BitLockerProtection", &pkey); err != nil {
		panic(err)
	}

	item, err := shell32.SHCreateItemFromParsingName("C:")
	if err != nil {
		panic(err)
	}

	defer item.Release()

	var v ole.VARIANT

	if err := item.GetProperty(&pkey, &v); err != nil {
		panic(err)
	}
	defer v.Clear()

	fmt.Printf("BitLocker status for C: %d\n", v.Val)
}
