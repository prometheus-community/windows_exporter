// SPDX-License-Identifier: MIT
//
// Copyright The Prometheus Authors

//go:build windows

package testutils

import (
	"errors"
	"fmt"
	"testing"

	"github.com/prometheus-community/windows_exporter/internal/pdh"
	"github.com/prometheus-community/windows_exporter/internal/types"
	"github.com/stretchr/testify/require"
)

func TestNoDataOnly(t *testing.T) {
	t.Parallel()

	failure := errors.New("collector failed")
	for _, tc := range []struct {
		name string
		err  error
		want bool
	}{
		{name: "nil"},
		{name: "PDH no data", err: pdh.ErrNoData, want: true},
		{name: "collector no data", err: types.ErrNoData, want: true},
		{name: "wrapped", err: fmt.Errorf("empty instances: %w", pdh.ErrNoData), want: true},
		{name: "joined empty groups", err: errors.Join(pdh.ErrNoData, types.ErrNoData), want: true},
		{name: "failure", err: failure},
		{name: "mixed failures", err: errors.Join(pdh.ErrNoData, failure)},
		{name: "nested mixed failures", err: fmt.Errorf("child: %w", errors.Join(pdh.ErrNoData, failure))},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			require.Equal(t, tc.want, noDataOnly(tc.err))
		})
	}
}
