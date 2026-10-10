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
	"reflect"
	"strings"
	"sync"
	"syscall"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"
)

// OperationOptionsTimeout is the key for the timeout option.
//
// https://github.com/microsoft/win32metadata/blob/527806d20d83d3abd43d16cd3fa8795d8deba343/generation/WinSDK/RecompiledIdlHeaders/um/mi.h#L9240
//
//nolint:gochecknoglobals
var OperationOptionsTimeout = UTF16PtrFromString[*uint16]("__MI_OPERATIONOPTIONS_TIMEOUT")

// OperationFlags represents the flags for an operation.
//
// https://learn.microsoft.com/en-us/previous-versions/windows/desktop/wmi_v2/mi-flags
type OperationFlags uint32

const (
	OperationFlagsStandardRTTI OperationFlags = 0x0800
)

// Operation represents an operation.
// https://learn.microsoft.com/en-us/windows/win32/api/mi/ns-mi-mi_operation
type Operation struct {
	reserved1 uint64
	reserved2 uintptr
	ft        *OperationFT

	// completed is set once GetInstance has returned the final result. It is
	// Go-only state after the MI_Operation fields, which MI never writes past;
	// every Operation is allocated in Go.
	completed bool
}

// OperationFT represents the function table for Operation.
// https://learn.microsoft.com/en-us/windows/win32/api/mi/ns-mi-mi_operationft
type OperationFT struct {
	Close         uintptr
	Cancel        uintptr
	GetSession    uintptr
	GetInstance   uintptr
	GetIndication uintptr
	GetClass      uintptr
}

type OperationOptions struct {
	reserved1 uint64
	reserved2 uintptr
	ft        *OperationOptionsFT
}

type OperationOptionsFT struct {
	Delete             uintptr
	SetString          uintptr
	SetNumber          uintptr
	SetCustomOption    uintptr
	GetString          uintptr
	GetNumber          uintptr
	GetOptionCount     uintptr
	GetOptionAt        uintptr
	GetOption          uintptr
	GetEnabledChannels uintptr
	Clone              uintptr
	SetInterval        uintptr
	GetInterval        uintptr
}

type OperationCallbacks[T any] struct {
	CallbackContext         *T
	PromptUser              uintptr
	WriteError              uintptr
	WriteMessage            uintptr
	WriteProgress           uintptr
	InstanceResult          uintptr
	IndicationResult        uintptr
	ClassResult             uintptr
	StreamedParameterResult uintptr
}

// Close closes an operation handle.
//
// https://learn.microsoft.com/en-us/windows/win32/api/mi/nf-mi-mi_operation_close
func (o *Operation) Close() error {
	if o == nil || o.ft == nil {
		return ErrNotInitialized
	}

	// MI_Operation_Close blocks until the final result has been delivered, so
	// pending results must be read first. Reading them would cost the full
	// remaining result set, so the operation is cancelled, which makes the
	// final result arrive right away. A fully read operation needs neither.
	if !o.completed {
		_ = o.Cancel()

		for !o.completed {
			if _, _, err := o.GetInstance(); err != nil {
				break
			}
		}
	}

	r0, _, _ := syscall.SyscallN(o.ft.Close, uintptr(unsafe.Pointer(o)))

	if result := ResultError(r0); !errors.Is(result, MI_RESULT_OK) {
		return result
	}

	return nil
}

func (o *Operation) Cancel() error {
	if o == nil || o.ft == nil {
		return ErrNotInitialized
	}

	r0, _, _ := syscall.SyscallN(o.ft.Cancel, uintptr(unsafe.Pointer(o)), 0)

	if result := ResultError(r0); !errors.Is(result, MI_RESULT_OK) {
		return result
	}

	return nil
}

func (o *Operation) GetInstance() (*Instance, bool, error) {
	if o == nil || o.ft == nil {
		return nil, false, ErrNotInitialized
	}

	var (
		instance          *Instance
		errorDetails      *Instance
		moreResults       Boolean
		instanceResult    ResultError
		errorMessageUTF16 *uint16
	)

	r0, _, _ := syscall.SyscallN(
		o.ft.GetInstance,
		uintptr(unsafe.Pointer(o)),
		uintptr(unsafe.Pointer(&instance)),
		uintptr(unsafe.Pointer(&moreResults)),
		uintptr(unsafe.Pointer(&instanceResult)),
		uintptr(unsafe.Pointer(&errorMessageUTF16)),
		uintptr(unsafe.Pointer(&errorDetails)),
	)

	//nolint:nestif
	if !errors.Is(instanceResult, MI_RESULT_OK) {
		errorMessage := strings.TrimSpace(windows.UTF16PtrToString(errorMessageUTF16))

		// We need a language neutral way to detect an operation timeout, because MI_RESULT_OPERATION_TIMED_OUT
		// is not returned by the API, but instead we get MI_RESULT_INVALID_OPERATION_TIMEOUT with a specific error code
		// in the error details.
		if errorDetails != nil {
			count, _ := errorDetails.GetElementCount()
			if count != 0 {
				errorCodeRaw, err := errorDetails.GetElement("error_Code")
				if err == nil {
					errorCodeValue, _ := errorCodeRaw.GetValue()

					errorCode, ok := errorCodeValue.(uint32)
					if ok && errorCode == 262148 {
						instanceResult = MI_RESULT_INVALID_OPERATION_TIMEOUT
						errorMessage = ""
					}
				}
			}
		}

		if errorMessage != "" {
			errorMessage = fmt.Sprintf(" (%s)", errorMessage)
		}

		// A failed result is always the final one.
		o.completed = true

		return nil, false, fmt.Errorf("instance result: %w%s", instanceResult, errorMessage)
	}

	if result := ResultError(r0); !errors.Is(result, MI_RESULT_OK) {
		return nil, false, result
	}

	if moreResults != True {
		o.completed = true
	}

	return instance, moreResults == True, nil
}

func (o *Operation) Unmarshal[T any](dst *[]T) error {
	if o == nil || o.ft == nil {
		return ErrNotInitialized
	}

	fields, err := prepareUnmarshal(dst)
	if err != nil {
		return err
	}

	return o.unmarshal(dst, fields, false)
}

// miField is a struct field that is populated from the MI element named by its `mi` tag.
type miField struct {
	index int
	tag   string
	name  ElementName
}

// miFieldsCache maps a struct type to its []miField. Collectors query the same
// types on every scrape, so the tags are parsed and converted to UTF-16 once.
//
//nolint:gochecknoglobals
var miFieldsCache sync.Map

// prepareUnmarshal checks that dst is non-nil and T is a struct, resets *dst to empty
// and returns the `mi`-tagged fields of T. Callers run it before starting a query,
// so invalid input is rejected without a WMI round trip.
func prepareUnmarshal[T any](dst *[]T) ([]miField, error) {
	if dst == nil {
		return nil, ErrInvalidEntityType
	}

	fields, err := miFieldsOf[T]()
	if err != nil {
		return nil, err
	}

	*dst = (*dst)[:0]

	return fields, nil
}

// miFieldsOf returns the fields of T that carry an `mi` tag.
// It returns ErrInvalidEntityType if T is not a struct.
// The returned slice is shared and must not be modified.
func miFieldsOf[T any]() ([]miField, error) {
	structType := reflect.TypeFor[T]()

	if fields, ok := miFieldsCache.Load(structType); ok {
		return fields.([]miField), nil //nolint:forcetypeassert
	}

	if structType.Kind() != reflect.Struct {
		return nil, ErrInvalidEntityType
	}

	fields := make([]miField, 0, structType.NumField())

	for i := range structType.NumField() {
		miTag := structType.Field(i).Tag.Get("mi")
		if miTag == "" {
			continue
		}

		name, err := NewElementName(miTag)
		if err != nil {
			return nil, fmt.Errorf("invalid mi tag %q: %w", miTag, err)
		}

		fields = append(fields, miField{index: i, tag: miTag, name: name})
	}

	miFieldsCache.Store(structType, fields)

	return fields, nil
}

// unmarshal iterates over the operation's instances and appends them to dst.
// fields must come from prepareUnmarshal.
// skipMissing controls how missing elements are handled (see unmarshalInstance).
func (o *Operation) unmarshal[T any](dst *[]T, fields []miField, skipMissing bool) error {
	for {
		instance, moreResults, err := o.GetInstance()
		if err != nil {
			return fmt.Errorf("failed to get instance: %w", err)
		}

		// If WMI returns nil, it means there are no more results.
		if instance == nil {
			break
		}

		counter, err := instance.GetElementCount()
		if err != nil {
			return fmt.Errorf("failed to get element count: %w", err)
		}

		if counter == 0 {
			break
		}

		var elem T

		if err := unmarshalInstance(instance, fields, reflect.ValueOf(&elem).Elem(), skipMissing); err != nil {
			return err
		}

		*dst = append(*dst, elem)

		if !moreResults {
			break
		}
	}

	return nil
}

func (o *OperationOptions) SetTimeout(timeout time.Duration) error {
	if o == nil || o.ft == nil {
		return ErrNotInitialized
	}

	r0, _, _ := syscall.SyscallN(
		o.ft.SetInterval,
		uintptr(unsafe.Pointer(o)),
		uintptr(unsafe.Pointer(OperationOptionsTimeout)),
		uintptr(unsafe.Pointer(NewInterval(timeout))),
		0,
	)

	if result := ResultError(r0); !errors.Is(result, MI_RESULT_OK) {
		return result
	}

	return nil
}

// Delete deletes the operation options. MI_OperationOptions_Delete returns
// void, so the only error is ErrNotInitialized.
//
// https://learn.microsoft.com/en-us/windows/win32/api/mi/nf-mi-mi_operationoptions_delete
func (o *OperationOptions) Delete() error {
	if o == nil || o.ft == nil {
		return ErrNotInitialized
	}

	_, _, _ = syscall.SyscallN(o.ft.Delete, uintptr(unsafe.Pointer(o)))

	return nil
}
