// SPDX-License-Identifier: MIT
//
// Copyright The Prometheus Authors

//go:build windows

package types

import "regexp"

var (
	RegExpAny   = regexp.MustCompile("^.+$")
	RegExpEmpty = regexp.MustCompile("^$")
)
