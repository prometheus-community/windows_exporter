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

package perfdata

import (
	"errors"
	"fmt"
	"slices"
	"strings"
	"sync"
	"unsafe"

	"github.com/prometheus/client_golang/prometheus"
	"golang.org/x/sys/windows"
)

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

	collectCh       chan struct{}
	counterValuesCh chan CounterValues
	errorCh         chan error
}

type Counter struct {
	Name      string
	Desc      string
	Instances map[string]pdhCounterHandle
	Type      uint32
	Frequency int64
}

func NewCollector(object string, instances []string, counters []string) (*Collector, error) {
	var handle pdhQueryHandle

	if ret := PdhOpenQuery(0, 0, &handle); ret != ErrorSuccess {
		return nil, NewPdhError(ret)
	}

	if len(instances) == 0 {
		instances = []string{InstanceEmpty}
	}

	collector := &Collector{
		object:                object,
		counters:              make(map[string]Counter, len(counters)),
		handle:                handle,
		totalCounterRequested: slices.Contains(instances, InstanceTotal),
		mu:                    sync.RWMutex{},
	}

	errs := make([]error, 0, len(counters))

	for _, counterName := range counters {
		if counterName == "*" {
			return nil, errors.New("wildcard counters are not supported")
		}

		counter := Counter{
			Name:      counterName,
			Instances: make(map[string]pdhCounterHandle, len(instances)),
		}

		var counterPath string

		for _, instance := range instances {
			counterPath = formatCounterPath(object, instance, counterName)

			var counterHandle pdhCounterHandle

			if ret := PdhAddEnglishCounter(handle, counterPath, 0, &counterHandle); ret != ErrorSuccess {
				errs = append(errs, fmt.Errorf("failed to add counter %s: %w", counterPath, NewPdhError(ret)))

				continue
			}

			counter.Instances[instance] = counterHandle

			if counter.Type != 0 {
				continue
			}

			// Get the info with the current buffer size
			bufLen := uint32(0)

			if ret := PdhGetCounterInfo(counterHandle, 0, &bufLen, nil); ret != PdhMoreData {
				errs = append(errs, fmt.Errorf("PdhGetCounterInfo: %w", NewPdhError(ret)))

				continue
			}

			buf := make([]byte, bufLen)
			if ret := PdhGetCounterInfo(counterHandle, 0, &bufLen, &buf[0]); ret != ErrorSuccess {
				errs = append(errs, fmt.Errorf("PdhGetCounterInfo: %w", NewPdhError(ret)))

				continue
			}

			ci := (*PdhCounterInfo)(unsafe.Pointer(&buf[0]))
			counter.Type = ci.DwType
			counter.Desc = windows.UTF16PtrToString(ci.SzExplainText)

			if counter.Type == PERF_ELAPSED_TIME {
				if ret := PdhGetCounterTimeBase(counterHandle, &counter.Frequency); ret != ErrorSuccess {
					errs = append(errs, fmt.Errorf("PdhGetCounterTimeBase: %w", NewPdhError(ret)))

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

	collector.collectCh = make(chan struct{})
	collector.counterValuesCh = make(chan CounterValues)
	collector.errorCh = make(chan error)

	go collector.collectRoutine()

	if _, err := collector.Collect(); err != nil && !errors.Is(err, ErrNoData) {
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

func (c *Collector) Collect() (CounterValues, error) {
	if c == nil {
		return CounterValues{}, ErrPerformanceCounterNotInitialized
	}

	c.mu.RLock()
	defer c.mu.RUnlock()

	if len(c.counters) == 0 || c.handle == 0 || c.collectCh == nil || c.counterValuesCh == nil || c.errorCh == nil {
		return nil, ErrPerformanceCounterNotInitialized
	}

	c.collectCh <- struct{}{}

	return <-c.counterValuesCh, <-c.errorCh
}

func (c *Collector) collectRoutine() {
	for range c.collectCh {
		if ret := PdhCollectQueryData(c.handle); ret != ErrorSuccess {
			c.counterValuesCh <- nil
			c.errorCh <- fmt.Errorf("failed to collect query data: %w", NewPdhError(ret))

			continue
		}

		counterValues, err := (func() (CounterValues, error) {
			var data CounterValues

			for _, counter := range c.counters {
				for _, instance := range counter.Instances {
					// Get the info with the current buffer size
					var itemCount uint32

					// Get the info with the current buffer size
					bufLen := uint32(0)

					ret := PdhGetRawCounterArray(instance, &bufLen, &itemCount, nil)
					if ret != PdhMoreData {
						return nil, fmt.Errorf("PdhGetRawCounterArray: %w", NewPdhError(ret))
					}

					buf := make([]byte, bufLen)

					ret = PdhGetRawCounterArray(instance, &bufLen, &itemCount, &buf[0])
					if ret != ErrorSuccess {
						if err := NewPdhError(ret); !isKnownCounterDataError(err) {
							return nil, fmt.Errorf("PdhGetRawCounterArray: %w", err)
						}

						continue
					}

					items := unsafe.Slice((*PdhRawCounterItem)(unsafe.Pointer(&buf[0])), itemCount)

					if data == nil {
						data = make(CounterValues, itemCount)
					}

					var metricType prometheus.ValueType
					if val, ok := SupportedCounterTypes[counter.Type]; ok {
						metricType = val
					} else {
						metricType = prometheus.GaugeValue
					}

					for _, item := range items {
						if item.RawValue.CStatus == PdhCstatusValidData || item.RawValue.CStatus == PdhCstatusNewData {
							instanceName := windows.UTF16PtrToString(item.SzName)
							if strings.HasSuffix(instanceName, InstanceTotal) && !c.totalCounterRequested {
								continue
							}

							if instanceName == "" || instanceName == "*" {
								instanceName = InstanceEmpty
							}

							if _, ok := data[instanceName]; !ok {
								data[instanceName] = make(map[string]CounterValue, len(c.counters))
							}

							values := CounterValue{
								Type: metricType,
							}

							// This is a workaround for the issue with the elapsed time counter type.
							// Source: https://github.com/prometheus-community/windows_exporter/pull/335/files#diff-d5d2528f559ba2648c2866aec34b1eaa5c094dedb52bd0ff22aa5eb83226bd8dR76-R83
							// Ref: https://learn.microsoft.com/en-us/windows/win32/perfctrs/calculating-counter-values

							switch counter.Type {
							case PERF_ELAPSED_TIME:
								values.FirstValue = float64((item.RawValue.FirstValue - WindowsEpoch) / counter.Frequency)
							case PERF_100NSEC_TIMER, PERF_PRECISION_100NS_TIMER:
								values.FirstValue = float64(item.RawValue.FirstValue) * TicksToSecondScaleFactor
							case PERF_AVERAGE_BULK, PERF_RAW_FRACTION:
								values.FirstValue = float64(item.RawValue.FirstValue)
								values.SecondValue = float64(item.RawValue.SecondValue)
							default:
								values.FirstValue = float64(item.RawValue.FirstValue)
							}

							data[instanceName][counter.Name] = values
						}
					}
				}
			}

			return data, nil
		})()

		if err == nil && len(counterValues) == 0 {
			err = ErrNoData
		}

		c.counterValuesCh <- counterValues
		c.errorCh <- err
	}

	return
}

func (c *Collector) Close() {
	if c == nil {
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	PdhCloseQuery(c.handle)

	c.handle = 0

	close(c.collectCh)
	close(c.counterValuesCh)
	close(c.errorCh)

	c.counterValuesCh = nil
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

	return errors.As(err, &pdhErr) && (pdhErr.ErrorCode == PdhInvalidData ||
		pdhErr.ErrorCode == PdhCalcNegativeDenominator ||
		pdhErr.ErrorCode == PdhCalcNegativeValue ||
		pdhErr.ErrorCode == PdhCstatusInvalidData ||
		pdhErr.ErrorCode == PdhCstatusNoInstance ||
		pdhErr.ErrorCode == PdhNoData)
}
