// SPDX-License-Identifier: Apache-2.0
//
// Copyright The Prometheus Authors
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

package collector

import (
	"maps"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/alecthomas/kingpin/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestRegistrations verifies that every place that lists the collectors agrees
// with the registrations: the flag builders, NewWithConfig, the fields and
// YAML keys of Config, ConfigDefaults and the collector documentation.
func TestRegistrations(t *testing.T) {
	t.Parallel()

	names := make([]string, 0, len(registrations))
	for _, r := range registrations {
		require.NotContains(t, names, r.name, "collector %s is registered twice", r.name)
		names = append(names, r.name)
	}

	slices.Sort(names)

	t.Run("flag builders", func(t *testing.T) {
		t.Parallel()

		require.Equal(t, names, Available())

		app := kingpin.New("test", "test")
		for name, builder := range BuildersWithFlags {
			assert.Equal(t, name, builder(app).GetName(), "flag builder of collector %s", name)
		}
	})

	t.Run("NewWithConfig", func(t *testing.T) {
		t.Parallel()

		collection := NewWithConfig(ConfigDefaults)
		require.Equal(t, names, slices.Sorted(maps.Keys(collection.collectors)))

		for name, c := range collection.collectors {
			assert.Equal(t, name, c.GetName(), "NewWithConfig collector %s", name)
		}
	})

	t.Run("Config fields", func(t *testing.T) {
		t.Parallel()

		var config Config

		value := reflect.ValueOf(&config).Elem()

		// Fields of size zero can share an address, so the type is part of
		// the key. Every collector has its own configuration type.
		type fieldKey struct {
			addr uintptr
			typ  reflect.Type
		}

		keyOf := func(ptr reflect.Value) fieldKey {
			return fieldKey{addr: ptr.Pointer(), typ: ptr.Type().Elem()}
		}

		fields := map[fieldKey]reflect.StructField{}
		for i := range value.NumField() {
			fields[keyOf(value.Field(i).Addr())] = value.Type().Field(i)
		}

		for _, r := range registrations {
			key := keyOf(reflect.ValueOf(r.field(&config)))

			field, ok := fields[key]
			if !assert.True(t, ok, "collector %s has no field in Config", r.name) {
				continue
			}

			configKey, _, _ := strings.Cut(field.Tag.Get("yaml"), ",")
			assert.Equal(t, r.name, configKey, "YAML key of Config.%s", field.Name)

			delete(fields, key)
		}

		for _, field := range fields {
			assert.Fail(t, "Config field without collector registration", "Config.%s", field.Name)
		}
	})

	t.Run("ConfigDefaults", func(t *testing.T) {
		t.Parallel()

		for _, r := range registrations {
			var want Config

			r.setDefaults(&want)

			assert.Equal(t, r.field(&want), r.field(&ConfigDefaults), "ConfigDefaults of collector %s", r.name)
		}
	})

	t.Run("documentation", func(t *testing.T) {
		t.Parallel()

		files, err := filepath.Glob(filepath.Join("..", "..", "docs", "collector.*.md"))
		require.NoError(t, err)

		documented := make([]string, 0, len(files))

		for _, file := range files {
			documented = append(documented, strings.TrimSuffix(strings.TrimPrefix(filepath.Base(file), "collector."), ".md"))
		}

		require.ElementsMatch(t, names, documented, "docs/collector.<name>.md files")
	})
}
