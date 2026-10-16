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

package wtsapi32_test

import (
	"log/slog"
	"testing"

	"github.com/prometheus-community/windows_exporter/internal/headers/wtsapi32"
	"github.com/stretchr/testify/require"
)

func TestWTSEnumerateSessionsEx(t *testing.T) {
	t.Parallel()

	server, err := wtsapi32.WTSOpenServer("")
	require.NoError(t, err)

	t.Cleanup(func() { wtsapi32.WTSCloseServer(server) })

	sessions, err := wtsapi32.WTSEnumerateSessionsEx(server, slog.New(slog.DiscardHandler))
	require.NoError(t, err)
	require.NotEmpty(t, sessions, "session 0 always exists")
}

func BenchmarkWTSEnumerateSessionsEx(b *testing.B) {
	server, err := wtsapi32.WTSOpenServer("")
	require.NoError(b, err)

	b.Cleanup(func() { wtsapi32.WTSCloseServer(server) })

	logger := slog.New(slog.DiscardHandler)

	b.ReportAllocs()

	for b.Loop() {
		_, _ = wtsapi32.WTSEnumerateSessionsEx(server, logger)
	}
}
