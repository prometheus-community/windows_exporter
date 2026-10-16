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
