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
	"testing"
	"time"
)

// BenchmarkReadResourceProperties replays the property list work readResource
// does for one resource: two property list controls into the shared buffer,
// then the DWORD values and the type name copied out of the parsed lists.
func BenchmarkReadResourceProperties(b *testing.B) {
	lists := [][]byte{
		readPropertyFixture(b, "resource_ro_common_generic_service.hex"),
		readPropertyFixture(b, "resource_common_generic_service.hex"),
	}

	b.ReportAllocs()

	for b.Loop() {
		resource := Resource{Values: make(map[string]uint32)}
		buffer := propertyLists.Get().(*[]byte) //nolint:forcetypeassert

		for _, list := range lists {
			data, err := controlBufferIn(time.Time{}, *buffer, func(buffer []byte) (uint32, error) {
				return uint32(copy(buffer, list)), nil
			})
			if err != nil {
				b.Fatal(err)
			}

			properties, err := ParseProperties(data)
			if err != nil {
				b.Fatal(err)
			}

			copyResourceProperties(&resource, properties)
		}

		propertyLists.Put(buffer)
	}
}
