//go:build windows
// +build windows

package collector

import (
	"errors"
	"time"

	"github.com/StackExchange/wmi"
	"github.com/prometheus-community/windows_exporter/log"
	"github.com/prometheus/client_golang/prometheus"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	idlePeriodFlag = kingpin.Flag(
		"collector.idle.period",
		"Period of time in seconds of inactivity to consider for idle detection",
	).Default("480").Int()
)

func init() {
	registerCollector("idle", newIdleCollector)
}

const (
	idleQuery = "SELECT Handle, ReadOperationCount FROM Win32_Process where name=\"csrss.exe\""
)

type IdleInfoCollector struct {
	IdleFlag *prometheus.Desc

	isIdle          byte
	idlePeriod      int
	opCount         map[string]uint64 // a map between csrss.exe Handle and ReadOperationCount
	lastChangedTime int64
}

func newIdleCollector() (Collector, error) {
	return &IdleInfoCollector{
		IdleFlag: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, "", "idle"),
			"Idle status",
			[]string{},
			nil,
		),
		isIdle:          1,
		idlePeriod:      *idlePeriodFlag,
		opCount:         make(map[string]uint64),
		lastChangedTime: time.Now().UnixMilli(),
	}, nil
}

// Collect sends the metric values for each metric
// to the provided prometheus Metric channel.
func (c *IdleInfoCollector) Collect(ctx *ScrapeContext, ch chan<- prometheus.Metric) error {
	if desc, err := c.collect(ch); err != nil {
		log.Error("failed collecting cpu_info metrics:", desc, err)
		return err
	}
	return nil
}

type win32_Process struct {
	Handle             string
	ReadOperationCount uint64
}

func (c *IdleInfoCollector) collect(ch chan<- prometheus.Metric) (*prometheus.Desc, error) {
	var dst []win32_Process

	if err := wmi.Query(idleQuery, &dst); err != nil {
		return nil, err
	}
	if len(dst) == 0 {
		return nil, errors.New("WMI query returned empty result set")
	}

	cTime := time.Now().UnixMilli()
	dataChanged := false
	for _, processor := range dst {
		if opCount, found := c.opCount[processor.Handle]; found {
			if opCount != processor.ReadOperationCount {
				c.opCount[processor.Handle] = processor.ReadOperationCount
				dataChanged = true
			} // else no change
		} else {
			// new handle
			c.opCount[processor.Handle] = processor.ReadOperationCount
			dataChanged = true
		}
	}
	// need to check if we have less handles now
	for handle, _ := range c.opCount {
		found := false
		for _, processor := range dst {
			if handle == processor.Handle {
				found = true
			}
		}
		if !found {
			delete(c.opCount, handle)
			dataChanged = true
		}
	}
	if dataChanged {
		c.isIdle = 0
		c.lastChangedTime = cTime
	} else {
		if cTime-c.lastChangedTime > int64(c.idlePeriod)*1000 {
			c.isIdle = 1
		}
	}

	ch <- prometheus.MustNewConstMetric(
		c.IdleFlag,
		prometheus.GaugeValue,
		float64(c.isIdle),
	)

	return nil, nil
}
