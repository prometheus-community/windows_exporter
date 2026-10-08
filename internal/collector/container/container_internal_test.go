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

package container

import (
	"log/slog"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/require"
)

// The annotation cache is keyed by the raw container ID, but was pruned by the
// prefixed ID, so entries of removed containers were never deleted.
func TestCollectJobContainersPrunesAnnotationsCache(t *testing.T) {
	t.Parallel()

	c := &Collector{
		config: Config{ContainerDStateDir: t.TempDir() + `\`},
		logger: slog.New(slog.DiscardHandler),
		annotationsCacheJob: map[string]containerInfo{
			"removed": {id: "containerd://removed"},
		},
	}

	ch := make(chan prometheus.Metric, 100)

	require.NoError(t, c.collectJobContainers(ch))
	require.Empty(t, c.annotationsCacheJob)
}
