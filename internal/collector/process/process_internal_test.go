// SPDX-License-Identifier: MIT
//
// Copyright The Prometheus Authors

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

	// On Windows, the Uid is the SID string.
	cached, ok := c.lookupCache.Load(currentUser.Uid)
	require.True(t, ok)
	require.Equal(t, ownerLookup{owner: owner}, cached)
}
