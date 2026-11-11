// SPDX-License-Identifier: MIT
//
// Copyright The Prometheus Authors

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
