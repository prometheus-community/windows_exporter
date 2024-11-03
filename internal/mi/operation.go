//go:build windows

package mi

import (
	"errors"
	"fmt"
	"reflect"
	"syscall"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"
)

// OperationOptionsTimeout is the key for the timeout option.
//
// https://github.com/microsoft/win32metadata/blob/527806d20d83d3abd43d16cd3fa8795d8deba343/generation/WinSDK/RecompiledIdlHeaders/um/mi.h#L9240
var OperationOptionsTimeout = UTF16PtrFromString[*uint16]("__MI_OPERATIONOPTIONS_TIMEOUT")

// OperationFlags represents the flags for an operation.
//
// https://learn.microsoft.com/en-us/previous-versions/windows/desktop/wmi_v2/mi-flags
type OperationFlags uint32

const (
	OperationFlagsDefaultRTTI  OperationFlags = 0x0000
	OperationFlagsBasicRTTI    OperationFlags = 0x0002
	OperationFlagsNoRTTI       OperationFlags = 0x0400
	OperationFlagsStandardRTTI OperationFlags = 0x0800
	OperationFlagsFullRTTI     OperationFlags = 0x0004
)

// Operation represents an operation.
// https://learn.microsoft.com/en-us/windows/win32/api/mi/ns-mi-mi_operation
type Operation struct {
	reserved1 uint64
	reserved2 uintptr
	ft        *OperationFT
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

	r0, _, _ := syscall.SyscallN(o.ft.Close, uintptr(unsafe.Pointer(o)), 0)

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

	if !errors.Is(instanceResult, MI_RESULT_OK) {
		return nil, false, fmt.Errorf("instance result: %w (%s)", instanceResult, windows.UTF16PtrToString(errorMessageUTF16))
	}

	if result := ResultError(r0); !errors.Is(result, MI_RESULT_OK) {
		return nil, false, result
	}

	return instance, moreResults == True, nil
}

func (o *Operation) Unmarshal(dst any) error {
	if o == nil || o.ft == nil {
		return ErrNotInitialized
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

		for i := range elemType.NumField() {
			field := elemValue.Field(i)

			// Check if the field has an `mi` tag
			miTag := elemType.Field(i).Tag.Get("mi")
			if miTag == "" {
				continue
			}

			element, err := instance.GetElement(miTag)
			if err != nil {
				return fmt.Errorf("failed to get element: %w", err)
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
					return fmt.Errorf("%s: invalid pointer: value is nil", miTag)
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

func (o *OperationOptions) Delete() error {
	if o == nil || o.ft == nil {
		return ErrNotInitialized
	}

	r0, _, _ := syscall.SyscallN(o.ft.Delete, uintptr(unsafe.Pointer(o)))

	if result := ResultError(r0); !errors.Is(result, MI_RESULT_OK) {
		return result
	}

	return nil
}
