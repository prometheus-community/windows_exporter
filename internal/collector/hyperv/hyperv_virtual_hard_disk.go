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
	"strings"
	"time"

	"github.com/prometheus-community/windows_exporter/internal/headers/virtdisk"
	"github.com/prometheus-community/windows_exporter/internal/mi"
	"github.com/prometheus-community/windows_exporter/internal/types"
	"github.com/prometheus/client_golang/prometheus"
)

// virtualHardDiskQueryTimeout bounds each WMI query, so a hung WMI service
// fails the sub-collector instead of blocking the scrape.
const virtualHardDiskQueryTimeout = 10 * time.Second

// collectorVirtualHardDisk Hyper-V virtual hard disk size metrics
type collectorVirtualHardDisk struct {
	virtualHardDiskVMQuery      mi.Query
	virtualHardDiskSettingQuery mi.Query

	virtualHardDiskSizeBytes     *prometheus.Desc
	virtualHardDiskFileSizeBytes *prometheus.Desc
}

// msvmVirtualMachine is the part of Msvm_ComputerSystem needed to name a VM.
// The host itself is also an Msvm_ComputerSystem; its Name is the host name,
// which never matches a disk's InstanceID.
type msvmVirtualMachine struct {
	Name        string `mi:"Name"`
	ElementName string `mi:"ElementName"`
}

// msvmStorageAllocationSettingData is the part of
// Msvm_StorageAllocationSettingData needed to find attached virtual hard disks.
type msvmStorageAllocationSettingData struct {
	InstanceID   string   `mi:"InstanceID"`
	HostResource []string `mi:"HostResource"`
}

func (c *Collector) buildVirtualHardDisk() error {
	if c.miSession == nil {
		return mi.ErrNotInitialized
	}

	if err := virtdisk.Available(); err != nil {
		return errors.Join(errors.ErrUnsupported, fmt.Errorf("virtdisk.dll: %w", err))
	}

	var err error

	c.virtualHardDiskVMQuery, err = mi.NewQuery("SELECT Name, ElementName FROM Msvm_ComputerSystem")
	if err != nil {
		return fmt.Errorf("failed to create WMI query: %w", err)
	}

	// ResourceType 31 is a logical disk: virtual hard disks, not ISO images.
	c.virtualHardDiskSettingQuery, err = mi.NewQuery("SELECT InstanceID, HostResource FROM Msvm_StorageAllocationSettingData WHERE ResourceType = 31")
	if err != nil {
		return fmt.Errorf("failed to create WMI query: %w", err)
	}

	labels := []string{"vm", "vm_id", "path", "type"}

	c.virtualHardDiskSizeBytes = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "virtual_hard_disk_size_bytes"),
		"Maximum size of the virtual hard disk, as seen by the virtual machine",
		labels,
		nil,
	)
	c.virtualHardDiskFileSizeBytes = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "virtual_hard_disk_file_size_bytes"),
		"Size of the virtual hard disk file on storage",
		labels,
		nil,
	)

	return nil
}

func (c *Collector) collectVirtualHardDisk(ch chan<- prometheus.Metric) error {
	var vms []msvmVirtualMachine
	if err := c.miSession.Query(&vms, mi.NamespaceRootVirtualizationV2, c.virtualHardDiskVMQuery, virtualHardDiskQueryTimeout); err != nil {
		return fmt.Errorf("failed to query Msvm_ComputerSystem: %w", err)
	}

	var settings []msvmStorageAllocationSettingData
	if err := c.miSession.Query(&settings, mi.NamespaceRootVirtualizationV2, c.virtualHardDiskSettingQuery, virtualHardDiskQueryTimeout); err != nil {
		return fmt.Errorf("failed to query Msvm_StorageAllocationSettingData: %w", err)
	}

	var errs []error

	for _, disk := range attachedVirtualHardDisks(vms, settings) {
		info, err := virtdisk.GetInfo(disk.Path)
		if err != nil {
			errs = append(errs, err)

			continue
		}

		diskType := virtualHardDiskType(info.ProviderSubtype)

		ch <- prometheus.MustNewConstMetric(c.virtualHardDiskSizeBytes, prometheus.GaugeValue, float64(info.VirtualSize), disk.VMName, disk.VMID, disk.Path, diskType)

		ch <- prometheus.MustNewConstMetric(c.virtualHardDiskFileSizeBytes, prometheus.GaugeValue, float64(info.PhysicalSize), disk.VMName, disk.VMID, disk.Path, diskType)
	}

	return errors.Join(errs...)
}

type attachedVirtualHardDisk struct {
	VMID   string
	VMName string
	Path   string
}

// attachedVirtualHardDisks returns the disks of the VMs' current configuration.
// The InstanceID of a setting is "Microsoft:<owner GUID>\...". The owner is the
// VM for its current configuration and the checkpoint for checkpoint settings,
// so matching the VM GUID drops checkpoint settings.
func attachedVirtualHardDisks(vms []msvmVirtualMachine, settings []msvmStorageAllocationSettingData) []attachedVirtualHardDisk {
	names := make(map[string]msvmVirtualMachine, len(vms))
	for _, vm := range vms {
		names[strings.ToUpper(vm.Name)] = vm
	}

	var disks []attachedVirtualHardDisk

	for _, setting := range settings {
		owner, ok := strings.CutPrefix(setting.InstanceID, "Microsoft:")
		if !ok {
			continue
		}

		owner, _, _ = strings.Cut(owner, `\`)

		vm, ok := names[strings.ToUpper(owner)]
		if !ok {
			continue
		}

		for _, path := range setting.HostResource {
			if path == "" {
				continue
			}

			disks = append(disks, attachedVirtualHardDisk{VMID: vm.Name, VMName: vm.ElementName, Path: path})
		}
	}

	return disks
}

func virtualHardDiskType(subtype uint32) string {
	switch subtype {
	case virtdisk.ProviderSubtypeFixed:
		return "fixed"
	case virtdisk.ProviderSubtypeDynamic:
		return "dynamic"
	case virtdisk.ProviderSubtypeDifferencing:
		return "differencing"
	default:
		return "unknown"
	}
}
