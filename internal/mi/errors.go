// SPDX-License-Identifier: MIT
//
// Copyright The Prometheus Authors

//go:build windows

package mi

import "errors"

var (
	ErrNotInitialized     = errors.New("not initialized")
	ErrInvalidEntityType  = errors.New("invalid entity type")
	ErrInvalidElementName = errors.New("invalid element name")
	// ErrNullValue is returned by the converters of an element that has no
	// value. Check [Element.IsNull] first to tell it apart from a real value.
	ErrNullValue = errors.New("element value is null")
)
