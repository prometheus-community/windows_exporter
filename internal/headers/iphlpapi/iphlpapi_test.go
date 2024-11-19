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
