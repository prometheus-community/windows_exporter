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

package clusapi

import (
	"encoding/binary"
	"errors"
	"fmt"
	"time"

	"golang.org/x/sys/windows"
)

const (
	// CLUSCTL_RESOURCE_STORAGE_GET_DISK_INFO_EX, available since Windows
	// Server 2008. The EX2 variant needs Windows Server 2016 and adds nothing
	// that the collector publishes.
	// https://learn.microsoft.com/en-us/previous-versions/windows/desktop/mscs/clusctl-resource-storage-get-disk-info-ex
	resourceStorageGetDiskInfoEx = 0x010001f1
)

// Partition is one CLUS_PARTITION_INFO_EX entry of a storage class resource.
type Partition struct {
	Resource    string
	DeviceName  string
	VolumeLabel string
	VolumeGUID  windows.GUID
	TotalBytes  uint64
	FreeBytes   uint64
}

// DiskPartitions reads the partitions of every storage class resource, the
// source of the MSCluster_DiskPartition WMI class. Resources that do not
// support the disk information control code are skipped. Successful
// partitions are retained alongside joined failures.
func (c *Cluster) DiskPartitions(deadline time.Time) (_ []Partition, resultErr error) {
	for _, proc := range []*windows.LazyProc{openCluster, openEnum, nextEnum, closeEnum, openResource, closeResource, resourceControl} {
		if err := proc.Find(); err != nil {
			return nil, fmt.Errorf("load ClusAPI: %w: %w", errors.ErrUnsupported, err)
		}
	}

	if err := c.begin(deadline); err != nil {
		return nil, err
	}
	defer func() { resultErr = errors.Join(resultErr, c.end()) }()

	if err := c.open(); err != nil {
		return nil, err
	}

	var partitions []Partition

	err := c.enumerate(enumResource, "resource", deadline, func(name objectName) error {
		found, err := c.readResourcePartitions(name, deadline)
		partitions = append(partitions, found...)

		return err
	})

	return partitions, err
}

func (c *Cluster) readResourcePartitions(name objectName, deadline time.Time) (_ []Partition, resultErr error) {
	if err := checkDeadline(deadline); err != nil {
		return nil, err
	}

	handle, err := openObject(openResource, c.handle, name)
	if errors.Is(err, windows.ERROR_RESOURCE_NOT_FOUND) {
		return nil, errObjectDeleted
	}

	if err != nil {
		return nil, fmt.Errorf("OpenClusterResourceEx: %w", err)
	}
	defer func() {
		if err := closeObject(closeResource, handle); err != nil {
			resultErr = errors.Join(resultErr, fmt.Errorf("CloseClusterResource: %w", err))
		}
	}()

	data, err := resourceBuffer(handle, resourceGetClassInfo, 8, deadline)
	if err != nil {
		return nil, fmt.Errorf("class information: %w", err)
	}

	if len(data) != 8 {
		return nil, errors.New("invalid resource class information size")
	}

	if binary.LittleEndian.Uint32(data) != resourceClassStorage {
		return nil, nil
	}

	data, err = resourceBuffer(handle, resourceStorageGetDiskInfoEx, propertyListBufferSize, deadline)
	if errors.Is(err, windows.ERROR_INVALID_FUNCTION) || errors.Is(err, windows.ERROR_NOT_SUPPORTED) {
		// Storage class resources such as storage pools have no disk layout.
		return nil, nil
	}

	if err != nil {
		return nil, fmt.Errorf("CLUSCTL_RESOURCE_STORAGE_GET_DISK_INFO_EX: %w", err)
	}

	partitions, err := ParseDiskPartitions(data)
	for index := range partitions {
		partitions[index].Resource = name.name
	}

	return partitions, err
}

// CLUSPROP_SYNTAX_PARTITION_INFO_EX: CLUSPROP_TYPE_PARTITION_INFO_EX (13) with
// CLUSPROP_FORMAT_BINARY (1).
const syntaxPartitionInfoEx = 13<<16 | 1

// Byte offsets in CLUS_PARTITION_INFO_EX. MAX_PATH is 260 WCHARs; the
// structure is 1160 bytes with or without natural alignment of the
// ULARGE_INTEGER members.
// https://learn.microsoft.com/en-us/previous-versions/windows/desktop/api/clusapi/ns-clusapi-clus_partition_info_ex
const (
	partitionDeviceNameOffset  = 4
	partitionVolumeLabelOffset = partitionDeviceNameOffset + 260*2
	partitionTotalSizeOffset   = partitionVolumeLabelOffset + 260*2 + 3*4 + 32*2
	partitionFreeSizeOffset    = partitionTotalSizeOffset + 8
	partitionVolumeGUIDOffset  = partitionFreeSizeOffset + 8 + 2*4
	partitionInfoExSize        = partitionVolumeGUIDOffset + 16
)

// ParseDiskPartitions decodes the CLUSPROP_PARTITION_INFO_EX entries of a
// CLUSCTL_RESOURCE_STORAGE_GET_DISK_INFO_EX value list. Other entries, such as
// the disk signature or size, are skipped.
func ParseDiskPartitions(data []byte) ([]Partition, error) {
	values, err := ParseValueList(data)
	if err != nil {
		return nil, err
	}

	var partitions []Partition

	for _, value := range values {
		if value.Syntax != syntaxPartitionInfoEx {
			continue
		}

		partition, err := parsePartitionInfoEx(value.Data)
		if err != nil {
			return partitions, err
		}

		partitions = append(partitions, partition)
	}

	return partitions, nil
}

func parsePartitionInfoEx(data []byte) (Partition, error) {
	if len(data) < partitionInfoExSize {
		return Partition{}, fmt.Errorf("partition information has %d bytes, want at least %d", len(data), partitionInfoExSize)
	}

	// Fixed-size WCHAR arrays decode up to their first NUL, like WMI strings.
	deviceName := decodeString(data[partitionDeviceNameOffset:partitionVolumeLabelOffset])
	volumeLabel := decodeString(data[partitionVolumeLabelOffset : partitionVolumeLabelOffset+260*2])

	guid := data[partitionVolumeGUIDOffset:partitionInfoExSize]

	return Partition{
		DeviceName:  deviceName,
		VolumeLabel: volumeLabel,
		VolumeGUID: windows.GUID{
			Data1: binary.LittleEndian.Uint32(guid),
			Data2: binary.LittleEndian.Uint16(guid[4:]),
			Data3: binary.LittleEndian.Uint16(guid[6:]),
			Data4: [8]byte(guid[8:16]),
		},
		TotalBytes: binary.LittleEndian.Uint64(data[partitionTotalSizeOffset:]),
		FreeBytes:  binary.LittleEndian.Uint64(data[partitionFreeSizeOffset:]),
	}, nil
}

// Value is one entry of a CLUSPROP value list.
type Value struct {
	Syntax uint32
	Data   []byte
}

// ParseValueList decodes a value list: CLUSPROP_VALUE headers (syntax and
// byte length) with DWORD-aligned payloads, terminated by
// CLUSPROP_SYNTAX_ENDMARK. Zero padding after the endmark is accepted.
// https://learn.microsoft.com/en-us/previous-versions/windows/desktop/mscs/value-lists
func ParseValueList(data []byte) ([]Value, error) {
	var values []Value

	for {
		if len(data) < 4 {
			return nil, errors.New("value list is missing its endmark")
		}

		syntax := binary.LittleEndian.Uint32(data)
		if syntax == 0 {
			for _, b := range data[4:] {
				if b != 0 {
					return nil, fmt.Errorf("%d bytes of trailing value list data", len(data)-4)
				}
			}

			return values, nil
		}

		payload, rest, err := propertyPayload(data)
		if err != nil {
			return nil, fmt.Errorf("value syntax %#x: %w", syntax, err)
		}

		values = append(values, Value{Syntax: syntax, Data: payload})
		data = rest
	}
}
