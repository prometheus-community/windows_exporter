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
	"errors"
	"fmt"
	"log/slog"
	"testing"
	"time"

	"github.com/alecthomas/kingpin/v2"
	"github.com/prometheus-community/windows_exporter/internal/mi"
	"github.com/prometheus-community/windows_exporter/internal/pdh"
	"github.com/prometheus-community/windows_exporter/pkg/collector"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/require"
)

func TestCollectionConstructors(t *testing.T) {
	t.Parallel()

	t.Run("config includes every collector", func(t *testing.T) {
		t.Parallel()

		collection := collector.NewWithConfig(collector.Config{})
		_, err := collection.NewHandler(time.Second, slog.New(slog.DiscardHandler), collector.Available())
		require.NoError(t, err)
	})

	t.Run("flags include every collector", func(t *testing.T) {
		t.Parallel()

		app := kingpin.New("test", "test")
		collection := collector.NewWithFlags(app)
		_, err := app.Parse(nil)
		require.NoError(t, err)
		_, err = collection.NewHandler(time.Second, slog.New(slog.DiscardHandler), collector.Available())
		require.NoError(t, err)
	})
}

func TestCollectionBuildAndClose(t *testing.T) {
	t.Parallel()

	buildFailure := errors.New("build failed")

	closeFailure := errors.New("close failed")
	for _, tc := range []struct {
		name         string
		buildErr     error
		closeErr     error
		wantBuildErr error
	}{
		{name: "success"},
		{name: "build failure", buildErr: buildFailure, wantBuildErr: buildFailure},
		{name: "optional counter missing", buildErr: pdh.ErrNoData},
		{name: "unsupported on this host", buildErr: fmt.Errorf("feature missing: %w", errors.ErrUnsupported)},
		{name: "close failure", closeErr: closeFailure},
		{name: "all expected joined errors", buildErr: errors.Join(errors.ErrUnsupported, pdh.ErrNoData)},
		{name: "joined fatal error", buildErr: errors.Join(errors.ErrUnsupported, buildFailure), wantBuildErr: buildFailure},
		{name: "nested fatal error", buildErr: errors.Join(pdh.ErrNoData, fmt.Errorf("nested: %w", errors.Join(errors.ErrUnsupported, buildFailure))), wantBuildErr: buildFailure},
		{name: "multiple wrapped fatal error", buildErr: fmt.Errorf("%w and %w", errors.ErrUnsupported, buildFailure), wantBuildErr: buildFailure},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			test := &lifecycleCollector{name: "test", buildErr: tc.buildErr, closeErr: tc.closeErr}
			collection := collector.New(collector.Map{"test": test})
			err := collection.Build(t.Context(), slog.New(slog.DiscardHandler))
			t.Cleanup(func() {
				err := collection.Close()
				if tc.closeErr != nil {
					require.ErrorIs(t, err, tc.closeErr)
				} else {
					require.NoError(t, err)
				}

				require.True(t, test.closed)
			})

			if tc.wantBuildErr != nil {
				require.ErrorIs(t, err, tc.wantBuildErr)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

type lifecycleCollector struct {
	name     string
	buildErr error
	closeErr error
	closed   bool
}

func (c *lifecycleCollector) GetName() string { return c.name }

func (c *lifecycleCollector) Build(_ *slog.Logger, _ *mi.Session) error { return c.buildErr }

func (c *lifecycleCollector) Collect(_ chan<- prometheus.Metric, _ time.Duration) error { return nil }

func (c *lifecycleCollector) Close() error {
	c.closed = true

	return c.closeErr
}
