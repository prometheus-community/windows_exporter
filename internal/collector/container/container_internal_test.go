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
	"errors"
	"fmt"
	"log/slog"
	"testing"

	"github.com/prometheus-community/windows_exporter/internal/headers/cri"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/require"
	"golang.org/x/sys/windows"
)

func TestNewKubernetesContainers(t *testing.T) {
	t.Parallel()

	k := newKubernetesContainers("containerd",
		[]cri.PodSandbox{
			{ID: "sandbox-1", Name: "pod", Namespace: "default"},
			{ID: "sandbox-2", Name: "stopped", Namespace: "kube-system", State: cri.SandboxNotReady},
		},
		[]cri.Container{
			{ID: "abc", PodSandboxID: "sandbox-1", Name: "nanoserver", State: cri.ContainerRunning},
			{ID: "orphan", PodSandboxID: "unknown", Name: "orphan", State: cri.ContainerRunning},
			{ID: "exited", PodSandboxID: "sandbox-1", Name: "init", State: cri.ContainerExited},
		},
	)

	require.Equal(t, map[string]containerInfo{
		"abc": {
			id:        "containerd://abc",
			namespace: "default",
			pod:       "pod",
			container: "nanoserver",
		},
		"orphan": {
			id:        "containerd://orphan",
			container: "orphan",
		},
	}, k.containers)
	require.Equal(t, map[string]struct{}{"sandbox-1": {}, "sandbox-2": {}}, k.sandboxes)
}

func TestHandleCRIError(t *testing.T) {
	t.Parallel()

	for name, tc := range map[string]struct {
		err         error
		unavailable bool
	}{
		"pipe not found": {
			err:         fmt.Errorf("dial: %w", windows.ERROR_FILE_NOT_FOUND),
			unavailable: true,
		},
		"CRI plugin disabled": {
			err:         fmt.Errorf("CRI Version: %w", &cri.StatusError{Code: grpcUnimplemented}),
			unavailable: true,
		},
		"other gRPC status": {
			err: fmt.Errorf("CRI Version: %w", &cri.StatusError{Code: 14}),
		},
		"other error": {
			err: errors.New("timeout"),
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			c := &Collector{
				config:         Config{CRIEndpoint: ConfigDefaults.CRIEndpoint},
				logger:         slog.New(slog.DiscardHandler),
				criRuntimeName: "containerd",
			}

			err := c.handleCRIError(tc.err)
			require.Empty(t, c.criRuntimeName)
			require.Equal(t, tc.unavailable, c.criUnavailableLogged)

			if tc.unavailable {
				require.NoError(t, err)
			} else {
				require.ErrorIs(t, err, tc.err)
			}
		})
	}
}

// A missing CRI endpoint is expected on hosts without Kubernetes.
func TestGetKubernetesContainersWithoutCRI(t *testing.T) {
	t.Parallel()

	c := New(&Config{
		CollectorsEnabled: []string{subCollectorHostprocess},
		CRIEndpoint:       "npipe:////./pipe/windows_exporter-container-test-does-not-exist",
	})
	require.NoError(t, c.Build(slog.New(slog.DiscardHandler), nil))

	t.Cleanup(func() { _ = c.Close() })

	for range 2 {
		k, err := c.getKubernetesContainers(t.Context())
		require.NoError(t, err)
		require.Empty(t, k.containers)
		require.True(t, c.criUnavailableLogged)
	}
}

// Containers without a job object are not job containers and are skipped.
func TestCollectJobContainersSkipsHCSContainers(t *testing.T) {
	t.Parallel()

	c := &Collector{logger: slog.New(slog.DiscardHandler)}
	ch := make(chan prometheus.Metric, 100)

	require.NoError(t, c.collectJobContainers(ch, kubernetesContainers{
		containers: map[string]containerInfo{
			"windows-exporter-test-no-job-object": {id: "containerd://windows-exporter-test-no-job-object"},
		},
	}))
	require.Empty(t, ch)
}

func TestBuildInvalidCRIEndpoint(t *testing.T) {
	t.Parallel()

	c := New(&Config{
		CollectorsEnabled: []string{subCollectorHostprocess},
		CRIEndpoint:       "unix:///run/containerd/containerd.sock",
	})
	require.Error(t, c.Build(slog.New(slog.DiscardHandler), nil))
}
