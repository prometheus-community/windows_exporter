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

package netframework

import (
	"errors"
	"fmt"
	"log/slog"
	"maps"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/alecthomas/kingpin/v2"
	"github.com/prometheus-community/windows_exporter/internal/labelcache"
	"github.com/prometheus-community/windows_exporter/internal/mi"
	"github.com/prometheus-community/windows_exporter/internal/pdh"
	"github.com/prometheus-community/windows_exporter/internal/scrapecache"
	"github.com/prometheus-community/windows_exporter/internal/types"
	"github.com/prometheus-community/windows_exporter/internal/utils"
	"github.com/prometheus-community/windows_exporter/internal/utils/recovery"
	"github.com/prometheus/client_golang/prometheus"
)

type perfCollector[T any] interface {
	Collect(dst *[]T) error
	Close()
}

const Name = "netframework"

type Config struct {
	CollectorsEnabled []string       `yaml:"enabled"`
	ProcessInclude    *regexp.Regexp `yaml:"process-include"`
	ProcessExclude    *regexp.Regexp `yaml:"process-exclude"`
}

//nolint:gochecknoglobals
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
	ProcessInclude: types.RegExpAny,
	ProcessExclude: types.RegExpEmpty,
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

// A Collector collects .NET Framework performance counters.
type Collector struct {
	// labels keeps the label pairs of the series of the last scrape. The
	// counters change on every scrape, but their process labels stay the
	// same while the process runs.
	labels   labelcache.Cache
	config   Config
	logger   *slog.Logger
	closeFns []func()
	// perfQuery samples the counters of all enabled sub-collectors at once.
	// The .NET CLR counter provider serves one query at a time and walks
	// every .NET process for each, so separate queries wait for each other.
	perfQuery              *pdh.Query
	perfClrExceptions      perfCollector[perfDataClrExceptions]
	perfClrInterop         perfCollector[perfDataClrInterop]
	perfClrJIT             perfCollector[perfDataClrJIT]
	perfClrLoading         perfCollector[perfDataClrLoading]
	perfClrLocksAndThreads perfCollector[perfDataClrLocksAndThreads]
	perfClrMemory          perfCollector[perfDataClrMemory]
	perfClrRemoting        perfCollector[perfDataClrRemoting]
	perfClrSecurity        perfCollector[perfDataClrSecurity]
	// perfClrProcessID reads the Process ID counter of the .NET CLR Memory
	// object for the sub-collectors of the other objects, see processIDs.
	perfClrProcessID perfCollector[perfDataClrProcessID]
	processIDRows    []perfDataClrProcessID
	// processIDs maps the instances of the current sample to their process
	// ID. Collect fills it before it starts the sub-collectors, which only
	// read it.
	processIDs map[processKey]string

	collectorFns []func(ch chan<- prometheus.Metric, scrape *labelcache.Scrape, maxScrapeDuration time.Duration) error

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

	if config.CollectorsEnabled == nil {
		config.CollectorsEnabled = ConfigDefaults.CollectorsEnabled
	}

	if config.ProcessInclude == nil {
		config.ProcessInclude = ConfigDefaults.ProcessInclude
	}

	if config.ProcessExclude == nil {
		config.ProcessExclude = ConfigDefaults.ProcessExclude
	}

	c := &Collector{
		config: *config,
	}

	return c
}

func NewWithFlags(app *kingpin.Application) *Collector {
	c := &Collector{
		config: ConfigDefaults,
	}
	c.config.CollectorsEnabled = make([]string, 0)

	var collectorsEnabled string

	app.Flag(
		"collector.netframework.enabled",
		"Comma-separated list of collectors to use. Defaults to all, if not specified.",
	).Default(strings.Join(ConfigDefaults.CollectorsEnabled, ",")).StringVar(&collectorsEnabled)

	utils.RegexpFlag(app,
		"collector.netframework.process-exclude",
		"Regexp of processes to exclude. Process name must both match include and not match exclude to be included.",
		"", &c.config.ProcessExclude,
	)

	utils.RegexpFlag(app,
		"collector.netframework.process-include",
		"Regexp of processes to include. Process name must both match include and not match exclude to be included.",
		".+", &c.config.ProcessInclude,
	)

	app.Action(func(*kingpin.ParseContext) error {
		c.config.CollectorsEnabled = strings.Split(collectorsEnabled, ",")

		return nil
	})

	return c
}

func (c *Collector) GetName() string {
	return Name
}

// ScrapeCaches returns the caches that the collection resets and begins a scrape of around Collect.
func (c *Collector) ScrapeCaches() []scrapecache.Cache {
	return []scrapecache.Cache{&c.labels}
}

func (c *Collector) Close() error {
	for _, closeFn := range c.closeFns {
		closeFn()
	}

	c.closeFns = nil
	c.collectorFns = nil

	if c.perfClrProcessID != nil {
		c.perfClrProcessID.Close()
		c.perfClrProcessID = nil
	}

	c.processIDRows = nil
	c.processIDs = nil

	c.perfQuery.Close()
	c.perfQuery = nil

	return nil
}

func (c *Collector) Build(logger *slog.Logger, _ *mi.Session) error {
	if len(c.config.CollectorsEnabled) == 0 {
		return nil
	}

	c.logger = logger

	var err error

	c.perfQuery, err = pdh.NewQuery()
	if err != nil {
		return fmt.Errorf("failed to open .NET CLR performance counter query: %w", err)
	}

	c.collectorFns = make([]func(ch chan<- prometheus.Metric, scrape *labelcache.Scrape, maxScrapeDuration time.Duration) error, 0, len(c.config.CollectorsEnabled))

	subCollectors := map[string]struct {
		build   func() error
		collect func(ch chan<- prometheus.Metric, scrape *labelcache.Scrape, maxScrapeDuration time.Duration) error
		close   func()
	}{
		collectorClrExceptions: {
			build:   c.buildClrExceptions,
			collect: c.collectClrExceptions,
			close: func() {
				if c.perfClrExceptions != nil {
					c.perfClrExceptions.Close()
				}
			},
		},
		collectorClrJIT: {
			build:   c.buildClrJIT,
			collect: c.collectClrJIT,
			close: func() {
				if c.perfClrJIT != nil {
					c.perfClrJIT.Close()
				}
			},
		},
		collectorClrLoading: {
			build:   c.buildClrLoading,
			collect: c.collectClrLoading,
			close: func() {
				if c.perfClrLoading != nil {
					c.perfClrLoading.Close()
				}
			},
		},
		collectorClrInterop: {
			build:   c.buildClrInterop,
			collect: c.collectClrInterop,
			close: func() {
				if c.perfClrInterop != nil {
					c.perfClrInterop.Close()
				}
			},
		},
		collectorClrLocksAndThreads: {
			build:   c.buildClrLocksAndThreads,
			collect: c.collectClrLocksAndThreads,
			close: func() {
				if c.perfClrLocksAndThreads != nil {
					c.perfClrLocksAndThreads.Close()
				}
			},
		},
		collectorClrMemory: {
			build:   c.buildClrMemory,
			collect: c.collectClrMemory,
			close: func() {
				if c.perfClrMemory != nil {
					c.perfClrMemory.Close()
				}
			},
		},
		collectorClrRemoting: {
			build:   c.buildClrRemoting,
			collect: c.collectClrRemoting,
			close: func() {
				if c.perfClrRemoting != nil {
					c.perfClrRemoting.Close()
				}
			},
		},
		collectorClrSecurity: {
			build:   c.buildClrSecurity,
			collect: c.collectClrSecurity,
			close: func() {
				if c.perfClrSecurity != nil {
					c.perfClrSecurity.Close()
				}
			},
		},
	}

	// Sort a copy, to not modify the slice of the caller or ConfigDefaults.
	// Result must order, to prevent test failures.
	collectorsEnabled := slices.Compact(slices.Sorted(slices.Values(c.config.CollectorsEnabled)))

	for _, name := range collectorsEnabled {
		if _, ok := subCollectors[name]; !ok {
			return fmt.Errorf("unknown sub collector: %s. Possible values: %s", name,
				strings.Join(slices.Sorted(maps.Keys(subCollectors)), ", "),
			)
		}
	}

	for _, name := range collectorsEnabled {
		c.closeFns = append(c.closeFns, subCollectors[name].close)
		if err := subCollectors[name].build(); err != nil {
			_ = c.Close()

			return fmt.Errorf("failed to build %s: %w", name, err)
		}

		c.collectorFns = append(c.collectorFns, subCollectors[name].collect)
	}

	// Only the .NET CLR Memory object has a Process ID counter. The
	// clrmemory rows carry their own, the other sub-collectors look it up.
	// With clrmemory enabled, Process ID is read twice on purpose: the
	// sub-collectors run concurrently, so they cannot share the clrmemory rows.
	if slices.ContainsFunc(collectorsEnabled, func(name string) bool { return name != collectorClrMemory }) {
		c.perfClrProcessID, err = newPerfCollector[perfDataClrProcessID](c.logger, c.perfQuery, ".NET CLR Memory")
		if err != nil {
			_ = c.Close()

			return fmt.Errorf("failed to build the process ID lookup: %w", err)
		}
	}

	return nil
}

// Collect sends the metric values for each metric
// to the provided prometheus Metric channel.
func (c *Collector) Collect(ch chan<- prometheus.Metric, maxScrapeDuration time.Duration) error {
	// The sub-collectors read this sample. Without any .NET process there is
	// no data; the sub-collectors report that themselves.
	if c.perfQuery != nil {
		if err := c.perfQuery.Collect(); err != nil && !errors.Is(err, pdh.ErrNoData) {
			return fmt.Errorf("failed to collect .NET CLR performance counters: %w", err)
		}
	}

	// The sub-collectors publish their metrics even if the process IDs are
	// unknown, with an empty process_id label.
	processIDErr := c.collectProcessIDs()

	scrape := c.labels.Scrape()

	var g recovery.Group

	for _, fn := range c.collectorFns {
		g.Go(func() error {
			return fn(ch, scrape, maxScrapeDuration)
		})
	}

	return errors.Join(processIDErr, g.Wait())
}

// processKey identifies an instance of a sample by the name Windows reported
// and its occurrence, before suffixDuplicateProcesses changes the name.
type processKey struct {
	name       string
	occurrence int
}

type perfDataClrProcessID struct {
	Name       string
	Occurrence int
	ProcessID  float64 `perfdata:"Process ID"`
}

// collectProcessIDs maps the instances of the .NET CLR Memory object of the
// current sample to their process IDs. The join assumes, without
// documentation to back it, that the n-th instance with a name is the same
// process in every .NET CLR object of a sample: PdhCollectQueryData does one
// RegQueryValueEx(HKEY_PERFORMANCE_DATA) per collect for the whole query,
// and the CLR provider writes every object from the same instance list.
// Sampling the objects in separate queries would break the join for
// duplicate names. An instance without a valid Process ID has no
// entry. Without any .NET process, the map is empty and the sub-collectors
// report that there is no data.
func (c *Collector) collectProcessIDs() error {
	if c.processIDs == nil {
		c.processIDs = make(map[processKey]string)
	} else {
		clear(c.processIDs)
	}

	if c.perfClrProcessID == nil {
		return nil
	}

	if err := c.perfClrProcessID.Collect(&c.processIDRows); err != nil {
		if errors.Is(err, pdh.ErrNoData) {
			return nil
		}

		return fmt.Errorf("failed to collect the process IDs of .NET CLR Memory: %w", err)
	}

	for _, row := range c.processIDRows {
		if row.Name == "_Global_" {
			continue
		}

		c.processIDs[processKey{name: row.Name, occurrence: row.Occurrence}] = strconv.FormatUint(uint64(row.ProcessID), 10)
	}

	return nil
}

// isProcessIncluded reports whether name, the process name as Windows reports
// it without an occurrence suffix, passes the process filters.
func (c *Collector) isProcessIncluded(name string) bool {
	if c.config.ProcessInclude != nil && !c.config.ProcessInclude.MatchString(name) {
		return false
	}

	return c.config.ProcessExclude == nil || !c.config.ProcessExclude.MatchString(name)
}

// clrInstance is implemented by the rows of the .NET CLR objects.
type clrInstance interface {
	instance() (name *string, occurrence int)
}

// filterProcesses removes the _Global_ instance and the processes that do not
// pass the process filters from rows. The filters match the name that Windows
// reports, so they apply to every process with that name, like the filters
// of the process collector. Filtering before suffixDuplicateProcesses keeps
// the suffixes the same in every object of a sample.
func filterProcesses[T any, P interface {
	*T
	clrInstance
}](c *Collector, rows []T) []T {
	filter := !utils.IsMatchAll(c.config.ProcessInclude) || !utils.IsMatchNone(c.config.ProcessExclude)

	return slices.DeleteFunc(rows, func(row T) bool {
		name, _ := P(&row).instance()

		return *name == "_Global_" || (filter && !c.isProcessIncluded(*name))
	})
}

// prepareProcesses filters rows, looks up the process ID of every row and
// then adds occurrence suffixes to duplicate process names. The process ID
// of rows[i] is processIDs[i]; it is empty if the .NET CLR Memory object has
// no valid Process ID for the instance, e.g. if clrmemory is disabled and
// the object is unavailable. The process label still keeps the series apart.
func prepareProcesses[T any, P interface {
	*T
	clrInstance
}](c *Collector, rows []T) ([]T, []string) {
	rows = filterProcesses[T, P](c, rows)
	processIDs := make([]string, len(rows))

	for i := range rows {
		name, occurrence := P(&rows[i]).instance()
		processIDs[i] = c.processIDs[processKey{name: *name, occurrence: occurrence}]
	}

	suffixDuplicateProcesses[T, P](rows)

	return rows, processIDs
}

// suffixDuplicateProcesses keeps processes with the same name apart. The .NET
// CLR objects name instances after the process image and report a name once
// per process, e.g. "powershell" twice for two PowerShell processes. Later
// processes get the suffix "#1", "#2" and so on, which depends on the order
// of the sample and does not identify a process across scrapes.
func suffixDuplicateProcesses[T any, P interface {
	*T
	clrInstance
}](rows []T) {
	pdh.SuffixDuplicates(rows,
		func(row *T) *string {
			name, _ := P(row).instance()

			return name
		},
		func(row *T) int {
			_, occurrence := P(row).instance()

			return occurrence
		},
	)
}

// newPerfCollector adds the counters of object to query. On error,
// pdh.NewQueryCollector has already removed the counters it added. Missing
// objects are optional; missing individual counters remain errors.
func newPerfCollector[T any](logger *slog.Logger, query *pdh.Query, object string) (perfCollector[T], error) {
	c, err := pdh.NewQueryCollector[T](logger, query, pdh.CounterTypeRaw, object, pdh.InstancesAll)
	if err == nil {
		return c, nil
	}

	if onlyMissingObject(err) {
		return unavailablePerfCollector[T]{}, nil
	}

	return nil, err
}

func onlyMissingObject(err error) bool {
	if joined, ok := err.(interface{ Unwrap() []error }); ok {
		children := joined.Unwrap()
		if len(children) == 0 {
			return false
		}

		for _, child := range children {
			if !onlyMissingObject(child) {
				return false
			}
		}

		return true
	}

	if wrapped, ok := err.(interface{ Unwrap() error }); ok {
		return onlyMissingObject(wrapped.Unwrap())
	}

	return errors.Is(err, pdh.NewPdhError(pdh.CstatusNoObject))
}

type unavailablePerfCollector[T any] struct{}

func (unavailablePerfCollector[T]) Collect(dst *[]T) error {
	*dst = nil

	return pdh.ErrNoData
}

func (unavailablePerfCollector[T]) Close() {}
