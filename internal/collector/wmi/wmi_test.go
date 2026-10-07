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

package wmi_test

import (
	"log/slog"
	"testing"

	"github.com/alecthomas/kingpin/v2"
	"github.com/prometheus-community/windows_exporter/internal/collector/wmi"
	"github.com/prometheus-community/windows_exporter/internal/mi"
	"github.com/prometheus-community/windows_exporter/internal/utils/testutils"
	"github.com/stretchr/testify/require"
)

func BenchmarkCollector(b *testing.B) {
	queries := `[{"name":"logical_disk","class":"Win32_LogicalDisk","where":"DriveType = 3","label_properties":[{"name":"DeviceID"}],"properties":[{"name":"FreeSpace"},{"name":"Size"}]}]`

	testutils.FuncBenchmarkCollector(b, wmi.Name, wmi.NewWithFlags, func(app *kingpin.Application) {
		// The queries are only read by the application action, which runs on parse.
		_, err := app.Parse([]string{"--collector.wmi.queries", queries})
		require.NoError(b, err)
	})
}

func TestCollector(t *testing.T) {
	t.Parallel()

	testutils.TestCollector(t, wmi.New, &wmi.Config{
		Queries: []wmi.Query{
			{
				Name:  "os",
				Class: "Win32_OperatingSystem",
				Properties: []wmi.Property{
					{Name: "NumberOfProcesses"},
					{Name: "LastBootUpTime", Metric: "windows_wmi_os_last_boot_timestamp_seconds"},
				},
			},
			{
				Name:            "logical_disk",
				Namespace:       `root\CIMv2`,
				Class:           "Win32_LogicalDisk",
				Where:           "DriveType = 3",
				LabelProperties: []wmi.LabelProperty{{Name: "DeviceID", Label: "volume"}},
				Properties:      []wmi.Property{{Name: "FreeSpace"}, {Name: "Size"}},
			},
		},
	})
}

// newSession returns a live MI session. Build must be validated against a real
// session, because a nil session is rejected before the configuration is
// looked at.
func newSession(t *testing.T) *mi.Session {
	t.Helper()

	app, err := mi.ApplicationInitialize()
	require.NoError(t, err)

	t.Cleanup(func() { require.NoError(t, app.Close()) })

	session, err := app.NewSession(nil)
	require.NoError(t, err)

	t.Cleanup(func() { require.NoError(t, session.Close()) })

	return session
}

func TestCollectorBuildEmptyConfigWithoutSession(t *testing.T) {
	t.Parallel()

	require.NoError(t, wmi.New(nil).Build(slog.New(slog.DiscardHandler), nil))
	require.Error(t, wmi.New(&wmi.Config{Queries: []wmi.Query{
		{Name: "os", Class: "Win32_OperatingSystem", Properties: []wmi.Property{{Name: "NumberOfProcesses"}}},
	}}).Build(slog.New(slog.DiscardHandler), nil))
}

func TestCollectorBuildErrors(t *testing.T) {
	t.Parallel()

	session := newSession(t)

	valid := func(mutate func(*wmi.Query)) []wmi.Query {
		query := wmi.Query{
			Name:            "disk",
			Class:           "Win32_LogicalDisk",
			LabelProperties: []wmi.LabelProperty{{Name: "DeviceID"}},
			Properties:      []wmi.Property{{Name: "FreeSpace"}},
		}

		mutate(&query)

		return []wmi.Query{query}
	}

	// Guard against the table passing for the wrong reason.
	require.NoError(t, wmi.New(&wmi.Config{Queries: valid(func(*wmi.Query) {})}).Build(slog.New(slog.DiscardHandler), session))

	for _, tc := range []struct {
		name    string
		queries []wmi.Query
		err     string
	}{
		{
			name:    "missing name",
			queries: valid(func(q *wmi.Query) { q.Name = "" }),
			err:     "query name is required",
		},
		{
			name: "duplicate name",
			queries: append(
				valid(func(*wmi.Query) {}),
				valid(func(q *wmi.Query) { q.Class = "Win32_Volume" })...,
			),
			err: "query disk: name is duplicated",
		},
		{
			name: "names differing only by case",
			queries: append(
				valid(func(*wmi.Query) {}),
				valid(func(q *wmi.Query) { q.Name = "DISK" })...,
			),
			err: "query DISK: name produces the same metric names as query disk",
		},
		{
			name: "names differing only by special characters",
			queries: append(
				valid(func(q *wmi.Query) { q.Name = "my_disk" }),
				valid(func(q *wmi.Query) { q.Name = "my-disk" })...,
			),
			err: "query my-disk: name produces the same metric names as query my_disk",
		},
		{
			name:    "missing class",
			queries: valid(func(q *wmi.Query) { q.Class = "" }),
			err:     `class "" must be a valid WMI class name`,
		},
		{
			name:    "class with WQL injection",
			queries: valid(func(q *wmi.Query) { q.Class = "Win32_LogicalDisk WHERE 1=1" }),
			err:     "must be a valid WMI class name",
		},
		{
			name:    "no properties",
			queries: valid(func(q *wmi.Query) { q.Properties = nil }),
			err:     "no properties configured",
		},
		{
			name:    "invalid property name",
			queries: valid(func(q *wmi.Query) { q.Properties = []wmi.Property{{Name: "Free Space"}} }),
			err:     `property "Free Space" must be a valid WMI property name`,
		},
		{
			name: "duplicate property differing only by case",
			queries: valid(func(q *wmi.Query) {
				q.Properties = []wmi.Property{{Name: "FreeSpace", Metric: "windows_a"}, {Name: "freespace", Metric: "windows_b"}}
			}),
			err: `property "freespace" is duplicated`,
		},
		{
			name:    "invalid type",
			queries: valid(func(q *wmi.Query) { q.Properties[0].Type = "histogram" }),
			err:     `invalid type "histogram"`,
		},
		{
			name:    "invalid label property name",
			queries: valid(func(q *wmi.Query) { q.LabelProperties = []wmi.LabelProperty{{Name: ""}} }),
			err:     `label property "" must be a valid WMI property name`,
		},
		{
			name:    "reserved label name",
			queries: valid(func(q *wmi.Query) { q.LabelProperties[0].Label = "__volume" }),
			err:     `"__volume"`,
		},
		{
			name:    "reserved constant label name",
			queries: valid(func(q *wmi.Query) { q.Properties[0].Labels = map[string]string{"__foo": "bar"} }),
			err:     `"__foo"`,
		},
		{
			name: "duplicate label",
			queries: valid(func(q *wmi.Query) {
				q.LabelProperties = []wmi.LabelProperty{{Name: "DeviceID", Label: "volume"}, {Name: "VolumeName", Label: "volume"}}
			}),
			err: `label "volume" is duplicated`,
		},
		{
			name:    "constant label collides with label property",
			queries: valid(func(q *wmi.Query) { q.Properties[0].Labels = map[string]string{"deviceid": "x"} }),
			err:     `constant label "deviceid" is already set by a label property`,
		},
		{
			name: "properties producing identical series",
			queries: valid(func(q *wmi.Query) {
				q.Properties = []wmi.Property{{Name: "FreeSpace", Metric: "windows_disk"}, {Name: "Size", Metric: "windows_disk"}}
			}),
			err: `properties "FreeSpace" and "Size" produce identical series windows_disk`,
		},
		{
			name: "queries without label properties producing identical series",
			queries: append(
				valid(func(q *wmi.Query) {
					q.LabelProperties = nil
					q.Properties[0].Metric = "windows_disk"
				}),
				valid(func(q *wmi.Query) {
					q.Name = "volume"
					q.LabelProperties = nil
					q.Properties[0].Metric = "windows_disk"
				})...,
			),
			err: `query volume: property "FreeSpace" produces the same series windows_disk as query disk`,
		},
		{
			name:    "explicit metric reserved by the collector",
			queries: valid(func(q *wmi.Query) { q.Properties[0].Metric = "windows_wmi_query_success" }),
			err:     `metric "windows_wmi_query_success" is reserved`,
		},
		{
			name: "derived metric reserved by the collector",
			queries: valid(func(q *wmi.Query) {
				q.Name = "query"
				q.Properties = []wmi.Property{{Name: "Duration_Seconds"}}
			}),
			err: `metric "windows_wmi_query_duration_seconds" is reserved`,
		},
		{
			name: "shared metric name with inconsistent help text",
			queries: valid(func(q *wmi.Query) {
				q.Properties = []wmi.Property{
					{Name: "FreeSpace", Metric: "windows_disk", Help: "a", Labels: map[string]string{"kind": "free"}},
					{Name: "Size", Metric: "windows_disk", Help: "b", Labels: map[string]string{"kind": "size"}},
				}
			}),
			err: `metric "windows_disk" must have the same help text, type and label names`,
		},
		{
			name: "shared metric name with inconsistent label names",
			queries: append(
				valid(func(q *wmi.Query) { q.Properties[0].Metric = "windows_disk" }),
				valid(func(q *wmi.Query) {
					q.Name = "volume"
					q.LabelProperties = nil
					q.Properties[0].Metric = "windows_disk"
				})...,
			),
			err: `metric "windows_disk" must have the same help text, type and label names`,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			c := wmi.New(&wmi.Config{Queries: tc.queries})
			require.ErrorContains(t, c.Build(slog.New(slog.DiscardHandler), session), tc.err)
		})
	}
}

// TestCollectorBuildValid covers configurations close to the rejected ones in
// TestCollectorBuildErrors that must still be accepted.
func TestCollectorBuildValid(t *testing.T) {
	t.Parallel()

	session := newSession(t)

	for _, tc := range []struct {
		name    string
		queries []wmi.Query
	}{
		{
			name: "whitespace-only where",
			queries: []wmi.Query{
				{Name: "os", Class: "Win32_OperatingSystem", Where: "  \t ", Properties: []wmi.Property{{Name: "Primary"}}},
			},
		},
		{
			name: "queries with label properties sharing a metric",
			queries: []wmi.Query{
				{Name: "fixed", Class: "Win32_LogicalDisk", Where: "DriveType = 3", LabelProperties: []wmi.LabelProperty{{Name: "DeviceID"}}, Properties: []wmi.Property{{Name: "Size", Metric: "windows_disk_size_bytes"}}},
				{Name: "removable", Class: "Win32_LogicalDisk", Where: "DriveType = 2", LabelProperties: []wmi.LabelProperty{{Name: "DeviceID"}}, Properties: []wmi.Property{{Name: "Size", Metric: "windows_disk_size_bytes"}}},
			},
		},
		{
			name: "queries without label properties sharing a metric with distinct constant labels",
			queries: []wmi.Query{
				{Name: "a", Class: "Win32_OperatingSystem", Properties: []wmi.Property{{Name: "Primary", Metric: "windows_os_primary", Labels: map[string]string{"query": "a"}}}},
				{Name: "b", Class: "Win32_OperatingSystem", Properties: []wmi.Property{{Name: "Primary", Metric: "windows_os_primary", Labels: map[string]string{"query": "b"}}}},
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			c := wmi.New(&wmi.Config{Queries: tc.queries})
			require.NoError(t, c.Build(slog.New(slog.DiscardHandler), session))
		})
	}
}
