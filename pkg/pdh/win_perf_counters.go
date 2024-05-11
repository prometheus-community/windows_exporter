//go:build windows

package pdh

import (
	_ "embed"
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"golang.org/x/sys/windows"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
)

var defaultMaxBufferSize = 100 * 1024 * 1024

// Accumulator is a nested map structure used for storing metric data.
// The outermost map's keys are the object names, the middle map's keys are the counter names,
// and the innermost map's keys are the instance names. The float64 values represent the metric values.
//
// Example:
//
//	acc := Accumulator{
//	    "object1": {
//	        "counter1": {
//	            "instance1": 1.23,
//	            "instance2": 4.56,
//	        },
//	        "counter2": {
//	            "instance1": 7.89,
//	            "instance2": 0.12,
//	        },
//	    },
//	    "object2": {
//	        "counter1": {
//	            "instance1": 3.45,
//	            "instance2": 6.78,
//	        },
//	    },
//	}
type Accumulator map[string]map[string]map[string]float64

// CounterInfos is a nested map structure used for storing CounterInfo data.
// The outer map's keys are the object names, and the inner map's keys are the counter names.
// The CounterInfo values represent the detailed information about each counter.
//
// Example:
//
//	info := CounterInfos{
//	    "object1": {
//	        "counter1": CounterInfo{
//	            CounterType:  0x00010000,
//	            ObjectName:   "object1",
//	            InstanceName: "instance1",
//	            CounterName:  "counter1",
//	            FullPath:     "\\object1\\counter1",
//	            ExplainText:  "This is a sample counter",
//	        },
//	        "counter2": CounterInfo{
//	            CounterType:  0x00010100,
//	            ObjectName:   "object1",
//	            InstanceName: "instance2",
//	            CounterName:  "counter2",
//	            FullPath:     "\\object1\\counter2",
//	            ExplainText:  "This is another sample counter",
//	        },
//	    },
//	    "object2": {
//	        "counter1": CounterInfo{
//	            CounterType:  0x00012000,
//	            ObjectName:   "object2",
//	            InstanceName: "instance1",
//	            CounterName:  "counter1",
//	            FullPath:     "\\object2\\counter1",
//	            ExplainText:  "This is a sample counter for object2",
//	        },
//	    },
//	}
type CounterInfos map[string]map[string]CounterInfo

type WinPerfCounters struct {
	PrintValid                 bool
	PreVistaSupport            bool
	UsePerfCounterTime         bool
	Object                     []PerfObject
	UseWildcardsExpansion      bool
	LocalizeWildcardsExpansion bool
	IgnoredErrors              []string
	MaxBufferSize              int64
	Sources                    []string

	Log log.Logger

	lastRefreshed time.Time
	queryCreator  PerformanceQueryCreator
	hostCounters  map[string]*hostCountersInfo
	// cached os.Hostname()
	cachedHostname string
}

type hostCountersInfo struct {
	// computer name used as key and for printing
	computer string
	// computer name used in tag
	tag       string
	counters  []*counter
	query     PerformanceQuery
	timestamp time.Time
}

type CounterInfo struct {
	CounterType  uint32
	ObjectName   string
	InstanceName string
	CounterName  string
	FullPath     string
	ExplainText  string
}

type PerfObject struct {
	Sources       []string
	ObjectName    string   `yaml:"objectName"`
	Counters      []string `yaml:"counters"`
	Instances     []string `yaml:"instances"`
	Measurement   string
	WarnOnMissing bool
	FailOnMissing bool
	IncludeTotal  bool `yaml:"includeTotal"`
	UseRawValues  bool
}

type counter struct {
	counterPath   string
	computer      string
	objectName    string
	counter       string
	instance      string
	measurement   string
	includeTotal  bool
	useRawValue   bool
	counterHandle pdhCounterHandle
}

// extractCounterInfoFromCounterPath gets object name, instance name (if available) and counter name from counter path
// General Counter path pattern is: \\computer\object(parent/instance#index)\counter
// parent/instance#index part is skipped in single instance objects (e.g. Memory): \\computer\object\counter
//
//nolint:revive //function-result-limit conditionally 5 return results allowed
func extractCounterInfoFromCounterPath(counterPath string) (computer string, object string, instance string, counter string, err error) {
	leftComputerBorderIndex := -1
	rightObjectBorderIndex := -1
	leftObjectBorderIndex := -1
	leftCounterBorderIndex := -1
	rightInstanceBorderIndex := -1
	leftInstanceBorderIndex := -1
	var bracketLevel int

	for i := len(counterPath) - 1; i >= 0; i-- {
		switch counterPath[i] {
		case '\\':
			if bracketLevel == 0 {
				if leftCounterBorderIndex == -1 {
					leftCounterBorderIndex = i
				} else if leftObjectBorderIndex == -1 {
					leftObjectBorderIndex = i
				} else if leftComputerBorderIndex == -1 {
					leftComputerBorderIndex = i
				}
			}
		case '(':
			bracketLevel--
			if leftInstanceBorderIndex == -1 && bracketLevel == 0 && leftObjectBorderIndex == -1 && leftCounterBorderIndex > -1 {
				leftInstanceBorderIndex = i
				rightObjectBorderIndex = i
			}
		case ')':
			if rightInstanceBorderIndex == -1 && bracketLevel == 0 && leftCounterBorderIndex > -1 {
				rightInstanceBorderIndex = i
			}
			bracketLevel++
		}
	}
	if rightObjectBorderIndex == -1 {
		rightObjectBorderIndex = leftCounterBorderIndex
	}
	if rightObjectBorderIndex == -1 || leftObjectBorderIndex == -1 {
		return "", "", "", "", errors.New("cannot parse object from: " + counterPath)
	}

	if leftComputerBorderIndex > -1 {
		// validate there is leading \\ and not empty computer (\\\O)
		if leftComputerBorderIndex != 1 || leftComputerBorderIndex == leftObjectBorderIndex-1 {
			return "", "", "", "", errors.New("cannot parse computer from: " + counterPath)
		}
		computer = counterPath[leftComputerBorderIndex+1 : leftObjectBorderIndex]
	}

	if leftInstanceBorderIndex > -1 && rightInstanceBorderIndex > -1 {
		instance = counterPath[leftInstanceBorderIndex+1 : rightInstanceBorderIndex]
	} else if (leftInstanceBorderIndex == -1 && rightInstanceBorderIndex > -1) || (leftInstanceBorderIndex > -1 && rightInstanceBorderIndex == -1) {
		return "", "", "", "", errors.New("cannot parse instance from: " + counterPath)
	}
	object = counterPath[leftObjectBorderIndex+1 : rightObjectBorderIndex]
	counter = counterPath[leftCounterBorderIndex+1:]
	return computer, object, instance, counter, nil
}

func (m *WinPerfCounters) hostname() string {
	if m.cachedHostname != "" {
		return m.cachedHostname
	}
	hostname, err := os.Hostname()
	if err != nil {
		m.cachedHostname = "localhost"
	} else {
		m.cachedHostname = hostname
	}
	return m.cachedHostname
}

//nolint:revive //argument-limit conditionally more arguments allowed for helper function
func newCounter(
	counterHandle pdhCounterHandle,
	counterPath string,
	computer string,
	objectName string,
	instance string,
	counterName string,
	measurement string,
	includeTotal bool,
	useRawValue bool,
) *counter {
	return &counter{counterPath, computer, objectName, counterName, instance, measurement,
		includeTotal, useRawValue, counterHandle}
}

//nolint:revive //argument-limit conditionally more arguments allowed
func (m *WinPerfCounters) AddItem(counterPath, computer, objectName, instance, counterName, measurement string, includeTotal bool, useRawValue bool) error {
	origCounterPath := counterPath
	var err error
	var counterHandle pdhCounterHandle

	sourceTag := computer
	if computer == "localhost" {
		sourceTag = m.hostname()
	}
	if m.hostCounters == nil {
		m.hostCounters = make(map[string]*hostCountersInfo)
	}
	hostCounter, ok := m.hostCounters[computer]
	if !ok {
		hostCounter = &hostCountersInfo{computer: computer, tag: sourceTag}
		m.hostCounters[computer] = hostCounter
		hostCounter.query = m.queryCreator.NewPerformanceQuery(computer, uint32(m.MaxBufferSize))
		if err := hostCounter.query.Open(); err != nil {
			return err
		}
		hostCounter.counters = make([]*counter, 0)
	}

	if !hostCounter.query.IsVistaOrNewer() {
		counterHandle, err = hostCounter.query.AddCounterToQuery(counterPath)
		if err != nil {
			return err
		}
	} else {
		counterHandle, err = hostCounter.query.AddEnglishCounterToQuery(counterPath)
		if err != nil {
			return err
		}
	}

	if m.UseWildcardsExpansion {
		origInstance := instance
		counterPath, err = hostCounter.query.GetCounterPath(counterHandle)
		if err != nil {
			return err
		}
		counters, err := hostCounter.query.ExpandWildCardPath(counterPath)
		if err != nil {
			return err
		}

		_, origObjectName, _, origCounterName, err := extractCounterInfoFromCounterPath(origCounterPath)
		if err != nil {
			return err
		}

		for _, counterPath := range counters {
			_, err := hostCounter.query.AddCounterToQuery(counterPath)
			if err != nil {
				return err
			}

			computer, objectName, instance, counterName, err = extractCounterInfoFromCounterPath(counterPath)
			if err != nil {
				return err
			}

			var newItem *counter
			if !m.LocalizeWildcardsExpansion {
				// On localized installations of Windows, Telegraf
				// should return English metrics, but
				// ExpandWildCardPath returns localized counters. Undo
				// that by using the original object and counter
				// names, along with the expanded instance.

				var newInstance string
				if instance == "" {
					newInstance = emptyInstance
				} else {
					newInstance = instance
				}
				counterPath = formatPath(computer, origObjectName, newInstance, origCounterName)
				counterHandle, err = hostCounter.query.AddEnglishCounterToQuery(counterPath)
				if err != nil {
					return err
				}
				newItem = newCounter(
					counterHandle,
					counterPath,
					computer,
					origObjectName, instance,
					origCounterName,
					measurement,
					includeTotal,
					useRawValue,
				)
			} else {
				counterHandle, err = hostCounter.query.AddCounterToQuery(counterPath)
				if err != nil {
					return err
				}
				newItem = newCounter(
					counterHandle,
					counterPath,
					computer,
					objectName,
					instance,
					counterName,
					measurement,
					includeTotal,
					useRawValue,
				)
			}

			if instance == "_Total" && origInstance == "*" && !includeTotal {
				continue
			}

			hostCounter.counters = append(hostCounter.counters, newItem)
		}
	} else {
		newItem := newCounter(
			counterHandle,
			counterPath,
			computer,
			objectName,
			instance,
			counterName,
			measurement,
			includeTotal,
			useRawValue,
		)
		hostCounter.counters = append(hostCounter.counters, newItem)
	}

	return nil
}

const emptyInstance = "------"

func formatPath(computer, objectName, instance, counter string) string {
	path := ""
	if instance == emptyInstance {
		path = fmt.Sprintf(`\%s\%s`, objectName, counter)
	} else {
		path = fmt.Sprintf(`\%s(%s)\%s`, objectName, instance, counter)
	}
	if computer != "" && computer != "localhost" {
		path = fmt.Sprintf(`\\%s%s`, computer, path)
	}
	return path
}

func (m *WinPerfCounters) ParseConfig() error {
	var counterPath string
	m.queryCreator = &PerformanceQueryCreatorImpl{}

	if m.MaxBufferSize == 0 {
		m.MaxBufferSize = int64(defaultMaxBufferSize)
	}

	if len(m.Sources) == 0 {
		m.Sources = []string{"localhost"}
	}

	if len(m.Object) <= 0 {
		err := errors.New("no performance objects configured")
		return err
	}

	for _, PerfObject := range m.Object {
		computers := PerfObject.Sources
		if len(computers) == 0 {
			computers = m.Sources
		}
		for _, computer := range computers {
			if computer == "" {
				// localhost as a computer name in counter path doesn't work
				computer = "localhost"
			}
			for _, counter := range PerfObject.Counters {
				if len(PerfObject.Instances) == 0 {
					_ = level.Warn(m.Log).Log("msg", fmt.Sprintf("Missing 'Instances' param for object %q", PerfObject.ObjectName))
				}
				for _, instance := range PerfObject.Instances {
					objectName := PerfObject.ObjectName
					counterPath = formatPath(computer, objectName, instance, counter)

					err := m.AddItem(counterPath, computer, objectName, instance, counter,
						PerfObject.Measurement, PerfObject.IncludeTotal, PerfObject.UseRawValues)
					if err != nil {
						if PerfObject.FailOnMissing || PerfObject.WarnOnMissing {
							_ = level.Error(m.Log).Log("msg", fmt.Sprintf("Invalid counterPath %q: %s", counterPath, err.Error()))
						}
						if PerfObject.FailOnMissing {
							return err
						}
					}
				}
			}
		}
	}

	return nil
}

func (m *WinPerfCounters) checkError(err error) error {
	var pdhErr *PdhError
	if errors.As(err, &pdhErr) {
		for _, ignoredErrors := range m.IgnoredErrors {
			if PDHErrors[pdhErr.ErrorCode] == ignoredErrors {
				return nil
			}
		}

		return err
	}
	return err
}

func (m *WinPerfCounters) GetInfo() (map[string]CounterInfos, error) {
	var wg sync.WaitGroup
	info := map[string]CounterInfos{}

	// iterate over computers
	for _, hostCounterInfo := range m.hostCounters {
		wg.Add(1)
		go func(hostInfo *hostCountersInfo) {
			_ = level.Debug(m.Log).Log("msg", fmt.Sprintf("Gathering from %s", hostInfo.computer))

			info[hostInfo.computer] = make(CounterInfos)

			start := time.Now()
			for _, counter := range hostInfo.counters {
				counterInfo, err := hostInfo.query.GetCounterInfo(counter.counterHandle, 1)
				if err != nil {
					_ = level.Error(m.Log).Log("msg", fmt.Sprintf("error during collecting info on host %s", hostInfo.computer), "err", err)
				}

				objectName := windows.UTF16PtrToString(counterInfo.SzObjectName)
				counterName := windows.UTF16PtrToString(counterInfo.SzCounterName)

				if _, ok := info[hostInfo.computer][objectName]; !ok {
					info[hostInfo.computer][objectName] = make(map[string]CounterInfo)
				}
				info[hostInfo.computer][objectName][counterName] = CounterInfo{
					CounterType: counterInfo.DwType,
					ObjectName:  windows.UTF16PtrToString(counterInfo.SzObjectName),
					CounterName: windows.UTF16PtrToString(counterInfo.SzCounterName),
					FullPath:    windows.UTF16PtrToString(counterInfo.SzFullPath),
					ExplainText: windows.UTF16PtrToString(counterInfo.SzExplainText),
				}
			}

			_ = level.Debug(m.Log).Log("msg", fmt.Sprintf("Gathering info from %s finished in %v", hostInfo.computer, time.Since(start)))
			wg.Done()
		}(hostCounterInfo)
	}

	wg.Wait()
	return info, nil
}

func (m *WinPerfCounters) Init() error {
	if m.lastRefreshed.IsZero() {
		if err := m.cleanQueries(); err != nil {
			return err
		}

		if err := m.ParseConfig(); err != nil {
			return err
		}
		for _, hostCounterSet := range m.hostCounters {
			// some counters need two data samples before computing a value
			if err := hostCounterSet.query.CollectData(); err != nil {
				return m.checkError(err)
			}
		}
		m.lastRefreshed = time.Now()
	}

	return nil
}

func (m *WinPerfCounters) Gather() (map[string]Accumulator, error) {
	// Parse the config once
	var err error

	for _, hostCounterSet := range m.hostCounters {
		if m.UsePerfCounterTime && hostCounterSet.query.IsVistaOrNewer() {
			hostCounterSet.timestamp, err = hostCounterSet.query.CollectDataWithTime()
			if err != nil {
				return map[string]Accumulator{}, m.checkError(err)
			}
		} else {
			hostCounterSet.timestamp = time.Now()
			if err := hostCounterSet.query.CollectData(); err != nil {
				return map[string]Accumulator{}, m.checkError(err)
			}
		}
	}
	var wg sync.WaitGroup
	acc := map[string]Accumulator{}

	// iterate over computers
	for _, hostCounterInfo := range m.hostCounters {
		wg.Add(1)
		go func(hostInfo *hostCountersInfo) {
			var err error

			_ = level.Debug(m.Log).Log("msg", fmt.Sprintf("Gathering from %s", hostInfo.computer))
			start := time.Now()
			acc[hostInfo.computer], err = m.gatherComputerCounters(hostInfo)
			_ = level.Debug(m.Log).Log("msg", fmt.Sprintf("Gathering from %s finished in %v", hostInfo.computer, time.Since(start)))
			if err != nil {
				_ = level.Error(m.Log).Log("msg", fmt.Sprintf("error during collecting data on host %s", hostInfo.computer), "err", err)
			}
			wg.Done()
		}(hostCounterInfo)
	}

	wg.Wait()
	return acc, nil
}

func (m *WinPerfCounters) gatherComputerCounters(hostCounterInfo *hostCountersInfo) (Accumulator, error) {
	var value float64
	var err error

	acc := make(Accumulator)
	// For iterate over the known metrics and get the samples.
	for _, metric := range hostCounterInfo.counters {
		// collect
		if m.UseWildcardsExpansion {
			if metric.useRawValue {
				value, err = hostCounterInfo.query.GetRawCounterValue(metric.counterHandle)
			} else {
				value, err = hostCounterInfo.query.GetFormattedCounterValueDouble(metric.counterHandle)
			}
			if err != nil {
				// ignore invalid data  as some counters from process instances returns this sometimes
				if !isKnownCounterDataError(err) {
					return Accumulator{}, fmt.Errorf("error while getting value for counter %q: %w", metric.counterPath, err)
				}
				_ = level.Warn(m.Log).Log("msg", fmt.Sprintf("Error while getting value for counter %q, instance: %s, will skip metric: %v", metric.counterPath, metric.instance, err))
				continue
			}

			if !metric.includeTotal && strings.Contains(metric.instance, "_Total") {
				continue
			}

			if _, ok := acc[metric.objectName]; !ok {
				acc[metric.objectName] = make(map[string]map[string]float64)
			}

			if _, ok := acc[metric.objectName][metric.counter]; !ok {
				acc[metric.objectName][metric.counter] = make(map[string]float64)
			}

			acc[metric.objectName][metric.counter][metric.instance] = value
		} else {
			var counterValues []CounterValue
			if metric.useRawValue {
				counterValues, err = hostCounterInfo.query.GetRawCounterArray(metric.counterHandle)
			} else {
				counterValues, err = hostCounterInfo.query.GetFormattedCounterArrayDouble(metric.counterHandle)
			}
			if err != nil {
				// ignore invalid data  as some counters from process instances returns this sometimes
				if !isKnownCounterDataError(err) {
					return Accumulator{}, fmt.Errorf("error while getting value for counter %q: %w", metric.counterPath, err)
				}
				_ = level.Warn(m.Log).Log("msg", fmt.Sprintf("Error while getting value for counter %q, instance: %s, will skip metric: %v", metric.counterPath, metric.instance, err))
				continue
			}
			for _, cValue := range counterValues {
				if strings.Contains(metric.instance, "#") && strings.HasPrefix(metric.instance, cValue.InstanceName) {
					// If you are using a multiple instance identifier such as "w3wp#1"
					// phd.dll returns only the first 2 characters of the identifier.
					cValue.InstanceName = metric.instance
				}

				if !shouldIncludeMetric(metric, cValue) {
					continue
				}

				if _, ok := acc[metric.objectName]; !ok {
					acc[metric.objectName] = make(map[string]map[string]float64)
				}

				if _, ok := acc[metric.objectName][metric.counter]; !ok {
					acc[metric.objectName][metric.counter] = make(map[string]float64)
				}

				acc[metric.objectName][metric.counter][metric.instance] = value
			}
		}
	}

	return acc, nil
}

func (m *WinPerfCounters) cleanQueries() error {
	for _, hostCounterInfo := range m.hostCounters {
		if err := hostCounterInfo.query.Close(); err != nil {
			return err
		}
	}
	m.hostCounters = nil
	return nil
}

func shouldIncludeMetric(metric *counter, cValue CounterValue) bool {
	if metric.includeTotal {
		// If IncludeTotal is set, include all.
		return true
	}
	if metric.instance == "*" && !strings.Contains(cValue.InstanceName, "_Total") {
		// Catch if set to * and that it is not a '*_Total*' instance.
		return true
	}
	if metric.instance == cValue.InstanceName {
		// Catch if we set it to total or some form of it
		return true
	}
	if metric.instance == emptyInstance {
		return true
	}
	return false
}

func isKnownCounterDataError(err error) bool {
	var pdhErr *PdhError
	if errors.As(err, &pdhErr) && (pdhErr.ErrorCode == PdhInvalidData ||
		pdhErr.ErrorCode == PdhCalcNegativeDenominator ||
		pdhErr.ErrorCode == PdhCalcNegativeValue ||
		pdhErr.ErrorCode == PdhCstatusInvalidData ||
		pdhErr.ErrorCode == PdhCstatusNoInstance ||
		pdhErr.ErrorCode == PdhNoData) {
		return true
	}
	return false
}
