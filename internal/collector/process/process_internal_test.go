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

package process

import (
	"log/slog"
	"os/user"
	"testing"

	"github.com/alecthomas/kingpin/v2"
	"github.com/stretchr/testify/require"
	"golang.org/x/sys/windows"
)

// The flags store anchored expressions in the configuration of the collector.
func TestNewWithFlagsFilters(t *testing.T) {
	t.Parallel()

	app := kingpin.New("test", "test")
	c := NewWithFlags(app)

	_, err := app.Parse([]string{"--collector.process.include=foo|bar"})
	require.NoError(t, err)
	require.Equal(t, "^(?:foo|bar)$", c.config.ProcessInclude.String())
	require.Equal(t, "^(?:)$", c.config.ProcessExclude.String())
	require.True(t, c.config.ProcessInclude.MatchString("bar"))
	require.False(t, c.config.ProcessInclude.MatchString("foobar"))

	app = kingpin.New("test", "test")
	NewWithFlags(app)

	_, err = app.Parse([]string{"--collector.process.exclude=("})
	require.ErrorContains(t, err, "collector.process.exclude: error parsing regexp")
}

// The owner label must use the Windows DOMAIN\user form, as os/user reports it.
func TestGetProcessOwner(t *testing.T) {
	t.Parallel()

	currentUser, err := user.Current()
	require.NoError(t, err)

	c := &Collector{}

	owner, err := c.getProcessOwner(slog.New(slog.DiscardHandler), windows.CurrentProcess())
	require.NoError(t, err)
	require.Equal(t, currentUser.Username, owner)
}
