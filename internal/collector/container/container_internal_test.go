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
	"os"
	"path/filepath"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/require"
)

func newTestCollector(stateDir string) *Collector {
	return &Collector{
		config:              Config{ContainerDStateDir: stateDir},
		logger:              slog.New(slog.DiscardHandler),
		annotationsCacheHCS: map[string]containerInfo{},
		annotationsCacheJob: map[string]jobContainer{},
	}
}

func writeBundle(t *testing.T, stateDir, containerID, config string) {
	t.Helper()

	bundleDir := filepath.Join(stateDir, containerID)

	require.NoError(t, os.MkdirAll(bundleDir, 0o755))

	if config != "" {
		require.NoError(t, os.WriteFile(filepath.Join(bundleDir, "config.json"), []byte(config), 0o600))
	}
}

// The annotation cache is keyed by the raw container ID, but was pruned by the
// prefixed ID, so entries of removed containers were never deleted.
func TestCollectJobContainersPrunesAnnotationsCache(t *testing.T) {
	t.Parallel()

	c := newTestCollector(t.TempDir() + `\`)
	c.annotationsCacheJob["removed"] = jobContainer{info: containerInfo{id: "containerd://removed"}}

	ch := make(chan prometheus.Metric, 100)

	require.NoError(t, c.collectJobContainers(ch))
	require.Empty(t, c.annotationsCacheJob)
}

// Every bundle is parsed once and cached, and a bundle without config.json
// is retried later instead of being cached.
func TestCollectJobContainersCachesBundles(t *testing.T) {
	t.Parallel()

	// No trailing separator: the path used to be concatenated without one.
	stateDir := t.TempDir()

	writeBundle(t, stateDir, "hostprocess", `{"annotations":{
		"microsoft.com/hostprocess-container":"true",
		"io.kubernetes.cri.sandbox-namespace":"kube-system",
		"io.kubernetes.cri.sandbox-name":"pod",
		"io.kubernetes.cri.container-name":"container"
	}}`)
	writeBundle(t, stateDir, "process", `{"annotations":{"io.kubernetes.cri.container-name":"other"}}`)
	writeBundle(t, stateDir, "creating", "")

	c := newTestCollector(stateDir)
	ch := make(chan prometheus.Metric, 100)

	// The job objects do not exist, so no metrics are collected.
	require.NoError(t, c.collectJobContainers(ch))
	require.Equal(t, map[string]jobContainer{
		"hostprocess": {
			info: containerInfo{
				id:        "containerd://hostprocess",
				namespace: "kube-system",
				pod:       "pod",
				container: "container",
			},
			hostProcess: true,
		},
		"process": {
			info: containerInfo{
				id:        "containerd://process",
				container: "other",
			},
		},
	}, c.annotationsCacheJob)

	// config.json must be closed, or containerd cannot remove the bundle.
	require.NoError(t, os.RemoveAll(filepath.Join(stateDir, "process")))

	writeBundle(t, stateDir, "creating", `{"annotations":{}}`)

	require.NoError(t, c.collectJobContainers(ch))
	require.Contains(t, c.annotationsCacheJob, "creating")
	require.NotContains(t, c.annotationsCacheJob, "process")
}

func TestCollectJobContainersMissingStateDir(t *testing.T) {
	t.Parallel()

	c := newTestCollector(filepath.Join(t.TempDir(), "missing"))
	c.annotationsCacheJob["removed"] = jobContainer{}

	ch := make(chan prometheus.Metric, 100)

	require.NoError(t, c.collectJobContainers(ch))
	require.True(t, c.stateDirMissingLogged)
	require.Empty(t, c.annotationsCacheJob)
}
