//go:build windows

package netframework

import (
	"errors"
	"fmt"
	"log/slog"
	"slices"

	"github.com/alecthomas/kingpin/v2"
	"github.com/prometheus-community/windows_exporter/pkg/types"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/yusufpapurcu/wmi"
)

const Name = "netframework"

type Config struct {
	CollectorsEnabled []string `yaml:"collectors_enabled"`
}

var ConfigDefaults = Config{
	CollectorsEnabled: []string{
		collectorClrExceptions,
		collectorClrInterop,
		collectorClrJIT,
		collectorClrLoading,
		collectorClrLocksAndThreads,
		collectorClrMemory,
		collectorClrRemoting,
		collectorClrSecurity,
	},
}

const (
	collectorClrExceptions      = "clrexceptions"
	collectorClrInterop         = "clrinterop"
	collectorClrJIT             = "clrjit"
	collectorClrLoading         = "clrloading"
	collectorClrLocksAndThreads = "clrlocksandthreads"
	collectorClrMemory          = "clrmemory"
	collectorClrRemoting        = "clrremoting"
	collectorClrSecurity        = "clrsecurity"
)

// A Collector is a Prometheus Collector for WMI Win32_PerfRawData_NETFramework_NETCLRExceptions metrics.
type Collector struct {
	config    Config
	wmiClient *wmi.Client

	// clrexceptions
	numberOfExceptionsThrown *prometheus.Desc
	numberOfFilters          *prometheus.Desc
	numberOfFinally          *prometheus.Desc
	throwToCatchDepth        *prometheus.Desc

	// clrinterop
	numberOfCCWs        *prometheus.Desc
	numberOfMarshalling *prometheus.Desc
	numberOfStubs       *prometheus.Desc

	// clrjit
	numberOfMethodsJitted      *prometheus.Desc
	timeInJit                  *prometheus.Desc
	standardJitFailures        *prometheus.Desc
	totalNumberOfILBytesJitted *prometheus.Desc

	// clrloading
	bytesInLoaderHeap         *prometheus.Desc
	currentAppDomains         *prometheus.Desc
	currentAssemblies         *prometheus.Desc
	currentClassesLoaded      *prometheus.Desc
	totalAppDomains           *prometheus.Desc
	totalAppDomainsUnloaded   *prometheus.Desc
	totalAssemblies           *prometheus.Desc
	totalClassesLoaded        *prometheus.Desc
	totalNumberOfLoadFailures *prometheus.Desc

	// clrlocksandthreads
	currentQueueLength               *prometheus.Desc
	numberOfCurrentLogicalThreads    *prometheus.Desc
	numberOfCurrentPhysicalThreads   *prometheus.Desc
	numberOfCurrentRecognizedThreads *prometheus.Desc
	numberOfTotalRecognizedThreads   *prometheus.Desc
	queueLengthPeak                  *prometheus.Desc
	totalNumberOfContentions         *prometheus.Desc

	// clrmemory
	allocatedBytes            *prometheus.Desc
	finalizationSurvivors     *prometheus.Desc
	heapSize                  *prometheus.Desc
	promotedBytes             *prometheus.Desc
	numberGCHandles           *prometheus.Desc
	numberCollections         *prometheus.Desc
	numberInducedGC           *prometheus.Desc
	numberOfPinnedObjects     *prometheus.Desc
	numberOfSinkBlocksInUse   *prometheus.Desc
	numberTotalCommittedBytes *prometheus.Desc
	numberTotalReservedBytes  *prometheus.Desc
	timeInGC                  *prometheus.Desc

	// clrremoting
	channels                  *prometheus.Desc
	contextBoundClassesLoaded *prometheus.Desc
	contextBoundObjects       *prometheus.Desc
	contextProxies            *prometheus.Desc
	contexts                  *prometheus.Desc
	totalRemoteCalls          *prometheus.Desc

	// clrsecurity
	numberLinkTimeChecks *prometheus.Desc
	timeInRTChecks       *prometheus.Desc
	stackWalkDepth       *prometheus.Desc
	totalRuntimeChecks   *prometheus.Desc
}

func New(config *Config) *Collector {
	if config == nil {
		config = &ConfigDefaults
	}

	c := &Collector{
		config: *config,
	}

	return c
}

func NewWithFlags(_ *kingpin.Application) *Collector {
	return &Collector{}
}

func (c *Collector) GetName() string {
	return Name
}

func (c *Collector) GetPerfCounter(_ *slog.Logger) ([]string, error) {
	return []string{}, nil
}

func (c *Collector) Close(_ *slog.Logger) error {
	return nil
}

func (c *Collector) Build(_ *slog.Logger, wmiClient *wmi.Client) error {
	if wmiClient == nil || wmiClient.SWbemServicesClient == nil {
		return errors.New("wmiClient or SWbemServicesClient is nil")
	}

	c.wmiClient = wmiClient

	if slices.Contains(c.config.CollectorsEnabled, collectorClrExceptions) {
		c.buildClrExceptions()
	}

	if slices.Contains(c.config.CollectorsEnabled, collectorClrInterop) {
		c.buildClrInterop()
	}

	if slices.Contains(c.config.CollectorsEnabled, collectorClrJIT) {
		c.buildClrJIT()
	}

	if slices.Contains(c.config.CollectorsEnabled, collectorClrLoading) {
		c.buildClrLoading()
	}

	if slices.Contains(c.config.CollectorsEnabled, collectorClrLocksAndThreads) {
		c.buildClrLocksAndThreads()
	}

	if slices.Contains(c.config.CollectorsEnabled, collectorClrMemory) {
		c.buildClrMemory()
	}

	if slices.Contains(c.config.CollectorsEnabled, collectorClrRemoting) {
		c.buildClrRemoting()
	}

	if slices.Contains(c.config.CollectorsEnabled, collectorClrSecurity) {
		c.buildClrSecurity()
	}

	return nil
}

// Collect sends the metric values for each metric
// to the provided prometheus Metric channel.
func (c *Collector) Collect(_ *types.ScrapeContext, logger *slog.Logger, ch chan<- prometheus.Metric) error {
	logger = logger.With(slog.String("collector", Name))

	var (
		err  error
		errs []error
	)

	if slices.Contains(c.config.CollectorsEnabled, collectorClrExceptions) {
		if err = c.collectClrExceptions(ch); err != nil {
			errs = append(errs, fmt.Errorf("failed to collect %s metrics: %w", collectorClrExceptions, err))
		}
	}

	if slices.Contains(c.config.CollectorsEnabled, collectorClrInterop) {
		if err = c.collectClrInterop(ch); err != nil {
			errs = append(errs, fmt.Errorf("failed to collect %s metrics: %w", collectorClrInterop, err))
		}
	}

	if slices.Contains(c.config.CollectorsEnabled, collectorClrJIT) {
		if err = c.collectClrJIT(ch); err != nil {
			errs = append(errs, fmt.Errorf("failed to collect %s metrics: %w", collectorClrJIT, err))
		}
	}

	if slices.Contains(c.config.CollectorsEnabled, collectorClrLoading) {
		if err = c.collectClrLoading(ch); err != nil {
			errs = append(errs, fmt.Errorf("failed to collect %s metrics: %w", collectorClrLoading, err))
		}
	}

	if slices.Contains(c.config.CollectorsEnabled, collectorClrLocksAndThreads) {
		if err = c.collectClrLocksAndThreads(ch); err != nil {
			errs = append(errs, fmt.Errorf("failed to collect %s metrics: %w", collectorClrLocksAndThreads, err))
		}
	}

	if slices.Contains(c.config.CollectorsEnabled, collectorClrMemory) {
		if err = c.collectClrMemory(ch); err != nil {
			errs = append(errs, fmt.Errorf("failed to collect %s metrics: %w", collectorClrMemory, err))
		}
	}

	if slices.Contains(c.config.CollectorsEnabled, collectorClrRemoting) {
		if err = c.collectClrRemoting(ch); err != nil {
			errs = append(errs, fmt.Errorf("failed to collect %s metrics: %w", collectorClrRemoting, err))
		}
	}

	if slices.Contains(c.config.CollectorsEnabled, collectorClrSecurity) {
		if err = c.collectClrSecurity(ch); err != nil {
			errs = append(errs, fmt.Errorf("failed to collect %s metrics: %w", collectorClrSecurity, err))
		}
	}

	return errors.Join(errs...)
}
