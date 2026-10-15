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

package config

import (
	"fmt"
	"reflect"
	"slices"
	"strings"

	"github.com/alecthomas/kingpin/v2"
	"go.yaml.in/yaml/v3"
)

// validateKeys checks that node, the value of the dotted key path, only uses
// the keys of the schema typ.
//
// The configuration file sets flags, and the flags parse and validate the
// values (see [checkFlagValue]). So the schema only defines the keys: a leaf
// takes any scalar or a list of scalars, whatever the type of its field.
// Decoding the values into the field types would reject values that the flags
// accept, like compress: "true" or a comma-separated list for a list field.
//
// A struct that decodes itself, like the performancecounter block, is decoded
// to validate its contents.
func validateKeys(node *yaml.Node, path string, typ reflect.Type) error {
	switch node.Kind {
	case yaml.DocumentNode:
		for _, content := range node.Content {
			if err := validateKeys(content, path, typ); err != nil {
				return err
			}
		}

		return nil
	case yaml.AliasNode:
		return validateKeys(node.Alias, path, typ)
	case yaml.ScalarNode:
		if node.Tag == "!!null" {
			// A key without a value leaves the flags unchanged.
			return nil
		}
	case yaml.MappingNode, yaml.SequenceNode:
	}

	if reflect.PointerTo(typ).Implements(reflect.TypeFor[yaml.Unmarshaler]()) {
		//nolint:wrapcheck // the decoder errors name the line and field.
		return node.Decode(reflect.New(typ).Interface())
	}

	if typ.Kind() != reflect.Struct {
		if !isValue(node) {
			return fmt.Errorf("line %d: %s must be a value or a list of values", node.Line, displayPath(path))
		}

		return nil
	}

	if node.Kind != yaml.MappingNode {
		return fmt.Errorf("line %d: %s must be a mapping", node.Line, displayPath(path))
	}

	for i := 0; i+1 < len(node.Content); i += 2 {
		key, value := node.Content[i], node.Content[i+1]

		if isMerge(key) {
			// A merge key (<<) adds the keys of other mappings.
			for _, source := range mergeSources(value) {
				if err := validateKeys(source, path, typ); err != nil {
					return err
				}
			}

			continue
		}

		field, ok := schemaField(typ, key.Value)
		if !ok {
			return fmt.Errorf("line %d: field %s not found in %s", key.Line, key.Value, displayPath(path))
		}

		if err := validateKeys(value, joinPath(path, key.Value), field.Type); err != nil {
			return err
		}
	}

	return nil
}

// isValue reports whether node is the value of a flag: a scalar, or a list of
// scalars for a repeatable or list flag.
func isValue(node *yaml.Node) bool {
	switch node.Kind {
	case yaml.ScalarNode:
		return true
	case yaml.AliasNode:
		return isValue(node.Alias)
	case yaml.SequenceNode:
		for _, item := range node.Content {
			if item.Kind == yaml.AliasNode {
				item = item.Alias
			}

			if item.Kind != yaml.ScalarNode {
				return false
			}
		}

		return true
	case yaml.DocumentNode, yaml.MappingNode:
	}

	return false
}

// keyLines records the line of every dotted key path below node in lines.
// Explicit keys take precedence over the keys of merged mappings, like in the
// decoded values.
func keyLines(node *yaml.Node, path string, lines map[string]int) {
	switch node.Kind {
	case yaml.DocumentNode:
		for _, content := range node.Content {
			keyLines(content, path, lines)
		}
	case yaml.AliasNode:
		keyLines(node.Alias, path, lines)
	case yaml.MappingNode:
		for i := 0; i+1 < len(node.Content); i += 2 {
			if isMerge(node.Content[i]) {
				for _, source := range mergeSources(node.Content[i+1]) {
					keyLines(source, path, lines)
				}
			}
		}

		for i := 0; i+1 < len(node.Content); i += 2 {
			key, value := node.Content[i], node.Content[i+1]
			if isMerge(key) {
				continue
			}

			keyPath := joinPath(path, key.Value)
			lines[keyPath] = key.Line
			keyLines(value, keyPath, lines)
		}
	case yaml.ScalarNode, yaml.SequenceNode:
	}
}

func isMerge(key *yaml.Node) bool {
	return key.Kind == yaml.ScalarNode && key.Tag == "!!merge"
}

// mergeSources returns the mappings that a merge key merges: an alias, a
// mapping or a list of them.
func mergeSources(value *yaml.Node) []*yaml.Node {
	if value.Kind == yaml.SequenceNode {
		return value.Content
	}

	return []*yaml.Node{value}
}

func joinPath(path, key string) string {
	if path == "" {
		return key
	}

	return path + "." + key
}

func displayPath(path string) string {
	if path == "" {
		return "the configuration"
	}

	return path
}

// schemaField returns the field of the struct typ with the YAML key name,
// looking into the fields inlined with the yaml inline option.
func schemaField(typ reflect.Type, name string) (reflect.StructField, bool) {
	for field := range typ.Fields() {
		if !field.IsExported() {
			continue
		}

		key, options, _ := strings.Cut(field.Tag.Get("yaml"), ",")

		if slices.Contains(strings.Split(options, ","), "inline") && field.Type.Kind() == reflect.Struct {
			if inlined, ok := schemaField(field.Type, name); ok {
				return inlined, true
			}

			continue
		}

		if key == "" {
			key = strings.ToLower(field.Name)
		}

		if key != "-" && key == name {
			return field, true
		}
	}

	return reflect.StructField{}, false
}

// checkFlagValue checks that the flag accepts the values from the
// configuration file.
//
// The flags only parse their default values in kingpin.Application.Parse,
// which does not tell the configuration key, and does not parse them at all
// if the flag is also set on the command line. kingpin cannot copy a flag
// value, so the check sets the flag value itself. That is safe: Parse sets it
// again, to the same default or to the command-line value. A repeatable flag
// would append the values twice, so its list is restored after the check.
func checkFlagValue(flag *kingpin.FlagClause, values []string) error {
	value := flag.Model().Value

	if repeatable, ok := value.(interface{ IsCumulative() bool }); !ok || !repeatable.IsCumulative() {
		//nolint:wrapcheck // the caller adds the key and line.
		return value.Set(strings.Join(values, ","))
	}

	getter, ok := value.(kingpin.Getter)
	if !ok {
		return nil
	}

	list := reflect.ValueOf(getter.Get())
	if list.Kind() != reflect.Pointer || list.Elem().Kind() != reflect.Slice {
		// The list cannot be restored; Parse checks the values.
		return nil
	}

	saved := reflect.ValueOf(list.Elem().Interface())
	defer list.Elem().Set(saved)

	for _, v := range values {
		if err := value.Set(v); err != nil {
			//nolint:wrapcheck // the caller adds the key and line.
			return err
		}
	}

	return nil
}
