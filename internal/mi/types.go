package mi

import (
	"time"
	"unsafe"

	"golang.org/x/sys/windows"
)

type Boolean uint8

const (
	False Boolean = 0
	True  Boolean = 1
)

type QueryDialect = *uint16

var (
	QueryDialectWQL = UTF16PtrFromString[QueryDialect]("WQL")
	QueryDialectCQL = UTF16PtrFromString[QueryDialect]("CQL")
)

type Namespace = *uint16

var (
	NamespaceRootCIMv2             = UTF16PtrFromString[Namespace]("root/cimv2")
	NamespaceRootWindowsFSRM       = UTF16PtrFromString[Namespace]("root/microsoft/windows/fsrm")
	NamespaceRootWebAdministration = UTF16PtrFromString[Namespace]("root/WebAdministration")
	NamespaceRootMSCluster         = UTF16PtrFromString[Namespace]("root/MSCluster")
)

// UTF16PtrFromString converts a string to a UTF-16 pointer at initialization time.
//
//nolint:ireturn
func UTF16PtrFromString[T *uint16](s string) T {
	val, err := windows.UTF16PtrFromString(s)
	if err != nil {
		panic(err)
	}

	return val
}

type Timestamp struct {
	Year         uint32
	Month        uint32
	Day          uint32
	Hour         uint32
	Minute       uint32
	Second       uint32
	Microseconds uint32
	UTC          int32
}

type Interval struct {
	Days         uint32
	Hours        uint32
	Minutes      uint32
	Seconds      uint32
	Microseconds uint32
	Padding1     uint32
	Padding2     uint32
	Padding3     uint32
}

func NewInterval(interval time.Duration) *Interval {
	// Convert the duration to a number of microseconds
	microseconds := interval.Microseconds()

	// Create a new interval with the microseconds
	return &Interval{
		Days:         uint32(microseconds / (24 * 60 * 60 * 1000000)),
		Hours:        uint32(microseconds / (60 * 60 * 1000000)),
		Minutes:      uint32(microseconds / (60 * 1000000)),
		Seconds:      uint32(microseconds / 1000000),
		Microseconds: uint32(microseconds % 1000000),
	}
}

type Datetime struct {
	IsTimestamp bool
	Timestamp   *Timestamp // Used when IsTimestamp is true
	Interval    *Interval  // Used when IsTimestamp is false
}

type PropertyDecl struct {
	Flags         uint32
	Code          uint32
	Name          *uint16
	Mqualifiers   uintptr
	NumQualifiers uint32
	PropertyType  ValueType
	ClassName     *uint16
	Subscript     uint32
	Offset        uint32
	Origin        *uint16
	Propagator    *uint16
	Value         uintptr
}

func (c *ClassDecl) Properties() []*PropertyDecl {
	// Create a slice to hold the properties
	properties := make([]*PropertyDecl, c.NumProperties)

	// Mproperties is a pointer to an array of pointers to PropertyDecl
	propertiesArray := (**PropertyDecl)(unsafe.Pointer(c.Mproperties))

	// Iterate over the number of properties and fetch each property
	for i := range c.NumProperties {
		// Get the property pointer at index i
		propertyPtr := *(**PropertyDecl)(unsafe.Pointer(uintptr(unsafe.Pointer(propertiesArray)) + uintptr(i)*unsafe.Sizeof(uintptr(0))))

		// Append the property to the slice
		properties[i] = propertyPtr
	}

	// Return the slice of properties
	return properties
}
