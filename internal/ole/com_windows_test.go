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

package ole

import (
	"runtime"
	"testing"
	"unsafe"

	"github.com/stretchr/testify/require"
)

func TestInitialize(t *testing.T) {
	runtime.LockOSThread()

	defer runtime.UnlockOSThread()

	require.NoError(t, Initialize())

	defer Uninitialize()

	// Verify the apartment selected by Initialize, rather than just its status.
	var apartment, qualifier int32

	hr, _, _ := ole32.NewProc("CoGetApartmentType").Call(
		uintptr(unsafe.Pointer(&apartment)),
		uintptr(unsafe.Pointer(&qualifier)),
	)
	require.NoError(t, ResultError(hr))
	require.Equal(t, int32(1), apartment, "APTTYPE_MTA")

	// The second initialization returns S_FALSE; it still needs balancing.
	require.NoError(t, Initialize())

	defer Uninitialize()
}

func TestBSTR(t *testing.T) {
	for _, value := range []string{"", "windows_exporter", "before\x00after", "Unicode \U0001F600"} {
		t.Run(value, func(t *testing.T) {
			str, err := newBSTR(value)
			require.NoError(t, err)

			defer str.free()

			require.Equal(t, value, str.string())
		})
	}

	require.Empty(t, (bstr{}).string())
}

func TestVariantLayout(t *testing.T) {
	require.Equal(t, uintptr(24), unsafe.Sizeof(Variant{}))
	require.Equal(t, uintptr(8), unsafe.Offsetof(Variant{}.Value))
}
