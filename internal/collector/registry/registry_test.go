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

package registry_test

import (
	"log/slog"
	"testing"

	"github.com/alecthomas/kingpin/v2"
	"github.com/prometheus-community/windows_exporter/internal/collector/registry"
	"github.com/prometheus-community/windows_exporter/internal/utils/testutils"
	"github.com/stretchr/testify/require"
)

func BenchmarkCollector(b *testing.B) {
	keys := `[{"name":"windows_nt","key":"HKLM\\SOFTWARE\\Microsoft\\Windows NT\\CurrentVersion","values":[{"name":"CurrentMajorVersionNumber"}]}]`

	testutils.FuncBenchmarkCollector(b, registry.Name, registry.NewWithFlags, func(app *kingpin.Application) {
		app.GetFlag("collector.registry.keys").StringVar(&keys)
	})
}

func TestCollector(t *testing.T) {
	t.Parallel()

	testutils.TestCollector(t, registry.New, &registry.Config{
		Keys: []registry.Key{
			{
				Name:   "windows_nt",
				Key:    `HKLM\SOFTWARE\Microsoft\Windows NT\CurrentVersion`,
				Values: []registry.Value{{Name: "CurrentMajorVersionNumber"}},
			},
			{
				Name:   "memory_management",
				Key:    `HKEY_LOCAL_MACHINE\SYSTEM\CurrentControlSet\Control\Session Manager\Memory Management`,
				Values: []registry.Value{{Name: "ClearPageFileAtShutdown", Type: "gauge"}},
			},
		},
	})
}

func TestCollectorBuildErrors(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name   string
		config registry.Config
	}{
		{
			name:   "unknown hive",
			config: registry.Config{Keys: []registry.Key{{Key: `BOGUS\Foo`}}},
		},
		{
			name:   "empty key",
			config: registry.Config{Keys: []registry.Key{{Key: ""}}},
		},
		{
			name: "key with no values",
			config: registry.Config{Keys: []registry.Key{
				{Name: "windows_nt", Key: `HKLM\SOFTWARE\Microsoft\Windows NT\CurrentVersion`},
			}},
		},
		{
			name: "duplicate key",
			config: registry.Config{Keys: []registry.Key{
				{Name: "nt_a", Key: `HKLM\SOFTWARE\Microsoft\Windows NT\CurrentVersion`},
				{Name: "nt_b", Key: `HKEY_LOCAL_MACHINE/SOFTWARE/Microsoft/Windows NT/CurrentVersion`},
			}},
		},
		{
			name: "duplicate key differing only by case",
			config: registry.Config{Keys: []registry.Key{
				{Name: "nt_a", Key: `HKLM\SOFTWARE\Microsoft\Windows NT\CurrentVersion`},
				{Name: "nt_b", Key: `hklm\software\microsoft\windows nt\currentversion`},
			}},
		},
		{
			name: "duplicate value differing only by case",
			config: registry.Config{Keys: []registry.Key{
				{
					Name:   "windows_nt",
					Key:    `HKLM\SOFTWARE\Microsoft\Windows NT\CurrentVersion`,
					Values: []registry.Value{{Name: "CurrentMajorVersionNumber"}, {Name: "currentmajorversionnumber"}},
				},
			}},
		},
		{
			name: "missing value name",
			config: registry.Config{Keys: []registry.Key{
				{
					Name:   "windows_nt",
					Key:    `HKLM\SOFTWARE\Microsoft\Windows NT\CurrentVersion`,
					Values: []registry.Value{{Name: ""}},
				},
			}},
		},
		{
			name: "invalid value type",
			config: registry.Config{Keys: []registry.Key{
				{
					Name:   "windows_nt",
					Key:    `HKLM\SOFTWARE\Microsoft\Windows NT\CurrentVersion`,
					Values: []registry.Value{{Name: "CurrentMajorVersionNumber", Type: "histogram"}},
				},
			}},
		},
		{
			name: "shared metric name with inconsistent help text",
			config: registry.Config{Keys: []registry.Key{
				{
					Name: "windows_nt",
					Key:  `HKLM\SOFTWARE\Microsoft\Windows NT\CurrentVersion`,
					Values: []registry.Value{
						{Name: "CurrentMajorVersionNumber", Metric: "windows_version", Help: "version help a"},
						{Name: "CurrentMinorVersionNumber", Metric: "windows_version", Help: "version help b"},
					},
				},
			}},
		},
		{
			name: "shared metric name with inconsistent type",
			config: registry.Config{Keys: []registry.Key{
				{
					Name: "windows_nt",
					Key:  `HKLM\SOFTWARE\Microsoft\Windows NT\CurrentVersion`,
					Values: []registry.Value{
						{Name: "CurrentMajorVersionNumber", Metric: "windows_version", Type: "gauge"},
						{Name: "CurrentMinorVersionNumber", Metric: "windows_version", Type: "counter"},
					},
				},
			}},
		},
		{
			name: "missing name",
			config: registry.Config{Keys: []registry.Key{
				{
					Key:    `HKLM\SOFTWARE\Microsoft\Windows NT\CurrentVersion`,
					Values: []registry.Value{{Name: "CurrentMajorVersionNumber"}},
				},
			}},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			c := registry.New(&tc.config)
			require.Error(t, c.Build(slog.New(slog.DiscardHandler), nil))
		})
	}
}
