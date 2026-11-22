// SPDX-License-Identifier: MIT
//
// Copyright The Prometheus Authors

//go:build windows

// Package ntdsapi reads the replication state of a domain controller through
// the directory service API (ntdsapi.dll), the source that repadmin /showrepl
// uses. The structures follow ntdsapi.h:
// https://learn.microsoft.com/en-us/windows/win32/api/ntdsapi/ns-ntdsapi-ds_repl_neighborw
package ntdsapi

import (
	"errors"
	"fmt"
	"syscall"
	"time"
	"unsafe"

	"github.com/prometheus-community/windows_exporter/internal/headers/lazyproc"
	"golang.org/x/sys/windows"
)

//nolint:gochecknoglobals
var (
	modntdsapi = windows.NewLazySystemDLL("ntdsapi.dll")

	procDsBindW           = modntdsapi.NewProc("DsBindW")
	procDsUnBindW         = modntdsapi.NewProc("DsUnBindW")
	procDsReplicaGetInfoW = modntdsapi.NewProc("DsReplicaGetInfoW")
	procDsReplicaFreeInfo = modntdsapi.NewProc("DsReplicaFreeInfo")

	load = lazyproc.Once(modntdsapi,
		procDsBindW,
		procDsUnBindW,
		procDsReplicaGetInfoW,
		procDsReplicaFreeInfo,
	)
)

// Load checks that ntdsapi.dll exports every function this package calls.
// Call it in Build: a missing export panics on the first call otherwise.
func Load() error { return load() }

// dsReplInfoNeighbors is DS_REPL_INFO_NEIGHBORS of DS_REPL_INFO_TYPE.
const dsReplInfoNeighbors = 0

// replNeighbor is DS_REPL_NEIGHBORW: 144 bytes on amd64 and arm64, 128 on
// 386. The 64-bit fields start at a multiple of 8 on every architecture, so
// Go and C agree although Go aligns int64 to 4 bytes on 386. FILETIME is two
// 32-bit fields.
type replNeighbor struct {
	pszNamingContext                   *uint16
	pszSourceDsaDN                     *uint16
	pszSourceDsaAddress                *uint16
	pszAsyncIntersiteTransportDN       *uint16
	dwReplicaFlags                     uint32
	dwReserved                         uint32
	uuidNamingContextObjGUID           windows.GUID
	uuidSourceDsaObjGUID               windows.GUID
	uuidSourceDsaInvocationID          windows.GUID
	uuidAsyncIntersiteTransportObjGUID windows.GUID
	usnLastObjChangeSynced             int64
	usnAttributeFilter                 int64
	ftimeLastSyncSuccess               windows.Filetime
	ftimeLastSyncAttempt               windows.Filetime
	dwLastSyncResult                   uint32
	cNumConsecutiveSyncFailures        uint32
}

// replNeighbors is the fixed part of DS_REPL_NEIGHBORSW. The array of
// cNumNeighbors entries follows at offset 8, the alignment of replNeighbor.
type replNeighbors struct {
	cNumNeighbors uint32
	dwReserved    uint32
}

// replNeighborsArrayOffset is the offset of rgNeighbor in DS_REPL_NEIGHBORSW.
const replNeighborsArrayOffset = 8

// Neighbor is the inbound replication state of one naming context from one
// source domain controller.
type Neighbor struct {
	// NamingContext is the distinguished name of the replicated naming context.
	NamingContext string
	// SourceDsaDN is the distinguished name of the NTDS Settings object of the
	// source domain controller.
	SourceDsaDN string
	// SourceDsaAddress is the transport address of the source, usually
	// <objectGUID>._msdcs.<forest root domain>.
	SourceDsaAddress string
	// ReplicaFlags holds the DS_REPL_NBR_* flags.
	ReplicaFlags uint32
	// LastSyncSuccess is zero if the naming context never replicated
	// successfully from this source.
	LastSyncSuccess time.Time
	// LastSyncAttempt is zero if replication was never attempted.
	LastSyncAttempt time.Time
	// LastSyncResult is the Win32 error code of the last attempt, 0 on success.
	LastSyncResult uint32
	// ConsecutiveSyncFailures is the number of failed attempts since the last
	// successful one.
	ConsecutiveSyncFailures uint32
}

// Handle is a binding to a domain controller, returned by Bind.
type Handle windows.Handle

// Bind binds to the directory service on the domain controller dc, a DNS or
// NetBIOS name. Close the handle with Unbind.
func Bind(dc string) (Handle, error) {
	dcName, err := windows.UTF16PtrFromString(dc)
	if err != nil {
		return 0, fmt.Errorf("DsBindW: %w", err)
	}

	var handle Handle

	r0, _, _ := syscall.SyscallN(procDsBindW.Addr(), uintptr(unsafe.Pointer(dcName)), 0, uintptr(unsafe.Pointer(&handle)))
	if r0 != 0 {
		return 0, fmt.Errorf("DsBindW: %w", windows.Errno(r0))
	}

	return handle, nil
}

// Unbind releases a handle returned by Bind.
func Unbind(handle Handle) error {
	r0, _, _ := syscall.SyscallN(procDsUnBindW.Addr(), uintptr(unsafe.Pointer(&handle)))
	if r0 != 0 {
		return fmt.Errorf("DsUnBindW: %w", windows.Errno(r0))
	}

	return nil
}

// InboundNeighbors returns the inbound replication partners of every naming
// context that the bound domain controller holds.
func InboundNeighbors(handle Handle) ([]Neighbor, error) {
	var info *replNeighbors

	r0, _, _ := syscall.SyscallN(procDsReplicaGetInfoW.Addr(),
		uintptr(handle),
		dsReplInfoNeighbors,
		0, // All naming contexts, as for repadmin /showrepl (MS-DRSR IDL_DRSGetReplInfo, DS_REPL_INFO_NEIGHBORS).
		0, // All sources.
		uintptr(unsafe.Pointer(&info)),
	)
	if r0 != 0 {
		return nil, fmt.Errorf("DsReplicaGetInfoW: %w", windows.Errno(r0))
	}

	if info == nil {
		return nil, fmt.Errorf("DsReplicaGetInfoW: %w", windows.ERROR_INVALID_DATA)
	}

	defer syscall.SyscallN(procDsReplicaFreeInfo.Addr(), dsReplInfoNeighbors, uintptr(unsafe.Pointer(info))) //nolint:errcheck

	return parseNeighbors(info)
}

func parseNeighbors(info *replNeighbors) ([]Neighbor, error) {
	// C aligns the USN fields and the array entries to 8 bytes; Go does so on
	// 386 only by chance of the field order. Refuse a layout that differs.
	if unsafe.Offsetof(replNeighbor{}.usnLastObjChangeSynced)%8 != 0 || unsafe.Sizeof(replNeighbor{})%8 != 0 {
		return nil, fmt.Errorf("DsReplicaGetInfoW: %w: DS_REPL_NEIGHBORW layout of this architecture", errors.ErrUnsupported)
	}

	if info.cNumNeighbors == 0 {
		return []Neighbor{}, nil
	}

	entries := unsafe.Slice(
		(*replNeighbor)(unsafe.Add(unsafe.Pointer(info), replNeighborsArrayOffset)),
		info.cNumNeighbors,
	)

	neighbors := make([]Neighbor, 0, len(entries))

	for i := range entries {
		entry := &entries[i]

		neighbors = append(neighbors, Neighbor{
			NamingContext:           windows.UTF16PtrToString(entry.pszNamingContext),
			SourceDsaDN:             windows.UTF16PtrToString(entry.pszSourceDsaDN),
			SourceDsaAddress:        windows.UTF16PtrToString(entry.pszSourceDsaAddress),
			ReplicaFlags:            entry.dwReplicaFlags,
			LastSyncSuccess:         filetimeToTime(entry.ftimeLastSyncSuccess),
			LastSyncAttempt:         filetimeToTime(entry.ftimeLastSyncAttempt),
			LastSyncResult:          entry.dwLastSyncResult,
			ConsecutiveSyncFailures: entry.cNumConsecutiveSyncFailures,
		})
	}

	return neighbors, nil
}

// filetimeToTime returns the zero time for a zero FILETIME, which the
// directory service reports for events that never happened.
func filetimeToTime(ft windows.Filetime) time.Time {
	if ft.HighDateTime == 0 && ft.LowDateTime == 0 {
		return time.Time{}
	}

	return time.Unix(0, ft.Nanoseconds()).UTC()
}
