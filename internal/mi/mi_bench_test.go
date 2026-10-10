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

package mi_test

import (
	"testing"
	"time"

	"github.com/prometheus-community/windows_exporter/internal/mi"
	"github.com/prometheus-community/windows_exporter/internal/utils/testutils"
	"github.com/stretchr/testify/require"
	"golang.org/x/sys/windows"
)

func Benchmark_MI_Query_Unmarshal(b *testing.B) {
	application, err := mi.ApplicationInitialize()
	require.NoError(b, err)
	require.NotEmpty(b, application)

	session, err := application.NewSession(nil)
	require.NoError(b, err)
	require.NotEmpty(b, session)

	b.ResetTimer()

	var processes []win32Process

	query, err := mi.NewQuery("SELECT Name FROM Win32_Process WHERE Handle = 0 OR Handle = 4")
	require.NoError(b, err)

	for b.Loop() {
		err := session.QueryUnmarshal(&processes, mi.OperationFlagsStandardRTTI, nil, mi.NamespaceRootCIMv2, mi.QueryDialectWQL, query)
		require.NoError(b, err)
		require.Equal(b, []win32Process{{Name: "System Idle Process"}, {Name: "System"}}, processes)
	}

	b.StopTimer()

	err = session.Close()
	require.NoError(b, err)

	err = application.Close()
	require.NoError(b, err)

	b.ReportAllocs()
}

func Benchmark_MI_QueryFunc_GetElement(b *testing.B) {
	application, err := mi.ApplicationInitialize()
	require.NoError(b, err)

	session, err := application.NewSession(nil)
	require.NoError(b, err)

	b.Cleanup(func() {
		require.NoError(b, session.Close())
		require.NoError(b, application.Close())
	})

	query, err := mi.NewQuery("SELECT Name, PercentIdleTime, PercentProcessorTime, InterruptsPersec FROM Win32_PerfRawData_PerfOS_Processor")
	require.NoError(b, err)

	properties := []string{"Name", "PercentIdleTime", "PercentProcessorTime", "InterruptsPersec"}

	b.ReportAllocs()

	for b.Loop() {
		err := session.QueryFunc(mi.NamespaceRootCIMv2, query, -1, func(instance *mi.Instance) error {
			for _, property := range properties {
				if _, err := instance.GetElement(property); err != nil {
					return err
				}
			}

			return nil
		})
		require.NoError(b, err)
	}
}

func Benchmark_MI_Query_Unmarshal_Processor(b *testing.B) {
	application, err := mi.ApplicationInitialize()
	require.NoError(b, err)

	session, err := application.NewSession(nil)
	require.NoError(b, err)

	b.Cleanup(func() {
		require.NoError(b, session.Close())
		require.NoError(b, application.Close())
	})

	query, err := mi.NewQuery("SELECT Name, PercentIdleTime, PercentProcessorTime, InterruptsPersec FROM Win32_PerfRawData_PerfOS_Processor")
	require.NoError(b, err)

	var processors []struct {
		Name                 string `mi:"Name"`
		PercentIdleTime      uint64 `mi:"PercentIdleTime"`
		PercentProcessorTime uint64 `mi:"PercentProcessorTime"`
		InterruptsPersec     uint32 `mi:"InterruptsPersec"`
	}

	b.ReportAllocs()

	for b.Loop() {
		require.NoError(b, session.Query(&processors, mi.NamespaceRootCIMv2, query, -1))
	}
}

// Benchmark_MI_QueryFunc_GetElementByName reads elements by names converted
// once, as the wmi collector does. The zero timeout creates per-query
// operation options.
func Benchmark_MI_QueryFunc_GetElementByName(b *testing.B) {
	application, err := mi.ApplicationInitialize()
	require.NoError(b, err)

	session, err := application.NewSession(nil)
	require.NoError(b, err)

	b.Cleanup(func() {
		require.NoError(b, session.Close())
		require.NoError(b, application.Close())
	})

	query, err := mi.NewQuery("SELECT Name, PercentIdleTime, PercentProcessorTime, InterruptsPersec FROM Win32_PerfRawData_PerfOS_Processor")
	require.NoError(b, err)

	properties := make([]mi.ElementName, 0, 4)

	for _, property := range []string{"Name", "PercentIdleTime", "PercentProcessorTime", "InterruptsPersec"} {
		name, err := mi.NewElementName(property)
		require.NoError(b, err)

		properties = append(properties, name)
	}

	b.ReportAllocs()

	for b.Loop() {
		err := session.QueryFunc(mi.NamespaceRootCIMv2, query, 0, func(instance *mi.Instance) error {
			for _, property := range properties {
				if _, err := instance.GetElementByName(property); err != nil {
					return err
				}
			}

			return nil
		})
		require.NoError(b, err)
	}
}

// The Sync benchmarks read the same results with MI_Operation_GetInstance,
// the path the high-level queries used before they switched to callbacks.

func Benchmark_MI_Query_Unmarshal_Sync(b *testing.B) {
	session := newTestSession(b)

	b.ReportAllocs()

	var processes []win32Process

	for b.Loop() {
		operation, err := session.QueryInstances(mi.OperationFlagsStandardRTTI, nil, mi.NamespaceRootCIMv2, mi.QueryDialectWQL,
			"SELECT Name FROM Win32_Process WHERE Handle = 0 OR Handle = 4")
		require.NoError(b, err)
		require.NoError(b, operation.Unmarshal(&processes))
		require.NoError(b, operation.Close())
		require.Equal(b, []win32Process{{Name: "System Idle Process"}, {Name: "System"}}, processes)
	}
}

func Benchmark_MI_QueryFunc_GetElement_Sync(b *testing.B) {
	session := newTestSession(b)

	properties := []string{"Name", "PercentIdleTime", "PercentProcessorTime", "InterruptsPersec"}

	b.ReportAllocs()

	for b.Loop() {
		operation, err := session.QueryInstances(mi.OperationFlagsStandardRTTI, nil, mi.NamespaceRootCIMv2, mi.QueryDialectWQL,
			"SELECT Name, PercentIdleTime, PercentProcessorTime, InterruptsPersec FROM Win32_PerfRawData_PerfOS_Processor")
		require.NoError(b, err)

		for {
			instance, moreResults, err := operation.GetInstance()
			require.NoError(b, err)

			if instance == nil {
				break
			}

			for _, property := range properties {
				_, err := instance.GetElement(property)
				require.NoError(b, err)
			}

			if !moreResults {
				break
			}
		}

		require.NoError(b, operation.Close())
	}
}

// Benchmark_MI_Parallel_StoragePool runs parallel queries against a slow
// provider and reports the handle count growth per 1000 queries.
func Benchmark_MI_Parallel_StoragePool(b *testing.B) {
	for _, mode := range []string{"Async", "Sync"} {
		b.Run(mode, func(b *testing.B) {
			session := newTestSession(b)
			query := storagePoolQuery(b, session)

			start, err := testutils.GetProcessHandleCount(windows.CurrentProcess())
			require.NoError(b, err)

			b.ReportAllocs()
			b.SetParallelism(2)

			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					var pools []msftStoragePool

					if mode == "Async" {
						require.NoError(b, session.Query(&pools, mi.NamespaceRootStorage, query, 4*time.Second))

						continue
					}

					operation, err := session.QueryInstances(mi.OperationFlagsStandardRTTI, nil, mi.NamespaceRootStorage, mi.QueryDialectWQL,
						"SELECT FriendlyName FROM MSFT_StoragePool")
					require.NoError(b, err)
					require.NoError(b, operation.Unmarshal(&pools))
					require.NoError(b, operation.Close())
				}
			})

			end, err := testutils.GetProcessHandleCount(windows.CurrentProcess())
			require.NoError(b, err)

			b.ReportMetric(float64(int64(end)-int64(start))*1000/float64(b.N), "handles/1k-queries")
		})
	}
}
