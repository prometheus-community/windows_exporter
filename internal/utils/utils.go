//go:build windows

package utils

func MilliSecToSec(t float64) float64 {
	return t / 1000
}

func MBToBytes(mb float64) float64 {
	return mb * 1024 * 1024
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

func PercentageToRatio(percentage float64) float64 {
	return percentage / 100
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

// SplitError returns a slice of errors from the given error. It reverses the [errors.Join] function.
func SplitError(err error) []error {
	if errs, ok := err.(interface{ Unwrap() []error }); ok {
		return errs.Unwrap()
	}

	return []error{err}
}
