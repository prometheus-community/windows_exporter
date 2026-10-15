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
	"context"
	"errors"
	"fmt"
	"log/slog"
	"math"
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
	// query is the shared query of a collector from [NewQueryCollector], or
	// nil if the collector owns handle.
	query                 *Query
	object                string
	resultType            CounterType
	counters              []Counter
	counterIndex          map[string]int // index in counters by counter name
	handle                pdhQueryHandle
	totalCounterRequested bool
	mu                    sync.RWMutex
	logger                *slog.Logger

	rows rowAccessor[T]

	// partialRows keeps instances for which some counters have no valid value
	// in a sample and sets those values to NaN. Otherwise, such instances are
	// left out of the sample, so a missing value is never reported as zero.
	partialRows bool

	state *collectState
}

// Row holds the values of one instance collected by a collector from [NewDynamicCollector].
type Row struct {
	Name string
	// Values holds one value per counter, in the order the counters were passed to [NewDynamicCollector].
	// A value is NaN if the counter has no valid value for the instance in this sample.
	Values []float64
}

// rowAccessor initializes rows of type T and sets their names and counter values.
// field is the index the counterField of the counter was created with.
type rowAccessor[T any] struct {
	// resetRow initializes row for instance. row can hold the values of a
	// previous sample; resetRow can reuse its memory.
	resetRow func(row *T, instance string)
	setName  func(row *T, instance string)
	setValue func(row *T, field int, value float64)
}

// counterField maps a counter to the field index passed to rowAccessor.setValue.
type counterField struct {
	// tag is the counter name, optionally suffixed with ",secondvalue".
	tag string
	// index is the byte offset of the field in a struct row, or the index
	// in Row.Values.
	index      int
	minOSBuild string
}

type Counter struct {
	Name string
	Desc string
	// MetricType is the Prometheus metric type of the counter. It is zero
	// until the counter info of one of its instances has been read.
	MetricType prometheus.ValueType
	Instances  map[string]pdhCounterHandle
	Type       uint32
	Frequency  int64

	FieldIndexValue       int
	FieldIndexSecondValue int
}

// NewCollector creates a collector for the counters declared by the perfdata tags of the struct T.
//
// The optional field Name (string) of T receives the instance name. Occurrence
// suffixes distinguish duplicate names without colliding with literal names.
// These suffixes are not stable identities across samples.
//
// An instance is only collected if every counter has a valid value for it in that
// sample. Use [Collector.MetricType] for the metric type of a counter.
//
// If an error is returned together with a non-nil Collector, some counters
// could not be added or the initial collection failed. The caller owns that
// Collector and must Close it. If the Collector is nil, no resources are held.
func NewCollector[T any](logger *slog.Logger, resultType CounterType, object string, instances []string) (*Collector[T], error) {
	return newStructCollector[T](logger, nil, resultType, object, instances)
}

// NewQueryCollector creates a collector like [NewCollector] whose counters are
// part of query. [Query.Collect] samples them together with the other
// collectors of the query; [Collector.Collect] reads the latest sample. Close
// the collector before the query.
func NewQueryCollector[T any](logger *slog.Logger, query *Query, resultType CounterType, object string, instances []string) (*Collector[T], error) {
	if query == nil {
		return nil, ErrPerformanceCounterNotInitialized
	}

	return newStructCollector[T](logger, query, resultType, object, instances)
}

func newStructCollector[T any](logger *slog.Logger, query *Query, resultType CounterType, object string, instances []string) (*Collector[T], error) {
	fields, rows, errs, err := structFields[T]()
	if err != nil {
		return nil, err
	}

	return newCollector(logger, query, resultType, object, instances, fields, rows, false, errs)
}

// structFields returns the counter fields declared by the perfdata tags of
// the struct T and an accessor that writes them. The index of a counterField
// is the byte offset of its float64 field in T, so the accessor writes values
// without reflection.
func structFields[T any]() ([]counterField, rowAccessor[T], []error, error) {
	valueType := reflect.TypeFor[T]()
	if valueType.Kind() != reflect.Struct {
		return nil, rowAccessor[T]{}, nil, fmt.Errorf("expected a struct, got %s", valueType)
	}

	nameOffset := -1

	if f, ok := valueType.FieldByName("Name"); ok && f.Type.Kind() == reflect.String {
		if offset, ok := fieldOffset(valueType, f.Index); ok {
			nameOffset = offset
		}
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

		offset, ok := fieldOffset(valueType, f.Index)
		if !ok {
			errs = append(errs, fmt.Errorf("field %s must not be promoted through an embedded pointer", f.Name))

			continue
		}

		fields = append(fields, counterField{
			tag:        counterName,
			index:      offset,
			minOSBuild: f.Tag.Get("perfdata_min_build"),
		})
	}

	return fields, structRowAccessor[T](nameOffset), errs, nil
}

// fieldOffset returns the byte offset of the field with the index sequence
// index in the struct t. ok is false if the field is reached through an
// embedded pointer and is therefore not stored in t itself.
func fieldOffset(t reflect.Type, index []int) (int, bool) {
	var offset uintptr

	for _, i := range index {
		if t.Kind() != reflect.Struct {
			return 0, false
		}

		f := t.Field(i)
		offset += f.Offset
		t = f.Type
	}

	return int(offset), true
}

// structRowAccessor returns the accessor of the struct T. nameOffset is the
// byte offset of its string field Name, or -1. The offsets passed to setValue
// come from [structFields] and address float64 fields of T, so values are
// written without reflection.
func structRowAccessor[T any](nameOffset int) rowAccessor[T] {
	setName := func(row *T, instance string) {
		if nameOffset >= 0 {
			*(*string)(unsafe.Add(unsafe.Pointer(row), nameOffset)) = instance
		}
	}

	return rowAccessor[T]{
		resetRow: func(row *T, instance string) {
			var zero T

			*row = zero

			setName(row, instance)
		},
		setName: setName,
		setValue: func(row *T, offset int, value float64) {
			*(*float64)(unsafe.Add(unsafe.Pointer(row), offset)) = value
		},
	}
}

// NewDynamicCollector creates a collector for counters that are only known at runtime.
// Row.Values of the collected rows holds one value per entry of counters, in the same order.
// Unlike [NewCollector], an instance is kept if only some of its counters have
// a valid value; the others are NaN. Errors are returned as described for [NewCollector].
func NewDynamicCollector(logger *slog.Logger, resultType CounterType, object string, instances []string, counters []string) (*Collector[Row], error) {
	fields := make([]counterField, len(counters))
	for i, counter := range counters {
		fields[i] = counterField{tag: counter, index: i}
	}

	return newCollector(logger, nil, resultType, object, instances, fields, dynamicRowAccessor(len(counters)), true, nil)
}

// dynamicRowAccessor returns the accessor of rows with n values. A reused row
// keeps its Values array if it is large enough.
func dynamicRowAccessor(n int) rowAccessor[Row] {
	return rowAccessor[Row]{
		resetRow: func(row *Row, instance string) {
			row.Name = instance

			if cap(row.Values) < n {
				row.Values = make([]float64, n)

				return
			}

			row.Values = row.Values[:n]
			clear(row.Values)
		},
		setName: func(row *Row, instance string) {
			row.Name = instance
		},
		setValue: func(row *Row, field int, value float64) {
			row.Values[field] = value
		},
	}
}

func newCollector[T any](logger *slog.Logger, query *Query, resultType CounterType, object string, instances []string,
	fields []counterField, rows rowAccessor[T], partialRows bool, errs []error,
) (*Collector[T], error) {
	if resultType != CounterTypeRaw && resultType != CounterTypeFormatted {
		return nil, fmt.Errorf("invalid result type: %v", resultType)
	}

	var handle pdhQueryHandle

	if query != nil {
		// Adding counters while the query samples is not safe.
		query.mu.Lock()
		defer query.mu.Unlock()

		if query.handle == 0 {
			return nil, ErrPerformanceCounterNotInitialized
		}

		handle = query.handle
	} else if ret := OpenQuery(0, 0, &handle); ret != ErrorSuccess {
		return nil, NewPdhError(ret)
	}

	if len(instances) == 0 {
		instances = []string{InstanceEmpty}
	}

	collector := &Collector[T]{
		query:                 query,
		object:                object,
		resultType:            resultType,
		counters:              make([]Counter, 0, len(fields)),
		counterIndex:          make(map[string]int, len(fields)),
		handle:                handle,
		totalCounterRequested: slices.ContainsFunc(instances, func(instance string) bool { return IsTotalInstance(object, instance) }),
		mu:                    sync.RWMutex{},
		logger:                logger,
		rows:                  rows,
		partialRows:           partialRows,
	}

	for _, f := range fields {
		counterName, secondValue := strings.CutSuffix(f.tag, ",secondvalue")

		i, ok := collector.counterIndex[counterName]
		if !ok {
			i = len(collector.counters)
			collector.counterIndex[counterName] = i
			collector.counters = append(collector.counters, Counter{
				Name:                  counterName,
				Instances:             make(map[string]pdhCounterHandle, len(instances)),
				FieldIndexSecondValue: -1,
				FieldIndexValue:       -1,
			})
		}

		counter := &collector.counters[i]

		if secondValue {
			counter.FieldIndexSecondValue = f.index
		} else {
			counter.FieldIndexValue = f.index
		}

		if len(counter.Instances) != 0 {
			continue
		}

		var counterPath string

		for _, instance := range instances {
			counterPath = formatCounterPath(object, instance, counterName)

			var counterHandle pdhCounterHandle

			//nolint:nestif
			if ret := addEnglishCounter(handle, object, instance, counterName, &counterHandle); ret != ErrorSuccess {
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

			if counter.MetricType != 0 {
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
			counter.MetricType = metricType(resultType, counter.Type)

			switch counter.Type {
			case PERF_ELAPSED_TIME:
				if ret := GetCounterTimeBase(counterHandle, &counter.Frequency); ret != ErrorSuccess && ret != NoData {
					errs = append(errs, fmt.Errorf("GetCounterTimeBase: %w", NewPdhError(ret)))

					continue
				}
			case PERF_AVERAGE_TIMER:
				// Only [Collector.Frequency] uses the time base of an average
				// timer, so a missing one is not an error. Collect retries it.
				if ret := GetCounterTimeBase(counterHandle, &counter.Frequency); ret != ErrorSuccess {
					counter.Frequency = 0
				}
			}
		}
	}

	if len(collector.counters) == 0 {
		// A shared query stays open for its other collectors, and its lock
		// is held here, so only close a query of our own.
		if query == nil {
			collector.Close()
		}

		errs = append(errs, errors.New("no counters configured"))

		return nil, errors.Join(errs...)
	}

	if err := errors.Join(errs...); err != nil {
		return collector, fmt.Errorf("failed to initialize collector: %w", err)
	}

	collector.state = &collectState{}

	// Collect initial data because some counters need to be read twice to get the correct value.
	if query != nil {
		if ret := CollectQueryData(handle); ret != ErrorSuccess && ret != NoData {
			return collector, fmt.Errorf("failed to collect initial data: %w", NewPdhError(ret))
		}

		return collector, nil
	}

	var collectValues []T
	if err := collector.Collect(&collectValues); err != nil && !errors.Is(err, ErrNoData) {
		return collector, fmt.Errorf("failed to collect initial data: %w", err)
	}

	return collector, nil
}

// metricType returns the Prometheus metric type for a PDH counter type.
// Formatted values are already rates or ratios, so they are always gauges.
func metricType(resultType CounterType, counterType uint32) prometheus.ValueType {
	if resultType == CounterTypeRaw {
		if val, ok := SupportedCounterTypes[counterType]; ok {
			return val
		}
	}

	return prometheus.GaugeValue
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

// MetricType returns the Prometheus metric type of the counter named
// counterName, derived from its PDH counter type. Formatted values are always
// gauges. It returns false if the counter is unknown or none of its instances
// could be added.
func (c *Collector[T]) MetricType(counterName string) (prometheus.ValueType, bool) {
	if c == nil {
		return 0, false
	}

	c.mu.RLock()
	defer c.mu.RUnlock()

	i, ok := c.counterIndex[counterName]
	if !ok {
		return 0, false
	}

	metricType := c.counters[i].MetricType

	return metricType, metricType != 0
}

// Frequency returns the time base in ticks per second of the raw timer counter
// named counterName. The raw value of a PERF_AVERAGE_TIMER counter is the sum
// of the measured durations in these ticks; divide it by the frequency to get
// seconds. It returns false if the counter is unknown, is not such a timer or
// its time base is not known yet. [Collector.Collect] reads a missing time
// base again.
func (c *Collector[T]) Frequency(counterName string) (int64, bool) {
	if c == nil {
		return 0, false
	}

	c.mu.RLock()
	defer c.mu.RUnlock()

	i, ok := c.counterIndex[counterName]
	if !ok {
		return 0, false
	}

	frequency := c.counters[i].Frequency

	return frequency, frequency > 0
}

// Collect replaces the content of dst with one row per collected instance.
//
// Collect reuses the backing array of dst and, for [Row], the Values slices of
// its rows, so a sample does not allocate them again. Rows, sub-slices of dst
// and Row.Values are only valid until the next Collect into the same slice;
// copy what must be kept longer.
func (c *Collector[T]) Collect(dst *[]T) error {
	if c == nil {
		return ErrPerformanceCounterNotInitialized
	}

	if dst == nil {
		return errors.New("dst must not be nil")
	}

	if c.query != nil {
		// Reading the counter arrays while the query samples is not safe.
		c.query.mu.Lock()
		defer c.query.mu.Unlock()
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if len(c.counters) == 0 || c.handle == 0 || c.state == nil {
		return ErrPerformanceCounterNotInitialized
	}

	return c.collect(dst, c.state)
}

// collectState holds the buffers that the collector reuses between samples.
type collectState struct {
	// buf starts empty so that the first call only queries the required size with a nil buffer.
	// PdhGetRawCounterArrayW writes 8 bytes into a non-nil buffer, even if lpdwBufferSize is smaller,
	// which corrupts the neighboring heap memory of a tiny allocation.
	buf []byte

	// valid records for every row and counter whether the counter had a valid value.
	valid []bool

	names instanceNameCache

	// index, occurrences and seen back the maps of rowSet. They are cleared
	// for every sample, so their buckets are not allocated again.
	index       map[instanceKey]int
	occurrences map[string]int
	seen        map[string]int
}

// collect replaces the content of dst with one sample.
func (c *Collector[T]) collect(dst *[]T, state *collectState) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = c.panicError(r)
		}
	}()

	// A collector in a shared query reads the sample taken by Query.Collect.
	if c.query == nil {
		if ret := CollectQueryData(c.handle); ret != ErrorSuccess {
			return fmt.Errorf("failed to collect query data: %w", NewPdhError(ret))
		}
	}

	// Keep the backing array; rowSet initializes every row it reuses.
	previousLen := len(*dst)
	*dst = (*dst)[:0]

	defer func() {
		// Release the instance names of rows that this sample no longer uses.
		if n := len(*dst); n < previousLen {
			clear((*dst)[n:previousLen])
		}
	}()

	rows := newRowSet(c, dst, state)

	// Also after an error or panic, so the cache stays bounded while samples keep failing.
	// A name that the sample did not reach is decoded again by the next sample.
	defer state.names.removeUnseen()

	for counterIndex := range c.counters {
		c.loadTimeBase(&c.counters[counterIndex])

		for _, instance := range c.counters[counterIndex].Instances {
			itemCount, ok, err := c.getCounterArray(instance, &state.buf)
			if err != nil {
				return err
			}

			if !ok {
				continue
			}

			if c.resultType == CounterTypeRaw {
				rows.addRawItems(counterIndex, state.buf, itemCount)
			} else {
				rows.addFormattedItems(counterIndex, state.buf, itemCount)
			}
		}
	}

	rows.finish()

	state.valid = rows.valid

	if len(*dst) == 0 {
		return ErrNoData
	}

	return nil
}

// hasTimeBase reports whether the raw value of a counter of counterType is
// measured in ticks of the counter's time base.
func hasTimeBase(counterType uint32) bool {
	return counterType == PERF_ELAPSED_TIME || counterType == PERF_AVERAGE_TIMER
}

// loadTimeBase reads the time base of a raw timer counter that had none when
// the collector was created. Early after a reboot, PDH can report no data
// for it, e.g. for APP_POOL_WAS while WAS is still starting (#2313). Without a
// time base, an elapsed time counter has no valid value and every instance is
// omitted.
func (c *Collector[T]) loadTimeBase(counter *Counter) {
	if c.resultType != CounterTypeRaw || !hasTimeBase(counter.Type) || counter.Frequency > 0 {
		return
	}

	for _, instance := range counter.Instances {
		var frequency int64

		if ret := GetCounterTimeBase(instance, &frequency); ret == ErrorSuccess && frequency > 0 {
			counter.Frequency = frequency

			return
		}
	}
}

// pdhBufferSlack is extra space allocated beyond the buffer size reported by PDH.
// On Windows Server 2012 R2, PdhGetRawCounterArrayW reports the size of the item array of a single-instance
// object (e.g. System) without the terminator of the empty instance name, and writes that terminator right after
// the array: 48 bytes reported, 2 bytes written past them. Without the slack, the write corrupts the neighboring
// heap object and SzName points into it, which the GC reports as "found pointer to free object".
const pdhBufferSlack = 8

// getCounterArray reads the counter array of instance into buf and returns the number of items.
// ok is false if the counter has no data in this sample.
func (c *Collector[T]) getCounterArray(instance pdhCounterHandle, buf *[]byte) (uint32, bool, error) {
	getArray, name := GetRawCounterArray, "GetRawCounterArray"
	if c.resultType == CounterTypeFormatted {
		getArray, name = GetFormattedCounterArrayDouble, "GetFormattedCounterArrayDouble"
	}

	var itemCount uint32

	// Get the info with the current buffer size
	bytesNeeded := uint32(len(*buf))

	for {
		ret := getArray(instance, &bytesNeeded, &itemCount, unsafe.SliceData(*buf))

		if ret == ErrorSuccess {
			return itemCount, true, nil
		}

		// Check the status before creating an error: NewPdhError formats the
		// message text, which a growing buffer or a counter without data in
		// this sample does not need.
		if ret != MoreData {
			if isKnownCounterDataStatus(ret) {
				return 0, false, nil
			}

			return 0, false, fmt.Errorf("%s: %w", name, NewPdhError(ret))
		}

		if bytesNeeded <= uint32(len(*buf)) {
			return 0, false, fmt.Errorf("%s reports buffer too small (%d), but buffer is large enough (%d): %w", name, uint32(len(*buf)), bytesNeeded, NewPdhError(ret))
		}

		*buf = make([]byte, bytesNeeded+pdhBufferSlack)
	}
}

// setRawValue stores the value of a raw counter item in row. It returns false
// if the value cannot be computed.
func (c *Collector[T]) setRawValue(row *T, counter *Counter, value RawCounter) bool {
	// This is a workaround for the issue with the elapsed time counter type.
	// Source: https://github.com/prometheus-community/windows_exporter/pull/335/files#diff-d5d2528f559ba2648c2866aec34b1eaa5c094dedb52bd0ff22aa5eb83226bd8dR76-R83
	// Ref: https://learn.microsoft.com/en-us/windows/win32/perfctrs/calculating-counter-values
	switch counter.Type {
	case PERF_ELAPSED_TIME:
		// A zero frequency would divide by zero.
		if counter.Frequency <= 0 {
			return false
		}

		if counter.FieldIndexValue != -1 {
			c.rows.setValue(row, counter.FieldIndexValue, float64(value.SecondValue-value.FirstValue)/float64(counter.Frequency))
		}
	case PERF_100NSEC_TIMER, PERF_PRECISION_100NS_TIMER:
		if counter.FieldIndexValue != -1 {
			c.rows.setValue(row, counter.FieldIndexValue, float64(value.FirstValue)*TicksToSecondScaleFactor)
		}
	default:
		if counter.FieldIndexSecondValue != -1 {
			c.rows.setValue(row, counter.FieldIndexSecondValue, float64(value.SecondValue))
		}

		if counter.FieldIndexValue != -1 {
			c.rows.setValue(row, counter.FieldIndexValue, float64(value.FirstValue))
		}
	}

	return true
}

// instanceKey matches the nth occurrence of an instance name across counter arrays.
type instanceKey struct {
	name       string
	occurrence int
}

// rowSet appends one row per instance to dst and tracks which counters have
// a valid value for each row.
type rowSet[T any] struct {
	c         *Collector[T]
	dst       *[]T
	index     map[instanceKey]int
	nameCache *instanceNameCache
	// occurrences includes invalid items to reserve all literal names and keep occurrence positions.
	occurrences map[string]int
	// seen counts names in the current counter array and is cleared between arrays.
	seen map[string]int
	// valid has len(c.counters) entries per row.
	valid []bool
}

func newRowSet[T any](c *Collector[T], dst *[]T, state *collectState) rowSet[T] {
	state.names.startSample()

	if state.index == nil {
		state.index = map[instanceKey]int{}
		state.occurrences = map[string]int{}
		state.seen = map[string]int{}
	} else {
		// A failed sample can leave entries behind.
		clear(state.index)
		clear(state.occurrences)
		clear(state.seen)
	}

	return rowSet[T]{
		c:           c,
		dst:         dst,
		index:       state.index,
		nameCache:   &state.names,
		occurrences: state.occurrences,
		seen:        state.seen,
		valid:       state.valid[:0],
	}
}

// row returns the index of the row for the instance named by szName,
// appending the row if needed. ok is false if the item has no valid data or
// its instance is not collected.
func (r *rowSet[T]) row(counter *Counter, szName *uint16, status uint32) (int, bool) {
	instanceName := r.nameCache.decode(szName)
	if instanceName == "" || instanceName == "*" {
		instanceName = InstanceEmpty
	}

	// Count before checking status so invalid items do not shift subsequent duplicates.
	occurrence := r.seen[instanceName]
	r.seen[instanceName]++
	r.occurrences[instanceName] = max(r.occurrences[instanceName], occurrence+1)

	if status != CstatusValidData && status != CstatusNewData {
		// Check the level first, so the attributes are not built for every
		// invalid item while debug logging is off.
		if r.c.logger.Enabled(context.Background(), slog.LevelDebug) {
			r.c.logger.LogAttrs(context.Background(), slog.LevelDebug, "skipping counter item with invalid data status",
				slog.String("counter", counter.Name),
				slog.String("instance", instanceName),
				slog.Uint64("status", uint64(status)),
			)
		}

		return 0, false
	}

	if IsTotalInstance(r.c.object, instanceName) && !r.c.totalCounterRequested {
		return 0, false
	}

	key := instanceKey{name: instanceName, occurrence: occurrence}
	if index, ok := r.index[key]; ok {
		return index, true
	}

	rows := *r.dst
	index := len(rows)
	r.index[key] = index

	// Reuse a row of a previous sample within the capacity of dst.
	if index < cap(rows) {
		rows = rows[:index+1]
	} else {
		var zero T

		rows = append(rows, zero)
	}

	r.c.rows.resetRow(&rows[index], instanceName)
	*r.dst = rows

	n := len(r.c.counters)
	r.valid = slices.Grow(r.valid, n)[:len(r.valid)+n]
	clear(r.valid[len(r.valid)-n:])

	return index, true
}

func (r *rowSet[T]) setValid(row, counterIndex int) {
	r.valid[row*len(r.c.counters)+counterIndex] = true
}

// addRawItems stores the values of the raw counter items in buf.
func (r *rowSet[T]) addRawItems(counterIndex int, buf []byte, itemCount uint32) {
	counter := &r.c.counters[counterIndex]
	items := unsafe.Slice((*RawCounterItem)(unsafe.Pointer(unsafe.SliceData(buf))), itemCount)

	clear(r.seen)

	for _, item := range items {
		row, ok := r.row(counter, item.SzName, item.RawValue.CStatus)
		if ok && r.c.setRawValue(&(*r.dst)[row], counter, item.RawValue) {
			r.setValid(row, counterIndex)
		}
	}
}

// addFormattedItems stores the values of the formatted counter items in buf.
func (r *rowSet[T]) addFormattedItems(counterIndex int, buf []byte, itemCount uint32) {
	counter := &r.c.counters[counterIndex]
	items := unsafe.Slice((*FmtCounterValueItemDouble)(unsafe.Pointer(unsafe.SliceData(buf))), itemCount)

	clear(r.seen)

	for _, item := range items {
		row, ok := r.row(counter, item.SzName, item.FmtValue.CStatus)
		if !ok {
			continue
		}

		if counter.FieldIndexValue != -1 {
			r.c.rows.setValue(&(*r.dst)[row], counter.FieldIndexValue, item.FmtValue.DoubleValue)
		}

		r.setValid(row, counterIndex)
	}
}

// nameDuplicates assigns suffixes after all literal names in the sample are known.
func (r *rowSet[T]) nameDuplicates() {
	for name, count := range r.occurrences {
		suffix := 0

		for occurrence := 1; occurrence < count; occurrence++ {
			var instanceName string

			for {
				suffix++
				instanceName = name + "#" + strconv.Itoa(suffix)

				if _, literalName := r.occurrences[instanceName]; !literalName {
					break
				}
			}

			if row, ok := r.index[instanceKey{name: name, occurrence: occurrence}]; ok {
				r.c.rows.setName(&(*r.dst)[row], instanceName)
			}
		}
	}
}

// finish handles rows that lack a valid value for some counter. Counters
// without any instance, e.g. those skipped by perfdata_min_build, are not
// required. Without partial rows, such rows are removed. Otherwise, the
// missing values are set to NaN.
func (r *rowSet[T]) finish() {
	r.nameDuplicates()

	rows := *r.dst
	n := len(r.c.counters)
	kept := 0

	for row := range rows {
		complete := true

		for counterIndex := range r.c.counters {
			counter := &r.c.counters[counterIndex]

			if len(counter.Instances) == 0 || r.valid[row*n+counterIndex] {
				continue
			}

			complete = false

			if !r.c.partialRows {
				break
			}

			for _, field := range []int{counter.FieldIndexValue, counter.FieldIndexSecondValue} {
				if field != -1 {
					r.c.rows.setValue(&rows[row], field, math.NaN())
				}
			}
		}

		if !complete && !r.c.partialRows {
			continue
		}

		rows[kept] = rows[row]
		kept++
	}

	if dropped := len(rows) - kept; dropped > 0 {
		r.c.logger.LogAttrs(context.Background(), slog.LevelDebug, "omitting instances without a valid value for every counter",
			slog.String("object", r.c.object),
			slog.Int("instances", dropped),
		)

		clear(rows[kept:])
		*r.dst = rows[:kept]
	}
}

// panicError logs a panic recovered while collecting and returns it as an error.
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

	if c.query != nil {
		c.query.mu.Lock()
		defer c.query.mu.Unlock()
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	switch {
	case c.handle == 0:
	case c.query == nil:
		CloseQuery(c.handle)
	case c.query.handle != 0:
		// The query belongs to its other collectors; only remove the
		// counters of this one.
		for _, counter := range c.counters {
			for _, instance := range counter.Instances {
				RemoveCounter(instance)
			}
		}
	}

	c.handle = 0
	c.state = nil
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

// instanceNameCache maps the raw UTF-16 content of instance names to their decoded names.
// Names are cached by content and not by pointer, because the pointers point into the item
// buffer, which is overwritten by every Get*CounterArray call. With explicit instances, every
// call writes its single name at the same address.
//
// The cache is kept between samples, so a name is decoded once while its instance exists.
// Names that a sample did not decode are removed at its end, because instance names that
// contain process IDs, like those of GPU Engine, change all the time.
type instanceNameCache struct {
	names  map[string]*cachedInstanceName
	sample uint64
}

type cachedInstanceName struct {
	name string
	// sample is the last sample that contained the name.
	sample uint64
}

// startSample starts a new sample.
func (cache *instanceNameCache) startSample() {
	cache.sample++
}

// decode returns the instance name p points to.
func (cache *instanceNameCache) decode(p *uint16) string {
	if p == nil {
		return ""
	}

	n := 0
	for *(*uint16)(unsafe.Add(unsafe.Pointer(p), n*2)) != 0 {
		n++
	}

	raw := unsafe.Slice((*byte)(unsafe.Pointer(p)), n*2)

	// The compiler does not allocate for the string conversion in a map lookup.
	if cached, ok := cache.names[string(raw)]; ok {
		cached.sample = cache.sample

		return cached.name
	}

	if cache.names == nil {
		cache.names = map[string]*cachedInstanceName{}
	}

	name := windows.UTF16ToString(unsafe.Slice(p, n))
	cache.names[string(raw)] = &cachedInstanceName{name: name, sample: cache.sample}

	return name
}

// removeUnseen removes the names that the current sample did not decode.
func (cache *instanceNameCache) removeUnseen() {
	for raw, cached := range cache.names {
		if cached.sample != cache.sample {
			delete(cache.names, raw)
		}
	}
}

// isKnownCounterDataStatus reports whether status means that a counter has no
// valid data in a sample, which is not a collection error.
func isKnownCounterDataStatus(status uint32) bool {
	switch status {
	case InvalidData, CalcNegativeDenominator, CalcNegativeValue, CstatusInvalidData, CstatusNoInstance, NoData:
		return true
	default:
		return false
	}
}
