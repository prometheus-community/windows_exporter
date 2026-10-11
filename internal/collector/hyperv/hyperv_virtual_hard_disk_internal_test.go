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
	"testing"

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
			settings: []msvmStorageAllocationSettingData{{InstanceID: `Microsoft:` + vmID + `\83F8638B-8DCA-4152-9EDA-2CA8B33039B4\0\0\L`, HostResource: []string{`C:\ClusterStorage\Volume1\VM01\disk.vhdx`}}},
			want:     []attachedVirtualHardDisk{{VMID: vmID, VMName: "VM01", Path: `C:\ClusterStorage\Volume1\VM01\disk.vhdx`}},
		},
		"checkpoint settings are dropped": {
			vms: vms,
			settings: []msvmStorageAllocationSettingData{
				{InstanceID: `Microsoft:` + snapshotID + `\83F8638B-8DCA-4152-9EDA-2CA8B33039B4\0\0\L`, HostResource: []string{`C:\VM01\disk.vhdx`}},
				{InstanceID: `Microsoft:` + vmID + `\83F8638B-8DCA-4152-9EDA-2CA8B33039B4\0\0\L`, HostResource: []string{`C:\VM01\disk_ABC.avhdx`}},
			},
			want: []attachedVirtualHardDisk{{VMID: vmID, VMName: "VM01", Path: `C:\VM01\disk_ABC.avhdx`}},
		},
		"guid case differs": {
			vms:      vms,
			settings: []msvmStorageAllocationSettingData{{InstanceID: `Microsoft:8e3a2c14-5b7d-4f11-9c2e-0a1b2c3d4e5f\X\0\0\L`, HostResource: []string{`C:\VM01\disk.vhdx`}}},
			want:     []attachedVirtualHardDisk{{VMID: vmID, VMName: "VM01", Path: `C:\VM01\disk.vhdx`}},
		},
		"empty host resource": {
			vms: vms,
			settings: []msvmStorageAllocationSettingData{
				{InstanceID: `Microsoft:` + vmID + `\X\0\0\L`, HostResource: nil},
				{InstanceID: `Microsoft:` + vmID + `\Y\0\0\L`, HostResource: []string{""}},
			},
			want: nil,
		},
		"unknown instance id format": {
			vms:      vms,
			settings: []msvmStorageAllocationSettingData{{InstanceID: `garbage`, HostResource: []string{`C:\x.vhdx`}}},
			want:     nil,
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
