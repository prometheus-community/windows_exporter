//go:build windows

package main

import (
	"fmt"
	"log"
	"maps"
	"math"
	"reflect"
	"slices"
	"strings"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	modkernel32 = syscall.NewLazyDLL("kernel32.dll")
	modntdll    = syscall.NewLazyDLL("ntdll.dll")
	modMi       = windows.NewLazySystemDLL("mi.dll")

	procGetProcessHandleCount   = modkernel32.NewProc("GetProcessHandleCount")
	procNtQueryObject           = modntdll.NewProc("NtQueryObject")
	procMIApplicationInitialize = modMi.NewProc("MI_Application_InitializeV1")
)

// SystemHandleEntry has been derived from the SYSTEM_HANDLE_ENTRY struct definition.
type SystemHandleEntry struct {
	OwnerPid      uint32
	ObjectType    byte
	HandleFlags   byte
	HandleValue   uint16
	ObjectPointer *byte
	AccessMask    uint32
}

// SystemHandleInformationT has been derived from the SYSTEM_HANDLE_INFORMATION struct definition.
type SystemHandleInformationT struct {
	Count   uint32
	Handles [1]SystemHandleEntry
}

type ObjectTypeInformationT struct {
	TypeName                   UNICODE_STRING
	TotalNumberOfObjects       uint32
	TotalNumberOfHandles       uint32
	TotalPagedPoolUsage        uint32
	TotalNonPagedPoolUsage     uint32
	TotalNamePoolUsage         uint32
	TotalHandleTableUsage      uint32
	HighWaterNumberOfObjects   uint32
	HighWaterNumberOfHandles   uint32
	HighWaterPagedPoolUsage    uint32
	HighWaterNonPagedPoolUsage uint32
	HighWaterNamePoolUsage     uint32
	HighWaterHandleTableUsage  uint32
	InvalidAttributes          uint32
	GenericMapping             uintptr
	ValidAccessMask            uint32
	SecurityRequired           bool
	MaintainHandleCount        bool
	TypeIndex                  byte
	ReservedByte               byte
	PoolType                   uint32
	DefaultPagedPoolCharge     uint32
	DefaultNonPagedPoolCharge  uint32
}

type UNICODE_STRING struct {
	Length        uint16
	AllocatedSize uint16
	WString       *byte
}

func (self UNICODE_STRING) String() string {
	defer recover()

	var data []uint16

	sh := (*reflect.SliceHeader)(unsafe.Pointer(&data))
	sh.Data = uintptr(unsafe.Pointer(self.WString))
	sh.Len = int(self.Length * 2)
	sh.Cap = int(self.Length * 2)

	return windows.UTF16ToString(data[:])
}

func printHandleTypeStats() {
	var handles []SystemHandleEntry
	for bufferSize := uint32(128 * 1024); ; {
		moduleBuffer := make([]byte, bufferSize)
		err := windows.NtQuerySystemInformation(windows.SystemHandleInformation, unsafe.Pointer(&moduleBuffer[0]), bufferSize, &bufferSize)
		switch err {
		case windows.STATUS_INFO_LENGTH_MISMATCH:
			continue
		case nil:
			break
		default:
			return
		}
		mods := (*SystemHandleInformationT)(unsafe.Pointer(&moduleBuffer[0]))
		handles = unsafe.Slice(&mods.Handles[0], mods.Count)
		break
	}

	currentPID := windows.GetCurrentProcessId()

	handleSummary := map[string]uint32{}

	for _, handle := range handles {
		if handle.OwnerPid == currentPID {
			objectInfo := NtQueryObject(windows.Handle(handle.HandleValue))

			handleSummary[objectInfo.TypeName.String()]++
		}
	}

	fmt.Printf("Handle type summary:\n")
	for _, handleType := range slices.Sorted(maps.Keys(handleSummary)) {
		fmt.Printf("%s:%s%d\n", handleType, strings.Repeat("\t", int(math.Max(0, float64(3-(len(handleType)/7))))), handleSummary[handleType])
	}
}

func GetProcessHandleCount(handle windows.Handle) uint32 {
	var count uint32
	r1, _, err := procGetProcessHandleCount.Call(
		uintptr(handle),
		uintptr(unsafe.Pointer(&count)),
	)
	if r1 != 1 {
		panic(err)
	} else {
		return count
	}
}

func NtQueryObject(handle windows.Handle) *ObjectTypeInformationT {
	var returnLength uint32

	buffer := make([]byte, 32*1024)

	r1, _, err := procNtQueryObject.Call(
		uintptr(handle),
		uintptr(2),
		uintptr(unsafe.Pointer(&buffer[0])),
		1024,
		uintptr(unsafe.Pointer(&returnLength)),
	)
	if r1 != 0 {
		panic(err)
	} else {
		return (*ObjectTypeInformationT)(unsafe.Pointer(&buffer[0]))
	}
}

// Application represents the MI application.
// https://learn.microsoft.com/de-de/windows/win32/api/mi/ns-mi-mi_application
type Application struct {
	reserved1 uint64
	reserved2 uintptr
	ft        *ApplicationFT
}

// ApplicationFT represents the function table of the MI application.
// https://learn.microsoft.com/de-de/windows/win32/api/mi/ns-mi-mi_applicationft
type ApplicationFT struct {
	Close                          uintptr
	NewSession                     uintptr
	NewHostedProvider              uintptr
	NewInstance                    uintptr
	NewDestinationOptions          uintptr
	NewOperationOptions            uintptr
	NewSubscriptionDeliveryOptions uintptr
	NewSerializer                  uintptr
	NewDeserializer                uintptr
	NewInstanceFromClass           uintptr
	NewClass                       uintptr
}

// Session represents a session.
//
// https://learn.microsoft.com/en-us/windows/win32/api/mi/ns-mi-mi_session
type Session struct {
	reserved1 uint64
	reserved2 uintptr
	ft        *SessionFT
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

type Instance struct {
	ft         *InstanceFT
	classDecl  uintptr
	serverName *uint16
	nameSpace  *uint16
	_          [4]uintptr
}

type InstanceFT struct {
	Clone           uintptr
	Destruct        uintptr
	Delete          uintptr
	IsA             uintptr
	GetClassName    uintptr
	SetNameSpace    uintptr
	GetNameSpace    uintptr
	GetElementCount uintptr
	AddElement      uintptr
	SetElement      uintptr
	SetElementAt    uintptr
	GetElement      uintptr
	GetElementAt    uintptr
	ClearElement    uintptr
	ClearElementAt  uintptr
	GetServerName   uintptr
	SetServerName   uintptr
	GetClass        uintptr
}

// Application_Initialize initializes the MI [Application].
// It is recommended to have only one Application per process.
//
// https://learn.microsoft.com/en-us/windows/win32/api/mi/nf-mi-mi_application_initializev1
func Application_Initialize() (*Application, error) {
	application := &Application{}

	r0, _, _ := procMIApplicationInitialize.Call(0, 0, 0, uintptr(unsafe.Pointer(application)))

	if r0 != 0 {
		return nil, fmt.Errorf("failed: %d", r0)
	}

	return application, nil
}

// Close deinitializes the management infrastructure client API that was initialized through a call to Application_Initialize.
//
// https://learn.microsoft.com/en-us/windows/win32/api/mi/nf-mi-mi_application_close
func (application *Application) Close() error {
	r0, _, _ := syscall.SyscallN(application.ft.Close, uintptr(unsafe.Pointer(application)))

	if r0 != 0 {
		return fmt.Errorf("failed: %d", r0)
	}

	return nil
}

// NewSession creates a session used to share connections for a set of operations to a single destination.
//
// https://learn.microsoft.com/en-us/windows/win32/api/mi/nf-mi-mi_application_newsession
func (application *Application) NewSession() (*Session, error) {
	session := &Session{}

	r0, _, _ := syscall.SyscallN(
		application.ft.NewSession, uintptr(unsafe.Pointer(application)), 0, 0, 0, 0, 0, uintptr(unsafe.Pointer(session)),
	)

	if r0 != 0 {
		return nil, fmt.Errorf("failed: %d", r0)
	}

	return session, nil
}

// Close closes a session and releases all associated memory.
//
// https://learn.microsoft.com/en-us/windows/win32/api/mi/nf-mi-mi_session_close
func (s *Session) Close() error {
	r0, _, _ := syscall.SyscallN(s.ft.Close, uintptr(unsafe.Pointer(s)), 0, 0)

	if r0 != 0 {
		return fmt.Errorf("failed: %d", r0)
	}

	return nil
}

// QueryInstances queries for a set of instances based on a query expression.
//
// https://learn.microsoft.com/en-us/windows/win32/api/mi/nf-mi-mi_session_queryinstances
func (s *Session) QueryInstances(namespaceName, queryDialect, queryExpression string) (*Operation, error) {
	namespaceNameUTF16, err := windows.UTF16PtrFromString(namespaceName)
	if err != nil {
		return nil, err
	}

	queryDialectUTF16, err := windows.UTF16PtrFromString(queryDialect)
	if err != nil {
		return nil, err
	}

	queryExpressionUTF16, err := windows.UTF16PtrFromString(queryExpression)
	if err != nil {
		return nil, err
	}

	operation := &Operation{}

	r0, _, _ := syscall.SyscallN(
		s.ft.QueryInstances,
		uintptr(unsafe.Pointer(s)),
		0,
		0,
		uintptr(unsafe.Pointer(namespaceNameUTF16)),
		uintptr(unsafe.Pointer(queryDialectUTF16)),
		uintptr(unsafe.Pointer(queryExpressionUTF16)),
		0,
		uintptr(unsafe.Pointer(operation)),
	)

	if r0 != 0 {
		return nil, fmt.Errorf("failed: %d", r0)
	}

	return operation, nil
}

// Close closes an operation handle.
//
// https://learn.microsoft.com/en-us/windows/win32/api/mi/nf-mi-mi_operation_close
func (o *Operation) Close() error {
	r0, _, _ := syscall.SyscallN(o.ft.Close, uintptr(unsafe.Pointer(o)))

	if r0 != 0 {
		return fmt.Errorf("failed: %d", r0)
	}

	return nil
}

func (o *Operation) GetInstance() (*Instance, bool, error) {
	var (
		instance          *Instance
		moreResults       uint8
		instanceResult    uint32
		errorMessageUTF16 *uint16
	)

	r0, _, _ := syscall.SyscallN(
		o.ft.GetInstance,
		uintptr(unsafe.Pointer(o)),
		uintptr(unsafe.Pointer(&instance)),
		uintptr(unsafe.Pointer(&moreResults)),
		uintptr(unsafe.Pointer(&instanceResult)),
		uintptr(unsafe.Pointer(&errorMessageUTF16)),
		0,
	)

	if instanceResult != 0 {
		return nil, false, fmt.Errorf("instance result: %d (%s)", instanceResult, windows.UTF16PtrToString(errorMessageUTF16))
	}

	if r0 != 0 {
		return nil, false, fmt.Errorf("failed: %d", r0)
	}

	return instance, moreResults == 1, nil
}

func (instance *Instance) GetElement(elementName string) (uintptr, error) {
	elementNameUTF16, err := windows.UTF16PtrFromString(elementName)
	if err != nil {
		return 0, fmt.Errorf("failed to convert element name %s to UTF-16: %w", elementName, err)
	}

	var value uintptr

	r0, _, _ := syscall.SyscallN(
		instance.ft.GetElement,
		uintptr(unsafe.Pointer(instance)),
		uintptr(unsafe.Pointer(elementNameUTF16)),
		uintptr(unsafe.Pointer(&value)),
		0,
		0,
		0,
	)

	if r0 != 0 {
		return 0, fmt.Errorf("failed: %d", r0)
	}

	return value, nil
}

func (instance *Instance) Delete() error {
	r0, _, _ := syscall.SyscallN(instance.ft.Delete, uintptr(unsafe.Pointer(instance)))

	if r0 != 0 {
		return fmt.Errorf("failed: %d", r0)
	}

	return nil
}

func main() {
	log.Printf("Process handle count: %d", GetProcessHandleCount(windows.CurrentProcess()))

	app, err := Application_Initialize()
	if err != nil {
		panic(err)
	}

	session, err := app.NewSession()
	if err != nil {
		panic(err)
	}

	for range 1000 {
		operation, err := session.QueryInstances("root/cimv2", "WQL",
			"SELECT Architecture, DeviceId, Description, Family, L2CacheSize, L3CacheSize, Name, ThreadCount, NumberOfCores, NumberOfEnabledCore, NumberOfLogicalProcessors FROM Win32_Processor",
		)
		if err != nil {
			panic(err)
		}

		for {
			instance, moreResults, err := operation.GetInstance()
			if err != nil {
				panic(err)
			}

			_, _ = instance.GetElement("Name")

			if !moreResults {
				break
			}
		}

		if err = operation.Close(); err != nil {
			panic(err)
		}
	}

	if err = session.Close(); err != nil {
		panic(err)
	}

	if err = app.Close(); err != nil {
		panic(err)
	}

	log.Printf("Process handle count: %d", GetProcessHandleCount(windows.CurrentProcess()))
	printHandleTypeStats()
	log.Printf("Process handle count: %d", GetProcessHandleCount(windows.CurrentProcess()))
}
