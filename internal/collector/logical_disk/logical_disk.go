// Copyright 2024 The Prometheus Authors
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
	"fmt"
	"log/slog"
	"regexp"
	"slices"
	"strconv"
	"strings"

	"github.com/alecthomas/kingpin/v2"
	"github.com/prometheus-community/windows_exporter/internal/mi"
	"github.com/prometheus-community/windows_exporter/internal/perfdata"
	"github.com/prometheus-community/windows_exporter/internal/types"
	"github.com/prometheus/client_golang/prometheus"
	"golang.org/x/sys/windows"
)

const Name = "logical_disk"

type Config struct {
	VolumeInclude *regexp.Regexp `yaml:"volume_include"`
	VolumeExclude *regexp.Regexp `yaml:"volume_exclude"`
}

//nolint:gochecknoglobals
var ConfigDefaults = Config{
	VolumeInclude: types.RegExpAny,
	VolumeExclude: types.RegExpEmpty,
}

// A Collector is a Prometheus Collector for perflib logicalDisk metrics.
type Collector struct {
	config Config
	logger *slog.Logger

	perfDataCollector *perfdata.Collector

	avgReadQueue     *prometheus.Desc
	avgWriteQueue    *prometheus.Desc
	freeSpace        *prometheus.Desc
	idleTime         *prometheus.Desc
	information      *prometheus.Desc
	readBytesTotal   *prometheus.Desc
	readLatency      *prometheus.Desc
	readOnly         *prometheus.Desc
	readsTotal       *prometheus.Desc
	readTime         *prometheus.Desc
	readWriteLatency *prometheus.Desc
	requestsQueued   *prometheus.Desc
	splitIOs         *prometheus.Desc
	totalSpace       *prometheus.Desc
	writeBytesTotal  *prometheus.Desc
	writeLatency     *prometheus.Desc
	writesTotal      *prometheus.Desc
	writeTime        *prometheus.Desc
}

type volumeInfo struct {
	diskIDs      string
	filesystem   string
	serialNumber string
	label        string
	volumeType   string
	readonly     float64
}

func New(config *Config) *Collector {
	if config == nil {
		config = &ConfigDefaults
	}

	if config.VolumeExclude == nil {
		config.VolumeExclude = ConfigDefaults.VolumeExclude
	}

	if config.VolumeInclude == nil {
		config.VolumeInclude = ConfigDefaults.VolumeInclude
	}

	c := &Collector{
		config: *config,
	}

	return c
}

func NewWithFlags(app *kingpin.Application) *Collector {
	c := &Collector{
		config: ConfigDefaults,
	}

	var volumeExclude, volumeInclude string

	app.Flag(
		"collector.logical_disk.volume-exclude",
		"Regexp of volumes to exclude. Volume name must both match include and not match exclude to be included.",
	).Default("").StringVar(&volumeExclude)

	app.Flag(
		"collector.logical_disk.volume-include",
		"Regexp of volumes to include. Volume name must both match include and not match exclude to be included.",
	).Default(".+").StringVar(&volumeInclude)

	app.Action(func(*kingpin.ParseContext) error {
		var err error

		c.config.VolumeExclude, err = regexp.Compile(fmt.Sprintf("^(?:%s)$", volumeExclude))
		if err != nil {
			return fmt.Errorf("collector.logical_disk.volume-exclude: %w", err)
		}

		c.config.VolumeInclude, err = regexp.Compile(fmt.Sprintf("^(?:%s)$", volumeInclude))
		if err != nil {
			return fmt.Errorf("collector.logical_disk.volume-include: %w", err)
		}

		return nil
	})

	return c
}

func (c *Collector) GetName() string {
	return Name
}

func (c *Collector) Close() error {
	return nil
}

func (c *Collector) Build(logger *slog.Logger, _ *mi.Session) error {
	c.logger = logger.With(slog.String("collector", Name))

	var err error

	c.perfDataCollector, err = perfdata.NewCollector("LogicalDisk", perfdata.InstancesAll, []string{
		currentDiskQueueLength,
		avgDiskReadQueueLength,
		avgDiskWriteQueueLength,
		diskReadBytesPerSec,
		diskReadsPerSec,
		diskWriteBytesPerSec,
		diskWritesPerSec,
		percentDiskReadTime,
		percentDiskWriteTime,
		percentFreeSpace,
		freeSpace,
		percentIdleTime,
		splitIOPerSec,
		avgDiskSecPerRead,
		avgDiskSecPerWrite,
		avgDiskSecPerTransfer,
	})
	if err != nil {
		return fmt.Errorf("failed to create LogicalDisk collector: %w", err)
	}

	c.information = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "info"),
		"A metric with a constant '1' value labeled with logical disk information",
		[]string{"disk", "type", "volume", "volume_name", "filesystem", "serial_number"},
		nil,
	)
	c.readOnly = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "readonly"),
		"Whether the logical disk is read-only",
		[]string{"volume"},
		nil,
	)
	c.requestsQueued = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "requests_queued"),
		"The number of requests queued to the disk (LogicalDisk.CurrentDiskQueueLength)",
		[]string{"volume"},
		nil,
	)

	c.avgReadQueue = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "avg_read_requests_queued"),
		"Average number of read requests that were queued for the selected disk during the sample interval (LogicalDisk.AvgDiskReadQueueLength)",
		[]string{"volume"},
		nil,
	)

	c.avgWriteQueue = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "avg_write_requests_queued"),
		"Average number of write requests that were queued for the selected disk during the sample interval (LogicalDisk.AvgDiskWriteQueueLength)",
		[]string{"volume"},
		nil,
	)

	c.readBytesTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "read_bytes_total"),
		"The number of bytes transferred from the disk during read operations (LogicalDisk.DiskReadBytesPerSec)",
		[]string{"volume"},
		nil,
	)

	c.readsTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "reads_total"),
		"The number of read operations on the disk (LogicalDisk.DiskReadsPerSec)",
		[]string{"volume"},
		nil,
	)

	c.writeBytesTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "write_bytes_total"),
		"The number of bytes transferred to the disk during write operations (LogicalDisk.DiskWriteBytesPerSec)",
		[]string{"volume"},
		nil,
	)

	c.writesTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "writes_total"),
		"The number of write operations on the disk (LogicalDisk.DiskWritesPerSec)",
		[]string{"volume"},
		nil,
	)

	c.readTime = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "read_seconds_total"),
		"Seconds that the disk was busy servicing read requests (LogicalDisk.PercentDiskReadTime)",
		[]string{"volume"},
		nil,
	)

	c.writeTime = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "write_seconds_total"),
		"Seconds that the disk was busy servicing write requests (LogicalDisk.PercentDiskWriteTime)",
		[]string{"volume"},
		nil,
	)

	c.freeSpace = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "free_bytes"),
		"Free space in bytes, updates every 10-15 min (LogicalDisk.PercentFreeSpace)",
		[]string{"volume"},
		nil,
	)

	c.totalSpace = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "size_bytes"),
		"Total space in bytes, updates every 10-15 min (LogicalDisk.PercentFreeSpace_Base)",
		[]string{"volume"},
		nil,
	)

	c.idleTime = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "idle_seconds_total"),
		"Seconds that the disk was idle (LogicalDisk.PercentIdleTime)",
		[]string{"volume"},
		nil,
	)

	c.splitIOs = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "split_ios_total"),
		"The number of I/Os to the disk were split into multiple I/Os (LogicalDisk.SplitIOPerSec)",
		[]string{"volume"},
		nil,
	)

	c.readLatency = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "read_latency_seconds_total"),
		"Shows the average time, in seconds, of a read operation from the disk (LogicalDisk.AvgDiskSecPerRead)",
		[]string{"volume"},
		nil,
	)

	c.writeLatency = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "write_latency_seconds_total"),
		"Shows the average time, in seconds, of a write operation to the disk (LogicalDisk.AvgDiskSecPerWrite)",
		[]string{"volume"},
		nil,
	)

	c.readWriteLatency = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "read_write_latency_seconds_total"),
		"Shows the time, in seconds, of the average disk transfer (LogicalDisk.AvgDiskSecPerTransfer)",
		[]string{"volume"},
		nil,
	)

	return nil
}

// Collect sends the metric values for each metric
// to the provided prometheus Metric channel.
func (c *Collector) Collect(ch chan<- prometheus.Metric) error {
	var (
		err  error
		info volumeInfo
	)

	perfData, err := c.perfDataCollector.Collect()
	if err != nil {
		return fmt.Errorf("failed to collect LogicalDisk metrics: %w", err)
	}

	for name, volume := range perfData {
		if c.config.VolumeExclude.MatchString(name) || !c.config.VolumeInclude.MatchString(name) {
			continue
		}

		info, err = getVolumeInfo(name)
		if err != nil {
			c.logger.Warn("failed to get volume information for "+name,
				slog.Any("err", err),
			)
		}

		ch <- prometheus.MustNewConstMetric(
			c.information,
			prometheus.GaugeValue,
			1,
			info.diskIDs,
			info.volumeType,
			name,
			info.label,
			info.filesystem,
			info.serialNumber,
		)

		ch <- prometheus.MustNewConstMetric(
			c.requestsQueued,
			prometheus.GaugeValue,
			volume[currentDiskQueueLength].FirstValue,
			name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.avgReadQueue,
			prometheus.GaugeValue,
			volume[avgDiskReadQueueLength].FirstValue*perfdata.TicksToSecondScaleFactor,
			name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.avgWriteQueue,
			prometheus.GaugeValue,
			volume[avgDiskWriteQueueLength].FirstValue*perfdata.TicksToSecondScaleFactor,
			name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.readBytesTotal,
			prometheus.CounterValue,
			volume[diskReadBytesPerSec].FirstValue,
			name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.readsTotal,
			prometheus.CounterValue,
			volume[diskReadsPerSec].FirstValue,
			name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.writeBytesTotal,
			prometheus.CounterValue,
			volume[diskWriteBytesPerSec].FirstValue,
			name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.writesTotal,
			prometheus.CounterValue,
			volume[diskWritesPerSec].FirstValue,
			name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.readTime,
			prometheus.CounterValue,
			volume[percentDiskReadTime].FirstValue,
			name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.writeTime,
			prometheus.CounterValue,
			volume[percentDiskWriteTime].FirstValue,
			name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.freeSpace,
			prometheus.GaugeValue,
			volume[freeSpace].FirstValue*1024*1024,
			name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.totalSpace,
			prometheus.GaugeValue,
			volume[percentFreeSpace].SecondValue*1024*1024,
			name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.idleTime,
			prometheus.CounterValue,
			volume[percentIdleTime].FirstValue,
			name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.splitIOs,
			prometheus.CounterValue,
			volume[splitIOPerSec].FirstValue,
			name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.readLatency,
			prometheus.CounterValue,
			volume[avgDiskSecPerRead].FirstValue*perfdata.TicksToSecondScaleFactor,
			name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.writeLatency,
			prometheus.CounterValue,
			volume[avgDiskSecPerWrite].FirstValue*perfdata.TicksToSecondScaleFactor,
			name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.readWriteLatency,
			prometheus.CounterValue,
			volume[avgDiskSecPerTransfer].FirstValue*perfdata.TicksToSecondScaleFactor,
			name,
		)
	}

	return nil
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

// diskExtentSize Size of the DiskExtent structure in bytes.
const diskExtentSize = 24

// getDiskIDByVolume returns the disk ID for a given volume.
func getVolumeInfo(rootDrive string) (volumeInfo, error) {
	volumePath := rootDrive

	// If rootDrive is a NTFS directory, convert it to a volume GUID.
	if strings.Contains(volumePath, `\`) {
		// GetVolumeNameForVolumeMountPoint expects a trailing backslash.
		volumePath += `\`

		volumePathName, err := windows.UTF16PtrFromString(volumePath)
		if err != nil {
			return volumeInfo{}, fmt.Errorf("could not convert rootDrive to volume path %s: %w", volumePath, err)
		}

		volumeGUIDPtr := make([]uint16, 50)
		if err := windows.GetVolumeNameForVolumeMountPoint(volumePathName, &volumeGUIDPtr[0], uint32(len(volumeGUIDPtr))); err != nil {
			return volumeInfo{}, fmt.Errorf("could not get volume GUID for volume %s: %w", volumePath, err)
		}

		volumePath = windows.UTF16ToString(volumeGUIDPtr)

		// GetVolumeNameForVolumeMountPoint returns the volume GUID path as \\?\Volume{GUID}\
		// According https://learn.microsoft.com/en-us/windows/win32/api/ioapiset/nf-ioapiset-deviceiocontrol#remarks
		// Win32 Drive Namespace is prefixed with \\.\, so we need to remove the \\?\ prefix.
		volumePath, _ = strings.CutPrefix(volumePath, `\\?\`)

		// https://stackoverflow.com/questions/55710326/how-to-get-the-physical-device-that-a-volume-guid-path-belongs-to#comment98104360_55710326
		// DeviceIoControl expects no trailing backslash in the volume GUID path.
		volumePath = strings.TrimRight(volumePath, `\`)
	}

	volumePathPtr := windows.StringToUTF16Ptr(`\\.\` + volumePath)

	// mode has to include FILE_SHARE permission to allow concurrent access to the disk.
	// use 0 as access mode to avoid admin permission.
	mode := uint32(windows.FILE_SHARE_READ | windows.FILE_SHARE_WRITE | windows.FILE_SHARE_DELETE)
	attr := uint32(windows.FILE_ATTRIBUTE_READONLY)

	volumeHandle, err := windows.CreateFile(volumePathPtr, 0, mode, nil, windows.OPEN_EXISTING, attr, 0)
	if err != nil {
		return volumeInfo{}, fmt.Errorf("could not open volume for %s: %w", rootDrive, err)
	}

	defer windows.Close(volumeHandle)

	controlCode := uint32(5636096) // IOCTL_VOLUME_GET_VOLUME_DISK_EXTENTS
	volumeDiskExtents := make([]byte, 16*1024)

	var bytesReturned uint32

	err = windows.DeviceIoControl(volumeHandle, controlCode, nil, 0, &volumeDiskExtents[0], uint32(len(volumeDiskExtents)), &bytesReturned, nil)
	if err != nil {
		return volumeInfo{}, fmt.Errorf("could not identify physical drive for %s: %w", rootDrive, err)
	}

	numDiskIDs := uint(binary.LittleEndian.Uint32(volumeDiskExtents))
	if numDiskIDs < 1 {
		return volumeInfo{}, fmt.Errorf("could not identify physical drive for %s: no disk IDs returned", rootDrive)
	}

	diskIDs := make([]string, numDiskIDs)

	for i := range numDiskIDs {
		diskIDs[i] = strconv.FormatUint(uint64(binary.LittleEndian.Uint32(volumeDiskExtents[8+i*diskExtentSize:])), 10)
	}

	slices.Sort(diskIDs)
	diskIDs = slices.Compact(diskIDs)

	volumeInformationRootDrive := volumePath + `\`

	if strings.Contains(volumePath, `Volume`) {
		volumeInformationRootDrive = `\\?\` + volumeInformationRootDrive
	}

	volumeInformationRootDrivePtr := windows.StringToUTF16Ptr(volumeInformationRootDrive)
	driveType := windows.GetDriveType(volumeInformationRootDrivePtr)
	volBufLabel := make([]uint16, windows.MAX_PATH+1)
	volSerialNum := uint32(0)
	fsFlags := uint32(0)
	volBufType := make([]uint16, windows.MAX_PATH+1)

	err = windows.GetVolumeInformation(
		volumeInformationRootDrivePtr,
		&volBufLabel[0], uint32(len(volBufLabel)),
		&volSerialNum, nil, &fsFlags,
		&volBufType[0], uint32(len(volBufType)),
	)
	if err != nil {
		if driveType == windows.DRIVE_CDROM || driveType == windows.DRIVE_REMOVABLE {
			return volumeInfo{}, nil
		}

		return volumeInfo{}, fmt.Errorf("could not get volume information for %s: %w", volumeInformationRootDrive, err)
	}

	return volumeInfo{
		diskIDs:      strings.Join(diskIDs, ";"),
		volumeType:   getDriveType(driveType),
		label:        windows.UTF16PtrToString(&volBufLabel[0]),
		filesystem:   windows.UTF16PtrToString(&volBufType[0]),
		serialNumber: fmt.Sprintf("%X", volSerialNum),
		readonly:     float64(fsFlags & windows.FILE_READ_ONLY_VOLUME),
	}, nil
}
