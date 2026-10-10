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
	"testing"
	"time"
	"unicode/utf16"

	"golang.org/x/sys/windows"
)

func putFixedString(data []byte, value string) {
	for index, unit := range utf16.Encode([]rune(value)) {
		binary.LittleEndian.PutUint16(data[index*2:], unit)
	}
}

func testPartitionInfoEx() []byte {
	data := make([]byte, partitionInfoExSize)
	putFixedString(data[partitionDeviceNameOffset:], `C:\ClusterStorage\Volume1`)
	putFixedString(data[partitionVolumeLabelOffset:], "CSV01  ")
	binary.LittleEndian.PutUint64(data[partitionTotalSizeOffset:], 2<<40)
	binary.LittleEndian.PutUint64(data[partitionFreeSizeOffset:], 19008585728+12345)

	guid := data[partitionVolumeGUIDOffset:]
	binary.LittleEndian.PutUint32(guid, 0xd3dbada3)
	binary.LittleEndian.PutUint16(guid[4:], 0x448f)
	binary.LittleEndian.PutUint16(guid[6:], 0x4c4c)
	copy(guid[8:], []byte{0xbe, 0x4d, 0x12, 0x2b, 0x7f, 0x29, 0x60, 0x07})

	return data
}

func appendValue(list []byte, syntax uint32, payload []byte) []byte {
	list = binary.LittleEndian.AppendUint32(list, syntax)
	list = binary.LittleEndian.AppendUint32(list, uint32(len(payload)))

	list = append(list, payload...)
	for len(list)%4 != 0 {
		list = append(list, 0)
	}

	return list
}

// testDiskInfo builds the documented GET_DISK_INFO_EX value list: a disk
// signature, a disk size, a partition and the endmark.
func testDiskInfo() []byte {
	list := appendValue(nil, 5<<16|2, []byte{1, 2, 3, 4})
	list = appendValue(list, 12<<16|6, make([]byte, 8))
	list = appendValue(list, syntaxPartitionInfoEx, testPartitionInfoEx())

	return binary.LittleEndian.AppendUint32(list, 0)
}

func TestParseDiskPartitions(t *testing.T) {
	partitions, err := ParseDiskPartitions(testDiskInfo())
	if err != nil {
		t.Fatal(err)
	}

	if len(partitions) != 1 {
		t.Fatalf("partitions = %d", len(partitions))
	}

	partition := partitions[0]
	want := windows.GUID{Data1: 0xd3dbada3, Data2: 0x448f, Data3: 0x4c4c, Data4: [8]byte{0xbe, 0x4d, 0x12, 0x2b, 0x7f, 0x29, 0x60, 0x07}}

	if partition.DeviceName != `C:\ClusterStorage\Volume1` || partition.VolumeLabel != "CSV01  " || partition.VolumeGUID != want ||
		partition.TotalBytes != 2<<40 || partition.FreeBytes != 19008585728+12345 {
		t.Fatalf("partition = %+v", partition)
	}

	data := testDiskInfo()
	for i := range data {
		if _, err := ParseDiskPartitions(data[:i]); err == nil {
			t.Errorf("accepted truncation at %d", i)
		}
	}
}

func TestParseDiskPartitionsInvalid(t *testing.T) {
	short := appendValue(nil, syntaxPartitionInfoEx, make([]byte, partitionInfoExSize-1))
	if _, err := ParseDiskPartitions(binary.LittleEndian.AppendUint32(short, 0)); err == nil {
		t.Fatal("accepted short partition information")
	}

	unterminated := testPartitionInfoEx()
	for index := partitionVolumeLabelOffset; index < partitionVolumeLabelOffset+260*2; index += 2 {
		binary.LittleEndian.PutUint16(unterminated[index:], 'A')
	}

	list := binary.LittleEndian.AppendUint32(appendValue(nil, syntaxPartitionInfoEx, unterminated), 0)
	if _, err := ParseDiskPartitions(list); err == nil {
		t.Fatal("accepted unterminated volume label")
	}

	surrogate := testPartitionInfoEx()
	binary.LittleEndian.PutUint16(surrogate[partitionDeviceNameOffset:], 0xdc00)

	list = binary.LittleEndian.AppendUint32(appendValue(nil, syntaxPartitionInfoEx, surrogate), 0)
	if _, err := ParseDiskPartitions(list); err == nil {
		t.Fatal("accepted unpaired surrogate")
	}

	oversized := testDiskInfo()
	binary.LittleEndian.PutUint32(oversized[4:], ^uint32(0))

	if _, err := ParseDiskPartitions(oversized); err == nil {
		t.Fatal("accepted oversized value length")
	}

	if _, err := ParseDiskPartitions(append(testDiskInfo(), 1, 0, 0, 0)); err == nil {
		t.Fatal("accepted trailing data")
	}

	if partitions, err := ParseDiskPartitions(append(testDiskInfo(), 0, 0, 0, 0)); err != nil || len(partitions) != 1 {
		t.Fatalf("zero padding: %v", err)
	}
}

func TestParseValueListEmpty(t *testing.T) {
	values, err := ParseValueList([]byte{0, 0, 0, 0})
	if err != nil || len(values) != 0 {
		t.Fatalf("values = %v, err = %v", values, err)
	}

	if _, err := ParseValueList(nil); err == nil {
		t.Fatal("accepted missing endmark")
	}
}

func TestClusterClosedDiskPartitions(t *testing.T) {
	c := &Cluster{}
	if err := c.Close(); err != nil {
		t.Fatal(err)
	}

	if _, err := c.DiskPartitions(time.Time{}); err == nil {
		t.Fatal("read closed cluster")
	}
}

func FuzzParseDiskPartitions(f *testing.F) {
	f.Add(testDiskInfo())
	f.Add([]byte{0, 0, 0, 0})
	f.Fuzz(func(t *testing.T, data []byte) { _, _ = ParseDiskPartitions(data) })
}
