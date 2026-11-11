// SPDX-License-Identifier: MIT
//
// Copyright The Prometheus Authors

//go:build windows

package types

type Collector[T any] interface {
	Collect(dst *[]T) error
	Close()
}
