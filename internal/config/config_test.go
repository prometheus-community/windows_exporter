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

package config_test

import (
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/alecthomas/kingpin/v2"
	"github.com/prometheus-community/windows_exporter/internal/config"
	"github.com/prometheus-community/windows_exporter/pkg/collector"
	"github.com/stretchr/testify/require"
	"go.yaml.in/yaml/v3"
)

func TestParseConfigFile(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name string
		args []string
		want string
	}{
		{name: "absent"},
		{name: "long joined", args: []string{"--config.file=test.yaml"}, want: "test.yaml"},
		{name: "short joined", args: []string{"-config.file=test.yaml"}, want: "test.yaml"},
		{name: "long separate", args: []string{"--config.file", "test.yaml"}, want: "test.yaml"},
		{name: "short separate", args: []string{"-config.file", "test.yaml"}, want: "test.yaml"},
		{name: "after other flags", args: []string{"--log.level=debug", "--config.file", "test.yaml"}, want: "test.yaml"},
		{name: "missing value", args: []string{"--config.file"}},
		{name: "empty value", args: []string{"--config.file="}},
		{name: "unrelated suffix", args: []string{"--other-config.file", "test.yaml"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			require.Equal(t, tc.want, config.ParseConfigFile(tc.args))
		})
	}
}

func TestNewConfigFileResolver(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name      string
		contents  string
		errorText string
	}{
		{name: "empty"},
		{name: "valid", contents: "log:\n  level: debug\nweb:\n  listen-address: ['127.0.0.1:9182', '[::1]:9182']\n"},
		{name: "unknown key", contents: "unknown: value", errorText: "configuration file validation error"},
		{name: "unknown collector option", contents: "collector:\n  cpu:\n    unknown: true\n", errorText: "configuration file validation error"},
		{name: "malformed YAML", contents: "log: [", errorText: "configuration file validation error"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			path := filepath.Join(t.TempDir(), "config.yaml")
			require.NoError(t, os.WriteFile(path, []byte(tc.contents), 0o600))

			resolver, err := config.NewConfigFileResolver(path)
			if tc.errorText != "" {
				require.ErrorContains(t, err, tc.errorText)

				return
			}

			require.NoError(t, err)
			require.NotNil(t, resolver)
		})
	}

	t.Run("missing file", func(t *testing.T) {
		t.Parallel()
		_, err := config.NewConfigFileResolver(filepath.Join(t.TempDir(), "missing.yaml"))
		require.ErrorContains(t, err, "failed to open configuration file")
	})
}

// TestNewConfigFileResolverCollectorFlags verifies that every collector flag can
// be set from the configuration file: the nested key derived from the flag name
// must exist in collector.Config, pass the strict validation, and be bound back
// to the same flag.
func TestNewConfigFileResolverCollectorFlags(t *testing.T) {
	t.Parallel()

	app := kingpin.New("test", "test")
	collector.NewWithFlags(app)

	for _, flag := range app.Model().Flags {
		// collector.exchange.list prints the available sub-collectors and exits;
		// it is not a configuration option.
		if !strings.HasPrefix(flag.Name, "collector.") || flag.Name == "collector.exchange.list" {
			continue
		}

		t.Run(flag.Name, func(t *testing.T) {
			t.Parallel()

			keys := strings.Split(flag.Name, ".")

			field, ok := configFieldType(reflect.TypeFor[collector.Config](), keys[1:])
			require.True(t, ok, "collector.Config has no field with the yaml path %s", flag.Name)

			value, want := sampleConfigValue(t, field)

			contents := value
			for _, key := range slices.Backward(keys) {
				contents = map[string]any{key: contents}
			}

			data, err := yaml.Marshal(contents)
			require.NoError(t, err)

			path := filepath.Join(t.TempDir(), "config.yaml")
			require.NoError(t, os.WriteFile(path, data, 0o600))

			resolver, err := config.NewConfigFileResolver(path)
			require.NoError(t, err, "config:\n%s", data)

			bindApp := kingpin.New("test", "test")
			collector.NewWithFlags(bindApp)

			require.NoError(t, resolver.Bind(bindApp, nil))
			require.Equal(t, []string{want}, bindApp.GetFlag(flag.Name).Model().Default)
		})
	}
}

// configFieldType returns the type of the struct field reached by following
// the yaml tags in keys, starting at typ.
func configFieldType(typ reflect.Type, keys []string) (reflect.Type, bool) {
	if len(keys) == 0 {
		return typ, true
	}

	if typ.Kind() != reflect.Struct {
		return nil, false
	}

	for field := range typ.Fields() {
		if name, _, _ := strings.Cut(field.Tag.Get("yaml"), ","); name == keys[0] {
			return configFieldType(field.Type, keys[1:])
		}
	}

	return nil, false
}

// sampleConfigValue returns a configuration file value that decodes into typ,
// together with the flag value the resolver flattens it to.
func sampleConfigValue(t *testing.T, typ reflect.Type) (any, string) {
	t.Helper()

	switch {
	case typ == reflect.TypeFor[time.Duration]():
		return "1m", "1m"
	case typ == reflect.TypeFor[*regexp.Regexp]():
		return "test", "test"
	}

	switch typ.Kind() {
	case reflect.Bool:
		return true, "true"
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return 1, "1"
	case reflect.Slice:
		return []string{"test"}, "test"
	case reflect.String:
		return "test", "test"
	default:
		t.Fatalf("no sample configuration value for type %s", typ)

		return nil, ""
	}
}

func TestParse(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name      string
		extraArgs []string
		want      string
		errorText string
	}{
		{name: "config defaults", want: "debug"},
		{name: "CLI overrides config", extraArgs: []string{"--log.level=error"}, want: "error"},
		{name: "invalid CLI flag", extraArgs: []string{"--unknown"}, errorText: "failed to bind configuration"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			path := filepath.Join(t.TempDir(), "config.yaml")
			require.NoError(t, os.WriteFile(path, []byte("log:\n  level: debug\n"), 0o600))

			app := kingpin.New("test", "test")
			app.Flag("config.file", "config").String()
			level := app.Flag("log.level", "level").Default("info").String()

			args := append([]string{"--config.file=" + path}, tc.extraArgs...)

			err := config.Parse(app, args)
			if tc.errorText != "" {
				require.ErrorContains(t, err, tc.errorText)

				return
			}

			require.NoError(t, err)
			require.Equal(t, tc.want, *level)
		})
	}
}
