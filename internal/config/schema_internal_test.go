// SPDX-License-Identifier: MIT
//
// Copyright The Prometheus Authors

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
