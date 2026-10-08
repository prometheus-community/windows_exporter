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

// Initialize initializes the MTA and disables legacy OLE 1.0 DDE. The caller
// must lock its OS thread before calling and release all objects before
// Uninitialize on that same thread. S_FALSE is successful initialization.
func Initialize() error {
	hr, _, _ := coInitializeEx.Call(0, windows.COINIT_MULTITHREADED|windows.COINIT_DISABLE_OLE1DDE)
	if err := ResultError(hr); err != nil {
		return fmt.Errorf("initialize COM: %w", err)
	}

	return nil
}

func Uninitialize() {
	// CoUninitialize returns void; GetLastError is irrelevant.
	_, _, _ = coUninitialize.Call()
}

// Call invokes a native COM method and returns its primary result. COM methods
// report HRESULTs or reference counts directly; GetLastError is irrelevant.
// Convert pointers to uintptr in the call expression so uintptrescapes moves
// their storage to the heap, where reentrant Go callbacks cannot invalidate
// native pointers by growing and moving the goroutine stack.
//
//go:uintptrescapes
func Call(fn uintptr, args ...uintptr) uintptr {
	result, _, _ := syscall.SyscallN(fn, args...)

	return result
}

// Object is the common native COM interface layout. VTable points to the native
// function table and is owned by COM. Object must be the first field of every
// interface wrapper; never copy it or move it between apartments.
type Object struct{ VTable *uintptr }

// Method returns the address of the zero-based native vtable slot.
func (o *Object) Method(slot uintptr) uintptr {
	return *(*uintptr)(unsafe.Add(unsafe.Pointer(o.VTable), slot*unsafe.Sizeof(uintptr(0))))
}

// Release relinquishes an owned interface reference. Do not call it on borrowed
// iterator items. Like IUnknown::Release, it does not return an HRESULT.
func (o *Object) Release() {
	_ = Call(o.Method(2), uintptr(unsafe.Pointer(o)))
	runtime.KeepAlive(o)
}

// Get calls a native getter. T must exactly match its scalar or interface
// output type. Pointer conversions stay in the Call expression so their storage
// remains alive and stable during native calls and reentrant Go callbacks.
//
//nolint:ireturn // T is a native scalar or interface output, not a Go interface.
func (o *Object) Get[T any](slot uintptr) (T, error) {
	var value T

	hr := Call(o.Method(slot), uintptr(unsafe.Pointer(o)), uintptr(unsafe.Pointer(&value)))
	runtime.KeepAlive(o)

	return value, ResultError(hr)
}

// GetArg calls a native getter with one input argument. T must exactly match
// its native output type; arg must use the native argument representation.
//
//nolint:ireturn // T is a native scalar or interface output, not a Go interface.
func (o *Object) GetArg[T any](slot, arg uintptr) (T, error) {
	var value T

	hr := Call(
		o.Method(slot),
		uintptr(unsafe.Pointer(o)),
		arg,
		uintptr(unsafe.Pointer(&value)),
	)
	runtime.KeepAlive(o)

	return value, ResultError(hr)
}

// Put calls a native setter with a scalar argument.
func (o *Object) Put(slot, arg uintptr) error {
	hr := Call(o.Method(slot), uintptr(unsafe.Pointer(o)), arg)
	runtime.KeepAlive(o)

	return ResultError(hr)
}

// Create activates class and returns an owned interface reference for iid.
// T must be the matching native interface wrapper, with Object as its first field.
func Create[T any](class, iid windows.GUID) (*T, error) {
	var value *T

	hr, _, _ := coCreateInstance.Call(
		uintptr(unsafe.Pointer(&class)),
		0,
		0x1|0x4, // CLSCTX_INPROC_SERVER | CLSCTX_LOCAL_SERVER
		uintptr(unsafe.Pointer(&iid)),
		uintptr(unsafe.Pointer(&value)),
	)
	if err := ResultError(hr); err != nil {
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

// String calls a BSTR getter, copies its value, and frees the native string.
func (o *Object) String(slot uintptr) (string, error) {
	value, err := o.Get[bstr](slot)
	defer value.free()

	if err != nil {
		return "", err
	}

	return value.string(), nil
}

// Variant is the 24-byte VARIANT layout on both supported Windows architectures.
// Native 64-bit ABIs pass this aggregate by address. Only VT_EMPTY and VT_I4
// are needed for local Connect and one-based Task Scheduler collection indices.
type Variant struct {
	Type     uint16
	reserved [3]uint16
	Value    int64
	padding  [8]byte
}

// The x64 ABI requires caller-owned aggregate arguments to be 16-byte aligned.
// Go only guarantees eight-byte alignment for VARIANT, so reserve padding and
// align explicitly. The same storage is also valid on arm64.
type variantStorage [39]byte

func (s *variantStorage) variant() *Variant {
	offset := -uintptr(unsafe.Pointer(s)) & 15

	return (*Variant)(unsafe.Add(unsafe.Pointer(s), offset))
}

// GetStringArg calls a method with one BSTR input and a native output value.
// The temporary BSTR is freed before returning.
//
//nolint:ireturn // T is a native scalar or interface output, not a Go interface.
func (o *Object) GetStringArg[T any](slot uintptr, input string) (T, error) {
	value, err := newBSTR(input)
	if err != nil {
		var zero T

		return zero, err
	}
	defer value.free()

	return o.GetArg[T](slot, uintptr(unsafe.Pointer(value.ptr)))
}

// PutString calls a setter with a BSTR input, freeing the temporary string.
func (o *Object) PutString(slot uintptr, input string) error {
	value, err := newBSTR(input)
	if err != nil {
		return err
	}
	defer value.free()

	return o.Put(slot, uintptr(unsafe.Pointer(value.ptr)))
}

// NewEmptyVariant returns caller-owned, 16-byte-aligned VT_EMPTY storage.
// It contains no resources requiring VariantClear.
func NewEmptyVariant() *Variant { return new(variantStorage).variant() }

// NewInt32Variant returns caller-owned, 16-byte-aligned VT_I4 storage.
// It contains no resources requiring VariantClear.
func NewInt32Variant(value int32) *Variant {
	v := NewEmptyVariant()
	v.Type = 3
	v.Value = int64(value)

	return v
}
