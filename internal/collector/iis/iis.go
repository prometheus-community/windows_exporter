//go:build windows

package iis

import (
	"errors"
	"fmt"
	"log/slog"
	"maps"
	"regexp"
	"slices"
	"strings"

	"github.com/alecthomas/kingpin/v2"
	"github.com/prometheus-community/windows_exporter/internal/mi"
	"github.com/prometheus-community/windows_exporter/internal/perfdata"
	"github.com/prometheus-community/windows_exporter/internal/types"
	"github.com/prometheus/client_golang/prometheus"
	"golang.org/x/sys/windows/registry"
)

const Name = "iis"

type Config struct {
	SiteInclude *regexp.Regexp `yaml:"site_include"`
	SiteExclude *regexp.Regexp `yaml:"site_exclude"`
	AppInclude  *regexp.Regexp `yaml:"app_include"`
	AppExclude  *regexp.Regexp `yaml:"app_exclude"`
}

var ConfigDefaults = Config{
	SiteInclude: types.RegExpAny,
	SiteExclude: types.RegExpEmpty,
	AppInclude:  types.RegExpAny,
	AppExclude:  types.RegExpEmpty,
}

type Collector struct {
	config     Config
	iisVersion simpleVersion

	info *prometheus.Desc
	collectorWebService
	collectorAppPoolWAS
	collectorW3SVCW3WP
	collectorWebServiceCache
}

func New(config *Config) *Collector {
	if config == nil {
		config = &ConfigDefaults
	}

	if config.AppExclude == nil {
		config.AppExclude = ConfigDefaults.AppExclude
	}

	if config.AppInclude == nil {
		config.AppInclude = ConfigDefaults.AppInclude
	}

	if config.SiteExclude == nil {
		config.SiteExclude = ConfigDefaults.SiteExclude
	}

	if config.SiteInclude == nil {
		config.SiteInclude = ConfigDefaults.SiteInclude
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

	var appExclude, appInclude, siteExclude, siteInclude string

	app.Flag(
		"collector.iis.app-exclude",
		"Regexp of apps to exclude. App name must both match include and not match exclude to be included.",
	).Default("").StringVar(&appExclude)

	app.Flag(
		"collector.iis.app-include",
		"Regexp of apps to include. App name must both match include and not match exclude to be included.",
	).Default(".+").StringVar(&appInclude)

	app.Flag(
		"collector.iis.site-exclude",
		"Regexp of sites to exclude. Site name must both match include and not match exclude to be included.",
	).Default("").StringVar(&siteExclude)

	app.Flag(
		"collector.iis.site-include",
		"Regexp of sites to include. Site name must both match include and not match exclude to be included.",
	).Default(".+").StringVar(&siteInclude)

	app.Action(func(*kingpin.ParseContext) error {
		var err error

		c.config.AppExclude, err = regexp.Compile(fmt.Sprintf("^(?:%s)$", appExclude))
		if err != nil {
			return fmt.Errorf("collector.iis.app-exclude: %w", err)
		}

		c.config.AppInclude, err = regexp.Compile(fmt.Sprintf("^(?:%s)$", appInclude))
		if err != nil {
			return fmt.Errorf("collector.iis.app-include: %w", err)
		}

		c.config.SiteExclude, err = regexp.Compile(fmt.Sprintf("^(?:%s)$", siteExclude))
		if err != nil {
			return fmt.Errorf("collector.iis.site-exclude: %w", err)
		}

		c.config.SiteInclude, err = regexp.Compile(fmt.Sprintf("^(?:%s)$", siteInclude))
		if err != nil {
			return fmt.Errorf("collector.iis.site-include: %w", err)
		}

		return nil
	})

	return c
}

func (c *Collector) GetName() string {
	return Name
}

func (c *Collector) Close() error {
	c.perfDataCollectorWebService.Close()
	c.perfDataCollectorAppPoolWAS.Close()
	c.perfDataCollectorW3SVCW3WP.Close()
	c.perfDataCollectorWebServiceCache.Close()

	return nil
}

func (c *Collector) Build(logger *slog.Logger, _ *mi.Session) error {
	logger = logger.With(slog.String("collector", Name))

	c.iisVersion = c.getIISVersion(logger)

	c.info = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "info"),
		"ISS information",
		[]string{},
		prometheus.Labels{"version": fmt.Sprintf("%d.%d", c.iisVersion.major, c.iisVersion.minor)},
	)

	if err := c.buildWebService(); err != nil {
		return fmt.Errorf("failed to build Web Service collector: %w", err)
	}

	if err := c.buildAppPoolWAS(); err != nil {
		return fmt.Errorf("failed to build APP_POOL_WAS collector: %w", err)
	}

	if err := c.buildW3SVCW3WP(); err != nil {
		return fmt.Errorf("failed to build W3SVC_W3WP collector: %w", err)
	}

	if err := c.buildWebServiceCache(); err != nil {
		return fmt.Errorf("failed to build Web Service Cache collector: %w", err)
	}

	return nil
}

type simpleVersion struct {
	major uint64
	minor uint64
}

func (c *Collector) getIISVersion(logger *slog.Logger) simpleVersion {
	k, err := registry.OpenKey(registry.LOCAL_MACHINE, `SOFTWARE\Microsoft\InetStp\`, registry.QUERY_VALUE)
	if err != nil {
		logger.Warn("Couldn't open registry to determine IIS version",
			slog.Any("err", err),
		)

		return simpleVersion{}
	}

	defer func() {
		err = k.Close()
		if err != nil {
			logger.Warn("Failed to close registry key",
				slog.Any("err", err),
			)
		}
	}()

	major, _, err := k.GetIntegerValue("MajorVersion")
	if err != nil {
		logger.Warn("Couldn't open registry to determine IIS version",
			slog.Any("err", err),
		)

		return simpleVersion{}
	}

	minor, _, err := k.GetIntegerValue("MinorVersion")
	if err != nil {
		logger.Warn("Couldn't open registry to determine IIS version",
			slog.Any("err", err),
		)

		return simpleVersion{}
	}

	logger.Debug(fmt.Sprintf("Detected IIS %d.%d\n", major, minor))

	return simpleVersion{
		major: major,
		minor: minor,
	}
}

// Collect sends the metric values for each metric
// to the provided prometheus Metric channel.
func (c *Collector) Collect(ch chan<- prometheus.Metric) error {
	ch <- prometheus.MustNewConstMetric(
		c.info,
		prometheus.GaugeValue,
		1,
	)

	errs := make([]error, 0, 4)

	if err := c.collectWebService(ch); err != nil {
		errs = append(errs, fmt.Errorf("failed to collect Web Service metrics: %w", err))
	}

	if err := c.collectAppPoolWAS(ch); err != nil {
		errs = append(errs, fmt.Errorf("failed to collect APP_POOL_WAS metrics: %w", err))
	}

	if err := c.collectW3SVCW3WP(ch); err != nil {
		errs = append(errs, fmt.Errorf("failed to collect W3SVC_W3WP metrics: %w", err))
	}

	if err := c.collectWebServiceCache(ch); err != nil {
		errs = append(errs, fmt.Errorf("failed to collect Web Service Cache metrics: %w", err))
	}

	return errors.Join(errs...)
}

// deduplicateIISNames deduplicate IIS site names from various IIS perflib objects.
//
// E.G. Given the following list of site names, "Site_B" would be
// discarded, and "Site_B#2" would be kept and presented as "Site_B" in the
// Collector metrics.
// [ "Site_A", "Site_B", "Site_C", "Site_B#2" ].
func deduplicateIISNames(counterValues map[string]map[string]perfdata.CounterValues) {
	services := slices.Collect(maps.Keys(counterValues))

	// Ensure IIS entry with the highest suffix occurs last
	slices.Sort(services)

	// Use map to deduplicate IIS entries
	for _, entry := range services {
		name := strings.Split(entry, "#")[0]
		if name == entry {
			continue
		}

		counterValues[name] = counterValues[entry]
		delete(counterValues, entry)
	}
}
