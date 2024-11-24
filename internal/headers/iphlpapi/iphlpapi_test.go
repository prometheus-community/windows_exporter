// Copyright 2024 The Prometheus Authors
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

package iphlpapi_test

import (
	"net"
	"os"
	"testing"

	"github.com/prometheus-community/windows_exporter/internal/headers/iphlpapi"
	"github.com/stretchr/testify/require"
	"golang.org/x/sys/windows"
)

func TestGetTCPConnectionStates(t *testing.T) {
	t.Parallel()

	pid, err := iphlpapi.GetTCPConnectionStates(windows.AF_INET)
	require.NoError(t, err)
	require.NotEmpty(t, pid)
}

func TestGetOwnerPIDOfTCPPort(t *testing.T) {
	t.Parallel()

	lister, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	t.Cleanup(func() {
		require.NoError(t, lister.Close())
	})

	pid, err := iphlpapi.GetOwnerPIDOfTCPPort(windows.AF_INET, uint16(lister.Addr().(*net.TCPAddr).Port))
	require.NoError(t, err)
	require.EqualValues(t, os.Getpid(), pid)
}
