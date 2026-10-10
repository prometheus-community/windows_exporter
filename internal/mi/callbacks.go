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
	"errors"
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"
	"syscall"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"
)

// Queries deliver their results through MI_OperationCallbacks instead of the
// synchronous MI_Operation_GetInstance. With GetInstance, the WMI client hands
// every result from its delivery thread to the reading thread, and creates
// Event handles it never closes whenever that hand-over has to wait, e.g. for
// parallel queries or slow providers. A callback receives the result on the
// delivery thread, so there is no such hand-over.
//
// The callback runs on MI threads, which are foreign to the Go runtime:
//
//   - It is created once. windows.NewCallback never frees a callback and
//     supports only a limited number of them per process.
//   - The asyncQuery, which holds the MI_OperationCallbacks and is the
//     callback context, is pinned until the operation is closed.
//   - It never calls into the operation. Cancel and Close are called by the
//     goroutine that started the query, after the final callback for Close.
//   - It recovers panics, which must not unwind into native code, and never
//     blocks on a goroutine that stopped serving it.
//
// MI reports some failures, e.g. invalid parameters, from within
// MI_Session_QueryInstances. The runtime runs such a callback on the
// goroutine that started the query.

// operationCallbacks is MI_OperationCallbacks.
//
// https://learn.microsoft.com/en-us/windows/win32/api/mi/ns-mi-mi_operationcallbacks
type operationCallbacks struct {
	callbackContext         unsafe.Pointer
	promptUser              uintptr
	writeError              uintptr
	writeMessage            uintptr
	writeProgress           uintptr
	instanceResult          uintptr
	indicationResult        uintptr
	classResult             uintptr
	streamedParameterResult uintptr
}

// instanceResultCallback is the MI_OperationCallback_Instance of all queries.
// The callback context identifies the query.
//
// https://learn.microsoft.com/en-us/windows/win32/api/mi/nc-mi-mi_operationcallback_instance
//
//nolint:gochecknoglobals
var instanceResultCallback = sync.OnceValue(func() uintptr {
	return windows.NewCallback(func(
		_ *Operation,
		query *asyncQuery,
		instance *Instance,
		moreResults Boolean,
		resultCode ResultError,
		errorMessage *uint16,
		errorDetails *Instance,
		_ uintptr, // resultAcknowledgement, only set with MI_OPERATIONFLAGS_MANUAL_ACK_RESULTS.
	) uintptr {
		query.instanceResult(instance, moreResults == True, resultCode, errorMessage, errorDetails)

		return 0
	})
})

// deadlockGuardInterval is the period of the timer a query keeps pending
// while it waits for its callbacks.
//
// The runtime does not know that an MI thread will call back, so while every
// goroutine waits for MI, it declares a deadlock and exits
// (https://github.com/golang/go/issues/55015). A pending timer suppresses
// that check. The interval does not bound the query; MI timeouts do.
const deadlockGuardInterval = time.Minute

// guardTimers holds stopped deadlock guard timers for reuse.
//
//nolint:gochecknoglobals
var guardTimers = sync.Pool{
	New: func() any {
		timer := time.NewTimer(deadlockGuardInterval)
		timer.Stop()

		return timer
	},
}

// newGuardTimer returns a running deadlock guard timer. Return it with
// releaseGuardTimer.
func newGuardTimer() *time.Timer {
	timer := guardTimers.Get().(*time.Timer) //nolint:forcetypeassert
	timer.Reset(deadlockGuardInterval)

	return timer
}

func releaseGuardTimer(timer *time.Timer) {
	// Since Go 1.23, a stopped timer delivers no stale value after Reset.
	timer.Stop()
	guardTimers.Put(timer)
}

var (
	// errQueryAborted is returned to an instance callback when the waiting
	// goroutine stopped handling instances, e.g. because its handler panicked.
	errQueryAborted = errors.New("query aborted")

	// errStopQuery cancels a query without an error.
	errStopQuery = errors.New("stop query")

	// errInlineInstance fails a [Session.QueryFunc] query whose instance MI
	// delivers from within MI_Session_QueryInstances. MI only reports
	// parameter errors that way.
	errInlineInstance = errors.New("MI delivered an instance from within MI_Session_QueryInstances")
)

// callbackPanicError reports a panic recovered in an MI callback. The value
// is formatted by Error, outside the callback, because formatting can panic
// as well.
type callbackPanicError struct {
	value any
}

func (e *callbackPanicError) Error() string {
	return fmt.Sprintf("panic in MI instance callback: %v", e.value)
}

// asyncQuery is the state of one MI_Session_QueryInstances operation that
// delivers its results to instanceResultCallback. It is pinned while MI may
// use it.
type asyncQuery struct {
	callbacks operationCallbacks
	operation Operation
	pinner    runtime.Pinner

	// onInstance is called for every instance, inside the callback. An error
	// cancels the query and is returned by wait; errStopQuery cancels it
	// without an error.
	onInstance func(*Instance) error

	// handler is called for every instance on the waiting goroutine, which
	// receives them on instances and replies on handled. aborted is closed
	// when it stops doing so. All are nil unless created by newHandOverQuery.
	handler   func(*Instance) error
	instances chan *Instance
	handled   chan error
	aborted   chan struct{}

	// startThread is the ID of the thread that runs MI_Session_QueryInstances
	// while it runs, zero otherwise.
	startThread atomic.Uint32

	// cancelRequested asks the waiting goroutine to cancel the operation.
	cancelRequested chan struct{}
	// done is closed by the final callback. MI does not use the query any
	// more afterward, and err is final. finished guards against closing it
	// twice, which would panic into MI.
	done     chan struct{}
	finished atomic.Bool

	// mu serializes the callbacks. MI delivers the results of an operation
	// one at a time, but possibly on different threads, and its own ordering
	// is invisible to Go. The race detector cannot point out a missing lock
	// here: it treats every callback as synchronized with every native call.
	mu  sync.Mutex
	err error
}

// newAsyncQuery returns a query that calls onInstance inside the callback.
func newAsyncQuery(onInstance func(*Instance) error) *asyncQuery {
	return &asyncQuery{
		onInstance:      onInstance,
		cancelRequested: make(chan struct{}, 1),
		done:            make(chan struct{}),
	}
}

// newHandOverQuery returns a query that calls handler for every instance on
// the goroutine running [asyncQuery.wait]. The callback waits until handler
// returns, so the instance stays valid without a copy.
func newHandOverQuery(handler func(*Instance) error) *asyncQuery {
	query := newAsyncQuery(nil)
	query.onInstance = query.handOver
	query.handler = handler
	query.instances = make(chan *Instance)
	query.handled = make(chan error)
	query.aborted = make(chan struct{})

	return query
}

// start starts the query as an MI_Session_QueryInstances operation. The
// caller must call close afterward. Nil operationOptions fall back to the
// session defaults.
//
// MI_Session_QueryInstances returns void: a failed query still yields an
// operation, and its status, e.g. a timeout or an invalid class, is reported
// by the final callback.
func (q *asyncQuery) start(session *Session, flags OperationFlags, operationOptions *OperationOptions,
	namespaceName Namespace, queryDialect QueryDialect, queryExpression Query,
) {
	q.callbacks.callbackContext = unsafe.Pointer(q)
	q.callbacks.instanceResult = instanceResultCallback()

	q.pinner.Pin(q)

	if operationOptions == nil {
		operationOptions = session.defaultOperationOptions
	}

	// The thread ID identifies callbacks from within the call, see handOver.
	runtime.LockOSThread()
	q.startThread.Store(windows.GetCurrentThreadId())

	_, _, _ = syscall.SyscallN(
		session.ft.QueryInstances,
		uintptr(unsafe.Pointer(session)),
		uintptr(flags),
		uintptr(unsafe.Pointer(operationOptions)),
		uintptr(unsafe.Pointer(namespaceName)),
		uintptr(unsafe.Pointer(queryDialect)),
		uintptr(unsafe.Pointer(queryExpression)),
		uintptr(unsafe.Pointer(&q.callbacks)),
		uintptr(unsafe.Pointer(&q.operation)),
	)

	q.startThread.Store(0)
	runtime.UnlockOSThread()
}

// instanceResult handles one MI_OperationCallback_Instance call.
func (q *asyncQuery) instanceResult(instance *Instance, moreResults bool, resultCode ResultError,
	errorMessage *uint16, errorDetails *Instance,
) {
	q.mu.Lock()

	defer func() {
		// A panic must not unwind into MI. It fails the query instead.
		if r := recover(); r != nil {
			q.fail(&callbackPanicError{value: r})
		}

		q.mu.Unlock()

		// The final result is the last call MI makes with this context.
		if !moreResults && q.finished.CompareAndSwap(false, true) {
			close(q.done)
		}
	}()

	if !errors.Is(resultCode, MI_RESULT_OK) {
		// The first error wins. A query cancelled because of an error ends
		// with a failed result, which must not replace that error.
		if q.err == nil {
			q.err = fmt.Errorf("failed to get instance: %w", instanceResultError(resultCode, errorMessage, errorDetails))
		}

		return
	}

	// After an error or an early stop, the remaining results are skipped
	// until the cancellation takes effect.
	if q.err != nil || instance == nil {
		return
	}

	if err := q.onInstance(instance); err != nil {
		q.fail(err)
	}
}

// fail records err and asks the waiting goroutine to cancel the operation.
// q.mu must be held.
func (q *asyncQuery) fail(err error) {
	if q.err == nil {
		q.err = err
	}

	select {
	case q.cancelRequested <- struct{}{}:
	default:
	}
}

// handOver passes instance to the goroutine running wait and blocks until
// handler has returned.
func (q *asyncQuery) handOver(instance *Instance) error {
	// A callback from within MI_Session_QueryInstances runs on the goroutine
	// that would have to receive the instance, so a hand-over would deadlock.
	// Calling handler here would run it inside native frames, where a panic
	// cannot reach the caller and runtime.Goexit is fatal.
	if thread := q.startThread.Load(); thread != 0 && thread == windows.GetCurrentThreadId() {
		return errInlineInstance
	}

	select {
	case q.instances <- instance:
	case <-q.aborted:
		return errQueryAborted
	}

	// Every instance received is answered, also if handler panics.
	return <-q.handled
}

// wait blocks until the final callback and returns the query result. It
// cancels the operation when a callback asks for it, and runs the handler of
// a hand-over query.
func (q *asyncQuery) wait() error {
	guard := newGuardTimer()
	defer releaseGuardTimer(guard)

	for {
		select {
		case <-q.done:
			if errors.Is(q.err, errStopQuery) {
				return nil
			}

			return q.err
		case <-q.cancelRequested:
			_ = q.operation.Cancel()
		case instance := <-q.instances:
			q.handle(instance)
		case <-guard.C:
			guard.Reset(deadlockGuardInterval)
		}
	}
}

// handle runs handler for instance and replies to the callback waiting in
// handOver, also if handler panics or exits the goroutine.
func (q *asyncQuery) handle(instance *Instance) {
	err := errQueryAborted

	defer func() {
		q.handled <- err
	}()

	err = q.handler(instance)
}

// close cancels the operation unless it has finished, waits for the final
// callback, closes the operation and unpins the query. The goroutine that
// started the query must call it, also when wait did not return because a
// handler panicked.
func (q *asyncQuery) close() {
	if q.aborted != nil {
		close(q.aborted)
	}

	select {
	case <-q.done:
	default:
		_ = q.operation.Cancel()

		guard := newGuardTimer()
		defer releaseGuardTimer(guard)

		// A callback may have started a hand-over before aborted was closed.
		for finished := false; !finished; {
			select {
			case <-q.done:
				finished = true
			case <-q.instances:
				q.handled <- errQueryAborted
			case <-guard.C:
				guard.Reset(deadlockGuardInterval)
			}
		}
	}

	// MI_Operation_Close blocks until the final result has been delivered,
	// which has happened.
	_ = q.operation.close()

	q.pinner.Unpin()
}

// query runs a query whose instances are passed to onInstance inside the MI
// callback.
func (s *Session) query(flags OperationFlags, operationOptions *OperationOptions, namespaceName Namespace,
	queryDialect QueryDialect, queryExpression Query, onInstance func(*Instance) error,
) error {
	query := newAsyncQuery(onInstance)

	query.start(s, flags, operationOptions, namespaceName, queryDialect, queryExpression)
	defer query.close()

	return query.wait()
}
