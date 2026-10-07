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
	"os"
	"path/filepath"
	"testing"

	"github.com/alecthomas/kingpin/v2"
	"github.com/prometheus-community/windows_exporter/internal/config"
	"github.com/stretchr/testify/require"
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
