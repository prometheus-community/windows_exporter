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

package dhcp

import (
	"errors"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestBuildWithoutDHCPServerAPI checks that a missing DHCP server API only
// disables the scope metrics, unless they are the only sub-collector.
//
//nolint:paralleltest // replaces loadDHCPServerAPI.
func TestBuildWithoutDHCPServerAPI(t *testing.T) {
	original := loadDHCPServerAPI

	t.Cleanup(func() { loadDHCPServerAPI = original })

	loadDHCPServerAPI = func() error { return errors.ErrUnsupported }

	t.Run("server and scope metrics", func(t *testing.T) {
		enabled := []string{subCollectorServerMetrics, subCollectorScopeMetrics}

		c := New(&Config{CollectorsEnabled: enabled})

		// The DHCP Server counters may be missing on this host; only the
		// scope metrics decision is checked.
		_ = c.Build(slog.New(slog.DiscardHandler), nil)

		t.Cleanup(func() { _ = c.Close() })

		require.Equal(t, []string{subCollectorServerMetrics}, c.config.CollectorsEnabled)
		require.Equal(t, []string{subCollectorServerMetrics, subCollectorScopeMetrics}, enabled, "the caller's slice must not change")
		require.Nil(t, c.scopeInfo)
	})

	t.Run("scope metrics only", func(t *testing.T) {
		c := New(&Config{CollectorsEnabled: []string{subCollectorScopeMetrics}})

		err := c.Build(slog.New(slog.DiscardHandler), nil)
		require.ErrorIs(t, err, errors.ErrUnsupported)
	})
}
