//go:build windows

package logical_disk

import (
	"encoding/binary"
	"fmt"
	"regexp"
	"slices"
	"strconv"
	"strings"

	"github.com/alecthomas/kingpin/v2"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/prometheus-community/windows_exporter/pkg/perflib"
	"github.com/prometheus-community/windows_exporter/pkg/types"
	"github.com/prometheus/client_golang/prometheus"
	"golang.org/x/sys/windows"
)

const Name = "logical_disk"

type Config struct {
	VolumeInclude *regexp.Regexp `yaml:"volume_include"`
	VolumeExclude *regexp.Regexp `yaml:"volume_exclude"`
}

var ConfigDefaults = Config{
	VolumeInclude: types.RegExpAny,
	VolumeExclude: types.RegExpEmpty,
}

// A Collector is a Prometheus Collector for perflib logicalDisk metrics.
type Collector struct {
	config Config
	logger log.Logger

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
	filesystem   string
	serialNumber string
	label        string
	volumeType   string
	readonly     float64
}

func New(logger log.Logger, config *Config) *Collector {
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

	c.SetLogger(logger)

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
	).Default(c.config.VolumeExclude.String()).StringVar(&volumeExclude)

	app.Flag(
		"collector.logical_disk.volume-include",
		"Regexp of volumes to include. Volume name must both match include and not match exclude to be included.",
	).Default(c.config.VolumeInclude.String()).StringVar(&volumeInclude)

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

func (c *Collector) SetLogger(logger log.Logger) {
	c.logger = log.With(logger, "collector", Name)
}

func (c *Collector) GetPerfCounter() ([]string, error) {
	return []string{"LogicalDisk"}, nil
}

func (c *Collector) Close() error {
	return nil
}

func (c *Collector) Build() error {
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
func (c *Collector) Collect(ctx *types.ScrapeContext, ch chan<- prometheus.Metric) error {
	if err := c.collect(ctx, ch); err != nil {
		_ = level.Error(c.logger).Log("msg", "failed collecting logical_disk metrics", "err", err)
		return err
	}
	return nil
}

// Win32_PerfRawData_PerfDisk_LogicalDisk docs:
// - https://msdn.microsoft.com/en-us/windows/hardware/aa394307(v=vs.71) - Win32_PerfRawData_PerfDisk_LogicalDisk class
// - https://msdn.microsoft.com/en-us/library/ms803973.aspx - LogicalDisk object reference.
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

func (c *Collector) collect(ctx *types.ScrapeContext, ch chan<- prometheus.Metric) error {
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
			c.config.VolumeExclude.MatchString(volume.Name) ||
			!c.config.VolumeInclude.MatchString(volume.Name) {
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
			c.information,
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
			c.requestsQueued,
			prometheus.GaugeValue,
			volume.CurrentDiskQueueLength,
			volume.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.avgReadQueue,
			prometheus.GaugeValue,
			volume.AvgDiskReadQueueLength*perflib.TicksToSecondScaleFactor,
			volume.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.avgWriteQueue,
			prometheus.GaugeValue,
			volume.AvgDiskWriteQueueLength*perflib.TicksToSecondScaleFactor,
			volume.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.readBytesTotal,
			prometheus.CounterValue,
			volume.DiskReadBytesPerSec,
			volume.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.readsTotal,
			prometheus.CounterValue,
			volume.DiskReadsPerSec,
			volume.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.writeBytesTotal,
			prometheus.CounterValue,
			volume.DiskWriteBytesPerSec,
			volume.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.writesTotal,
			prometheus.CounterValue,
			volume.DiskWritesPerSec,
			volume.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.readTime,
			prometheus.CounterValue,
			volume.PercentDiskReadTime,
			volume.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.writeTime,
			prometheus.CounterValue,
			volume.PercentDiskWriteTime,
			volume.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.freeSpace,
			prometheus.GaugeValue,
			volume.PercentFreeSpace_Base*1024*1024,
			volume.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.totalSpace,
			prometheus.GaugeValue,
			volume.PercentFreeSpace*1024*1024,
			volume.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.idleTime,
			prometheus.CounterValue,
			volume.PercentIdleTime,
			volume.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.splitIOs,
			prometheus.CounterValue,
			volume.SplitIOPerSec,
			volume.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.readLatency,
			prometheus.CounterValue,
			volume.AvgDiskSecPerRead*perflib.TicksToSecondScaleFactor,
			volume.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.writeLatency,
			prometheus.CounterValue,
			volume.AvgDiskSecPerWrite*perflib.TicksToSecondScaleFactor,
			volume.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.readWriteLatency,
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

// diskExtentSize Size of the DiskExtent structure in bytes.
const diskExtentSize = 24

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
		return "", fmt.Errorf("could not identify physical drive for %s: %w", rootDrive, err)
	}

	numDiskIDs := uint(binary.LittleEndian.Uint32(volumeDiskExtents))
	if numDiskIDs < 1 {
		return "", fmt.Errorf("could not identify physical drive for %s: no disk IDs returned", rootDrive)
	}

	diskIDs := make([]string, numDiskIDs)

	for i := range numDiskIDs {
		diskIDs[i] = strconv.FormatUint(uint64(binary.LittleEndian.Uint32(volumeDiskExtents[8+i*diskExtentSize:])), 10)
	}

	slices.Sort(diskIDs)
	diskIDs = slices.Compact(diskIDs)

	return strings.Join(diskIDs, ";"), nil
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
