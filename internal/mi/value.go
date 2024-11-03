//go:build windows

package mi

import (
	"errors"
	"fmt"
	"unsafe"

	"golang.org/x/sys/windows"
)

type ValueType int

const (
	ValueTypeBOOLEAN ValueType = iota
	ValueTypeUINT8
	ValueTypeSINT8
	ValueTypeUINT16
	ValueTypeSINT16
	ValueTypeUINT32
	ValueTypeSINT32
	ValueTypeUINT64
	ValueTypeSINT64
	ValueTypeREAL32
	ValueTypeREAL64
	ValueTypeCHAR16
	ValueTypeDATETIME
	ValueTypeSTRING
	ValueTypeREFERENCE
	ValueTypeINSTANCE
	ValueTypeBOOLEANA
	ValueTypeUINT8A
	ValueTypeSINT8A
	ValueTypeUINT16A
	ValueTypeSINT16A
	ValueTypeUINT32A
	ValueTypeSINT32A
	ValueTypeUINT64A
	ValueTypeSINT64A
	ValueTypeREAL32A
	ValueTypeREAL64A
	ValueTypeCHAR16A
	ValueTypeDATETIMEA
	ValueTypeSTRINGA
	ValueTypeREFERENCEA
	ValueTypeINSTANCEA
	ValueTypeARRAY ValueType = 16
)

type Element struct {
	value     uintptr
	valueType ValueType
}

func (e *Element) GetValue() (any, error) {
	switch e.valueType {
	case ValueTypeBOOLEAN:
		return e.value == 1, nil
	case ValueTypeUINT8:
		return uint8(e.value), nil
	case ValueTypeSINT8:
		return int8(e.value), nil
	case ValueTypeUINT16:
		return uint16(e.value), nil
	case ValueTypeSINT16:
		return int16(e.value), nil
	case ValueTypeUINT32:
		return uint32(e.value), nil
	case ValueTypeSINT32:
		return int32(e.value), nil
	case ValueTypeUINT64:
		return uint64(e.value), nil
	case ValueTypeSINT64:
		return int64(e.value), nil
	case ValueTypeREAL32:
		return float32(e.value), nil
	case ValueTypeREAL64:
		return float64(e.value), nil
	case ValueTypeCHAR16:
		return uint16(e.value), nil
	case ValueTypeDATETIME:
		if e.value == 0 {
			return nil, errors.New("invalid pointer: value is nil")
		}

		return *(*Datetime)(unsafe.Pointer(e.value)), nil
	case ValueTypeSTRING:
		if e.value == 0 {
			return nil, errors.New("invalid pointer: value is nil")
		}

		// Convert the UTF-16 string to a Go string
		return windows.UTF16PtrToString((*uint16)(unsafe.Pointer(e.value))), nil
	case ValueTypeSTRINGA:
		if e.value == 0 {
			return nil, errors.New("invalid pointer: value is nil")
		}

		// Assuming array of pointers to UTF-16 strings
		ptrArray := *(*[]*uint16)(unsafe.Pointer(e.value))
		strArray := make([]string, len(ptrArray))

		for i, ptr := range ptrArray {
			strArray[i] = windows.UTF16PtrToString(ptr)
		}

		return strArray, nil
	default:
		return nil, fmt.Errorf("unsupported value type: %d", e.valueType)
	}
}
