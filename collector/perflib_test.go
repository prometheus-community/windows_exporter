package collector

import (
	"reflect"
	"testing"

	"github.com/prometheus-community/windows_exporter/perflib"

	"github.com/go-kit/log"
)

type simple struct {
	ValA float64 `perflib:"Something"`
	ValB float64 `perflib:"Something Else"`
	ValC float64 `perflib:"Something Else,secondvalue"`
}

func TestUnmarshalPerflib(t *testing.T) {
	cases := []struct {
		name string
		obj  *perflib.PerfObject

		expectedOutput []simple
		expectError    bool
	}{
		{
			name:           "nil check",
			obj:            nil,
			expectedOutput: []simple{},
			expectError:    true,
		},
		{
			name: "Simple",
			obj: &perflib.PerfObject{
				Instances: []*perflib.PerfInstance{
					{
						Counters: []*perflib.PerfCounter{
							{
								Def: &perflib.PerfCounterDef{
									Name:        "Something",
									CounterType: perflib.PERF_COUNTER_COUNTER,
								},
								Value: 123,
							},
						},
					},
				},
			},
			expectedOutput: []simple{{ValA: 123}},
			expectError:    false,
		},
		{
			name: "Multiple properties",
			obj: &perflib.PerfObject{
				Instances: []*perflib.PerfInstance{
					{
						Counters: []*perflib.PerfCounter{
							{
								Def: &perflib.PerfCounterDef{
									Name:        "Something",
									CounterType: perflib.PERF_COUNTER_COUNTER,
								},
								Value: 123,
							},
							{
								Def: &perflib.PerfCounterDef{
									Name:           "Something Else",
									CounterType:    perflib.PERF_COUNTER_COUNTER,
									HasSecondValue: true,
								},
								Value:       256,
								SecondValue: 222,
							},
						},
					},
				},
			},
			expectedOutput: []simple{{ValA: 123, ValB: 256, ValC: 222}},
			expectError:    false,
		},
		{
			name: "Multiple instances",
			obj: &perflib.PerfObject{
				Instances: []*perflib.PerfInstance{
					{
						Counters: []*perflib.PerfCounter{
							{
								Def: &perflib.PerfCounterDef{
									Name:        "Something",
									CounterType: perflib.PERF_COUNTER_COUNTER,
								},
								Value: 321,
							},
						},
					},
					{
						Counters: []*perflib.PerfCounter{
							{
								Def: &perflib.PerfCounterDef{
									Name:        "Something",
									CounterType: perflib.PERF_COUNTER_COUNTER,
								},
								Value: 231,
							},
						},
					},
				},
			},
			expectedOutput: []simple{{ValA: 321}, {ValA: 231}},
			expectError:    false,
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			output := make([]simple, 0)
			err := unmarshalObject(c.obj, &output, log.NewNopLogger())
			if err != nil && !c.expectError {
				t.Errorf("Did not expect error, got %q", err)
			}
			if err == nil && c.expectError {
				t.Errorf("Expected an error, but got ok")
			}

			if err == nil && !reflect.DeepEqual(output, c.expectedOutput) {
				t.Errorf("Output mismatch, expected %+v, got %+v", c.expectedOutput, output)
			}
		})
	}
}
