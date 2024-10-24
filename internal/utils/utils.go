//go:build windows

package utils

import "golang.org/x/sys/windows"

func MilliSecToSec(t float64) float64 {
	return t / 1000
}

func BoolToFloat(b bool) float64 {
	if b {
		return 1.0
	}

	return 0.0
}

func ToPTR[t any](v t) *t {
	return &v
}

// MustUTF16PtrFromString converts a string to a UTF-16 pointer at initialization time.
//
//nolint:ireturn
func MustUTF16PtrFromString[T ~*uint16](s string) T {
	val, err := windows.UTF16PtrFromString(s)
	if err != nil {
		panic(err)
	}

	return val
}
