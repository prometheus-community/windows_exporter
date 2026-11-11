// SPDX-License-Identifier: MIT
//
// Copyright The Prometheus Authors

//go:build windows

package lazyproc_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/prometheus-community/windows_exporter/internal/headers/lazyproc"
	"github.com/stretchr/testify/require"
	"golang.org/x/sys/windows"
)

func TestFind(t *testing.T) {
	t.Parallel()

	kernel32 := windows.NewLazySystemDLL("kernel32.dll")
	missingDLL := windows.NewLazySystemDLL("windows_exporter_missing.dll")

	t.Run("all present", func(t *testing.T) {
		t.Parallel()

		require.NoError(t, lazyproc.Find(kernel32, kernel32.NewProc("GetTickCount64"), kernel32.NewProc("GetSystemInfo")))
		require.NoError(t, lazyproc.Find(kernel32))
	})

	t.Run("missing export", func(t *testing.T) {
		t.Parallel()

		err := lazyproc.Find(kernel32, kernel32.NewProc("GetTickCount64"), kernel32.NewProc("WindowsExporterMissingExport"))
		require.ErrorIs(t, err, errors.ErrUnsupported)
		require.ErrorContains(t, err, "WindowsExporterMissingExport")
	})

	t.Run("missing DLL is reported once", func(t *testing.T) {
		t.Parallel()

		err := lazyproc.Find(missingDLL, missingDLL.NewProc("A"), missingDLL.NewProc("B"))
		require.ErrorIs(t, err, errors.ErrUnsupported)
		require.Equal(t, 1, strings.Count(err.Error(), "windows_exporter_missing.dll"), err.Error())

		// The loader error stays text only, so the error has a single cause.
		require.NotErrorIs(t, err, windows.ERROR_MOD_NOT_FOUND)

		var lazyErr *lazyproc.Error

		require.ErrorAs(t, err, &lazyErr)
		require.Equal(t, "windows_exporter_missing.dll", lazyErr.DLL)
		require.Len(t, lazyErr.Missing, 1)
	})

	t.Run("Once", func(t *testing.T) {
		t.Parallel()

		load := lazyproc.Once(kernel32, kernel32.NewProc("WindowsExporterMissingExport"))

		first := load()
		require.ErrorIs(t, first, errors.ErrUnsupported)
		// The cached error is returned again.
		require.Same(t, first, load())
	})
}
