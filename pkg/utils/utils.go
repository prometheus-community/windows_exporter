//go:build windows

package utils

func MilliSecToSec(t float64) float64 {
	return t / 1000
}

func BoolToFloat(b bool) float64 {
	if b {
		return 1.0
	}
	return 0.0
}

func HasValue(v *string) bool {
	return !IsEmpty(v)
}

func IsEmpty(v *string) bool {
	return v == nil || *v == ""
}
