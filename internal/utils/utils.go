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

func ToPTR[t any](v t) *t {
	return &v
}

// Must panics if the error is not nil.
//
//nolint:ireturn
func Must[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}

	return v
}
