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
