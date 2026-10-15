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
	"context"
	"errors"
	"fmt"
	"math"
	"reflect"
	"sync"
	"syscall"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"
)

// BuildQueryTimeout bounds the probe queries collectors run in Build. Build
// has no scrape budget, and the session default, [DefaultQueryTimeout], is
// too short for slow providers. A timed-out query fails with
// MI_RESULT_INVALID_OPERATION_TIMEOUT.
const BuildQueryTimeout = 30 * time.Second

// DefaultQueryTimeout is the MI timeout of a query that is started with a
// non-positive timeout or with nil operation options. Every query started
// through [Session.Query] or [Session.QueryFunc] has an MI timeout: MI treats
// a zero timeout as unset, which would let a stalled provider block the query
// forever.
const DefaultQueryTimeout = 5 * time.Second

// minQueryTimeout is the shortest MI timeout. MI ignores timeouts below 1ms,
// so shorter positive timeouts are raised to it instead of disabling the
// timeout.
const minQueryTimeout = time.Millisecond

// maxCachedOperationOptions bounds the number of fixed query timeouts whose
// operation options a session keeps.
const maxCachedOperationOptions = 16

// Session represents a session.
//
// https://learn.microsoft.com/en-us/windows/win32/api/mi/ns-mi-mi_session
type Session struct {
	reserved1 uint64
	reserved2 uintptr
	ft        *SessionFT

	// Go-only state after the MI_Session fields, which MI never writes past.
	//
	// application created the session and outlives it, so per-query
	// operation options are created without MI_Session_GetApplication.
	application *Application
	// defaultOperationOptions carry DefaultQueryTimeout. They are used for
	// nil operation options and non-positive query timeouts.
	defaultOperationOptions *OperationOptions

	// optionsMu guards cachedOperationOptions and closed. The cached options
	// carry fixed timeouts, see operationOptions. Like the default options,
	// they are shared by concurrent operations and deleted by Close after
	// MI_Session_Close has closed every operation.
	optionsMu              sync.Mutex
	cachedOperationOptions map[time.Duration]*OperationOptions
	closed                 bool
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

	// Query and QueryFunc calls from now on fail instead of using options
	// that are about to be deleted.
	s.optionsMu.Lock()
	s.closed = true
	s.optionsMu.Unlock()

	// MI_Session_Close cancels running operations and blocks until their
	// handles are closed. The default and cached options are deleted
	// afterwards, so an operation that is still open never uses freed options.
	r0, _, _ := syscall.SyscallN(s.ft.Close,
		uintptr(unsafe.Pointer(s)),
		0,
		0,
	)

	s.optionsMu.Lock()

	for _, operationOptions := range s.cachedOperationOptions {
		_ = operationOptions.Delete()
	}

	s.cachedOperationOptions = nil

	if s.defaultOperationOptions != nil {
		_ = s.defaultOperationOptions.Delete()
		s.defaultOperationOptions = nil
	}

	s.optionsMu.Unlock()

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

	// MI_Session_TestConnection returns void. The outcome, e.g. a timeout, is
	// reported by GetInstance, and the operation must be closed either way.
	//
	// ref: https://github.com/KurtDeGreeff/omi/blob/9caa55032a1070a665e14fd282a091f6247d13c3/Unix/scriptext/py/PMI_Session.c#L92-L105
	_, _, _ = syscall.SyscallN(
		s.ft.TestConnection,
		uintptr(unsafe.Pointer(s)),
		0,
		0,
		uintptr(unsafe.Pointer(operation)),
	)

	_, _, err := operation.GetInstance()

	if closeErr := operation.Close(); err == nil && closeErr != nil {
		return fmt.Errorf("failed to close operation: %w", closeErr)
	}

	if err != nil {
		return fmt.Errorf("failed to get instance: %w", err)
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
// The results are read synchronously with [Operation.GetInstance]. Prefer
// [Session.Query] or [Session.QueryFunc], which receive them through callbacks;
// the synchronous path leaks WMI client handles under parallel load.
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

	return s.queryInstances(flags, operationOptions, namespaceName, queryDialect, queryExpressionUTF16), nil
}

// queryInstances starts an MI_Session_QueryInstances operation. Nil operationOptions
// fall back to the session defaults. The caller must close the returned operation.
//
// MI_Session_QueryInstances returns void: a failed query still yields an
// operation, and its status, e.g. a timeout or an invalid class, is reported
// by [Operation.GetInstance].
func (s *Session) queryInstances(flags OperationFlags, operationOptions *OperationOptions, namespaceName Namespace,
	queryDialect QueryDialect, queryExpression Query,
) *Operation {
	operation := &Operation{}

	if operationOptions == nil {
		operationOptions = s.defaultOperationOptions
	}

	_, _, _ = syscall.SyscallN(
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

	return operation
}

// unmarshalInstance populates structValue from instance using the `mi` struct tags.
//
// skipMissing controls what happens when the instance has no element for a tagged field.
// QueryUnmarshal passes true, tolerating classes whose MOF gained fields the query did not
// select. Operation.Unmarshal passes false, so a missing element is reported as an error
// instead of being silently left at its zero value.
func unmarshalInstance(instance *Instance, fields []miField, structValue reflect.Value, skipMissing bool) error {
	for _, f := range fields {
		element, err := instance.GetElementByName(f.name)
		if err != nil {
			if skipMissing && errors.Is(err, MI_RESULT_NO_SUCH_PROPERTY) {
				continue
			}

			return fmt.Errorf("failed to get element %s: %w", f.tag, err)
		}

		if err := setField(f.tag, structValue.Field(f.index), &element); err != nil {
			return err
		}
	}

	return nil
}

// setField assigns the value of element to field, rejecting Go field types that
// cannot hold the MI value instead of letting the reflect package panic.
func setField(miTag string, field reflect.Value, element *Element) error {
	if element.IsNull() {
		// MI makes no promise about the contents of a NULL element's union.
		// Check the Go type against a zeroed copy, so a mismatched field fails
		// whether or not the property is NULL, but leave the field untouched.
		return setField(miTag, reflect.New(field.Type()).Elem(), &Element{valueType: element.valueType})
	}

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
	case ValueTypeSTRINGA:
		if field.Type() != reflect.TypeFor[[]string]() {
			return fmt.Errorf("cannot unmarshal STRINGA into field of type %s, expected []string", field.Type())
		}

		field.Set(reflect.ValueOf(element.getStringArray()))
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
func (s *Session) QueryUnmarshal[T any](dst *[]T,
	flags OperationFlags, operationOptions *OperationOptions,
	namespaceName Namespace, queryDialect QueryDialect, queryExpression Query,
) error {
	if s == nil || s.ft == nil {
		return ErrNotInitialized
	}

	fields, err := prepareUnmarshal(dst)
	if err != nil {
		return err
	}

	return s.query(flags, operationOptions, namespaceName, queryDialect, queryExpression, func(instance *Instance) error {
		return appendInstance(dst, instance, fields, true)
	})
}

// Query queries for a set of instances based on a query expression and
// unmarshals them into dst. It is [Session.QueryContext] without
// cancellation.
func (s *Session) Query[T any](dst *[]T, namespaceName Namespace, queryExpression Query, queryTimeout time.Duration) error {
	return s.QueryContext(context.Background(), dst, namespaceName, queryExpression, queryTimeout)
}

// QueryContext queries for a set of instances based on a query expression and
// unmarshals them into dst.
//
// queryTimeout is the MI timeout of the query. A non-positive timeout selects
// [DefaultQueryTimeout], so MI never runs the query without a timeout. A
// positive timeout below 1ms is raised to 1ms, the shortest one MI honors.
//
// If ctx is done before the query starts, the query is not started. If ctx is
// done while the query runs, the operation is cancelled. Either way the
// context's cause is returned, unless the query failed before. QueryContext
// returns only after MI has delivered the final result of a started query.
func (s *Session) QueryContext[T any](ctx context.Context, dst *[]T, namespaceName Namespace, queryExpression Query,
	queryTimeout time.Duration,
) error {
	if s == nil || s.ft == nil {
		return ErrNotInitialized
	}

	fields, err := prepareUnmarshal(dst)
	if err != nil {
		return err
	}

	if ctx.Err() != nil {
		return context.Cause(ctx)
	}

	operationOptions, release, err := s.operationOptions(queryTimeout)
	if err != nil {
		return err
	}

	// Deferred first so the options outlive the operation that uses them.
	defer release()

	return s.queryContext(ctx, OperationFlagsStandardRTTI, operationOptions, namespaceName, QueryDialectWQL, queryExpression,
		func(instance *Instance) error {
			return appendInstance(dst, instance, fields, true)
		})
}

// QueryFunc runs a WQL query and calls fn for every returned instance. It is
// [Session.QueryFuncContext] without cancellation.
func (s *Session) QueryFunc(namespaceName Namespace, queryExpression Query, queryTimeout time.Duration, fn func(*Instance) error) error {
	return s.QueryFuncContext(context.Background(), namespaceName, queryExpression, queryTimeout, fn)
}

// QueryFuncContext runs a WQL query and calls fn for every returned instance,
// in order. It is meant for queries whose result shape is only known at
// runtime, where [Session.QueryContext] cannot unmarshal into a struct.
//
// queryTimeout and ctx behave as for [Session.QueryContext]. Once ctx is
// done, fn is not called any more.
//
// The instance passed to fn, and every element read from it, is only valid
// until fn returns. If fn returns an error, the query is cancelled and that
// error is returned.
//
// fn runs on the calling goroutine, for a copy of each instance: MI delivers
// the instances to a callback on one of its threads, which copies them and
// returns right away. A panic in fn cancels the query and propagates to the
// caller once MI has finished the operation.
//
// MI does not wait for fn. With a slow fn, copies of the remaining result set
// queue up in native memory, and the query timeout does not stop fn; check a
// deadline in fn where that matters.
func (s *Session) QueryFuncContext(ctx context.Context, namespaceName Namespace, queryExpression Query,
	queryTimeout time.Duration, fn func(*Instance) error,
) error {
	if s == nil || s.ft == nil {
		return ErrNotInitialized
	}

	if ctx.Err() != nil {
		return context.Cause(ctx)
	}

	operationOptions, release, err := s.operationOptions(queryTimeout)
	if err != nil {
		return err
	}

	// Deferred first so the options outlive the operation that uses them.
	defer release()

	query := newCopyQuery(fn)

	query.start(s, OperationFlagsStandardRTTI, operationOptions, namespaceName, QueryDialectWQL, queryExpression)

	// close cancels the operation if fn aborted it early or panicked.
	defer query.close()

	return query.wait(ctx)
}

// operationOptions returns the operation options for queryTimeout and a
// function that releases them once the operation that uses them is closed.
//
//   - A non-positive timeout, or DefaultQueryTimeout itself, returns the
//     session defaults.
//   - A timeout of whole seconds is taken as a fixed, configured timeout, e.g.
//     BuildQueryTimeout. Its options are created once and kept until Close,
//     for up to maxCachedOperationOptions distinct timeouts. Every query with
//     such a timeout saves three native calls.
//   - Any other timeout, typically the remaining scrape budget, gets options
//     of its own, which release deletes.
//
// Every path fails with ErrNotInitialized once Close has started.
func (s *Session) operationOptions(queryTimeout time.Duration) (*OperationOptions, func(), error) {
	s.optionsMu.Lock()
	defer s.optionsMu.Unlock()

	if s.closed || s.defaultOperationOptions == nil {
		return nil, nil, ErrNotInitialized
	}

	if queryTimeout <= 0 || queryTimeout == DefaultQueryTimeout {
		return s.defaultOperationOptions, func() {}, nil
	}

	queryTimeout = max(queryTimeout, minQueryTimeout)

	if queryTimeout%time.Second != 0 {
		return s.ownOperationOptions(queryTimeout)
	}

	if operationOptions, ok := s.cachedOperationOptions[queryTimeout]; ok {
		return operationOptions, func() {}, nil
	}

	if len(s.cachedOperationOptions) >= maxCachedOperationOptions {
		return s.ownOperationOptions(queryTimeout)
	}

	operationOptions, err := s.newOperationOptions(queryTimeout)
	if err != nil {
		return nil, nil, err
	}

	if s.cachedOperationOptions == nil {
		s.cachedOperationOptions = make(map[time.Duration]*OperationOptions, 1)
	}

	s.cachedOperationOptions[queryTimeout] = operationOptions

	return operationOptions, func() {}, nil
}

// ownOperationOptions returns new operation options for queryTimeout and a
// function that deletes them.
func (s *Session) ownOperationOptions(queryTimeout time.Duration) (*OperationOptions, func(), error) {
	operationOptions, err := s.newOperationOptions(queryTimeout)
	if err != nil {
		return nil, nil, err
	}

	return operationOptions, func() { _ = operationOptions.Delete() }, nil
}

// newOperationOptions creates operation options carrying queryTimeout, which
// must be at least minQueryTimeout. The caller must delete them once the
// operation that uses them is closed.
func (s *Session) newOperationOptions(queryTimeout time.Duration) (*OperationOptions, error) {
	operationOptions, err := s.application.NewOperationOptions()
	if err != nil {
		return nil, fmt.Errorf("failed to create operation options: %w", err)
	}

	if err = operationOptions.SetTimeout(queryTimeout); err != nil {
		_ = operationOptions.Delete()

		return nil, fmt.Errorf("failed to set timeout: %w", err)
	}

	return operationOptions, nil
}
