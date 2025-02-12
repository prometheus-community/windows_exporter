// Copyright 2024 The Prometheus Authors
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
	"sync"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"
)

// operationUnmarshalCallbacksInstanceResult registers a global callback function.
// The amount of system callbacks is limited to 2000.
//
//nolint:gochecknoglobals
var operationUnmarshalCallbacksInstanceResult = sync.OnceValue[uintptr](func() uintptr {
	// Workaround for a deadlock issue in go.
	// Ref: https://github.com/golang/go/issues/55015
	go time.Sleep(time.Duration(math.MaxInt64))

	return windows.NewCallback(func(
		operation *Operation,
		callbacks *OperationUnmarshalCallbacks,
		instance *Instance,
		moreResults Boolean,
		instanceResult ResultError,
		errorMessageUTF16 *uint16,
		errorDetails *Instance,
		_ uintptr,
	) uintptr {
		if moreResults == False {
			defer operation.Close()
		}

		return callbacks.InstanceResult(operation, instance, moreResults, instanceResult, errorMessageUTF16, errorDetails)
	})
})

type OperationUnmarshalCallbacks struct {
	dst   any
	dv    reflect.Value
	errCh chan<- error

	elemType  reflect.Type
	elemValue reflect.Value
}

func NewUnmarshalOperationsCallbacks(dst any, errCh chan<- error) (*OperationCallbacks[OperationUnmarshalCallbacks], error) {
	dv := reflect.ValueOf(dst)
	if dv.Kind() != reflect.Ptr || dv.IsNil() {
		return nil, ErrInvalidEntityType
	}

	dv = dv.Elem()

	elemType := dv.Type().Elem()
	elemValue := reflect.ValueOf(reflect.New(elemType).Interface()).Elem()

	if dv.Kind() != reflect.Slice || elemType.Kind() != reflect.Struct {
		return nil, ErrInvalidEntityType
	}

	dv.Set(reflect.MakeSlice(dv.Type(), 0, 0))

	return &OperationCallbacks[OperationUnmarshalCallbacks]{
		CallbackContext: &OperationUnmarshalCallbacks{
			errCh:     errCh,
			dst:       dst,
			dv:        dv,
			elemType:  elemType,
			elemValue: elemValue,
		},
		InstanceResult: operationUnmarshalCallbacksInstanceResult(),
	}, nil
}

func (o *OperationUnmarshalCallbacks) InstanceResult(
	_ *Operation,
	instance *Instance,
	moreResults Boolean,
	instanceResult ResultError,
	errorMessageUTF16 *uint16,
	_ *Instance,
) uintptr {
	defer func() {
		if moreResults == False {
			close(o.errCh)
		}
	}()

	if !errors.Is(instanceResult, MI_RESULT_OK) {
		o.errCh <- fmt.Errorf("%w: %s", instanceResult, windows.UTF16PtrToString(errorMessageUTF16))

		return 0
	}

	if instance == nil {
		return 0
	}

	counter, err := instance.GetElementCount()
	if err != nil {
		o.errCh <- fmt.Errorf("failed to get element count: %w", err)

		return 0
	}

	if counter == 0 {
		return 0
	}

	for i := range o.elemType.NumField() {
		field := o.elemValue.Field(i)

		// Check if the field has an `mi` tag
		miTag := o.elemType.Field(i).Tag.Get("mi")
		if miTag == "" {
			continue
		}

		element, err := instance.GetElement(miTag)
		if err != nil {
			if errors.Is(err, MI_RESULT_NO_SUCH_PROPERTY) {
				continue
			}

			o.errCh <- fmt.Errorf("failed to get element %s: %w", miTag, err)

			return 0
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
			o.errCh <- fmt.Errorf("unsupported value type: %d", element.valueType)

			return 0
		}
	}

	o.dv.Set(reflect.Append(o.dv, o.elemValue))

	return 0
}
