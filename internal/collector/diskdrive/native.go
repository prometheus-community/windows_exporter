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

package diskdrive

// This file rebuilds the Win32_DiskDrive properties used by this collector
// without WMI. The logic mirrors CWin32DiskDrive in cimwin32.dll
// (onecore\admin\wmi\wbem\providers\win32provider\providers\diskdrive.cpp),
// reverse engineered from build 10.0.26100.8875:
//
//   - Enumeration: SetupDiGetClassDevs(GUID_DEVINTERFACE_DISK,
//     DIGCF_PRESENT|DIGCF_DEVICEINTERFACE), one instance per interface. The
//     interface path is opened with CreateFile; ERROR_FILE_NOT_FOUND skips the
//     disk. Any other open error still yields an instance that only has
//     Status "OK".
//   - DeviceID, Name: `\\.\PHYSICALDRIVE` + _itow(DeviceNumber) from
//     IOCTL_STORAGE_GET_DEVICE_NUMBER. Unset if the IOCTL fails.
//   - Caption: the same path, replaced by the REG_SZ CM_DRP_DEVICEDESC and then
//     by the REG_SZ CM_DRP_FRIENDLYNAME device registry property.
//   - Model: CM_DRP_FRIENDLYNAME only. Vendor and product IDs from
//     IOCTL_STORAGE_QUERY_PROPERTY are not used; that IOCTL only feeds
//     FirmwareRevision and SerialNumber.
//   - Size: Cylinders * TracksPerCylinder * SectorsPerTrack * BytesPerSector
//     from IOCTL_DISK_GET_DRIVE_GEOMETRY, a cylinder-rounded value that is
//     smaller than DISK_GEOMETRY_EX.DiskSize.
//   - Partitions: IOCTL_DISK_GET_DRIVE_LAYOUT_EX entries. MBR layouts count
//     entries with RecognizedPartition set, GPT layouts count entries that are
//     not the Microsoft reserved partition, RAW layouts count nothing.
//   - Status: "OK", then derived from the CM_Get_DevNode_Status flags, then
//     overridden by IOCTL_STORAGE_PREDICT_FAILURE ("Pred Fail" or "OK").
//   - Availability: never set by the provider, so WMI returns NULL.

import (
	"bytes"
	"encoding/binary"
	"errors"
	"strconv"
	"unsafe"

	"github.com/prometheus-community/windows_exporter/internal/headers/cfgmgr32"
	"github.com/prometheus-community/windows_exporter/internal/headers/setupapi"
	"golang.org/x/sys/windows"
)

// Device I/O control codes from winioctl.h. All of them use FILE_ANY_ACCESS,
// so the device can be opened without requesting read access.
const (
	ioctlStorageGetDeviceNumber  = 0x002D1080
	ioctlStoragePredictFailure   = 0x002D1100
	ioctlDiskGetDriveGeometry    = 0x00070000
	ioctlDiskGetDriveLayoutEx    = 0x00070050
	physicalDrivePrefix          = `\\.\PHYSICALDRIVE`
	diskGeometrySize             = 24
	storageDeviceNumberSize      = 12
	storagePredictFailureSize    = 4 + 512
	driveLayoutHeaderSize        = 48
	partitionInformationExSize   = 144
	partitionTypeOffset          = 0x20
	mbrRecognizedPartitionOffset = 0x22
	// initialDriveLayoutSize is the first buffer size cimwin32 uses.
	initialDriveLayoutSize = 0x10C0
	// maxDriveLayoutSize bounds the retries for IOCTL_DISK_GET_DRIVE_LAYOUT_EX.
	maxDriveLayoutSize = 1 << 20
	// registryPropertySize is the buffer size cimwin32 passes to
	// CM_Get_DevNode_Registry_PropertyW. Longer values are treated as absent.
	registryPropertySize = 0x802
)

// PARTITION_STYLE values.
const (
	partitionStyleMBR = 0
	partitionStyleGPT = 1
)

// Device node status flags from cfg.h.
const (
	dnRootEnumerated = 0x00000001
	dnDriverLoaded   = 0x00000002
	dnEnumLoaded     = 0x00000004
	dnStarted        = 0x00000008
	dnHasProblem     = 0x00000400
	dnMoved          = 0x00001000
	dnPrivateProblem = 0x00008000
	dnWillBeRemoved  = 0x00040000
)

const (
	statusOK       = "OK"
	statusError    = "Error"
	statusDegraded = "Degraded"
	statusUnknown  = "Unknown"
	statusPredFail = "Pred Fail"
)

// partitionMsftReservedGUID is PARTITION_MSFT_RESERVED_GUID
// {E3C9E316-0B5C-4DB8-817D-F92DF00215AE} in its in-memory byte order.
//
//nolint:gochecknoglobals
var partitionMsftReservedGUID = [16]byte{
	0x16, 0xe3, 0xc9, 0xe3, 0x5c, 0x0b, 0xb8, 0x4d,
	0x81, 0x7d, 0xf9, 0x2d, 0xf0, 0x02, 0x15, 0xae,
}

// readNativeDiskDrives returns the disk drives the way Win32_DiskDrive
// reports them. Only enumeration failures are returned as errors; per-disk
// IOCTL failures leave the affected property empty, as in WMI.
func readNativeDiskDrives() ([]diskDrive, error) {
	interfaces, err := setupapi.GetDeviceInterfaces(&setupapi.GUID_DEVINTERFACE_DISK)
	if err != nil {
		return nil, err
	}

	drives := make([]diskDrive, 0, len(interfaces))

	for _, deviceInterface := range interfaces {
		drive, ok := readNativeDiskDrive(deviceInterface)
		if ok {
			drives = append(drives, drive)
		}
	}

	return drives, nil
}

func readNativeDiskDrive(deviceInterface setupapi.DeviceInterface) (diskDrive, bool) {
	// cimwin32 sets Status before it opens the device.
	drive := diskDrive{Status: statusOK}

	path, err := windows.UTF16PtrFromString(deviceInterface.Path)
	if err != nil {
		return drive, true
	}

	// cimwin32 tries GENERIC_READ first and retries with 0 access after
	// ERROR_ACCESS_DENIED. The IOCTLs below use FILE_ANY_ACCESS, so 0 access is
	// enough and doesn't need administrative rights.
	handle, err := windows.CreateFile(
		path,
		0,
		windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE,
		nil,
		windows.OPEN_EXISTING,
		0,
		0,
	)
	if err != nil {
		// WBEM_E_NOT_FOUND for a vanished device: the instance is skipped.
		// Any other failure still yields an instance with Status "OK".
		return drive, !errors.Is(err, windows.ERROR_FILE_NOT_FOUND)
	}

	defer func() {
		_ = windows.CloseHandle(handle)
	}()

	if deviceNumber, ok := storageDeviceNumber(handle); ok {
		drive.DeviceID = physicalDriveName(deviceNumber)
		drive.Name = drive.DeviceID
		drive.Caption = drive.DeviceID
	}

	if description, ok := registryStringProperty(deviceInterface.DevInst, cfgmgr32.CM_DRP_DEVICEDESC); ok {
		drive.Caption = description
	}

	if friendlyName, ok := registryStringProperty(deviceInterface.DevInst, cfgmgr32.CM_DRP_FRIENDLYNAME); ok {
		drive.Caption = friendlyName
		drive.Model = friendlyName
	}

	drive.Status = deviceStatus(handle, deviceInterface.DevInst)

	if partitions, ok := partitionCount(handle); ok {
		drive.Partitions = partitions
	}

	if size, ok := geometrySize(handle); ok {
		drive.Size = size
	}

	return drive, true
}

// physicalDriveName formats the device number with _itow semantics, which
// treat the DWORD as a signed int.
func physicalDriveName(deviceNumber uint32) string {
	return physicalDrivePrefix + strconv.Itoa(int(int32(deviceNumber)))
}

func deviceIoControl(handle windows.Handle, code uint32, out []byte) (uint32, error) {
	var returned uint32

	err := windows.DeviceIoControl(handle, code, nil, 0, &out[0], uint32(len(out)), &returned, nil)

	return returned, err
}

func storageDeviceNumber(handle windows.Handle) (uint32, bool) {
	buf := make([]byte, storageDeviceNumberSize)

	if _, err := deviceIoControl(handle, ioctlStorageGetDeviceNumber, buf); err != nil {
		return 0, false
	}

	// STORAGE_DEVICE_NUMBER: DeviceType, DeviceNumber, PartitionNumber.
	return binary.LittleEndian.Uint32(buf[4:8]), true
}

func registryStringProperty(devInst uint32, property uint32) (string, bool) {
	var (
		buf      [registryPropertySize / 2]uint16
		dataType uint32
	)

	size := uint32(registryPropertySize)

	err := cfgmgr32.CMGetDevNodeRegistryProperty(devInst, property, &dataType, unsafe.Pointer(&buf[0]), &size)
	if err != nil || dataType != windows.REG_SZ {
		return "", false
	}

	return windows.UTF16ToString(buf[:]), true
}

// deviceStatus mirrors CWin32DiskDrive's Status handling. The default "OK"
// stays when CM_Get_DevNode_Status fails or reports no flags; in that case the
// predict failure IOCTL isn't sent either.
func deviceStatus(handle windows.Handle, devInst uint32) string {
	var status, problem uint32

	if err := windows.CM_Get_DevNode_Status(&status, &problem, windows.DEVINST(devInst), 0); err != nil || status == 0 {
		return statusOK
	}

	result := statusFromDevNodeStatus(status)

	buf := make([]byte, storagePredictFailureSize)
	if _, err := deviceIoControl(handle, ioctlStoragePredictFailure, buf); err == nil {
		result = statusFromPredictFailure(binary.LittleEndian.Uint32(buf[0:4]))
	}

	return result
}

// statusFromDevNodeStatus maps CM_Get_DevNode_Status flags like
// CConfigMgrDevice::GetStatus in cimwin32. Problem codes aren't used; later
// checks win over earlier ones.
func statusFromDevNodeStatus(status uint32) string {
	result := statusUnknown

	if status&(dnRootEnumerated|dnDriverLoaded|dnEnumLoaded|dnStarted) != 0 {
		result = statusOK
	}

	if status&(dnMoved|dnWillBeRemoved) != 0 {
		result = statusDegraded
	}

	if status&(dnHasProblem|dnPrivateProblem) != 0 {
		result = statusError
	}

	return result
}

// statusFromPredictFailure maps STORAGE_PREDICT_FAILURE.PredictFailure. A
// successful IOCTL replaces the device node status, including "Error".
func statusFromPredictFailure(predictFailure uint32) string {
	if predictFailure != 0 {
		return statusPredFail
	}

	return statusOK
}

func geometrySize(handle windows.Handle) (uint64, bool) {
	buf := make([]byte, diskGeometrySize)

	returned, err := deviceIoControl(handle, ioctlDiskGetDriveGeometry, buf)
	if err != nil {
		return 0, false
	}

	return sizeFromGeometry(buf[:min(int(returned), len(buf))])
}

// sizeFromGeometry multiplies the DISK_GEOMETRY fields with 64-bit wrap-around
// arithmetic, like cimwin32.
func sizeFromGeometry(geometry []byte) (uint64, bool) {
	if len(geometry) < diskGeometrySize {
		return 0, false
	}

	cylinders := binary.LittleEndian.Uint64(geometry[0:8])
	tracksPerCylinder := uint64(binary.LittleEndian.Uint32(geometry[12:16]))
	sectorsPerTrack := uint64(binary.LittleEndian.Uint32(geometry[16:20]))
	bytesPerSector := uint64(binary.LittleEndian.Uint32(geometry[20:24]))

	return cylinders * tracksPerCylinder * sectorsPerTrack * bytesPerSector, true
}

func partitionCount(handle windows.Handle) (uint32, bool) {
	for size := initialDriveLayoutSize; size <= maxDriveLayoutSize; size *= 2 {
		buf := make([]byte, size)

		returned, err := deviceIoControl(handle, ioctlDiskGetDriveLayoutEx, buf)
		if err == nil {
			return countPartitions(buf[:min(int(returned), len(buf))])
		}

		if !errors.Is(err, windows.ERROR_INSUFFICIENT_BUFFER) && !errors.Is(err, windows.ERROR_MORE_DATA) {
			return 0, false
		}
	}

	return 0, false
}

// countPartitions counts DRIVE_LAYOUT_INFORMATION_EX entries like cimwin32.
// The check depends on the layout's partition style. For MBR layouts, an
// entry with RecognizedPartition set is also compared against the reserved
// partition GUID at the same offset, which can't match a valid MBR entry.
func countPartitions(layout []byte) (uint32, bool) {
	if len(layout) < driveLayoutHeaderSize {
		return 0, false
	}

	style := binary.LittleEndian.Uint32(layout[0:4])
	count := binary.LittleEndian.Uint32(layout[4:8])

	var partitions uint32

	for i := range uint64(count) {
		offset := driveLayoutHeaderSize + i*partitionInformationExSize
		if offset+partitionInformationExSize > uint64(len(layout)) {
			return 0, false
		}

		entry := layout[offset : offset+partitionInformationExSize]

		switch style {
		case partitionStyleMBR:
			if entry[mbrRecognizedPartitionOffset] == 0 {
				continue
			}
		case partitionStyleGPT:
		default:
			continue
		}

		if !bytes.Equal(entry[partitionTypeOffset:partitionTypeOffset+16], partitionMsftReservedGUID[:]) {
			partitions++
		}
	}

	return partitions, true
}
