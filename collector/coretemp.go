package collector

import (
	"bytes"
	"fmt"
	"strings"
	"syscall"
	"unsafe"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
	"golang.org/x/sys/windows"
)

// technical documentation for the shared memory structure: https://www.alcpu.com/CoreTemp/developers.html

var (
	kernel32         = syscall.NewLazyDLL("KERNEL32.dll")
	msvcrt           = syscall.NewLazyDLL("msvcrt.dll")
	createMutexW     = kernel32.NewProc("CreateMutexW")
	releaseMutex     = kernel32.NewProc("ReleaseMutex")
	openFileMappingW = kernel32.NewProc("OpenFileMappingW")
	closeHandle      = kernel32.NewProc("CloseHandle")
	mapViewOfFile    = kernel32.NewProc("MapViewOfFile")
	unmapViewOfFile  = kernel32.NewProc("UnmapViewOfFile")
	memcpy_s         = msvcrt.NewProc("memcpy_s")
)

func init() {
	registerCollector("coretemp", NewCoreTempCollector)
}

// A coreTempCollector is a Prometheus collector for CoreTemp shared data metrics
type coreTempCollector struct {
	Temperature *prometheus.Desc
	Load        *prometheus.Desc
}

// NewCoreTempCollector ...
func NewCoreTempCollector() (Collector, error) {
	const subsystem = "coretemp"
	return &coreTempCollector{

		Temperature: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "temperature_celsius"),
			"(Temperature)",
			[]string{"name", "core"},
			nil,
		),

		Load: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "load"),
			"(Load)",
			[]string{"name", "core"},
			nil,
		),
	}, nil
}

// Collect sends the metric values for each metric
// to the provided prometheus Metric channel.
func (c *coreTempCollector) Collect(ctx *ScrapeContext, ch chan<- prometheus.Metric) error {
	if desc, err := c.collect(ch); err != nil {
		log.Error("failed collecting coretemp metrics:", desc, err)
		return err
	}
	return nil
}

func (c *coreTempCollector) collect(ch chan<- prometheus.Metric) (*prometheus.Desc, error) {

	s, err := readCoreTempSharedData()
	if err != nil {
		return nil, err
	}

	cpuName := s.GetCPUName()

	for i := uint32(0); i < s.CoreCount; i++ {

		// convert to celsius if internal unit is fahrenheit
		temp := float64(s.Temp[i])
		if s.IsFahrenheit {
			temp = (temp - 32) * 5 / 9
		}

		ch <- prometheus.MustNewConstMetric(
			c.Temperature,
			prometheus.GaugeValue,
			temp,
			cpuName,
			fmt.Sprintf("%d", i),
		)

		ch <- prometheus.MustNewConstMetric(
			c.Load,
			prometheus.GaugeValue,
			float64(s.Load[i]),
			cpuName,
			fmt.Sprintf("%d", i),
		)
	}

	return nil, nil
}

// read memory from CoreTemp to fetch cpu data
func readCoreTempSharedData() (*coreTempSharedData, error) {

	mutexName, _ := windows.UTF16PtrFromString("CoreTempMutexObject")
	mutexObject, _, err := createMutexW.Call(0, 0, uintptr(unsafe.Pointer(mutexName)))
	if err == nil {
		return nil, fmt.Errorf("CoreTempMutexObject not found. make sure Core Temp is running")
	}
	defer releaseMutex.Call(mutexObject)

	mappingName, _ := windows.UTF16PtrFromString("CoreTempMappingObject")
	mappingObject, _, err := openFileMappingW.Call(4, 1, uintptr(unsafe.Pointer(mappingName)))
	if mappingObject == uintptr(0) {
		return nil, err
	}
	defer closeHandle.Call(mappingObject)

	mapView, _, err := mapViewOfFile.Call(mappingObject, 4, 0, 0, 0)
	if mapView == uintptr(0) {
		return nil, err
	}
	defer unmapViewOfFile.Call(mapView)

	data := coreTempSharedData{}
	_, _, _ = memcpy_s.Call(uintptr(unsafe.Pointer(&data)), 0xa80, mapView, 0xa80)

	return &data, nil
}

type coreTempSharedData struct {
	Load           [256]uint32
	TjMax          [128]uint32
	CoreCount      uint32
	CPUCount       uint32
	Temp           [256]float32
	VID            float32
	CPUSpeed       float32
	FSBSpeed       float32
	Multipier      float32
	CPUName        [100]byte
	IsFahrenheit   bool // if true, true, the temperature is reported in Fahrenheit
	IsDeltaToTjMax bool // if true, the temperature reported represents the distance from TjMax
}

func (s *coreTempSharedData) GetCPUName() string {
	n := bytes.IndexByte(s.CPUName[:], 0)
	return strings.TrimSpace(string(s.CPUName[:n]))
}
