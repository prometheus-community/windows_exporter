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

//go:generate go run github.com/tc-hib/go-winres@v0.3.3 make --product-version=git-tag --file-version=git-tag --arch=amd64,arm64

package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"maps"
	"math"
	"net/http"
	"net/http/httptest"
	"net/http/pprof"
	"os"
	"os/signal"
	"os/user"
	"runtime"
	"runtime/debug"
	"slices"
	"strings"
	"syscall"
	"time"

	"github.com/alecthomas/kingpin/v2"
	"github.com/prometheus-community/windows_exporter/internal/config"
	"github.com/prometheus-community/windows_exporter/internal/httphandler"
	"github.com/prometheus-community/windows_exporter/internal/log"
	"github.com/prometheus-community/windows_exporter/internal/log/flag"
	"github.com/prometheus-community/windows_exporter/internal/utils"
	"github.com/prometheus-community/windows_exporter/pkg/collector"
	"github.com/prometheus/common/version"
	"github.com/prometheus/exporter-toolkit/web"
	webflag "github.com/prometheus/exporter-toolkit/web/kingpinflag"
	"golang.org/x/sys/windows"
)

// collectionCloseTimeout bounds how long the exporter waits for the collectors to close on shutdown.
const collectionCloseTimeout = 10 * time.Second

// healthPath is the URL path of the health endpoint.
const healthPath = "/health"

func main() {
	// os.Kill can't be caught. On Windows, Go delivers console close, logoff and shutdown events as SIGTERM.
	//nolint:forbidigo // os/signal only accepts syscall.Signal values, not windows.Signal.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)

	// A stop request from the service manager cancels the root context,
	// so that it aborts the startup as well as the running exporter.
	ctx, cancel := context.WithCancelCause(ctx)

	go func() {
		select {
		case <-serviceStop.Done():
			cancel(errServiceStop)
		case <-ctx.Done():
		}
	}()

	exitCode := run(ctx, os.Args[1:])

	cancel(nil)
	stop()

	// If we are running as a service, we need to signal the service control manager that we are done.
	if !IsService {
		os.Exit(exitCode)
	}

	exitCodeCh <- exitCode

	// Wait for the service control manager to signal that we are done.
	<-serviceManagerFinishedCh
}

func run(ctx context.Context, args []string) int {
	startTime := time.Now()

	app := kingpin.New("windows_exporter", "A metrics collector for Windows.")

	var (
		scrapeStdout = app.Flag(
			"scrape.stdout",
			"Collect metrics once and write them to stdout in Prometheus text format, then exit. "+
				"Logs are sent to stderr unless --log.file is set explicitly. "+
				"PDH rate counters need two samples; the first scrape after startup covers only a short interval.",
		).Default("false").Bool()
		configFile = app.Flag(
			"config.file",
			"YAML configuration file to use. Values set in this file will be overridden by CLI flags.",
		).String()
		webConfig   = webflag.AddFlags(app, ":9182")
		metricsPath = app.Flag(
			"telemetry.path",
			"URL path for surfacing collected metrics.",
		).Default("/metrics").String()
		disableExporterMetrics = app.Flag(
			"web.disable-exporter-metrics",
			"Exclude metrics about the exporter itself (promhttp_*, process_*, go_*).",
		).Bool()
		enabledCollectors = app.Flag(
			"collectors.enabled",
			"Comma-separated list of collectors to use. Use '[defaults]' as a placeholder for all the collectors enabled by default.").
			Default(collector.DefaultCollectors).String()
		disabledCollectors = app.Flag(
			"collectors.disabled",
			"Comma-separated list of collectors to exclude. Can be used to disable collector from the defaults.").
			Default("").String()
		timeoutMargin = app.Flag(
			"scrape.timeout-margin",
			"Seconds to subtract from the timeout allowed by the client. The margin takes at most half of the client's timeout. Tune to allow for overhead or high loads.",
		).Default("0.5").Float64()
		debugEnabled = app.Flag(
			"debug.enabled",
			"If true, windows_exporter will expose debug endpoints under /debug/pprof.",
		).Default("false").Bool()
		processPriority = app.Flag(
			"process.priority",
			"Priority of the exporter process. Higher priorities may improve exporter responsiveness during periods of system load. Can be one of [\"realtime\", \"high\", \"abovenormal\", \"normal\", \"belownormal\", \"low\"]",
		).Default("normal").String()
		memoryLimit = app.Flag(
			"process.memory-limit",
			"Limit memory usage in bytes. This is a soft-limit and not guaranteed. 0 means no limit. Negative values are invalid. Read more at https://pkg.go.dev/runtime/debug#SetMemoryLimit .",
		).Default("200000000").Int64()
	)

	logFile := &log.AllowedFile{}

	_ = logFile.Set("stdout")
	if IsService {
		_ = logFile.Set("eventlog")
	}

	logConfig := &log.Config{File: logFile}
	flag.AddFlags(app, logConfig)

	app.Version(version.Print("windows_exporter"))
	app.HelpFlag.Short('h')

	// Initialize collectors before loading and parsing CLI arguments
	collectors := collector.NewWithFlags(app)

	if err := config.Parse(app, args); err != nil {
		logStartupError(ctx, "Failed to load configuration", err)

		return 1
	}

	if err := setProcessMemoryLimit(*memoryLimit); err != nil {
		logStartupError(ctx, "Invalid process memory limit", err)

		return 1
	}

	if err := validateTimeoutMargin(*timeoutMargin); err != nil {
		logStartupError(ctx, "Invalid scrape timeout margin", err)

		return 1
	}

	// Check the routes before the collectors are built, so that an invalid
	// telemetry path fails the startup instead of panicking in the router.
	if _, err := newServeMux(*metricsPath, *debugEnabled, http.NotFoundHandler()); err != nil {
		logStartupError(ctx, "Invalid telemetry path", err)

		return 1
	}

	// When scraping to stdout, redirect logs to stderr so they don't
	// mix with the metrics output. Only override if the user hasn't
	// explicitly set --log.file to something other than the default.
	if *scrapeStdout && logFile.String() == "stdout" {
		_ = logFile.Set("stderr")
	}

	logger, logCloser, err := log.New(logConfig)
	if err != nil {
		logStartupError(ctx, "failed to create logger", err)

		return 1
	}

	if logCloser != nil {
		defer func() {
			_ = logCloser.Close()
		}()
	}

	logger.LogAttrs(ctx, slog.LevelDebug, "logging has Started")

	if configFile != nil && *configFile != "" {
		logger.LogAttrs(ctx, slog.LevelInfo, "using configuration file: "+*configFile)
	}

	// Log the build and the account before the collectors are built,
	// so that this information is available if a collector fails to initialize.
	logger.LogAttrs(ctx, slog.LevelInfo, "starting windows_exporter",
		slog.String("version", version.Version),
		slog.String("branch", version.Branch),
		slog.String("revision", version.GetRevision()),
		slog.String("goversion", version.GoVersion),
		slog.String("builddate", version.BuildDate),
		slog.Int("maxprocs", runtime.GOMAXPROCS(0)),
	)

	logCurrentUser(ctx, logger)

	if err = setPriorityWindows(ctx, logger, os.Getpid(), *processPriority); err != nil {
		logger.LogAttrs(ctx, slog.LevelError, "failed to set process priority",
			slog.Any("err", err),
		)

		return 1
	}

	enabledCollectorList := expandEnabledCollectors(*enabledCollectors)
	if err := collectors.Enable(enabledCollectorList); err != nil {
		logger.LogAttrs(ctx, slog.LevelError, "couldn't enable collectors",
			slog.Any("err", err),
		)

		return 1
	}

	var disabledCollectorList []string
	if *disabledCollectors != "" {
		disabledCollectorList = slices.Compact(strings.Split(*disabledCollectors, ","))
		collectors.Disable(disabledCollectorList)

		// An unknown name in --collectors.enabled fails the startup. Only warn about one in
		// --collectors.disabled: disabling a collector that does not exist leaves nothing
		// running that should not, and a failure would break configurations that disable a
		// collector which a later release removed.
		if unknown := unknownCollectors(disabledCollectorList); len(unknown) > 0 {
			logger.LogAttrs(ctx, slog.LevelWarn, "ignoring unknown collectors in --collectors.disabled",
				slog.Any("collectors", unknown),
			)
		}
	}

	// Close the collectors on every return from here on. On a regular shutdown, this runs after the HTTP server has shut down.
	defer closeCollection(ctx, logger, collectors)

	// Initialize collectors before loading
	if err = collectors.Build(ctx, logger); err != nil {
		if ctx.Err() != nil {
			logger.LogAttrs(ctx, slog.LevelInfo, "windows_exporter startup aborted",
				slog.Any("reason", context.Cause(ctx)),
			)

			return 0
		}

		for _, err := range utils.SplitError(err) {
			logger.LogAttrs(ctx, slog.LevelError, "couldn't initialize collector",
				slog.Any("err", err),
			)
		}

		return 1
	}

	logger.InfoContext(ctx, "Enabled collectors: "+strings.Join(effectiveCollectors(enabledCollectorList, disabledCollectorList), ", "))

	handlerOpts := &httphandler.Options{
		DisableExporterMetrics: *disableExporterMetrics,
		TimeoutMargin:          *timeoutMargin,
	}

	if *scrapeStdout {
		return scrapeToStdout(ctx, os.Stdout, logger, collectors, handlerOpts)
	}

	mux, err := newServeMux(*metricsPath, *debugEnabled, httphandler.New(logger, collectors, handlerOpts))
	if err != nil {
		logger.LogAttrs(ctx, slog.LevelError, "Invalid telemetry path",
			slog.Any("err", err),
		)

		return 1
	}

	logger.LogAttrs(ctx, slog.LevelInfo, fmt.Sprintf("started windows_exporter in %s", time.Since(startTime)))

	server := &http.Server{
		ReadHeaderTimeout: 5 * time.Second,
		IdleTimeout:       60 * time.Second,
		ReadTimeout:       5 * time.Second,
		WriteTimeout:      5 * time.Minute,
		Handler:           mux,
	}

	errCh := make(chan error, 1)

	go func() {
		if err := web.ListenAndServe(server, webConfig, logger); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}

		close(errCh)
	}()

	select {
	case <-ctx.Done():
		if errors.Is(context.Cause(ctx), errServiceStop) {
			logger.LogAttrs(ctx, slog.LevelInfo, "Shutting down windows_exporter via service control")
		} else {
			logger.LogAttrs(ctx, slog.LevelInfo, "Shutting down windows_exporter via kill signal")
		}
	case err := <-errCh:
		if err != nil {
			logger.LogAttrs(ctx, slog.LevelError, "Failed to start windows_exporter",
				slog.Any("err", err),
			)

			return 1
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	//nolint:contextcheck // create a new context for server shutdown
	if err = server.Shutdown(ctx); err != nil {
		//nolint:contextcheck
		logger.LogAttrs(ctx, slog.LevelError, "Failed to shutdown windows_exporter",
			slog.Any("err", err),
		)
	} else {
		//nolint:contextcheck
		logger.LogAttrs(ctx, slog.LevelInfo, "windows_exporter has shut down")
	}

	return 0
}

func logCurrentUser(ctx context.Context, logger *slog.Logger) {
	u, err := user.Current()
	if err != nil {
		logger.LogAttrs(ctx, slog.LevelWarn, "Unable to determine which user is running this exporter. More info: https://github.com/golang/go/issues/37348",
			slog.Any("err", err),
		)

		return
	}

	logger.LogAttrs(ctx, slog.LevelInfo, "Running as "+u.Username)

	if strings.Contains(u.Username, "ContainerAdministrator") || strings.Contains(u.Username, "ContainerUser") {
		logger.LogAttrs(ctx, slog.LevelWarn, "Running as a preconfigured Windows Container user. This may mean you do not have Windows HostProcess containers configured correctly and some functionality will not work as expected.")
	}
}

// setPriorityWindows sets the priority of the current process to the specified value.
func setPriorityWindows(ctx context.Context, logger *slog.Logger, pid int, priority string) error {
	// Mapping of priority names to uin32 values required by windows.SetPriorityClass.
	priorityStringToInt := map[string]uint32{
		"realtime":    windows.REALTIME_PRIORITY_CLASS,
		"high":        windows.HIGH_PRIORITY_CLASS,
		"abovenormal": windows.ABOVE_NORMAL_PRIORITY_CLASS,
		"normal":      windows.NORMAL_PRIORITY_CLASS,
		"belownormal": windows.BELOW_NORMAL_PRIORITY_CLASS,
		"low":         windows.IDLE_PRIORITY_CLASS,
	}

	winPriority, ok := priorityStringToInt[priority]
	if !ok {
		return fmt.Errorf("unknown process priority %q, must be one of %s", priority, strings.Join(slices.Sorted(maps.Keys(priorityStringToInt)), ", "))
	}

	// Only set process priority if a non-default value has been set
	if winPriority == windows.NORMAL_PRIORITY_CLASS {
		return nil
	}

	logger.LogAttrs(ctx, slog.LevelDebug, "setting process priority to "+priority)

	// https://learn.microsoft.com/en-us/windows/win32/procthread/process-security-and-access-rights
	handle, err := windows.OpenProcess(
		windows.STANDARD_RIGHTS_REQUIRED|windows.SYNCHRONIZE|windows.SPECIFIC_RIGHTS_ALL,
		false, uint32(pid),
	)
	if err != nil {
		return fmt.Errorf("failed to open own process: %w", err)
	}

	if err = windows.SetPriorityClass(handle, winPriority); err != nil {
		return fmt.Errorf("failed to set priority class: %w", err)
	}

	if err = windows.CloseHandle(handle); err != nil {
		return fmt.Errorf("failed to close handle: %w", err)
	}

	return nil
}

// closeCollection closes the collectors. It waits at most collectionCloseTimeout,
// so that a collector that doesn't return can't block the shutdown.
func closeCollection(ctx context.Context, logger *slog.Logger, collection *collector.Collection) {
	errCh := make(chan error, 1)

	//nolint:contextcheck // Close has no context parameter. The select below bounds the wait instead.
	go func() {
		errCh <- collection.Close()
	}()

	timer := time.NewTimer(collectionCloseTimeout)
	defer timer.Stop()

	select {
	case err := <-errCh:
		if err == nil {
			return
		}

		for _, err := range utils.SplitError(err) {
			logger.LogAttrs(ctx, slog.LevelWarn, "couldn't close collector",
				slog.Any("err", err),
			)
		}
	case <-timer.C:
		logger.LogAttrs(ctx, slog.LevelWarn, fmt.Sprintf("collectors didn't close within %s", collectionCloseTimeout))
	}
}

// effectiveCollectors returns the sorted list of enabled collectors without the disabled ones.
func effectiveCollectors(enabled, disabled []string) []string {
	return slices.DeleteFunc(slices.Compact(slices.Sorted(slices.Values(enabled))), func(name string) bool {
		return slices.Contains(disabled, name)
	})
}

// unknownCollectors returns the names that are not available collectors.
// Empty names, as in "cpu,,net", are skipped.
func unknownCollectors(names []string) []string {
	available := collector.Available()

	var unknown []string

	for _, name := range names {
		if name != "" && !slices.Contains(available, name) && !slices.Contains(unknown, name) {
			unknown = append(unknown, name)
		}
	}

	return unknown
}

// validateTimeoutMargin rejects a margin that would extend the scrape timeout
// beyond the timeout of the client.
func validateTimeoutMargin(margin float64) error {
	if math.IsNaN(margin) || math.IsInf(margin, 0) || margin < 0 {
		return fmt.Errorf("--scrape.timeout-margin must be a non-negative number of seconds, got %v", margin)
	}

	return nil
}

// newServeMux returns the router of the exporter. http.ServeMux panics on a
// pattern it cannot parse or that conflicts with another route, so newServeMux
// returns an error for a telemetry path that cannot be registered instead.
func newServeMux(metricsPath string, debugEnabled bool, metricsHandler http.Handler) (*http.ServeMux, error) {
	if !strings.HasPrefix(metricsPath, "/") {
		return nil, fmt.Errorf("--telemetry.path %q must start with /", metricsPath)
	}

	if metricsPath == healthPath {
		return nil, fmt.Errorf("--telemetry.path %q is reserved for the health endpoint", metricsPath)
	}

	mux := http.NewServeMux()
	mux.Handle("GET "+healthPath, httphandler.NewHealthHandler())

	if debugEnabled {
		mux.HandleFunc("GET /debug/pprof/", pprof.Index)
		mux.HandleFunc("GET /debug/pprof/cmdline", pprof.Cmdline)
		mux.HandleFunc("GET /debug/pprof/profile", pprof.Profile)
		mux.HandleFunc("GET /debug/pprof/symbol", pprof.Symbol)
		mux.HandleFunc("GET /debug/pprof/trace", pprof.Trace)
	}

	// The metrics route is registered last, so that a conflict with any other route is detected.
	if err := handle(mux, "GET "+metricsPath, metricsHandler); err != nil {
		return nil, fmt.Errorf("--telemetry.path %q cannot be registered: %w", metricsPath, err)
	}

	return mux, nil
}

// handle registers handler for pattern and returns the panic of
// http.ServeMux.Handle as an error.
func handle(mux *http.ServeMux, pattern string, handler http.Handler) error {
	var err error

	func() {
		defer func() {
			if r := recover(); r != nil {
				err = fmt.Errorf("%v", r)
			}
		}()

		mux.Handle(pattern, handler)
	}()

	return err
}

func expandEnabledCollectors(enabled string) []string {
	expanded := strings.ReplaceAll(enabled, "[defaults]", collector.DefaultCollectors)

	return slices.Compact(strings.Split(expanded, ","))
}

// scrapeToStdout performs a single scrape through the HTTP handler and writes
// the response body to stdout. It reuses the same handler that serves /metrics
// so the output is identical, including exporter self-instrumentation and
// timeout margin.
func scrapeToStdout(ctx context.Context, w io.Writer, logger *slog.Logger, collectors *collector.Collection, opts *httphandler.Options) int {
	handler := httphandler.New(logger, collectors, opts)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "/metrics", nil)
	if err != nil {
		logger.LogAttrs(ctx, slog.LevelError, "couldn't create request",
			slog.Any("err", err),
		)

		return 1
	}

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	result := rec.Result()
	defer result.Body.Close()

	if result.StatusCode != http.StatusOK {
		logger.LogAttrs(ctx, slog.LevelError, "scrape failed",
			slog.Int("status", result.StatusCode),
		)

		return 1
	}

	if _, err := io.Copy(w, result.Body); err != nil {
		logger.LogAttrs(ctx, slog.LevelError, "failed to write metrics to stdout",
			slog.Any("err", err),
		)

		return 1
	}

	return 0
}

// setProcessMemoryLimit applies the CLI contract: zero disables the soft limit.
func setProcessMemoryLimit(limit int64) error {
	if limit < 0 {
		return errors.New("process.memory-limit must be non-negative")
	}

	if limit == 0 {
		limit = math.MaxInt64
	}

	debug.SetMemoryLimit(limit)

	return nil
}
