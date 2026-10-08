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

package registry

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/prometheus-community/windows_exporter/internal/pdh"
)

type Collector[T any] struct {
	object string
	query  string

	counters       map[string]Counter
	nameIndexValue int
}

type Counter struct {
	Name      string
	Desc      string
	Instances map[string]uint32
	Type      uint32
	Frequency float64

	FieldIndexValue       int
	FieldIndexSecondValue int
}

// NewCollector creates a collector for the counters declared by the perfdata_v1 or perfdata tags of the struct T.
func NewCollector[T any](object string, _ []string) (*Collector[T], error) {
	valueType := reflect.TypeFor[T]()
	if valueType.Kind() != reflect.Struct {
		return nil, fmt.Errorf("expected a struct, got %s", valueType)
	}

	collector := &Collector[T]{
		object:         object,
		query:          MapCounterToIndex(object),
		nameIndexValue: -1,
		counters:       make(map[string]Counter),
	}

	if f, ok := valueType.FieldByName("Name"); ok {
		if f.Type.Kind() == reflect.String {
			collector.nameIndexValue = f.Index[0]
		}
	}

	for _, f := range reflect.VisibleFields(valueType) {
		tag, ok := f.Tag.Lookup("perfdata_v1")
		if !ok {
			tag, ok = f.Tag.Lookup("perfdata")
			if !ok {
				continue
			}
		}

		if f.Type.Kind() != reflect.Float64 {
			return nil, fmt.Errorf("field %s must be a float64", f.Name)
		}

		counterName, secondValue := strings.CutSuffix(tag, ",secondvalue")

		counter, ok := collector.counters[counterName]
		if !ok {
			counter = Counter{
				Name:                  counterName,
				FieldIndexSecondValue: -1,
				FieldIndexValue:       -1,
			}
		}

		if secondValue {
			counter.FieldIndexSecondValue = f.Index[0]
		} else {
			counter.FieldIndexValue = f.Index[0]
		}

		collector.counters[counterName] = counter
	}

	var collectValues []T

	if err := collector.Collect(&collectValues); err != nil {
		return nil, fmt.Errorf("failed to collect initial data: %w", err)
	}

	return collector, nil
}

func (c *Collector[T]) Describe() map[string]string {
	return map[string]string{}
}

// Collect replaces the content of dst with one row per collected instance.
func (c *Collector[T]) Collect(dst *[]T) error {
	if dst == nil {
		return errors.New("dst must not be nil")
	}

	perfObjects, err := QueryPerformanceData(c.query, c.object)
	if err != nil {
		return fmt.Errorf("QueryPerformanceData: %w", err)
	}

	if len(perfObjects) == 0 || perfObjects[0] == nil || len(perfObjects[0].Instances) == 0 {
		return nil
	}

	*dst = make([]T, 0, len(perfObjects[0].Instances))

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

			var row T

			rv := reflect.ValueOf(&row).Elem()

			if c.nameIndexValue != -1 {
				rv.Field(c.nameIndexValue).SetString(instanceName)
			}

			for _, perfCounter := range perfInstance.Counters {
				if perfCounter.Def.IsBaseValue && !perfCounter.Def.IsNanosecondCounter {
					continue
				}

				counter, ok := c.counters[perfCounter.Def.Name]
				if !ok {
					continue
				}

				switch perfCounter.Def.CounterType {
				case pdh.PERF_ELAPSED_TIME:
					rv.Field(counter.FieldIndexValue).
						SetFloat(float64((perfCounter.Value - pdh.WindowsEpoch) / perfObject.Frequency))
				case pdh.PERF_100NSEC_TIMER, pdh.PERF_PRECISION_100NS_TIMER:
					rv.Field(counter.FieldIndexValue).
						SetFloat(float64(perfCounter.Value) * pdh.TicksToSecondScaleFactor)
				default:
					if counter.FieldIndexSecondValue != -1 {
						rv.Field(counter.FieldIndexSecondValue).
							SetFloat(float64(perfCounter.SecondValue))
					}

					if counter.FieldIndexValue != -1 {
						rv.Field(counter.FieldIndexValue).
							SetFloat(float64(perfCounter.Value))
					}
				}
			}

			*dst = append(*dst, row)
		}
	}

	return nil
}

func (c *Collector[T]) Close() {}
