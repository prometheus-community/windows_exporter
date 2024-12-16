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

package exchange

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"sync"

	"github.com/alecthomas/kingpin/v2"
	"github.com/prometheus-community/windows_exporter/internal/mi"
	"github.com/prometheus/client_golang/prometheus"
)

const Name = "exchange"

const (
	subCollectorADAccessProcesses   = "ADAccessProcesses"
	subCollectorTransportQueues     = "TransportQueues"
	subCollectorHttpProxy           = "HttpProxy"
	subCollectorActiveSync          = "ActiveSync"
	subCollectorAvailabilityService = "AvailabilityService"
	subCollectorOutlookWebAccess    = "OutlookWebAccess"
	subCollectorAutoDiscover        = "Autodiscover"
	subCollectorWorkloadManagement  = "WorkloadManagement"
	subCollectorRpcClientAccess     = "RpcClientAccess"
	subCollectorMapiHttpEmsmdb      = "MapiHttpEmsmdb"
)

type Config struct {
	CollectorsEnabled []string `yaml:"collectors_enabled"`
}

//nolint:gochecknoglobals
var ConfigDefaults = Config{
	CollectorsEnabled: []string{
		subCollectorADAccessProcesses,
		subCollectorTransportQueues,
		subCollectorHttpProxy,
		subCollectorActiveSync,
		subCollectorAvailabilityService,
		subCollectorOutlookWebAccess,
		subCollectorAutoDiscover,
		subCollectorWorkloadManagement,
		subCollectorRpcClientAccess,
		subCollectorMapiHttpEmsmdb,
	},
}

type Collector struct {
	config Config

	collectorFns []func(ch chan<- prometheus.Metric) error
	closeFns     []func()

	collectorADAccessProcesses
	collectorActiveSync
	collectorAutoDiscover
	collectorAvailabilityService
	collectorHTTPProxy
	collectorMapiHttpEmsmdb
	collectorOWA
	collectorRpcClientAccess
	collectorTransportQueues
	collectorWorkloadManagementWorkloads
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

	var listAllCollectors bool

	var collectorsEnabled string

	app.Flag(
		"collector.exchange.list",
		"List the collectors along with their perflib object name/ids",
	).BoolVar(&listAllCollectors)

	app.Flag(
		"collector.exchange.enabled",
		"Comma-separated list of collectors to use. Defaults to all, if not specified.",
	).Default(strings.Join(ConfigDefaults.CollectorsEnabled, ",")).StringVar(&collectorsEnabled)

	app.PreAction(func(*kingpin.ParseContext) error {
		if listAllCollectors {
			collectorDesc := map[string]string{
				subCollectorADAccessProcesses:   "[19108] MSExchange ADAccess Processes",
				subCollectorTransportQueues:     "[20524] MSExchangeTransport Queues",
				subCollectorHttpProxy:           "[36934] MSExchange HttpProxy",
				subCollectorActiveSync:          "[25138] MSExchange ActiveSync",
				subCollectorAvailabilityService: "[24914] MSExchange Availability Service",
				subCollectorOutlookWebAccess:    "[24618] MSExchange OWA",
				subCollectorAutoDiscover:        "[29240] MSExchange Autodiscover",
				subCollectorWorkloadManagement:  "[19430] MSExchange WorkloadManagement Workloads",
				subCollectorRpcClientAccess:     "[29336] MSExchange RpcClientAccess",
				subCollectorMapiHttpEmsmdb:      "[26463] MSExchange MapiHttp Emsmdb",
			}

			sb := strings.Builder{}
			sb.WriteString(fmt.Sprintf("%-32s %-32s\n", "Collector Name", "[PerfID] Perflib Object"))

			for _, cname := range ConfigDefaults.CollectorsEnabled {
				sb.WriteString(fmt.Sprintf("%-32s %-32s\n", cname, collectorDesc[cname]))
			}

			app.UsageTemplate(sb.String()).Usage(nil)

			os.Exit(0)
		}

		return nil
	})

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
	for _, fn := range c.closeFns {
		fn()
	}

	return nil
}

func (c *Collector) Build(_ *slog.Logger, _ *mi.Session) error {
	subCollectors := map[string]struct {
		build   func() error
		collect func(ch chan<- prometheus.Metric) error
		close   func()
	}{
		subCollectorADAccessProcesses: {
			build:   c.buildADAccessProcesses,
			collect: c.collectADAccessProcesses,
			close:   c.perfDataCollectorADAccessProcesses.Close,
		},
		subCollectorTransportQueues: {
			build:   c.buildTransportQueues,
			collect: c.collectTransportQueues,
			close:   c.perfDataCollectorTransportQueues.Close,
		},
		subCollectorHttpProxy: {
			build:   c.buildHTTPProxy,
			collect: c.collectHTTPProxy,
			close:   c.perfDataCollectorHTTPProxy.Close,
		},
		subCollectorActiveSync: {
			build:   c.buildActiveSync,
			collect: c.collectActiveSync,
			close:   c.perfDataCollectorActiveSync.Close,
		},
		subCollectorAvailabilityService: {
			build:   c.buildAvailabilityService,
			collect: c.collectAvailabilityService,
			close:   c.perfDataCollectorAvailabilityService.Close,
		},
		subCollectorOutlookWebAccess: {
			build:   c.buildOWA,
			collect: c.collectOWA,
			close:   c.perfDataCollectorOWA.Close,
		},
		subCollectorAutoDiscover: {
			build:   c.buildAutoDiscover,
			collect: c.collectAutoDiscover,
			close:   c.perfDataCollectorAutoDiscover.Close,
		},
		subCollectorWorkloadManagement: {
			build:   c.buildWorkloadManagementWorkloads,
			collect: c.collectWorkloadManagementWorkloads,
			close:   c.perfDataCollectorWorkloadManagementWorkloads.Close,
		},
		subCollectorRpcClientAccess: {
			build:   c.buildRpcClientAccess,
			collect: c.collectRpcClientAccess,
			close:   c.perfDataCollectorRpcClientAccess.Close,
		},
		subCollectorMapiHttpEmsmdb: {
			build:   c.buildMapiHttpEmsmdb,
			collect: c.collectMapiHttpEmsmdb,
			close:   c.perfDataCollectorMapiHttpEmsmdb.Close,
		},
	}

	errs := make([]error, 0, len(c.config.CollectorsEnabled))

	for _, name := range c.config.CollectorsEnabled {
		if _, ok := subCollectors[name]; !ok {
			return fmt.Errorf("unknown collector: %s", name)
		}

		if err := subCollectors[name].build(); err != nil {
			errs = append(errs, fmt.Errorf("failed to build %s collector: %w", name, err))

			continue
		}

		c.collectorFns = append(c.collectorFns, subCollectors[name].collect)
		c.closeFns = append(c.closeFns, subCollectors[name].close)
	}

	return errors.Join(errs...)
}

// Collect collects exchange metrics and sends them to prometheus.
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

// toLabelName converts strings to lowercase and replaces all whitespaces and dots with underscores.
func (c *Collector) toLabelName(name string) string {
	s := strings.ReplaceAll(strings.Join(strings.Fields(strings.ToLower(name)), "_"), ".", "_")
	s = strings.ReplaceAll(s, "__", "_")

	return s
}
