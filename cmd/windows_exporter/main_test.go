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

package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"golang.org/x/sys/windows"
)

//nolint:tparallel
func TestRun(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name            string
		args            []string
		config          string
		metricsEndpoint string
		exitCode        int
	}{
		{
			name:            "default",
			args:            []string{},
			metricsEndpoint: "http://127.0.0.1:9182/metrics",
		},
		{
			name:            "web.listen-address",
			args:            []string{"--web.listen-address=127.0.0.1:8080"},
			metricsEndpoint: "http://127.0.0.1:8080/metrics",
		},
		{
			name:            "web.listen-address",
			args:            []string{"--web.listen-address=127.0.0.1:8081", "--web.listen-address=[::1]:8081"},
			metricsEndpoint: "http://[::1]:8081/metrics",
		},
		{
			name:            "config",
			args:            []string{"--config.file=config.yaml"},
			config:          `{"web":{"listen-address":"127.0.0.1:8082"}}`,
			metricsEndpoint: "http://127.0.0.1:8082/metrics",
		},
		{
			name:            "web.listen-address with config",
			args:            []string{"--config.file=config.yaml", "--web.listen-address=127.0.0.1:8084"},
			config:          `{"web":{"listen-address":"127.0.0.1:8083"}}`,
			metricsEndpoint: "http://127.0.0.1:8084/metrics",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(t.Context())
			defer cancel()

			if tc.config != "" {
				// Create a temporary config file.
				tmpfile, err := os.CreateTemp(t.TempDir(), "config-*.yaml")
				require.NoError(t, err)

				t.Cleanup(func() {
					require.NoError(t, tmpfile.Close())
				})

				_, err = tmpfile.WriteString(tc.config)
				require.NoError(t, err)

				for i, arg := range tc.args {
					tc.args[i] = strings.ReplaceAll(arg, "config.yaml", tmpfile.Name())
				}
			}

			exitCodeCh := make(chan int)

			go func() {
				exitCodeCh <- run(ctx, tc.args)
			}()

			t.Cleanup(func() {
				// run closes the collectors before it returns. That takes milliseconds,
				// but native teardown (PDH, MI) can stall on a loaded CI runner, so
				// wait for the bound closeCollection enforces.
				select {
				case exitCode := <-exitCodeCh:
					require.Equal(t, tc.exitCode, exitCode)
				case <-time.After(collectionCloseTimeout + 5*time.Second):
					t.Fatalf("timed out waiting for exit code, want %d", tc.exitCode)
				}
			})

			if tc.exitCode != 0 {
				return
			}

			uri, err := url.Parse(tc.metricsEndpoint)
			require.NoError(t, err)

			err = waitUntilListening(t, "tcp", uri.Host)
			require.NoError(t, err)

			req, err := http.NewRequestWithContext(ctx, http.MethodGet, tc.metricsEndpoint, nil)
			require.NoError(t, err)

			resp, err := http.DefaultClient.Do(req)
			require.NoError(t, err)
			require.Equal(t, http.StatusOK, resp.StatusCode)

			body, err := io.ReadAll(resp.Body)
			require.NoError(t, err)

			err = resp.Body.Close()
			require.NoError(t, err)

			require.NotEmpty(t, body)
			require.Contains(t, string(body), "# HELP windows_exporter_build_info")

			cancel()
		})
	}
}

func waitUntilListening(tb testing.TB, network, address string) error {
	tb.Helper()

	var (
		conn net.Conn
		err  error
	)

	dialer := &net.Dialer{Timeout: 100 * time.Millisecond}

	deadline := time.Now().Add(30 * time.Second)

	// Before the listener starts, dials fail with connection refused or, on a
	// busy Windows host, with a dial timeout. Retry both until the deadline.
	for time.Now().Before(deadline) {
		conn, err = dialer.DialContext(tb.Context(), network, address)
		if err == nil {
			_ = conn.Close()

			return nil
		}

		if tb.Context().Err() != nil {
			break
		}

		time.Sleep(50 * time.Millisecond)
	}

	if winErr, ok := errors.AsType[windows.Errno](err); ok {
		return fmt.Errorf("listener not listening: %w (#%d)", winErr, uint32(winErr))
	}

	return fmt.Errorf("listener not listening: %w", err)
}

func TestSetPriorityWindows(t *testing.T) {
	t.Parallel()

	logger := slog.New(slog.DiscardHandler)

	require.NoError(t, setPriorityWindows(t.Context(), logger, os.Getpid(), "normal"))
	require.ErrorContains(t, setPriorityWindows(t.Context(), logger, os.Getpid(), "highest"), `unknown process priority "highest"`)
}
