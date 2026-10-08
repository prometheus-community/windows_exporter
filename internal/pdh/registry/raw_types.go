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

package registry

import (
	"io"

	"golang.org/x/sys/windows"
)

/*
perfDataBlock
See: https://msdn.microsoft.com/de-de/library/windows/desktop/aa373157(v=vs.85).aspx

	typedef struct _PERF_DATA_BLOCK {
	  WCHAR         Signature[4];
	  DWORD         LittleEndian;
	  DWORD         Version;
	  DWORD         Revision;
	  DWORD         TotalByteLength;
	  DWORD         HeaderLength;
	  DWORD         NumObjectTypes;
	  DWORD         DefaultObject;
	  SYSTEMTIME    SystemTime;
	  LARGE_INTEGER PerfTime;
	  LARGE_INTEGER PerfFreq;
	  LARGE_INTEGER PerfTime100nSec;
	  DWORD         SystemNameLength;
	  DWORD         SystemNameOffset;
	} PERF_DATA_BLOCK;
*/
type perfDataBlock struct {
	Signature        [4]uint16
	LittleEndian     uint32
	Version          uint32
	Revision         uint32
	TotalByteLength  uint32
	HeaderLength     uint32
	NumObjectTypes   uint32
	DefaultObject    int32
	SystemTime       windows.Systemtime
	_                uint32 // unknown field
	PerfTime         int64
	PerfFreq         int64
	PerfTime100nSec  int64
	SystemNameLength uint32
	SystemNameOffset uint32
}

// decode reads p from the start of b. The fields are decoded directly, since
// binary.Read allocates and walks the struct by reflection on every call.
func (p *perfDataBlock) decode(b []byte) error {
	if int64(len(b)) < perfDataBlockSize {
		return io.ErrUnexpectedEOF
	}

	*p = perfDataBlock{
		Signature:       [4]uint16{bo.Uint16(b[0:]), bo.Uint16(b[2:]), bo.Uint16(b[4:]), bo.Uint16(b[6:])},
		LittleEndian:    bo.Uint32(b[8:]),
		Version:         bo.Uint32(b[12:]),
		Revision:        bo.Uint32(b[16:]),
		TotalByteLength: bo.Uint32(b[20:]),
		HeaderLength:    bo.Uint32(b[24:]),
		NumObjectTypes:  bo.Uint32(b[28:]),
		DefaultObject:   int32(bo.Uint32(b[32:])),
		SystemTime: windows.Systemtime{
			Year:         bo.Uint16(b[36:]),
			Month:        bo.Uint16(b[38:]),
			DayOfWeek:    bo.Uint16(b[40:]),
			Day:          bo.Uint16(b[42:]),
			Hour:         bo.Uint16(b[44:]),
			Minute:       bo.Uint16(b[46:]),
			Second:       bo.Uint16(b[48:]),
			Milliseconds: bo.Uint16(b[50:]),
		},
		PerfTime:         int64(bo.Uint64(b[56:])),
		PerfFreq:         int64(bo.Uint64(b[64:])),
		PerfTime100nSec:  int64(bo.Uint64(b[72:])),
		SystemNameLength: bo.Uint32(b[80:]),
		SystemNameOffset: bo.Uint32(b[84:]),
	}

	return nil
}

/*
perfObjectType
See: https://msdn.microsoft.com/en-us/library/windows/desktop/aa373160(v=vs.85).aspx

	typedef struct _PERF_OBJECT_TYPE {
	  DWORD         TotalByteLength;
	  DWORD         DefinitionLength;
	  DWORD         HeaderLength;
	  DWORD         ObjectNameTitleIndex;
	  LPWSTR        ObjectNameTitle;
	  DWORD         ObjectHelpTitleIndex;
	  LPWSTR        ObjectHelpTitle;
	  DWORD         DetailLevel;
	  DWORD         NumCounters;
	  DWORD         DefaultCounter;
	  DWORD         NumInstances;
	  DWORD         CodePage;
	  LARGE_INTEGER PerfTime;
	  LARGE_INTEGER PerfFreq;
	} PERF_OBJECT_TYPE;
*/
type perfObjectType struct {
	TotalByteLength      uint32
	DefinitionLength     uint32
	HeaderLength         uint32
	ObjectNameTitleIndex uint32
	ObjectNameTitle      uint32
	ObjectHelpTitleIndex uint32
	ObjectHelpTitle      uint32
	DetailLevel          uint32
	NumCounters          uint32
	DefaultCounter       int32
	NumInstances         int32
	CodePage             uint32
	PerfTime             int64
	PerfFreq             int64
}

// decode reads p from the start of b.
func (p *perfObjectType) decode(b []byte) error {
	if int64(len(b)) < perfObjectTypeSize {
		return io.ErrUnexpectedEOF
	}

	*p = perfObjectType{
		TotalByteLength:      bo.Uint32(b[0:]),
		DefinitionLength:     bo.Uint32(b[4:]),
		HeaderLength:         bo.Uint32(b[8:]),
		ObjectNameTitleIndex: bo.Uint32(b[12:]),
		ObjectNameTitle:      bo.Uint32(b[16:]),
		ObjectHelpTitleIndex: bo.Uint32(b[20:]),
		ObjectHelpTitle:      bo.Uint32(b[24:]),
		DetailLevel:          bo.Uint32(b[28:]),
		NumCounters:          bo.Uint32(b[32:]),
		DefaultCounter:       int32(bo.Uint32(b[36:])),
		NumInstances:         int32(bo.Uint32(b[40:])),
		CodePage:             bo.Uint32(b[44:]),
		PerfTime:             int64(bo.Uint64(b[48:])),
		PerfFreq:             int64(bo.Uint64(b[56:])),
	}

	return nil
}

/*
perfCounterDefinition
See: https://msdn.microsoft.com/en-us/library/windows/desktop/aa373150(v=vs.85).aspx

	typedef struct _PERF_COUNTER_DEFINITION {
	  DWORD  ByteLength;
	  DWORD  CounterNameTitleIndex;
	  LPWSTR CounterNameTitle;
	  DWORD  CounterHelpTitleIndex;
	  LPWSTR CounterHelpTitle;
	  LONG   DefaultScale;
	  DWORD  DetailLevel;
	  DWORD  CounterType;
	  DWORD  CounterSize;
	  DWORD  CounterOffset;
	} PERF_COUNTER_DEFINITION;
*/
type perfCounterDefinition struct {
	ByteLength            uint32
	CounterNameTitleIndex uint32
	CounterNameTitle      uint32
	CounterHelpTitleIndex uint32
	CounterHelpTitle      uint32
	DefaultScale          int32
	DetailLevel           uint32
	CounterType           uint32
	CounterSize           uint32
	CounterOffset         uint32
}

// decode reads p from the start of b.
func (p *perfCounterDefinition) decode(b []byte) error {
	if int64(len(b)) < perfCounterDefinitionSize {
		return io.ErrUnexpectedEOF
	}

	*p = perfCounterDefinition{
		ByteLength:            bo.Uint32(b[0:]),
		CounterNameTitleIndex: bo.Uint32(b[4:]),
		CounterNameTitle:      bo.Uint32(b[8:]),
		CounterHelpTitleIndex: bo.Uint32(b[12:]),
		CounterHelpTitle:      bo.Uint32(b[16:]),
		DefaultScale:          int32(bo.Uint32(b[20:])),
		DetailLevel:           bo.Uint32(b[24:]),
		CounterType:           bo.Uint32(b[28:]),
		CounterSize:           bo.Uint32(b[32:]),
		CounterOffset:         bo.Uint32(b[36:]),
	}

	return nil
}

func (p *perfCounterDefinition) LookupName() string {
	return CounterNameTable.LookupString(p.CounterNameTitleIndex)
}

/*
perfCounterBlock
See: https://msdn.microsoft.com/en-us/library/windows/desktop/aa373147(v=vs.85).aspx

	typedef struct _PERF_COUNTER_BLOCK {
	  DWORD ByteLength;
	} PERF_COUNTER_BLOCK;
*/
type perfCounterBlock struct {
	ByteLength uint32
}

// decode reads p from the start of b.
func (p *perfCounterBlock) decode(b []byte) error {
	if int64(len(b)) < perfCounterBlockSize {
		return io.ErrUnexpectedEOF
	}

	p.ByteLength = bo.Uint32(b)

	return nil
}

/*
perfInstanceDefinition
See: https://msdn.microsoft.com/en-us/library/windows/desktop/aa373159(v=vs.85).aspx

	typedef struct _PERF_INSTANCE_DEFINITION {
	  DWORD ByteLength;
	  DWORD ParentObjectTitleIndex;
	  DWORD ParentObjectInstance;
	  DWORD UniqueID;
	  DWORD NameOffset;
	  DWORD NameLength;
	} PERF_INSTANCE_DEFINITION;
*/
type perfInstanceDefinition struct {
	ByteLength             uint32
	ParentObjectTitleIndex uint32
	ParentObjectInstance   uint32
	UniqueID               uint32
	NameOffset             uint32
	NameLength             uint32
}

// decode reads p from the start of b.
func (p *perfInstanceDefinition) decode(b []byte) error {
	if int64(len(b)) < perfInstanceDefinitionSize {
		return io.ErrUnexpectedEOF
	}

	*p = perfInstanceDefinition{
		ByteLength:             bo.Uint32(b[0:]),
		ParentObjectTitleIndex: bo.Uint32(b[4:]),
		ParentObjectInstance:   bo.Uint32(b[8:]),
		UniqueID:               bo.Uint32(b[12:]),
		NameOffset:             bo.Uint32(b[16:]),
		NameLength:             bo.Uint32(b[20:]),
	}

	return nil
}
