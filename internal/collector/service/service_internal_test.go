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

package service

import (
	"fmt"
	"log/slog"
	"os"
	"regexp"
	"slices"
	"strings"
	"testing"
	"time"
	"unsafe"

	"github.com/prometheus-community/windows_exporter/internal/scrapecache"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/svc/mgr"
)

type serviceMetricCollector struct {
	t         assert.TestingT
	collector *Collector
}

func (c serviceMetricCollector) Describe(ch chan<- *prometheus.Desc) {
	prometheus.DescribeByCollect(c, ch)
}

func (c serviceMetricCollector) Collect(ch chan<- prometheus.Metric) {
	// Collect within a scrape of the caches, as the collection does.
	cacheScrape := scrapecache.Begin(c.collector)
	err := c.collector.Collect(ch, 0)
	cacheScrape.End(err)

	assert.NoError(c.t, err)
}

// serviceMetrics scrapes the collector and returns the value of each series of
// the service as "<metric>{<label values>}". It only reports failures to t, so
// that it can be retried inside EventuallyWithT.
func serviceMetrics(t assert.TestingT, c *Collector, name string) map[string]float64 {
	registry := prometheus.NewPedanticRegistry()
	// require must not be used inside EventuallyWithT.
	//nolint:testifylint
	if !assert.NoError(t, registry.Register(serviceMetricCollector{t: t, collector: c})) {
		return nil
	}

	families, err := registry.Gather()
	//nolint:testifylint
	if !assert.NoError(t, err) {
		return nil
	}

	metrics := make(map[string]float64)

	for _, family := range families {
		for _, metric := range family.GetMetric() {
			var (
				labels  []string
				matches bool
			)

			for _, label := range metric.GetLabel() {
				if label.GetName() == "service" {
					matches = label.GetValue() == name

					continue
				}

				labels = append(labels, label.GetName()+"="+label.GetValue())
			}

			if matches {
				metrics[fmt.Sprintf("%s{%s}", family.GetName(), strings.Join(labels, ","))] = metric.GetGauge().GetValue()
			}
		}
	}

	return metrics
}

// setStartTrigger gives a service a trigger that starts it once the first IP
// address becomes available. golang.org/x/sys/windows/svc/mgr can't set triggers.
func setStartTrigger(t *testing.T, service *mgr.Service) {
	t.Helper()

	// NETWORK_MANAGER_FIRST_IP_ADDRESS_ARRIVAL_GUID.
	subtype := windows.GUID{Data1: 0x4f27f2de, Data2: 0x14e2, Data3: 0x430b, Data4: [8]byte{0xa5, 0x49, 0x7c, 0xd4, 0x8c, 0xbc, 0x82, 0x45}}

	// SERVICE_TRIGGER.
	trigger := struct {
		triggerType uint32
		action      uint32
		subtype     *windows.GUID
		dataItems   uint32
		data        uintptr
	}{
		triggerType: 3, // SERVICE_TRIGGER_TYPE_IP_ADDRESS_AVAILABILITY
		action:      1, // SERVICE_TRIGGER_ACTION_SERVICE_START
		subtype:     &subtype,
	}

	// SERVICE_TRIGGER_INFO.
	info := struct {
		triggers uint32
		trigger  unsafe.Pointer
		reserved uintptr
	}{
		triggers: 1,
		trigger:  unsafe.Pointer(&trigger),
	}

	require.NoError(t, windows.ChangeServiceConfig2(service.Handle, windows.SERVICE_CONFIG_TRIGGER_INFO, (*byte)(unsafe.Pointer(&info))))
}

// TestCollectorPicksUpConfigChanges creates a service, changes its configuration
// through the service control manager and checks that the cached configuration
// is replaced. It needs an elevated process, which the Windows CI runners are.
func TestCollectorPicksUpConfigChanges(t *testing.T) {
	t.Parallel()

	required := slices.Contains(strings.Split(os.Getenv("WINDOWS_EXPORTER_TEST_COLLECTORS"), ","), Name)
	if !windows.GetCurrentProcessToken().IsElevated() {
		if required {
			t.Fatal("creating a test service needs an elevated process")
		}

		t.Skip("creating a test service needs an elevated process")
	}

	manager, err := mgr.Connect()
	require.NoError(t, err)
	t.Cleanup(func() { assert.NoError(t, manager.Disconnect()) })

	name := fmt.Sprintf("windows_exporter_test_%d", time.Now().UnixNano())

	// The service is never started.
	service, err := manager.CreateService(name, `C:\Windows\System32\cmd.exe`, mgr.Config{
		StartType:   mgr.StartManual,
		DisplayName: "before",
	})
	require.NoError(t, err)

	deleted := false

	t.Cleanup(func() {
		if !deleted {
			assert.NoError(t, service.Delete())
			assert.NoError(t, service.Close())
		}
	})

	c := New(&Config{
		ServiceInclude: regexp.MustCompile("^" + regexp.QuoteMeta(name) + "$"),
		ServiceExclude: regexp.MustCompile("^$"),
	})
	require.NoError(t, c.Build(slog.New(slog.DiscardHandler), nil))
	t.Cleanup(func() { assert.NoError(t, c.Close()) })
	require.True(t, c.configCache.enabled(), "the configuration cache must be enabled")

	// A configuration is only cached once its key is older than a clock tick.
	time.Sleep(configCacheMinKeyAge + 100*time.Millisecond)

	metrics := serviceMetrics(t, c, name)
	assert.InDelta(t, 1.0, metrics["windows_service_start_mode{start_mode=manual}"], 0)
	assert.Contains(t, metrics, `windows_service_info{delayed_start=false,display_name=before,path_name=C:\Windows\System32\cmd.exe,run_as=LocalSystem,trigger_start=false}`)

	_, _, ok := c.configCache.get(name)
	require.True(t, ok, "the configuration must be cached after a scrape")

	config, err := service.Config()
	require.NoError(t, err)

	// The service is never started. It is switched back to manual start below,
	// so that a killed test doesn't leave an auto-start service behind.
	config.StartType = mgr.StartAutomatic
	config.DelayedAutoStart = true
	config.DisplayName = "after"
	require.NoError(t, service.UpdateConfig(config))

	// The change notification is delivered asynchronously.
	require.EventuallyWithT(t, func(collect *assert.CollectT) {
		metrics := serviceMetrics(collect, c, name)
		assert.InDelta(collect, 1.0, metrics["windows_service_start_mode{start_mode=auto}"], 0)
		assert.InDelta(collect, 0.0, metrics["windows_service_start_mode{start_mode=manual}"], 0)
		assert.Contains(collect, metrics, `windows_service_info{delayed_start=true,display_name=after,path_name=C:\Windows\System32\cmd.exe,run_as=LocalSystem,trigger_start=false}`)
	}, 5*time.Second, 50*time.Millisecond)

	// Windows ignores the delayed auto-start flag of a manual service, and so does the label.
	config.StartType = mgr.StartManual
	require.NoError(t, service.UpdateConfig(config))

	require.EventuallyWithT(t, func(collect *assert.CollectT) {
		metrics := serviceMetrics(collect, c, name)
		assert.InDelta(collect, 1.0, metrics["windows_service_start_mode{start_mode=manual}"], 0)
		assert.Contains(collect, metrics, `windows_service_info{delayed_start=false,display_name=after,path_name=C:\Windows\System32\cmd.exe,run_as=LocalSystem,trigger_start=false}`)
	}, 5*time.Second, 50*time.Millisecond)

	// Cache the changed configuration, so that the trigger must invalidate it.
	time.Sleep(configCacheMinKeyAge + 100*time.Millisecond)
	serviceMetrics(t, c, name)

	_, _, ok = c.configCache.get(name)
	require.True(t, ok, "the changed configuration must be cached after a scrape")

	// Triggers are stored in the TriggerInfo subkey of the service key. Creating
	// it changes the last write time of the service key.
	setStartTrigger(t, service)

	require.EventuallyWithT(t, func(collect *assert.CollectT) {
		metrics := serviceMetrics(collect, c, name)
		assert.Contains(collect, metrics, `windows_service_info{delayed_start=false,display_name=after,path_name=C:\Windows\System32\cmd.exe,run_as=LocalSystem,trigger_start=true}`)
	}, 5*time.Second, 50*time.Millisecond)

	// The service control manager removes the service once the last handle is closed.
	require.NoError(t, service.Delete())

	deleted = true

	require.NoError(t, service.Close())

	require.EventuallyWithT(t, func(collect *assert.CollectT) {
		assert.Empty(collect, serviceMetrics(collect, c, name))

		_, _, ok := c.configCache.get(name)
		assert.False(collect, ok, "the configuration of a deleted service must be dropped")
	}, 5*time.Second, 50*time.Millisecond)
}

// TestServiceTriggerInfoLayout checks the mirror of SERVICE_TRIGGER_INFO
// (winsvc.h): DWORD cTriggers, PSERVICE_TRIGGER pTriggers and PBYTE pReserved.
// The pointers are pointer-aligned, so cTriggers is padded to a pointer word.
func TestServiceTriggerInfoLayout(t *testing.T) {
	t.Parallel()

	var info serviceTriggerInfo

	pointerSize := unsafe.Sizeof(uintptr(0))

	assert.Equal(t, uintptr(0), unsafe.Offsetof(info.triggers))
	assert.Equal(t, pointerSize, unsafe.Offsetof(info.trigger))
	assert.Equal(t, 3*pointerSize, unsafe.Sizeof(info))
}

// TestQueryAllServicesResumes enumerates with a buffer that holds only a few
// services, so that EnumServicesStatusEx returns ERROR_MORE_DATA with a part
// of the services several times, and compares the result with one call.
func TestQueryAllServicesResumes(t *testing.T) {
	t.Parallel()

	c := New(&Config{
		ServiceInclude: regexp.MustCompile(".+"),
		ServiceExclude: regexp.MustCompile("^$"),
	})
	require.NoError(t, c.Build(slog.New(slog.DiscardHandler), nil))
	t.Cleanup(func() { assert.NoError(t, c.Close()) })

	names := func(entries []serviceEntry) []string {
		result := make([]string, 0, len(entries))
		for _, entry := range entries {
			result = append(result, entry.name)
		}

		slices.Sort(result)

		return slices.Compact(result)
	}

	all, err := c.queryAllServices()
	require.NoError(t, err)
	require.NotEmpty(t, all)

	// A service can be added or removed between the two enumerations.
	c.queryAllServicesBuffer = make([]byte, 2048)

	resumed, err := c.queryAllServices()
	require.NoError(t, err)
	require.Len(t, c.queryAllServicesBuffer, 2048, "a buffer that holds a service is not grown")
	require.InDelta(t, len(all), len(resumed), 5)
	require.Len(t, names(resumed), len(resumed), "no service is returned twice")

	// A buffer that holds no service is grown.
	c.queryAllServicesBuffer = make([]byte, 16)

	grown, err := c.queryAllServices()
	require.NoError(t, err)
	require.InDelta(t, len(all), len(grown), 5)
	require.Greater(t, len(c.queryAllServicesBuffer), 16)
}

// TestCollectorPublishesServiceWithDeniedConfig creates a service whose DACL
// denies SERVICE_QUERY_CONFIG to administrators. Its state must still be
// published. It needs an elevated process, which the Windows CI runners are.
func TestCollectorPublishesServiceWithDeniedConfig(t *testing.T) {
	t.Parallel()

	required := slices.Contains(strings.Split(os.Getenv("WINDOWS_EXPORTER_TEST_COLLECTORS"), ","), Name)
	if !windows.GetCurrentProcessToken().IsElevated() {
		if required {
			t.Fatal("creating a test service needs an elevated process")
		}

		t.Skip("creating a test service needs an elevated process")
	}

	manager, err := mgr.Connect()
	require.NoError(t, err)
	t.Cleanup(func() { assert.NoError(t, manager.Disconnect()) })

	name := fmt.Sprintf("windows_exporter_test_denied_%d", time.Now().UnixNano())

	// The service is never started.
	service, err := manager.CreateService(name, `C:\Windows\System32\cmd.exe`, mgr.Config{
		StartType:   mgr.StartManual,
		DisplayName: "denied",
	})
	require.NoError(t, err)
	t.Cleanup(func() {
		assert.NoError(t, service.Delete())
		assert.NoError(t, service.Close())
	})

	// Deny SERVICE_QUERY_CONFIG (CC) to administrators, allow everything else.
	sd, err := windows.SecurityDescriptorFromString("D:(D;;CC;;;BA)(A;;GA;;;BA)(A;;GA;;;SY)")
	require.NoError(t, err)

	dacl, _, err := sd.DACL()
	require.NoError(t, err)
	require.NoError(t, windows.SetSecurityInfo(service.Handle, windows.SE_SERVICE, windows.DACL_SECURITY_INFORMATION, nil, nil, dacl, nil))

	c := New(&Config{
		ServiceInclude: regexp.MustCompile("^" + regexp.QuoteMeta(name) + "$"),
		ServiceExclude: regexp.MustCompile("^$"),
	})
	require.NoError(t, c.Build(slog.New(slog.DiscardHandler), nil))
	t.Cleanup(func() { assert.NoError(t, c.Close()) })

	_, available, err := c.queryServiceConfig(name)
	require.NoError(t, err)
	require.False(t, available, "the configuration cannot be read")

	metrics := serviceMetrics(t, c, name)
	assert.InDelta(t, 1.0, metrics["windows_service_state{state=stopped}"], 0)

	for series := range metrics {
		assert.NotContains(t, series, "windows_service_start_mode", "the start mode is part of the configuration")
	}
}
