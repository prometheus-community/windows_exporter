// SPDX-License-Identifier: Apache-2.0
//
// Copyright The Prometheus Authors
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
