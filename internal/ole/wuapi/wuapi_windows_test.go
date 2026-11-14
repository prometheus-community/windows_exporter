// SPDX-License-Identifier: MIT
//
// Copyright The Prometheus Authors

//go:build windows && (amd64 || arm64)

package wuapi

import (
	"testing"

	"github.com/prometheus-community/windows_exporter/internal/ole"
	"github.com/stretchr/testify/require"
	"golang.org/x/sys/windows"
)

func TestNativeUpdateCollection(t *testing.T) {
	var methods [11]uintptr

	// IUpdateCollection::get_Count slot from the Windows SDK.
	methods[10] = windows.NewCallback(func(_ uintptr, out *int32) uintptr {
		*out = 1

		return 0
	})

	var index uintptr

	// IUpdateCollection::get_Item slot from the Windows SDK.
	methods[7] = windows.NewCallback(func(_ uintptr, i uintptr, _ **Update) uintptr {
		index = i

		return 0x80070005
	})
	collection := &UpdateCollection{VTable: &methods[0]}

	errors := 0
	for item, err := range collection.All() {
		errors++

		require.Nil(t, item)
		require.ErrorIs(t, err, ole.HRESULT(0x80070005))
	}

	require.Equal(t, 1, errors)
	require.Zero(t, index)
}
