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
	"encoding/binary"
	"errors"
	"fmt"

	"github.com/prometheus-community/windows_exporter/internal/headers/clusapi"
	"github.com/prometheus-community/windows_exporter/internal/metriccache"
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
	// Absent, if set, is published when the cluster service does not report
	// the property on any build. It reproduces a value that the WMI provider
	// synthesized; see the cluster field list for each case.
	absent *uint32
}

// resolve returns the raw value that publishObjectFields publishes for field,
// whether it publishes the field, and whether the field is missing: not
// reported by the cluster service although it should be, which is an error.
func (field objectField) resolve(object clusapi.Object, build uint16) (uint32, bool, bool) {
	raw, exists := object.Values[field.name]
	older := build < field.minBuild

	switch {
	case older && field.older == neverPublish:
		return 0, false, false
	case exists:
		return raw, true, false
	case field.absent != nil:
		return *field.absent, true, false
	case older && field.older == zeroIfMissing:
		return 0, true, false
	case older:
		return 0, false, false
	default:
		return 0, false, true
	}
}

// metric builds the metric of field for a raw value of the object name.
//
//nolint:ireturn // Const metrics have no exported type.
func (field objectField) metric(raw uint32, name string) prometheus.Metric {
	value := float64(raw)
	if field.signed {
		value = float64(int32(raw))
	}

	return prometheus.MustNewConstMetric(field.desc, prometheus.GaugeValue, value, name)
}

// publishObjectFields publishes the fields of one object with its name as the
// only label. Missing values are omitted and reported, never published as 0.
func publishObjectFields(ch chan<- prometheus.Metric, kind string, object clusapi.Object, fields []objectField, build uint16, resultErr error) error {
	for _, field := range fields {
		raw, publish, missing := field.resolve(object, build)
		if missing {
			resultErr = errors.Join(resultErr, fmt.Errorf("%s %q: missing property %s", kind, object.Name, field.name))

			continue
		}

		if publish {
			ch <- field.metric(raw, object.Name)
		}
	}

	return resultErr
}

// objectMetricCache reuses the field metrics of cluster objects whose values
// did not change since the last scrape. Building the const metrics is most of
// the allocations of a scrape, while the values are still read through ClusAPI
// on every scrape. The key is the object name. The value encodes every input
// of the cached metrics: the published fields with their raw values (see
// appendObjectFieldsKey) and, for groups, the owner node and the node names.
type objectMetricCache = metriccache.Cache[string, string]

// appendObjectFieldsKey appends the published fields of object and their raw
// values to key. It returns false if a field is missing: such an object
// returns an error on every scrape and is not cached.
func appendObjectFieldsKey(key []byte, object clusapi.Object, fields []objectField, build uint16) ([]byte, bool) {
	for _, field := range fields {
		raw, publish, missing := field.resolve(object, build)

		switch {
		case missing:
			return key, false
		case publish:
			key = append(key, 1)
			key = binary.LittleEndian.AppendUint32(key, raw)
		default:
			key = append(key, 0)
		}
	}

	return key, true
}

// appendKeyString appends s with its length to key, so that consecutive
// strings cannot run into each other.
func appendKeyString(key []byte, s string) []byte {
	key = binary.AppendUvarint(key, uint64(len(s)))

	return append(key, s...)
}

// objectFieldMetrics builds the field metrics of an object without missing
// fields, with room for extra metrics.
func objectFieldMetrics(object clusapi.Object, fields []objectField, build uint16, extra int) []prometheus.Metric {
	metrics := make([]prometheus.Metric, 0, len(fields)+extra)

	for _, field := range fields {
		if raw, publish, _ := field.resolve(object, build); publish {
			metrics = append(metrics, field.metric(raw, object.Name))
		}
	}

	return metrics
}

// publishCachedObjectFields publishes the fields of object like
// publishObjectFields and reuses the metrics of the last scrape if they were
// built from the same values. It returns key for reuse as a buffer.
func publishCachedObjectFields(ch chan<- prometheus.Metric, scrape *metriccache.Scrape[string, string], key []byte, kind string, object clusapi.Object, fields []objectField, build uint16, resultErr error) ([]byte, error) {
	key, ok := appendObjectFieldsKey(key[:0], object, fields, build)
	if !ok {
		return key, publishObjectFields(ch, kind, object, fields, build, resultErr)
	}

	metrics := scrape.Get(object.Name, string(key), func() []prometheus.Metric {
		return objectFieldMetrics(object, fields, build, 0)
	})

	for _, metric := range metrics {
		ch <- metric
	}

	return key, resultErr
}
