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

package gpu

import (
	"fmt"
	"strings"
	"sync"

	"github.com/prometheus-community/windows_exporter/internal/headers/gdi32"
)

// isBasicRenderDriver reports whether the adapter is the Microsoft Basic Render Driver.
// https://devicehunt.com/view/type/pci/vendor/1414/device/008C
// The D3DKMT_ADAPTERTYPE SoftwareDevice flag is not set for it on every system,
// e.g. Windows Server 2022, so the device ID is checked as well.
func isBasicRenderDriver(device gdi32.GPUDevice) bool {
	return strings.HasPrefix(device.DeviceID, `PCI\VEN_1414&DEV_008C&`)
}

type Instance struct {
	Pid     string
	Luid    string
	Phys    string
	Eng     string
	Engtype string
	Part    string
}

func parseGPUCounterInstanceString(s string) Instance {
	// Example: "pid_1234_luid_0x00000000_0x00005678_phys_0_eng_0_engtype_3D"
	// Example: "luid_0x00000000_0x00005678_phys_0"
	// Example: "luid_0x00000000_0x00005678_phys_0_part_0"
	// Example: "pid_1234_luid_0x00000000_0x00005678_phys_0_eng_1_engtype_Compute_0"
	parts := strings.Split(s, "_")

	var instance Instance

	for i, part := range parts {
		switch part {
		case "pid":
			if i+1 < len(parts) {
				instance.Pid = parts[i+1]
			}
		case "luid":
			if i+2 < len(parts) {
				instance.Luid = fmt.Sprintf("%s_%s", parts[i+1], parts[i+2])
			}
		case "phys":
			if i+1 < len(parts) {
				instance.Phys = parts[i+1]
			}
		case "eng":
			if i+1 < len(parts) {
				instance.Eng = parts[i+1]
			}
		case "engtype":
			// The engine type is always the last element and may contain underscores, e.g. "Compute_0".
			if i+1 < len(parts) {
				instance.Engtype = strings.Join(parts[i+1:], "_")
			}

			return instance
		case "part":
			if i+1 < len(parts) {
				instance.Part = parts[i+1]
			}
		}
	}

	return instance
}

// instanceCache keeps the parsed counter instance names of the last scrape.
// The same instances appear in every scrape while a GPU process or engine
// exists, so they are only parsed once.
type instanceCache struct {
	mu      sync.Mutex
	entries map[string]Instance
}

// instanceScrape parses the instance names of one scrape.
type instanceScrape struct {
	cache *instanceCache
	// prev is the last committed generation. It is never modified, so it is
	// read without holding mu.
	prev map[string]Instance

	mu   sync.Mutex
	next map[string]Instance
}

func (c *instanceCache) begin() *instanceScrape {
	c.mu.Lock()
	prev := c.entries
	c.mu.Unlock()

	return &instanceScrape{cache: c, prev: prev, next: make(map[string]Instance, len(prev))}
}

func (c *instanceCache) reset() {
	c.mu.Lock()
	c.entries = nil
	c.mu.Unlock()
}

// parse returns parseGPUCounterInstanceString(name), reusing the result of
// the previous scrape.
func (s *instanceScrape) parse(name string) Instance {
	instance, ok := s.prev[name]

	s.mu.Lock()
	if !ok {
		instance, ok = s.next[name]
	}
	s.mu.Unlock()

	if !ok {
		instance = parseGPUCounterInstanceString(name)
	}

	s.mu.Lock()
	if s.next != nil {
		s.next[name] = instance
	}
	s.mu.Unlock()

	return instance
}

// commit makes the instances of this scrape the base of the next one.
func (s *instanceScrape) commit() {
	s.mu.Lock()
	next := s.next
	s.next = nil
	s.mu.Unlock()

	if next == nil {
		return
	}

	s.cache.mu.Lock()
	s.cache.entries = next
	s.cache.mu.Unlock()
}
