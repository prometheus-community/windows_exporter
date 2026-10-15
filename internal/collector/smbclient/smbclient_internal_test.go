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
