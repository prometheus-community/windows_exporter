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
	"github.com/prometheus-community/windows_exporter/internal/labelcache"
	"github.com/prometheus-community/windows_exporter/internal/mi"
	"github.com/prometheus-community/windows_exporter/internal/types"
	"github.com/prometheus-community/windows_exporter/internal/utils/recovery"
	"github.com/prometheus/client_golang/prometheus"
)

// virtualHardDiskQueryTimeout bounds each WMI query of a refresh. It also
// bounds Close, which waits for a running query.
const virtualHardDiskQueryTimeout = 10 * time.Second

// virtualHardDiskFirstRefreshTimeout bounds how long Build waits for the first
// refresh, so the first scrape after a start already has data.
const virtualHardDiskFirstRefreshTimeout = 10 * time.Second

// virtualHardDiskWorkers bounds how many disks are opened at once. Opening a
// disk on a CSV takes a few milliseconds, so hosts with many disks read them
// in parallel.
const virtualHardDiskWorkers = 8

// virtualHardDiskResourceSubType marks virtual hard disks. ISO and floppy
// images share ResourceType 31 and differ only in ResourceSubType.
const virtualHardDiskResourceSubType = "Microsoft:Hyper-V:Virtual Hard Disk"

// collectorVirtualHardDisk Hyper-V virtual hard disk size metrics
type collectorVirtualHardDisk struct {
	virtualHardDiskVMQuery      mi.Query
	virtualHardDiskSettingQuery mi.Query

	virtualHardDiskRefresher *virtualHardDiskRefresher

	virtualHardDiskSizeBytes                   *prometheus.Desc
	virtualHardDiskFileSizeBytes               *prometheus.Desc
	virtualHardDiskReadErrors                  *prometheus.Desc
	virtualHardDiskLastRefreshTimestampSeconds *prometheus.Desc
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
	InstanceID      string   `mi:"InstanceID"`
	ResourceSubType string   `mi:"ResourceSubType"`
	HostResource    []string `mi:"HostResource"`
}

func (c *Collector) buildVirtualHardDisk() error {
	if c.config.VirtualHardDisk.RefreshInterval <= 0 {
		return fmt.Errorf("collector.hyperv.virtual-hard-disk.refresh-interval must be positive, got %s", c.config.VirtualHardDisk.RefreshInterval)
	}

	if c.miSession == nil {
		return mi.ErrNotInitialized
	}

	if err := virtdisk.Available(); err != nil {
		return fmt.Errorf("%w: virtdisk.dll: %w", errors.ErrUnsupported, err)
	}

	var err error

	c.virtualHardDiskVMQuery, err = mi.NewQuery("SELECT Name, ElementName FROM Msvm_ComputerSystem")
	if err != nil {
		return fmt.Errorf("failed to create WMI query: %w", err)
	}

	// ResourceType 31 is a logical disk: virtual hard disks, ISO and floppy
	// images. attachedVirtualHardDisks keeps the virtual hard disks.
	c.virtualHardDiskSettingQuery, err = mi.NewQuery("SELECT InstanceID, ResourceSubType, HostResource FROM Msvm_StorageAllocationSettingData WHERE ResourceType = 31")
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
	c.virtualHardDiskReadErrors = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "virtual_hard_disk_read_errors"),
		"Number of attached virtual hard disks that could not be read in the last refresh",
		nil,
		nil,
	)
	c.virtualHardDiskLastRefreshTimestampSeconds = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "virtual_hard_disk_last_refresh_timestamp_seconds"),
		"Time of the last successful refresh of the virtual hard disk metrics",
		nil,
		nil,
	)

	c.virtualHardDiskRefresher = newVirtualHardDiskRefresher(c.logger, c.config.VirtualHardDisk.RefreshInterval, c.listVirtualHardDisks, virtdisk.GetInfo)
	c.virtualHardDiskRefresher.start(virtualHardDiskFirstRefreshTimeout)

	return nil
}

func (c *Collector) closeVirtualHardDisk() {
	if c.virtualHardDiskRefresher != nil {
		c.virtualHardDiskRefresher.stop()
		c.virtualHardDiskRefresher = nil
	}
}

func (c *Collector) listVirtualHardDisks() ([]attachedVirtualHardDisk, error) {
	var vms []msvmVirtualMachine
	if err := c.miSession.Query(&vms, mi.NamespaceRootVirtualizationV2, c.virtualHardDiskVMQuery, virtualHardDiskQueryTimeout); err != nil {
		return nil, fmt.Errorf("failed to query Msvm_ComputerSystem: %w", err)
	}

	var settings []msvmStorageAllocationSettingData
	if err := c.miSession.Query(&settings, mi.NamespaceRootVirtualizationV2, c.virtualHardDiskSettingQuery, virtualHardDiskQueryTimeout); err != nil {
		return nil, fmt.Errorf("failed to query Msvm_StorageAllocationSettingData: %w", err)
	}

	return attachedVirtualHardDisks(vms, settings), nil
}

// collectVirtualHardDisk publishes the result of the last background refresh.
// Disks that can't be read are logged and counted, only a failed disk listing
// is returned as an error.
func (c *Collector) collectVirtualHardDisk(ch chan<- prometheus.Metric, scrape *labelcache.Scrape) error {
	snapshot := c.virtualHardDiskRefresher.get()

	for _, result := range snapshot.results {
		disk := result.disk
		diskType := virtualHardDiskType(result.info.ProviderSubtype)

		ch <- scrape.MustNewConstMetric(c.virtualHardDiskSizeBytes, prometheus.GaugeValue, float64(result.info.VirtualSize), disk.VMName, disk.VMID, disk.Path, diskType)

		ch <- scrape.MustNewConstMetric(c.virtualHardDiskFileSizeBytes, prometheus.GaugeValue, float64(result.info.PhysicalSize), disk.VMName, disk.VMID, disk.Path, diskType)
	}

	if !snapshot.refreshed.IsZero() {
		ch <- scrape.MustNewConstMetric(c.virtualHardDiskReadErrors, prometheus.GaugeValue, float64(snapshot.readErrors))

		ch <- scrape.MustNewConstMetric(c.virtualHardDiskLastRefreshTimestampSeconds, prometheus.GaugeValue, float64(snapshot.refreshed.UnixNano())/1e9)
	}

	if snapshot.err != nil {
		return fmt.Errorf("last virtual hard disk refresh failed: %w", snapshot.err)
	}

	return nil
}

type virtualHardDiskResult struct {
	disk attachedVirtualHardDisk
	info virtdisk.Info
}

// readVirtualHardDisks reads the disks with at most workers calls of getInfo
// at once. A disk that fails is left out and its error, or a recovered panic,
// is returned joined with the others; the order of the disks is kept.
func readVirtualHardDisks(disks []attachedVirtualHardDisk, workers int, getInfo func(path string) (virtdisk.Info, error)) ([]virtualHardDiskResult, error) {
	infos := make([]virtdisk.Info, len(disks))
	read := make([]bool, len(disks))
	slots := make(chan struct{}, workers)

	var group recovery.Group

	for i, disk := range disks {
		slots <- struct{}{}

		group.Go(func() error {
			defer func() { <-slots }()

			info, err := getInfo(disk.Path)
			if err != nil {
				return err
			}

			infos[i], read[i] = info, true

			return nil
		})
	}

	err := group.Wait()

	results := make([]virtualHardDiskResult, 0, len(disks))

	for i, disk := range disks {
		if read[i] {
			results = append(results, virtualHardDiskResult{disk: disk, info: infos[i]})
		}
	}

	return results, err
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
		if setting.ResourceSubType != virtualHardDiskResourceSubType {
			continue
		}

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
