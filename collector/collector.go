package collector

import (
	"fmt"
	"gopkg.in/alecthomas/kingpin.v2"
	"sort"
	"strconv"
	"strings"

	"github.com/leoluk/perflib_exporter/perflib"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
	"golang.org/x/sys/windows/registry"
)

// ...
const (
	// TODO: Make package-local
	Namespace = "windows"

	// Conversion factors
	ticksToSecondsScaleFactor = 1 / 1e7
	windowsEpoch              = 116444736000000000
)

// getWindowsVersion reads the version number of the OS from the Registry
// See https://docs.microsoft.com/en-us/windows/desktop/sysinfo/operating-system-version
func getWindowsVersion() float64 {
	k, err := registry.OpenKey(registry.LOCAL_MACHINE, `SOFTWARE\Microsoft\Windows NT\CurrentVersion`, registry.QUERY_VALUE)
	if err != nil {
		log.Warn("Couldn't open registry", err)
		return 0
	}
	defer func() {
		err = k.Close()
		if err != nil {
			log.Warnf("Failed to close registry key: %v", err)
		}
	}()

	currentv, _, err := k.GetStringValue("CurrentVersion")
	if err != nil {
		log.Warn("Couldn't open registry to determine current Windows version:", err)
		return 0
	}

	currentv_flt, err := strconv.ParseFloat(currentv, 64)

	log.Debugf("Detected Windows version %f\n", currentv_flt)

	return currentv_flt
}

/*
Builders contains the function necessary to build a new collector
ConfigMap contains the collection of Configuration options used for the kingpin integration
ConfigInstanceMap contains the actual values for the config, when used as a standalone the Instance is a singleton,
	when used as a Library it is not used at all but instead of map is passed into the NewWindowsCollector to instantiate
	from that specific configuration
*/
var (
	builders                = make(map[string]builderFactory)
	configMap               = make(map[string]Config)
	configInstanceMap       = make(map[string]*ConfigInstance)
	perfCounterDependencies = make(map[string]string)

)

/*
These whole build* interfaces and functions are to enforce compile time checks for the RegisterCollector classes.
*/
type builderFactory interface {
	build() (Collector, error)
}

type buildInstance struct {
	instanceBuilder func() (Collector, error)
}

func (b *buildInstance) build() (Collector, error) {
	return b.instanceBuilder()
}


type buildConfigInstance struct {
	instanceBuilder func() (ConfigurableCollector, error)
}

func (b *buildConfigInstance) build() (Collector, error) {
	return b.instanceBuilder()
}

func registerCollector(name string, builder func() (Collector, error), perfCounterNames ...string) {
	instance := buildInstance{instanceBuilder: builder}
	builders[name] = &instance
	perfCounterDependencies[name] = addPerfCounterDependencies(perfCounterNames)
}

func registerCollectorWithConfig(name string, builder func() (ConfigurableCollector, error), config []Config, perfCounterNames ...string) {
	instance := buildConfigInstance{instanceBuilder: builder}
	builders[name] = &instance
	for _,v := range config {
		ci := &ConfigInstance{
			Value:  "",
			Config: v,
		}
		configInstanceMap[v.Name] = ci
		configMap[v.Name] = v
	}
	perfCounterDependencies[name] = addPerfCounterDependencies(perfCounterNames)
}

func ApplyKingpinConfig(app *kingpin.Application) map[string]*ConfigInstance {
	for _,v := range configInstanceMap {
		app.Flag(v.Name,v.HelpText).Default(v.Default).Action(setExists).StringVar(&v.Value)
	}
	return configInstanceMap
}

/*
This exists mostly to support the Bool parameter
*/
func setExists(ctx *kingpin.ParseContext) error {
	for _,v := range ctx.Elements {
		name := ""
		if c, ok := v.Clause.(*kingpin.CmdClause); ok {
			name = c.Model().Name
		} else if c, ok := v.Clause.(*kingpin.FlagClause); ok {
			name = c.Model().Name
		} else if c, ok := v.Clause.(*kingpin.ArgClause); ok {
			name = c.Model().Name
		} else {
			continue
		}
		configInstanceMap[name].Exists = true
	}

	return nil
}

/*
Used to hole the metadata about a configuration option
*/
type Config struct {
	Name string
	HelpText string
	Default string
}

/*
Used to hold the actual values for a configuration option
*/
type ConfigInstance struct {
	Value string
	Exists bool
	Config
}

func Available() []string {
	cs := make([]string, 0, len(builders))
	for c := range builders {
		cs = append(cs, c)
	}
	return cs
}

func Build(collector string, settings map[string]*ConfigInstance) (Collector, error) {
	builder, exists := builders[collector]
	if !exists {
		return nil, fmt.Errorf("Unknown collector %q", collector)
	}
	c, err := builder.build()
	if err != nil {
		return nil, err
	}
	if v, ok := c.(ConfigurableCollector) ; ok {
		v.ApplyConfig(settings)
	}
	return c, err
}

func addPerfCounterDependencies(perfCounterNames []string) string {
	perfIndicies := make([]string, 0, len(perfCounterNames))
	for _, cn := range perfCounterNames {
		perfIndicies = append(perfIndicies, MapCounterToIndex(cn))
	}
	return strings.Join(perfIndicies, " ")
}


func getPerfQuery(collectors []string) string {
	parts := make([]string, 0, len(collectors))
	for _, c := range collectors {
		if p := perfCounterDependencies[c]; p != "" {
			parts = append(parts, p)
		}
	}
	return strings.Join(parts, " ")
}

type Collector interface {
	// Get new metrics and expose them via prometheus registry.
	Collect(ctx *ScrapeContext, ch chan<- prometheus.Metric) (err error)
}

/*
This interface is used when a Collector needs to have configuration. This code should support multiple collectors of
the same type which means we cannot use the global var based configuration.

Setup is used for any checking that needs to happen before the collector starts
*/
type ConfigurableCollector interface {
	Collect(ctx *ScrapeContext, ch chan<- prometheus.Metric) (err error)
	ApplyConfig(map[string]*ConfigInstance)
	Setup()

}


type ScrapeContext struct {
	perfObjects map[string]*perflib.PerfObject
}

// PrepareScrapeContext creates a ScrapeContext to be used during a single scrape
func PrepareScrapeContext(collectors []string) (*ScrapeContext, error) {
	q := getPerfQuery(collectors) // TODO: Memoize
	objs, err := getPerflibSnapshot(q)
	if err != nil {
		return nil, err
	}

	return &ScrapeContext{objs}, nil
}

func boolToFloat(b bool) float64 {
	if b {
		return 1.0
	}
	return 0.0
}

func find(slice []string, val string) bool {
	for _, item := range slice {
		if item == val {
			return true
		}
	}
	return false
}

// Used by more complex collectors where user input specifies enabled child collectors.
// Splits provided child collectors and deduplicate.
func expandEnabledChildCollectors(enabled string) []string {
	separated := strings.Split(enabled, ",")
	unique := map[string]bool{}
	for _, s := range separated {
		if s != "" {
			unique[s] = true
		}
	}
	result := make([]string, 0, len(unique))
	for s := range unique {
		result = append(result, s)
	}
	// Ensure result is ordered, to prevent test failure
	sort.Strings(result)
	return result
}

func getValueFromMap(m map[string]*ConfigInstance, key string) string {
	if v, exists := m[key]; exists {
		if v.Exists {
			return v.Value
		}
		return v.Default
	}
	return ""
}