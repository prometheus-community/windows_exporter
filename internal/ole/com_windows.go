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
	"errors"
	"fmt"
	"runtime"
	"syscall"
	"unicode/utf16"
	"unsafe"

	"golang.org/x/sys/windows"
)

//nolint:gochecknoglobals
var (
	ole32             = windows.NewLazySystemDLL("ole32.dll")
	oleaut32          = windows.NewLazySystemDLL("oleaut32.dll")
	coInitializeEx    = ole32.NewProc("CoInitializeEx")
	coUninitialize    = ole32.NewProc("CoUninitialize")
	coCreateInstance  = ole32.NewProc("CoCreateInstance")
	sysAllocStringLen = oleaut32.NewProc("SysAllocStringLen")
	sysFreeString     = oleaut32.NewProc("SysFreeString")
	sysStringLen      = oleaut32.NewProc("SysStringLen")
)

// Initialize initializes an STA and disables legacy OLE 1.0 DDE. The caller
// must lock its OS thread before calling and release all objects before
// Uninitialize on that same thread. S_FALSE is successful initialization.
func Initialize() error {
	hr, _, _ := coInitializeEx.Call(0, 0x2|0x4)
	if err := resultError(hr); err != nil {
		return fmt.Errorf("initialize COM: %w", err)
	}

	return nil
}

func Uninitialize() {
	// CoUninitialize returns void; GetLastError is irrelevant.
	_, _, _ = coUninitialize.Call()
}

// object is the common native COM interface layout. It must stay the first
// field of every interface wrapper; never copy it or move it between apartments.
type object struct{ vtable *uintptr }

func (o *object) method(slot uintptr) uintptr {
	return *(*uintptr)(unsafe.Add(unsafe.Pointer(o.vtable), slot*unsafe.Sizeof(uintptr(0))))
}

// Release relinquishes an owned interface reference. Do not call it on borrowed
// iterator items. Like IUnknown::Release, it does not return an HRESULT.
func (o *object) Release() {
	_, _, _ = syscall.SyscallN(o.method(2), uintptr(unsafe.Pointer(o)))
	runtime.KeepAlive(o)
}

// get and getArg use generic methods for scalar and interface output values.
// All pointer conversions stay in the syscall expression so the Go runtime
// keeps their storage alive and stable during the call.
//
//nolint:ireturn // T is a native scalar or interface output, not a Go interface.
func (o *object) get[T any](slot uintptr) (T, error) {
	var value T

	hr, _, _ := syscall.SyscallN(o.method(slot), uintptr(unsafe.Pointer(o)), uintptr(unsafe.Pointer(&value)))
	runtime.KeepAlive(o)

	return value, resultError(hr)
}

//nolint:ireturn // T is a native scalar or interface output, not a Go interface.
func (o *object) getArg[T any](slot, arg uintptr) (T, error) {
	var value T

	hr, _, _ := syscall.SyscallN(
		o.method(slot),
		uintptr(unsafe.Pointer(o)),
		arg,
		uintptr(unsafe.Pointer(&value)),
	)
	runtime.KeepAlive(o)

	return value, resultError(hr)
}

func (o *object) put(slot, arg uintptr) error {
	hr, _, _ := syscall.SyscallN(o.method(slot), uintptr(unsafe.Pointer(o)), arg)
	runtime.KeepAlive(o)

	return resultError(hr)
}

func create[T any](class, iid windows.GUID) (*T, error) {
	var value *T

	hr, _, _ := coCreateInstance.Call(
		uintptr(unsafe.Pointer(&class)),
		0,
		0x1|0x4, // CLSCTX_INPROC_SERVER | CLSCTX_LOCAL_SERVER
		uintptr(unsafe.Pointer(&iid)),
		uintptr(unsafe.Pointer(&value)),
	)
	if err := resultError(hr); err != nil {
		return nil, fmt.Errorf("create COM instance: %w", err)
	}

	if value == nil {
		return nil, errors.New("create COM instance returned a nil interface")
	}

	return value, nil
}

type bstr struct{ ptr *uint16 }

func newBSTR(value string) (bstr, error) {
	units := utf16.Encode([]rune(value))
	// Keep a non-nil input buffer even for an empty string.
	units = append(units, 0)

	p, _, _ := sysAllocStringLen.Call(uintptr(unsafe.Pointer(&units[0])), uintptr(len(units)-1))
	if p == 0 {
		return bstr{}, fmt.Errorf("allocate BSTR: %w", HRESULT(0x8007000e))
	}

	return bstr{ptr: (*uint16)(unsafe.Pointer(p))}, nil
}

func (b bstr) free() {
	// SysFreeString returns void and accepts a nil BSTR.
	_, _, _ = sysFreeString.Call(uintptr(unsafe.Pointer(b.ptr)))
}

func (b bstr) string() string {
	if b.ptr == nil {
		return ""
	}

	n, _, _ := sysStringLen.Call(uintptr(unsafe.Pointer(b.ptr)))

	return string(utf16.Decode(unsafe.Slice(b.ptr, int(n))))
}

func (o *object) string(slot uintptr) (string, error) {
	value, err := o.get[bstr](slot)
	defer value.free()

	if err != nil {
		return "", err
	}

	return value.string(), nil
}

// variant is the 24-byte VARIANT layout on both supported Windows architectures.
// Native 64-bit ABIs pass this aggregate by address. Only VT_EMPTY and VT_I4
// are needed for local Connect and one-based Task Scheduler collection indices.
type variant struct {
	vt       uint16
	reserved [3]uint16
	value    int64
	padding  [8]byte
}

// The x64 ABI requires caller-owned aggregate arguments to be 16-byte aligned.
// Go only guarantees eight-byte alignment for VARIANT, so reserve padding and
// align explicitly. The same storage is also valid on arm64.
type variantStorage [39]byte

func (s *variantStorage) variant() *variant {
	offset := -uintptr(unsafe.Pointer(s)) & 15

	return (*variant)(unsafe.Add(unsafe.Pointer(s), offset))
}
