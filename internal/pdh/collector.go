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

package pdh

import (
	"errors"
	"fmt"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"sync"
	"unsafe"

	"github.com/Microsoft/hcsshim/osversion"
	"github.com/prometheus-community/windows_exporter/internal/mi"
	"github.com/prometheus/client_golang/prometheus"
	"golang.org/x/sys/windows"
)

//nolint:gochecknoglobals
var (
	InstancesAll   = []string{"*"}
	InstancesTotal = []string{InstanceTotal}
)

type CounterValues = map[string]map[string]CounterValue

type Collector struct {
	object                string
	counters              map[string]Counter
	handle                pdhQueryHandle
	totalCounterRequested bool
	mu                    sync.RWMutex

	nameIndexValue        int
	metricsTypeIndexValue int

	collectCh chan any
	errorCh   chan error
}

type Counter struct {
	Name       string
	Desc       string
	MetricType prometheus.ValueType
	Instances  map[string]pdhCounterHandle
	Type       uint32
	Frequency  int64

	FieldIndexValue       int
	FieldIndexSecondValue int
}

func NewCollector[T any](resultType CounterType, object string, instances []string) (*Collector, error) {
	valueType := reflect.TypeFor[T]()

	return NewCollectorWithReflection(resultType, object, instances, valueType)
}

func NewCollectorWithReflection(resultType CounterType, object string, instances []string, valueType reflect.Type) (*Collector, error) {
	var handle pdhQueryHandle

	if ret := OpenQuery(0, 0, &handle); ret != ErrorSuccess {
		return nil, NewPdhError(ret)
	}

	if len(instances) == 0 {
		instances = []string{InstanceEmpty}
	}

	if resultType != CounterTypeRaw && resultType != CounterTypeFormatted {
		return nil, fmt.Errorf("invalid result type: %v", resultType)
	}

	collector := &Collector{
		object:                object,
		counters:              make(map[string]Counter, valueType.NumField()),
		handle:                handle,
		totalCounterRequested: slices.Contains(instances, InstanceTotal),
		mu:                    sync.RWMutex{},
		nameIndexValue:        -1,
		metricsTypeIndexValue: -1,
	}

	errs := make([]error, 0, valueType.NumField())

	if f, ok := valueType.FieldByName("Name"); ok {
		if f.Type.Kind() == reflect.String {
			collector.nameIndexValue = f.Index[0]
		}
	}

	if f, ok := valueType.FieldByName("MetricType"); ok {
		if f.Type.Kind() == reflect.TypeOf(prometheus.ValueType(0)).Kind() {
			collector.metricsTypeIndexValue = f.Index[0]
		}
	}

	for _, f := range reflect.VisibleFields(valueType) {
		counterName, ok := f.Tag.Lookup("perfdata")
		if !ok {
			continue
		}

		if f.Type.Kind() != reflect.Float64 {
			errs = append(errs, fmt.Errorf("field %s must be a float64", f.Name))

			continue
		}

		secondValue := strings.HasSuffix(counterName, ",secondvalue")
		if secondValue {
			counterName = strings.TrimSuffix(counterName, ",secondvalue")
		}

		var counter Counter
		if counter, ok = collector.counters[counterName]; !ok {
			counter = Counter{
				Name:                  counterName,
				Instances:             make(map[string]pdhCounterHandle, len(instances)),
				FieldIndexSecondValue: -1,
				FieldIndexValue:       -1,
			}
		}

		if secondValue {
			counter.FieldIndexSecondValue = f.Index[0]
		} else {
			counter.FieldIndexValue = f.Index[0]
		}

		if len(counter.Instances) != 0 {
			collector.counters[counterName] = counter

			continue
		}

		var counterPath string

		for _, instance := range instances {
			counterPath = formatCounterPath(object, instance, counterName)

			var counterHandle pdhCounterHandle

			//nolint:nestif
			if ret := AddEnglishCounter(handle, counterPath, 0, &counterHandle); ret != ErrorSuccess {
				if ret == CstatusNoCounter {
					if minOSBuildTag, ok := f.Tag.Lookup("perfdata_min_build"); ok {
						if minOSBuild, err := strconv.Atoi(minOSBuildTag); err == nil {
							if uint16(minOSBuild) > osversion.Build() {
								continue
							}
						}
					}
				}

				errs = append(errs, fmt.Errorf("failed to add counter %s: %w", counterPath, NewPdhError(ret)))

				continue
			}

			counter.Instances[instance] = counterHandle

			if counter.Type != 0 {
				continue
			}

			// Get the info with the current buffer size
			var bufLen uint32

			if ret := GetCounterInfo(counterHandle, 0, &bufLen, nil); ret != MoreData {
				errs = append(errs, fmt.Errorf("GetCounterInfo: %w", NewPdhError(ret)))

				continue
			}

			buf := make([]byte, bufLen)
			if len(buf) == 0 {
				errs = append(errs, errors.New("GetCounterInfo: buffer length is zero"))

				continue
			}

			if ret := GetCounterInfo(counterHandle, 0, &bufLen, &buf[0]); ret != ErrorSuccess {
				errs = append(errs, fmt.Errorf("GetCounterInfo: %w", NewPdhError(ret)))

				continue
			}

			counterInfo := (*CounterInfo)(unsafe.Pointer(&buf[0]))
			if counterInfo == nil {
				errs = append(errs, errors.New("GetCounterInfo: counter info is nil"))

				continue
			}

			counter.Type = counterInfo.DwType
			if val, ok := SupportedCounterTypes[counter.Type]; ok {
				counter.MetricType = val
			} else {
				counter.MetricType = prometheus.GaugeValue
			}

			if counter.Type == PERF_ELAPSED_TIME {
				if ret := GetCounterTimeBase(counterHandle, &counter.Frequency); ret != ErrorSuccess {
					errs = append(errs, fmt.Errorf("GetCounterTimeBase: %w", NewPdhError(ret)))

					continue
				}
			}
		}

		collector.counters[counterName] = counter
	}

	if err := errors.Join(errs...); err != nil {
		return collector, fmt.Errorf("failed to initialize collector: %w", err)
	}

	if len(collector.counters) == 0 {
		return nil, errors.New("no counters configured")
	}

	collector.collectCh = make(chan any)
	collector.errorCh = make(chan error)

	if resultType == CounterTypeRaw {
		go collector.collectWorkerRaw()
	} else {
		go collector.collectWorkerFormatted()
	}

	// Collect initial data because some counters need to be read twice to get the correct value.
	collectValues := reflect.New(reflect.SliceOf(valueType)).Elem()
	if err := collector.Collect(collectValues.Addr().Interface()); err != nil && !errors.Is(err, ErrNoData) {
		return collector, fmt.Errorf("failed to collect initial data: %w", err)
	}

	return collector, nil
}

func (c *Collector) Describe() map[string]string {
	if c == nil {
		return map[string]string{}
	}

	c.mu.RLock()
	defer c.mu.RUnlock()

	desc := make(map[string]string, len(c.counters))

	for _, counter := range c.counters {
		desc[counter.Name] = counter.Desc
	}

	return desc
}

func (c *Collector) Collect(dst any) error {
	if c == nil {
		return ErrPerformanceCounterNotInitialized
	}

	c.mu.RLock()
	defer c.mu.RUnlock()

	if len(c.counters) == 0 || c.handle == 0 || c.collectCh == nil || c.errorCh == nil {
		return ErrPerformanceCounterNotInitialized
	}

	c.collectCh <- dst

	return <-c.errorCh
}

func (c *Collector) collectWorkerRaw() {
	var (
		err         error
		itemCount   uint32
		items       []RawCounterItem
		bytesNeeded uint32
	)

	buf := make([]byte, 1)

	for data := range c.collectCh {
		err = (func() error {
			if ret := CollectQueryData(c.handle); ret != ErrorSuccess {
				return fmt.Errorf("failed to collect query data: %w", NewPdhError(ret))
			}

			dv := reflect.ValueOf(data)
			if dv.Kind() != reflect.Ptr || dv.IsNil() {
				return fmt.Errorf("expected a pointer, got %s: %w", dv.Kind(), mi.ErrInvalidEntityType)
			}

			dv = dv.Elem()

			if dv.Kind() != reflect.Slice {
				return fmt.Errorf("expected a pointer to a slice, got %s: %w", dv.Kind(), mi.ErrInvalidEntityType)
			}

			elemType := dv.Type().Elem()

			if elemType.Kind() != reflect.Struct {
				return fmt.Errorf("expected a pointer to a slice of structs, got a slice of %s: %w", elemType.Kind(), mi.ErrInvalidEntityType)
			}

			if dv.Len() != 0 {
				dv.Set(reflect.MakeSlice(dv.Type(), 0, 0))
			}

			dv.Clear()

			elemValue := reflect.ValueOf(reflect.New(elemType).Interface()).Elem()

			indexMap := map[string]int{}
			stringMap := map[*uint16]string{}

			for _, counter := range c.counters {
				for _, instance := range counter.Instances {
					// Get the info with the current buffer size
					bytesNeeded = uint32(cap(buf))

					for {
						ret := GetRawCounterArray(instance, &bytesNeeded, &itemCount, &buf[0])

						if ret == ErrorSuccess {
							break
						}

						if err := NewPdhError(ret); ret != MoreData {
							if isKnownCounterDataError(err) {
								break
							}

							return fmt.Errorf("GetRawCounterArray: %w", err)
						}

						if bytesNeeded <= uint32(cap(buf)) {
							return fmt.Errorf("GetRawCounterArray reports buffer too small (%d), but buffer is large enough (%d): %w", uint32(cap(buf)), bytesNeeded, NewPdhError(ret))
						}

						buf = make([]byte, bytesNeeded)
					}

					items = unsafe.Slice((*RawCounterItem)(unsafe.Pointer(&buf[0])), itemCount)

					var (
						instanceName string
						ok           bool
					)

					for _, item := range items {
						if item.RawValue.CStatus != CstatusValidData && item.RawValue.CStatus != CstatusNewData {
							continue
						}

						if instanceName, ok = stringMap[item.SzName]; !ok {
							instanceName = windows.UTF16PtrToString(item.SzName)
							stringMap[item.SzName] = instanceName
						}

						if strings.HasSuffix(instanceName, InstanceTotal) && !c.totalCounterRequested {
							continue
						}

						if instanceName == "" || instanceName == "*" {
							instanceName = InstanceEmpty
						}

						var (
							index int
							ok    bool
						)

						if index, ok = indexMap[instanceName]; !ok {
							index = dv.Len()
							indexMap[instanceName] = index

							if c.nameIndexValue != -1 {
								elemValue.Field(c.nameIndexValue).SetString(instanceName)
							}

							if c.metricsTypeIndexValue != -1 {
								var metricsType prometheus.ValueType
								if metricsType, ok = SupportedCounterTypes[counter.Type]; !ok {
									metricsType = prometheus.GaugeValue
								}

								elemValue.Field(c.metricsTypeIndexValue).Set(reflect.ValueOf(metricsType))
							}

							dv.Set(reflect.Append(dv, elemValue))
						}

						// This is a workaround for the issue with the elapsed time counter type.
						// Source: https://github.com/prometheus-community/windows_exporter/pull/335/files#diff-d5d2528f559ba2648c2866aec34b1eaa5c094dedb52bd0ff22aa5eb83226bd8dR76-R83
						// Ref: https://learn.microsoft.com/en-us/windows/win32/perfctrs/calculating-counter-values
						switch counter.Type {
						case PERF_ELAPSED_TIME:
							dv.Index(index).
								Field(counter.FieldIndexValue).
								SetFloat(float64((item.RawValue.SecondValue - item.RawValue.FirstValue) / counter.Frequency))
						case PERF_100NSEC_TIMER, PERF_PRECISION_100NS_TIMER:
							dv.Index(index).
								Field(counter.FieldIndexValue).
								SetFloat(float64(item.RawValue.FirstValue) * TicksToSecondScaleFactor)
						default:
							if counter.FieldIndexSecondValue != -1 {
								dv.Index(index).
									Field(counter.FieldIndexSecondValue).
									SetFloat(float64(item.RawValue.SecondValue))
							}

							if counter.FieldIndexValue != -1 {
								dv.Index(index).
									Field(counter.FieldIndexValue).
									SetFloat(float64(item.RawValue.FirstValue))
							}
						}
					}
				}
			}

			if dv.Len() == 0 {
				return ErrNoData
			}

			return nil
		})()

		c.errorCh <- err
	}
}

func (c *Collector) collectWorkerFormatted() {
	var (
		err         error
		itemCount   uint32
		items       []FmtCounterValueItemDouble
		bytesNeeded uint32
	)

	buf := make([]byte, 1)

	for data := range c.collectCh {
		err = (func() error {
			if ret := CollectQueryData(c.handle); ret != ErrorSuccess {
				return fmt.Errorf("failed to collect query data: %w", NewPdhError(ret))
			}

			dv := reflect.ValueOf(data)
			if dv.Kind() != reflect.Ptr || dv.IsNil() {
				return fmt.Errorf("expected a pointer, got %s: %w", dv.Kind(), mi.ErrInvalidEntityType)
			}

			dv = dv.Elem()

			if dv.Kind() != reflect.Slice {
				return fmt.Errorf("expected a pointer to a slice, got %s: %w", dv.Kind(), mi.ErrInvalidEntityType)
			}

			elemType := dv.Type().Elem()

			if elemType.Kind() != reflect.Struct {
				return fmt.Errorf("expected a pointer to a slice of structs, got a slice of %s: %w", elemType.Kind(), mi.ErrInvalidEntityType)
			}

			if dv.Len() != 0 {
				dv.Set(reflect.MakeSlice(dv.Type(), 0, 0))
			}

			dv.Clear()

			elemValue := reflect.ValueOf(reflect.New(elemType).Interface()).Elem()

			indexMap := map[string]int{}
			stringMap := map[*uint16]string{}

			for _, counter := range c.counters {
				for _, instance := range counter.Instances {
					// Get the info with the current buffer size
					bytesNeeded = uint32(cap(buf))

					for {
						ret := GetFormattedCounterArrayDouble(instance, &bytesNeeded, &itemCount, &buf[0])

						if ret == ErrorSuccess {
							break
						}

						if err := NewPdhError(ret); ret != MoreData {
							if isKnownCounterDataError(err) {
								break
							}

							return fmt.Errorf("GetFormattedCounterArrayDouble: %w", err)
						}

						if bytesNeeded <= uint32(cap(buf)) {
							return fmt.Errorf("GetFormattedCounterArrayDouble reports buffer too small (%d), but buffer is large enough (%d): %w", uint32(cap(buf)), bytesNeeded, NewPdhError(ret))
						}

						buf = make([]byte, bytesNeeded)
					}

					items = unsafe.Slice((*FmtCounterValueItemDouble)(unsafe.Pointer(&buf[0])), itemCount)

					var (
						instanceName string
						ok           bool
					)

					for _, item := range items {
						if item.FmtValue.CStatus != CstatusValidData && item.FmtValue.CStatus != CstatusNewData {
							continue
						}

						if instanceName, ok = stringMap[item.SzName]; !ok {
							instanceName = windows.UTF16PtrToString(item.SzName)
							stringMap[item.SzName] = instanceName
						}

						if strings.HasSuffix(instanceName, InstanceTotal) && !c.totalCounterRequested {
							continue
						}

						if instanceName == "" || instanceName == "*" {
							instanceName = InstanceEmpty
						}

						var (
							index int
							ok    bool
						)

						if index, ok = indexMap[instanceName]; !ok {
							index = dv.Len()
							indexMap[instanceName] = index

							if c.nameIndexValue != -1 {
								elemValue.Field(c.nameIndexValue).SetString(instanceName)
							}

							if c.metricsTypeIndexValue != -1 {
								elemValue.Field(c.metricsTypeIndexValue).Set(reflect.ValueOf(prometheus.GaugeValue))
							}

							dv.Set(reflect.Append(dv, elemValue))
						}

						if counter.FieldIndexValue != -1 {
							dv.Index(index).
								Field(counter.FieldIndexValue).
								SetFloat(item.FmtValue.DoubleValue)
						}
					}
				}
			}

			if dv.Len() == 0 {
				return ErrNoData
			}

			return nil
		})()

		c.errorCh <- err
	}
}

func (c *Collector) Close() {
	if c == nil {
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	CloseQuery(c.handle)

	c.handle = 0

	if c.collectCh != nil {
		close(c.collectCh)
	}

	if c.errorCh != nil {
		close(c.errorCh)
	}

	c.collectCh = nil
	c.errorCh = nil
}

func formatCounterPath(object, instance, counterName string) string {
	var counterPath string

	if instance == InstanceEmpty {
		counterPath = fmt.Sprintf(`\%s\%s`, object, counterName)
	} else {
		counterPath = fmt.Sprintf(`\%s(%s)\%s`, object, instance, counterName)
	}

	return counterPath
}

func isKnownCounterDataError(err error) bool {
	var pdhErr *Error

	return errors.As(err, &pdhErr) && (pdhErr.ErrorCode == InvalidData ||
		pdhErr.ErrorCode == CalcNegativeDenominator ||
		pdhErr.ErrorCode == CalcNegativeValue ||
		pdhErr.ErrorCode == CstatusInvalidData ||
		pdhErr.ErrorCode == CstatusNoInstance ||
		pdhErr.ErrorCode == NoData)
}
