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

package mi_test

import (
	"testing"

	"github.com/prometheus-community/windows_exporter/internal/mi"
	"github.com/stretchr/testify/require"
)

func Benchmark_MI_Query_Unmarshal(b *testing.B) {
	application, err := mi.Application_Initialize()
	require.NoError(b, err)
	require.NotEmpty(b, application)

	session, err := application.NewSession(nil)
	require.NoError(b, err)
	require.NotEmpty(b, session)

	b.ResetTimer()

	var processes []win32Process

	query, err := mi.NewQuery("SELECT Name FROM Win32_Process WHERE Handle = 0 OR Handle = 4")
	require.NoError(b, err)

	for i := 0; i < b.N; i++ {
		err := session.QueryUnmarshal(&processes, mi.OperationFlagsStandardRTTI, nil, mi.NamespaceRootCIMv2, mi.QueryDialectWQL, query)
		require.NoError(b, err)
		require.Equal(b, []win32Process{{Name: "System Idle Process"}, {Name: "System"}}, processes)
	}

	b.StopTimer()

	err = session.Close()
	require.NoError(b, err)

	err = application.Close()
	require.NoError(b, err)

	b.ReportAllocs()
}
