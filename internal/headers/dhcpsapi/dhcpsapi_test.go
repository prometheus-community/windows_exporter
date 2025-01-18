package dhcpsapi

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	"golang.org/x/sys/windows"
)

func TestGetDHCPV4ScopeStatistics(t *testing.T) {
	t.Parallel()

	if procDhcpGetSuperScopeInfoV4.Find() != nil {
		t.Skip("DhcpGetSuperScopeInfoV4 is not available")
	}

	_, err := GetDHCPV4ScopeStatistics()
	if errors.Is(err, windows.Errno(1753)) {
		t.Skip(err.Error())
	}

	require.NoError(t, err)
}
