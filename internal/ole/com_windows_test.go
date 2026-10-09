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
	"golang.org/x/sys/windows"
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
	for _, value := range []string{"", "windows_exporter", "before\x00after", "Unicode \U0001F600", "Grüße"} {
		t.Run(value, func(t *testing.T) {
			str, err := newBSTR(value)
			require.NoError(t, err)

			defer str.free()

			require.Equal(t, value, str.string())
		})
	}

	require.Empty(t, (bstr{}).string())
}

func TestNativeGetterCallbackStackGrowth(t *testing.T) {
	for _, name := range []string{"Get", "GetArg"} {
		t.Run(name, func(t *testing.T) {
			var methods [8]uintptr

			if name == "Get" {
				methods[7] = windows.NewCallback(func(_ uintptr, out *int32) uintptr {
					// A native caller retains the original address even when the
					// callback grows and relocates the caller's Go stack.
					growCallbackStack(64)

					*out = 1234

					return 0
				})
			} else {
				methods[7] = windows.NewCallback(func(_ uintptr, arg uintptr, out *int32) uintptr {
					growCallbackStack(64)

					*out = int32(arg)

					return 0
				})
			}

			object := &Object{VTable: &methods[0]}

			var (
				value int32
				err   error
			)

			if name == "Get" {
				value, err = object.Get[int32](7)
			} else {
				value, err = object.GetArg[int32](7, 1234)
			}

			require.NoError(t, err)
			require.Equal(t, int32(1234), value)
		})
	}
}

type nativeTestObject struct{ Object }

func TestNativeObjectGetters(t *testing.T) {
	tests := []struct {
		name   string
		status uintptr
		value  *nativeTestObject
	}{
		{name: "success", value: &nativeTestObject{}},
		{name: "S_FALSE", status: 1, value: &nativeTestObject{}},
		{name: "null success"},
		{name: "null S_FALSE", status: 1},
		{name: "HRESULT failure", status: 0x80070005},
	}

	for _, getter := range []string{"GetObject", "GetObjectArg", "GetObjectStringArg"} {
		t.Run(getter, func(t *testing.T) {
			for _, test := range tests {
				t.Run(test.name, func(t *testing.T) {
					var methods [8]uintptr

					if getter == "GetObject" {
						methods[7] = windows.NewCallback(func(_ uintptr, out **nativeTestObject) uintptr {
							*out = test.value

							return test.status
						})
					} else {
						methods[7] = windows.NewCallback(func(_ uintptr, _ uintptr, out **nativeTestObject) uintptr {
							*out = test.value

							return test.status
						})
					}

					object := &Object{VTable: &methods[0]}

					var (
						value *nativeTestObject
						err   error
					)

					switch getter {
					case "GetObject":
						value, err = object.GetObject[nativeTestObject](7)
					case "GetObjectArg":
						value, err = object.GetObjectArg[nativeTestObject](7, 1)
					case "GetObjectStringArg":
						value, err = object.GetObjectStringArg[nativeTestObject](7, "criteria")
					}

					if int32(test.status) < 0 {
						require.Nil(t, value)
						require.ErrorIs(t, err, HRESULT(test.status))

						return
					}

					if test.value == nil {
						require.Nil(t, value)
						require.ErrorContains(t, err, "returned a nil interface")

						return
					}

					require.NoError(t, err)
					require.Same(t, test.value, value)
				})
			}
		})
	}
}

func TestNativeObjectGetterCallbackStackGrowth(t *testing.T) {
	for _, getter := range []string{"GetObject", "GetObjectArg", "GetObjectStringArg"} {
		t.Run(getter, func(t *testing.T) {
			var methods [8]uintptr

			expected := &nativeTestObject{}
			inputValid := true

			switch getter {
			case "GetObject":
				methods[7] = windows.NewCallback(func(_ uintptr, out **nativeTestObject) uintptr {
					growCallbackStack(64)

					*out = expected

					return 0
				})
			case "GetObjectArg":
				methods[7] = windows.NewCallback(func(_ uintptr, input *int32, out **nativeTestObject) uintptr {
					growCallbackStack(64)

					inputValid = *input == 1234
					*out = expected

					return 0
				})
			case "GetObjectStringArg":
				methods[7] = windows.NewCallback(func(_ uintptr, input *uint16, out **nativeTestObject) uintptr {
					growCallbackStack(64)

					inputValid = (bstr{ptr: input}).string() == "before\x00after"
					*out = expected

					return 0
				})
			}

			object := &Object{VTable: &methods[0]}

			var (
				value *nativeTestObject
				err   error
			)

			switch getter {
			case "GetObject":
				value, err = object.GetObject[nativeTestObject](7)
			case "GetObjectArg":
				input := int32(1234)
				value, err = object.GetObjectArg[nativeTestObject](7, uintptr(unsafe.Pointer(&input)))
			case "GetObjectStringArg":
				value, err = object.GetObjectStringArg[nativeTestObject](7, "before\x00after")
			}

			require.NoError(t, err)
			require.Same(t, expected, value)
			require.True(t, inputValid)
		})
	}
}

//go:noinline
func growCallbackStack(depth int) byte {
	var padding [4096]byte
	for i := range padding {
		padding[i] = byte(i + depth)
	}

	if depth > 0 {
		padding[0] = growCallbackStack(depth - 1)
	}

	runtime.KeepAlive(&padding)

	return padding[0] + padding[len(padding)-1]
}

func TestVariantLayout(t *testing.T) {
	require.Equal(t, uintptr(24), unsafe.Sizeof(Variant{}))
	require.Equal(t, uintptr(8), unsafe.Offsetof(Variant{}.Value))
}
