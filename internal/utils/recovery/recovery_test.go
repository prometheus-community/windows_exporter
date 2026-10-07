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

package recovery_test

import (
	"errors"
	"sync/atomic"
	"testing"

	"github.com/prometheus-community/windows_exporter/internal/utils/recovery"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRun(t *testing.T) {
	t.Parallel()

	errExpected := errors.New("expected")

	require.NoError(t, recovery.Run(func() error { return nil }))
	require.ErrorIs(t, recovery.Run(func() error { return errExpected }), errExpected)

	err := recovery.Run(func() error {
		panic("boom")
	})
	require.ErrorIs(t, err, recovery.ErrPanic)
	assert.Contains(t, err.Error(), "boom")
	assert.Contains(t, err.Error(), "recovery_test.go", "error must contain the stack trace")
}

func TestGroup(t *testing.T) {
	t.Parallel()

	errExpected := errors.New("expected")

	var (
		g    recovery.Group
		done atomic.Int32
	)

	g.Go(func() error {
		done.Add(1)

		return nil
	})
	g.Go(func() error {
		done.Add(1)

		return errExpected
	})
	g.Go(func() error {
		done.Add(1)

		panic("boom")
	})
	g.Go(func() error {
		// Nested Go calls must be waited for, too.
		g.Go(func() error {
			done.Add(1)

			return nil
		})

		done.Add(1)

		return nil
	})

	err := g.Wait()
	require.ErrorIs(t, err, errExpected)
	require.ErrorIs(t, err, recovery.ErrPanic)
	assert.Contains(t, err.Error(), "boom")
	assert.Equal(t, int32(5), done.Load())

	var empty recovery.Group
	require.NoError(t, empty.Wait())
}
