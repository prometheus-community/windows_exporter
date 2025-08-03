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
)

type Instance struct {
	Pid     string
	Luid    string
	Phys    string
	Eng     string
	Engtype string
	Part    string
}

type PidPhys struct {
	Pid  string
	Luid string
	Phys string
}

type PidPhysEngEngType struct {
	Pid     string
	Luid    string
	Phys    string
	Eng     string
	Engtype string
}

func parseGPUCounterInstanceString(s string) Instance {
	// Example: "pid_1234_luid_0x00000000_0x00005678_phys_0_eng_0_engtype_3D"
	// Example: "luid_0x00000000_0x00005678_phys_0"
	// Example: "luid_0x00000000_0x00005678_phys_0_part_0"
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
			if i+1 < len(parts) {
				instance.Engtype = parts[i+1]
			}
		case "part":
			if i+1 < len(parts) {
				instance.Part = parts[i+1]
			}
		}
	}

	return instance
}
