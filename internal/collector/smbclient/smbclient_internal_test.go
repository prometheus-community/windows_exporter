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

package smbclient

import (
	"math"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseInstanceName(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name   string
		server string
		share  string
		ok     bool
	}{
		{name: `\\fileserver\data`, server: "fileserver", share: "data", ok: true},
		// Windows 11 publishes the instances with a single leading backslash.
		{name: `\localhost\C$`, server: "localhost", share: "C$", ok: true},
		{name: `\\domain\dfs\folder`},
		{name: `_Total`},
		{name: ``},
	} {
		server, share, ok := parseInstanceName(tc.name)

		require.Equal(t, tc.ok, ok, tc.name)
		require.Equal(t, tc.server, server, tc.name)
		require.Equal(t, tc.share, share, tc.name)
	}
}

func TestExtendCounters32(t *testing.T) {
	t.Parallel()

	c := &Collector{counters32: make(map[string]*shareCounters32)}

	sample := func(data perfDataCounterValues) perfDataCounterValues {
		c.generation++

		data.Name = `\\server\share`
		c.extendCounters32(&data)
		c.pruneCounters32()

		return data
	}

	got := sample(perfDataCounterValues{AvgSecPerRead: 1 << 20, ReadRequestsPerSec: 7, ReadBytesPerSec: 100})
	require.Equal(t, float64(1<<20), got.AvgSecPerRead) //nolint:testifylint
	require.Equal(t, 7.0, got.ReadRequestsPerSec)       //nolint:testifylint

	// A timer sum can wrap from anywhere: 2^32 ticks are about 430 s of
	// summed request time.
	got = sample(perfDataCounterValues{AvgSecPerRead: 1 << 10, ReadRequestsPerSec: 9, ReadBytesPerSec: 200})
	require.Equal(t, float64(math.MaxUint32)+1+(1<<10), got.AvgSecPerRead) //nolint:testifylint
	require.Equal(t, 9.0, got.ReadRequestsPerSec)                          //nolint:testifylint

	// A drop of a 64-bit byte counter shows a reconnect, even if DataBytes grew.
	got = sample(perfDataCounterValues{AvgSecPerRead: 5, ReadRequestsPerSec: 1, ReadBytesPerSec: 50, DataBytesPerSec: 1000})
	require.Equal(t, 5.0, got.AvgSecPerRead)      //nolint:testifylint
	require.Equal(t, 1.0, got.ReadRequestsPerSec) //nolint:testifylint

	// A metadata-only share has no byte counter to show the reconnect; a
	// request count dropping from below 2^31 is taken as one.
	got = sample(perfDataCounterValues{MetadataRequestsPerSec: 50, ReadBytesPerSec: 50, DataBytesPerSec: 1000})
	require.Equal(t, 50.0, got.MetadataRequestsPerSec) //nolint:testifylint

	got = sample(perfDataCounterValues{MetadataRequestsPerSec: 3, ReadBytesPerSec: 50, DataBytesPerSec: 1000})
	require.Equal(t, 3.0, got.MetadataRequestsPerSec) //nolint:testifylint
}
