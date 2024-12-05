package registry

import (
	"fmt"
	"strings"

	"github.com/prometheus-community/windows_exporter/internal/pdh"
	"github.com/prometheus/client_golang/prometheus"
)

type Collector struct {
	object string
	query  string
}

type Counter struct {
	Name      string
	Desc      string
	Instances map[string]uint32
	Type      uint32
	Frequency float64
}

func NewCollector(object string, _ []string, _ []string) (*Collector, error) {
	collector := &Collector{
		object: object,
		query:  MapCounterToIndex(object),
	}

	if _, err := collector.Collect(); err != nil {
		return nil, fmt.Errorf("failed to collect initial data: %w", err)
	}

	return collector, nil
}

func (c *Collector) Describe() map[string]string {
	return map[string]string{}
}

func (c *Collector) Collect() (pdh.CounterValues, error) {
	perfObjects, err := QueryPerformanceData(c.query, c.object)
	if err != nil {
		return nil, fmt.Errorf("QueryPerformanceData: %w", err)
	}

	if len(perfObjects) == 0 || perfObjects[0] == nil || len(perfObjects[0].Instances) == 0 {
		return pdh.CounterValues{}, nil
	}

	data := make(pdh.CounterValues, len(perfObjects[0].Instances))

	for _, perfObject := range perfObjects {
		if perfObject.Name != c.object {
			continue
		}

		for _, perfInstance := range perfObject.Instances {
			instanceName := perfInstance.Name
			if strings.HasSuffix(instanceName, "_Total") {
				continue
			}

			if instanceName == "" || instanceName == "*" {
				instanceName = pdh.InstanceEmpty
			}

			if _, ok := data[instanceName]; !ok {
				if _, ok := data[instanceName]; !ok {
					data[instanceName] = make(map[string]pdh.CounterValue, len(perfInstance.Counters))
				}
			}

			for _, perfCounter := range perfInstance.Counters {
				if perfCounter.Def.IsBaseValue && !perfCounter.Def.IsNanosecondCounter {
					continue
				}

				if _, ok := data[instanceName][perfCounter.Def.Name]; !ok {
					data[instanceName][perfCounter.Def.Name] = pdh.CounterValue{
						Type: prometheus.GaugeValue,
					}
				}

				var metricType prometheus.ValueType
				if val, ok := pdh.SupportedCounterTypes[perfCounter.Def.CounterType]; ok {
					metricType = val
				} else {
					metricType = prometheus.GaugeValue
				}

				values := pdh.CounterValue{
					Type: metricType,
				}

				switch perfCounter.Def.CounterType {
				case pdh.PERF_ELAPSED_TIME:
					values.FirstValue = float64(perfCounter.Value-pdh.WindowsEpoch) / float64(perfObject.Frequency)
					values.SecondValue = float64(perfCounter.SecondValue-pdh.WindowsEpoch) / float64(perfObject.Frequency)
				case pdh.PERF_100NSEC_TIMER, pdh.PERF_PRECISION_100NS_TIMER:
					values.FirstValue = float64(perfCounter.Value) * pdh.TicksToSecondScaleFactor
					values.SecondValue = float64(perfCounter.SecondValue) * pdh.TicksToSecondScaleFactor
				case pdh.PERF_AVERAGE_BULK, pdh.PERF_RAW_FRACTION:
					values.FirstValue = float64(perfCounter.Value)
					values.SecondValue = float64(perfCounter.SecondValue)
				default:
					values.FirstValue = float64(perfCounter.Value)
				}

				data[instanceName][perfCounter.Def.Name] = values
			}
		}
	}

	return data, nil
}

func (c *Collector) Close() {
}
