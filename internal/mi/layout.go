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

import "unsafe"

// Compile-time checks of the Go mirrors of native MI structures against
// mi.h. MI reads and writes these structures by offset, so a missing or
// reordered field corrupts memory instead of failing.
//
// Each check indexes a one-element array with the difference of two
// constants: it compiles only when both are equal. A smaller left side
// overflows uintptr, a larger one is out of bounds.

const ptrSize = unsafe.Sizeof(uintptr(0))

// handleSize is the size of the MI handle structures MI_Application,
// MI_Session, MI_Operation, MI_OperationOptions and MI_DestinationOptions:
// MI_Uint64 reserved1, ptrdiff_t reserved2 and a function table pointer.
const handleSize = 8 + 2*ptrSize

// miValueSize is the size of the MI_Value union. Its largest member is
// MI_Datetime: an MI_Boolean, padded to 4 bytes, and a 32-byte MI_Timestamp or
// MI_Interval union, 36 bytes, rounded up to the 8-byte alignment of the
// 64-bit members. Instance.GetElementByName passes a buffer of this size.
const miValueSize = 40

var (
	// Handles without Go-only fields.
	_ = [1]struct{}{}[unsafe.Sizeof(Application{})-handleSize]
	_ = [1]struct{}{}[unsafe.Sizeof(OperationOptions{})-handleSize]
	_ = [1]struct{}{}[unsafe.Sizeof(DestinationOptions{})-handleSize]

	// Session and Operation append Go-only fields after the native layout,
	// which MI never writes past. The first Go-only field must start right
	// after the native fields, with the function table at the native offset.
	_ = [1]struct{}{}[unsafe.Offsetof(Session{}.ft)-(8+ptrSize)]
	_ = [1]struct{}{}[unsafe.Offsetof(Session{}.application)-handleSize]
	_ = [1]struct{}{}[unsafe.Offsetof(Operation{}.ft)-(8+ptrSize)]
	_ = [1]struct{}{}[unsafe.Offsetof(Operation{}.completed)-handleSize]

	// MI_Instance: four pointers and ptrdiff_t reserved[4].
	_ = [1]struct{}{}[unsafe.Sizeof(Instance{})-8*ptrSize]

	// MI_OperationCallbacks: the callback context and eight callbacks.
	_ = [1]struct{}{}[unsafe.Sizeof(operationCallbacks{})-9*ptrSize]
	_ = [1]struct{}{}[unsafe.Offsetof(operationCallbacks{}.instanceResult)-5*ptrSize]

	// MI_Timestamp and MI_Interval: eight 32-bit fields each.
	_ = [1]struct{}{}[unsafe.Sizeof(Timestamp{})-32]
	_ = [1]struct{}{}[unsafe.Sizeof(Interval{})-32]

	// The MI_Value buffer of Instance.GetElementByName, kept in Element.raw.
	_ = [1]struct{}{}[unsafe.Sizeof(Element{}.raw)-miValueSize]
	_ = [1]struct{}{}[unsafe.Sizeof(ValueType(0))-4]

	// Function tables, by the number of entries in mi.h. Calls use the slot
	// offsets, so a missing entry would call the wrong function.
	_ = [1]struct{}{}[unsafe.Sizeof(ApplicationFT{})-11*ptrSize]
	_ = [1]struct{}{}[unsafe.Sizeof(SessionFT{})-15*ptrSize]
	_ = [1]struct{}{}[unsafe.Sizeof(OperationFT{})-6*ptrSize]
	_ = [1]struct{}{}[unsafe.Sizeof(OperationOptionsFT{})-13*ptrSize]
	_ = [1]struct{}{}[unsafe.Sizeof(DestinationOptionsFT{})-15*ptrSize]
	_ = [1]struct{}{}[unsafe.Sizeof(InstanceFT{})-18*ptrSize]
)
