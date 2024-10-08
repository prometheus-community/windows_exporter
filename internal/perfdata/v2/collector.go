//go:build windows

package v2

import (
	"errors"
	"fmt"
	"strings"
	"unsafe"

	"github.com/prometheus-community/windows_exporter/internal/perfdata/perftypes"
	"github.com/prometheus/client_golang/prometheus"
	"golang.org/x/sys/windows"
)

type Collector struct {
	object   string
	counters map[string]Counter
	handle   pdhQueryHandle
}

type Counter struct {
	Name      string
	Desc      string
	Instances map[string]pdhCounterHandle
	Type      uint32
	Frequency float64
}

func NewCollector(object string, instances []string, counters []string) (*Collector, error) {
	var handle pdhQueryHandle

	if ret := PdhOpenQuery(0, 0, &handle); ret != ErrorSuccess {
		return nil, NewPdhError(ret)
	}

	if len(instances) == 0 {
		instances = []string{perftypes.EmptyInstance}
	}

	collector := &Collector{
		object:   object,
		counters: make(map[string]Counter, len(counters)),
		handle:   handle,
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

			if counter.Type == 0 {
				// Get the info with the current buffer size
				bufLen := uint32(0)

				if ret := PdhGetCounterInfo(counterHandle, 1, &bufLen, nil); ret != PdhMoreData {
					return nil, fmt.Errorf("PdhGetCounterInfo: %w", NewPdhError(ret))
				}

				buf := make([]byte, bufLen)
				if ret := PdhGetCounterInfo(counterHandle, 1, &bufLen, &buf[0]); ret != ErrorSuccess {
					return nil, fmt.Errorf("PdhGetCounterInfo: %w", NewPdhError(ret))
				}

				ci := (*PdhCounterInfo)(unsafe.Pointer(&buf[0]))
				counter.Type = ci.DwType
				counter.Desc = windows.UTF16PtrToString(ci.SzExplainText)

				frequency := float64(0)

				if ret := PdhGetCounterTimeBase(counterHandle, &frequency); ret != ErrorSuccess {
					return nil, fmt.Errorf("PdhGetCounterTimeBase: %w", NewPdhError(ret))
				}

				counter.Frequency = frequency
			}
		}

		collector.counters[counterName] = counter
	}

	if len(collector.counters) == 0 {
		return nil, errors.New("no counters configured")
	}

	if _, err := collector.Collect(); err != nil {
		return nil, fmt.Errorf("failed to collect initial data: %w", err)
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

func (c *Collector) Collect() (map[string]map[string]perftypes.CounterValues, error) {
	if len(c.counters) == 0 {
		return map[string]map[string]perftypes.CounterValues{}, nil
	}

	if ret := PdhCollectQueryData(c.handle); ret != ErrorSuccess {
		return nil, fmt.Errorf("failed to collect query data: %w", NewPdhError(ret))
	}

	var data map[string]map[string]perftypes.CounterValues

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

			items := (*[1 << 20]PdhRawCounterItem)(unsafe.Pointer(&buf[0]))[:itemCount]

			if data == nil {
				data = make(map[string]map[string]perftypes.CounterValues, itemCount)
			}

			var metricType prometheus.ValueType
			if val, ok := perftypes.SupportedCounterTypes[counter.Type]; ok {
				metricType = val
			} else {
				metricType = prometheus.GaugeValue
			}

			for _, item := range items {
				if item.RawValue.CStatus == PdhCstatusValidData || item.RawValue.CStatus == PdhCstatusNewData {
					instanceName := windows.UTF16PtrToString(item.SzName)
					if strings.HasSuffix(instanceName, "_Total") {
						continue
					}

					if instanceName == "" || instanceName == "*" {
						instanceName = perftypes.EmptyInstance
					}

					if _, ok := data[instanceName]; !ok {
						data[instanceName] = make(map[string]perftypes.CounterValues, len(c.counters))
					}

					values := perftypes.CounterValues{
						Type: metricType,
					}

					// This is a workaround for the issue with the elapsed time counter type.
					// Source: https://github.com/prometheus-community/windows_exporter/pull/335/files#diff-d5d2528f559ba2648c2866aec34b1eaa5c094dedb52bd0ff22aa5eb83226bd8dR76-R83
					// Ref: https://learn.microsoft.com/en-us/windows/win32/perfctrs/calculating-counter-values

					switch counter.Type {
					case perftypes.PERF_ELAPSED_TIME:
						values.FirstValue = float64(item.RawValue.FirstValue-perftypes.WindowsEpoch) / counter.Frequency
						values.SecondValue = float64(item.RawValue.SecondValue-perftypes.WindowsEpoch) / counter.Frequency
					case perftypes.PERF_100NSEC_TIMER, perftypes.PERF_PRECISION_100NS_TIMER:
						values.FirstValue = float64(item.RawValue.FirstValue) * perftypes.TicksToSecondScaleFactor
						values.SecondValue = float64(item.RawValue.SecondValue) * perftypes.TicksToSecondScaleFactor
					default:
						values.FirstValue = float64(item.RawValue.FirstValue)
						values.SecondValue = float64(item.RawValue.SecondValue)
					}

					data[instanceName][counter.Name] = values
				}
			}
		}
	}

	return data, nil
}

func (c *Collector) Close() {
	PdhCloseQuery(c.handle)
}

func formatCounterPath(object, instance, counterName string) string {
	var counterPath string

	if instance == perftypes.EmptyInstance {
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
