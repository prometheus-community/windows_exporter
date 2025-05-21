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

package hcn_test

import (
	"testing"

	"github.com/prometheus-community/windows_exporter/internal/headers/hcn"
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
