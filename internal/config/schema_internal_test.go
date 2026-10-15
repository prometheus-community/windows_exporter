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
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
	"go.yaml.in/yaml/v3"
)

// TestValidateKeysInline verifies that the keys of a field inlined with the
// yaml inline option belong to the enclosing mapping.
func TestValidateKeysInline(t *testing.T) {
	t.Parallel()

	type Inner struct {
		Level string `yaml:"level"`
	}

	type schema struct {
		Log struct {
			Inner `yaml:",inline"`

			Format string `yaml:"format"`
		} `yaml:"log"`
	}

	for _, tc := range []struct {
		name      string
		contents  string
		errorText string
	}{
		{name: "inlined key", contents: "log:\n  level: debug\n  format: json\n"},
		{name: "unknown key", contents: "log:\n  inner: debug\n", errorText: "line 2: field inner not found in log"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var document yaml.Node
			require.NoError(t, yaml.Unmarshal([]byte(tc.contents), &document))

			err := validateKeys(&document, "", reflect.TypeFor[schema]())
			if tc.errorText != "" {
				require.EqualError(t, err, tc.errorText)

				return
			}

			require.NoError(t, err)
		})
	}
}
