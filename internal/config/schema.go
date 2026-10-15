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
	"strings"

	"go.yaml.in/yaml/v3"
)

// validateKeys checks that node, the value of the key name, only uses the keys
// of the schema typ.
//
// The configuration file sets flags, and the flags parse and validate the
// values. So the schema only defines the keys: a leaf takes any scalar or a
// list of scalars, whatever the type of its field. Decoding the values into
// the field types would reject values that the flags accept, like
// compress: "true" or a comma-separated list for a list field.
//
// A struct that decodes itself, like the performancecounter block, is decoded
// to validate its contents.
func validateKeys(node *yaml.Node, name string, typ reflect.Type) error {
	switch node.Kind {
	case yaml.DocumentNode:
		for _, content := range node.Content {
			if err := validateKeys(content, name, typ); err != nil {
				return err
			}
		}

		return nil
	case yaml.AliasNode:
		return validateKeys(node.Alias, name, typ)
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
		return validateValue(node)
	}

	if node.Kind != yaml.MappingNode {
		return fmt.Errorf("line %d: %s must be a mapping", node.Line, name)
	}

	for i := 0; i+1 < len(node.Content); i += 2 {
		key, value := node.Content[i], node.Content[i+1]

		field, ok := schemaField(typ, key.Value)
		if !ok {
			return fmt.Errorf("line %d: field %s not found in type %s", key.Line, key.Value, typ)
		}

		if err := validateKeys(value, key.Value, field.Type); err != nil {
			return err
		}
	}

	return nil
}

// validateValue checks that node is the value of a flag: a scalar, or a list
// of scalars for a repeatable or list flag.
func validateValue(node *yaml.Node) error {
	switch node.Kind {
	case yaml.ScalarNode:
		return nil
	case yaml.AliasNode:
		return validateValue(node.Alias)
	case yaml.SequenceNode:
		for _, item := range node.Content {
			if item.Kind == yaml.AliasNode {
				item = item.Alias
			}

			if item.Kind != yaml.ScalarNode {
				return fmt.Errorf("line %d: expected a value in the list", item.Line)
			}
		}

		return nil
	case yaml.DocumentNode, yaml.MappingNode:
	}

	return fmt.Errorf("line %d: expected a value or a list of values", node.Line)
}

// schemaField returns the field of the struct typ with the YAML key name.
func schemaField(typ reflect.Type, name string) (reflect.StructField, bool) {
	for field := range typ.Fields() {
		if !field.IsExported() {
			continue
		}

		key, _, _ := strings.Cut(field.Tag.Get("yaml"), ",")
		if key == "" {
			key = strings.ToLower(field.Name)
		}

		if key != "-" && key == name {
			return field, true
		}
	}

	return reflect.StructField{}, false
}
