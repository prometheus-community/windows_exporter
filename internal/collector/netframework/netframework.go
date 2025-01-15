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

package netframework

import (
	"errors"
	"fmt"
	"log/slog"
	"sort"
	"strings"
	"sync"

	"github.com/alecthomas/kingpin/v2"
	"github.com/prometheus-community/windows_exporter/internal/mi"
	"github.com/prometheus/client_golang/prometheus"
)

const Name = "netframework"

type Config struct {
	CollectorsEnabled []string `yaml:"collectors_enabled"`
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

// A Collector is a Prometheus Collector for WMI Win32_PerfRawData_NETFramework_NETCLRExceptions metrics.
type Collector struct {
	config    Config
	miSession *mi.Session

	collectorFns []func(ch chan<- prometheus.Metric) error

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
	return nil
}

func (c *Collector) Build(_ *slog.Logger, miSession *mi.Session) error {
	if len(c.config.CollectorsEnabled) == 0 {
		return nil
	}

	if miSession == nil {
		return errors.New("miSession is nil")
	}

	c.miSession = miSession

	c.collectorFns = make([]func(ch chan<- prometheus.Metric) error, 0, len(c.config.CollectorsEnabled))

	subCollectors := map[string]struct {
		build   func()
		collect func(ch chan<- prometheus.Metric) error
		close   func()
	}{
		collectorClrExceptions: {
			build:   c.buildClrExceptions,
			collect: c.collectClrExceptions,
		},
		collectorClrJIT: {
			build:   c.buildClrJIT,
			collect: c.collectClrJIT,
		},
		collectorClrLoading: {
			build:   c.buildClrLoading,
			collect: c.collectClrLoading,
		},
		collectorClrInterop: {
			build:   c.buildClrInterop,
			collect: c.collectClrInterop,
		},
		collectorClrLocksAndThreads: {
			build:   c.buildClrLocksAndThreads,
			collect: c.collectClrLocksAndThreads,
		},
		collectorClrMemory: {
			build:   c.buildClrMemory,
			collect: c.collectClrMemory,
		},
		collectorClrRemoting: {
			build:   c.buildClrRemoting,
			collect: c.collectClrRemoting,
		},
		collectorClrSecurity: {
			build:   c.buildClrSecurity,
			collect: c.collectClrSecurity,
		},
	}

	// Result must order, to prevent test failures.
	sort.Strings(c.config.CollectorsEnabled)

	for _, name := range c.config.CollectorsEnabled {
		if _, ok := subCollectors[name]; !ok {
			return fmt.Errorf("unknown collector: %s", name)
		}

		subCollectors[name].build()

		c.collectorFns = append(c.collectorFns, subCollectors[name].collect)
	}

	return nil
}

// Collect sends the metric values for each metric
// to the provided prometheus Metric channel.
func (c *Collector) Collect(ch chan<- prometheus.Metric) error {
	errCh := make(chan error, len(c.collectorFns))
	errs := make([]error, 0, len(c.collectorFns))

	wg := sync.WaitGroup{}

	for _, fn := range c.collectorFns {
		wg.Add(1)

		go func(fn func(ch chan<- prometheus.Metric) error) {
			defer wg.Done()

			if err := fn(ch); err != nil {
				errCh <- err
			}
		}(fn)
	}

	wg.Wait()

	close(errCh)

	for err := range errCh {
		errs = append(errs, err)
	}

	return errors.Join(errs...)
}
