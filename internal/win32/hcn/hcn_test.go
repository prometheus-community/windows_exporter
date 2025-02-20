package hcn_test

import (
	"testing"

	"github.com/prometheus-community/windows_exporter/internal/win32/hcn"
	"github.com/stretchr/testify/require"
)

func TestEnumerateEndpoints(t *testing.T) {
	t.Parallel()

	endpoints, err := hcn.EnumerateEndpoints()
	require.NoError(t, err)
	require.NotNil(t, endpoints)
}

func TestQueryEndpointProperties(t *testing.T) {
	t.Parallel()

	endpoints, err := hcn.EnumerateEndpoints()
	require.NoError(t, err)

	if len(endpoints) == 0 {
		t.Skip("No endpoints found")
	}

	_, err = hcn.GetEndpointProperties(endpoints[0])
	require.NoError(t, err)
}
