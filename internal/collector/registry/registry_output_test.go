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
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/prometheus-community/windows_exporter/internal/collector/registry"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/stretchr/testify/require"
	winregistry "golang.org/x/sys/windows/registry"
)

// collectorAdapter bridges registry.Collector's Collect(ch, time.Duration) error
// signature to the prometheus.Collector interface, mirroring the adapter in the
// performancecounter collector's tests.
type collectorAdapter struct {
	registry.Collector
}

// Describe implements the prometheus.Collector interface.
func (collectorAdapter) Describe(chan<- *prometheus.Desc) {}

// Collect implements the prometheus.Collector interface.
//
// Unlike the performancecounter adapter, the Collect error is intentionally
// discarded rather than panicked on: a missing key is an expected, non-fatal
// condition reported via key_success=0, and this test deliberately configures a
// bogus key, so Collect returns a non-nil joined error while still emitting
// every metric.
func (a collectorAdapter) Collect(ch chan<- prometheus.Metric) {
	_ = a.Collector.Collect(ch, 0)
}

// TestCollectorMetrics asserts the full value→Desc→metric wiring end to end by
// reading known REG_DWORD/REG_QWORD values from a throwaway HKCU key. It exercises
// the auto-generated metric name, an explicit metric: override, gauge vs counter
// types, constant labels, exact 32- and 64-bit values, and the key_success 1-vs-0
// semantics (a missing key fails just that key, not the whole scrape). Because the
// test controls the input values, it asserts exact values rather than the numeric
// patterns the performancecounter test must use. Using HKCU keeps it admin-free
// and reliable in CI.
func TestCollectorMetrics(t *testing.T) {
	suffix := time.Now().UnixNano()

	sub := fmt.Sprintf(`Software\windows_exporter_registry_test_%d`, suffix)
	keyPath := `HKCU\` + sub

	k, _, err := winregistry.CreateKey(winregistry.CURRENT_USER, sub, winregistry.SET_VALUE)
	require.NoError(t, err)

	// REG_DWORD fits in 32 bits; REG_QWORD is deliberately > math.MaxUint32 to prove
	// 64-bit handling. Both are well under 2^53, so they convert to float64 exactly.
	require.NoError(t, k.SetDWordValue("TestDword", 1234))
	require.NoError(t, k.SetQWordValue("TestQword", 5_000_000_000))
	require.NoError(t, k.Close())

	t.Cleanup(func() {
		_ = winregistry.DeleteKey(winregistry.CURRENT_USER, sub)
	})

	// A second key that does not exist exercises graceful per-key failure: the
	// collector reports key_success=0 for it instead of aborting the scrape.
	missingSub := fmt.Sprintf(`Software\windows_exporter_missing_%d`, suffix)
	missingPath := `HKCU\` + missingSub

	c := registry.New(&registry.Config{
		Keys: []registry.Key{
			{
				Name: "testgroup",
				Key:  keyPath,
				Values: []registry.Value{
					// Auto-named gauge → windows_registry_testgroup_testdword.
					{Name: "TestDword"},
					// Explicit metric: override + counter type + constant label.
					{
						Name:   "TestQword",
						Metric: "windows_registry_test_custom_metric",
						Type:   "counter",
						Labels: map[string]string{"foo": "bar"},
					},
				},
			},
			{
				Key: missingPath,
			},
		},
	})

	require.NoError(t, c.Build(slog.New(slog.DiscardHandler), nil))

	reg := prometheus.NewRegistry()
	reg.MustRegister(collectorAdapter{*c})

	rw := httptest.NewRecorder()
	promhttp.HandlerFor(reg, promhttp.HandlerOpts{ErrorHandling: promhttp.ContinueOnError}).ServeHTTP(rw, &http.Request{})
	got := rw.Body.String()

	require.NotEmpty(t, got)

	// key_success labels are the normalized, lowercased path (short hive +
	// backslashes); in the text exposition format the backslashes are escaped.
	// Metric families are emitted sorted by name, and metrics within the
	// key_success family sorted by label value, so "missing" (0) precedes
	// "registry_test" (1).
	successLabel := strings.ReplaceAll(strings.ToLower(keyPath), `\`, `\\`)
	missingLabel := strings.ReplaceAll(strings.ToLower(missingPath), `\`, `\\`)

	expected := fmt.Sprintf(`# HELP windows_registry_key_success Whether the registry key could be read successfully.
# TYPE windows_registry_key_success gauge
windows_registry_key_success{key="%s"} 0
windows_registry_key_success{key="%s"} 1
# HELP windows_registry_test_custom_metric windows_exporter: custom registry metric
# TYPE windows_registry_test_custom_metric counter
windows_registry_test_custom_metric{foo="bar"} 5e+09
# HELP windows_registry_testgroup_testdword windows_exporter: custom registry metric
# TYPE windows_registry_testgroup_testdword gauge
windows_registry_testgroup_testdword 1234
`, missingLabel, successLabel)

	// QuoteMeta escapes the backslashes/braces in the exposition so the exact
	// values, types, labels, and names are matched literally.
	require.Regexp(t, "^"+regexp.QuoteMeta(expected), got)
}
