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

package pdh

import (
	"errors"
	"fmt"
	"log/slog"
	"reflect"
	"runtime/debug"
	"slices"
	"strconv"
	"strings"
	"sync"
	"unsafe"

	"github.com/prometheus-community/windows_exporter/internal/osversion"
	"github.com/prometheus/client_golang/prometheus"
	"golang.org/x/sys/windows"
)

//nolint:gochecknoglobals
var (
	InstancesAll   = []string{"*"}
	InstancesTotal = []string{InstanceTotal}
)

type CounterValues = map[string]map[string]CounterValue

type Collector[T any] struct {
	object                string
	counters              map[string]Counter
	handle                pdhQueryHandle
	totalCounterRequested bool
	mu                    sync.RWMutex
	logger                *slog.Logger

	rows rowAccessor[T]

	collectCh chan *[]T
	errorCh   chan error
}

// Row holds the values of one instance collected by a collector from [NewDynamicCollector].
type Row struct {
	Name       string
	MetricType prometheus.ValueType
	// Values holds one value per counter, in the order the counters were passed to [NewDynamicCollector].
	Values []float64
}

// rowAccessor creates rows of type T and sets their counter values.
// field is the index the counterField of the counter was created with.
type rowAccessor[T any] struct {
	newRow   func(instance string, metricType prometheus.ValueType) T
	setValue func(row *T, field int, value float64)
}

// counterField maps a counter to the field index passed to rowAccessor.setValue.
type counterField struct {
	// tag is the counter name, optionally suffixed with ",secondvalue".
	tag        string
	index      int
	minOSBuild string
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

// NewCollector creates a collector for the counters declared by the perfdata tags of the struct T.
//
// The optional fields Name (string) and MetricType (prometheus.ValueType) of T
// receive the instance name and the metric type of the counter.
//
// If an error is returned together with a non-nil Collector, some counters
// could not be added or the initial collection failed. The caller owns that
// Collector and must Close it. If the Collector is nil, no resources are held.
func NewCollector[T any](logger *slog.Logger, resultType CounterType, object string, instances []string) (*Collector[T], error) {
	valueType := reflect.TypeFor[T]()
	if valueType.Kind() != reflect.Struct {
		return nil, fmt.Errorf("expected a struct, got %s", valueType)
	}

	nameIndex, metricTypeIndex := -1, -1

	if f, ok := valueType.FieldByName("Name"); ok && f.Type.Kind() == reflect.String {
		nameIndex = f.Index[0]
	}

	if f, ok := valueType.FieldByName("MetricType"); ok && f.Type == reflect.TypeFor[prometheus.ValueType]() {
		metricTypeIndex = f.Index[0]
	}

	var errs []error

	fields := make([]counterField, 0, valueType.NumField())

	for _, f := range reflect.VisibleFields(valueType) {
		counterName, ok := f.Tag.Lookup("perfdata")
		if !ok {
			continue
		}

		if f.Type.Kind() != reflect.Float64 {
			errs = append(errs, fmt.Errorf("field %s must be a float64", f.Name))

			continue
		}

		fields = append(fields, counterField{
			tag:        counterName,
			index:      f.Index[0],
			minOSBuild: f.Tag.Get("perfdata_min_build"),
		})
	}

	rows := rowAccessor[T]{
		newRow: func(instance string, metricType prometheus.ValueType) T {
			var row T

			rv := reflect.ValueOf(&row).Elem()

			if nameIndex != -1 {
				rv.Field(nameIndex).SetString(instance)
			}

			if metricTypeIndex != -1 {
				rv.Field(metricTypeIndex).SetInt(int64(metricType))
			}

			return row
		},
		setValue: func(row *T, field int, value float64) {
			reflect.ValueOf(row).Elem().Field(field).SetFloat(value)
		},
	}

	return newCollector(logger, resultType, object, instances, fields, rows, errs)
}

// NewDynamicCollector creates a collector for counters that are only known at runtime.
// Row.Values of the collected rows holds one value per entry of counters, in the same order.
// Errors are returned as described for [NewCollector].
func NewDynamicCollector(logger *slog.Logger, resultType CounterType, object string, instances []string, counters []string) (*Collector[Row], error) {
	fields := make([]counterField, len(counters))
	for i, counter := range counters {
		fields[i] = counterField{tag: counter, index: i}
	}

	rows := rowAccessor[Row]{
		newRow: func(instance string, metricType prometheus.ValueType) Row {
			return Row{Name: instance, MetricType: metricType, Values: make([]float64, len(counters))}
		},
		setValue: func(row *Row, field int, value float64) {
			row.Values[field] = value
		},
	}

	return newCollector(logger, resultType, object, instances, fields, rows, nil)
}

func newCollector[T any](logger *slog.Logger, resultType CounterType, object string, instances []string,
	fields []counterField, rows rowAccessor[T], errs []error,
) (*Collector[T], error) {
	if resultType != CounterTypeRaw && resultType != CounterTypeFormatted {
		return nil, fmt.Errorf("invalid result type: %v", resultType)
	}

	var handle pdhQueryHandle

	if ret := OpenQuery(0, 0, &handle); ret != ErrorSuccess {
		return nil, NewPdhError(ret)
	}

	if len(instances) == 0 {
		instances = []string{InstanceEmpty}
	}

	collector := &Collector[T]{
		object:                object,
		counters:              make(map[string]Counter, len(fields)),
		handle:                handle,
		totalCounterRequested: slices.Contains(instances, InstanceTotal),
		mu:                    sync.RWMutex{},
		logger:                logger,
		rows:                  rows,
	}

	for _, f := range fields {
		counterName, secondValue := strings.CutSuffix(f.tag, ",secondvalue")

		counter, ok := collector.counters[counterName]
		if !ok {
			counter = Counter{
				Name:                  counterName,
				Instances:             make(map[string]pdhCounterHandle, len(instances)),
				FieldIndexSecondValue: -1,
				FieldIndexValue:       -1,
			}
		}

		if secondValue {
			counter.FieldIndexSecondValue = f.index
		} else {
			counter.FieldIndexValue = f.index
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
					if f.minOSBuild != "" {
						if minOSBuild, err := strconv.Atoi(f.minOSBuild); err == nil {
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
				if ret := GetCounterTimeBase(counterHandle, &counter.Frequency); ret != ErrorSuccess && ret != NoData {
					errs = append(errs, fmt.Errorf("GetCounterTimeBase: %w", NewPdhError(ret)))

					continue
				}
			}
		}

		collector.counters[counterName] = counter
	}

	if len(collector.counters) == 0 {
		collector.Close()

		errs = append(errs, errors.New("no counters configured"))

		return nil, errors.Join(errs...)
	}

	if err := errors.Join(errs...); err != nil {
		return collector, fmt.Errorf("failed to initialize collector: %w", err)
	}

	collector.collectCh = make(chan *[]T)
	collector.errorCh = make(chan error)

	if resultType == CounterTypeRaw {
		go collector.collectWorkerRaw()
	} else {
		go collector.collectWorkerFormatted()
	}

	// Collect initial data because some counters need to be read twice to get the correct value.
	var collectValues []T
	if err := collector.Collect(&collectValues); err != nil && !errors.Is(err, ErrNoData) {
		return collector, fmt.Errorf("failed to collect initial data: %w", err)
	}

	return collector, nil
}

func (c *Collector[T]) Describe() map[string]string {
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

// Collect replaces the content of dst with one row per collected instance.
func (c *Collector[T]) Collect(dst *[]T) error {
	if c == nil {
		return ErrPerformanceCounterNotInitialized
	}

	if dst == nil {
		return errors.New("dst must not be nil")
	}

	c.mu.RLock()
	defer c.mu.RUnlock()

	if len(c.counters) == 0 || c.handle == 0 || c.collectCh == nil || c.errorCh == nil {
		return ErrPerformanceCounterNotInitialized
	}

	c.collectCh <- dst

	return <-c.errorCh
}

func (c *Collector[T]) collectWorkerRaw() {
	var (
		err         error
		itemCount   uint32
		items       []RawCounterItem
		bytesNeeded uint32
	)

	// buf starts empty so that the first call only queries the required size with a nil buffer.
	// PdhGetRawCounterArrayW writes 8 bytes into a non-nil buffer, even if lpdwBufferSize is smaller,
	// which corrupts the neighboring heap memory of a tiny allocation.
	var buf []byte

	for dst := range c.collectCh {
		err = (func() (err error) {
			defer func() {
				if r := recover(); r != nil {
					err = c.panicError(r)
				}
			}()

			if ret := CollectQueryData(c.handle); ret != ErrorSuccess {
				return fmt.Errorf("failed to collect query data: %w", NewPdhError(ret))
			}

			*dst = (*dst)[:0:0]

			indexMap := map[string]int{}
			nameCache := map[string]string{}

			for _, counter := range c.counters {
			instances:
				for _, instance := range counter.Instances {
					// Get the info with the current buffer size
					bytesNeeded = uint32(len(buf))

					for {
						ret := GetRawCounterArray(instance, &bytesNeeded, &itemCount, unsafe.SliceData(buf))

						if ret == ErrorSuccess {
							break
						}

						if err := NewPdhError(ret); ret != MoreData {
							if isKnownCounterDataError(err) {
								continue instances
							}

							return fmt.Errorf("GetRawCounterArray: %w", err)
						}

						if bytesNeeded <= uint32(len(buf)) {
							return fmt.Errorf("GetRawCounterArray reports buffer too small (%d), but buffer is large enough (%d): %w", uint32(len(buf)), bytesNeeded, NewPdhError(ret))
						}

						buf = make([]byte, bytesNeeded)
					}

					items = unsafe.Slice((*RawCounterItem)(unsafe.Pointer(unsafe.SliceData(buf))), itemCount)

					for _, item := range items {
						if item.RawValue.CStatus != CstatusValidData && item.RawValue.CStatus != CstatusNewData {
							c.logger.Debug("skipping counter item with invalid data status",
								slog.String("counter", counter.Name),
								slog.String("instance", windows.UTF16PtrToString(item.SzName)),
								slog.Uint64("status", uint64(item.RawValue.CStatus)),
							)

							continue
						}

						instanceName := decodeInstanceName(nameCache, item.SzName)

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
							index = len(*dst)
							indexMap[instanceName] = index

							var metricsType prometheus.ValueType
							if metricsType, ok = SupportedCounterTypes[counter.Type]; !ok {
								metricsType = prometheus.GaugeValue
							}

							*dst = append(*dst, c.rows.newRow(instanceName, metricsType))
						}

						row := &(*dst)[index]

						// This is a workaround for the issue with the elapsed time counter type.
						// Source: https://github.com/prometheus-community/windows_exporter/pull/335/files#diff-d5d2528f559ba2648c2866aec34b1eaa5c094dedb52bd0ff22aa5eb83226bd8dR76-R83
						// Ref: https://learn.microsoft.com/en-us/windows/win32/perfctrs/calculating-counter-values
						switch counter.Type {
						case PERF_ELAPSED_TIME:
							// A zero frequency would divide by zero. The value is left unset.
							if counter.Frequency <= 0 || counter.FieldIndexValue == -1 {
								continue
							}

							c.rows.setValue(row, counter.FieldIndexValue, float64(item.RawValue.SecondValue-item.RawValue.FirstValue)/float64(counter.Frequency))
						case PERF_100NSEC_TIMER, PERF_PRECISION_100NS_TIMER:
							c.rows.setValue(row, counter.FieldIndexValue, float64(item.RawValue.FirstValue)*TicksToSecondScaleFactor)
						default:
							if counter.FieldIndexSecondValue != -1 {
								c.rows.setValue(row, counter.FieldIndexSecondValue, float64(item.RawValue.SecondValue))
							}

							if counter.FieldIndexValue != -1 {
								c.rows.setValue(row, counter.FieldIndexValue, float64(item.RawValue.FirstValue))
							}
						}
					}
				}
			}

			if len(*dst) == 0 {
				return ErrNoData
			}

			return nil
		})()

		c.errorCh <- err
	}
}

func (c *Collector[T]) collectWorkerFormatted() {
	var (
		err         error
		itemCount   uint32
		items       []FmtCounterValueItemDouble
		bytesNeeded uint32
	)

	// buf starts empty so that the first call only queries the required size with a nil buffer.
	// See collectWorkerRaw.
	var buf []byte

	for dst := range c.collectCh {
		err = (func() (err error) {
			defer func() {
				if r := recover(); r != nil {
					err = c.panicError(r)
				}
			}()

			if ret := CollectQueryData(c.handle); ret != ErrorSuccess {
				return fmt.Errorf("failed to collect query data: %w", NewPdhError(ret))
			}

			*dst = (*dst)[:0:0]

			indexMap := map[string]int{}
			nameCache := map[string]string{}

			for _, counter := range c.counters {
			instances:
				for _, instance := range counter.Instances {
					// Get the info with the current buffer size
					bytesNeeded = uint32(len(buf))

					for {
						ret := GetFormattedCounterArrayDouble(instance, &bytesNeeded, &itemCount, unsafe.SliceData(buf))

						if ret == ErrorSuccess {
							break
						}

						if err := NewPdhError(ret); ret != MoreData {
							if isKnownCounterDataError(err) {
								continue instances
							}

							return fmt.Errorf("GetFormattedCounterArrayDouble: %w", err)
						}

						if bytesNeeded <= uint32(len(buf)) {
							return fmt.Errorf("GetFormattedCounterArrayDouble reports buffer too small (%d), but buffer is large enough (%d): %w", uint32(len(buf)), bytesNeeded, NewPdhError(ret))
						}

						buf = make([]byte, bytesNeeded)
					}

					items = unsafe.Slice((*FmtCounterValueItemDouble)(unsafe.Pointer(unsafe.SliceData(buf))), itemCount)

					for _, item := range items {
						if item.FmtValue.CStatus != CstatusValidData && item.FmtValue.CStatus != CstatusNewData {
							continue
						}

						instanceName := decodeInstanceName(nameCache, item.SzName)

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
							index = len(*dst)
							indexMap[instanceName] = index

							*dst = append(*dst, c.rows.newRow(instanceName, prometheus.GaugeValue))
						}

						if counter.FieldIndexValue != -1 {
							c.rows.setValue(&(*dst)[index], counter.FieldIndexValue, item.FmtValue.DoubleValue)
						}
					}
				}
			}

			if len(*dst) == 0 {
				return ErrNoData
			}

			return nil
		})()

		c.errorCh <- err
	}
}

// panicError logs a panic recovered in the collect worker and returns it as an
// error. The worker goroutine has no other recovery, so a panic would end the process.
func (c *Collector[T]) panicError(r any) error {
	c.logger.Error("recovered from panic while collecting performance counters",
		slog.String("object", c.object),
		slog.Any("panic", r),
		slog.String("stack", string(debug.Stack())),
	)

	return fmt.Errorf("panic while collecting performance counters of %s: %v", c.object, r)
}

func (c *Collector[T]) Close() {
	if c == nil {
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if c.handle != 0 {
		CloseQuery(c.handle)
	}

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

// decodeInstanceName returns the instance name p points to.
// Decoded names are cached by their raw UTF-16 content and not by p, because p points
// into the item buffer, which is overwritten by every Get*CounterArray call. With explicit
// instances, every call writes its single name at the same address.
func decodeInstanceName(cache map[string]string, p *uint16) string {
	if p == nil {
		return ""
	}

	n := 0
	for *(*uint16)(unsafe.Add(unsafe.Pointer(p), n*2)) != 0 {
		n++
	}

	raw := unsafe.Slice((*byte)(unsafe.Pointer(p)), n*2)

	// The compiler does not allocate for the string conversion in a map lookup.
	if name, ok := cache[string(raw)]; ok {
		return name
	}

	name := windows.UTF16ToString(unsafe.Slice(p, n))
	cache[string(raw)] = name

	return name
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
