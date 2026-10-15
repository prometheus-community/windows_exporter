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

package mi

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// newInternalTestSession returns a session and its application. The caller
// closes the session; the application is closed with the test.
func newInternalTestSession(tb testing.TB) *Session {
	tb.Helper()

	application, err := ApplicationInitialize()
	require.NoError(tb, err)

	tb.Cleanup(func() { require.NoError(tb, application.Close()) })

	session, err := application.NewSession(nil)
	require.NoError(tb, err)

	return session
}

func TestSession_OperationOptions(t *testing.T) {
	session := newInternalTestSession(t)

	t.Run("non-positive and default timeouts use the session defaults", func(t *testing.T) {
		for _, timeout := range []time.Duration{0, -1, -time.Hour, DefaultQueryTimeout} {
			operationOptions, release, err := session.operationOptions(timeout)
			require.NoError(t, err)
			require.Same(t, session.defaultOperationOptions, operationOptions)

			release()
		}
	})

	t.Run("fixed timeouts are cached", func(t *testing.T) {
		first, release, err := session.operationOptions(BuildQueryTimeout)
		require.NoError(t, err)
		release()

		second, release, err := session.operationOptions(BuildQueryTimeout)
		require.NoError(t, err)
		release()

		require.Same(t, first, second)
		require.NotSame(t, session.defaultOperationOptions, first)
		require.Same(t, first, session.cachedOperationOptions[BuildQueryTimeout])
	})

	t.Run("other timeouts get options of their own", func(t *testing.T) {
		for _, timeout := range []time.Duration{time.Microsecond, 1500 * time.Millisecond} {
			operationOptions, release, err := session.operationOptions(timeout)
			require.NoError(t, err)
			require.NotSame(t, session.defaultOperationOptions, operationOptions)
			require.NotContains(t, session.cachedOperationOptions, timeout)

			release()
		}
	})

	t.Run("the cache is bounded", func(t *testing.T) {
		for i := range maxCachedOperationOptions + 2 {
			_, release, err := session.operationOptions(time.Duration(100+i) * time.Second)
			require.NoError(t, err)

			release()
		}

		require.Len(t, session.cachedOperationOptions, maxCachedOperationOptions)
	})

	require.NoError(t, session.Close())
	require.Nil(t, session.cachedOperationOptions)
	require.Nil(t, session.defaultOperationOptions)

	for _, timeout := range []time.Duration{0, BuildQueryTimeout, 1500 * time.Millisecond} {
		_, _, err := session.operationOptions(timeout)
		require.ErrorIs(t, err, ErrNotInitialized, timeout)
	}
}

// TestSession_CloseWaitsForStartingQuery checks that Close does not delete
// operation options while a query is being started, and that queries fail
// after Close instead of starting with deleted options.
func TestSession_CloseWaitsForStartingQuery(t *testing.T) {
	session := newInternalTestSession(t)

	queryExpression, err := NewQuery("SELECT Name FROM Win32_Process")
	require.NoError(t, err)

	// A query between checking closed and MI_Session_QueryInstances holds
	// this read lock.
	session.lifecycle.RLock()

	closed := make(chan error, 1)

	go func() { closed <- session.Close() }()

	select {
	case err := <-closed:
		session.lifecycle.RUnlock()
		require.Failf(t, "Close returned while a query was being started", "err: %v", err)
	case <-time.After(200 * time.Millisecond):
	}

	require.NotNil(t, session.defaultOperationOptions, "options deleted while a query was being started")

	session.lifecycle.RUnlock()

	select {
	case err := <-closed:
		require.NoError(t, err)
	case <-time.After(time.Minute):
		require.FailNow(t, "Close did not return")
	}

	require.ErrorIs(t, session.QueryFunc(NamespaceRootCIMv2, queryExpression, BuildQueryTimeout, func(*Instance) error { return nil }),
		ErrNotInitialized)
	require.ErrorIs(t, session.QueryUnmarshal(&[]struct {
		Name string `mi:"Name"`
	}{}, OperationFlagsStandardRTTI, nil, NamespaceRootCIMv2, QueryDialectWQL, queryExpression), ErrNotInitialized)
	require.ErrorIs(t, session.Close(), ErrNotInitialized)
}

// TestSession_CloseDuringQueries closes a session while queries with
// default, cached and per-query options start and run on other goroutines.
// Every query either runs, is cancelled by MI_Session_Close, or fails with
// ErrNotInitialized; none may start with deleted options.
func TestSession_CloseDuringQueries(t *testing.T) {
	session := newInternalTestSession(t)

	queryExpression, err := NewQuery("SELECT Name FROM Win32_Process")
	require.NoError(t, err)

	var (
		wg      sync.WaitGroup
		started = make(chan struct{})
		once    sync.Once
	)

	for _, timeout := range []time.Duration{0, BuildQueryTimeout, 1500 * time.Millisecond} {
		for range 2 {
			wg.Go(func() {
				for {
					err := session.QueryFunc(NamespaceRootCIMv2, queryExpression, timeout, func(*Instance) error {
						once.Do(func() { close(started) })

						return nil
					})
					if errors.Is(err, ErrNotInitialized) {
						return
					}
				}
			})
		}
	}

	<-started
	time.Sleep(50 * time.Millisecond)

	require.NoError(t, session.Close())

	wg.Wait()
}

// Test_asyncQuery_ContextDoneAfterFinalResult checks a context that is done
// after MI's final callback. Instances the handler did not see fail the
// query, while a result that was complete stays successful.
func Test_asyncQuery_ContextDoneAfterFinalResult(t *testing.T) {
	session := newInternalTestSession(t)

	t.Cleanup(func() { require.NoError(t, session.Close()) })

	queryExpression, err := NewQuery("SELECT Name FROM Win32_Process")
	require.NoError(t, err)

	t.Run("unhandled copies", func(t *testing.T) {
		ctx, cancel := context.WithCancel(t.Context())
		defer cancel()

		var handled int

		query := newCopyQuery(func(*Instance) error {
			handled++

			cancel()

			return nil
		})

		query.start(session, OperationFlagsStandardRTTI, nil, NamespaceRootCIMv2, QueryDialectWQL, queryExpression)
		defer query.close()

		// Every copy is queued and the final callback has run before wait.
		<-query.done

		require.ErrorIs(t, query.wait(ctx), context.Canceled)
		require.Equal(t, 1, handled)
	})

	t.Run("complete result", func(t *testing.T) {
		ctx, cancel := context.WithCancel(t.Context())

		var instances int

		query := newAsyncQuery(func(*Instance) error {
			instances++

			return nil
		})

		query.start(session, OperationFlagsStandardRTTI, nil, NamespaceRootCIMv2, QueryDialectWQL, queryExpression)
		defer query.close()

		<-query.done
		cancel()

		require.NoError(t, query.wait(ctx))
		require.Positive(t, instances)
	})
}

// Test_asyncQuery_ContextCancel checks that a cancelled query returns only
// after MI's final callback, stops calling the handler, and leaves no queued
// copies behind once closed.
func Test_asyncQuery_ContextCancel(t *testing.T) {
	session := newInternalTestSession(t)

	t.Cleanup(func() { require.NoError(t, session.Close()) })

	queryExpression, err := NewQuery("SELECT Name FROM Win32_Process")
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	var handled int

	query := newCopyQuery(func(*Instance) error {
		handled++

		cancel()

		return nil
	})

	query.start(session, OperationFlagsStandardRTTI, nil, NamespaceRootCIMv2, QueryDialectWQL, queryExpression)

	err = query.wait(ctx)
	require.ErrorIs(t, err, context.Canceled)
	require.Equal(t, 1, handled, "the handler must not run after the context is done")
	require.True(t, query.finished.Load(), "wait returned before the final callback")

	select {
	case <-query.done:
	default:
		require.Fail(t, "wait returned before the final callback")
	}

	query.close()

	require.Nil(t, query.pending)
	require.Nil(t, query.batch)
}

// Benchmark_Session_OperationOptions compares cached operation options for a
// fixed timeout with options created for every query.
func Benchmark_Session_OperationOptions(b *testing.B) {
	session := newInternalTestSession(b)

	b.Cleanup(func() { require.NoError(b, session.Close()) })

	for _, bc := range []struct {
		name    string
		timeout time.Duration
	}{
		{name: "cached", timeout: 5 * time.Second},
		{name: "per-query", timeout: 5*time.Second + time.Microsecond},
	} {
		b.Run(bc.name, func(b *testing.B) {
			b.ReportAllocs()

			for b.Loop() {
				_, release, err := session.operationOptions(bc.timeout)
				if err != nil {
					b.Fatal(err)
				}

				release()
			}
		})
	}
}
