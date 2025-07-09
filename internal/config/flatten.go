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
	"strings"
)

// convertMap converts a map with any comparable key type to a map with string keys.
func convertMap[K comparable, V any](originalMap map[K]V) map[string]V {
	convertedMap := make(map[string]V, len(originalMap))
	for key, value := range originalMap {
		if keyString, ok := any(key).(string); ok {
			convertedMap[keyString] = value
		}
	}

	return convertedMap
}

// flatten flattens a nested map, joining keys with dots.
// e.g. {"a": {"b":"c"}} => {"a.b":"c"}
func flatten(data map[string]any) map[string]string {
	result := make(map[string]string)

	flattenHelper("", data, result)

	return result
}

func flattenHelper(prefix string, data map[string]any, result map[string]string) {
	for k, v := range data {
		fullKey := k
		if prefix != "" {
			fullKey = prefix + "." + k
		}
		switch val := v.(type) {
		case map[any]any:
			flattenHelper(fullKey, convertMap(val), result)
		case map[string]any:
			flattenHelper(fullKey, val, result)
		case []any:
			strSlice := make([]string, len(val))
			for i, elem := range val {
				strSlice[i] = fmt.Sprint(elem)
			}
			result[fullKey] = strings.Join(strSlice, ",")
		default:
			result[fullKey] = fmt.Sprint(val)
		}
	}
}
