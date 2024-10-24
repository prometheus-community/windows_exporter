//go:build windows

package utils

import (
	"os"
	"strings"

	"github.com/prometheus-community/windows_exporter/internal/types"
)

func ExpandEnabledCollectors(enabled string) []string {
	expanded := strings.ReplaceAll(enabled, types.DefaultCollectorsPlaceholder, types.DefaultCollectors)
	separated := strings.Split(expanded, ",")
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

	return result
}

func PDHEnabled() bool {
	if v, ok := os.LookupEnv("WINDOWS_EXPORTER_PERF_COUNTERS_ENGINE"); ok && v == "pdh" {
		return true
	}

	return false
}

func MIEnabled() bool {
	if v, ok := os.LookupEnv("WINDOWS_EXPORTER_WMI_ENGINE"); ok && v == "mi" {
		return true
	}

	return false
}
