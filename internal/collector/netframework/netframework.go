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
	"slices"
	"strings"
	"time"
	"unsafe"

	"github.com/alecthomas/kingpin/v2"
	"github.com/prometheus-community/windows_exporter/internal/mi"
	"github.com/prometheus-community/windows_exporter/internal/pdh"
	"github.com/prometheus-community/windows_exporter/internal/utils/recovery"
	"github.com/prometheus/client_golang/prometheus"
	"golang.org/x/sys/windows"
)

type perfCollector[T any] interface {
	Collect(dst *[]T) error
	Close()
}

const Name = "netframework"

type Config struct {
	CollectorsEnabled []string `yaml:"enabled"`
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
	config                 Config
	logger                 *slog.Logger
	perfFrequency          float64
	closeFns               []func()
	perfClrExceptions      perfCollector[perfDataClrExceptions]
	perfClrInterop         perfCollector[perfDataClrInterop]
	perfClrJIT             perfCollector[perfDataClrJIT]
	perfClrLoading         perfCollector[perfDataClrLoading]
	perfClrLocksAndThreads perfCollector[perfDataClrLocksAndThreads]
	perfClrMemory          perfCollector[perfDataClrMemory]
	perfClrRemoting        perfCollector[perfDataClrRemoting]
	perfClrSecurity        perfCollector[perfDataClrSecurity]

	collectorFns []func(ch chan<- prometheus.Metric, maxScrapeDuration time.Duration) error

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

	app.Action(func(*kingpin.ParseContext) error {
		c.config.CollectorsEnabled = strings.Split(collectorsEnabled, ",")

		return nil
	})

	return c
}

func (c *Collector) GetName() string {
	return Name
}

func (c *Collector) Close() error {
	for _, closeFn := range c.closeFns {
		closeFn()
	}

	c.closeFns = nil
	c.collectorFns = nil

	return nil
}

func (c *Collector) Build(logger *slog.Logger, _ *mi.Session) error {
	if len(c.config.CollectorsEnabled) == 0 {
		return nil
	}

	c.logger = logger
	if slices.Contains(c.config.CollectorsEnabled, collectorClrJIT) || slices.Contains(c.config.CollectorsEnabled, collectorClrSecurity) {
		var frequency int64

		ret, _, err := windows.NewLazySystemDLL("kernel32.dll").NewProc("QueryPerformanceFrequency").Call(uintptr(unsafe.Pointer(&frequency)))
		if ret == 0 || frequency <= 0 {
			return fmt.Errorf("QueryPerformanceFrequency: %w", err)
		}

		c.perfFrequency = float64(frequency)
	}

	c.collectorFns = make([]func(ch chan<- prometheus.Metric, maxScrapeDuration time.Duration) error, 0, len(c.config.CollectorsEnabled))

	subCollectors := map[string]struct {
		build   func() error
		collect func(ch chan<- prometheus.Metric, maxScrapeDuration time.Duration) error
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

	return nil
}

// Collect sends the metric values for each metric
// to the provided prometheus Metric channel.
func (c *Collector) Collect(ch chan<- prometheus.Metric, maxScrapeDuration time.Duration) error {
	var g recovery.Group

	for _, fn := range c.collectorFns {
		g.Go(func() error {
			return fn(ch, maxScrapeDuration)
		})
	}

	return g.Wait()
}

// newPerfCollector retains ownership of partially initialized PDH queries.
// Missing objects are optional; missing individual counters remain errors.
func newPerfCollector[T any](logger *slog.Logger, object string) (perfCollector[T], error) {
	c, err := pdh.NewCollector[T](logger, pdh.CounterTypeRaw, object, pdh.InstancesAll)
	if err == nil {
		return c, nil
	}

	if c != nil {
		c.Close()
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
