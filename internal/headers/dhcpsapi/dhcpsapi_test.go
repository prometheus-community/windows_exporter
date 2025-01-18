package dhcpsapi

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetDHCPV4ScopeStatistics(t *testing.T) {
	t.Parallel()

	if procDhcpGetSuperScopeInfoV4.Find() != nil {
		t.Skip("DhcpGetSuperScopeInfoV4 is not available")
	}

	_, err := GetDHCPV4ScopeStatistics()
	require.NoError(t, err)
}
