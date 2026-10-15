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

package collector

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"maps"
	"runtime/pprof"
	"slices"
	gotime "time"

	"github.com/alecthomas/kingpin/v2"
	"github.com/prometheus-community/windows_exporter/internal/mi"
	"github.com/prometheus-community/windows_exporter/internal/pdh"
	"github.com/prometheus-community/windows_exporter/internal/types"
	"github.com/prometheus/client_golang/prometheus"
	winregistry "golang.org/x/sys/windows/registry"
)

// NewWithFlags To be called by the exporter for collector initialization before running kingpin.Parse.
func NewWithFlags(app *kingpin.Application) *Collection {
	collectors := map[string]Collector{}

	for name, builder := range BuildersWithFlags {
		collectors[name] = builder(app)
	}

	return New(collectors)
}

// NewWithConfig To be called by the external libraries for collector initialization without running [kingpin.Parse].
//
//goland:noinspection GoUnusedExportedFunction
func NewWithConfig(config Config) *Collection {
	collectors := make(Map, len(registrations))
	for _, r := range registrations {
		collectors[r.name] = r.newWithConfig(&config)
	}

	return New(collectors)
}

// New To be called by the external libraries for collector initialization.
func New(collectors Map) *Collection {
	return &Collection{
		collectors: collectors,
		state:      newCollectionState(),
		scrapeDurationDesc: prometheus.NewDesc(
			prometheus.BuildFQName(types.Namespace, "exporter", "scrape_duration_seconds"),
			"windows_exporter: Total scrape duration.",
			nil,
			nil,
		),
		collectorScrapeDurationDesc: prometheus.NewDesc(
			prometheus.BuildFQName(types.Namespace, "exporter", "collector_duration_seconds"),
			"windows_exporter: Duration of a collection.",
			[]string{"collector"},
			nil,
		),
		collectorScrapeSuccessDesc: prometheus.NewDesc(
			prometheus.BuildFQName(types.Namespace, "exporter", "collector_success"),
			"windows_exporter: Whether the collector was successful.",
			[]string{"collector"},
			nil,
		),
		collectorScrapeTimeoutDesc: prometheus.NewDesc(
			prometheus.BuildFQName(types.Namespace, "exporter", "collector_timeout"),
			"windows_exporter: Whether the collector timed out.",
			[]string{"collector"},
			nil,
		),
	}
}

// Enable removes all collectors that not enabledCollectors.
func (c *Collection) Enable(enabledCollectors []string) error {
	for _, name := range enabledCollectors {
		if _, ok := c.collectors[name]; !ok {
			return fmt.Errorf("unknown collector %s", name)
		}
	}

	for name := range c.collectors {
		if !slices.Contains(enabledCollectors, name) {
			delete(c.collectors, name)
		}
	}

	return nil
}

// Disable removes all collectors that are listed in disabledCollectors.
func (c *Collection) Disable(disabledCollectors []string) {
	for name := range c.collectors {
		if slices.Contains(disabledCollectors, name) {
			delete(c.collectors, name)
		}
	}
}

// closeTimeout bounds how long Close waits for running Build and Collect calls to return.
const closeTimeout = 5 * gotime.Second

// Build To be called by the exporter for collector initialization.
// Instead, fail fast, it will try to build all collectors and return all errors.
// errors are joined with errors.Join.
//
// A collector whose Build fails is not collected. Scrapes report it with
// windows_exporter_collector_success 0. A panic in a collector's Build is returned as an error.
//
// If ctx is done before all collectors are built, Build returns without waiting for them.
// Collectors that are still building finish in the background, and Close waits for them.
func (c *Collection) Build(ctx context.Context, logger *slog.Logger) error {
	c.startTime = gotime.Now()

	if err := ctx.Err(); err != nil {
		return fmt.Errorf("collector initialization aborted: %w", context.Cause(ctx))
	}

	miSession, err := c.initMI()
	if err != nil {
		return fmt.Errorf("error from initialize MI: %w", err)
	}

	type buildResult struct {
		name string
		err  error
	}

	// The channel is buffered, so that builds abandoned by a done ctx don't block.
	resultCh := make(chan buildResult, len(c.collectors))

	for name, collector := range c.collectors {
		state := c.state.collector(name)

		go func() {
			// Workers started by Build inherit the collector's profiling label.
			pprof.Do(ctx, pprof.Labels("collector", name), func(ctx context.Context) {
				resultCh <- buildResult{
					name: collector.GetName(),
					err:  state.build(ctx, logger, miSession, collector),
				}
			})
		}()
	}

	errs := make([]error, 0, len(c.collectors))

	for range len(c.collectors) {
		var result buildResult

		select {
		case result = <-resultCh:
		case <-ctx.Done():
			errs = append(errs, fmt.Errorf("collector initialization aborted: %w", context.Cause(ctx)))

			return errors.Join(errs...)
		}

		if result.err == nil {
			continue
		}

		err := fmt.Errorf("error build collector %s: %w", result.name, result.err)

		// errors.ErrUnsupported marks a collector whose subsystem is not
		// available on this host, e.g. a Windows feature that is not installed.
		if expectedBuildError(err) {
			logger.LogAttrs(ctx, slog.LevelWarn, "couldn't initialize collector",
				slog.Any("err", err),
			)

			continue
		}

		errs = append(errs, err)
	}

	return errors.Join(errs...)
}

// Close To be called by the exporter for collector cleanup.
//
// Close waits up to 5 seconds for running Build and Collect calls to return.
// A collector that is still running after that is left open and reported in the returned error,
// because closing it would release resources that are still in use.
// The MI session is then left open as well.
//
// Close does nothing on a filtered view returned by WithCollectors. The original
// Collection owns the shared collectors and MI resources and must be closed.
//
// Close is safe to call after a failed or partial Build and more than once.
// Collectors that were never built are not closed. A closed collector is never built or collected again.
func (c *Collection) Close() error {
	if c.isView {
		return nil
	}

	c.state.closeMu.Lock()
	defer c.state.closeMu.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), closeTimeout)
	defer cancel()

	errs := make([]error, 0, len(c.collectors))
	collectorBusy := false

	for name, collector := range c.collectors {
		if err := c.state.collector(name).close(ctx, collector); err != nil {
			collectorBusy = collectorBusy || errors.Is(err, errCollectorBusy)

			errs = append(errs, fmt.Errorf("error from close collector %s: %w", collector.GetName(), err))
		}
	}

	if collectorBusy {
		// Closing the MI session would block on, or break, the operations of running collectors.
		errs = append(errs, errors.New("MI session left open, because collectors are still running"))
	} else {
		errs = append(errs, c.state.closeMI()...)
	}

	return errors.Join(errs...)
}

// initMI To be called by the exporter for collector initialization.
// It returns the session of an earlier call, if there is one.
func (c *Collection) initMI() (*mi.Session, error) {
	if session := c.state.session(); session != nil {
		return session, nil
	}

	app, err := mi.ApplicationInitialize()
	if err != nil {
		return nil, fmt.Errorf("error from initialize MI application: %w", err)
	}

	destinationOptions, err := app.NewDestinationOptions()
	if err != nil {
		_ = app.Close()

		return nil, fmt.Errorf("error from create NewDestinationOptions: %w", err)
	}

	session, err := newMISession(app, destinationOptions)
	if err != nil {
		_ = app.Close()

		return nil, err
	}

	c.state.setMI(app, session)

	return session, nil
}

// newMISession consumes destinationOptions. The session retains its own options;
// release the caller's options before initMI can close the application on failure.
func newMISession(app *mi.Application, destinationOptions *mi.DestinationOptions) (*mi.Session, error) {
	defer func() { _ = destinationOptions.Delete() }()

	if err := destinationOptions.SetLocale(mi.LocaleEnglish); err != nil {
		return nil, fmt.Errorf("error from set locale: %w", err)
	}

	if err := destinationOptions.SetTimeout(gotime.Second); err != nil {
		return nil, fmt.Errorf("error from set timeout: %w", err)
	}

	session, err := app.NewSession(destinationOptions)
	if err != nil {
		return nil, fmt.Errorf("error from create NewSession: %w", err)
	}

	return session, nil
}

// WithCollectors To be called by the exporter for collector initialization.
// The returned Collection shares the collector instances and their state with c,
// so a collector never runs twice at the same time, whichever Collection scrapes it.
// It does not own these resources: Close on a view does nothing. Close the original
// Collection when all of its views are no longer needed.
func (c *Collection) WithCollectors(collectors []string) (*Collection, error) {
	metricCollectors := &Collection{
		startTime:                   c.startTime,
		state:                       c.state,
		isView:                      true,
		scrapeDurationDesc:          c.scrapeDurationDesc,
		collectorScrapeDurationDesc: c.collectorScrapeDurationDesc,
		collectorScrapeSuccessDesc:  c.collectorScrapeSuccessDesc,
		collectorScrapeTimeoutDesc:  c.collectorScrapeTimeoutDesc,
		collectors:                  maps.Clone(c.collectors),
	}

	if err := metricCollectors.Enable(collectors); err != nil {
		return nil, err
	}

	return metricCollectors, nil
}

func (c *Collection) GetStartTime() gotime.Time {
	return c.startTime
}

// expectedBuildError tolerates an error tree only if every independent cause is
// an expected initialization failure for an optional Windows subsystem.
func expectedBuildError(err error) bool {
	if joined, ok := err.(interface{ Unwrap() []error }); ok {
		causes := joined.Unwrap()
		for _, cause := range causes {
			if !expectedBuildError(cause) {
				return false
			}
		}

		return len(causes) != 0
	}

	if wrapped, ok := err.(interface{ Unwrap() error }); ok {
		return expectedBuildError(wrapped.Unwrap())
	}

	return errors.Is(err, errors.ErrUnsupported) ||
		errors.Is(err, pdh.ErrNoData) ||
		errors.Is(err, winregistry.ErrNotExist) ||
		errors.Is(err, pdh.NewPdhError(pdh.CstatusNoObject)) ||
		errors.Is(err, pdh.NewPdhError(pdh.CstatusNoCounter)) ||
		errors.Is(err, mi.MI_RESULT_INVALID_OPERATION_TIMEOUT) ||
		errors.Is(err, mi.MI_RESULT_INVALID_NAMESPACE) ||
		errors.Is(err, mi.MI_RESULT_INVALID_CLASS)
}
