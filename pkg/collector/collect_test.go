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

package collector_test

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"testing/synctest"
	"time"

	"github.com/prometheus-community/windows_exporter/internal/mi"
	"github.com/prometheus-community/windows_exporter/internal/pdh"
	"github.com/prometheus-community/windows_exporter/internal/types"
	"github.com/prometheus-community/windows_exporter/pkg/collector"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/expfmt"
	"github.com/stretchr/testify/require"
	"golang.org/x/sys/windows"
)

// The tests in this file share the package-level scrape slot, so they don't run in parallel.

func TestCollectErrorStatus(t *testing.T) {
	failure := errors.New("collector failed")

	for _, tc := range []struct {
		name    string
		err     error
		success string
	}{
		{name: "success", success: "1"},
		{name: "failure", err: failure, success: "0"},
		{name: "no data", err: types.ErrNoData, success: "1"},
		{name: "PDH no data", err: pdh.ErrNoData, success: "1"},
		{name: "RPC endpoint absent", err: windows.EPT_S_NOT_REGISTERED, success: "1"},
		{name: "unexpected no data", err: types.ErrNoDataUnexpected, success: "0"},
		{
			name:    "wrapped no data",
			err:     fmt.Errorf("empty instances: %w", pdh.ErrNoData),
			success: "1",
		},
		{
			name:    "joined expected errors",
			err:     errors.Join(pdh.ErrNoData, types.ErrNoData, windows.EPT_S_NOT_REGISTERED),
			success: "1",
		},
		{
			name: "nested expected errors",
			err: fmt.Errorf(
				"groups: %w",
				errors.Join(pdh.ErrNoData, errors.Join(types.ErrNoData, windows.EPT_S_NOT_REGISTERED)),
			),
			success: "1",
		},
		{
			name:    "classified child warning",
			err:     fmt.Errorf("child failed: %s: %w", failure.Error(), types.ErrNoData),
			success: "1",
		},
		{
			name:    "failure joined with no data",
			err:     errors.Join(types.ErrNoData, failure),
			success: "0",
		},
		{
			name:    "failure joined with PDH no data",
			err:     errors.Join(pdh.ErrNoData, failure),
			success: "0",
		},
		{
			name:    "failure joined with absent RPC endpoint",
			err:     errors.Join(windows.EPT_S_NOT_REGISTERED, failure),
			success: "0",
		},
		{
			name:    "nested mixed errors",
			err:     fmt.Errorf("groups: %w", errors.Join(pdh.ErrNoData, errors.Join(types.ErrNoData, failure))),
			success: "0",
		},
		{
			name:    "multiple wrapped errors",
			err:     fmt.Errorf("first: %w; second: %w", pdh.ErrNoData, failure),
			success: "0",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			logs := &recordHandler{}
			collection := collector.New(collector.Map{"test": &fakeCollector{name: "test", collectErr: tc.err}})
			handler, err := collection.NewHandler(time.Second, slog.New(logs), nil)
			require.NoError(t, err)

			out := scrape(t, handler)
			require.Contains(t, out, `windows_exporter_collector_success{collector="test"} `+tc.success)
			require.Contains(t, out, `windows_exporter_collector_timeout{collector="test"} 0`)
			require.Contains(t, out, "windows_test_test 42", "partial metrics must still be exported")

			warnings := 0
			if tc.success == "0" {
				warnings = 1
			}

			require.Equal(t, warnings, logs.warnings("failed"))
		})
	}
}

func TestCollectTimedOutCollectorDoesNotOverlap(t *testing.T) {
	// The collector's delay and the scrape deadlines use synthetic time.
	synctest.Test(t, func(t *testing.T) {
		logs := &recordHandler{}
		logger := slog.New(logs)

		slow := &fakeCollector{name: "slow", delay: 2 * time.Second}
		fast := &fakeCollector{name: "fast"}
		collection := collector.New(collector.Map{"slow": slow, "fast": fast})

		all, err := collection.NewHandler(300*time.Millisecond, logger, nil)
		require.NoError(t, err)

		// A filtered view must share the running state with the full collection.
		filtered, err := collection.NewHandler(300*time.Millisecond, logger, []string{"slow"})
		require.NoError(t, err)

		start := time.Now()

		for i := range 5 {
			handler := all
			if i%2 == 1 {
				handler = filtered
			}

			out := scrape(t, handler)
			require.Contains(t, out, `windows_exporter_collector_timeout{collector="slow"} 1`)
			require.Contains(t, out, `windows_exporter_collector_success{collector="slow"} 0`)
			require.NotContains(t, out, "windows_test_slow")

			if handler == all {
				require.Contains(t, out, `windows_exporter_collector_success{collector="fast"} 1`)
				require.Contains(t, out, "windows_test_fast 42")
			}
		}

		// Only the first scrape waited for its deadline, because the others skipped the running collector.
		require.Equal(t, 300*time.Millisecond, time.Since(start))

		// The first call is still running, and no second call was started.
		require.Equal(t, int32(1), slow.calls.Load())
		require.Equal(t, int32(1), slow.maxActive.Load())
		require.Equal(t, 1, logs.warnings("timeouted"))
		require.Equal(t, 1, logs.warnings("is still running"))

		// After the first call returns, the next scrape starts a new one.
		time.Sleep(2 * time.Second)
		synctest.Wait()

		scrape(t, all)
		require.Equal(t, int32(2), slow.calls.Load())
		require.Equal(t, int32(1), slow.maxActive.Load())

		// Let the second call finish before leaving the bubble.
		time.Sleep(2 * time.Second)
		synctest.Wait()
	})
}

func TestCollectAdmission(t *testing.T) {
	t.Run("canceled while queued", func(t *testing.T) {
		running := newBlockingCollector()
		collection := collector.New(collector.Map{"test": running})
		first := scrapeInBackground(t.Context(), collection, time.Minute)

		<-running.started

		ctx, cancel := context.WithCancel(t.Context())
		second := scrapeInBackground(ctx, collection, time.Minute)

		// Give the second scrape time to queue behind the first one. The result is the same if it hasn't yet.
		time.Sleep(50 * time.Millisecond)
		cancel()

		select {
		case out := <-second:
			require.Contains(t, out, `windows_exporter_collector_timeout{collector="test"} 1`)
		case <-time.After(10 * time.Second):
			t.Fatal("canceled scrape didn't return")
		}

		close(running.block)
		require.Contains(t, <-first, `windows_exporter_collector_success{collector="test"} 1`)
		require.Equal(t, int32(1), running.calls.Load())
	})

	t.Run("deadline passes while queued", func(t *testing.T) {
		running := newBlockingCollector()
		collection := collector.New(collector.Map{"test": running})
		first := scrapeInBackground(t.Context(), collection, time.Minute)

		<-running.started

		start := time.Now()
		second := scrapeInBackground(t.Context(), collection, 100*time.Millisecond)

		select {
		case out := <-second:
			require.Contains(t, out, `windows_exporter_collector_timeout{collector="test"} 1`)
			require.Less(t, time.Since(start), 10*time.Second)
		case <-time.After(10 * time.Second):
			t.Fatal("queued scrape didn't return at its deadline")
		}

		close(running.block)
		<-first
		require.Equal(t, int32(1), running.calls.Load())
	})

	t.Run("canceled before admission", func(t *testing.T) {
		test := &fakeCollector{name: "test"}
		collection := collector.New(collector.Map{"test": test})

		ctx, cancel := context.WithCancel(t.Context())
		cancel()

		handler, err := collection.NewHandlerWithContext(ctx, time.Minute, slog.New(slog.DiscardHandler), nil)
		require.NoError(t, err)

		require.Contains(t, scrape(t, handler), `windows_exporter_collector_timeout{collector="test"} 1`)
		require.Equal(t, int32(0), test.calls.Load())
	})

	t.Run("deadline starts when the handler is created", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			test := &fakeCollector{name: "test"}
			collection := collector.New(collector.Map{"test": test})

			handler, err := collection.NewHandlerWithContext(t.Context(), 10*time.Second, slog.New(slog.DiscardHandler), nil)
			require.NoError(t, err)

			time.Sleep(4 * time.Second)

			require.Contains(t, scrape(t, handler), `windows_exporter_collector_success{collector="test"} 1`)
			require.Equal(t, 6*time.Second, time.Duration(test.budget.Load()))
		})
	})
}

func TestCollectFailedBuild(t *testing.T) {
	for _, tc := range []struct {
		name     string
		buildErr error
		panics   bool
		wantErr  string
	}{
		{name: "tolerated error", buildErr: pdh.ErrNoData},
		{name: "error", buildErr: errors.New("build failed"), wantErr: "build failed"},
		{name: "panic", panics: true, wantErr: "panic: build panic"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			logs := &recordHandler{}
			logger := slog.New(logs)

			failing := &fakeCollector{name: "failing", buildErr: tc.buildErr, buildPanic: tc.panics}
			working := &fakeCollector{name: "working"}
			collection := collector.New(collector.Map{"failing": failing, "working": working})

			err := collection.Build(t.Context(), logger)
			if tc.wantErr == "" {
				require.NoError(t, err)
				require.Equal(t, 1, logs.warnings("couldn't initialize collector"))
			} else {
				require.ErrorContains(t, err, tc.wantErr)
			}

			if tc.panics {
				// The error carries the stack of the panic.
				require.ErrorContains(t, err, "runtime/debug.Stack")
			}

			handler, err := collection.NewHandler(time.Minute, logger, nil)
			require.NoError(t, err)

			logs.reset()

			for range 3 {
				out := scrape(t, handler)
				require.Contains(t, out, `windows_exporter_collector_success{collector="failing"} 0`)
				require.Contains(t, out, `windows_exporter_collector_timeout{collector="failing"} 0`)
				require.Contains(t, out, `windows_exporter_collector_success{collector="working"} 1`)
				require.NotContains(t, out, "windows_test_failing")
			}

			require.Equal(t, int32(0), failing.calls.Load())
			require.Equal(t, int32(3), working.calls.Load())
			require.Zero(t, logs.warnings(""), "a collector that failed to build must not log on every scrape")

			// The collector may hold resources from its partial Build, so it is closed.
			require.NoError(t, collection.Close())
			require.Equal(t, int32(1), failing.closes.Load())
		})
	}
}

func TestBuildCanceled(t *testing.T) {
	building := &fakeCollector{name: "building", buildBlock: make(chan struct{}), buildStarted: make(chan struct{}, 1)}
	collection := collector.New(collector.Map{"building": building})

	ctx, cancel := context.WithCancel(t.Context())

	errCh := make(chan error, 1)

	go func() {
		errCh <- collection.Build(ctx, slog.New(slog.DiscardHandler))
	}()

	<-building.buildStarted
	cancel()

	select {
	case err := <-errCh:
		require.ErrorIs(t, err, context.Canceled)
	case <-time.After(10 * time.Second):
		t.Fatal("Build didn't return after its context was canceled")
	}

	// Close waits for the abandoned Build to return before it closes the collector.
	closeErr := make(chan error, 1)

	go func() {
		closeErr <- collection.Close()
	}()

	close(building.buildBlock)
	require.NoError(t, <-closeErr)
	require.Equal(t, int32(1), building.closes.Load())

	require.ErrorIs(t, collection.Build(canceledContext(), slog.New(slog.DiscardHandler)), context.Canceled)
}

func TestCollectionClose(t *testing.T) {
	t.Run("without Build", func(t *testing.T) {
		test := &fakeCollector{name: "test"}
		collection := collector.New(collector.Map{"test": test})

		require.NoError(t, collection.Close())
		require.NoError(t, collection.Close())
		require.Equal(t, int32(0), test.closes.Load())
	})

	t.Run("idempotent", func(t *testing.T) {
		test := &fakeCollector{name: "test"}
		collection := collector.New(collector.Map{"test": test})
		require.NoError(t, collection.Build(t.Context(), slog.New(slog.DiscardHandler)))

		require.NoError(t, collection.Close())
		require.NoError(t, collection.Close())
		require.Equal(t, int32(1), test.closes.Load())

		// A closed collector is not collected anymore.
		handler, err := collection.NewHandler(time.Minute, slog.New(slog.DiscardHandler), nil)
		require.NoError(t, err)
		require.Contains(t, scrape(t, handler), `windows_exporter_collector_success{collector="test"} 0`)
		require.Equal(t, int32(0), test.calls.Load())
	})

	t.Run("waits for a running Collect", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			slow := &fakeCollector{name: "slow", delay: 2 * time.Second}
			collection := collector.New(collector.Map{"slow": slow})
			require.NoError(t, collection.Build(t.Context(), slog.New(slog.DiscardHandler)))

			handler, err := collection.NewHandler(100*time.Millisecond, slog.New(slog.DiscardHandler), nil)
			require.NoError(t, err)
			require.Contains(t, scrape(t, handler), `windows_exporter_collector_timeout{collector="slow"} 1`)

			require.NoError(t, collection.Close())
			require.Equal(t, int32(1), slow.closes.Load())
			require.Equal(t, int32(0), slow.closedWhileActive.Load())
		})
	})

	t.Run("leaves a hung collector open", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			hung := &fakeCollector{name: "hung", delay: time.Minute}
			collection := collector.New(collector.Map{"hung": hung})
			require.NoError(t, collection.Build(t.Context(), slog.New(slog.DiscardHandler)))

			handler, err := collection.NewHandler(100*time.Millisecond, slog.New(slog.DiscardHandler), nil)
			require.NoError(t, err)
			scrape(t, handler)

			require.ErrorContains(t, collection.Close(), "still running")
			require.Equal(t, int32(0), hung.closes.Load())

			// Once the call returns, a later Close closes the collector.
			time.Sleep(time.Minute)
			synctest.Wait()

			require.NoError(t, collection.Close())
			require.Equal(t, int32(1), hung.closes.Load())
			require.Equal(t, int32(0), hung.closedWhileActive.Load())
		})
	})
}

func TestBuildCloseCyclesDoNotLeakGoroutines(t *testing.T) {
	// The bubble also fails the test if any goroutine of it is still blocked at the end.
	synctest.Test(t, func(t *testing.T) {
		logger := slog.New(slog.DiscardHandler)

		cycle := func() {
			slow := &fakeCollector{name: "slow", delay: time.Second}
			fast := &fakeCollector{name: "fast"}
			collection := collector.New(collector.Map{"slow": slow, "fast": fast})
			require.NoError(t, collection.Build(t.Context(), logger))

			handler, err := collection.NewHandler(100*time.Millisecond, logger, nil)
			require.NoError(t, err)

			out := scrape(t, handler)
			require.Contains(t, out, `windows_exporter_collector_timeout{collector="slow"} 1`)
			require.Contains(t, out, `windows_exporter_collector_success{collector="fast"} 1`)

			require.NoError(t, collection.Close())
			require.Equal(t, int32(1), slow.closes.Load())
			require.Equal(t, int32(1), fast.closes.Load())
		}

		cycle()
		synctest.Wait()

		baseline := runtime.NumGoroutine()

		for range 20 {
			cycle()
		}

		synctest.Wait()

		// One leaked goroutine per cycle would add 20.
		require.LessOrEqual(t, runtime.NumGoroutine(), baseline+2)
	})
}

// fakeCollector is a collector whose Build and Collect behave as configured.
type fakeCollector struct {
	name         string
	buildErr     error
	collectErr   error
	buildPanic   bool
	buildStarted chan struct{}
	buildBlock   chan struct{}
	delay        time.Duration

	calls             atomic.Int32
	active            atomic.Int32
	maxActive         atomic.Int32
	closes            atomic.Int32
	closedWhileActive atomic.Int32
	budget            atomic.Int64
}

func (c *fakeCollector) GetName() string { return c.name }

func (c *fakeCollector) Build(_ *slog.Logger, _ *mi.Session) error {
	if c.buildStarted != nil {
		c.buildStarted <- struct{}{}
	}

	if c.buildBlock != nil {
		<-c.buildBlock
	}

	if c.buildPanic {
		panic("build panic")
	}

	return c.buildErr
}

func (c *fakeCollector) Collect(ch chan<- prometheus.Metric, maxScrapeDuration time.Duration) error {
	c.calls.Add(1)
	c.budget.Store(int64(maxScrapeDuration))

	active := c.active.Add(1)
	defer c.active.Add(-1)

	for {
		maxActive := c.maxActive.Load()
		if active <= maxActive || c.maxActive.CompareAndSwap(maxActive, active) {
			break
		}
	}

	time.Sleep(c.delay)

	ch <- prometheus.MustNewConstMetric(prometheus.NewDesc("windows_test_"+c.name, "Test metric", nil, nil), prometheus.GaugeValue, 42)

	return c.collectErr
}

func (c *fakeCollector) Close() error {
	c.closes.Add(1)

	if c.active.Load() != 0 {
		c.closedWhileActive.Add(1)
	}

	return nil
}

// blockingCollector is a collector whose Collect signals started and waits until block is closed.
type blockingCollector struct {
	*fakeCollector

	started chan struct{}
	block   chan struct{}
}

func newBlockingCollector() *blockingCollector {
	c := &blockingCollector{
		started: make(chan struct{}, 1),
		block:   make(chan struct{}),
	}
	c.fakeCollector = &fakeCollector{name: "test"}

	return c
}

func (c *blockingCollector) Collect(ch chan<- prometheus.Metric, maxScrapeDuration time.Duration) error {
	c.started <- struct{}{}

	<-c.block

	return c.fakeCollector.Collect(ch, maxScrapeDuration)
}

// scrapeInBackground starts a scrape with a handler bound to ctx and returns its output, or the error text.
func scrapeInBackground(ctx context.Context, collection *collector.Collection, timeout time.Duration) <-chan string {
	out := make(chan string, 1)

	handler, err := collection.NewHandlerWithContext(ctx, timeout, slog.New(slog.DiscardHandler), nil)
	if err != nil {
		out <- err.Error()

		return out
	}

	go func() {
		text, err := gather(handler)
		if err != nil {
			text = err.Error()
		}

		out <- text
	}()

	return out
}

func scrape(t *testing.T, handler *collector.Handler) string {
	t.Helper()

	out, err := gather(handler)
	require.NoError(t, err)

	return out
}

// gather collects handler once and returns the metrics in the text format.
func gather(handler *collector.Handler) (string, error) {
	registry := prometheus.NewRegistry()
	if err := registry.Register(handler); err != nil {
		return "", err
	}

	families, err := registry.Gather()
	if err != nil {
		return "", err
	}

	var out strings.Builder

	for _, family := range families {
		if _, err := expfmt.MetricFamilyToText(&out, family); err != nil {
			return "", err
		}
	}

	return out.String(), nil
}

func canceledContext() context.Context {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	return ctx
}

// recordHandler is a [slog.Handler] that records all messages.
type recordHandler struct {
	mu      sync.Mutex
	records []slog.Record
}

func (h *recordHandler) Enabled(context.Context, slog.Level) bool { return true }

func (h *recordHandler) Handle(_ context.Context, record slog.Record) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.records = append(h.records, record)

	return nil
}

func (h *recordHandler) WithAttrs([]slog.Attr) slog.Handler { return h }

func (h *recordHandler) WithGroup(string) slog.Handler { return h }

// warnings returns the number of warnings whose message contains substr.
func (h *recordHandler) warnings(substr string) int {
	h.mu.Lock()
	defer h.mu.Unlock()

	n := 0

	for _, record := range h.records {
		if record.Level == slog.LevelWarn && strings.Contains(record.Message, substr) {
			n++
		}
	}

	return n
}

func (h *recordHandler) reset() {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.records = nil
}
