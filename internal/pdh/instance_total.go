// SPDX-License-Identifier: MIT
//
// Copyright The Prometheus Authors

//go:build windows

package pdh

import (
	"strconv"
	"strings"
)

// IsTotalInstance identifies aggregate instances without discarding ordinary
// application names ending in _Total. Processor Information also publishes
// processor-group aggregates named <group>,_Total.
func IsTotalInstance(object, instance string) bool {
	if instance == InstanceTotal {
		return true
	}

	if !strings.EqualFold(object, "Processor Information") {
		return false
	}

	group, ok := strings.CutSuffix(instance, ","+InstanceTotal)
	if !ok {
		return false
	}

	_, err := strconv.ParseUint(group, 10, 32)

	return err == nil
}
