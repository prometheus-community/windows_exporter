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

package mi_test

import (
	"errors"
	"os"
	"os/exec"
	"sync"
	"testing"
	"time"

	"github.com/prometheus-community/windows_exporter/internal/mi"
	"github.com/prometheus-community/windows_exporter/internal/utils/testutils"
	"github.com/stretchr/testify/require"
	"golang.org/x/sys/windows"
)

type msftStoragePool struct {
	FriendlyName string `mi:"FriendlyName"`
}

// newTestSession returns a session that is closed with the test.
func newTestSession(tb testing.TB) *mi.Session {
	tb.Helper()

	application, err := mi.ApplicationInitialize()
	require.NoError(tb, err)

	tb.Cleanup(func() { require.NoError(tb, application.Close()) })

	session, err := application.NewSession(nil)
	require.NoError(tb, err)

	tb.Cleanup(func() { require.NoError(tb, session.Close()) })

	return session
}

// storagePoolQueryTimeout is generous: the storage provider is slow, and CI
// runs several test binaries at once. The tests check handles, not latency.
const storagePoolQueryTimeout = 30 * time.Second

// storagePoolQuery returns the MSFT_StoragePool query, or skips if the storage
// provider is not installed or not accessible, e.g. without elevation. Other
// errors fail the test.
func storagePoolQuery(tb testing.TB, session *mi.Session) mi.Query {
	tb.Helper()

	query, err := mi.NewQuery("SELECT FriendlyName FROM MSFT_StoragePool")
	require.NoError(tb, err)

	var pools []msftStoragePool

	err = session.Query(&pools, mi.NamespaceRootStorage, query, storagePoolQueryTimeout)
	if errors.Is(err, mi.MI_RESULT_INVALID_NAMESPACE) || errors.Is(err, mi.MI_RESULT_INVALID_CLASS) ||
		errors.Is(err, mi.MI_RESULT_NOT_SUPPORTED) || errors.Is(err, mi.MI_RESULT_ACCESS_DENIED) {
		tb.Skipf("MSFT_StoragePool is not available: %v", err)
	}

	require.NoError(tb, err)

	return query
}

// Test_MI_ParallelQuery_HandleGrowth runs parallel queries against the slow
// MSFT_StoragePool provider. With MI_Operation_GetInstance, the WMI client
// leaked Event handles whenever its result hand-over had to wait, which
// parallel queries against a slow provider do: about one per three queries.
func Test_MI_ParallelQuery_HandleGrowth(t *testing.T) {
	const (
		workers           = 8
		queriesPerWorker  = 50
		rounds            = 4
		maxHandleIncrease = 100
	)

	session := newTestSession(t)
	query := storagePoolQuery(t, session)

	round := func() {
		var wg sync.WaitGroup

		for range workers {
			wg.Go(func() {
				for i := range queriesPerWorker {
					// Both query paths deliver results through the callback.
					if i%2 == 0 {
						var pools []msftStoragePool

						if err := session.Query(&pools, mi.NamespaceRootStorage, query, storagePoolQueryTimeout); err != nil {
							t.Error(err)
						}

						continue
					}

					err := session.QueryFunc(mi.NamespaceRootStorage, query, storagePoolQueryTimeout, func(*mi.Instance) error {
						return nil
					})
					if err != nil {
						t.Error(err)
					}
				}
			})
		}

		wg.Wait()
	}

	handleCount := func() int64 {
		count, err := testutils.GetProcessHandleCount(windows.CurrentProcess())
		require.NoError(t, err)

		return int64(count)
	}

	// The first parallel round grows MI and Go thread pools.
	round()

	baseline := handleCount()

	for range rounds {
		round()
	}

	current := handleCount()

	t.Logf("handle count after warm-up %d, after %d queries %d", baseline, workers*queriesPerWorker*rounds, current)

	require.LessOrEqual(t, current, baseline+maxHandleIncrease,
		"handle count grew from %d to %d after %d parallel queries", baseline, current, workers*queriesPerWorker*rounds)
}

// Test_MI_QueryFunc_Panic checks that a panic in fn reaches the caller, after
// the query has been cancelled and closed, and leaves the session usable.
func Test_MI_QueryFunc_Panic(t *testing.T) {
	session := newTestSession(t)

	query, err := mi.NewQuery("SELECT Name FROM Win32_Process")
	require.NoError(t, err)

	for range 3 {
		var calls int

		require.PanicsWithValue(t, "fn panicked", func() {
			_ = session.QueryFunc(mi.NamespaceRootCIMv2, query, 5*time.Second, func(*mi.Instance) error {
				calls++

				panic("fn panicked")
			})
		})
		require.Equal(t, 1, calls)
	}

	var processes []win32Process

	require.NoError(t, session.Query(&processes, mi.NamespaceRootCIMv2, query, 5*time.Second))
	require.NotEmpty(t, processes)
}

// Test_MI_QueryFunc_Timeout checks that a timeout reaches QueryFunc as the
// native MI result.
func Test_MI_QueryFunc_Timeout(t *testing.T) {
	session := newTestSession(t)

	query, err := mi.NewQuery("SELECT * FROM Win32_Process")
	require.NoError(t, err)

	err = session.QueryFunc(mi.NamespaceRootCIMv2, query, time.Millisecond, func(*mi.Instance) error {
		return nil
	})
	require.ErrorIs(t, err, mi.MI_RESULT_INVALID_OPERATION_TIMEOUT)
}

// Test_MI_QueryFunc_Nested runs a query from within fn, which runs on the
// calling goroutine while MI waits for it.
func Test_MI_QueryFunc_Nested(t *testing.T) {
	session := newTestSession(t)

	outer, err := mi.NewQuery("SELECT Name FROM Win32_Process WHERE Handle = 0 OR Handle = 4")
	require.NoError(t, err)

	inner, err := mi.NewQuery("SELECT Name FROM Win32_Process WHERE Handle = 4")
	require.NoError(t, err)

	var calls int

	err = session.QueryFunc(mi.NamespaceRootCIMv2, outer, 5*time.Second, func(instance *mi.Instance) error {
		calls++

		var processes []win32Process

		if err := session.Query(&processes, mi.NamespaceRootCIMv2, inner, 5*time.Second); err != nil {
			return err
		}

		if len(processes) != 1 || processes[0].Name != "System" {
			return errors.New("unexpected inner result")
		}

		// The outer instance is still valid.
		element, err := instance.GetElement("Name")
		if err != nil {
			return err
		}

		_, err = element.String()

		return err
	})
	require.NoError(t, err)
	require.Equal(t, 2, calls)
}

// Test_MI_Query_InvalidParameter checks a failure MI reports from within
// MI_Session_QueryInstances, on the calling goroutine.
func Test_MI_Query_InvalidParameter(t *testing.T) {
	session := newTestSession(t)

	var processes []win32Process

	require.ErrorIs(t, session.Query(&processes, mi.NamespaceRootCIMv2, nil, 5*time.Second), mi.MI_RESULT_INVALID_PARAMETER)

	err := session.QueryFunc(nil, nil, 5*time.Second, func(*mi.Instance) error {
		return errors.New("unexpected instance")
	})
	require.ErrorIs(t, err, mi.MI_RESULT_INVALID_PARAMETER)
}

// Test_MI_Query_DeadlockDetector runs a query in a test binary without the
// test timeout, so no timer is pending while the only goroutine waits for MI
// callbacks. The runtime must not declare a deadlock
// (https://github.com/golang/go/issues/55015).
func Test_MI_Query_DeadlockDetector(t *testing.T) {
	if os.Getenv("MI_TEST_DEADLOCK_DETECTOR") == "1" {
		session := newTestSession(t)

		query, err := mi.NewQuery("SELECT * FROM Win32_Process")
		require.NoError(t, err)

		for range 5 {
			var processes []win32Process

			require.NoError(t, session.Query(&processes, mi.NamespaceRootCIMv2, query, -1))
			require.NoError(t, session.QueryFunc(mi.NamespaceRootCIMv2, query, -1, func(*mi.Instance) error { return nil }))
		}

		return
	}

	executable, err := os.Executable()
	require.NoError(t, err)

	cmd := exec.CommandContext(t.Context(), executable, "-test.run=^Test_MI_Query_DeadlockDetector$", "-test.timeout=0", "-test.count=1")

	cmd.Env = append(os.Environ(), "MI_TEST_DEADLOCK_DETECTOR=1")

	output, err := cmd.CombinedOutput()
	require.NoError(t, err, "%s", output)
}
