//go:build windows

package utils

import (
	"slices"
	"strings"

	"github.com/prometheus-community/windows_exporter/pkg/collector"
)

func ExpandEnabledCollectors(enabled string) []string {
	expanded := strings.ReplaceAll(enabled, collector.DefaultCollectorsPlaceholder, collector.DefaultCollectors)

	return slices.Compact(strings.Split(expanded, ","))
}
