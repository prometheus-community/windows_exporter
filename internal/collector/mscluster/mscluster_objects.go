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

package mscluster

import (
	"errors"
	"fmt"

	"github.com/prometheus-community/windows_exporter/internal/headers/clusapi"
	"github.com/prometheus/client_golang/prometheus"
)

// objectField maps a ClusAPI value of a group, node or network to the metric
// that previously published the WMI property of the same name.
type objectField struct {
	name string
	desc *prometheus.Desc
	// Signed fields are sint32 in WMI, so 0xFFFFFFFF is published as -1.
	signed bool
	// MinBuild is the first Windows build that has the property. On older
	// builds a missing value is omitted without an error.
	minBuild uint16
	// ZeroBeforeMinBuild keeps the previous behavior for properties that the
	// WMI query did not select on older builds and therefore published as 0.
	zeroBeforeMinBuild bool
}

// publishObjectFields publishes the fields of one object with its name as the
// only label. Missing values are omitted and reported, never published as 0.
func publishObjectFields(ch chan<- prometheus.Metric, kind string, object clusapi.Object, fields []objectField, build uint16, resultErr error) error {
	for _, field := range fields {
		raw, exists := object.Values[field.name]

		switch {
		case exists:
		case build < field.minBuild && field.zeroBeforeMinBuild:
			raw = 0
		case build < field.minBuild:
			continue
		default:
			resultErr = errors.Join(resultErr, fmt.Errorf("%s %q: missing property %s", kind, object.Name, field.name))

			continue
		}

		value := float64(raw)
		if field.signed {
			value = float64(int32(raw))
		}

		ch <- prometheus.MustNewConstMetric(field.desc, prometheus.GaugeValue, value, object.Name)
	}

	return resultErr
}
