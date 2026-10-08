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
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"golang.org/x/sys/windows/svc"
)

func TestServiceExecuteStop(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name                string
		exitCode            int
		wantServiceSpecific bool
		wantExitCode        uint32
	}{
		{name: "clean stop", exitCode: 0},
		// The main function fails, for example because the stop aborted its startup.
		{name: "failed stop", exitCode: 1, wantServiceSpecific: true, wantExitCode: 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			exitCodeCh := make(chan int, 1)
			service := &windowsExporterService{
				stop:       newBroadcast(),
				exitCodeCh: exitCodeCh,
				logEvent:   func(uint16, string) error { return nil },
			}

			requests := make(chan svc.ChangeRequest)
			changes := make(chan svc.Status, 10)

			type result struct {
				serviceSpecific bool
				exitCode        uint32
			}

			done := make(chan result, 1)

			go func() {
				serviceSpecific, exitCode := service.Execute(nil, requests, changes)
				done <- result{serviceSpecific, exitCode}
			}()

			requests <- svc.ChangeRequest{Cmd: svc.Stop}

			// Execute signals the stop without waiting for the main function to receive it.
			select {
			case <-service.stop.Done():
			case <-time.After(10 * time.Second):
				t.Fatal("Execute didn't signal the stop")
			}

			exitCodeCh <- tc.exitCode

			select {
			case got := <-done:
				require.Equal(t, result{tc.wantServiceSpecific, tc.wantExitCode}, got)
			case <-time.After(10 * time.Second):
				t.Fatal("Execute didn't return the exit code")
			}

			require.Equal(t, svc.StartPending, (<-changes).State)
			require.Equal(t, svc.Running, (<-changes).State)
			require.Equal(t, svc.StopPending, (<-changes).State)
		})
	}
}

func TestBroadcast(t *testing.T) {
	t.Parallel()

	b := newBroadcast()

	select {
	case <-b.Done():
		t.Fatal("broadcast is done before Close")
	default:
	}

	b.Close()
	b.Close()

	<-b.Done()
}

func TestRunStartupCanceled(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancelCause(t.Context())
	cancel(errServiceStop)

	exitCodeCh := make(chan int, 1)

	go func() {
		exitCodeCh <- run(ctx, []string{"--web.listen-address=127.0.0.1:0"})
	}()

	select {
	case exitCode := <-exitCodeCh:
		// A stop during startup is a clean stop.
		require.Equal(t, 0, exitCode)
	case <-time.After(30 * time.Second):
		t.Fatal("run didn't return after its context was canceled")
	}
}

func TestEffectiveCollectors(t *testing.T) {
	t.Parallel()

	require.Equal(t,
		[]string{"cpu", "memory", "os"},
		effectiveCollectors([]string{"os", "cpu", "net", "memory", "cpu"}, []string{"net", "unknown"}),
	)
	require.Equal(t,
		[]string{"cpu", "net"},
		effectiveCollectors([]string{"net", "cpu"}, nil),
	)
}
