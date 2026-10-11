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

package hyperv

import (
	"errors"
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/prometheus-community/windows_exporter/internal/headers/virtdisk"
	"github.com/stretchr/testify/require"
)

func TestAttachedVirtualHardDisks(t *testing.T) {
	t.Parallel()

	const (
		vmID       = "8E3A2C14-5B7D-4F11-9C2E-0A1B2C3D4E5F"
		snapshotID = "11111111-2222-3333-4444-555555555555"
	)

	vms := []msvmVirtualMachine{
		{Name: "HOSTNAME", ElementName: "HOSTNAME"}, // the host itself
		{Name: vmID, ElementName: "VM01"},
	}

	for name, tc := range map[string]struct {
		vms      []msvmVirtualMachine
		settings []msvmStorageAllocationSettingData
		want     []attachedVirtualHardDisk
	}{
		"current disk": {
			vms:      vms,
			settings: []msvmStorageAllocationSettingData{{ResourceSubType: virtualHardDiskResourceSubType, InstanceID: `Microsoft:` + vmID + `\83F8638B-8DCA-4152-9EDA-2CA8B33039B4\0\0\L`, HostResource: []string{`C:\ClusterStorage\Volume1\VM01\disk.vhdx`}}},
			want:     []attachedVirtualHardDisk{{VMID: vmID, VMName: "VM01", Path: `C:\ClusterStorage\Volume1\VM01\disk.vhdx`}},
		},
		"checkpoint settings are dropped": {
			vms: vms,
			settings: []msvmStorageAllocationSettingData{
				{ResourceSubType: virtualHardDiskResourceSubType, InstanceID: `Microsoft:` + snapshotID + `\83F8638B-8DCA-4152-9EDA-2CA8B33039B4\0\0\L`, HostResource: []string{`C:\VM01\disk.vhdx`}},
				{ResourceSubType: virtualHardDiskResourceSubType, InstanceID: `Microsoft:` + vmID + `\83F8638B-8DCA-4152-9EDA-2CA8B33039B4\0\0\L`, HostResource: []string{`C:\VM01\disk_ABC.avhdx`}},
			},
			want: []attachedVirtualHardDisk{{VMID: vmID, VMName: "VM01", Path: `C:\VM01\disk_ABC.avhdx`}},
		},
		"guid case differs": {
			vms:      vms,
			settings: []msvmStorageAllocationSettingData{{ResourceSubType: virtualHardDiskResourceSubType, InstanceID: `Microsoft:8e3a2c14-5b7d-4f11-9c2e-0a1b2c3d4e5f\X\0\0\L`, HostResource: []string{`C:\VM01\disk.vhdx`}}},
			want:     []attachedVirtualHardDisk{{VMID: vmID, VMName: "VM01", Path: `C:\VM01\disk.vhdx`}},
		},
		"empty host resource": {
			vms: vms,
			settings: []msvmStorageAllocationSettingData{
				{ResourceSubType: virtualHardDiskResourceSubType, InstanceID: `Microsoft:` + vmID + `\X\0\0\L`, HostResource: nil},
				{ResourceSubType: virtualHardDiskResourceSubType, InstanceID: `Microsoft:` + vmID + `\Y\0\0\L`, HostResource: []string{""}},
			},
			want: nil,
		},
		"unknown instance id format": {
			vms:      vms,
			settings: []msvmStorageAllocationSettingData{{ResourceSubType: virtualHardDiskResourceSubType, InstanceID: `garbage`, HostResource: []string{`C:\x.vhdx`}}},
			want:     nil,
		},
		"iso and floppy images are dropped": {
			vms: vms,
			settings: []msvmStorageAllocationSettingData{
				{ResourceSubType: "Microsoft:Hyper-V:Virtual CD/DVD Disk", InstanceID: `Microsoft:` + vmID + `\X\0\0\D`, HostResource: []string{`C:\ISO\install.iso`}},
				{ResourceSubType: "Microsoft:Hyper-V:Virtual Floppy Disk", InstanceID: `Microsoft:` + vmID + `\Y\0\0\F`, HostResource: []string{`C:\VM01\boot.vfd`}},
			},
			want: nil,
		},
		"no vms": {
			vms:      nil,
			settings: nil,
			want:     nil,
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			require.Equal(t, tc.want, attachedVirtualHardDisks(tc.vms, tc.settings))
		})
	}
}

func TestVirtualHardDiskType(t *testing.T) {
	t.Parallel()

	require.Equal(t, "fixed", virtualHardDiskType(virtdisk.ProviderSubtypeFixed))
	require.Equal(t, "dynamic", virtualHardDiskType(virtdisk.ProviderSubtypeDynamic))
	require.Equal(t, "differencing", virtualHardDiskType(virtdisk.ProviderSubtypeDifferencing))
	require.Equal(t, "unknown", virtualHardDiskType(0))
	require.Equal(t, "unknown", virtualHardDiskType(99))
}

func TestReadVirtualHardDisks(t *testing.T) {
	t.Parallel()

	disks := make([]attachedVirtualHardDisk, 20)
	for i := range disks {
		disks[i] = attachedVirtualHardDisk{VMID: "id", VMName: "vm", Path: fmt.Sprintf(`C:\disk%d.vhdx`, i)}
	}

	errBroken := errors.New("broken")

	var running, maxRunning atomic.Int32

	getInfo := func(path string) (virtdisk.Info, error) {
		current := running.Add(1)
		defer running.Add(-1)

		for {
			highest := maxRunning.Load()
			if current <= highest || maxRunning.CompareAndSwap(highest, current) {
				break
			}
		}

		time.Sleep(5 * time.Millisecond)

		if path == `C:\disk7.vhdx` {
			return virtdisk.Info{}, fmt.Errorf("%s: %w", path, errBroken)
		}

		if path == `C:\disk9.vhdx` {
			panic("boom")
		}

		return virtdisk.Info{VirtualSize: uint64(len(path)), ProviderSubtype: virtdisk.ProviderSubtypeDynamic}, nil
	}

	results, err := readVirtualHardDisks(disks, 4, getInfo)

	// One disk fails, one panics: both are reported, the others are read.
	require.ErrorIs(t, err, errBroken)
	require.ErrorContains(t, err, `C:\disk7.vhdx`)
	require.ErrorContains(t, err, "boom")
	require.Len(t, results, len(disks)-2)

	for _, result := range results {
		require.NotEqual(t, `C:\disk7.vhdx`, result.disk.Path)
		require.NotEqual(t, `C:\disk9.vhdx`, result.disk.Path)
		require.Equal(t, uint64(len(result.disk.Path)), result.info.VirtualSize)
	}

	require.LessOrEqual(t, maxRunning.Load(), int32(4), "more disks were opened at once than allowed")
	require.Greater(t, maxRunning.Load(), int32(1), "disks were not opened in parallel")
}

func TestReadVirtualHardDisksEmpty(t *testing.T) {
	t.Parallel()

	results, err := readVirtualHardDisks(nil, 4, func(string) (virtdisk.Info, error) {
		t.Fatal("no disk to read")

		return virtdisk.Info{}, nil
	})
	require.NoError(t, err)
	require.Empty(t, results)
}
