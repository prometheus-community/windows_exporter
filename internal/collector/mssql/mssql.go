//go:build windows

package mssql

import (
	"errors"
	"fmt"
	"log/slog"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
	"unsafe"

	"github.com/Microsoft/go-winio/pkg/process"
	"github.com/alecthomas/kingpin/v2"
	"github.com/prometheus-community/windows_exporter/internal/headers/iphlpapi"
	"github.com/prometheus-community/windows_exporter/internal/mi"
	"github.com/prometheus-community/windows_exporter/internal/perfdata"
	"github.com/prometheus-community/windows_exporter/internal/types"
	"github.com/prometheus/client_golang/prometheus"
	"golang.org/x/sys/windows"
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
	subCollectorLocks               = "locks"
	subCollectorMemoryManager       = "memmgr"
	subCollectorSQLErrors           = "sqlerrors"
	subCollectorSQLStats            = "sqlstats"
	subCollectorTransactions        = "transactions"
	subCollectorWaitStats           = "waitstats"
)

type Config struct {
	CollectorsEnabled []string `yaml:"collectors_enabled"`
	Port              uint16   `yaml:"port"`
}

var ConfigDefaults = Config{
	CollectorsEnabled: []string{
		subCollectorAccessMethods,
		subCollectorAvailabilityReplica,
		subCollectorBufferManager,
		subCollectorDatabases,
		subCollectorDatabaseReplica,
		subCollectorGeneralStatistics,
		subCollectorLocks,
		subCollectorMemoryManager,
		subCollectorSQLErrors,
		subCollectorSQLStats,
		subCollectorTransactions,
		subCollectorWaitStats,
	},
	Port: 1433,
}

// A Collector is a Prometheus Collector for various WMI Win32_PerfRawData_MSSQLSERVER_* metrics.
type Collector struct {
	config Config

	logger *slog.Logger

	mssqlInstances mssqlInstancesType
	collectorFns   []func(ch chan<- prometheus.Metric) error
	closeFns       []func()

	fileVersion    string
	productVersion string

	// meta
	mssqlScrapeDurationDesc *prometheus.Desc
	mssqlScrapeSuccessDesc  *prometheus.Desc
	mssqlInfoDesc           *prometheus.Desc

	collectorAccessMethods
	collectorAvailabilityReplica
	collectorBufferManager
	collectorDatabaseReplica
	collectorDatabases
	collectorGeneralStatistics
	collectorLocks
	collectorMemoryManager
	collectorSQLErrors
	collectorSQLStats
	collectorTransactions
	collectorWaitStats
}

type mssqlInstancesType map[string]string

func New(config *Config) *Collector {
	if config == nil {
		config = &ConfigDefaults
	}

	if config.CollectorsEnabled == nil {
		config.CollectorsEnabled = ConfigDefaults.CollectorsEnabled
	}

	if config.Port == 0 {
		config.Port = ConfigDefaults.Port
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

	app.Flag(
		"collector.mssql.port",
		"Port of MSSQL server used for windows_mssql_info metric.",
	).Default(strconv.FormatUint(uint64(c.config.Port), 10)).Uint16Var(&c.config.Port)

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
	c.mssqlInstances = c.getMSSQLInstances()

	fileVersion, productVersion, err := c.getMSSQLServerVersion(c.config.Port)
	if err != nil {
		logger.Warn("Failed to get MSSQL server version",
			slog.Any("err", err),
			slog.String("collector", Name),
		)
	}

	c.fileVersion = fileVersion
	c.productVersion = productVersion

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

	for _, name := range c.config.CollectorsEnabled {
		if _, ok := subCollectors[name]; !ok {
			return fmt.Errorf("unknown collector: %s", name)
		}

		if err := subCollectors[name].build(); err != nil {
			return fmt.Errorf("failed to build %s collector: %w", name, err)
		}

		c.collectorFns = append(c.collectorFns, subCollectors[name].collect)
		c.closeFns = append(c.closeFns, subCollectors[name].close)
	}

	// meta
	c.mssqlInfoDesc = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "info"),
		"mssql server information",
		[]string{"file_version", "version"},
		nil,
	)

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

func (c *Collector) getMSSQLInstances() mssqlInstancesType {
	sqlInstances := make(mssqlInstancesType)

	// in case querying the registry fails, return the default instance
	sqlDefaultInstance := make(mssqlInstancesType)
	sqlDefaultInstance["MSSQLSERVER"] = ""

	regKey := `Software\Microsoft\Microsoft SQL Server\Instance Names\SQL`

	k, err := registry.OpenKey(registry.LOCAL_MACHINE, regKey, registry.QUERY_VALUE)
	if err != nil {
		c.logger.Warn("Couldn't open registry to determine SQL instances",
			slog.Any("err", err),
		)

		return sqlDefaultInstance
	}

	defer func(key registry.Key) {
		if err := key.Close(); err != nil {
			c.logger.Warn("Failed to close registry key",
				slog.Any("err", err),
			)
		}
	}(k)

	instanceNames, err := k.ReadValueNames(0)
	if err != nil {
		c.logger.Warn("Can't ReadSubKeyNames",
			slog.Any("err", err),
		)

		return sqlDefaultInstance
	}

	for _, instanceName := range instanceNames {
		if instanceVersion, _, err := k.GetStringValue(instanceName); err == nil {
			sqlInstances[instanceName] = instanceVersion
		}
	}

	c.logger.Debug(fmt.Sprintf("Detected MSSQL Instances: %#v\n", sqlInstances))

	return sqlInstances
}

// mssqlGetPerfObjectName returns the name of the Windows Performance
// Counter object for the given SQL instance and Collector.
func (c *Collector) mssqlGetPerfObjectName(sqlInstance string, collector string) string {
	sb := strings.Builder{}
	sb.WriteString("SQLServer:")

	if sqlInstance != "MSSQLSERVER" {
		sb.WriteString("MSSQL$")
		sb.WriteString(sqlInstance)
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
	perfDataCollectors map[string]*perfdata.Collector,
	collectFn func(ch chan<- prometheus.Metric, sqlInstance string, perfDataCollector *perfdata.Collector) error,
) error {
	errs := make([]error, 0, len(perfDataCollectors))

	for sqlInstance, perfDataCollector := range perfDataCollectors {
		begin := time.Now()
		success := 1.0
		err := collectFn(ch, sqlInstance, perfDataCollector)
		duration := time.Since(begin).Seconds()

		if err != nil {
			errs = append(errs, err)
			success = 0.0

			c.logger.Error(fmt.Sprintf("mssql class collector %s failed after %fs", collector, duration),
				slog.Any("err", err),
			)
		} else {
			c.logger.Debug(fmt.Sprintf("mssql class collector %s succeeded after %fs.", collector, duration))
		}

		ch <- prometheus.MustNewConstMetric(
			c.mssqlScrapeDurationDesc,
			prometheus.GaugeValue,
			duration,
			collector, sqlInstance,
		)
		ch <- prometheus.MustNewConstMetric(
			c.mssqlScrapeSuccessDesc,
			prometheus.GaugeValue,
			success,
			collector, sqlInstance,
		)
	}

	return errors.Join(errs...)
}

// getMSSQLServerVersion get the version of the SQL Server instance by
// reading the version information from the process running the SQL Server instance port.
func (c *Collector) getMSSQLServerVersion(port uint16) (string, string, error) {
	pid, err := iphlpapi.GetOwnerPIDOfTCPPort(windows.AF_INET, port)
	if err != nil {
		return "", "", fmt.Errorf("failed to get the PID of the process running on port 1433: %w", err)
	}

	hProcess, err := windows.OpenProcess(windows.PROCESS_QUERY_LIMITED_INFORMATION, false, pid)
	if err != nil {
		return "", "", fmt.Errorf("failed to open the process with PID %d: %w", pid, err)
	}

	defer windows.CloseHandle(hProcess) //nolint:errcheck

	processFilePath, err := process.QueryFullProcessImageName(hProcess, process.ImageNameFormatWin32Path)
	if err != nil {
		return "", "", fmt.Errorf("failed to query the full path of the process with PID %d: %w", pid, err)
	}

	// Load the file version information
	size, err := windows.GetFileVersionInfoSize(processFilePath, nil)
	if err != nil {
		return "", "", fmt.Errorf("failed to get the size of the file version information: %w", err)
	}

	fileVersionInfo := make([]byte, size)

	err = windows.GetFileVersionInfo(processFilePath, 0, size, unsafe.Pointer(&fileVersionInfo[0]))
	if err != nil {
		return "", "", fmt.Errorf("failed to get the file version information: %w", err)
	}

	var (
		verData *byte
		verSize uint32
	)

	err = windows.VerQueryValue(
		unsafe.Pointer(&fileVersionInfo[0]),
		`\StringFileInfo\040904b0\ProductVersion`,
		unsafe.Pointer(&verData),
		&verSize,
	)
	if err != nil {
		return "", "", fmt.Errorf("failed to query the product version: %w", err)
	}

	productVersion := windows.UTF16ToString((*[1 << 16]uint16)(unsafe.Pointer(verData))[:verSize])

	err = windows.VerQueryValue(
		unsafe.Pointer(&fileVersionInfo[0]),
		`\StringFileInfo\040904b0\FileVersion`,
		unsafe.Pointer(&verData),
		&verSize,
	)
	if err != nil {
		return "", "", fmt.Errorf("failed to query the file version: %w", err)
	}

	fileVersion := windows.UTF16ToString((*[1 << 16]uint16)(unsafe.Pointer(verData))[:verSize])

	return fileVersion, productVersion, nil
}
