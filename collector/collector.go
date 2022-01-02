package collector

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/leoluk/perflib_exporter/perflib"
	"github.com/prometheus-community/windows_exporter/log"
	"github.com/prometheus/client_golang/prometheus"
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

type collectorBuilder func() (Collector, error)

var (
	builders                = make(map[string]collectorBuilder)
	perfCounterDependencies = make(map[string]string)
)

func registerCollector(name string, builder collectorBuilder, perfCounterNames ...string) {
	builders[name] = builder
	addPerfCounterDependencies(name, perfCounterNames)
}

func addPerfCounterDependencies(name string, perfCounterNames []string) {
	perfIndicies := make([]string, 0, len(perfCounterNames))
	for _, cn := range perfCounterNames {
		perfIndicies = append(perfIndicies, MapCounterToIndex(cn))
	}
	perfCounterDependencies[name] = strings.Join(perfIndicies, " ")
}

func Available() []string {
	cs := make([]string, 0, len(builders))
	for c := range builders {
		cs = append(cs, c)
	}
	return cs
}
func Build(collector string) (Collector, error) {
	builder, exists := builders[collector]
	if !exists {
		return nil, fmt.Errorf("Unknown collector %q", collector)
	}
	return builder()
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

// Collector is the interface a collector has to implement.
type Collector interface {
	// Get new metrics and expose them via prometheus registry.
	Collect(ctx *ScrapeContext, ch chan<- prometheus.Metric) (err error)
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

func milliSecToSec(t float64) float64 {
	return t / 1000
}
