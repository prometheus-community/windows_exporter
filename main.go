package main

import (
	"fmt"
	"unsafe"

	"github.com/go-ole/go-ole"
	"github.com/prometheus-community/windows_exporter/internal/headers/win32"
	"golang.org/x/sys/windows"
)

const (
	// Configuration Manager return codes
	CR_SUCCESS                  = 0x00
	CR_DEFAULT                  = 0x01
	CR_OUT_OF_MEMORY            = 0x02
	CR_INVALID_POINTER          = 0x03
	CR_INVALID_FLAG             = 0x04
	CR_INVALID_DEVNODE          = 0x05
	CR_INVALID_DEVINST          = CR_INVALID_DEVNODE
	CR_INVALID_RES_DES          = 0x06
	CR_INVALID_LOG_CONF         = 0x07
	CR_INVALID_ARBITRATOR       = 0x08
	CR_INVALID_NODELIST         = 0x09
	CR_DEVNODE_HAS_REQS         = 0x0a
	CR_DEVINST_HAS_REQS         = CR_DEVNODE_HAS_REQS
	CR_INVALID_RESOURCEID       = 0x0b
	CR_DLVXD_NOT_FOUND          = 0x0c
	CR_NO_SUCH_DEVNODE          = 0x0d
	CR_NO_SUCH_DEVINST          = CR_NO_SUCH_DEVNODE
	CR_NO_MORE_LOG_CONF         = 0x0e
	CR_NO_MORE_RES_DES          = 0x0f
	CR_ALREADY_SUCH_DEVNODE     = 0x10
	CR_ALREADY_SUCH_DEVINST     = CR_ALREADY_SUCH_DEVNODE
	CR_INVALID_RANGE_LIST       = 0x11
	CR_INVALID_RANGE            = 0x12
	CR_FAILURE                  = 0x13
	CR_NO_SUCH_LOGICAL_DEV      = 0x14
	CR_CREATE_BLOCKED           = 0x15
	CR_NOT_SYSTEM_VM            = 0x16
	CR_REMOVE_VETOED            = 0x17
	CR_APM_VETOED               = 0x18
	CR_INVALID_LOAD_TYPE        = 0x19
	CR_BUFFER_SMALL             = 0x1a
	CR_NO_ARBITRATOR            = 0x1b
	CR_NO_REGISTRY_HANDLE       = 0x1c
	CR_REGISTRY_ERROR           = 0x1d
	CR_INVALID_DEVICE_ID        = 0x1e
	CR_INVALID_DATA             = 0x1f
	CR_INVALID_API              = 0x20
	CR_DEVLOADER_NOT_READY      = 0x21
	CR_NEED_RESTART             = 0x22
	CR_NO_MORE_HW_PROFILES      = 0x23
	CR_DEVICE_NOT_THERE         = 0x24
	CR_NO_SUCH_VALUE            = 0x25
	CR_WRONG_TYPE               = 0x26
	CR_INVALID_PRIORITY         = 0x27
	CR_NOT_DISABLEABLE          = 0x28
	CR_FREE_RESOURCES           = 0x29
	CR_QUERY_VETOED             = 0x2a
	CR_CANT_SHARE_IRQ           = 0x2b
	CR_NO_DEPENDENT             = 0x2c
	CR_SAME_RESOURCES           = 0x2d
	CR_NO_SUCH_REGISTRY_KEY     = 0x2e
	CR_INVALID_MACHINENAME      = 0x2f
	CR_REMOTE_COMM_FAILURE      = 0x30
	CR_MACHINE_UNAVAILABLE      = 0x31
	CR_NO_CM_SERVICES           = 0x32
	CR_ACCESS_DENIED            = 0x33
	CR_CALL_NOT_IMPLEMENTED     = 0x34
	CR_INVALID_PROPERTY         = 0x35
	CR_DEVICE_INTERFACE_ACTIVE  = 0x36
	CR_NO_SUCH_DEVICE_INTERFACE = 0x37
	CR_INVALID_REFERENCE_STRING = 0x38
	CR_INVALID_CONFLICT_LIST    = 0x39
	CR_INVALID_INDEX            = 0x3a
	CR_INVALID_STRUCTURE_SIZE   = 0x3b

	// Filter flags
	CM_GETIDLIST_FILTER_NONE               = 0x00000000
	CM_GETIDLIST_FILTER_ENUMERATOR         = 0x00000001
	CM_GETIDLIST_FILTER_SERVICE            = 0x00000002
	CM_GETIDLIST_FILTER_EJECTRELATIONS     = 0x00000004
	CM_GETIDLIST_FILTER_REMOVALRELATIONS   = 0x00000008
	CM_GETIDLIST_FILTER_POWERRELATIONS     = 0x00000010
	CM_GETIDLIST_FILTER_BUSRELATIONS       = 0x00000020
	CM_GETIDLIST_DONOTGENERATE             = 0x10000040
	CM_GETIDLIST_FILTER_PRESENT            = 0x00000100
	CM_GETIDLIST_FILTER_CLASS              = 0x00000200
	CM_GETIDLIST_FILTER_TRANSPORTRELATIONS = 0x00000080
	CM_GETIDLIST_FILTER_BITS               = 0x100003FF

	// Device registry properties
	CM_DRP_DEVICEDESC           = 0x00000001
	CM_DRP_HARDWAREID           = 0x00000002
	CM_DRP_LOCATION_INFORMATION = 0x0000000E
	CM_DRP_BUSNUMBER            = 0x00000016
	CM_DRP_ADDRESS              = 0x0000001C

	// Constants
	MAX_DEVICE_ID_LEN = 200

	DEVPROP_TYPEMOD_ARRAY uint32 = 0x00001000
	DEVPROP_TYPEMOD_LIST  uint32 = 0x00002000

	DEVPROP_TYPE_EMPTY    uint32 = 0x00000000
	DEVPROP_TYPE_NULL     uint32 = 0x00000001
	DEVPROP_TYPE_SBYTE    uint32 = 0x00000002
	DEVPROP_TYPE_BYTE     uint32 = 0x00000003
	DEVPROP_TYPE_INT16    uint32 = 0x00000004
	DEVPROP_TYPE_UINT16   uint32 = 0x00000005
	DEVPROP_TYPE_INT32    uint32 = 0x00000006
	DEVPROP_TYPE_UINT32   uint32 = 0x00000007
	DEVPROP_TYPE_INT64    uint32 = 0x00000008
	DEVPROP_TYPE_UINT64   uint32 = 0x00000009
	DEVPROP_TYPE_FLOAT    uint32 = 0x0000000A
	DEVPROP_TYPE_DOUBLE   uint32 = 0x0000000B
	DEVPROP_TYPE_DECIMAL  uint32 = 0x0000000C
	DEVPROP_TYPE_GUID     uint32 = 0x0000000D
	DEVPROP_TYPE_CURRENCY uint32 = 0x0000000E
	DEVPROP_TYPE_DATE     uint32 = 0x0000000F
	DEVPROP_TYPE_FILETIME uint32 = 0x00000010
	DEVPROP_TYPE_BOOLEAN  uint32 = 0x00000011
	DEVPROP_TYPE_STRING   uint32 = 0x00000012

	DEVPROP_TYPE_STRING_LIST uint32 = DEVPROP_TYPE_STRING | DEVPROP_TYPEMOD_LIST

	DEVPROP_TYPE_SECURITY_DESCRIPTOR        uint32 = 0x00000013
	DEVPROP_TYPE_SECURITY_DESCRIPTOR_STRING uint32 = 0x00000014
	DEVPROP_TYPE_DEVPROPKEY                 uint32 = 0x00000015
	DEVPROP_TYPE_DEVPROPTYPE                uint32 = 0x00000016

	DEVPROP_TYPE_BINARY uint32 = DEVPROP_TYPE_BYTE | DEVPROP_TYPEMOD_ARRAY

	DEVPROP_TYPE_ERROR           uint32 = 0x00000017
	DEVPROP_TYPE_NTSTATUS        uint32 = 0x00000018
	DEVPROP_TYPE_STRING_INDIRECT uint32 = 0x00000019
)

var (
	cfgmgr32                  = windows.NewLazySystemDLL("cfgmgr32.dll")
	procCMGetDeviceIDListW    = cfgmgr32.NewProc("CM_Get_Device_ID_ListW")
	procCMGetDeviceIDListSize = cfgmgr32.NewProc("CM_Get_Device_ID_List_SizeW")
	procCMLocateDevNodeW      = cfgmgr32.NewProc("CM_Locate_DevNodeW")
	procCMGetDevNodePropertyW = cfgmgr32.NewProc("CM_Get_DevNode_PropertyW")
)

// DEVPROPKEY represents a device property key (GUID + pid)
type DEVPROPKEY struct {
	FmtID ole.GUID
	PID   uint32
}

// https://github.com/Infinidat/infi.devicemanager/blob/8be9ead6b04ff45c63d9e3bc70d82cceafb75c47/src/infi/devicemanager/setupapi/properties.py#L138C1-L143C34
var (
	DEVPKEY_Device_BusNumber = DEVPROPKEY{
		FmtID: ole.GUID{
			Data1: 0xa45c254e,
			Data2: 0xdf1c,
			Data3: 0x4efd,
			Data4: [8]byte{0x80, 0x20, 0x67, 0xd1, 0x46, 0xa8, 0x50, 0xe0},
		},
		PID: 23, // DEVPROP_TYPE_UINT32
	}

	DEVPKEY_Device_Address = DEVPROPKEY{
		FmtID: ole.GUID{
			Data1: 0xa45c254e,
			Data2: 0xdf1c,
			Data3: 0x4efd,
			Data4: [8]byte{0x80, 0x20, 0x67, 0xd1, 0x46, 0xa8, 0x50, 0xe0},
		},
		PID: 30, // DEVPROP_TYPE_UINT32
	}
)

// https://github.com/XZiar/RayRenderer/blob/4645d4e5b1b24e04576dac5800d68d8929a41042/XComputeBase/DeviceDiscoveryWin32.cpp#L163
func main() {
	var size uint32

	ret, _, lastErr := procCMGetDeviceIDListSize.Call(
		uintptr(unsafe.Pointer(&size)),
		win32.NewLPWSTR("PCI\\VEN_10DE&DEV_1B81&SUBSYS_61733842&REV_A1").Pointer(),
		uintptr(CM_GETIDLIST_FILTER_PRESENT|CM_GETIDLIST_FILTER_ENUMERATOR),
	)

	if ret != CR_SUCCESS {
		fmt.Printf("Return: 0x%02X\n", ret)
		fmt.Printf("Size: %d\n", size)
		fmt.Printf("LastError: %v\n", lastErr)

		return
	}

	buf := make([]uint16, size)
	ret, _, lastErr = procCMGetDeviceIDListW.Call(
		win32.NewLPWSTR("PCI\\VEN_10DE&DEV_1B81&SUBSYS_61733842&REV_A1").Pointer(),
		uintptr(unsafe.Pointer(&buf[0])),
		uintptr(size),
		uintptr(CM_GETIDLIST_FILTER_PRESENT|CM_GETIDLIST_FILTER_ENUMERATOR),
	)

	if ret != CR_SUCCESS {
		fmt.Printf("Return: 0x%02X\n", ret)
		fmt.Printf("Size: %d\n", size)
		fmt.Printf("LastError: %v\n", lastErr)

		return
	}

	for _, deviceID := range multiSzToStrings(buf) {
		var handle *windows.Handle

		ret, _, lastErr := procCMLocateDevNodeW.Call(
			uintptr(unsafe.Pointer(&handle)),
			uintptr(unsafe.Pointer(&deviceID[0])),
			uintptr(0),
		)

		if ret != CR_SUCCESS {
			fmt.Printf("Return: 0x%02X\n", ret)
			fmt.Printf("Size: %d\n", size)
			fmt.Printf("LastError: %v\n", lastErr)

			return
		}

		var propType uint32
		var buf uint32
		bufLen := uint32(4)

		ret, _, lastErr = procCMGetDevNodePropertyW.Call(
			uintptr(unsafe.Pointer(handle)),
			uintptr(unsafe.Pointer(&DEVPKEY_Device_BusNumber)),
			uintptr(unsafe.Pointer(&propType)),
			uintptr(unsafe.Pointer(&buf)),
			uintptr(unsafe.Pointer(&bufLen)),
			0,
		)

		if ret != CR_SUCCESS {
			fmt.Printf("Return: 0x%02X\n", ret)
			fmt.Printf("Size: %d\n", size)
			fmt.Printf("LastError: %v\n", lastErr)

			return
		}

		if propType != DEVPROP_TYPE_UINT32 {
			fmt.Printf("Unexpected property type: 0x%08X\n", propType)
			return
		}

		fmt.Printf("BusNumber: %d\n", buf)
	}
}

func multiSzToStrings(buf []uint16) [][]uint16 {
	var result [][]uint16
	start := 0

	for i := 0; i < len(buf); i++ {
		if buf[i] == 0 {
			// Wenn wir auf ein Nullterminierungszeichen stoßen
			if i == start {
				// Zwei aufeinanderfolgende Nulls -> Ende der Liste
				break
			}
			// Slice vom letzten Start bis vor das Null hinzufügen
			result = append(result, buf[start:i])
			start = i + 1
		}
	}

	return result
}
