//go:build windows

package utils

import (
	"strings"

	"github.com/prometheus-community/windows_exporter/pkg/types"
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
