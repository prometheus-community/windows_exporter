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
	}{
		{
			name:    "missing name",
			queries: valid(func(q *wmi.Query) { q.Name = "" }),
		},
		{
			name: "duplicate name",
			queries: append(
				valid(func(*wmi.Query) {}),
				valid(func(q *wmi.Query) { q.Class = "Win32_Volume" })...,
			),
		},
		{
			name:    "missing class",
			queries: valid(func(q *wmi.Query) { q.Class = "" }),
		},
		{
			name:    "class with WQL injection",
			queries: valid(func(q *wmi.Query) { q.Class = "Win32_LogicalDisk WHERE 1=1" }),
		},
		{
			name:    "no properties",
			queries: valid(func(q *wmi.Query) { q.Properties = nil }),
		},
		{
			name:    "invalid property name",
			queries: valid(func(q *wmi.Query) { q.Properties = []wmi.Property{{Name: "Free Space"}} }),
		},
		{
			name: "duplicate property differing only by case",
			queries: valid(func(q *wmi.Query) {
				q.Properties = []wmi.Property{{Name: "FreeSpace", Metric: "windows_a"}, {Name: "freespace", Metric: "windows_b"}}
			}),
		},
		{
			name:    "invalid type",
			queries: valid(func(q *wmi.Query) { q.Properties[0].Type = "histogram" }),
		},
		{
			name:    "invalid label property name",
			queries: valid(func(q *wmi.Query) { q.LabelProperties = []wmi.LabelProperty{{Name: ""}} }),
		},
		{
			name:    "reserved label name",
			queries: valid(func(q *wmi.Query) { q.LabelProperties[0].Label = "__volume" }),
		},
		{
			name:    "reserved constant label name",
			queries: valid(func(q *wmi.Query) { q.Properties[0].Labels = map[string]string{"__foo": "bar"} }),
		},
		{
			name: "duplicate label",
			queries: valid(func(q *wmi.Query) {
				q.LabelProperties = []wmi.LabelProperty{{Name: "DeviceID", Label: "volume"}, {Name: "VolumeName", Label: "volume"}}
			}),
		},
		{
			name:    "constant label collides with label property",
			queries: valid(func(q *wmi.Query) { q.Properties[0].Labels = map[string]string{"deviceid": "x"} }),
		},
		{
			name: "properties producing identical series",
			queries: valid(func(q *wmi.Query) {
				q.Properties = []wmi.Property{{Name: "FreeSpace", Metric: "windows_disk"}, {Name: "Size", Metric: "windows_disk"}}
			}),
		},
		{
			name: "shared metric name with inconsistent help text",
			queries: valid(func(q *wmi.Query) {
				q.Properties = []wmi.Property{
					{Name: "FreeSpace", Metric: "windows_disk", Help: "a", Labels: map[string]string{"kind": "free"}},
					{Name: "Size", Metric: "windows_disk", Help: "b", Labels: map[string]string{"kind": "size"}},
				}
			}),
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
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			c := wmi.New(&wmi.Config{Queries: tc.queries})
			require.Error(t, c.Build(slog.New(slog.DiscardHandler), session))
		})
	}
}
