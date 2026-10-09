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

package cpu_info

import (
	"context"
	"encoding/binary"
	"errors"
	"log/slog"
	"reflect"
	"testing"
	"time"

	"github.com/prometheus-community/windows_exporter/internal/mi"
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func fixtureFirmware() []byte {
	processor := make([]byte, 0x30, 0x30+len("Fixture CPU\x00\x00"))
	processor[0], processor[1], processor[5], processor[6] = 4, 0x30, 3, 107
	binary.LittleEndian.PutUint16(processor[2:], 1)
	processor[0x10], processor[0x18] = 1, 0x41
	binary.LittleEndian.PutUint16(processor[0x1c:], 2)
	binary.LittleEndian.PutUint16(processor[0x1e:], 3)
	processor[0x23], processor[0x24], processor[0x25] = 12, 12, 24
	processor = append(processor, []byte("Fixture CPU\x00\x00")...)
	table := processor

	for _, cache := range []struct{ handle, level, size uint16 }{{2, 2, 6144}, {3, 3, 0x8400}} {
		record := make([]byte, 0x1b)
		record[0], record[1] = 7, 0x1b
		binary.LittleEndian.PutUint16(record[2:], cache.handle)
		binary.LittleEndian.PutUint16(record[5:], cache.level-1)
		binary.LittleEndian.PutUint16(record[9:], cache.size)
		table = append(table, record...)
		table = append(table, 0, 0)
	}

	table = append(table, 127, 4, 0xff, 0xff, 0, 0)
	header := make([]byte, 8, 8+len(table))
	header[1], header[2] = 3, 5
	binary.LittleEndian.PutUint32(header[4:], uint32(len(table)))

	return append(header, table...)
}

func fixturePackage(pointerSize int, masks ...uint64) []byte {
	size := 32 + len(masks)*(pointerSize+8)
	record := make([]byte, size)
	binary.LittleEndian.PutUint32(record, 3)
	binary.LittleEndian.PutUint32(record[4:], uint32(size))
	binary.LittleEndian.PutUint16(record[30:], uint16(len(masks)))

	for i, mask := range masks {
		affinity := record[32+i*(pointerSize+8):]
		if pointerSize == 8 {
			binary.LittleEndian.PutUint64(affinity, mask)
		} else {
			binary.LittleEndian.PutUint32(affinity, uint32(mask))
		}

		binary.LittleEndian.PutUint16(affinity[pointerSize:], uint16(i))
	}

	return record
}

func TestProcessorFirmware(t *testing.T) {
	processor, err := parseProcessorFirmware(fixtureFirmware())
	require.NoError(t, err)
	assert.Equal(t, miProcessor{
		Family: 107, Name: "Fixture CPU", NumberOfCores: 12,
		NumberOfEnabledCore: 12, ThreadCount: 24, L2CacheSize: 6144, L3CacheSize: 65536,
	}, processor)
}

func TestProcessorFirmwareExtendedFields(t *testing.T) {
	data := fixtureFirmware()
	record := data[8:]
	record[6] = 0xfe
	binary.LittleEndian.PutUint16(record[0x28:], 280)
	record[0x23], record[0x24], record[0x25] = 0xff, 0xff, 0xff
	binary.LittleEndian.PutUint16(record[0x2a:], 300)
	binary.LittleEndian.PutUint16(record[0x2c:], 260)
	binary.LittleEndian.PutUint16(record[0x2e:], 520)
	// First cache begins after the processor and its string area.
	cache := record[0x30+len("Fixture CPU")+2:]
	binary.LittleEndian.PutUint16(cache[9:], 0xffff)
	binary.LittleEndian.PutUint32(cache[0x17:], 0x80000000|2048)

	processor, err := parseProcessorFirmware(data)
	require.NoError(t, err)
	assert.Equal(t, uint16(280), processor.Family)
	assert.Equal(t, uint32(300), processor.NumberOfCores)
	assert.Equal(t, uint32(260), processor.NumberOfEnabledCore)
	assert.Equal(t, uint32(520), processor.ThreadCount)
	assert.Equal(t, uint32(131072), processor.L2CacheSize)
}

func TestProcessorFirmwareRejectsIncompleteData(t *testing.T) {
	for length := range len(fixtureFirmware()) {
		_, err := parseProcessorFirmware(fixtureFirmware()[:length])
		require.Error(t, err, "prefix length %d", length)
	}

	cases := map[string]func([]byte){
		"unpopulated":        func(b []byte) { b[8+0x18] = 1 },
		"disabled":           func(b []byte) { b[8+0x18] = 0x42 },
		"unknown cores":      func(b []byte) { b[8+0x23] = 0 },
		"missing name":       func(b []byte) { b[8+0x10] = 9 },
		"missing cache":      func(b []byte) { binary.LittleEndian.PutUint16(b[8+0x1c:], 55) },
		"wrong cache level":  func(b []byte) { binary.LittleEndian.PutUint16(b[8+0x1c:], 3) },
		"inconsistent count": func(b []byte) { b[8+0x24] = 13 },
		"short structure":    func(b []byte) { b[9] = 3 },
		"duplicate handle":   func(b []byte) { binary.LittleEndian.PutUint16(b[8+0x30+len("Fixture CPU")+2+2:], 1) },
	}
	for name, mutate := range cases {
		t.Run(name, func(t *testing.T) {
			data := fixtureFirmware()
			mutate(data)
			_, err := parseProcessorFirmware(data)
			require.Error(t, err)
		})
	}
}

func TestSinglePackageGroups(t *testing.T) {
	for _, pointerSize := range []int{4, 8} {
		count, err := parseSinglePackage(fixturePackage(pointerSize, 0xf, 0xff), pointerSize)
		require.NoError(t, err)
		assert.Equal(t, uint32(12), count)
	}

	multiple := append(fixturePackage(8, 0xff), fixturePackage(8, 0xff)...)
	_, err := parseSinglePackage(multiple, 8)
	require.Error(t, err)

	duplicate := fixturePackage(8, 0xff, 0xff)
	binary.LittleEndian.PutUint16(duplicate[32+16+8:], 0)
	_, err = parseSinglePackage(duplicate, 8)
	require.Error(t, err)
	_, err = parseSinglePackage(fixturePackage(8, 0), 8)
	require.Error(t, err)

	valid := fixturePackage(8, 0xff)
	for i := range valid {
		_, err = parseSinglePackage(valid[:i], 8)
		require.Error(t, err, "prefix %d", i)
	}
}

func TestCompatibilityComparesEveryPublishedField(t *testing.T) {
	baseline, err := parseProcessorFirmware(fixtureFirmware())
	require.NoError(t, err)

	baseline.Architecture, baseline.DeviceID, baseline.Description, baseline.NumberOfLogicalProcessors = 9, "CPU0", "Description", 24
	assert.True(t, compatibleProcessors([]miProcessor{baseline}, []miProcessor{baseline}))

	value := reflect.ValueOf(&baseline).Elem()
	for i := range value.NumField() {
		field := value.Type().Field(i)
		if field.Name == "Total" {
			continue
		}

		t.Run(field.Name, func(t *testing.T) {
			changed := baseline

			changedField := reflect.ValueOf(&changed).Elem().Field(i)
			if changedField.Kind() == reflect.String {
				changedField.SetString(changedField.String() + "changed")
			} else {
				changedField.SetUint(changedField.Uint() + 1)
			}

			assert.False(t, compatibleProcessors([]miProcessor{changed}, []miProcessor{baseline}))
		})
	}

	assert.False(t, compatibleProcessors([]miProcessor{baseline, baseline}, []miProcessor{baseline, baseline}))
	padded := baseline
	padded.Name += "  "
	assert.True(t, compatibleProcessors([]miProcessor{padded}, []miProcessor{baseline}))
}

type gatherCollector struct {
	collector *Collector
	err       error
}

func (g *gatherCollector) Describe(ch chan<- *prometheus.Desc) {
	c := g.collector
	for _, desc := range []*prometheus.Desc{
		c.cpuInfo, c.cpuCoreCount, c.cpuEnabledCoreCount,
		c.cpuLogicalProcessorsCount, c.cpuThreadCount, c.cpuL2CacheSize, c.cpuL3CacheSize,
	} {
		ch <- desc
	}
}

func (g *gatherCollector) Collect(ch chan<- prometheus.Metric) {
	g.err = g.collector.Collect(ch, time.Second)
}

func TestWholeHostFallbackAndGatheredMetrics(t *testing.T) {
	firmware, err := parseProcessorFirmware(fixtureFirmware())
	require.NoError(t, err)

	firmware.Architecture, firmware.DeviceID, firmware.Description, firmware.NumberOfLogicalProcessors = 9, "CPU0", "Description", 24
	baseline := []miProcessor{firmware}

	for _, mode := range []string{"native", "failure", "changed", "unverified", "multisocket"} {
		t.Run(mode, func(t *testing.T) {
			c := New(nil)
			c.buildDescriptors()
			c.nativeBaseline = baseline
			nativeCalls, wmiCalls := 0, 0
			c.nativeRead = func() ([]miProcessor, error) {
				nativeCalls++

				if mode == "failure" {
					return nil, errors.New("native failure")
				}

				current := firmware
				if mode == "changed" {
					current.L2CacheSize = 1
				}

				return []miProcessor{current}, nil
			}
			fallback := firmware
			fallback.Name = "WMI result"
			fallback.L2CacheSize = 999
			expected := fallback

			c.wmiRead = func(budget time.Duration) ([]miProcessor, error) {
				wmiCalls++

				assert.Greater(t, budget, time.Duration(0))
				assert.LessOrEqual(t, budget, time.Second)

				if mode == "multisocket" {
					second := fallback
					second.DeviceID = "CPU1"

					return []miProcessor{fallback, second}, nil
				}

				return []miProcessor{fallback}, nil
			}
			if mode == "unverified" || mode == "multisocket" {
				c.nativeBaseline = nil
			}

			if mode == "native" {
				expected = firmware
			}

			gatherer := &gatherCollector{collector: c}
			registry := prometheus.NewPedanticRegistry()
			require.NoError(t, registry.Register(gatherer))
			families, err := registry.Gather()
			require.NoError(t, err)
			require.NoError(t, gatherer.err)
			require.Len(t, families, 7)

			values := map[string]float64{
				"windows_cpu_info":                   1,
				"windows_cpu_info_core":              float64(expected.NumberOfCores),
				"windows_cpu_info_enabled_core":      float64(expected.NumberOfEnabledCore),
				"windows_cpu_info_thread":            float64(expected.ThreadCount),
				"windows_cpu_info_logical_processor": float64(expected.NumberOfLogicalProcessors),
				"windows_cpu_info_l2_cache_size":     float64(expected.L2CacheSize),
				"windows_cpu_info_l3_cache_size":     float64(expected.L3CacheSize),
			}

			for _, family := range families {
				assert.Equal(t, dto.MetricType_GAUGE, family.GetType())

				expectedSeries := 1
				if mode == "multisocket" {
					expectedSeries = 2
				}

				require.Len(t, family.GetMetric(), expectedSeries)

				for _, metric := range family.GetMetric() {
					assert.InDelta(t, values[family.GetName()], metric.GetGauge().GetValue(), 1e-12)

					labels := make(map[string]string)
					for _, label := range metric.GetLabel() {
						labels[label.GetName()] = label.GetValue()
					}

					if family.GetName() == "windows_cpu_info" {
						require.Len(t, labels, 5)
						assert.Equal(t, expected.Name, labels["name"])
						assert.Equal(t, "9", labels["architecture"])
						assert.Equal(t, "107", labels["family"])
						assert.Equal(t, "Description", labels["description"])
					} else {
						require.Len(t, labels, 1)
					}

					require.Contains(t, []string{"CPU0", "CPU1"}, labels["device_id"])
				}
			}

			if mode == "native" {
				assert.Equal(t, 0, wmiCalls)
			} else {
				assert.Equal(t, 1, wmiCalls)
			}

			if mode == "unverified" || mode == "multisocket" {
				assert.Equal(t, 0, nativeCalls)
			}
		})
	}
}

func TestNativeCompatibilityWithWMI(t *testing.T) {
	application, err := mi.ApplicationInitialize()
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, application.Close()) })

	session, err := application.NewSession(nil)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, session.Close()) })

	c := New(nil)
	require.NoError(t, c.Build(slog.New(slog.DiscardHandler), session))
	baseline, err := c.readWMI(time.Minute)
	require.NoError(t, err)

	native, nativeErr := readNativeProcessors()
	// A repeated Build must discard any previous eligibility decision, including
	// when the current host uses the WMI fallback. Nil loggers remain accepted.
	c.nativeBaseline = []miProcessor{{DeviceID: "obsolete CPU"}}
	require.NoError(t, c.Build(nil, session))
	t.Logf("WMI baseline: %+v; native candidate: %+v; native error: %v; native selected: %v",
		baseline, native, nativeErr, len(c.nativeBaseline) != 0)

	if nativeErr != nil || !compatibleProcessors(native, baseline) {
		assert.Empty(t, c.nativeBaseline, "nonmatching native result must retain whole WMI")
	} else {
		require.NotEmpty(t, c.nativeBaseline)
		actual, err := c.readProcessors(time.Minute)
		require.NoError(t, err)
		assert.True(t, compatibleProcessors(actual, baseline))
	}
}

func BenchmarkProcessorFirmware(b *testing.B) {
	data := fixtureFirmware()

	b.ReportAllocs()

	for b.Loop() {
		if _, err := parseProcessorFirmware(data); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkSinglePackage(b *testing.B) {
	data := fixturePackage(8, 0xffffffffffffffff, 0xffffffffffffffff)

	b.ReportAllocs()

	for b.Loop() {
		if _, err := parseSinglePackage(data, 8); err != nil {
			b.Fatal(err)
		}
	}
}

func FuzzProcessorFirmware(f *testing.F) {
	f.Add(fixtureFirmware())
	f.Add([]byte{})
	f.Fuzz(func(_ *testing.T, data []byte) { _, _ = parseProcessorFirmware(data) })
}

func FuzzSinglePackage(f *testing.F) {
	f.Add(fixturePackage(8, 0xffff, 0xffff))
	f.Add([]byte{})
	f.Fuzz(func(_ *testing.T, data []byte) { _, _ = parseSinglePackage(data, 8) })
}

func TestNativeFailureRespectsRemainingBudget(t *testing.T) {
	c := New(nil)
	c.nativeBaseline = []miProcessor{{DeviceID: "CPU0"}}
	c.nativeRead = func() ([]miProcessor, error) {
		time.Sleep(time.Millisecond)

		return nil, errors.New("native failure")
	}
	called := false
	c.wmiRead = func(_ time.Duration) ([]miProcessor, error) {
		called = true

		return nil, nil
	}
	_, err := c.readProcessors(time.Nanosecond)
	require.ErrorIs(t, err, context.DeadlineExceeded)
	assert.False(t, called)
}

func TestWMIFallbackFailurePublishesNoNativeMetrics(t *testing.T) {
	c := New(nil)
	c.buildDescriptors()
	c.nativeBaseline = []miProcessor{{DeviceID: "CPU0"}}
	c.nativeRead = func() ([]miProcessor, error) { return []miProcessor{{DeviceID: "CPU1"}}, nil }
	expectedErr := errors.New("WMI unavailable")
	c.wmiRead = func(_ time.Duration) ([]miProcessor, error) { return nil, expectedErr }
	metrics := make(chan prometheus.Metric, 10)
	err := c.Collect(metrics, time.Second)
	require.ErrorIs(t, err, expectedErr)
	assert.Empty(t, metrics)
}

func BenchmarkNativeProcessorQuery(b *testing.B) {
	if _, err := readNativeProcessors(); err != nil {
		b.Skipf("native candidate unavailable: %v", err)
	}

	b.ReportAllocs()

	for b.Loop() {
		if _, err := readNativeProcessors(); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkWMIProcessorQuery(b *testing.B) {
	application, err := mi.ApplicationInitialize()
	require.NoError(b, err)
	b.Cleanup(func() { assert.NoError(b, application.Close()) })

	session, err := application.NewSession(nil)
	require.NoError(b, err)
	b.Cleanup(func() { assert.NoError(b, session.Close()) })

	c := New(nil)
	require.NoError(b, c.Build(slog.New(slog.DiscardHandler), session))
	b.ReportAllocs()

	for b.Loop() {
		if _, err := c.readWMI(time.Minute); err != nil {
			b.Fatal(err)
		}
	}
}

func TestNativeFailureReducesWMIBudget(t *testing.T) {
	c := New(nil)
	c.nativeBaseline = []miProcessor{{DeviceID: "CPU0"}}
	c.nativeRead = func() ([]miProcessor, error) {
		time.Sleep(2 * time.Millisecond)

		return nil, errors.New("native failure")
	}
	called := false
	c.wmiRead = func(budget time.Duration) ([]miProcessor, error) {
		called = true

		require.Positive(t, budget)
		assert.LessOrEqual(t, budget, time.Second-time.Millisecond)

		return nil, nil
	}
	_, err := c.readProcessors(time.Second)
	require.NoError(t, err)
	assert.True(t, called)
}

func TestBuildClearsPreviousNativeEligibilityOnError(t *testing.T) {
	c := New(nil)
	c.nativeBaseline = []miProcessor{{DeviceID: "obsolete CPU"}}
	err := c.Build(nil, nil)
	require.Error(t, err)
	assert.Empty(t, c.nativeBaseline)
}

func TestUnboundedWMIBudgetIsPreserved(t *testing.T) {
	for _, budget := range []time.Duration{0, -time.Second} {
		c := New(nil)
		c.wmiRead = func(actual time.Duration) ([]miProcessor, error) {
			assert.Equal(t, budget, actual)

			return nil, nil
		}
		_, err := c.readProcessors(budget)
		require.NoError(t, err)
	}
}
