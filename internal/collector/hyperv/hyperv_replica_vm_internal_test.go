// SPDX-License-Identifier: MIT
//
// Copyright The Prometheus Authors

//go:build windows

package hyperv

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseReplicationRelationshipID(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		instanceID       string
		vmID             string
		relationshipType string
	}{
		{`Microsoft:6B1F2E1A-3C4D-4E5F-8A9B-0C1D2E3F4A5B\HVR\0`, "6B1F2E1A-3C4D-4E5F-8A9B-0C1D2E3F4A5B", "primary"},
		{`Microsoft:6B1F2E1A-3C4D-4E5F-8A9B-0C1D2E3F4A5B\HVR\1`, "6B1F2E1A-3C4D-4E5F-8A9B-0C1D2E3F4A5B", "extended"},
		{`Microsoft:6B1F2E1A-3C4D-4E5F-8A9B-0C1D2E3F4A5B\HVR\2`, "6B1F2E1A-3C4D-4E5F-8A9B-0C1D2E3F4A5B", "unknown"},
		{`unexpected`, "unexpected", "unknown"},
		{``, "", "unknown"},
	} {
		t.Run(tc.instanceID, func(t *testing.T) {
			t.Parallel()

			vmID, relationshipType := parseReplicationRelationshipID(tc.instanceID)
			require.Equal(t, tc.vmID, vmID)
			require.Equal(t, tc.relationshipType, relationshipType)
		})
	}
}
