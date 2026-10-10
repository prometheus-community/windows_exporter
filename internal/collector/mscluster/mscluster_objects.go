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

// beforeMinBuild selects how a field is published on Windows builds older than
// its objectField.minBuild.
type beforeMinBuild int

const (
	// omitIfMissing publishes a value that exists and omits a missing one
	// without an error. The property did not exist on older builds.
	omitIfMissing beforeMinBuild = iota
	// zeroIfMissing keeps the previous behavior for properties that the WMI
	// query did not select on older builds and therefore published as 0.
	zeroIfMissing
	// neverPublish keeps the previous behavior for properties that were not
	// published at all on older builds.
	neverPublish
)

// objectField maps a ClusAPI value of a cluster, group, node or network to the
// metric that previously published the WMI property of the same name.
type objectField struct {
	name string
	desc *prometheus.Desc
	// Signed fields are sint32 in WMI, so 0xFFFFFFFF is published as -1.
	signed bool
	// MinBuild is the first Windows build that has the property.
	minBuild uint16
	older    beforeMinBuild
}

// publishObjectFields publishes the fields of one object with its name as the
// only label. Missing values are omitted and reported, never published as 0.
func publishObjectFields(ch chan<- prometheus.Metric, kind string, object clusapi.Object, fields []objectField, build uint16, resultErr error) error {
	for _, field := range fields {
		raw, exists := object.Values[field.name]
		older := build < field.minBuild

		switch {
		case older && field.older == neverPublish:
			continue
		case exists:
		case older && field.older == zeroIfMissing:
			raw = 0
		case older:
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
