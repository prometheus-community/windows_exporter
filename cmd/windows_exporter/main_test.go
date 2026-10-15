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
	"bytes"
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

	"github.com/alecthomas/kingpin/v2"
	"github.com/prometheus-community/windows_exporter/internal/httphandler"
	"github.com/prometheus-community/windows_exporter/pkg/collector"
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

func TestScrapeStdout(t *testing.T) {
	t.Parallel()

	logger := slog.New(slog.DiscardHandler)

	collectors := collector.NewWithFlags(kingpin.New("test", "test"))
	require.NoError(t, collectors.Enable([]string{"cpu"}))
	require.NoError(t, collectors.Build(t.Context(), logger))

	t.Cleanup(func() {
		require.NoError(t, collectors.Close())
	})

	var buf bytes.Buffer

	exitCode := scrapeToStdout(t.Context(), &buf, logger, collectors, &httphandler.Options{})
	require.Equal(t, 0, exitCode)

	output := buf.String()

	// Must contain collector metrics.
	require.Contains(t, output, "windows_cpu_time_seconds_total")
	// Must contain exporter build info (proves the HTTP handler path is used).
	require.Contains(t, output, "windows_exporter_build_info")
	// Must contain collector success metadata.
	require.Contains(t, output, `windows_exporter_collector_success{collector="cpu"} 1`)
}

// serveTest serves handler with the exporter's server settings on a free loopback port.
func serveTest(t *testing.T, ctx context.Context, handler http.Handler) (*http.Server, string) {
	t.Helper()

	listener, err := (&net.ListenConfig{}).Listen(ctx, "tcp", "127.0.0.1:0")
	require.NoError(t, err)

	server := newServer(ctx, handler)

	go func() {
		_ = server.Serve(listener)
	}()

	t.Cleanup(func() {
		_ = server.Close()
	})

	return server, "http://" + listener.Addr().String() + "/"
}

func TestNewServerWriteTimeout(t *testing.T) {
	t.Parallel()

	server := newServer(t.Context(), http.NotFoundHandler())

	// A scrape that uses its full budget must still be able to write its response.
	require.Greater(t, server.WriteTimeout, httphandler.MaxScrapeTimeout)
}

func TestNewServerCancelsInFlightRequests(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	started := make(chan struct{})

	_, uri := serveTest(t, ctx, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		close(started)

		select {
		case <-r.Context().Done():
			_, _ = io.WriteString(w, "canceled")
		case <-time.After(30 * time.Second):
			_, _ = io.WriteString(w, "not canceled")
		}
	}))

	type result struct {
		body string
		err  error
	}

	resultCh := make(chan result, 1)

	go func() {
		req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, uri, nil)
		if err != nil {
			resultCh <- result{err: err}

			return
		}

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			resultCh <- result{err: err}

			return
		}

		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		resultCh <- result{body: string(body), err: err}
	}()

	<-started
	cancel()

	select {
	case res := <-resultCh:
		require.NoError(t, res.err)
		require.Equal(t, "canceled", res.body)
	case <-time.After(10 * time.Second):
		t.Fatal("in-flight request wasn't canceled with the server context")
	}
}

func TestShutdownServerClosesStuckRequests(t *testing.T) {
	t.Parallel()

	started := make(chan struct{})
	release := make(chan struct{})

	t.Cleanup(func() {
		close(release)
	})

	// The handler ignores its context, like a scrape stuck in a native call.
	server, uri := serveTest(t, t.Context(), http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		close(started)
		<-release

		_, _ = io.WriteString(w, "released")
	}))

	errCh := make(chan error, 1)

	go func() {
		req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, uri, nil)
		if err != nil {
			errCh <- err

			return
		}

		resp, err := http.DefaultClient.Do(req)
		if err == nil {
			_ = resp.Body.Close()
		}

		errCh <- err
	}()

	<-started

	start := time.Now()
	err := shutdownServer(t.Context(), server, 100*time.Millisecond)
	require.ErrorIs(t, err, context.DeadlineExceeded)
	require.Less(t, time.Since(start), 5*time.Second)

	// Close drops the connection of the stuck request, so the client doesn't wait for it.
	select {
	case err := <-errCh:
		require.Error(t, err)
	case <-time.After(10 * time.Second):
		t.Fatal("connection of the stuck request wasn't closed")
	}
}

func TestShutdownServerIdle(t *testing.T) {
	t.Parallel()

	server, _ := serveTest(t, t.Context(), http.NotFoundHandler())

	require.NoError(t, shutdownServer(t.Context(), server, time.Second))
}

func TestSetPriorityWindows(t *testing.T) {
	t.Parallel()

	logger := slog.New(slog.DiscardHandler)

	require.NoError(t, setPriorityWindows(t.Context(), logger, os.Getpid(), "normal"))
	require.ErrorContains(t, setPriorityWindows(t.Context(), logger, os.Getpid(), "highest"), `unknown process priority "highest"`)
}
