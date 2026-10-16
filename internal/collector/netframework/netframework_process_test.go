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

package netframework

import (
	"errors"
	"testing"
	"time"

	"github.com/prometheus-community/windows_exporter/internal/labelcache"
	"github.com/prometheus-community/windows_exporter/internal/utils"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/require"
)

// gatherProcessIDs gathers the metrics of collect and returns the process_id
// label of each process label of the metric family name.
func gatherProcessIDs(t *testing.T, collect func(chan<- prometheus.Metric) error, name string) (map[string]string, error) {
	t.Helper()

	var collectErr error

	registry := prometheus.NewRegistry()
	registry.MustRegister(fixturePromCollector{collect: func(ch chan<- prometheus.Metric) { collectErr = collect(ch) }})
	families, err := registry.Gather()
	require.NoError(t, err)

	got := map[string]string{}

	for _, family := range families {
		if family.GetName() != name {
			continue
		}

		for _, metric := range family.GetMetric() {
			labels := make(map[string]string)
			for _, label := range metric.GetLabel() {
				labels[label.GetName()] = label.GetValue()
			}

			require.Len(t, labels, 2)
			got[labels["process"]] = labels["process_id"]
		}
	}

	return got, collectErr
}

// gatherExceptionsThrown runs Collect with the clrexceptions sub-collector of
// c and returns the process_id label of each process of
// windows_netframework_clrexceptions_exceptions_thrown_total.
func gatherExceptionsThrown(t *testing.T, c *Collector) (map[string]string, error) {
	t.Helper()

	c.describeClrExceptions()
	c.collectorFns = []func(chan<- prometheus.Metric, *labelcache.Scrape, time.Duration) error{c.collectClrExceptions}

	return gatherProcessIDs(t, func(ch chan<- prometheus.Metric) error {
		return c.Collect(ch, time.Second)
	}, "windows_netframework_clrexceptions_exceptions_thrown_total")
}

// TestProcessIDLookup checks that the metrics of the other .NET CLR objects
// get the process ID of the .NET CLR Memory instance with the same name and
// occurrence, and an empty process_id if .NET CLR Memory has none.
func TestProcessIDLookup(t *testing.T) {
	c := New(nil)
	c.perfClrProcessID = &fixturePerfCollector[perfDataClrProcessID]{rows: []perfDataClrProcessID{
		{Name: "_Global_", ProcessID: 0},
		{Name: "w3wp", ProcessID: 10},
		{Name: "w3wp#1", ProcessID: 30},
		{Name: "w3wp", Occurrence: 1, ProcessID: 20},
	}}
	c.perfClrExceptions = &fixturePerfCollector[perfDataClrExceptions]{rows: []perfDataClrExceptions{
		{Name: "_Global_"},
		{Name: "w3wp"},
		{Name: "w3wp#1"},
		{Name: "w3wp", Occurrence: 1},
		{Name: "powershell"},
	}}

	got, err := gatherExceptionsThrown(t, c)
	require.NoError(t, err)

	// "w3wp#1" is the name of a process, so the second w3wp is "w3wp#2".
	require.Equal(t, map[string]string{"w3wp": "10", "w3wp#1": "30", "w3wp#2": "20", "powershell": ""}, got)

	// A scrape does not keep the process IDs of an earlier one.
	c.perfClrProcessID = &fixturePerfCollector[perfDataClrProcessID]{rows: []perfDataClrProcessID{{Name: "w3wp", ProcessID: 40}}}

	got, err = gatherExceptionsThrown(t, c)
	require.NoError(t, err)
	require.Equal(t, map[string]string{"w3wp": "40", "w3wp#1": "", "w3wp#2": "", "powershell": ""}, got)
}

// TestProcessIDLookupFailure checks that the metrics are published with an
// empty process_id if the process IDs are unavailable, and that only a real
// failure is reported.
func TestProcessIDLookupFailure(t *testing.T) {
	failure := errors.New("provider failed")

	for _, tc := range []struct {
		name    string
		lookup  perfCollector[perfDataClrProcessID]
		wantErr error
	}{
		{"missing object", unavailablePerfCollector[perfDataClrProcessID]{}, nil},
		{"failure", &fixturePerfCollector[perfDataClrProcessID]{err: failure}, failure},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := New(nil)
			c.perfClrProcessID = tc.lookup
			c.perfClrExceptions = &fixturePerfCollector[perfDataClrExceptions]{rows: []perfDataClrExceptions{{Name: "w3wp"}}}

			got, err := gatherExceptionsThrown(t, c)
			if tc.wantErr == nil {
				require.NoError(t, err)
			} else {
				require.ErrorIs(t, err, tc.wantErr)
			}

			require.Equal(t, map[string]string{"w3wp": ""}, got)
		})
	}
}

// TestProcessFilter checks that the process filters match the name that
// Windows reports, without the occurrence suffix, and that the process IDs
// still match after filtering.
func TestProcessFilter(t *testing.T) {
	for _, tc := range []struct {
		name     string
		include  string
		exclude  string
		expected map[string]string
	}{
		{"default", ".+", "", map[string]string{"w3wp": "10", "w3wp#1": "30", "w3wp#2": "20", "powershell": "40"}},
		// Without the process named "w3wp#1", the second w3wp is "w3wp#1".
		{"include", "w3wp", "", map[string]string{"w3wp": "10", "w3wp#1": "20"}},
		{"include name with #", "w3wp#1", "", map[string]string{"w3wp#1": "30"}},
		{"exclude", ".+", "w3wp.*", map[string]string{"powershell": "40"}},
		{"include and exclude", "w3wp.*", "w3wp", map[string]string{"w3wp#1": "30"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			include, err := utils.CompileAnchored(tc.include)
			require.NoError(t, err)

			exclude, err := utils.CompileAnchored(tc.exclude)
			require.NoError(t, err)

			c := New(&Config{ProcessInclude: include, ProcessExclude: exclude})
			c.perfClrProcessID = &fixturePerfCollector[perfDataClrProcessID]{rows: []perfDataClrProcessID{
				{Name: "w3wp", ProcessID: 10},
				{Name: "w3wp#1", ProcessID: 30},
				{Name: "powershell", ProcessID: 40},
				{Name: "w3wp", Occurrence: 1, ProcessID: 20},
			}}
			c.perfClrExceptions = &fixturePerfCollector[perfDataClrExceptions]{rows: []perfDataClrExceptions{
				{Name: "w3wp"},
				{Name: "w3wp#1"},
				{Name: "powershell"},
				{Name: "w3wp", Occurrence: 1},
			}}

			got, err := gatherExceptionsThrown(t, c)
			require.NoError(t, err)
			require.Equal(t, tc.expected, got)

			// The clrmemory metrics have the same process labels.
			c.describeClrMemory()
			c.perfClrMemory = &fixturePerfCollector[perfDataClrMemory]{rows: []perfDataClrMemory{
				{Name: "w3wp", ProcessID: 10},
				{Name: "w3wp#1", ProcessID: 30},
				{Name: "powershell", ProcessID: 40},
				{Name: "w3wp", Occurrence: 1, ProcessID: 20},
			}}

			got, err = gatherProcessIDs(t, func(ch chan<- prometheus.Metric) error {
				return c.collectClrMemory(ch, nil, time.Second)
			}, "windows_netframework_clrmemory_allocated_bytes_total")
			require.NoError(t, err)
			require.Equal(t, tc.expected, got)
		})
	}
}
