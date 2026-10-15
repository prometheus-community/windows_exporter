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

package collector

import (
	"maps"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/alecthomas/kingpin/v2"
	"github.com/prometheus-community/windows_exporter/internal/collector/ad"
	"github.com/prometheus-community/windows_exporter/internal/collector/adcs"
	"github.com/prometheus-community/windows_exporter/internal/collector/adfs"
	"github.com/prometheus-community/windows_exporter/internal/collector/cache"
	"github.com/prometheus-community/windows_exporter/internal/collector/container"
	"github.com/prometheus-community/windows_exporter/internal/collector/cpu"
	"github.com/prometheus-community/windows_exporter/internal/collector/cpu_info"
	"github.com/prometheus-community/windows_exporter/internal/collector/dfsr"
	"github.com/prometheus-community/windows_exporter/internal/collector/dhcp"
	"github.com/prometheus-community/windows_exporter/internal/collector/diskdrive"
	"github.com/prometheus-community/windows_exporter/internal/collector/dmi"
	"github.com/prometheus-community/windows_exporter/internal/collector/dns"
	"github.com/prometheus-community/windows_exporter/internal/collector/exchange"
	"github.com/prometheus-community/windows_exporter/internal/collector/file"
	"github.com/prometheus-community/windows_exporter/internal/collector/fsrmquota"
	"github.com/prometheus-community/windows_exporter/internal/collector/gpu"
	"github.com/prometheus-community/windows_exporter/internal/collector/hyperv"
	"github.com/prometheus-community/windows_exporter/internal/collector/iis"
	"github.com/prometheus-community/windows_exporter/internal/collector/license"
	"github.com/prometheus-community/windows_exporter/internal/collector/logical_disk"
	"github.com/prometheus-community/windows_exporter/internal/collector/memory"
	"github.com/prometheus-community/windows_exporter/internal/collector/mscluster"
	"github.com/prometheus-community/windows_exporter/internal/collector/msmq"
	"github.com/prometheus-community/windows_exporter/internal/collector/mssql"
	"github.com/prometheus-community/windows_exporter/internal/collector/net"
	"github.com/prometheus-community/windows_exporter/internal/collector/netframework"
	"github.com/prometheus-community/windows_exporter/internal/collector/nps"
	"github.com/prometheus-community/windows_exporter/internal/collector/os"
	"github.com/prometheus-community/windows_exporter/internal/collector/pagefile"
	"github.com/prometheus-community/windows_exporter/internal/collector/performancecounter"
	"github.com/prometheus-community/windows_exporter/internal/collector/physical_disk"
	"github.com/prometheus-community/windows_exporter/internal/collector/printer"
	"github.com/prometheus-community/windows_exporter/internal/collector/process"
	"github.com/prometheus-community/windows_exporter/internal/collector/registry"
	"github.com/prometheus-community/windows_exporter/internal/collector/remote_fx"
	"github.com/prometheus-community/windows_exporter/internal/collector/scheduled_task"
	"github.com/prometheus-community/windows_exporter/internal/collector/service"
	"github.com/prometheus-community/windows_exporter/internal/collector/smb"
	"github.com/prometheus-community/windows_exporter/internal/collector/smbclient"
	"github.com/prometheus-community/windows_exporter/internal/collector/smtp"
	"github.com/prometheus-community/windows_exporter/internal/collector/storage_spaces"
	"github.com/prometheus-community/windows_exporter/internal/collector/system"
	"github.com/prometheus-community/windows_exporter/internal/collector/tcp"
	"github.com/prometheus-community/windows_exporter/internal/collector/terminal_services"
	"github.com/prometheus-community/windows_exporter/internal/collector/textfile"
	"github.com/prometheus-community/windows_exporter/internal/collector/time"
	"github.com/prometheus-community/windows_exporter/internal/collector/udp"
	"github.com/prometheus-community/windows_exporter/internal/collector/update"
	"github.com/prometheus-community/windows_exporter/internal/collector/vmware"
	"github.com/prometheus-community/windows_exporter/internal/collector/wmi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestRegistrations verifies that every place that lists the collectors agrees
// with the registrations: the flag builders, NewWithConfig, the fields and
// YAML keys of Config, ConfigDefaults and the collector documentation.
func TestRegistrations(t *testing.T) {
	t.Parallel()

	names := make([]string, 0, len(registrations))
	for _, r := range registrations {
		require.NotContains(t, names, r.name, "collector %s is registered twice", r.name)
		names = append(names, r.name)
	}

	slices.Sort(names)

	t.Run("flag builders", func(t *testing.T) {
		t.Parallel()

		require.Equal(t, names, Available())

		app := kingpin.New("test", "test")
		for name, builder := range BuildersWithFlags {
			assert.Equal(t, name, builder(app).GetName(), "flag builder of collector %s", name)
		}
	})

	t.Run("NewWithConfig", func(t *testing.T) {
		t.Parallel()

		collection := NewWithConfig(ConfigDefaults)
		require.Equal(t, names, slices.Sorted(maps.Keys(collection.collectors)))

		for name, c := range collection.collectors {
			assert.Equal(t, name, c.GetName(), "NewWithConfig collector %s", name)
		}
	})

	t.Run("Config fields", func(t *testing.T) {
		t.Parallel()

		var config Config

		value := reflect.ValueOf(&config).Elem()

		// Fields of size zero can share an address, so the type is part of
		// the key. Every collector has its own configuration type.
		type fieldKey struct {
			addr uintptr
			typ  reflect.Type
		}

		keyOf := func(ptr reflect.Value) fieldKey {
			return fieldKey{addr: ptr.Pointer(), typ: ptr.Type().Elem()}
		}

		fields := map[fieldKey]reflect.StructField{}
		for i := range value.NumField() {
			fields[keyOf(value.Field(i).Addr())] = value.Type().Field(i)
		}

		for _, r := range registrations {
			key := keyOf(reflect.ValueOf(r.field(&config)))

			field, ok := fields[key]
			if !assert.True(t, ok, "collector %s has no field in Config", r.name) {
				continue
			}

			configKey, _, _ := strings.Cut(field.Tag.Get("yaml"), ",")
			assert.Equal(t, r.name, configKey, "YAML key of Config.%s", field.Name)

			delete(fields, key)
		}

		for _, field := range fields {
			assert.Fail(t, "Config field without collector registration", "Config.%s", field.Name)
		}
	})

	t.Run("ConfigDefaults", func(t *testing.T) {
		t.Parallel()

		// The defaults of every collector, listed independently of the
		// registrations.
		defaults := map[string]any{
			ad.Name: ad.ConfigDefaults, adcs.Name: adcs.ConfigDefaults, adfs.Name: adfs.ConfigDefaults,
			cache.Name: cache.ConfigDefaults, container.Name: container.ConfigDefaults, cpu.Name: cpu.ConfigDefaults,
			cpu_info.Name: cpu_info.ConfigDefaults, dfsr.Name: dfsr.ConfigDefaults, dhcp.Name: dhcp.ConfigDefaults,
			diskdrive.Name: diskdrive.ConfigDefaults, dmi.Name: dmi.ConfigDefaults, dns.Name: dns.ConfigDefaults,
			exchange.Name: exchange.ConfigDefaults, file.Name: file.ConfigDefaults, fsrmquota.Name: fsrmquota.ConfigDefaults,
			gpu.Name: gpu.ConfigDefaults, hyperv.Name: hyperv.ConfigDefaults, iis.Name: iis.ConfigDefaults,
			license.Name: license.ConfigDefaults, logical_disk.Name: logical_disk.ConfigDefaults, memory.Name: memory.ConfigDefaults,
			mscluster.Name: mscluster.ConfigDefaults, msmq.Name: msmq.ConfigDefaults, mssql.Name: mssql.ConfigDefaults,
			net.Name: net.ConfigDefaults, netframework.Name: netframework.ConfigDefaults, nps.Name: nps.ConfigDefaults,
			os.Name: os.ConfigDefaults, pagefile.Name: pagefile.ConfigDefaults, performancecounter.Name: performancecounter.ConfigDefaults,
			physical_disk.Name: physical_disk.ConfigDefaults, printer.Name: printer.ConfigDefaults, process.Name: process.ConfigDefaults,
			registry.Name: registry.ConfigDefaults, remote_fx.Name: remote_fx.ConfigDefaults, scheduled_task.Name: scheduled_task.ConfigDefaults,
			service.Name: service.ConfigDefaults, smb.Name: smb.ConfigDefaults, smbclient.Name: smbclient.ConfigDefaults,
			smtp.Name: smtp.ConfigDefaults, storage_spaces.Name: storage_spaces.ConfigDefaults, system.Name: system.ConfigDefaults,
			tcp.Name: tcp.ConfigDefaults, terminal_services.Name: terminal_services.ConfigDefaults, textfile.Name: textfile.ConfigDefaults,
			time.Name: time.ConfigDefaults, udp.Name: udp.ConfigDefaults, update.Name: update.ConfigDefaults,
			vmware.Name: vmware.ConfigDefaults, wmi.Name: wmi.ConfigDefaults,
		}

		require.Len(t, defaults, len(registrations))

		for _, r := range registrations {
			want, ok := defaults[r.name]
			if !assert.True(t, ok, "collector %s has no defaults in the test", r.name) {
				continue
			}

			got := reflect.ValueOf(r.field(&ConfigDefaults)).Elem().Interface()
			assert.Equal(t, want, got, "ConfigDefaults of collector %s", r.name)
		}
	})

	t.Run("documentation", func(t *testing.T) {
		t.Parallel()

		files, err := filepath.Glob(filepath.Join("..", "..", "docs", "collector.*.md"))
		require.NoError(t, err)

		documented := make([]string, 0, len(files))

		for _, file := range files {
			documented = append(documented, strings.TrimSuffix(strings.TrimPrefix(filepath.Base(file), "collector."), ".md"))
		}

		require.ElementsMatch(t, names, documented, "docs/collector.<name>.md files")
	})
}
