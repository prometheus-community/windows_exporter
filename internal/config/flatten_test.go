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
