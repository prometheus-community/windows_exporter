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
