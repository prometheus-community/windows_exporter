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

package dmi_test

import (
	"maps"
	"slices"
	"testing"

	"github.com/prometheus-community/windows_exporter/internal/collector/dmi"
	"github.com/prometheus-community/windows_exporter/internal/utils/testutils"
	"github.com/stretchr/testify/require"
)

func BenchmarkCollector(b *testing.B) {
	testutils.FuncBenchmarkCollector(b, dmi.Name, dmi.NewWithFlags)
}

func TestCollector(t *testing.T) {
	metrics := testutils.TestCollector(t, dmi.New, nil)

	metric := testutils.RequireFixtureMetric(t, metrics, dmi.Name, "windows_dmi_info", nil)
	if metric == nil {
		return
	}

	labels := make(map[string]string, len(metric.GetLabel()))
	for _, label := range metric.GetLabel() {
		labels[label.GetName()] = label.GetValue()
	}

	require.ElementsMatch(t, []string{
		"bios_date", "bios_release", "bios_vendor", "bios_version",
		"board_asset_tag", "board_name", "board_serial", "board_vendor", "board_version",
		"chassis_asset_tag", "chassis_serial", "chassis_vendor", "chassis_version",
		"product_family", "product_name", "product_serial", "product_sku",
		"product_uuid", "product_version", "system_vendor",
	}, slices.Collect(maps.Keys(labels)), "windows_dmi_info must carry node_exporter's node_dmi_info label names")
	require.NotEmpty(t, labels["product_uuid"], "SMBIOS system UUID is empty: %s", metric)
	require.NotEmpty(t, labels["bios_vendor"], "SMBIOS BIOS vendor is empty: %s", metric)
}
