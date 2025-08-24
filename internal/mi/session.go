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
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

// Session represents a session.
//
// https://learn.microsoft.com/en-us/windows/win32/api/mi/ns-mi-mi_session
type Session struct {
	reserved1 uint64
	reserved2 uintptr
	ft        *SessionFT

	defaultOperationOptions *OperationOptions
}

// SessionFT represents the function table for Session.
//
// https://learn.microsoft.com/en-us/windows/win32/api/mi/ns-mi-mi_session
type SessionFT struct {
	Close               uintptr
	GetApplication      uintptr
	GetInstance         uintptr
	ModifyInstance      uintptr
	CreateInstance      uintptr
	DeleteInstance      uintptr
	Invoke              uintptr
	EnumerateInstances  uintptr
	QueryInstances      uintptr
	AssociatorInstances uintptr
	ReferenceInstances  uintptr
	Subscribe           uintptr
	GetClass            uintptr
	EnumerateClasses    uintptr
	TestConnection      uintptr
}

// Close closes a session and releases all associated memory.
//
// https://learn.microsoft.com/en-us/windows/win32/api/mi/nf-mi-mi_session_close
func (s *Session) Close() error {
	if s == nil || s.ft == nil {
		return ErrNotInitialized
	}

	if s.defaultOperationOptions != nil {
		_ = s.defaultOperationOptions.Delete()
	}

	r0, _, _ := syscall.SyscallN(s.ft.Close,
		uintptr(unsafe.Pointer(s)),
		0,
		0,
	)

	if result := ResultError(r0); !errors.Is(result, MI_RESULT_OK) {
		return result
	}

	return nil
}

// TestConnection queries instances. It is used to test the connection.
// The function returns an operation that can be used to retrieve the result with [Operation.GetInstance]. The operation must be closed with [Operation.Close].
// The instance returned by [Operation.GetInstance] is always nil.
//
// https://learn.microsoft.com/en-us/windows/win32/api/mi/nf-mi-mi_session_testconnection
func (s *Session) TestConnection() error {
	if s == nil || s.ft == nil {
		return ErrNotInitialized
	}

	operation := &Operation{}

	// ref: https://github.com/KurtDeGreeff/omi/blob/9caa55032a1070a665e14fd282a091f6247d13c3/Unix/scriptext/py/PMI_Session.c#L92-L105
	r0, _, _ := syscall.SyscallN(
		s.ft.TestConnection,
		uintptr(unsafe.Pointer(s)),
		0,
		0,
		uintptr(unsafe.Pointer(operation)),
	)

	if result := ResultError(r0); !errors.Is(result, MI_RESULT_OK) {
		return result
	}

	if _, _, err := operation.GetInstance(); err != nil {
		return fmt.Errorf("failed to get instance: %w", err)
	}

	if err := operation.Close(); err != nil {
		return fmt.Errorf("failed to close operation: %w", err)
	}

	return nil
}

// GetApplication gets the Application handle that was used to create the specified session.
//
// https://learn.microsoft.com/en-us/windows/win32/api/mi/nf-mi-mi_session_getapplication
func (s *Session) GetApplication() (*Application, error) {
	if s == nil || s.ft == nil {
		return nil, ErrNotInitialized
	}

	application := &Application{}

	r0, _, _ := syscall.SyscallN(
		s.ft.GetApplication,
		uintptr(unsafe.Pointer(s)),
		uintptr(unsafe.Pointer(application)),
	)

	if result := ResultError(r0); !errors.Is(result, MI_RESULT_OK) {
		return nil, result
	}

	return application, nil
}

// QueryInstances queries for a set of instances based on a query expression.
//
// https://learn.microsoft.com/en-us/windows/win32/api/mi/nf-mi-mi_session_queryinstances
func (s *Session) QueryInstances(flags OperationFlags, operationOptions *OperationOptions, namespaceName Namespace,
	queryDialect QueryDialect, queryExpression string,
) (*Operation, error) {
	if s == nil || s.ft == nil {
		return nil, ErrNotInitialized
	}

	queryExpressionUTF16, err := windows.UTF16PtrFromString(queryExpression)
	if err != nil {
		return nil, err
	}

	operation := &Operation{}

	if operationOptions == nil {
		operationOptions = s.defaultOperationOptions
	}

	r0, _, _ := syscall.SyscallN(
		s.ft.QueryInstances,
		uintptr(unsafe.Pointer(s)),
		uintptr(flags),
		uintptr(unsafe.Pointer(operationOptions)),
		uintptr(unsafe.Pointer(namespaceName)),
		uintptr(unsafe.Pointer(queryDialect)),
		uintptr(unsafe.Pointer(queryExpressionUTF16)),
		0,
		uintptr(unsafe.Pointer(operation)),
	)

	if result := ResultError(r0); !errors.Is(result, MI_RESULT_OK) {
		return nil, result
	}

	return operation, nil
}

// QueryUnmarshal queries for a set of instances based on a query expression.
//
// https://learn.microsoft.com/en-us/windows/win32/api/mi/nf-mi-mi_session_queryinstances
func (s *Session) QueryUnmarshal(dst any,
	flags OperationFlags, operationOptions *OperationOptions,
	namespaceName Namespace, queryDialect QueryDialect, queryExpression Query,
) error {
	if s == nil || s.ft == nil {
		return ErrNotInitialized
	}

	operation := &Operation{}

	if operationOptions == nil {
		operationOptions = s.defaultOperationOptions
	}

	dv := reflect.ValueOf(dst)
	if dv.Kind() != reflect.Ptr || dv.IsNil() {
		return ErrInvalidEntityType
	}

	dv = dv.Elem()

	elemType := dv.Type().Elem()
	elemValue := reflect.ValueOf(reflect.New(elemType).Interface()).Elem()

	if dv.Kind() != reflect.Slice || elemType.Kind() != reflect.Struct {
		return ErrInvalidEntityType
	}

	dv.Set(reflect.MakeSlice(dv.Type(), 0, 0))

	r0, _, _ := syscall.SyscallN(
		s.ft.QueryInstances,
		uintptr(unsafe.Pointer(s)),
		uintptr(flags),
		uintptr(unsafe.Pointer(operationOptions)),
		uintptr(unsafe.Pointer(namespaceName)),
		uintptr(unsafe.Pointer(queryDialect)),
		uintptr(unsafe.Pointer(queryExpression)),
		0,
		uintptr(unsafe.Pointer(operation)),
	)

	if result := ResultError(r0); !errors.Is(result, MI_RESULT_OK) {
		return result
	}

	defer func() {
		_ = operation.Close()
	}()

	for {
		instance, moreResults, err := operation.GetInstance()
		if err != nil {
			return fmt.Errorf("failed to get instance: %w", err)
		}

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

		for i := range elemType.NumField() {
			field := elemValue.Field(i)

			// Check if the field has an `mi` tag
			miTag := elemType.Field(i).Tag.Get("mi")
			if miTag == "" {
				continue
			}

			element, err := instance.GetElement(miTag)
			if err != nil {
				if errors.Is(err, MI_RESULT_NO_SUCH_PROPERTY) {
					continue
				}

				return fmt.Errorf("failed to get element %s: %w", miTag, err)
			}

			switch element.valueType {
			case ValueTypeBOOLEAN:
				field.SetBool(element.value == 1)
			case ValueTypeUINT8, ValueTypeUINT16, ValueTypeUINT32, ValueTypeUINT64:
				field.SetUint(uint64(element.value))
			case ValueTypeSINT8, ValueTypeSINT16, ValueTypeSINT32, ValueTypeSINT64:
				field.SetInt(int64(element.value))
			case ValueTypeSTRING:
				if element.value == 0 {
					// value is null
					continue
				}

				// Convert the UTF-16 string to a Go string
				stringValue := windows.UTF16PtrToString((*uint16)(unsafe.Pointer(element.value)))

				field.SetString(stringValue)
			case ValueTypeREAL32, ValueTypeREAL64:
				field.SetFloat(float64(element.value))
			default:
				return fmt.Errorf("unsupported value type: %d", element.valueType)
			}
		}

		dv.Set(reflect.Append(dv, elemValue))

		if !moreResults {
			break
		}
	}

	return nil
}

// Query queries for a set of instances based on a query expression.
func (s *Session) Query(dst any, namespaceName Namespace, queryExpression Query) error {
	err := s.QueryUnmarshal(dst, OperationFlagsStandardRTTI, nil, namespaceName, QueryDialectWQL, queryExpression)
	if err != nil {
		return fmt.Errorf("WMI query failed: %w", err)
	}

	return nil
}
