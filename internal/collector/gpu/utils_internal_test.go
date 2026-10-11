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
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseGPUCounterInstanceString(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name  string
		input string
		want  Instance
	}{
		{
			name:  "engine with pid",
			input: "pid_1234_luid_0x00000000_0x0001136A_phys_0_eng_0_engtype_3D",
			want:  Instance{Pid: "1234", Luid: "0x00000000_0x0001136A", Phys: "0", Eng: "0", Engtype: "3D"},
		},
		{
			name:  "engine type with underscore",
			input: "pid_42_luid_0x00000000_0x00012394_phys_1_eng_12_engtype_Compute_0",
			want:  Instance{Pid: "42", Luid: "0x00000000_0x00012394", Phys: "1", Eng: "12", Engtype: "Compute_0"},
		},
		{
			name:  "process memory",
			input: "pid_10232_luid_0x00000000_0x0001136A_phys_0",
			want:  Instance{Pid: "10232", Luid: "0x00000000_0x0001136A", Phys: "0"},
		},
		{
			name:  "adapter memory",
			input: "luid_0x00000000_0x0001136A_phys_0",
			want:  Instance{Luid: "0x00000000_0x0001136A", Phys: "0"},
		},
		{
			name:  "adapter memory with part",
			input: "luid_0x00000000_0x0001136A_phys_0_part_0",
			want:  Instance{Luid: "0x00000000_0x0001136A", Phys: "0", Part: "0"},
		},
		{
			name:  "truncated luid",
			input: "luid_0x00000000",
			want:  Instance{},
		},
		{
			name:  "empty",
			input: "",
			want:  Instance{},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			require.Equal(t, tc.want, parseGPUCounterInstanceString(tc.input))
		})
	}
}

func TestInstanceCache(t *testing.T) {
	t.Parallel()

	var cache instanceCache

	names := []string{
		"pid_1234_luid_0x00000000_0x00005678_phys_0_eng_0_engtype_3D",
		"pid_1234_luid_0x00000000_0x00005678_phys_0_eng_1_engtype_Compute_0",
		"luid_0x00000000_0x00005678_phys_0",
	}

	for range 3 {
		scrape := cache.begin()

		for _, name := range names {
			require.Equal(t, parseGPUCounterInstanceString(name), scrape.parse(name))
		}

		scrape.commit()
		scrape.commit()
	}

	require.Len(t, cache.entries, len(names))

	// Instances that a scrape doesn't see are dropped.
	scrape := cache.begin()
	scrape.parse(names[0])
	scrape.commit()

	require.Len(t, cache.entries, 1)

	cache.reset()
	require.Empty(t, cache.entries)
}
