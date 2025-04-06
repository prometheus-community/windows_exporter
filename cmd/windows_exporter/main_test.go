// Copyright 2024 The Prometheus Authors
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
			ctx, cancel := context.WithCancel(context.Background())
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

			var stdout string

			go func() {
				stdout = captureOutput(t, func() {
					// Simulate the service control manager signaling that we are done.
					exitCodeCh <- run(ctx, tc.args)
				})
			}()

			t.Cleanup(func() {
				select {
				case exitCode := <-exitCodeCh:
					require.Equal(t, tc.exitCode, exitCode)
				case <-time.After(2 * time.Second):
					t.Fatalf("timed out waiting for exit code, want %d", tc.exitCode)
				}
			})

			if tc.exitCode != 0 {
				return
			}

			uri, err := url.Parse(tc.metricsEndpoint)
			require.NoError(t, err)

			err = waitUntilListening(t, "tcp", uri.Host)
			require.NoError(t, err, "LOGS:\n%s", stdout)

			req, err := http.NewRequestWithContext(ctx, http.MethodGet, tc.metricsEndpoint, nil)
			require.NoError(t, err)

			resp, err := http.DefaultClient.Do(req)
			require.NoError(t, err, "LOGS:\n%s", stdout)
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

func captureOutput(tb testing.TB, f func()) string {
	tb.Helper()

	orig := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	f()

	os.Stdout = orig

	_ = w.Close()

	out, _ := io.ReadAll(r)

	return string(out)
}

func waitUntilListening(tb testing.TB, network, address string) error {
	tb.Helper()

	var (
		conn net.Conn
		err  error
	)

	for range 10 {
		conn, err = net.DialTimeout(network, address, 100*time.Millisecond)
		if err == nil {
			_ = conn.Close()

			return nil
		}

		if errors.Is(err, windows.Errno(10061)) {
			time.Sleep(50 * time.Millisecond)

			continue
		}
	}

	return fmt.Errorf("listener not listening: %w", err)
}
