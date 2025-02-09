// Copyright 2024 The Prometheus Authors
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
	"log/slog"
	"net/http"
	"net/http/pprof"
	"os"
	"os/signal"
	"os/user"
	"runtime"
	"runtime/debug"
	"slices"
	"strings"
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

func main() {
	exitCode := run()

	// If we are running as a service, we need to signal the service control manager that we are done.
	if !IsService {
		os.Exit(exitCode)
	}

	exitCodeCh <- exitCode

	// Wait for the service control manager to signal that we are done.
	<-serviceManagerFinishedCh
}

func run() int {
	startTime := time.Now()
	ctx := context.Background()

	app := kingpin.New("windows_exporter", "A metrics collector for Windows.")

	var (
		configFile = app.Flag(
			"config.file",
			"YAML configuration file to use. Values set in this file will be overridden by CLI flags.",
		).String()
		insecureSkipVerify = app.Flag(
			"config.file.insecure-skip-verify",
			"Skip TLS verification in loading YAML configuration.",
		).Default("false").Bool()
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
		timeoutMargin = app.Flag(
			"scrape.timeout-margin",
			"Seconds to subtract from the timeout allowed by the client. Tune to allow for overhead or high loads.",
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
			"Limit memory usage in bytes. This is a soft-limit and not guaranteed. 0 means no limit. Read more at https://pkg.go.dev/runtime/debug#SetMemoryLimit .",
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

	// Load values from configuration file(s). Executable flags must first be parsed, in order
	// to load the specified file(s).
	if _, err := app.Parse(os.Args[1:]); err != nil {
		//nolint:sloglint // we do not have an logger yet
		slog.Error("Failed to parse CLI args",
			slog.Any("err", err),
		)

		return 1
	}

	debug.SetMemoryLimit(*memoryLimit)

	logger, err := log.New(logConfig)
	if err != nil {
		//nolint:sloglint // we do not have an logger yet
		slog.Error("failed to create logger",
			slog.Any("err", err),
		)

		return 1
	}

	if *configFile != "" {
		resolver, err := config.NewResolver(ctx, *configFile, logger, *insecureSkipVerify)
		if err != nil {
			logger.Error("could not load config file",
				slog.Any("err", err),
			)

			return 1
		}

		if err = resolver.Bind(app, os.Args[1:]); err != nil {
			logger.ErrorContext(ctx, "failed to bind configuration",
				slog.Any("err", err),
			)

			return 1
		}

		// NOTE: This is temporary fix for issue #1092, calling kingpin.Parse
		// twice makes slices flags duplicate its value, this clean up
		// the first parse before the second call.
		*webConfig.WebListenAddresses = slices.Compact(*webConfig.WebListenAddresses)

		// Parse flags once more to include those discovered in configuration file(s).
		if _, err = app.Parse(os.Args[1:]); err != nil {
			logger.ErrorContext(ctx, "failed to parse CLI args from YAML file",
				slog.Any("err", err),
			)

			return 1
		}

		logger, err = log.New(logConfig)
		if err != nil {
			//nolint:sloglint // we do not have an logger yet
			slog.Error("failed to create logger",
				slog.Any("err", err),
			)

			return 1
		}
	}

	logger.LogAttrs(ctx, slog.LevelDebug, "logging has Started")

	if err = setPriorityWindows(logger, os.Getpid(), *processPriority); err != nil {
		logger.Error("failed to set process priority",
			slog.Any("err", err),
		)

		return 1
	}

	enabledCollectorList := expandEnabledCollectors(*enabledCollectors)
	if err := collectors.Enable(enabledCollectorList); err != nil {
		logger.Error("couldn't enable collectors",
			slog.Any("err", err),
		)

		return 1
	}

	// Initialize collectors before loading
	if err = collectors.Build(logger); err != nil {
		for _, err := range utils.SplitError(err) {
			logger.Error("couldn't initialize collector",
				slog.Any("err", err),
			)

			return 1
		}
	}

	logCurrentUser(logger)

	logger.InfoContext(ctx, "Enabled collectors: "+strings.Join(enabledCollectorList, ", "))

	mux := http.NewServeMux()
	mux.Handle("GET /health", httphandler.NewHealthHandler())
	mux.Handle("GET /version", httphandler.NewVersionHandler())
	mux.Handle("GET "+*metricsPath, httphandler.New(logger, collectors, &httphandler.Options{
		DisableExporterMetrics: *disableExporterMetrics,
		TimeoutMargin:          *timeoutMargin,
	}))

	if *debugEnabled {
		mux.HandleFunc("GET /debug/pprof/", pprof.Index)
		mux.HandleFunc("GET /debug/pprof/cmdline", pprof.Cmdline)
		mux.HandleFunc("GET /debug/pprof/profile", pprof.Profile)
		mux.HandleFunc("GET /debug/pprof/symbol", pprof.Symbol)
		mux.HandleFunc("GET /debug/pprof/trace", pprof.Trace)
	}

	logger.LogAttrs(ctx, slog.LevelInfo, fmt.Sprintf("starting windows_exporter in %s", time.Since(startTime)),
		slog.String("version", version.Version),
		slog.String("branch", version.Branch),
		slog.String("revision", version.GetRevision()),
		slog.String("goversion", version.GoVersion),
		slog.String("builddate", version.BuildDate),
		slog.Int("maxprocs", runtime.GOMAXPROCS(0)),
	)

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

	ctx, stop := signal.NotifyContext(ctx, os.Interrupt, os.Kill)
	defer stop()

	select {
	case <-ctx.Done():
		logger.Info("Shutting down windows_exporter via kill signal")
	case <-stopCh:
		logger.Info("Shutting down windows_exporter via service control")
	case err := <-errCh:
		if err != nil {
			logger.ErrorContext(ctx, "Failed to start windows_exporter",
				slog.Any("err", err),
			)

			return 1
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_ = server.Shutdown(ctx)

	logger.InfoContext(ctx, "windows_exporter has shut down")

	return 0
}

func logCurrentUser(logger *slog.Logger) {
	u, err := user.Current()
	if err != nil {
		logger.Warn("Unable to determine which user is running this exporter. More info: https://github.com/golang/go/issues/37348",
			slog.Any("err", err),
		)

		return
	}

	logger.Info("Running as " + u.Username)

	if strings.Contains(u.Username, "ContainerAdministrator") || strings.Contains(u.Username, "ContainerUser") {
		logger.Warn("Running as a preconfigured Windows Container user. This may mean you do not have Windows HostProcess containers configured correctly and some functionality will not work as expected.")
	}
}

// setPriorityWindows sets the priority of the current process to the specified value.
func setPriorityWindows(logger *slog.Logger, pid int, priority string) error {
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

	// Only set process priority if a non-default and valid value has been set
	if !ok || winPriority == windows.NORMAL_PRIORITY_CLASS {
		return nil
	}

	logger.LogAttrs(context.Background(), slog.LevelDebug, "setting process priority to "+priority)

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

func expandEnabledCollectors(enabled string) []string {
	expanded := strings.ReplaceAll(enabled, "[defaults]", collector.DefaultCollectors)

	return slices.Compact(strings.Split(expanded, ","))
}
