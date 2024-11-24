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

package config

import (
	"reflect"
	"testing"

	"gopkg.in/yaml.v3"
)

// Unmarshal good configuration file and confirm data is flattened correctly.
func TestConfigFlattening(t *testing.T) {
	t.Parallel()

	goodYamlConfig := []byte(`---

    collectors:
      enabled: cpu,net,service

    log:
      level: debug`)

	var data map[string]interface{}

	err := yaml.Unmarshal(goodYamlConfig, &data)
	if err != nil {
		t.Error(err)
	}

	expectedResult := map[string]string{
		"collectors.enabled": "cpu,net,service",
		"log.level":          "debug",
	}
	flattenedValues := flatten(data)

	if !reflect.DeepEqual(expectedResult, flattenedValues) {
		t.Errorf("Flattened values do not match!\nExpected result: %s\nActual result: %s", expectedResult, flattenedValues)
	}
}
