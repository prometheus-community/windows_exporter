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
	collection := &updateCollection[Update]{VTable: &methods[0]}

	errors := 0
	for item, err := range collection.All() {
		errors++

		require.Nil(t, item)
		require.ErrorIs(t, err, ole.HRESULT(0x80070005))
	}

	require.Equal(t, 1, errors)
	require.Zero(t, index)
}
