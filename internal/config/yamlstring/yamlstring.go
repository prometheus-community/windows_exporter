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

// Package yamlstring decodes configuration file blocks whose value is a YAML or
// JSON document kept as a string, like the objects of the performancecounter
// collector.
package yamlstring

import (
	"errors"
	"fmt"
	"io"
	"strings"

	"go.yaml.in/yaml/v3"
)

// DecodeBlock decodes node, the configuration file block of a collector, into
// target. key must be the only key of the block, and its value a string holding
// a YAML or JSON list, the same value the collector's flag takes. A null value
// leaves target unchanged.
//
// Unknown keys, in the block and in the list, are rejected, like everywhere else
// in the configuration file. The errors are returned as a [yaml.TypeError], so
// the decoder reports them together with the other errors of the file.
func DecodeBlock[T any](node *yaml.Node, key string, target *[]T) error {
	if node.Kind != yaml.MappingNode {
		return &yaml.TypeError{Errors: []string{
			fmt.Sprintf("line %d: cannot unmarshal %s into a block with the key %s", node.Line, node.ShortTag(), key),
		}}
	}

	var errs []string

	for i := 0; i+1 < len(node.Content); i += 2 {
		name, value := node.Content[i], node.Content[i+1]

		if name.Value != key {
			errs = append(errs, fmt.Sprintf("line %d: field %s not found, the only field is %s", name.Line, name.Value, key))

			continue
		}

		errs = append(errs, decodeValue(value, key, target)...)
	}

	if len(errs) > 0 {
		return &yaml.TypeError{Errors: errs}
	}

	return nil
}

// decodeValue decodes the string value of key into target and returns the
// errors, if any. Line numbers inside the string count from its first line.
func decodeValue[T any](value *yaml.Node, key string, target *[]T) []string {
	if value.ShortTag() == "!!null" {
		return nil
	}

	if value.Kind != yaml.ScalarNode || value.ShortTag() != "!!str" {
		return []string{fmt.Sprintf("line %d: %s must be a string, use %s: |- followed by the list", value.Line, key, key)}
	}

	decoder := yaml.NewDecoder(strings.NewReader(value.Value))
	decoder.KnownFields(true)

	var list []T

	if err := decoder.Decode(&list); err != nil && !errors.Is(err, io.EOF) {
		var typeErr *yaml.TypeError
		if !errors.As(err, &typeErr) {
			return []string{fmt.Sprintf("line %d: in %s, %s", value.Line, key, err)}
		}

		errs := make([]string, len(typeErr.Errors))
		for i, e := range typeErr.Errors {
			errs[i] = fmt.Sprintf("line %d: in %s, %s", value.Line, key, e)
		}

		return errs
	}

	*target = list

	return nil
}
