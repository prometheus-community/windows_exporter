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
	InstanceAll   = []string{"*"}
	InstanceTotal = []string{"_Total"}
)

type Collector struct {
	object                string
	counters              map[string]Counter
	handle                pdhQueryHandle
	totalCounterRequested bool
	mu                    sync.RWMutex
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
		instances = []string{EmptyInstance}
	}

	collector := &Collector{
		object:                object,
		counters:              make(map[string]Counter, len(counters)),
		handle:                handle,
		totalCounterRequested: slices.Contains(instances, "_Total"),
		mu:                    sync.RWMutex{},
	}

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
				return nil, fmt.Errorf("failed to add counter %s: %w", counterPath, NewPdhError(ret))
			}

			counter.Instances[instance] = counterHandle

			if counter.Type != 0 {
				continue
			}

			// Get the info with the current buffer size
			bufLen := uint32(0)

			if ret := PdhGetCounterInfo(counterHandle, 0, &bufLen, nil); ret != PdhMoreData {
				return nil, fmt.Errorf("PdhGetCounterInfo: %w", NewPdhError(ret))
			}

			buf := make([]byte, bufLen)
			if ret := PdhGetCounterInfo(counterHandle, 0, &bufLen, &buf[0]); ret != ErrorSuccess {
				return nil, fmt.Errorf("PdhGetCounterInfo: %w", NewPdhError(ret))
			}

			ci := (*PdhCounterInfo)(unsafe.Pointer(&buf[0]))
			counter.Type = ci.DwType
			counter.Desc = windows.UTF16PtrToString(ci.SzExplainText)

			if counter.Type == PERF_ELAPSED_TIME {
				if ret := PdhGetCounterTimeBase(counterHandle, &counter.Frequency); ret != ErrorSuccess {
					return nil, fmt.Errorf("PdhGetCounterTimeBase: %w", NewPdhError(ret))
				}
			}
		}

		collector.counters[counterName] = counter
	}

	if len(collector.counters) == 0 {
		return nil, errors.New("no counters configured")
	}

	if _, err := collector.Collect(); err != nil {
		return collector, fmt.Errorf("failed to collect initial data: %w", err)
	}

	return collector, nil
}

func (c *Collector) Describe() map[string]string {
	desc := make(map[string]string, len(c.counters))

	for _, counter := range c.counters {
		desc[counter.Name] = counter.Desc
	}

	return desc
}

func (c *Collector) Collect() (map[string]map[string]CounterValues, error) {
	if len(c.counters) == 0 {
		return map[string]map[string]CounterValues{}, nil
	}

	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.handle == 0 {
		return map[string]map[string]CounterValues{}, nil
	}

	if ret := PdhCollectQueryData(c.handle); ret != ErrorSuccess {
		return nil, fmt.Errorf("failed to collect query data: %w", NewPdhError(ret))
	}

	var data map[string]map[string]CounterValues

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
				data = make(map[string]map[string]CounterValues, itemCount)
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
					if strings.HasSuffix(instanceName, "_Total") && !c.totalCounterRequested {
						continue
					}

					if instanceName == "" || instanceName == "*" {
						instanceName = EmptyInstance
					}

					if _, ok := data[instanceName]; !ok {
						data[instanceName] = make(map[string]CounterValues, len(c.counters))
					}

					values := CounterValues{
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
}

func (c *Collector) Close() {
	c.mu.Lock()
	defer c.mu.Unlock()

	PdhCloseQuery(c.handle)

	c.handle = 0
}

func formatCounterPath(object, instance, counterName string) string {
	var counterPath string

	if instance == EmptyInstance {
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
