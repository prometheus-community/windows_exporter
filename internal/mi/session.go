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
	"math"
	"reflect"
	"syscall"
	"time"
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

// unmarshalInstance populates structValue from instance using the `mi` struct tags.
//
// skipMissing controls what happens when the instance has no element for a tagged field.
// QueryUnmarshal passes true, tolerating classes whose MOF gained fields the query did not
// select. Operation.Unmarshal passes false, so a missing element is reported as an error
// instead of being silently left at its zero value.
func unmarshalInstance(instance *Instance, structType reflect.Type, structValue reflect.Value, skipMissing bool) error {
	for i := range structType.NumField() {
		field := structValue.Field(i)

		miTag := structType.Field(i).Tag.Get("mi")
		if miTag == "" {
			continue
		}

		element, err := instance.GetElement(miTag)
		if err != nil {
			if skipMissing && errors.Is(err, MI_RESULT_NO_SUCH_PROPERTY) {
				continue
			}

			return fmt.Errorf("failed to get element %s: %w", miTag, err)
		}

		if err := setField(miTag, field, element); err != nil {
			return err
		}
	}

	return nil
}

// setField assigns the value of element to field, rejecting Go field types that
// cannot hold the MI value instead of letting the reflect package panic.
func setField(miTag string, field reflect.Value, element *Element) error {
	switch element.valueType {
	case ValueTypeBOOLEAN:
		if field.Kind() != reflect.Bool {
			return fieldTypeError(miTag, field, "boolean")
		}

		field.SetBool(element.value == 1)
	case ValueTypeUINT8, ValueTypeUINT16, ValueTypeUINT32, ValueTypeUINT64:
		return setUintField(miTag, field, uint64(element.value))
	case ValueTypeSINT8:
		return setIntField(miTag, field, int64(int8(element.value)))
	case ValueTypeSINT16:
		return setIntField(miTag, field, int64(int16(element.value)))
	case ValueTypeSINT32:
		return setIntField(miTag, field, int64(int32(element.value)))
	case ValueTypeSINT64:
		return setIntField(miTag, field, int64(element.value))
	case ValueTypeSTRING:
		if field.Kind() != reflect.String {
			return fieldTypeError(miTag, field, "string")
		}

		if element.value == 0 {
			return nil
		}

		field.SetString(windows.UTF16PtrToString((*uint16)(unsafe.Pointer(element.value))))
	case ValueTypeCHAR16:
		// A CHAR16 is a single UTF-16 code unit, not a pointer to a string.
		// Note: lone UTF-16 surrogates (0xD800–0xDFFF) are replaced with U+FFFD
		// by string(rune(...)). Use a uint16 Go field when the raw code unit is needed.
		if field.Kind() == reflect.String {
			field.SetString(string(rune(uint16(element.value))))

			return nil
		}

		return setUintField(miTag, field, uint64(uint16(element.value)))
	case ValueTypeREAL32:
		if field.Kind() != reflect.Float32 && field.Kind() != reflect.Float64 {
			return fieldTypeError(miTag, field, "float")
		}

		field.SetFloat(float64(math.Float32frombits(uint32(element.value))))
	case ValueTypeREAL64:
		if field.Kind() != reflect.Float64 {
			return fieldTypeError(miTag, field, "float")
		}

		field.SetFloat(math.Float64frombits(uint64(element.value)))
	case ValueTypeUINT16A:
		if field.Type() != reflect.TypeFor[[]uint16]() {
			return fmt.Errorf("cannot unmarshal UINT16A into field of type %s, expected []uint16", field.Type())
		}

		field.Set(reflect.ValueOf(element.getUint16Array()))
	default:
		return fmt.Errorf("unsupported value type: %d", element.valueType)
	}

	return nil
}

// setUintField assigns value to a numeric field, rejecting fields that cannot hold it
// instead of letting the reflect package panic or silently wrap.
func setUintField(miTag string, field reflect.Value, value uint64) error {
	if !isUnsignedKind(field.Kind()) {
		return fieldTypeError(miTag, field, "unsigned integer")
	}

	if field.OverflowUint(value) {
		return fmt.Errorf("field %s of Go type %s cannot hold the MI unsigned value %d", miTag, field.Type(), value)
	}

	field.SetUint(value)

	return nil
}

// setIntField assigns value to a numeric field, rejecting fields that cannot hold it.
// If the Go field is unsigned, non-negative values that fit are accepted; negative
// values are rejected with an error rather than silently wrapping.
func setIntField(miTag string, field reflect.Value, value int64) error {
	if !isNumericKind(field.Kind()) {
		return fieldTypeError(miTag, field, "signed integer")
	}

	if isUnsignedKind(field.Kind()) {
		if value < 0 || field.OverflowUint(uint64(value)) {
			return fmt.Errorf("field %s of Go type %s cannot hold the MI signed value %d", miTag, field.Type(), value)
		}

		field.SetUint(uint64(value))

		return nil
	}

	if field.OverflowInt(value) {
		return fmt.Errorf("field %s of Go type %s cannot hold the MI signed value %d", miTag, field.Type(), value)
	}

	field.SetInt(value)

	return nil
}

func isNumericKind(k reflect.Kind) bool {
	//nolint:exhaustive // We only care about fixed-width integer kinds.
	switch k {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return true
	default:
		return false
	}
}

func isUnsignedKind(k reflect.Kind) bool {
	//nolint:exhaustive // We only care about unsigned integer kinds.
	switch k {
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return true
	default:
		return false
	}
}

func fieldTypeError(miTag string, field reflect.Value, miType string) error {
	return fmt.Errorf("field %s is of Go type %s but the MI value is a %s", miTag, field.Type(), miType)
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

	dv, err := structSlice(dst)
	if err != nil {
		return err
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
		uintptr(unsafe.Pointer(queryExpression)),
		0,
		uintptr(unsafe.Pointer(operation)),
	)

	if result := ResultError(r0); !errors.Is(result, MI_RESULT_OK) {
		return fmt.Errorf("failed to query instances: %w", result)
	}

	defer func() {
		_ = operation.Close()
	}()

	return operation.unmarshal(dv, true)
}

// Query queries for a set of instances based on a query expression.
func (s *Session) Query(dst any, namespaceName Namespace, queryExpression Query, queryTimeout time.Duration) error {
	operationOptions, err := s.newOperationOptions(queryTimeout)
	if err != nil {
		return err
	}

	if operationOptions != nil {
		defer func() {
			_ = operationOptions.Delete()
		}()
	}

	return s.QueryUnmarshal(dst, OperationFlagsStandardRTTI, operationOptions, namespaceName, QueryDialectWQL, queryExpression)
}

// QueryFunc runs a WQL query and calls fn for every returned instance, in
// order. It is meant for queries whose result shape is only known at runtime,
// where [Session.Query] cannot unmarshal into a struct.
//
// The instance, and every element read from it, is owned by the operation and
// only valid until fn returns. If fn returns an error, the query is aborted
// and that error is returned.
func (s *Session) QueryFunc(namespaceName Namespace, queryExpression Query, queryTimeout time.Duration, fn func(*Instance) error) error {
	if s == nil || s.ft == nil {
		return ErrNotInitialized
	}

	operationOptions, err := s.newOperationOptions(queryTimeout)
	if err != nil {
		return err
	}

	if operationOptions != nil {
		// Deferred first so the options outlive the operation that uses them.
		defer func() {
			_ = operationOptions.Delete()
		}()
	} else {
		operationOptions = s.defaultOperationOptions
	}

	operation := &Operation{}

	r0, _, _ := syscall.SyscallN(
		s.ft.QueryInstances,
		uintptr(unsafe.Pointer(s)),
		uintptr(OperationFlagsStandardRTTI),
		uintptr(unsafe.Pointer(operationOptions)),
		uintptr(unsafe.Pointer(namespaceName)),
		uintptr(unsafe.Pointer(QueryDialectWQL)),
		uintptr(unsafe.Pointer(queryExpression)),
		0,
		uintptr(unsafe.Pointer(operation)),
	)

	if result := ResultError(r0); !errors.Is(result, MI_RESULT_OK) {
		return fmt.Errorf("failed to query instances: %w", result)
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
			return nil
		}

		if err := fn(instance); err != nil {
			return err
		}

		if !moreResults {
			return nil
		}
	}
}

// newOperationOptions creates operation options carrying queryTimeout. A zero
// timeout leaves the MI default in place; a negative timeout returns nil
// options, so the session defaults are used. The caller must delete non-nil
// options once the operation is closed.
func (s *Session) newOperationOptions(queryTimeout time.Duration) (*OperationOptions, error) {
	if queryTimeout < 0 {
		return nil, nil //nolint:nilnil
	}

	app, err := s.GetApplication()
	if err != nil {
		return nil, fmt.Errorf("failed to get application: %w", err)
	}

	operationOptions, err := app.NewOperationOptions()
	if err != nil {
		return nil, fmt.Errorf("failed to create operation options: %w", err)
	}

	if queryTimeout > 0 {
		if err = operationOptions.SetTimeout(queryTimeout); err != nil {
			_ = operationOptions.Delete()

			return nil, fmt.Errorf("failed to set timeout: %w", err)
		}
	}

	return operationOptions, nil
}
