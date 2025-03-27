package hcs_test

import (
	"testing"

	"github.com/prometheus-community/windows_exporter/internal/win32/hcs"
	"github.com/stretchr/testify/require"
)

func TestGetContainers(t *testing.T) {
	t.Parallel()

	containers, err := hcs.GetContainers()
	require.NoError(t, err)
	require.NotNil(t, containers)
}

func TestOpenContainer(t *testing.T) {
	t.Parallel()

	containers, err := hcs.GetContainers()
	require.NoError(t, err)

	if len(containers) == 0 {
		t.Skip("No containers found")
	}

	statistics, err := hcs.GetContainerStatistics(containers[0].ID)
	require.NoError(t, err)
	require.NotNil(t, statistics)
}

func TestOpenContainerNotFound(t *testing.T) {
	t.Parallel()

	_, err := hcs.GetContainerStatistics("f3056b79b36ddfe203376473e2aeb4922a8ca7c5d8100764e5829eb5552fe09b")
	require.ErrorIs(t, err, hcs.ErrIDNotFound)
}
