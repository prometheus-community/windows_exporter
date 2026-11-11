// SPDX-License-Identifier: MIT
//
// Copyright The Prometheus Authors

//go:build windows

package types

import "errors"

var (
	ErrCollectorNotInitialized = errors.New("collector not initialized")
	ErrNoData                  = errors.New("no data")
	ErrNoDataUnexpected        = errors.New("no data")
)
