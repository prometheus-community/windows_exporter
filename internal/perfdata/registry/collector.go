package registry

import (
	"fmt"
	"time"

	"github.com/prometheus-community/windows_exporter/internal/perfdata/perftypes"
	"github.com/prometheus/client_golang/prometheus"
)

type Collector struct {
	time   time.Time
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

func NewCollector(object string, instances []string, _ []string) (*Collector, error) {
	if len(instances) == 0 {
		instances = []string{perftypes.EmptyInstance}
	}

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

func (c *Collector) Collect() (map[string]map[string]perftypes.CounterValues, error) {
	perfObjects, err := QueryPerformanceData(c.query)
	if err != nil {
		return nil, fmt.Errorf("QueryPerformanceData: %w", err)
	}

	c.time = time.Now()

	data := make(map[string]map[string]perftypes.CounterValues, len(perfObjects[0].Instances))

	for _, perfObject := range perfObjects {
		for _, perfInstance := range perfObject.Instances {
			instanceName := perfInstance.Name
			if instanceName == "" || instanceName == "*" {
				instanceName = perftypes.EmptyInstance
			}

			if _, ok := data[instanceName]; !ok {
				data[instanceName] = make(map[string]perftypes.CounterValues, len(perfInstance.Counters))
			}

			for _, perfCounter := range perfInstance.Counters {
				if _, ok := data[instanceName][perfCounter.Def.Name]; !ok {
					data[instanceName][perfCounter.Def.Name] = perftypes.CounterValues{
						Type: prometheus.GaugeValue,
					}
				}

				var metricType prometheus.ValueType
				if val, ok := perftypes.SupportedCounterTypes[perfCounter.Def.CounterType]; ok {
					metricType = val
				} else {
					metricType = prometheus.GaugeValue
				}

				values := perftypes.CounterValues{
					Type: metricType,
				}

				switch perfCounter.Def.CounterType {
				case perftypes.PERF_ELAPSED_TIME:
					values.FirstValue = float64(perfCounter.Value-perftypes.WindowsEpoch) / float64(perfObject.Frequency)
					values.SecondValue = float64(perfCounter.SecondValue-perftypes.WindowsEpoch) / float64(perfObject.Frequency)
				case perftypes.PERF_100NSEC_TIMER, perftypes.PERF_PRECISION_100NS_TIMER:
					values.FirstValue = float64(perfCounter.Value) * perftypes.TicksToSecondScaleFactor
					values.SecondValue = float64(perfCounter.SecondValue) * perftypes.TicksToSecondScaleFactor
				default:
					values.FirstValue = float64(perfCounter.Value)
					values.SecondValue = float64(perfCounter.SecondValue)
				}

				data[instanceName][perfCounter.Def.Name] = values
			}
		}
	}

	return data, nil
}

func (c *Collector) Close() {

}
