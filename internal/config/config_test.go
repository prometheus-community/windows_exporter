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
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/alecthomas/kingpin/v2"
	"github.com/prometheus-community/windows_exporter/internal/config"
	"github.com/prometheus-community/windows_exporter/internal/log"
	logflag "github.com/prometheus-community/windows_exporter/internal/log/flag"
	"github.com/prometheus-community/windows_exporter/pkg/collector"
	webflag "github.com/prometheus/exporter-toolkit/web/kingpinflag"
	"github.com/stretchr/testify/assert"
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
		{name: "empty value", contents: "log:\n  file:\n"},
		{name: "objects as JSON", contents: "collector:\n  performancecounter:\n    objects: '[{\"name\": \"memory\", \"object\": \"Memory\"}]'\n"},
		{name: "objects without value", contents: "collector:\n  performancecounter:\n    objects:\n"},
		{
			name:      "unknown performancecounter option",
			contents:  "collector:\n  performancecounter:\n    invalid: |-\n      - name: memory\n        object: Memory\n",
			errorText: "line 3: field invalid not found",
		},
		{
			name:      "unknown performancecounter object field",
			contents:  "collector:\n  performancecounter:\n    objects: |-\n      - name: memory\n        invalid: Memory\n",
			errorText: "field invalid not found in type performancecounter.Object",
		},
		{
			name:      "performancecounter objects as list",
			contents:  "collector:\n  performancecounter:\n    objects:\n      - name: memory\n        object: Memory\n",
			errorText: "line 4: objects must be a string",
		},
		{
			name:      "unknown registry option",
			contents:  "collector:\n  registry:\n    key: |-\n      - name: test\n",
			errorText: "line 3: field key not found",
		},
		{
			name:      "unknown registry value field",
			contents:  "collector:\n  registry:\n    keys: |-\n      - name: test\n        key: HKLM\\SOFTWARE\n        values:\n          - name: test\n            metrc: test\n",
			errorText: "field metrc not found in type registry.Value",
		},
		{
			name:      "unknown wmi option",
			contents:  "collector:\n  wmi:\n    query: |-\n      - name: test\n",
			errorText: "line 3: field query not found",
		},
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

// TestNewConfigFileResolverFlags verifies that every flag can be set from the
// configuration file: the nested key derived from the flag name must exist in
// the configuration file schema, pass the strict validation, and be bound back
// to the same flag.
func TestNewConfigFileResolverFlags(t *testing.T) {
	t.Parallel()

	mainFlags := mainFlagNames(t)

	// kingpin's built-in flags, like --help, are not configuration options.
	builtin := kingpin.New("test", "test")

	for _, flag := range newApp(mainFlags).Model().Flags {
		switch {
		case builtin.GetFlag(flag.Name) != nil,
			// selects the configuration file itself
			flag.Name == "config.file",
			// prints the available sub-collectors and exits
			flag.Name == "collector.exchange.list":
			continue
		}

		t.Run(flag.Name, func(t *testing.T) {
			t.Parallel()

			keys := strings.Split(flag.Name, ".")

			field, block, ok := configFieldType(reflect.TypeFor[config.ConfigFile](), keys)
			require.True(t, ok, "the configuration file has no key %s", flag.Name)

			value, values := sampleConfigValue(t, field, block)

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

			app := newApp(mainFlags)
			require.NoError(t, resolver.Bind(app, nil))

			want := values
			if repeatable, ok := flag.Value.(interface{ IsCumulative() bool }); !ok || !repeatable.IsCumulative() {
				want = []string{strings.Join(values, ",")}
			}

			require.Equal(t, want, app.GetFlag(flag.Name).Model().Default, "config:\n%s", data)
		})
	}
}

// TestConfigFileKeys verifies the reverse: every key of the configuration file
// schema sets a flag, so no key passes the validation only to be ignored.
func TestConfigFileKeys(t *testing.T) {
	t.Parallel()

	app := newApp(mainFlagNames(t))

	for _, key := range configKeys(reflect.TypeFor[config.ConfigFile](), "") {
		assert.NotNil(t, app.GetFlag(key), "the configuration file key %s sets no flag", key)
	}
}

// TestNewConfigFileResolverExamples verifies that the example configuration
// files, and the configuration examples in the documentation of the collectors
// that take their list as a string, load and parse.
func TestNewConfigFileResolverExamples(t *testing.T) {
	t.Parallel()

	examples := map[string]string{}

	for _, path := range []string{"../../config.yaml", "../../docs/example_config.yml"} {
		data, err := os.ReadFile(path)
		require.NoError(t, err)

		examples[path] = string(data)
	}

	codeBlock := regexp.MustCompile("(?ms)^```(?:yaml|json)\n(.*?)^```")

	for _, doc := range []struct{ path, block, key string }{
		{"../../docs/collector.performancecounter.md", "performancecounter", "objects"},
		{"../../docs/collector.registry.md", "registry", "keys"},
		{"../../docs/collector.wmi.md", "wmi", "queries"},
	} {
		data, err := os.ReadFile(doc.path)
		require.NoError(t, err)

		for i, match := range codeBlock.FindAllStringSubmatch(strings.ReplaceAll(string(data), "\r\n", "\n"), -1) {
			example := match[1]

			switch {
			case strings.HasPrefix(example, "collector:"):
				// a configuration file
			case strings.HasPrefix(example, "- "), strings.HasPrefix(example, "["):
				// a schema example, which is the value of the flag
				example = fmt.Sprintf("collector:\n  %s:\n    %s: |-\n%s", doc.block, doc.key, indent(example, "      "))
			default:
				// a fragment or an alerting rule
				continue
			}

			examples[fmt.Sprintf("%s block %d", doc.path, i+1)] = example
		}
	}

	mainFlags := mainFlagNames(t)

	for name, contents := range examples {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			path := filepath.Join(t.TempDir(), "config.yaml")
			require.NoError(t, os.WriteFile(path, []byte(contents), 0o600))

			resolver, err := config.NewConfigFileResolver(path)
			require.NoError(t, err, "config:\n%s", contents)

			app := newApp(mainFlags)
			require.NoError(t, resolver.Bind(app, nil))

			_, err = app.Parse(nil)
			require.NoError(t, err, "config:\n%s", contents)
		})
	}
}

// newApp returns a kingpin application with the flags of windows_exporter: the
// flags declared in cmd/windows_exporter, named by mainFlags, and the flags of
// the packages it uses.
func newApp(mainFlags []string) *kingpin.Application {
	app := kingpin.New("test", "test")

	for _, name := range mainFlags {
		app.Flag(name, "").String()
	}

	logFile := &log.AllowedFile{}
	_ = logFile.Set("stdout")

	webflag.AddFlags(app, ":9182")
	logflag.AddFlags(app, &log.Config{File: logFile})
	collector.NewWithFlags(app)

	return app
}

// mainFlagNames returns the names of the flags declared in cmd/windows_exporter.
// It is a main package, so the names are read from its source.
func mainFlagNames(t *testing.T) []string {
	t.Helper()

	files, err := filepath.Glob(filepath.Join("..", "..", "cmd", "windows_exporter", "*.go"))
	require.NoError(t, err)

	var names []string

	fset := token.NewFileSet()

	for _, path := range files {
		if strings.HasSuffix(path, "_test.go") {
			continue
		}

		file, err := parser.ParseFile(fset, path, nil, parser.SkipObjectResolution)
		require.NoError(t, err)

		ast.Inspect(file, func(node ast.Node) bool {
			call, ok := node.(*ast.CallExpr)
			if !ok || len(call.Args) == 0 {
				return true
			}

			if fun, ok := call.Fun.(*ast.SelectorExpr); !ok || fun.Sel.Name != "Flag" {
				return true
			}

			lit, ok := call.Args[0].(*ast.BasicLit)
			require.True(t, ok && lit.Kind == token.STRING, "%s: the flag name is not a string literal", fset.Position(call.Pos()))

			name, err := strconv.Unquote(lit.Value)
			require.NoError(t, err)

			names = append(names, name)

			return true
		})
	}

	require.Contains(t, names, "collectors.enabled", "no flags found in cmd/windows_exporter")

	return names
}

// configFieldType returns the type of the struct field reached by following the
// yaml keys, starting at typ. block reports whether the field belongs to a block
// with its own decoding, which takes the value as a string.
func configFieldType(typ reflect.Type, keys []string) (reflect.Type, bool, bool) {
	if len(keys) == 0 {
		return typ, false, true
	}

	if typ.Kind() != reflect.Struct {
		return nil, false, false
	}

	for field := range typ.Fields() {
		if yamlName(field) == keys[0] {
			fieldType, block, ok := configFieldType(field.Type, keys[1:])

			return fieldType, block || reflect.PointerTo(typ).Implements(reflect.TypeFor[yaml.Unmarshaler]()), ok
		}
	}

	return nil, false, false
}

// configKeys returns the dotted keys of the leaves of the configuration file
// schema typ.
func configKeys(typ reflect.Type, prefix string) []string {
	var keys []string

	for field := range typ.Fields() {
		key := yamlName(field)
		if key == "" {
			continue
		}

		if prefix != "" {
			key = prefix + "." + key
		}

		if field.Type.Kind() == reflect.Struct {
			keys = append(keys, configKeys(field.Type, key)...)
		} else {
			keys = append(keys, key)
		}
	}

	return keys
}

// yamlName returns the key of field in YAML, or "" if it has none.
func yamlName(field reflect.StructField) string {
	if !field.IsExported() {
		return ""
	}

	name, _, _ := strings.Cut(field.Tag.Get("yaml"), ",")

	switch name {
	case "-":
		return ""
	case "":
		return strings.ToLower(field.Name)
	default:
		return name
	}
}

// sampleConfigValue returns a configuration file value that decodes into typ,
// together with the values the resolver flattens it to. A field of a block with
// its own decoding takes the value as a string.
func sampleConfigValue(t *testing.T, typ reflect.Type, block bool) (any, []string) {
	t.Helper()

	switch {
	case block:
		return "[]", []string{"[]"}
	case typ == reflect.TypeFor[time.Duration]():
		return "1m", []string{"1m"}
	case typ == reflect.TypeFor[*regexp.Regexp]():
		return "test", []string{"test"}
	}

	switch typ.Kind() {
	case reflect.Bool:
		return true, []string{"true"}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return 1, []string{"1"}
	case reflect.Slice, reflect.Interface:
		return []string{"test1", "test2"}, []string{"test1", "test2"}
	case reflect.String:
		return "test", []string{"test"}
	default:
		t.Fatalf("no sample configuration value for type %s", typ)

		return nil, nil
	}
}

// indent prefixes every non-empty line of s.
func indent(s, prefix string) string {
	lines := strings.Split(strings.TrimRight(s, "\n"), "\n")
	for i, line := range lines {
		if line != "" {
			lines[i] = prefix + line
		}
	}

	return strings.Join(lines, "\n") + "\n"
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
