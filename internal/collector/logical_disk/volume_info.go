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

package logical_disk

import (
	"encoding/binary"
	"errors"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/prometheus-community/windows_exporter/internal/utils"
	"golang.org/x/sys/windows"
)

const (
	// ioctlVolumeGetVolumeDiskExtents is IOCTL_VOLUME_GET_VOLUME_DISK_EXTENTS.
	ioctlVolumeGetVolumeDiskExtents = 0x00560000

	// diskExtentsHeaderSize is the offset of the first DISK_EXTENT in VOLUME_DISK_EXTENTS:
	// NumberOfDiskExtents (DWORD) padded to the 8-byte alignment of DISK_EXTENT.
	diskExtentsHeaderSize = 8
	// diskExtentSize is the size of the DISK_EXTENT structure in bytes.
	diskExtentSize = 24

	// staticVolumeInfoMaxAge bounds how long the disk IDs and the drive type of a volume are reused.
	// Disk numbers can change while a volume stays mounted, e.g. after an iSCSI or USB reconnect.
	staticVolumeInfoMaxAge = 10 * time.Minute
)

type volumeInfo struct {
	diskIDs      string
	filesystem   string
	serialNumber string
	label        string
	volumeType   string
	readonly     float64
	// hasReadonly is false if the volume information, including the read-only flag, is unknown.
	hasReadonly bool

	freeBytes float64
	sizeBytes float64
	// hasSpace is false if freeBytes and sizeBytes are unknown.
	hasSpace bool
}

// volumeState is what volumeInfoReader keeps per volume between scrapes.
type volumeState struct {
	// generation is the scrape that last read the volume. Volumes not read in a scrape are dropped.
	generation uint64

	// devicePath, diskIDs and driveType are the cached disk extents and drive type, read at readAt.
	// A different device path (another volume mounted under the same name) invalidates them.
	hasStatic  bool
	devicePath string
	diskIDs    string
	driveType  uint32
	readAt     time.Time

	// lastErr is the last error logged as warning, to log repeats at debug level.
	lastErr string
}

// volumeInfoReader reads the information of volumes. It caches the disk IDs and the drive type,
// which rarely change, and reuses its buffers. It must not be used concurrently; the Collect calls
// of a collector never overlap.
type volumeInfoReader struct {
	volumes    map[string]*volumeState
	generation uint64
	now        func() time.Time

	extents []byte
	label   []uint16
	fsName  []uint16
}

func newVolumeInfoReader() *volumeInfoReader {
	return &volumeInfoReader{
		volumes: make(map[string]*volumeState),
		now:     time.Now,
		// Room for 4 extents. Volumes spanning more disks grow the buffer once.
		extents: make([]byte, diskExtentsHeaderSize+4*diskExtentSize),
		label:   make([]uint16, windows.MAX_PATH+1),
		fsName:  make([]uint16, windows.MAX_PATH+1),
	}
}

// beginScrape starts a scrape. Every volume read before the next endScrape is kept.
func (r *volumeInfoReader) beginScrape() {
	r.generation++
}

// endScrape drops the state of all volumes that were not read since beginScrape.
func (r *volumeInfoReader) endScrape() {
	for name, state := range r.volumes {
		if state.generation != r.generation {
			delete(r.volumes, name)
		}
	}
}

func (r *volumeInfoReader) state(name string) *volumeState {
	state, ok := r.volumes[name]
	if !ok {
		state = &volumeState{}
		r.volumes[name] = state
	}

	state.generation = r.generation

	return state
}

// shouldWarn reports whether err should be logged as warning for the volume name. It returns
// false if the same error was already logged since the volume last read without error.
func (r *volumeInfoReader) shouldWarn(name string, err error) bool {
	state := r.state(name)

	if err == nil {
		state.lastErr = ""

		return false
	}

	msg := err.Error()
	if state.lastErr == msg {
		return false
	}

	state.lastErr = msg

	return true
}

// read returns the information of the volume name, a LogicalDisk instance name.
// mountedVolumes maps mount points to volume GUID paths, see getAllMountedVolumes.
// With withSpace, it also reads the free and total bytes of the volume.
//
// The returned volumeInfo holds everything that could be read, also if err is not nil.
func (r *volumeInfoReader) read(mountedVolumes map[string]string, name string, withSpace bool) (volumeInfo, error) {
	state := r.state(name)
	devicePath, rootPath := getVolumePaths(mountedVolumes, name)

	rootPathPtr, err := windows.UTF16PtrFromString(rootPath)
	if err != nil {
		return volumeInfo{}, fmt.Errorf("invalid volume path %q: %w", rootPath, err)
	}

	var errs []error

	if !state.hasStatic || state.devicePath != devicePath || r.now().Sub(state.readAt) >= staticVolumeInfoMaxAge {
		state.hasStatic = false

		diskIDs, err := r.readDiskIDs(devicePath)
		if err != nil {
			errs = append(errs, fmt.Errorf("could not identify physical drive for %s: %w", name, err))
		} else {
			state.hasStatic = true
			state.devicePath = devicePath
			state.diskIDs = diskIDs
			state.driveType = windows.GetDriveType(rootPathPtr)
			state.readAt = r.now()
		}
	}

	info := volumeInfo{}

	driveType := state.driveType
	if state.hasStatic {
		info.diskIDs = state.diskIDs
	} else {
		driveType = windows.GetDriveType(rootPathPtr)
	}

	info.volumeType = getDriveType(driveType)

	// Removable and CD-ROM drives may have no media or no recognized file system.
	mediaOptional := driveType == windows.DRIVE_CDROM || driveType == windows.DRIVE_REMOVABLE

	var serialNumber, fsFlags uint32

	err = windows.GetVolumeInformation(
		rootPathPtr,
		&r.label[0], uint32(len(r.label)),
		&serialNumber, nil, &fsFlags,
		&r.fsName[0], uint32(len(r.fsName)),
	)
	if err != nil {
		if !mediaOptional {
			errs = append(errs, fmt.Errorf("could not get volume information for %s: %w", rootPath, err))
		}

		return info, errors.Join(errs...)
	}

	info.label = windows.UTF16ToString(r.label)
	info.filesystem = windows.UTF16ToString(r.fsName)
	info.serialNumber = strings.ToUpper(strconv.FormatUint(uint64(serialNumber), 16))
	info.readonly = utils.BoolToFloat(fsFlags&windows.FILE_READ_ONLY_VOLUME != 0)
	info.hasReadonly = true

	if withSpace {
		var callerFreeBytes, sizeBytes, freeBytes uint64

		// The free bytes are those of the volume. The total bytes are reduced to the quota
		// if NTFS disk quotas apply to the account of windows_exporter.
		err = windows.GetDiskFreeSpaceEx(rootPathPtr, &callerFreeBytes, &sizeBytes, &freeBytes)

		switch {
		case err != nil:
			if !mediaOptional {
				errs = append(errs, fmt.Errorf("could not get free space for %s: %w", rootPath, err))
			}
		case freeBytes <= sizeBytes:
			info.freeBytes = float64(freeBytes)
			info.sizeBytes = float64(sizeBytes)
			info.hasSpace = true
		}
	}

	return info, errors.Join(errs...)
}

// getVolumePaths returns the paths to open the volume name with CreateFile and to pass to
// GetVolumeInformation and GetDiskFreeSpaceEx, which require the root directory.
func getVolumePaths(mountedVolumes map[string]string, name string) (string, string) {
	// GetVolumePathNamesForVolumeName returns the volume GUID path as \\?\Volume{GUID}.
	// According to https://learn.microsoft.com/en-us/windows/win32/api/ioapiset/nf-ioapiset-deviceiocontrol#remarks
	// a volume is opened in the Win32 device namespace \\.\, so the \\?\ prefix is replaced.
	if volumeGUID, ok := mountedVolumes[name]; ok {
		volumePath := strings.TrimPrefix(volumeGUID, `\\?\`)

		return `\\.\` + volumePath, `\\?\` + volumePath + `\`
	}

	// Volumes without a mount point are reported by their device name, e.g. HarddiskVolume1.
	if strings.HasPrefix(name, "HarddiskVolume") {
		return `\\.\GLOBALROOT\Device\` + name, `\\?\GLOBALROOT\Device\` + name + `\`
	}

	return `\\.\` + name, name + `\`
}

// readDiskIDs returns the semicolon-separated numbers of the disks the volume is located on.
func (r *volumeInfoReader) readDiskIDs(devicePath string) (string, error) {
	devicePathPtr, err := windows.UTF16PtrFromString(devicePath)
	if err != nil {
		return "", fmt.Errorf("invalid device path %q: %w", devicePath, err)
	}

	// mode has to include FILE_SHARE permission to allow concurrent access to the disk.
	// use 0 as access mode to avoid admin permission.
	mode := uint32(windows.FILE_SHARE_READ | windows.FILE_SHARE_WRITE | windows.FILE_SHARE_DELETE)

	volumeHandle, err := windows.CreateFile(devicePathPtr, 0, mode, nil, windows.OPEN_EXISTING, windows.FILE_ATTRIBUTE_READONLY, 0)
	if err != nil {
		return "", fmt.Errorf("could not open volume %s: %w", devicePath, err)
	}

	defer func(fd windows.Handle) {
		_ = windows.Close(fd)
	}(volumeHandle)

	for {
		var bytesReturned uint32

		err = windows.DeviceIoControl(volumeHandle, ioctlVolumeGetVolumeDiskExtents, nil, 0,
			&r.extents[0], uint32(len(r.extents)), &bytesReturned, nil)
		if errors.Is(err, windows.ERROR_MORE_DATA) {
			// The buffer holds NumberOfDiskExtents.
			required := diskExtentsHeaderSize + int(binary.LittleEndian.Uint32(r.extents))*diskExtentSize
			if required <= len(r.extents) {
				return "", fmt.Errorf("invalid required disk extents buffer size: %d", required)
			}

			r.extents = make([]byte, required)

			continue
		}

		if err != nil {
			return "", err
		}

		return parseDiskIDs(r.extents[:bytesReturned])
	}
}

// parseDiskIDs returns the sorted, unique disk numbers of a VOLUME_DISK_EXTENTS structure, separated by semicolons.
func parseDiskIDs(volumeDiskExtents []byte) (string, error) {
	if len(volumeDiskExtents) < diskExtentsHeaderSize {
		return "", fmt.Errorf("disk extents too short: %d bytes", len(volumeDiskExtents))
	}

	numDiskExtents := int(binary.LittleEndian.Uint32(volumeDiskExtents))
	if numDiskExtents < 1 {
		return "", errors.New("no disk IDs returned")
	}

	if len(volumeDiskExtents) < diskExtentsHeaderSize+numDiskExtents*diskExtentSize {
		return "", fmt.Errorf("disk extents truncated: %d extents in %d bytes", numDiskExtents, len(volumeDiskExtents))
	}

	diskIDs := make([]string, numDiskExtents)
	for i := range diskIDs {
		diskIDs[i] = strconv.FormatUint(uint64(binary.LittleEndian.Uint32(volumeDiskExtents[diskExtentsHeaderSize+i*diskExtentSize:])), 10)
	}

	slices.Sort(diskIDs)

	return strings.Join(slices.Compact(diskIDs), ";"), nil
}

func getDriveType(driveType uint32) string {
	switch driveType {
	case windows.DRIVE_UNKNOWN:
		return "unknown"
	case windows.DRIVE_NO_ROOT_DIR:
		return "norootdir"
	case windows.DRIVE_REMOVABLE:
		return "removable"
	case windows.DRIVE_FIXED:
		return "fixed"
	case windows.DRIVE_REMOTE:
		return "remote"
	case windows.DRIVE_CDROM:
		return "cdrom"
	case windows.DRIVE_RAMDISK:
		return "ramdisk"
	default:
		return "unknown"
	}
}
