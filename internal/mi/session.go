//go:build windows

package mi

import (
	"errors"
	"fmt"
	"runtime"
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

	var err error

	if _, _, err = operation.GetInstance(); err != nil {
		return fmt.Errorf("failed to get instance: %w", err)
	}

	if err = operation.Close(); err != nil {
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

	errCh := make(chan error, 1)

	operationCallbacks, err := NewUnmarshalOperationsCallbacks(dst, errCh)
	if err != nil {
		return err
	}

	r0, _, _ := syscall.SyscallN(
		s.ft.QueryInstances,
		uintptr(unsafe.Pointer(s)),
		uintptr(flags),
		uintptr(unsafe.Pointer(operationOptions)),
		uintptr(unsafe.Pointer(namespaceName)),
		uintptr(unsafe.Pointer(queryDialect)),
		uintptr(unsafe.Pointer(queryExpression)),
		uintptr(unsafe.Pointer(operationCallbacks)),
		uintptr(unsafe.Pointer(operation)),
	)

	if result := ResultError(r0); !errors.Is(result, MI_RESULT_OK) {
		return result
	}

	errs := make([]error, 0)

	// We need an active go routine to prevent a
	// fatal error: all goroutines are asleep - deadlock!
	// ref: https://github.com/golang/go/issues/55015
	// go time.Sleep(5 * time.Second)

	for {
		if err, ok := <-errCh; err != nil {
			errs = append(errs, err)
		} else if !ok {
			break
		}
	}

	// KeepAlive is used to ensure that the callbacks are not garbage collected before the operation is closed.
	runtime.KeepAlive(operationCallbacks.CallbackContext)

	return errors.Join(errs...)
}

// Query queries for a set of instances based on a query expression.
func (s *Session) Query(dst any, namespaceName Namespace, queryExpression Query) error {
	err := s.QueryUnmarshal(dst, OperationFlagsStandardRTTI, nil, namespaceName, QueryDialectWQL, queryExpression)
	if err != nil {
		return fmt.Errorf("WMI query failed: %w", err)
	}

	return nil
}
