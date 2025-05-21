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

package hcs_test

import (
	"testing"

	"github.com/prometheus-community/windows_exporter/internal/headers/hcs"
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
