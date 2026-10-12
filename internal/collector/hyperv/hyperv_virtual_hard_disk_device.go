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

package hyperv

import (
	"path/filepath"
	"slices"
	"strings"

	"github.com/prometheus-community/windows_exporter/internal/headers/virtdisk"
)

// virtualHardDiskMaxChainDepth bounds how many parents of a differencing disk
// are followed to find its "Hyper-V Virtual Storage Device" counter instance.
const virtualHardDiskMaxChainDepth = 16

//nolint:gochecknoglobals
var virtualStorageDeviceNameReplacer = strings.NewReplacer(`\`, "-", "(", "[", ")", "]")

// virtualStorageDeviceName returns the name that the "Hyper-V Virtual Storage
// Device" counter set uses for the disk at path: counter paths reserve "\"
// and "(...)", so Hyper-V replaces them with "-" and "[...]". A "\\?\" prefix
// becomes "--?-".
func virtualStorageDeviceName(path string) string {
	return virtualStorageDeviceNameReplacer.Replace(path)
}

// virtualStorageDeviceMatcher finds the counter instance of a disk among the
// instances that exist. Only running virtual machines have instances.
type virtualStorageDeviceMatcher struct {
	instances map[string]string // instance name and lower case instance name -> instance name
}

func newVirtualStorageDeviceMatcher(instances []string) virtualStorageDeviceMatcher {
	matcher := virtualStorageDeviceMatcher{instances: make(map[string]string, 2*len(instances))}

	for _, instance := range instances {
		matcher.instances[instance] = instance
	}

	// Exact names win over names that only match without case.
	for _, instance := range instances {
		if _, ok := matcher.instances[strings.ToLower(instance)]; !ok {
			matcher.instances[strings.ToLower(instance)] = instance
		}
	}

	return matcher
}

// match returns the instance of the first candidate path that has one, or ""
// if none has. The candidates are a disk and then its parents: the instance of
// a VHD (not VHDX) checkpoint .avhd is named after the base disk of its chain.
func (m virtualStorageDeviceMatcher) match(candidates []string) string {
	for _, candidate := range candidates {
		name := virtualStorageDeviceName(candidate)

		if instance, ok := m.instances[name]; ok {
			return instance
		}

		if instance, ok := m.instances[strings.ToLower(name)]; ok {
			return instance
		}
	}

	return ""
}

// virtualHardDiskChain returns path followed by its parents, the base disk
// last. Only differencing disks have parents. Relative parent locations are
// resolved against the directory of the child.
func virtualHardDiskChain(
	path string,
	subtype uint32,
	getInfo func(path string) (virtdisk.Info, error),
	parentLocations func(path string) ([]string, error),
) []string {
	chain := []string{path}
	seen := map[string]bool{strings.ToLower(path): true}

	for current := path; subtype == virtdisk.ProviderSubtypeDifferencing && len(chain) < virtualHardDiskMaxChainDepth; {
		locations, err := parentLocations(current)
		if err != nil {
			break
		}

		next := ""

		// Absolute locations first: they match the counter instance names,
		// relative ones may contain "..".
		locations = slices.Clone(locations)
		slices.SortStableFunc(locations, func(a, b string) int {
			switch {
			case filepath.IsAbs(a) == filepath.IsAbs(b):
				return 0
			case filepath.IsAbs(a):
				return -1
			default:
				return 1
			}
		})

		for _, location := range locations {
			if !filepath.IsAbs(location) {
				location = filepath.Join(filepath.Dir(current), location)
			}

			if seen[strings.ToLower(location)] {
				continue
			}

			info, err := getInfo(location)
			if err != nil {
				continue
			}

			next, subtype = location, info.ProviderSubtype

			break
		}

		if next == "" {
			break
		}

		chain = append(chain, next)
		seen[strings.ToLower(next)] = true
		current = next
	}

	return chain
}
