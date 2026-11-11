// SPDX-License-Identifier: MIT
//
// Copyright The Prometheus Authors

//go:build windows

package registry_test

import (
	"testing"

	"github.com/alecthomas/kingpin/v2"
	"github.com/prometheus-community/windows_exporter/internal/collector/registry"
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
		{name: "valid", value: `- name: crash
  key: HKLM\SYSTEM\CurrentControlSet\Control\CrashControl
  values:
    - name: AutoReboot`},
		{name: "comments only", value: "# nothing configured"},
		{name: "unknown key", value: `- name: crash
  key: HKLM\SYSTEM\CurrentControlSet\Control\CrashControl
  invalid: x
  values:
    - name: AutoReboot`, wantErr: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			app := kingpin.New("test", "")
			registry.NewWithFlags(app)

			// The = form is required, because kingpin reads a separate value
			// starting with "-", like a YAML list, as another flag.
			_, err := app.Parse([]string{"--collector.registry.keys=" + tc.value})
			if tc.wantErr {
				require.ErrorContains(t, err, "invalid")

				return
			}

			require.NoError(t, err)
		})
	}
}
