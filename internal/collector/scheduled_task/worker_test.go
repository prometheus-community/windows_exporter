// SPDX-License-Identifier: MIT
//
// Copyright The Prometheus Authors

//go:build windows

package scheduled_task

import (
	"context"
	"testing"

	"github.com/prometheus-community/windows_exporter/internal/ole"
)

// HangWorker blocks the COM worker of c, like a Task Scheduler call that does
// not return, until the returned function is called. It returns once the
// worker is blocked.
func HangWorker(t *testing.T, c *Collector) func() {
	t.Helper()

	running := make(chan struct{})
	unblock := make(chan struct{})
	done := make(chan struct{})

	go func() {
		defer close(done)

		_, _ = ole.Do(context.Background(), c.worker, func() (struct{}, error) {
			close(running)
			<-unblock

			return struct{}{}, nil
		})
	}()

	<-running

	return func() {
		close(unblock)
		<-done
	}
}
