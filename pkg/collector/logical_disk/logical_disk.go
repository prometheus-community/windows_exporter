//go:build windows

package logical_disk

import (
	"encoding/binary"
	"fmt"
	"golang.org/x/sys/windows"
	"regexp"
	"strconv"
	"strings"

	"github.com/alecthomas/kingpin/v2"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/prometheus-community/windows_exporter/pkg/perflib"
	"github.com/prometheus-community/windows_exporter/pkg/types"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	Name = "logical_disk"

	FlagLogicalDiskVolumeExclude = "collector.logical_disk.volume-exclude"
	FlagLogicalDiskVolumeInclude = "collector.logical_disk.volume-include"
)

type Config struct {
	VolumeInclude string `yaml:"volume_include"`
	VolumeExclude string `yaml:"volume_exclude"`
}

var ConfigDefaults = Config{
	VolumeInclude: ".+",
	VolumeExclude: "",
}

// A collector is a Prometheus collector for perflib logicalDisk metrics
type collector struct {
	logger log.Logger

	volumeInclude *string
	volumeExclude *string

	Information      *prometheus.Desc
	ReadOnly         *prometheus.Desc
	RequestsQueued   *prometheus.Desc
	AvgReadQueue     *prometheus.Desc
	AvgWriteQueue    *prometheus.Desc
	ReadBytesTotal   *prometheus.Desc
	ReadsTotal       *prometheus.Desc
	WriteBytesTotal  *prometheus.Desc
	WritesTotal      *prometheus.Desc
	ReadTime         *prometheus.Desc
	WriteTime        *prometheus.Desc
	TotalSpace       *prometheus.Desc
	FreeSpace        *prometheus.Desc
	IdleTime         *prometheus.Desc
	SplitIOs         *prometheus.Desc
	ReadLatency      *prometheus.Desc
	WriteLatency     *prometheus.Desc
	ReadWriteLatency *prometheus.Desc

	volumeIncludePattern *regexp.Regexp
	volumeExcludePattern *regexp.Regexp
}

type volumeInfo struct {
	filesystem   string
	serialNumber string
	label        string
	volumeType   string
	readonly     float64
}

func New(logger log.Logger, config *Config) types.Collector {
	if config == nil {
		config = &ConfigDefaults
	}

	c := &collector{
		volumeExclude: &config.VolumeExclude,
		volumeInclude: &config.VolumeInclude,
	}
	c.SetLogger(logger)
	return c
}

func NewWithFlags(app *kingpin.Application) types.Collector {
	c := &collector{
		volumeInclude: app.Flag(
			FlagLogicalDiskVolumeInclude,
			"Regexp of volumes to include. Volume name must both match include and not match exclude to be included.",
		).Default(ConfigDefaults.VolumeInclude).String(),
		volumeExclude: app.Flag(
			FlagLogicalDiskVolumeExclude,
			"Regexp of volumes to exclude. Volume name must both match include and not match exclude to be included.",
		).Default(ConfigDefaults.VolumeExclude).String(),
	}

	return c
}

func (c *collector) GetName() string {
	return Name
}

func (c *collector) SetLogger(logger log.Logger) {
	c.logger = log.With(logger, "collector", Name)
}

func (c *collector) GetPerfCounter() ([]string, error) {
	return []string{"LogicalDisk"}, nil
}

func (c *collector) Build() error {
	c.Information = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "info"),
		"A metric with a constant '1' value labeled with logical disk information",
		[]string{"disk", "type", "volume", "volume_name", "filesystem", "serial_number"},
		nil,
	)
	c.ReadOnly = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "readonly"),
		"Whether the logical disk is read-only",
		[]string{"volume"},
		nil,
	)
	c.RequestsQueued = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "requests_queued"),
		"The number of requests queued to the disk (LogicalDisk.CurrentDiskQueueLength)",
		[]string{"volume"},
		nil,
	)

	c.AvgReadQueue = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "avg_read_requests_queued"),
		"Average number of read requests that were queued for the selected disk during the sample interval (LogicalDisk.AvgDiskReadQueueLength)",
		[]string{"volume"},
		nil,
	)

	c.AvgWriteQueue = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "avg_write_requests_queued"),
		"Average number of write requests that were queued for the selected disk during the sample interval (LogicalDisk.AvgDiskWriteQueueLength)",
		[]string{"volume"},
		nil,
	)

	c.ReadBytesTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "read_bytes_total"),
		"The number of bytes transferred from the disk during read operations (LogicalDisk.DiskReadBytesPerSec)",
		[]string{"volume"},
		nil,
	)

	c.ReadsTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "reads_total"),
		"The number of read operations on the disk (LogicalDisk.DiskReadsPerSec)",
		[]string{"volume"},
		nil,
	)

	c.WriteBytesTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "write_bytes_total"),
		"The number of bytes transferred to the disk during write operations (LogicalDisk.DiskWriteBytesPerSec)",
		[]string{"volume"},
		nil,
	)

	c.WritesTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "writes_total"),
		"The number of write operations on the disk (LogicalDisk.DiskWritesPerSec)",
		[]string{"volume"},
		nil,
	)

	c.ReadTime = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "read_seconds_total"),
		"Seconds that the disk was busy servicing read requests (LogicalDisk.PercentDiskReadTime)",
		[]string{"volume"},
		nil,
	)

	c.WriteTime = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "write_seconds_total"),
		"Seconds that the disk was busy servicing write requests (LogicalDisk.PercentDiskWriteTime)",
		[]string{"volume"},
		nil,
	)

	c.FreeSpace = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "free_bytes"),
		"Free space in bytes, updates every 10-15 min (LogicalDisk.PercentFreeSpace)",
		[]string{"volume"},
		nil,
	)

	c.TotalSpace = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "size_bytes"),
		"Total space in bytes, updates every 10-15 min (LogicalDisk.PercentFreeSpace_Base)",
		[]string{"volume"},
		nil,
	)

	c.IdleTime = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "idle_seconds_total"),
		"Seconds that the disk was idle (LogicalDisk.PercentIdleTime)",
		[]string{"volume"},
		nil,
	)

	c.SplitIOs = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "split_ios_total"),
		"The number of I/Os to the disk were split into multiple I/Os (LogicalDisk.SplitIOPerSec)",
		[]string{"volume"},
		nil,
	)

	c.ReadLatency = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "read_latency_seconds_total"),
		"Shows the average time, in seconds, of a read operation from the disk (LogicalDisk.AvgDiskSecPerRead)",
		[]string{"volume"},
		nil,
	)

	c.WriteLatency = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "write_latency_seconds_total"),
		"Shows the average time, in seconds, of a write operation to the disk (LogicalDisk.AvgDiskSecPerWrite)",
		[]string{"volume"},
		nil,
	)

	c.ReadWriteLatency = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "read_write_latency_seconds_total"),
		"Shows the time, in seconds, of the average disk transfer (LogicalDisk.AvgDiskSecPerTransfer)",
		[]string{"volume"},
		nil,
	)

	var err error
	c.volumeIncludePattern, err = regexp.Compile(fmt.Sprintf("^(?:%s)$", *c.volumeInclude))
	if err != nil {
		return err
	}

	c.volumeExcludePattern, err = regexp.Compile(fmt.Sprintf("^(?:%s)$", *c.volumeExclude))
	if err != nil {
		return err
	}

	return nil
}

// Collect sends the metric values for each metric
// to the provided prometheus Metric channel.
func (c *collector) Collect(ctx *types.ScrapeContext, ch chan<- prometheus.Metric) error {
	if err := c.collect(ctx, ch); err != nil {
		_ = level.Error(c.logger).Log("msg", "failed collecting logical_disk metrics", "err", err)
		return err
	}
	return nil
}

// Win32_PerfRawData_PerfDisk_LogicalDisk docs:
// - https://msdn.microsoft.com/en-us/windows/hardware/aa394307(v=vs.71) - Win32_PerfRawData_PerfDisk_LogicalDisk class
// - https://msdn.microsoft.com/en-us/library/ms803973.aspx - LogicalDisk object reference
type logicalDisk struct {
	Name                    string
	CurrentDiskQueueLength  float64 `perflib:"Current Disk Queue Length"`
	AvgDiskReadQueueLength  float64 `perflib:"Avg. Disk Read Queue Length"`
	AvgDiskWriteQueueLength float64 `perflib:"Avg. Disk Write Queue Length"`
	DiskReadBytesPerSec     float64 `perflib:"Disk Read Bytes/sec"`
	DiskReadsPerSec         float64 `perflib:"Disk Reads/sec"`
	DiskWriteBytesPerSec    float64 `perflib:"Disk Write Bytes/sec"`
	DiskWritesPerSec        float64 `perflib:"Disk Writes/sec"`
	PercentDiskReadTime     float64 `perflib:"% Disk Read Time"`
	PercentDiskWriteTime    float64 `perflib:"% Disk Write Time"`
	PercentFreeSpace        float64 `perflib:"% Free Space_Base"`
	PercentFreeSpace_Base   float64 `perflib:"Free Megabytes"`
	PercentIdleTime         float64 `perflib:"% Idle Time"`
	SplitIOPerSec           float64 `perflib:"Split IO/Sec"`
	AvgDiskSecPerRead       float64 `perflib:"Avg. Disk sec/Read"`
	AvgDiskSecPerWrite      float64 `perflib:"Avg. Disk sec/Write"`
	AvgDiskSecPerTransfer   float64 `perflib:"Avg. Disk sec/Transfer"`
}

func (c *collector) collect(ctx *types.ScrapeContext, ch chan<- prometheus.Metric) error {
	var (
		err    error
		diskID string
		info   volumeInfo
		dst    []logicalDisk
	)

	if err = perflib.UnmarshalObject(ctx.PerfObjects["LogicalDisk"], &dst, c.logger); err != nil {
		return err
	}

	for _, volume := range dst {
		if volume.Name == "_Total" ||
			c.volumeExcludePattern.MatchString(volume.Name) ||
			!c.volumeIncludePattern.MatchString(volume.Name) {
			continue
		}

		diskID, err = getDiskIDByVolume(volume.Name)
		if err != nil {
			_ = level.Warn(c.logger).Log("msg", "failed to get disk ID for "+volume.Name, "err", err)
		}

		info, err = getVolumeInfo(volume.Name)
		if err != nil {
			_ = level.Warn(c.logger).Log("msg", "failed to get volume information for %s"+volume.Name, "err", err)
		}

		ch <- prometheus.MustNewConstMetric(
			c.Information,
			prometheus.GaugeValue,
			1,
			diskID,
			info.volumeType,
			volume.Name,
			info.label,
			info.filesystem,
			info.serialNumber,
		)

		ch <- prometheus.MustNewConstMetric(
			c.RequestsQueued,
			prometheus.GaugeValue,
			volume.CurrentDiskQueueLength,
			volume.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.AvgReadQueue,
			prometheus.GaugeValue,
			volume.AvgDiskReadQueueLength*perflib.TicksToSecondScaleFactor,
			volume.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.AvgWriteQueue,
			prometheus.GaugeValue,
			volume.AvgDiskWriteQueueLength*perflib.TicksToSecondScaleFactor,
			volume.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.ReadBytesTotal,
			prometheus.CounterValue,
			volume.DiskReadBytesPerSec,
			volume.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.ReadsTotal,
			prometheus.CounterValue,
			volume.DiskReadsPerSec,
			volume.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.WriteBytesTotal,
			prometheus.CounterValue,
			volume.DiskWriteBytesPerSec,
			volume.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.WritesTotal,
			prometheus.CounterValue,
			volume.DiskWritesPerSec,
			volume.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.ReadTime,
			prometheus.CounterValue,
			volume.PercentDiskReadTime,
			volume.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.WriteTime,
			prometheus.CounterValue,
			volume.PercentDiskWriteTime,
			volume.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.FreeSpace,
			prometheus.GaugeValue,
			volume.PercentFreeSpace_Base*1024*1024,
			volume.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.TotalSpace,
			prometheus.GaugeValue,
			volume.PercentFreeSpace*1024*1024,
			volume.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.IdleTime,
			prometheus.CounterValue,
			volume.PercentIdleTime,
			volume.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.SplitIOs,
			prometheus.CounterValue,
			volume.SplitIOPerSec,
			volume.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.ReadLatency,
			prometheus.CounterValue,
			volume.AvgDiskSecPerRead*perflib.TicksToSecondScaleFactor,
			volume.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.WriteLatency,
			prometheus.CounterValue,
			volume.AvgDiskSecPerWrite*perflib.TicksToSecondScaleFactor,
			volume.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.ReadWriteLatency,
			prometheus.CounterValue,
			volume.AvgDiskSecPerTransfer*perflib.TicksToSecondScaleFactor,
			volume.Name,
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

// getDiskIDByVolume returns the disk ID for a given volume.
func getDiskIDByVolume(rootDrive string) (string, error) {
	// Open a volume handle to the Disk Root.
	var err error
	var f windows.Handle

	// mode has to include FILE_SHARE permission to allow concurrent access to the disk.
	// use 0 as access mode to avoid admin permission.
	mode := uint32(windows.FILE_SHARE_READ | windows.FILE_SHARE_WRITE | windows.FILE_SHARE_DELETE)
	f, err = windows.CreateFile(
		windows.StringToUTF16Ptr(`\\.\`+rootDrive),
		0, mode, nil, windows.OPEN_EXISTING, uint32(windows.FILE_ATTRIBUTE_READONLY), 0)

	if err != nil {
		return "", err
	}

	defer windows.Close(f)

	controlCode := uint32(5636096) // IOCTL_VOLUME_GET_VOLUME_DISK_EXTENTS
	volumeDiskExtents := make([]byte, 16*1024)

	var bytesReturned uint32
	err = windows.DeviceIoControl(f, controlCode, nil, 0, &volumeDiskExtents[0], uint32(len(volumeDiskExtents)), &bytesReturned, nil)
	if err != nil {
		return "", err
	}

	if uint(binary.LittleEndian.Uint32(volumeDiskExtents)) != 1 {
		return "", fmt.Errorf("could not identify physical drive for %s", rootDrive)
	}

	diskId := strconv.FormatUint(uint64(binary.LittleEndian.Uint32(volumeDiskExtents[8:])), 10)

	return diskId, nil
}

func getVolumeInfo(rootDrive string) (volumeInfo, error) {
	if !strings.HasSuffix(rootDrive, ":") {
		return volumeInfo{}, nil
	}

	volPath := windows.StringToUTF16Ptr(rootDrive + `\`)

	volBufLabel := make([]uint16, windows.MAX_PATH+1)
	volSerialNum := uint32(0)
	fsFlags := uint32(0)
	volBufType := make([]uint16, windows.MAX_PATH+1)

	driveType := windows.GetDriveType(volPath)

	err := windows.GetVolumeInformation(volPath, &volBufLabel[0], uint32(len(volBufLabel)),
		&volSerialNum, nil, &fsFlags, &volBufType[0], uint32(len(volBufType)))

	if err != nil {
		if driveType != windows.DRIVE_CDROM && driveType != windows.DRIVE_REMOVABLE {
			return volumeInfo{}, err
		}

		return volumeInfo{}, nil
	}

	return volumeInfo{
		volumeType:   getDriveType(driveType),
		label:        windows.UTF16PtrToString(&volBufLabel[0]),
		filesystem:   windows.UTF16PtrToString(&volBufType[0]),
		serialNumber: fmt.Sprintf("%X", volSerialNum),
		readonly:     float64(fsFlags & windows.FILE_READ_ONLY_VOLUME),
	}, nil
}
