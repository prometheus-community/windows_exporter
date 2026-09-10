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

//go:build windows

package mi

import (
	"errors"
	"fmt"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

type Instance struct {
	ft         *InstanceFT
	classDecl  *ClassDecl
	serverName *uint16
	nameSpace  *uint16
	_          [4]uintptr
}

type InstanceFT struct {
	Clone           uintptr
	Destruct        uintptr
	Delete          uintptr
	IsA             uintptr
	GetClassName    uintptr
	SetNameSpace    uintptr
	GetNameSpace    uintptr
	GetElementCount uintptr
	AddElement      uintptr
	SetElement      uintptr
	SetElementAt    uintptr
	GetElement      uintptr
	GetElementAt    uintptr
	ClearElement    uintptr
	ClearElementAt  uintptr
	GetServerName   uintptr
	SetServerName   uintptr
	GetClass        uintptr
}

type ClassDecl struct {
	Flags          uint32
	Code           uint32
	Name           *uint16
	Mqualifiers    uintptr
	NumQualifiers  uint32
	Mproperties    uintptr
	NumProperties  uint32
	Size           uint32
	SuperClass     *uint16
	SuperClassDecl uintptr
	Methods        uintptr
	NumMethods     uint32

	Schema      uintptr
	ProviderFT  uintptr
	OwningClass uintptr
}

func (instance *Instance) Delete() error {
	if instance == nil || instance.ft == nil {
		return ErrNotInitialized
	}

	r0, _, _ := syscall.SyscallN(instance.ft.Delete, uintptr(unsafe.Pointer(instance)))

	if result := ResultError(r0); !errors.Is(result, MI_RESULT_OK) {
		return result
	}

	return nil
}

func (instance *Instance) GetElement(elementName string) (*Element, error) {
	if instance == nil || instance.ft == nil {
		return nil, ErrNotInitialized
	}

	elementNameUTF16, err := windows.UTF16PtrFromString(elementName)
	if err != nil {
		return nil, fmt.Errorf("failed to convert element name %s to UTF-16: %w", elementName, err)
	}

	// MI_Value is a union whose largest members (MI_Datetime, the array
	// descriptors) exceed a single machine word. Provide a full-width buffer so
	// the provider cannot write past it and corrupt the stack when the element
	// is an array or datetime. Word 0 holds the scalar/pointer/array-data; word
	// 1 holds the array element count for array types.
	var (
		valueBuf  [4]uint64
		valueType ValueType
	)

	r0, _, _ := syscall.SyscallN(
		instance.ft.GetElement,
		uintptr(unsafe.Pointer(instance)),
		uintptr(unsafe.Pointer(elementNameUTF16)),
		uintptr(unsafe.Pointer(&valueBuf)),
		uintptr(unsafe.Pointer(&valueType)),
		0,
		0,
	)

	if result := ResultError(r0); !errors.Is(result, MI_RESULT_OK) {
		return nil, result
	}

	return &Element{
		value:     uintptr(valueBuf[0]),
		arrayLen:  uint32(valueBuf[1]),
		valueType: valueType,
	}, nil
}

func (instance *Instance) GetElementCount() (uint32, error) {
	if instance == nil || instance.ft == nil {
		return 0, ErrNotInitialized
	}

	var count uint32

	r0, _, _ := syscall.SyscallN(
		instance.ft.GetElementCount,
		uintptr(unsafe.Pointer(instance)),
		uintptr(unsafe.Pointer(&count)),
	)

	if result := ResultError(r0); !errors.Is(result, MI_RESULT_OK) {
		return 0, result
	}

	return count, nil
}

func (instance *Instance) GetClassName() (string, error) {
	if instance == nil || instance.ft == nil {
		return "", ErrNotInitialized
	}

	var classNameUTF16 *uint16

	r0, _, _ := syscall.SyscallN(
		instance.ft.GetClassName,
		uintptr(unsafe.Pointer(instance)),
		uintptr(unsafe.Pointer(&classNameUTF16)),
	)

	if result := ResultError(r0); !errors.Is(result, MI_RESULT_OK) {
		return "", result
	}

	if classNameUTF16 == nil {
		return "", errors.New("class name is nil")
	}

	return windows.UTF16PtrToString(classNameUTF16), nil
}
