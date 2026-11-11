// SPDX-License-Identifier: MIT
//
// Copyright The Prometheus Authors

//go:build windows

package performancecounter_test

import (
	"testing"

	"github.com/alecthomas/kingpin/v2"
	"github.com/prometheus-community/windows_exporter/internal/collector/performancecounter"
	"github.com/stretchr/testify/require"
)

// The flag value must be as strict as the configuration file block: unknown
// keys fail instead of being dropped silently.
func TestNewWithFlagsRejectsUnknownKeys(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name    string
		value   string
		wantErr bool
	}{
		{name: "valid", value: `- name: memory
  object: Memory
  counters:
    - name: Available Bytes`},
		{name: "comments only", value: "# nothing configured"},
		{name: "unknown key", value: `- name: memory
  object: Memory
  invalid: x
  counters:
    - name: Available Bytes`, wantErr: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			app := kingpin.New("test", "")
			performancecounter.NewWithFlags(app)

			// The = form is required, because kingpin reads a separate value
			// starting with "-", like a YAML list, as another flag.
			_, err := app.Parse([]string{"--collector.performancecounter.objects=" + tc.value})
			if tc.wantErr {
				require.ErrorContains(t, err, "invalid")

				return
			}

			require.NoError(t, err)
		})
	}
}
