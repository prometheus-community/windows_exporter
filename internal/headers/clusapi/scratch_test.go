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

package clusapi

import (
	"bytes"
	"maps"
	"testing"
	"time"

	"golang.org/x/sys/windows"
)

// controlFromList answers a control with list, like ClusterResourceControl
// filling the caller's buffer.
func controlFromList(list []byte) func([]byte) (uint32, error) {
	return func(buffer []byte) (uint32, error) {
		if len(buffer) < len(list) {
			return uint32(len(list)), windows.ERROR_MORE_DATA
		}

		return uint32(copy(buffer, list)), nil
	}
}

// readListsIntoResource replays the property list part of readResource: both
// lists go through the same scratch buffer, one after the other.
func readListsIntoResource(tb testing.TB, scratch *controlScratch, resource *Resource, lists ...[]byte) bool {
	tb.Helper()

	hasType := false

	for _, list := range lists {
		data, err := controlBufferIn(time.Time{}, scratch.get(propertyListBufferSize), controlFromList(list))
		if err != nil {
			tb.Fatal(err)
		}

		scratch.keep(data)

		properties, err := ParseProperties(data)
		if err != nil {
			tb.Fatal(err)
		}

		if copyResourceProperties(resource, properties) {
			hasType = true
		}
	}

	return hasType
}

func TestCopyResourceProperties(t *testing.T) {
	var scratch controlScratch

	resource := Resource{Values: make(map[string]uint32)}
	hasType := readListsIntoResource(t, &scratch, &resource,
		readPropertyFixture(t, "resource_ro_common_generic_service.hex"),
		readPropertyFixture(t, "resource_common_generic_service.hex"),
	)

	// The second list overwrote the buffer of the first; the values must have
	// been copied out before.
	want := map[string]uint32{
		"MonitorProcessId":       9092,
		"DeadlockTimeout":        300000,
		"EmbeddedFailureAction":  2,
		"IsAlivePollInterval":    ^uint32(0),
		"LooksAlivePollInterval": ^uint32(0),
		"PendingTimeout":         180000,
		"PersistentState":        1,
		"RestartAction":          2,
		"RestartDelay":           500,
		"RestartPeriod":          600000,
		"RestartThreshold":       1,
		"RetryPeriodOnFailure":   600000,
		"SeparateMonitor":        0,
	}
	if !maps.Equal(resource.Values, want) {
		t.Errorf("Values = %v, want %v", resource.Values, want)
	}

	if !hasType || resource.Type != "Generic Service" {
		t.Errorf("Type = %q, hasType = %t", resource.Type, hasType)
	}

	clear(scratch.buffer)

	if resource.Type != "Generic Service" {
		t.Errorf("Type changed with the buffer: %q", resource.Type)
	}

	// The read-only list has no Type property.
	resource = Resource{Values: make(map[string]uint32)}
	if readListsIntoResource(t, &scratch, &resource, readPropertyFixture(t, "resource_ro_common_generic_service.hex")) {
		t.Errorf("hasType for a list without Type, Type = %q", resource.Type)
	}
}

func TestControlScratch(t *testing.T) {
	var scratch controlScratch

	buffer := scratch.get(8)
	if len(buffer) != 8 || cap(buffer) != propertyListBufferSize {
		t.Fatalf("get(8): len %d, cap %d", len(buffer), cap(buffer))
	}

	// A reused buffer must look like a new one: the cluster service may
	// report more bytes than it wrote, and parsing relies on zero padding.
	copy(buffer, "leftover")

	if buffer := scratch.get(8); !bytes.Equal(buffer, make([]byte, 8)) {
		t.Errorf("get after use = % x, want zeros", buffer)
	}

	// A buffer grown by ERROR_MORE_DATA is kept for the next object, up to
	// maxScratchSize.
	grown := make([]byte, 16<<10)
	scratch.keep(grown[:100])

	if cap(scratch.buffer) != len(grown) {
		t.Errorf("keep: cap %d, want %d", cap(scratch.buffer), len(grown))
	}

	scratch.keep(make([]byte, maxScratchSize+1))

	if cap(scratch.buffer) != len(grown) {
		t.Errorf("kept a buffer above maxScratchSize: cap %d", cap(scratch.buffer))
	}
}

func TestControlBufferInGrowsWithinCapacity(t *testing.T) {
	var scratch controlScratch

	list := bytes.Repeat([]byte{1}, 1000)
	start := scratch.get(typeBufferSize)
	copy(start[:cap(start)], bytes.Repeat([]byte{0xff}, cap(start)))

	data, err := controlBufferIn(time.Time{}, start, controlFromList(list))
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(data, list) {
		t.Fatalf("data = % x", data[:16])
	}

	// ERROR_MORE_DATA reused the scratch buffer instead of allocating.
	if &data[0] != &scratch.buffer[0] {
		t.Error("grew into a new buffer although the capacity was enough")
	}
}

// BenchmarkReadResourceProperties replays the property list work readResource
// does for one resource with the captured CI lists.
func BenchmarkReadResourceProperties(b *testing.B) {
	lists := [][]byte{
		readPropertyFixture(b, "resource_ro_common_generic_service.hex"),
		readPropertyFixture(b, "resource_common_generic_service.hex"),
	}

	var scratch controlScratch

	b.ReportAllocs()

	for b.Loop() {
		resource := Resource{Values: make(map[string]uint32)}
		readListsIntoResource(b, &scratch, &resource, lists...)
	}
}
