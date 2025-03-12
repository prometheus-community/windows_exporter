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

package netframework

import (
	"fmt"

	"github.com/prometheus-community/windows_exporter/internal/mi"
	"github.com/prometheus-community/windows_exporter/internal/types"
	"github.com/prometheus-community/windows_exporter/internal/utils"
	"github.com/prometheus/client_golang/prometheus"
)

func (c *Collector) buildClrInterop() {
	c.numberOfCCWs = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, collectorClrInterop+"_com_callable_wrappers_total"),
		"Displays the current number of COM callable wrappers (CCWs). A CCW is a proxy for a managed object being referenced from an unmanaged COM client.",
		[]string{"process"},
		nil,
	)
	c.numberOfMarshalling = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, collectorClrInterop+"_interop_marshalling_total"),
		"Displays the total number of times arguments and return values have been marshaled from managed to unmanaged code, and vice versa, since the application started.",
		[]string{"process"},
		nil,
	)
	c.numberOfStubs = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, collectorClrInterop+"_interop_stubs_created_total"),
		"Displays the current number of stubs created by the common language runtime. Stubs are responsible for marshaling arguments and return values from managed to unmanaged code, and vice versa, during a COM interop call or a platform invoke call.",
		[]string{"process"},
		nil,
	)
}

type Win32_PerfRawData_NETFramework_NETCLRInterop struct {
	Name string `mi:"Name"`

	NumberofCCWs             uint32 `mi:"NumberofCCWs"`
	Numberofmarshalling      uint32 `mi:"Numberofmarshalling"`
	NumberofStubs            uint32 `mi:"NumberofStubs"`
	NumberofTLBexportsPersec uint32 `mi:"NumberofTLBexportsPersec"`
	NumberofTLBimportsPersec uint32 `mi:"NumberofTLBimportsPersec"`
}

func (c *Collector) collectClrInterop(ch chan<- prometheus.Metric) error {
	var dst []Win32_PerfRawData_NETFramework_NETCLRInterop
	if err := c.miSession.Query(&dst, mi.NamespaceRootCIMv2, utils.Must(mi.NewQuery("SELECT * FROM Win32_PerfRawData_NETFramework_NETCLRInterop"))); err != nil {
		return fmt.Errorf("WMI query failed: %w", err)
	}

	for _, process := range dst {
		if process.Name == "_Global_" {
			continue
		}

		ch <- prometheus.MustNewConstMetric(
			c.numberOfCCWs,
			prometheus.CounterValue,
			float64(process.NumberofCCWs),
			process.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.numberOfMarshalling,
			prometheus.CounterValue,
			float64(process.Numberofmarshalling),
			process.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.numberOfStubs,
			prometheus.CounterValue,
			float64(process.NumberofStubs),
			process.Name,
		)
	}

	return nil
}
