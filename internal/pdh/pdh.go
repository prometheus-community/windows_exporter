// Copyright (c) 2010-2024 The win Authors. All rights reserved.
// Copyright (c) 2024 The prometheus-community Authors. All rights reserved.
//
// Redistribution and use in source and binary forms, with or without
// modification, are permitted provided that the following conditions
// are met:
// 1. Redistributions of source code must retain the above copyright
//    notice, this list of conditions and the following disclaimer.
// 2. Redistributions in binary form must reproduce the above copyright
//    notice, this list of conditions and the following disclaimer in the
//    documentation and/or other materials provided with the distribution.
// 3. The names of the authors may not be used to endorse or promote products
//    derived from this software without specific prior written permission.
//
// THIS SOFTWARE IS PROVIDED BY THE AUTHORS ``AS IS'' AND ANY EXPRESS OR
// IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE IMPLIED WARRANTIES
// OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE DISCLAIMED.
// IN NO EVENT SHALL THE AUTHORS BE LIABLE FOR ANY DIRECT, INDIRECT,
// INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT
// NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE,
// DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY
// THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
// (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF
// THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
//
// This is the official list of 'win' authors for copyright purposes.
//
// Alexander Neumann <an2048@googlemail.com>
// Joseph Watson <jtwatson@linux-consulting.us>
// Kevin Pors <krpors@gmail.com>

//go:build windows

package pdh

import (
	"fmt"
	"time"
	"unsafe"

	"github.com/prometheus-community/windows_exporter/internal/headers/kernel32"
	"golang.org/x/sys/windows"
)

// Error codes.
const (
	ErrorSuccess         = 0
	ErrorFailure         = 1
	ErrorInvalidFunction = 1
)

type (
	HANDLE uintptr
)

// PDH error codes, which can be returned by all Pdh* functions. Taken from mingw-w64 pdhmsg.h

const (
	CstatusValidData                   uint32 = 0x00000000 // The returned data is valid.
	CstatusNewData                     uint32 = 0x00000001 // The return data value is valid and different from the last sample.
	CstatusNoMachine                   uint32 = 0x800007D0 // Unable to connect to the specified computer, or the computer is offline.
	CstatusNoInstance                  uint32 = 0x800007D1
	MoreData                           uint32 = 0x800007D2 // The PdhGetFormattedCounterArray* function can return this if there's 'more data to be displayed'.
	CstatusItemNotValidated            uint32 = 0x800007D3
	Retry                              uint32 = 0x800007D4
	NoData                             uint32 = 0x800007D5 // The query does not currently contain any counters (for example, limited access)
	CalcNegativeDenominator            uint32 = 0x800007D6
	CalcNegativeTimebase               uint32 = 0x800007D7
	CalcNegativeValue                  uint32 = 0x800007D8
	DialogCancelled                    uint32 = 0x800007D9
	EndOfLogFile                       uint32 = 0x800007DA
	AsyncQueryTimeout                  uint32 = 0x800007DB
	CannotSetDefaultRealtimeDatasource uint32 = 0x800007DC
	CstatusNoObject                    uint32 = 0xC0000BB8
	CstatusNoCounter                   uint32 = 0xC0000BB9 // The specified counter could not be found.
	CstatusInvalidData                 uint32 = 0xC0000BBA // The counter was successfully found, but the data returned is not valid.
	MemoryAllocationFailure            uint32 = 0xC0000BBB
	InvalidHandle                      uint32 = 0xC0000BBC
	InvalidArgument                    uint32 = 0xC0000BBD // Required argument is missing or incorrect.
	FunctionNotFound                   uint32 = 0xC0000BBE
	CstatusNoCountername               uint32 = 0xC0000BBF
	CstatusBadCountername              uint32 = 0xC0000BC0 // Unable to parse the counter path. Check the format and syntax of the specified path.
	InvalidBuffer                      uint32 = 0xC0000BC1
	InsufficientBuffer                 uint32 = 0xC0000BC2
	CannotConnectMachine               uint32 = 0xC0000BC3
	InvalidPath                        uint32 = 0xC0000BC4
	InvalidInstance                    uint32 = 0xC0000BC5
	InvalidData                        uint32 = 0xC0000BC6 // specified counter does not contain valid data or a successful status code.
	NoDialogData                       uint32 = 0xC0000BC7
	CannotReadNameStrings              uint32 = 0xC0000BC8
	LogFileCreateError                 uint32 = 0xC0000BC9
	LogFileOpenError                   uint32 = 0xC0000BCA
	LogTypeNotFound                    uint32 = 0xC0000BCB
	NoMoreData                         uint32 = 0xC0000BCC
	EntryNotInLogFile                  uint32 = 0xC0000BCD
	DataSourceIsLogFile                uint32 = 0xC0000BCE
	DataSourceIsRealTime               uint32 = 0xC0000BCF
	UnableReadLogHeader                uint32 = 0xC0000BD0
	FileNotFound                       uint32 = 0xC0000BD1
	FileAlreadyExists                  uint32 = 0xC0000BD2
	NotImplemented                     uint32 = 0xC0000BD3
	StringNotFound                     uint32 = 0xC0000BD4
	UnableMapNameFiles                 uint32 = 0x80000BD5
	UnknownLogFormat                   uint32 = 0xC0000BD6
	UnknownLogsvcCommand               uint32 = 0xC0000BD7
	LogsvcQueryNotFound                uint32 = 0xC0000BD8
	LogsvcNotOpened                    uint32 = 0xC0000BD9
	WbemError                          uint32 = 0xC0000BDA
	AccessDenied                       uint32 = 0xC0000BDB
	LogFileTooSmall                    uint32 = 0xC0000BDC
	InvalidDatasource                  uint32 = 0xC0000BDD
	InvalidSqldb                       uint32 = 0xC0000BDE
	NoCounters                         uint32 = 0xC0000BDF
	SQLAllocFailed                     uint32 = 0xC0000BE0
	SQLAllocconFailed                  uint32 = 0xC0000BE1
	SQLExecDirectFailed                uint32 = 0xC0000BE2
	SQLFetchFailed                     uint32 = 0xC0000BE3
	SQLRowcountFailed                  uint32 = 0xC0000BE4
	SQLMoreResultsFailed               uint32 = 0xC0000BE5
	SQLConnectFailed                   uint32 = 0xC0000BE6
	SQLBindFailed                      uint32 = 0xC0000BE7
	CannotConnectWmiServer             uint32 = 0xC0000BE8
	PlaCollectionAlreadyRunning        uint32 = 0xC0000BE9
	PlaErrorScheduleOverlap            uint32 = 0xC0000BEA
	PlaCollectionNotFound              uint32 = 0xC0000BEB
	PlaErrorScheduleElapsed            uint32 = 0xC0000BEC
	PlaErrorNostart                    uint32 = 0xC0000BED
	PlaErrorAlreadyExists              uint32 = 0xC0000BEE
	PlaErrorTypeMismatch               uint32 = 0xC0000BEF
	PlaErrorFilepath                   uint32 = 0xC0000BF0
	PlaServiceError                    uint32 = 0xC0000BF1
	PlaValidationError                 uint32 = 0xC0000BF2
	PlaValidationWarning               uint32 = 0x80000BF3
	PlaErrorNameTooLong                uint32 = 0xC0000BF4
	InvalidSQLLogFormat                uint32 = 0xC0000BF5
	CounterAlreadyInQuery              uint32 = 0xC0000BF6
	BinaryLogCorrupt                   uint32 = 0xC0000BF7
	LogSampleTooSmall                  uint32 = 0xC0000BF8
	OsLaterVersion                     uint32 = 0xC0000BF9
	OsEarlierVersion                   uint32 = 0xC0000BFA
	IncorrectAppendTime                uint32 = 0xC0000BFB
	UnmatchedAppendCounter             uint32 = 0xC0000BFC
	SQLAlterDetailFailed               uint32 = 0xC0000BFD
	QueryPerfDataTimeout               uint32 = 0xC0000BFE
)

//nolint:gochecknoglobals
var Errors = map[uint32]string{
	CstatusValidData:                   "PDH_CSTATUS_VALID_DATA",
	CstatusNewData:                     "PDH_CSTATUS_NEW_DATA",
	CstatusNoMachine:                   "PDH_CSTATUS_NO_MACHINE",
	CstatusNoInstance:                  "PDH_CSTATUS_NO_INSTANCE",
	MoreData:                           "PDH_MORE_DATA",
	CstatusItemNotValidated:            "PDH_CSTATUS_ITEM_NOT_VALIDATED",
	Retry:                              "PDH_RETRY",
	NoData:                             "PDH_NO_DATA",
	CalcNegativeDenominator:            "PDH_CALC_NEGATIVE_DENOMINATOR",
	CalcNegativeTimebase:               "PDH_CALC_NEGATIVE_TIMEBASE",
	CalcNegativeValue:                  "PDH_CALC_NEGATIVE_VALUE",
	DialogCancelled:                    "PDH_DIALOG_CANCELLED",
	EndOfLogFile:                       "PDH_END_OF_LOG_FILE",
	AsyncQueryTimeout:                  "PDH_ASYNC_QUERY_TIMEOUT",
	CannotSetDefaultRealtimeDatasource: "PDH_CANNOT_SET_DEFAULT_REALTIME_DATASOURCE",
	CstatusNoObject:                    "PDH_CSTATUS_NO_OBJECT",
	CstatusNoCounter:                   "PDH_CSTATUS_NO_COUNTER",
	CstatusInvalidData:                 "PDH_CSTATUS_INVALID_DATA",
	MemoryAllocationFailure:            "PDH_MEMORY_ALLOCATION_FAILURE",
	InvalidHandle:                      "PDH_INVALID_HANDLE",
	InvalidArgument:                    "PDH_INVALID_ARGUMENT",
	FunctionNotFound:                   "PDH_FUNCTION_NOT_FOUND",
	CstatusNoCountername:               "PDH_CSTATUS_NO_COUNTERNAME",
	CstatusBadCountername:              "PDH_CSTATUS_BAD_COUNTERNAME",
	InvalidBuffer:                      "PDH_INVALID_BUFFER",
	InsufficientBuffer:                 "PDH_INSUFFICIENT_BUFFER",
	CannotConnectMachine:               "PDH_CANNOT_CONNECT_MACHINE",
	InvalidPath:                        "PDH_INVALID_PATH",
	InvalidInstance:                    "PDH_INVALID_INSTANCE",
	InvalidData:                        "PDH_INVALID_DATA",
	NoDialogData:                       "PDH_NO_DIALOG_DATA",
	CannotReadNameStrings:              "PDH_CANNOT_READ_NAME_STRINGS",
	LogFileCreateError:                 "PDH_LOG_FILE_CREATE_ERROR",
	LogFileOpenError:                   "PDH_LOG_FILE_OPEN_ERROR",
	LogTypeNotFound:                    "PDH_LOG_TYPE_NOT_FOUND",
	NoMoreData:                         "PDH_NO_MORE_DATA",
	EntryNotInLogFile:                  "PDH_ENTRY_NOT_IN_LOG_FILE",
	DataSourceIsLogFile:                "PDH_DATA_SOURCE_IS_LOG_FILE",
	DataSourceIsRealTime:               "PDH_DATA_SOURCE_IS_REAL_TIME",
	UnableReadLogHeader:                "PDH_UNABLE_READ_LOG_HEADER",
	FileNotFound:                       "PDH_FILE_NOT_FOUND",
	FileAlreadyExists:                  "PDH_FILE_ALREADY_EXISTS",
	NotImplemented:                     "PDH_NOT_IMPLEMENTED",
	StringNotFound:                     "PDH_STRING_NOT_FOUND",
	UnableMapNameFiles:                 "PDH_UNABLE_MAP_NAME_FILES",
	UnknownLogFormat:                   "PDH_UNKNOWN_LOG_FORMAT",
	UnknownLogsvcCommand:               "PDH_UNKNOWN_LOGSVC_COMMAND",
	LogsvcQueryNotFound:                "PDH_LOGSVC_QUERY_NOT_FOUND",
	LogsvcNotOpened:                    "PDH_LOGSVC_NOT_OPENED",
	WbemError:                          "PDH_WBEM_ERROR",
	AccessDenied:                       "PDH_ACCESS_DENIED",
	LogFileTooSmall:                    "PDH_LOG_FILE_TOO_SMALL",
	InvalidDatasource:                  "PDH_INVALID_DATASOURCE",
	InvalidSqldb:                       "PDH_INVALID_SQLDB",
	NoCounters:                         "PDH_NO_COUNTERS",
	SQLAllocFailed:                     "PDH_SQL_ALLOC_FAILED",
	SQLAllocconFailed:                  "PDH_SQL_ALLOCCON_FAILED",
	SQLExecDirectFailed:                "PDH_SQL_EXEC_DIRECT_FAILED",
	SQLFetchFailed:                     "PDH_SQL_FETCH_FAILED",
	SQLRowcountFailed:                  "PDH_SQL_ROWCOUNT_FAILED",
	SQLMoreResultsFailed:               "PDH_SQL_MORE_RESULTS_FAILED",
	SQLConnectFailed:                   "PDH_SQL_CONNECT_FAILED",
	SQLBindFailed:                      "PDH_SQL_BIND_FAILED",
	CannotConnectWmiServer:             "PDH_CANNOT_CONNECT_WMI_SERVER",
	PlaCollectionAlreadyRunning:        "PDH_PLA_COLLECTION_ALREADY_RUNNING",
	PlaErrorScheduleOverlap:            "PDH_PLA_ERROR_SCHEDULE_OVERLAP",
	PlaCollectionNotFound:              "PDH_PLA_COLLECTION_NOT_FOUND",
	PlaErrorScheduleElapsed:            "PDH_PLA_ERROR_SCHEDULE_ELAPSED",
	PlaErrorNostart:                    "PDH_PLA_ERROR_NOSTART",
	PlaErrorAlreadyExists:              "PDH_PLA_ERROR_ALREADY_EXISTS",
	PlaErrorTypeMismatch:               "PDH_PLA_ERROR_TYPE_MISMATCH",
	PlaErrorFilepath:                   "PDH_PLA_ERROR_FILEPATH",
	PlaServiceError:                    "PDH_PLA_SERVICE_ERROR",
	PlaValidationError:                 "PDH_PLA_VALIDATION_ERROR",
	PlaValidationWarning:               "PDH_PLA_VALIDATION_WARNING",
	PlaErrorNameTooLong:                "PDH_PLA_ERROR_NAME_TOO_LONG",
	InvalidSQLLogFormat:                "PDH_INVALID_SQL_LOG_FORMAT",
	CounterAlreadyInQuery:              "PDH_COUNTER_ALREADY_IN_QUERY",
	BinaryLogCorrupt:                   "PDH_BINARY_LOG_CORRUPT",
	LogSampleTooSmall:                  "PDH_LOG_SAMPLE_TOO_SMALL",
	OsLaterVersion:                     "PDH_OS_LATER_VERSION",
	OsEarlierVersion:                   "PDH_OS_EARLIER_VERSION",
	IncorrectAppendTime:                "PDH_INCORRECT_APPEND_TIME",
	UnmatchedAppendCounter:             "PDH_UNMATCHED_APPEND_COUNTER",
	SQLAlterDetailFailed:               "PDH_SQL_ALTER_DETAIL_FAILED",
	QueryPerfDataTimeout:               "PDH_QUERY_PERF_DATA_TIMEOUT",
}

// Formatting options for GetFormattedCounterValue().
//
//goland:noinspection GoUnusedConst
const (
	FmtRaw             = 0x00000010
	FmtAnsi            = 0x00000020
	FmtUnicode         = 0x00000040
	FmtLong            = 0x00000100 // Return data as a long int.
	FmtDouble          = 0x00000200 // Return data as a double precision floating point real.
	FmtLarge           = 0x00000400 // Return data as a 64 bit integer.
	FmtNoscale         = 0x00001000 // can be OR-ed: Do not apply the counter's default scaling factor.
	Fmt1000            = 0x00002000 // can be OR-ed: multiply the actual value by 1,000.
	FmtNodata          = 0x00004000 // can be OR-ed: unknown what this is for, MSDN says nothing.
	FmtNocap100        = 0x00008000 // can be OR-ed: do not cap values > 100.
	PerfDetailCostly   = 0x00010000
	PerfDetailStandard = 0x0000FFFF
)

type (
	pdhQueryHandle   HANDLE // query handle
	pdhCounterHandle HANDLE // counter handle
)

//nolint:gochecknoglobals
var (
	libPdhDll = windows.NewLazySystemDLL("pdh.dll")

	pdhAddCounterW               = libPdhDll.NewProc("PdhAddCounterW")
	pdhAddEnglishCounterW        = libPdhDll.NewProc("PdhAddEnglishCounterW")
	pdhCloseQuery                = libPdhDll.NewProc("PdhCloseQuery")
	pdhCollectQueryData          = libPdhDll.NewProc("PdhCollectQueryData")
	pdhCollectQueryDataWithTime  = libPdhDll.NewProc("PdhCollectQueryDataWithTime")
	pdhGetFormattedCounterValue  = libPdhDll.NewProc("PdhGetFormattedCounterValue")
	pdhGetFormattedCounterArrayW = libPdhDll.NewProc("PdhGetFormattedCounterArrayW")
	pdhOpenQuery                 = libPdhDll.NewProc("PdhOpenQuery")
	pdhValidatePathW             = libPdhDll.NewProc("PdhValidatePathW")
	pdhExpandWildCardPathW       = libPdhDll.NewProc("PdhExpandWildCardPathW")
	pdhGetCounterInfoW           = libPdhDll.NewProc("PdhGetCounterInfoW")
	pdhGetRawCounterValue        = libPdhDll.NewProc("PdhGetRawCounterValue")
	pdhGetRawCounterArrayW       = libPdhDll.NewProc("PdhGetRawCounterArrayW")
	pdhPdhGetCounterTimeBase     = libPdhDll.NewProc("PdhGetCounterTimeBase")
)

// AddCounter adds the specified counter to the query. This is the internationalized version. Preferably, use the
// function AddEnglishCounter instead. hQuery is the query handle, which has been fetched by OpenQuery.
// szFullCounterPath is a full, internationalized counter path (this will differ per Windows language version).
// dwUserData is a 'user-defined value', which becomes part of the counter information. To retrieve this value
// later, call GetCounterInfo() and access dwQueryUserData of the CounterInfo structure.
//
// Examples of szFullCounterPath (in an English version of Windows):
//
//	\\Processor(_Total)\\% Idle Time
//	\\Processor(_Total)\\% Processor Time
//	\\LogicalDisk(C:)\% Free Space
//
// To view all (internationalized...) counters on a system, there are three non-programmatic ways: perfmon utility,
// the typeperf command, and the v1 editor. perfmon.exe is perhaps the easiest way, because it's basically a
// full implementation of the pdh.dll API, except with a GUI and all that. The v1 setting also provides an
// interface to the available counters, and can be found at the following key:
//
//	HKEY_LOCAL_MACHINE\SOFTWARE\Microsoft\Windows NT\CurrentVersion\Perflib\CurrentLanguage
//
// This v1 key contains several values as follows:
//
//	1
//	1847
//	2
//	System
//	4
//	Memory
//	6
//	% Processor Time
//	... many, many more
//
// Somehow, these numeric values can be used as szFullCounterPath too:
//
//	\2\6 will correspond to \\System\% Processor Time
//
// The typeperf command may also be pretty easy. To find all performance counters, simply execute:
//
//	typeperf -qx
func AddCounter(hQuery pdhQueryHandle, szFullCounterPath string, dwUserData uintptr, phCounter *pdhCounterHandle) uint32 {
	ptxt, _ := windows.UTF16PtrFromString(szFullCounterPath)
	ret, _, _ := pdhAddCounterW.Call(
		uintptr(hQuery),
		uintptr(unsafe.Pointer(ptxt)),
		dwUserData,
		uintptr(unsafe.Pointer(phCounter)))

	return uint32(ret)
}

// AddEnglishCounter adds the specified language-neutral counter to the query. See the AddCounter function. This function only exists on
// Windows versions higher than Vista.
func AddEnglishCounter(hQuery pdhQueryHandle, szFullCounterPath string, dwUserData uintptr, phCounter *pdhCounterHandle) uint32 {
	if pdhAddEnglishCounterW == nil {
		return ErrorInvalidFunction
	}

	ptxt, _ := windows.UTF16PtrFromString(szFullCounterPath)
	ret, _, _ := pdhAddEnglishCounterW.Call(
		uintptr(hQuery),
		uintptr(unsafe.Pointer(ptxt)),
		dwUserData,
		uintptr(unsafe.Pointer(phCounter)))

	return uint32(ret)
}

// CloseQuery closes all counters contained in the specified query, closes all handles related to the query,
// and frees all memory associated with the query.
func CloseQuery(hQuery pdhQueryHandle) uint32 {
	ret, _, _ := pdhCloseQuery.Call(uintptr(hQuery))

	return uint32(ret)
}

// CollectQueryData collects the current raw data value for all counters in the specified query and updates the status
// code of each counter. With some counters, this function needs to be repeatedly called before the value
// of the counter can be extracted with PdhGetFormattedCounterValue(). For example, the following code
// requires at least two calls:
//
//	var handle win.PDH_HQUERY
//	var counterHandle win.PDH_HCOUNTER
//	ret := win.OpenQuery(0, 0, &handle)
//	ret = win.AddEnglishCounter(handle, "\\Processor(_Total)\\% Idle Time", 0, &counterHandle)
//	var derp win.PDH_FMT_COUNTERVALUE_DOUBLE
//
//	ret = win.CollectQueryData(handle)
//	fmt.Printf("Collect return code is %x\n", ret) // return code will be PDH_CSTATUS_INVALID_DATA
//	ret = win.GetFormattedCounterValueDouble(counterHandle, 0, &derp)
//
//	ret = win.CollectQueryData(handle)
//	fmt.Printf("Collect return code is %x\n", ret) // return code will be ERROR_SUCCESS
//	ret = win.GetFormattedCounterValueDouble(counterHandle, 0, &derp)
//
// The CollectQueryData will return an error in the first call because it needs two values for
// displaying the correct data for the processor idle time. The second call will have a 0 return code.
func CollectQueryData(hQuery pdhQueryHandle) uint32 {
	ret, _, _ := pdhCollectQueryData.Call(uintptr(hQuery))

	return uint32(ret)
}

// CollectQueryDataWithTime queries data from perfmon, retrieving the device/windows timestamp from the node it was collected on.
// Converts the filetime structure to a GO time class and returns the native time.
func CollectQueryDataWithTime(hQuery pdhQueryHandle) (uint32, time.Time) {
	var localFileTime windows.Filetime

	ret, _, _ := pdhCollectQueryDataWithTime.Call(uintptr(hQuery), uintptr(unsafe.Pointer(&localFileTime)))

	if ret == ErrorSuccess {
		var utcFileTime windows.Filetime

		if ret := kernel32.LocalFileTimeToFileTime(&localFileTime, &utcFileTime); ret == 0 {
			return uint32(ErrorFailure), time.Now()
		}

		retTime := time.Unix(0, utcFileTime.Nanoseconds())

		return uint32(ErrorSuccess), retTime
	}

	return uint32(ret), time.Now()
}

// GetFormattedCounterValueDouble formats the given hCounter using a 'double'. The result is set into the specialized union struct pValue.
// This function does not directly translate to a Windows counterpart due to union specialization tricks.
func GetFormattedCounterValueDouble(hCounter pdhCounterHandle, lpdwType *uint32, pValue *FmtCounterValueDouble) uint32 {
	ret, _, _ := pdhGetFormattedCounterValue.Call(
		uintptr(hCounter),
		uintptr(FmtDouble|FmtNocap100),
		uintptr(unsafe.Pointer(lpdwType)),
		uintptr(unsafe.Pointer(pValue)))

	return uint32(ret)
}

// GetFormattedCounterArrayDouble returns an array of formatted counter values. Use this function when you want to format the counter values of a
// counter that contains a wildcard character for the instance name. The itemBuffer must a slice of type FmtCounterValueItemDouble.
// An example of how this function can be used:
//
//	okPath := "\\Process(*)\\% Processor Time" // notice the wildcard * character
//
//	// omitted all necessary stuff ...
//
//	var bufSize uint32
//	var bufCount uint32
//	var size uint32 = uint32(unsafe.Sizeof(win.PDH_FMT_COUNTERVALUE_ITEM_DOUBLE{}))
//	var emptyBuf [1]win.PDH_FMT_COUNTERVALUE_ITEM_DOUBLE // need at least 1 addressable null ptr.
//
//	for {
//		// collect
//		ret := win.CollectQueryData(queryHandle)
//		if ret == win.ERROR_SUCCESS {
//			ret = win.GetFormattedCounterArrayDouble(counterHandle, &bufSize, &bufCount, &emptyBuf[0]) // uses null ptr here according to MSDN.
//			if ret == win.PDH_MORE_DATA {
//				filledBuf := make([]win.PDH_FMT_COUNTERVALUE_ITEM_DOUBLE, bufCount*size)
//				ret = win.GetFormattedCounterArrayDouble(counterHandle, &bufSize, &bufCount, &filledBuf[0])
//				for i := 0; i < int(bufCount); i++ {
//					c := filledBuf[i]
//					var s string = win.UTF16PtrToString(c.SzName)
//					fmt.Printf("Index %d -> %s, value %v\n", i, s, c.FmtValue.DoubleValue)
//				}
//
//				filledBuf = nil
//				// Need to at least set bufSize to zero, because if not, the function will not
//				// return PDH_MORE_DATA and will not set the bufSize.
//				bufCount = 0
//				bufSize = 0
//			}
//
//			time.Sleep(2000 * time.Millisecond)
//		}
//	}
func GetFormattedCounterArrayDouble(hCounter pdhCounterHandle, lpdwBufferSize *uint32, lpdwBufferCount *uint32, itemBuffer *byte) uint32 {
	ret, _, _ := pdhGetFormattedCounterArrayW.Call(
		uintptr(hCounter),
		uintptr(FmtDouble),
		uintptr(unsafe.Pointer(lpdwBufferSize)),
		uintptr(unsafe.Pointer(lpdwBufferCount)),
		uintptr(unsafe.Pointer(itemBuffer)))

	return uint32(ret)
}

// OpenQuery creates a new query that is used to manage the collection of performance data.
// szDataSource is a null terminated string that specifies the name of the log file from which to
// retrieve the performance data. If 0, performance data is collected from a real-time data source.
// dwUserData is a user-defined value to associate with this query. To retrieve the user data later,
// call GetCounterInfo and access dwQueryUserData of the CounterInfo structure. phQuery is
// the handle to the query, and must be used in subsequent calls. This function returns a PDH_
// constant error code, or ErrorSuccess if the call succeeded.
func OpenQuery(szDataSource uintptr, dwUserData uintptr, phQuery *pdhQueryHandle) uint32 {
	ret, _, _ := pdhOpenQuery.Call(
		szDataSource,
		dwUserData,
		uintptr(unsafe.Pointer(phQuery)))

	return uint32(ret)
}

// ExpandWildCardPath examines the specified computer or log file and returns those counter paths that match the given counter path
// which contains wildcard characters. The general counter path format is as follows:
//
// \\computer\object(parent/instance#index)\counter
//
// The parent, instance, index, and counter components of the counter path may contain either a valid name or a wildcard character.
// The computer, parent, instance, and index components are not necessary for all counters.
//
// The following is a list of the possible formats:
//
// \\computer\object(parent/instance#index)\counter
// \\computer\object(parent/instance)\counter
// \\computer\object(instance#index)\counter
// \\computer\object(instance)\counter
// \\computer\object\counter
// \object(parent/instance#index)\counter
// \object(parent/instance)\counter
// \object(instance#index)\counter
// \object(instance)\counter
// \object\counter
// Use an asterisk (*) as the wildcard character, for example, \object(*)\counter.
//
// If a wildcard character is specified in the parent name, all instances of the specified object
// that match the specified instance and counter fields will be returned.
// For example, \object(*/instance)\counter.
//
// If a wildcard character is specified in the instance name, all instances of the specified object and parent object will be returned if all instance names
// corresponding to the specified index match the wildcard character. For example, \object(parent/*)\counter.
// If the object does not contain an instance, an error occurs.
//
// If a wildcard character is specified in the counter name, all counters of the specified object are returned.
//
// Partial counter path string matches (for example, "pro*") are supported.
func ExpandWildCardPath(szWildCardPath string, mszExpandedPathList *uint16, pcchPathListLength *uint32) uint32 {
	ptxt, _ := windows.UTF16PtrFromString(szWildCardPath)
	flags := uint32(0) // expand instances and counters
	ret, _, _ := pdhExpandWildCardPathW.Call(
		0, // search counters on local computer
		uintptr(unsafe.Pointer(ptxt)),
		uintptr(unsafe.Pointer(mszExpandedPathList)),
		uintptr(unsafe.Pointer(pcchPathListLength)),
		uintptr(unsafe.Pointer(&flags)))

	return uint32(ret)
}

// ValidatePath validates a path. Will return ErrorSuccess when ok, or PdhCstatusBadCountername when the path is erroneous.
func ValidatePath(path string) uint32 {
	ptxt, _ := windows.UTF16PtrFromString(path)
	ret, _, _ := pdhValidatePathW.Call(uintptr(unsafe.Pointer(ptxt)))

	return uint32(ret)
}

func FormatError(msgID uint32) string {
	var flags uint32 = windows.FORMAT_MESSAGE_FROM_HMODULE | windows.FORMAT_MESSAGE_ARGUMENT_ARRAY | windows.FORMAT_MESSAGE_IGNORE_INSERTS

	buf := make([]uint16, 300)
	_, err := windows.FormatMessage(flags, libPdhDll.Handle(), msgID, 0, buf, nil)

	if err == nil {
		return windows.UTF16PtrToString(&buf[0])
	}

	return fmt.Sprintf("(pdhErr=%d) %s", msgID, err.Error())
}

// GetCounterInfo retrieves information about a counter, such as data size, counter type, path, and user-supplied data values
// hCounter [in]
// Handle of the counter from which you want to retrieve information. The AddCounter function returns this handle.
//
// bRetrieveExplainText [in]
// Determines whether explain text is retrieved. If you set this parameter to TRUE, the explain text for the counter is retrieved.
// If you set this parameter to FALSE, the field in the returned buffer is NULL.
//
// pdwBufferSize [in, out]
// Size of the lpBuffer buffer, in bytes. If zero on input, the function returns PdhMoreData and sets this parameter to the required buffer size.
// If the buffer is larger than the required size, the function sets this parameter to the actual size of the buffer that was used.
// If the specified size on input is greater than zero but less than the required size, you should not rely on the returned size to reallocate the buffer.
//
// lpBuffer [out]
// Caller-allocated buffer that receives a CounterInfo structure.
// The structure is variable-length, because the string data is appended to the end of the fixed-format portion of the structure.
// This is done so that all data is returned in a single buffer allocated by the caller. Set to NULL if pdwBufferSize is zero.
func GetCounterInfo(hCounter pdhCounterHandle, bRetrieveExplainText int, pdwBufferSize *uint32, lpBuffer *byte) uint32 {
	ret, _, _ := pdhGetCounterInfoW.Call(
		uintptr(hCounter),
		uintptr(bRetrieveExplainText),
		uintptr(unsafe.Pointer(pdwBufferSize)),
		uintptr(unsafe.Pointer(lpBuffer)))

	return uint32(ret)
}

// GetRawCounterValue returns the current raw value of the counter.
// If the specified counter instance does not exist, this function will return ErrorSuccess
// and the CStatus member of the RawCounter structure will contain PdhCstatusNoInstance.
//
// hCounter [in]
// Handle of the counter from which to retrieve the current raw value. The AddCounter function returns this handle.
//
// lpdwType [out]
// Receives the counter type. For a list of counter types, see the Counter Types section of the Windows Server 2003 Deployment Kit.
// This parameter is optional.
//
// pValue [out]
// A RawCounter structure that receives the counter value.
func GetRawCounterValue(hCounter pdhCounterHandle, lpdwType *uint32, pValue *RawCounter) uint32 {
	ret, _, _ := pdhGetRawCounterValue.Call(
		uintptr(hCounter),
		uintptr(unsafe.Pointer(lpdwType)),
		uintptr(unsafe.Pointer(pValue)))

	return uint32(ret)
}

// GetRawCounterArray returns an array of raw values from the specified counter. Use this function when you want to retrieve the raw counter values
// of a counter that contains a wildcard character for the instance name.
// hCounter
// Handle of the counter for whose current raw instance values you want to retrieve. The AddCounter function returns this handle.
//
// lpdwBufferSize
// Size of the ItemBuffer buffer, in bytes. If zero on input, the function returns PdhMoreData and sets this parameter to the required buffer size.
// If the buffer is larger than the required size, the function sets this parameter to the actual size of the buffer that was used.
// If the specified size on input is greater than zero but less than the required size, you should not rely on the returned size to reallocate the buffer.
//
// lpdwItemCount
// Number of raw counter values in the ItemBuffer buffer.
//
// ItemBuffer
// Caller-allocated buffer that receives the array of RawCounterItem structures; the structures contain the raw instance counter values.
// Set to NULL if lpdwBufferSize is zero.
func GetRawCounterArray(hCounter pdhCounterHandle, lpdwBufferSize *uint32, lpdwBufferCount *uint32, itemBuffer *byte) uint32 {
	ret, _, _ := pdhGetRawCounterArrayW.Call(
		uintptr(hCounter),
		uintptr(unsafe.Pointer(lpdwBufferSize)),
		uintptr(unsafe.Pointer(lpdwBufferCount)),
		uintptr(unsafe.Pointer(itemBuffer)))

	return uint32(ret)
}

// GetCounterTimeBase returns the time base of the specified counter.
// hCounter
// Handle of the counter for whose current raw instance values you want to retrieve. The AddCounter function returns this handle.
//
// lpdwItemCount
// Time base that specifies the number of performance values a counter samples per second.
func GetCounterTimeBase(hCounter pdhCounterHandle, pTimeBase *int64) uint32 {
	ret, _, _ := pdhPdhGetCounterTimeBase.Call(
		uintptr(hCounter),
		uintptr(unsafe.Pointer(pTimeBase)))

	return uint32(ret)
}
