// SPDX-License-Identifier: MIT
//
// Copyright The Prometheus Authors

//go:build windows

package yamlstring_test

import (
	"testing"

	"github.com/prometheus-community/windows_exporter/internal/config/yamlstring"
	"github.com/stretchr/testify/require"
	"go.yaml.in/yaml/v3"
)

type item struct {
	Name string `yaml:"name"`
}

type block struct {
	Items []item
}

func (b *block) UnmarshalYAML(node *yaml.Node) error {
	return yamlstring.DecodeBlock(node, "items", &b.Items)
}

func TestDecodeBlock(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name      string
		contents  string
		want      []item
		errorText string
	}{
		{name: "YAML", contents: "items: |-\n  - name: a\n  - name: b\n", want: []item{{Name: "a"}, {Name: "b"}}},
		{name: "JSON", contents: `items: '[{"name": "a"}]'`, want: []item{{Name: "a"}}},
		{name: "empty string", contents: `items: ""`},
		{name: "null", contents: "items:\n"},
		{name: "unknown key", contents: "itemz: |-\n  - name: a\n", errorText: "line 1: field itemz not found, the only field is items"},
		{name: "unknown field", contents: "items: |-\n  - invalid: a\n", errorText: "line 1: in items, line 1: field invalid not found in type yamlstring_test.item"},
		{name: "list", contents: "items:\n  - name: a\n", errorText: "line 2: items must be a string"},
		{name: "not a list", contents: "items: |-\n  name: a\n", errorText: "cannot unmarshal !!map into []yamlstring_test.item"},
		{name: "not a mapping", contents: "[items]", errorText: "line 1: cannot unmarshal !!seq into a block with the key items"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var got block

			err := yaml.Unmarshal([]byte(tc.contents), &got)
			if tc.errorText != "" {
				require.ErrorContains(t, err, tc.errorText)

				return
			}

			require.NoError(t, err)
			require.Equal(t, tc.want, got.Items)
		})
	}
}
