// SPDX-License-Identifier: MIT
//
// Copyright The Prometheus Authors

//go:build windows

package utils_test

import (
	"math"
	"testing"

	"github.com/prometheus-community/windows_exporter/internal/utils"
	"github.com/stretchr/testify/require"
)

func TestCounter(t *testing.T) {
	t.Parallel()

	c := utils.NewCounter(0)
	require.Equal(t, 0.0, c.Value()) //nolint:testifylint

	c.AddValue(10)

	require.Equal(t, 10.0, c.Value()) //nolint:testifylint

	c.AddValue(50)

	require.Equal(t, 50.0, c.Value()) //nolint:testifylint

	c.AddValue(math.MaxUint32 - 10)

	require.Equal(t, float64(math.MaxUint32)-10, c.Value()) //nolint:testifylint

	c.AddValue(20)

	require.Equal(t, float64(math.MaxUint32)+21, c.Value()) //nolint:testifylint

	c.AddValue(40)

	require.Equal(t, float64(math.MaxUint32)+41, c.Value()) //nolint:testifylint

	c.AddValue(math.MaxUint32 - 10)

	require.Equal(t, float64(math.MaxUint32)*2-9, c.Value()) //nolint:testifylint
}

func TestCounter32(t *testing.T) {
	t.Parallel()

	c := utils.NewCounter32(100, 0)

	require.Equal(t, 100.0, c.Update(100, false))                                    //nolint:testifylint
	require.Equal(t, 150.0, c.Update(150, false))                                    //nolint:testifylint
	require.Equal(t, float64(math.MaxUint32)-10, c.Update(math.MaxUint32-10, false)) //nolint:testifylint
	require.Equal(t, float64(math.MaxUint32)+21, c.Update(20, false))                //nolint:testifylint
	require.Equal(t, 5.0, c.Update(5, true))                                         //nolint:testifylint
	require.Equal(t, 15.0, c.Update(15, false))                                      //nolint:testifylint
	// Without restartBelow, any decrease is a wrap.
	require.Equal(t, float64(math.MaxUint32)+4, c.Update(3, false)) //nolint:testifylint
}

func TestCounter32RestartBelow(t *testing.T) {
	t.Parallel()

	c := utils.NewCounter32(50, 1<<31)

	require.Equal(t, 3.0, c.Update(3, false))                                        //nolint:testifylint
	require.Equal(t, float64(math.MaxUint32)-10, c.Update(math.MaxUint32-10, false)) //nolint:testifylint
	require.Equal(t, float64(math.MaxUint32)+21, c.Update(20, false))                //nolint:testifylint
}
