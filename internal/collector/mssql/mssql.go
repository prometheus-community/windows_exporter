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

package mssql

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/alecthomas/kingpin/v2"
	"github.com/prometheus-community/windows_exporter/internal/mi"
	"github.com/prometheus-community/windows_exporter/internal/pdh"
	"github.com/prometheus-community/windows_exporter/internal/types"
	"github.com/prometheus/client_golang/prometheus"
	"golang.org/x/sys/windows/registry"
)

const (
	Name = "mssql"

	subCollectorAccessMethods       = "accessmethods"
	subCollectorAvailabilityReplica = "availreplica"
	subCollectorBufferManager       = "bufman"
	subCollectorDatabases           = "databases"
	subCollectorDatabaseReplica     = "dbreplica"
	subCollectorGeneralStatistics   = "genstats"
	subCollectorInfo                = "info"
	subCollectorLocks               = "locks"
	subCollectorMemoryManager       = "memmgr"
	subCollectorSQLErrors           = "sqlerrors"
	subCollectorSQLStats            = "sqlstats"
	subCollectorTransactions        = "transactions"
	subCollectorWaitStats           = "waitstats"
)

type Config struct {
	CollectorsEnabled []string `yaml:"collectors_enabled"`
}

//nolint:gochecknoglobals
var ConfigDefaults = Config{
	CollectorsEnabled: []string{
		subCollectorAccessMethods,
		subCollectorAvailabilityReplica,
		subCollectorBufferManager,
		subCollectorDatabases,
		subCollectorDatabaseReplica,
		subCollectorGeneralStatistics,
		subCollectorInfo,
		subCollectorLocks,
		subCollectorMemoryManager,
		subCollectorSQLErrors,
		subCollectorSQLStats,
		subCollectorTransactions,
		subCollectorWaitStats,
	},
}

// A Collector is a Prometheus Collector for various WMI Win32_PerfRawData_MSSQLSERVER_* metrics.
type Collector struct {
	config Config

	logger *slog.Logger

	mssqlInstances []mssqlInstance
	collectorFns   []func(ch chan<- prometheus.Metric) error
	closeFns       []func()

	// meta
	mssqlScrapeDurationDesc *prometheus.Desc
	mssqlScrapeSuccessDesc  *prometheus.Desc

	collectorAccessMethods
	collectorAvailabilityReplica
	collectorBufferManager
	collectorDatabaseReplica
	collectorDatabases
	collectorGeneralStatistics
	collectorInstance
	collectorLocks
	collectorMemoryManager
	collectorSQLErrors
	collectorSQLStats
	collectorTransactions
	collectorWaitStats
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

	var collectorsEnabled string

	app.Flag(
		"collector.mssql.enabled",
		"Comma-separated list of collectors to use.",
	).Default(strings.Join(c.config.CollectorsEnabled, ",")).StringVar(&collectorsEnabled)

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

func (c *Collector) Build(logger *slog.Logger, _ *mi.Session) error {
	c.logger = logger.With(slog.String("collector", Name))

	instances, err := c.getMSSQLInstances()
	if err != nil {
		return fmt.Errorf("couldn't get SQL instances: %w", err)
	}

	c.mssqlInstances = instances

	subCollectors := map[string]struct {
		build   func() error
		collect func(ch chan<- prometheus.Metric) error
		close   func()
	}{
		subCollectorAccessMethods: {
			build:   c.buildAccessMethods,
			collect: c.collectAccessMethods,
			close:   c.closeAccessMethods,
		},
		subCollectorAvailabilityReplica: {
			build:   c.buildAvailabilityReplica,
			collect: c.collectAvailabilityReplica,
			close:   c.closeAvailabilityReplica,
		},
		subCollectorBufferManager: {
			build:   c.buildBufferManager,
			collect: c.collectBufferManager,
			close:   c.closeBufferManager,
		},
		subCollectorDatabases: {
			build:   c.buildDatabases,
			collect: c.collectDatabases,
			close:   c.closeDatabases,
		},
		subCollectorDatabaseReplica: {
			build:   c.buildDatabaseReplica,
			collect: c.collectDatabaseReplica,
			close:   c.closeDatabaseReplica,
		},
		subCollectorGeneralStatistics: {
			build:   c.buildGeneralStatistics,
			collect: c.collectGeneralStatistics,
			close:   c.closeGeneralStatistics,
		},
		subCollectorInfo: {
			build:   c.buildInstance,
			collect: c.collectInstance,
			close:   c.closeInstance,
		},
		subCollectorLocks: {
			build:   c.buildLocks,
			collect: c.collectLocks,
			close:   c.closeLocks,
		},
		subCollectorMemoryManager: {
			build:   c.buildMemoryManager,
			collect: c.collectMemoryManager,
			close:   c.closeMemoryManager,
		},
		subCollectorSQLErrors: {
			build:   c.buildSQLErrors,
			collect: c.collectSQLErrors,
			close:   c.closeSQLErrors,
		},
		subCollectorSQLStats: {
			build:   c.buildSQLStats,
			collect: c.collectSQLStats,
			close:   c.closeSQLStats,
		},
		subCollectorTransactions: {
			build:   c.buildTransactions,
			collect: c.collectTransactions,
			close:   c.closeTransactions,
		},
		subCollectorWaitStats: {
			build:   c.buildWaitStats,
			collect: c.collectWaitStats,
			close:   c.closeWaitStats,
		},
	}

	c.collectorFns = make([]func(ch chan<- prometheus.Metric) error, 0, len(c.config.CollectorsEnabled))
	c.closeFns = make([]func(), 0, len(c.config.CollectorsEnabled))
	// Result must order, to prevent test failures.
	sort.Strings(c.config.CollectorsEnabled)

	errs := make([]error, 0, len(c.config.CollectorsEnabled))

	for _, name := range c.config.CollectorsEnabled {
		if _, ok := subCollectors[name]; !ok {
			return fmt.Errorf("unknown collector: %s", name)
		}

		if err := subCollectors[name].build(); err != nil {
			errs = append(errs, fmt.Errorf("failed to build %s collector: %w", name, err))
		}

		c.collectorFns = append(c.collectorFns, subCollectors[name].collect)
		c.closeFns = append(c.closeFns, subCollectors[name].close)
	}

	c.mssqlScrapeDurationDesc = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "collector_duration_seconds"),
		"windows_exporter: Duration of an mssql child collection.",
		[]string{"collector", "mssql_instance"},
		nil,
	)
	c.mssqlScrapeSuccessDesc = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "collector_success"),
		"windows_exporter: Whether a mssql child collector was successful.",
		[]string{"collector", "mssql_instance"},
		nil,
	)

	return errors.Join(errs...)
}

// Collect sends the metric values for each metric
// to the provided prometheus Metric channel.
func (c *Collector) Collect(ch chan<- prometheus.Metric) error {
	if len(c.mssqlInstances) == 0 {
		return fmt.Errorf("no SQL instances found: %w", pdh.ErrNoData)
	}

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

func (c *Collector) getMSSQLInstances() ([]mssqlInstance, error) {
	regKey := `Software\Microsoft\Microsoft SQL Server\Instance Names\SQL`

	k, err := registry.OpenKey(registry.LOCAL_MACHINE, regKey, registry.QUERY_VALUE)
	if err != nil {
		return nil, fmt.Errorf("couldn't open registry to determine SQL instances: %w", err)
	}

	defer func(key registry.Key) {
		if err := key.Close(); err != nil {
			c.logger.Warn("failed to close registry key",
				slog.Any("err", err),
			)
		}
	}(k)

	instanceNames, err := k.ReadValueNames(0)
	if err != nil {
		return nil, fmt.Errorf("couldn't read subkey names: %w", err)
	}

	sqlInstances := make([]mssqlInstance, 0, len(instanceNames))

	for _, instanceName := range instanceNames {
		instanceVersion, _, err := k.GetStringValue(instanceName)
		if err != nil {
			return nil, fmt.Errorf("couldn't get instance info: %w", err)
		}

		instance, err := newMssqlInstance(instanceName, instanceVersion)
		if err != nil {
			return nil, err
		}

		sqlInstances = append(sqlInstances, instance)
	}

	c.logger.Debug(fmt.Sprintf("detected MSSQL Instances: %#v\n", sqlInstances))

	return sqlInstances, nil
}

// mssqlGetPerfObjectName returns the name of the Windows Performance
// Counter object for the given SQL instance and Collector.
func (c *Collector) mssqlGetPerfObjectName(sqlInstance mssqlInstance, collector string) string {
	sb := strings.Builder{}

	if sqlInstance.isFirstInstance {
		sb.WriteString("SQLServer:")
	} else {
		sb.WriteString("MSSQL$")
		sb.WriteString(sqlInstance.name)
		sb.WriteString(":")
	}

	sb.WriteString(collector)

	return sb.String()
}

// mssqlGetPerfObjectName returns the name of the Windows Performance
// Counter object for the given SQL instance and Collector.
func (c *Collector) collect(
	ch chan<- prometheus.Metric,
	collector string,
	perfDataCollectors map[mssqlInstance]*pdh.Collector,
	collectFn func(ch chan<- prometheus.Metric, sqlInstance mssqlInstance, perfDataCollector *pdh.Collector) error,
) error {
	errs := make([]error, 0, len(perfDataCollectors))

	ctx := context.Background()

	for sqlInstance, perfDataCollector := range perfDataCollectors {
		begin := time.Now()
		success := 1.0
		err := collectFn(ch, sqlInstance, perfDataCollector)
		duration := time.Since(begin)

		if err != nil && !errors.Is(err, pdh.ErrNoData) {
			errs = append(errs, err)
			success = 0.0

			c.logger.LogAttrs(ctx, slog.LevelDebug, fmt.Sprintf("mssql class collector %s for instance %s failed after %s", collector, sqlInstance.name, duration),
				slog.Any("err", err),
			)
		} else {
			c.logger.LogAttrs(ctx, slog.LevelDebug, fmt.Sprintf("mssql class collector %s for instance %s succeeded after %s", collector, sqlInstance.name, duration))
		}

		if collector == "" {
			continue
		}

		ch <- prometheus.MustNewConstMetric(
			c.mssqlScrapeDurationDesc,
			prometheus.GaugeValue,
			duration.Seconds(),
			collector, sqlInstance.name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.mssqlScrapeSuccessDesc,
			prometheus.GaugeValue,
			success,
			collector, sqlInstance.name,
		)
	}

	return errors.Join(errs...)
}
