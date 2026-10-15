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

package utils_test

import (
	"regexp"
	"testing"

	"github.com/alecthomas/kingpin/v2"
	"github.com/prometheus-community/windows_exporter/internal/types"
	"github.com/prometheus-community/windows_exporter/internal/utils"
	"github.com/stretchr/testify/require"
)

func TestRegexpFlag(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name         string
		defaultValue string
		args         []string
		want         string
		matches      []string
		notMatches   []string
		matchAll     bool
		matchNone    bool
		errorText    string
	}{
		{
			name: "include default", defaultValue: ".+", want: "^(?:.+)$",
			matches: []string{"a", "svc host"}, notMatches: []string{""}, matchAll: true,
		},
		{
			name: "exclude default", defaultValue: "", want: "^(?:)$",
			matches: []string{""}, notMatches: []string{"a"}, matchNone: true,
		},
		{
			name: "match all with star", defaultValue: ".+", args: []string{"--collector.test.include=.*"},
			want: "^(?:.*)$", matchAll: true,
		},
		{
			name: "alternation is anchored as a whole", defaultValue: ".+", args: []string{"--collector.test.include=foo|bar"},
			want: "^(?:foo|bar)$", matches: []string{"foo", "bar"}, notMatches: []string{"foobar", "xfoo", "barx"},
		},
		{
			name: "default from configuration", defaultValue: "/Microsoft/.+", want: "^(?:/Microsoft/.+)$",
			matches: []string{"/Microsoft/Windows/Task"}, notMatches: []string{"/Vendor/Task"},
		},
		{
			name: "invalid value names the flag", defaultValue: ".+", args: []string{"--collector.test.include=("},
			errorText: "collector.test.include: error parsing regexp: missing closing ): `^(?:()$`",
		},
		{
			name: "invalid default names the flag", defaultValue: "[",
			errorText: "collector.test.include: error parsing regexp",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var re *regexp.Regexp

			app := kingpin.New("test", "test")
			utils.RegexpFlag(app, "collector.test.include", "help", tc.defaultValue, &re)

			_, err := app.Parse(tc.args)
			if tc.errorText != "" {
				require.ErrorContains(t, err, tc.errorText)

				return
			}

			require.NoError(t, err)
			require.NotNil(t, re)
			require.Equal(t, tc.want, re.String())
			require.Equal(t, tc.matchAll, utils.IsMatchAll(re))
			require.Equal(t, tc.matchNone, utils.IsMatchNone(re))

			for _, name := range tc.matches {
				require.True(t, re.MatchString(name), name)
			}

			for _, name := range tc.notMatches {
				require.False(t, re.MatchString(name), name)
			}
		})
	}
}

func TestRegexpFlagValue(t *testing.T) {
	t.Parallel()

	var re *regexp.Regexp

	app := kingpin.New("test", "test")
	utils.RegexpFlag(app, "collector.test.exclude", "help", "", &re)

	_, err := app.Parse([]string{"--collector.test.exclude=a.b"})
	require.NoError(t, err)
	require.Equal(t, "a.b", app.GetFlag("collector.test.exclude").Model().Value.String())
	require.Equal(t, []string{""}, app.GetFlag("collector.test.exclude").Model().Default)
}

func TestIsMatchAllAndNone(t *testing.T) {
	t.Parallel()

	require.True(t, utils.IsMatchAll(nil))
	require.True(t, utils.IsMatchAll(types.RegExpAny))
	require.True(t, utils.IsMatchAll(utils.MustCompileAnchored(".+")))
	require.True(t, utils.IsMatchAll(utils.MustCompileAnchored(".*")))
	require.False(t, utils.IsMatchAll(utils.MustCompileAnchored("foo")))
	require.False(t, utils.IsMatchAll(types.RegExpEmpty))

	require.True(t, utils.IsMatchNone(nil))
	require.True(t, utils.IsMatchNone(types.RegExpEmpty))
	require.True(t, utils.IsMatchNone(utils.MustCompileAnchored("")))
	require.False(t, utils.IsMatchNone(utils.MustCompileAnchored("foo")))
	require.False(t, utils.IsMatchNone(types.RegExpAny))
}

func TestCompileAnchored(t *testing.T) {
	t.Parallel()

	re, err := utils.CompileAnchored("svc")
	require.NoError(t, err)
	require.True(t, re.MatchString("svc"))
	require.False(t, re.MatchString("svchost"))

	_, err = utils.CompileAnchored("(")
	require.Error(t, err)

	require.Panics(t, func() { utils.MustCompileAnchored("(") })
}
