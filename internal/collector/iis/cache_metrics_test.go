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

package iis

import (
	"reflect"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"github.com/stretchr/testify/require"
)

func TestCacheCounterSourcesDistinct(t *testing.T) {
	t.Parallel()

	for _, counterType := range []reflect.Type{reflect.TypeFor[perfDataCounterServiceCache](), reflect.TypeFor[perfDataCounterValuesW3SVCW3WP]()} {
		seen := make(map[string]string)

		for _, field := range reflect.VisibleFields(counterType) {
			tag := field.Tag.Get("perfdata")
			if tag == "" {
				continue
			}

			require.Empty(t, seen[tag], "%s maps both %s and %s", tag, seen[tag], field.Name)
			seen[tag] = field.Name
		}
	}
}

func TestCacheEmitsDifferentSourceValues(t *testing.T) {
	t.Parallel()

	sources := map[string]float64{
		"URI Cache Flushes": 3, "Kernel: URI Cache Flushes": 5,
		"Total Flushed URIs": 7, "Kernel: Total Flushed URIs": 11,
		"Total URIs Cached": 20, "Kernel: Total URIs Cached": 40,
		"Metadata Cache Hits": 80, "Metadata Cache Misses": 20,
		"Output Cache Current Flushed Items": 13,
		"Output Cache Current Items":         17,
		"Output Cache Current Memory Usage":  19,
	}
	c := New(nil)
	c.buildWebServiceCacheDescriptors()
	c.buildW3SVCW3WPDescriptors()
	c.serviceCachePerfDataCollector = cacheFixture[perfDataCounterServiceCache]{sources: sources}
	c.w3SVCW3WPPerfDataCollector = cacheFixture[perfDataCounterValuesW3SVCW3WP]{sources: sources}
	metrics := make(chan prometheus.Metric, 200)
	require.NoError(t, c.collectWebServiceCache(metrics))
	require.NoError(t, c.collectW3SVCW3WPv7(metrics))
	close(metrics)

	values := make(map[*prometheus.Desc][]*dto.Metric)

	for metric := range metrics {
		value := &dto.Metric{}
		require.NoError(t, metric.Write(value))
		values[metric.Desc()] = append(values[metric.Desc()], value)
	}

	for _, tc := range []struct {
		desc *prometheus.Desc
		want []float64
	}{
		{c.serviceCacheURICacheFlushesTotal, []float64{3, 5}},
		{c.serviceCacheURIsFlushedTotal, []float64{7, 11}},
		{c.serviceCacheURIsCachedTotal, []float64{20, 40}},
		{c.serviceCacheMetadataCacheHitsTotal, []float64{80}},
		{c.w3SVCW3WPURICacheFlushesTotal, []float64{3}},
		{c.w3SVCW3WPURIsFlushedTotal, []float64{7}},
	} {
		require.Len(t, values[tc.desc], len(tc.want))

		for i, expected := range tc.want {
			require.InDelta(t, expected, values[tc.desc][i].GetCounter().GetValue(), 1e-9)
		}
	}

	for _, tc := range []struct {
		desc *prometheus.Desc
		want float64
	}{
		{c.serviceCacheOutputCacheActiveFlushedItems, 13},
		{c.serviceCacheOutputCacheItems, 17},
		{c.serviceCacheOutputCacheMemoryUsage, 19},
		{c.w3SVCW3WPOutputCacheActiveFlushedItems, 13},
		{c.w3SVCW3WPOutputCacheItems, 17},
		{c.w3SVCW3WPOutputCacheMemoryUsage, 19},
	} {
		require.Len(t, values[tc.desc], 1)
		require.NotNil(t, values[tc.desc][0].GetGauge())
		require.InDelta(t, tc.want, values[tc.desc][0].GetGauge().GetValue(), 1e-9)
	}
}

type cacheFixture[T any] struct{ sources map[string]float64 }

func (f cacheFixture[T]) Collect(dst *[]T) error {
	var row T

	value := reflect.ValueOf(&row).Elem()
	for _, field := range reflect.VisibleFields(value.Type()) {
		if field.Name == "Name" {
			value.FieldByIndex(field.Index).SetString("123_Example")

			continue
		}

		if tag := field.Tag.Get("perfdata"); tag != "" {
			value.FieldByIndex(field.Index).SetFloat(f.sources[tag])
		}
	}

	*dst = []T{row}

	return nil
}

func (f cacheFixture[T]) Close() {}

func TestCloseBeforeCacheProviderInitialization(t *testing.T) {
	t.Parallel()
	require.NoError(t, New(nil).Close())
}
