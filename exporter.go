//go:build windows

//go:generate go run github.com/tc-hib/go-winres@v0.3.3 make --product-version=git-tag --file-version=git-tag --arch=amd64,arm64

package main

//goland:noinspection GoUnsortedImport
//nolint:gofumpt
import (
	// Its important that we do these first so that we can register with the Windows service control ASAP to avoid timeouts.
	"github.com/prometheus-community/windows_exporter/pkg/initiate"

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
	"sort"
	"strings"
	"time"

	"github.com/alecthomas/kingpin/v2"
	"github.com/prometheus-community/windows_exporter/pkg/collector"
	"github.com/prometheus-community/windows_exporter/pkg/config"
	"github.com/prometheus-community/windows_exporter/pkg/httphandler"
	winlog "github.com/prometheus-community/windows_exporter/pkg/log"
	"github.com/prometheus-community/windows_exporter/pkg/log/flag"
	"github.com/prometheus-community/windows_exporter/pkg/types"
	"github.com/prometheus-community/windows_exporter/pkg/utils"
	"github.com/prometheus/common/version"
	"github.com/prometheus/exporter-toolkit/web"
	webflag "github.com/prometheus/exporter-toolkit/web/kingpinflag"
	"golang.org/x/sys/windows"
)

// Mapping of priority names to uin32 values required by windows.SetPriorityClass.
var priorityStringToInt = map[string]uint32{
	"realtime":    windows.REALTIME_PRIORITY_CLASS,
	"high":        windows.HIGH_PRIORITY_CLASS,
	"abovenormal": windows.ABOVE_NORMAL_PRIORITY_CLASS,
	"normal":      windows.NORMAL_PRIORITY_CLASS,
	"belownormal": windows.BELOW_NORMAL_PRIORITY_CLASS,
	"low":         windows.IDLE_PRIORITY_CLASS,
}

func main() {
	os.Exit(run())
}

func run() int {
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
		maxRequests = app.Flag(
			"telemetry.max-requests",
			"Maximum number of concurrent requests. 0 to disable.",
		).Default("5").Int()
		enabledCollectors = app.Flag(
			"collectors.enabled",
			"Comma-separated list of collectors to use. Use '[defaults]' as a placeholder for all the collectors enabled by default.").
			Default(types.DefaultCollectors).String()
		printCollectors = app.Flag(
			"collectors.print",
			"If true, print available collectors and exit.",
		).Bool()
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
	)

	winlogConfig := &winlog.Config{}
	flag.AddFlags(app, winlogConfig)

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

	logger, err := winlog.New(winlogConfig)
	if err != nil {
		//nolint:sloglint // we do not have an logger yet
		slog.Error("failed to create logger",
			slog.Any("err", err),
		)

		return 1
	}

	if *configFile != "" {
		resolver, err := config.NewResolver(*configFile, logger, *insecureSkipVerify)
		if err != nil {
			logger.Error("could not load config file",
				slog.Any("err", err),
			)

			return 1
		}

		if err = resolver.Bind(app, os.Args[1:]); err != nil {
			logger.Error("Failed to bind configuration",
				slog.Any("err", err),
			)

			return 1
		}

		// NOTE: This is temporary fix for issue #1092, calling kingpin.Parse
		// twice makes slices flags duplicate its value, this clean up
		// the first parse before the second call.
		*webConfig.WebListenAddresses = (*webConfig.WebListenAddresses)[1:]

		// Parse flags once more to include those discovered in configuration file(s).
		if _, err = app.Parse(os.Args[1:]); err != nil {
			logger.Error("Failed to parse CLI args from YAML file",
				slog.Any("err", err),
			)

			return 1
		}

		logger, err = winlog.New(winlogConfig)
		if err != nil {
			//nolint:sloglint // we do not have an logger yet
			slog.Error("failed to create logger",
				slog.Any("err", err),
			)

			return 1
		}
	}

	logger.Debug("Logging has Started")

	if *printCollectors {
		printCollectorsToStdout()

		return 0
	}

	// Only set process priority if a non-default and valid value has been set
	if priority, ok := priorityStringToInt[*processPriority]; ok && priority != windows.NORMAL_PRIORITY_CLASS {
		logger.Debug("setting process priority to " + *processPriority)

		if err = setPriorityWindows(os.Getpid(), priority); err != nil {
			logger.Error("failed to set process priority",
				slog.Any("err", err),
			)

			return 1
		}
	}

	enabledCollectorList := utils.ExpandEnabledCollectors(*enabledCollectors)
	collectors.Enable(enabledCollectorList)

	// Initialize collectors before loading
	if err = collectors.Build(logger); err != nil {
		logger.Error("Couldn't load collectors",
			slog.Any("err", err),
		)

		return 1
	}

	if err = collectors.SetPerfCounterQuery(logger); err != nil {
		logger.Error("Couldn't set performance counter query",
			slog.Any("err", err),
		)

		return 1
	}

	logCurrentUser(logger)

	logger.Info("Enabled collectors: " + strings.Join(enabledCollectorList, ", "))

	if utils.PDHEnabled() {
		logger.Info("Using performance data helper from PHD.dll for performance counter collection. This is in experimental state.")
	}

	mux := http.NewServeMux()
	mux.Handle("GET /health", httphandler.NewHealthHandler())
	mux.Handle("GET /version", httphandler.NewVersionHandler())
	mux.Handle("GET "+*metricsPath, httphandler.New(logger, collectors, &httphandler.Options{
		DisableExporterMetrics: *disableExporterMetrics,
		TimeoutMargin:          *timeoutMargin,
		MaxRequests:            *maxRequests,
	}))

	if *debugEnabled {
		mux.HandleFunc("GET /debug/pprof/", pprof.Index)
		mux.HandleFunc("GET /debug/pprof/cmdline", pprof.Cmdline)
		mux.HandleFunc("GET /debug/pprof/profile", pprof.Profile)
		mux.HandleFunc("GET /debug/pprof/symbol", pprof.Symbol)
		mux.HandleFunc("GET /debug/pprof/trace", pprof.Trace)
	}

	logger.Info("Starting windows_exporter",
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

		errCh <- nil
	}()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	defer stop()

	select {
	case <-ctx.Done():
		logger.Info("Shutting down windows_exporter via kill signal")
	case <-initiate.StopCh:
		logger.Info("Shutting down windows_exporter via service control")
	case err := <-errCh:
		if err != nil {
			logger.Error("Failed to start windows_exporter",
				slog.Any("err", err),
			)

			return 1
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_ = server.Shutdown(ctx)

	logger.Info("windows_exporter has shut down")

	return 0
}

func printCollectorsToStdout() {
	collectorNames := collector.Available()
	sort.Strings(collectorNames)

	fmt.Println("Available collectors:") //nolint:forbidigo

	for _, n := range collectorNames {
		fmt.Printf(" - %s\n", n) //nolint:forbidigo
	}
}

func logCurrentUser(logger *slog.Logger) {
	if u, err := user.Current(); err == nil {
		logger.Info("Running as " + u.Username)

		if strings.Contains(u.Username, "ContainerAdministrator") || strings.Contains(u.Username, "ContainerUser") {
			logger.Warn("Running as a preconfigured Windows Container user. This may mean you do not have Windows HostProcess containers configured correctly and some functionality will not work as expected.")
		}

		return
	}

	logger.Warn("Unable to determine which user is running this exporter. More info: https://github.com/golang/go/issues/37348")
}

func setPriorityWindows(pid int, priority uint32) error {
	// https://learn.microsoft.com/en-us/windows/win32/procthread/process-security-and-access-rights
	handle, err := windows.OpenProcess(
		windows.STANDARD_RIGHTS_REQUIRED|windows.SYNCHRONIZE|windows.SPECIFIC_RIGHTS_ALL,
		false, uint32(pid),
	)
	if err != nil {
		return fmt.Errorf("failed to open own process: %w", err)
	}

	if err = windows.SetPriorityClass(handle, priority); err != nil {
		return fmt.Errorf("failed to set priority class: %w", err)
	}

	if err = windows.CloseHandle(handle); err != nil {
		return fmt.Errorf("failed to close handle: %w", err)
	}

	return nil
}
