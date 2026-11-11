// SPDX-License-Identifier: MIT
//
// Copyright The Prometheus Authors

//go:build windows

package config

import (
	"maps"
	"slices"
	"testing"

	"go.yaml.in/yaml/v3"
)

// Unmarshal good configuration file and confirm data is flattened correctly.
func TestConfigFlattening(t *testing.T) {
	t.Parallel()

	goodYamlConfig := []byte(`---

    collectors:
      enabled: cpu,net,service

    collector:
      textfile:
        directories: [C:\a, C:\b]
      service:
        include:

    log:
      level: debug`)

	var data map[string]any

	err := yaml.Unmarshal(goodYamlConfig, &data)
	if err != nil {
		t.Error(err)
	}

	expectedResult := map[string][]string{
		"collectors.enabled":             {"cpu,net,service"},
		"collector.textfile.directories": {`C:\a`, `C:\b`},
		"log.level":                      {"debug"},
	}
	flattenedValues := flatten(data)

	if !maps.EqualFunc(expectedResult, flattenedValues, slices.Equal) {
		t.Errorf("Flattened values do not match!\nExpected result: %s\nActual result: %s", expectedResult, flattenedValues)
	}
}
