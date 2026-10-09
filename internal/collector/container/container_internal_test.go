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
	"maps"
	"slices"
	"testing"
	"time"

	"github.com/prometheus-community/windows_exporter/internal/headers/cri"
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
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
			{ID: "abc", PodSandboxID: "sandbox-1", Name: "nanoserver", State: cri.ContainerRunning, CreatedAt: time.Unix(1700000000, 0)},
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
			createdAt: time.Unix(1700000000, 0),
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

	jobContainers, err := c.collectJobContainers(ch, kubernetesContainers{
		containers: map[string]containerInfo{
			"windows-exporter-test-no-job-object": {id: "containerd://windows-exporter-test-no-job-object"},
		},
	})
	require.NoError(t, err)
	require.Empty(t, jobContainers)
	require.Empty(t, ch)
}

func TestSelectCRIStatsContainers(t *testing.T) {
	t.Parallel()

	kubernetes := kubernetesContainers{
		containers: map[string]containerInfo{
			"process-isolated": {},
			"hostprocess":      {},
			"hyperv":           {},
		},
	}
	hcsContainers := map[string]struct{}{"process-isolated": {}, "docker": {}}
	jobContainers := map[string]struct{}{"hostprocess": {}}

	for name, tc := range map[string]struct {
		hcsContainers map[string]struct{}
		jobContainers map[string]struct{}
		exported      []string
		hyperv        []string
	}{
		"all collectors": {
			hcsContainers: hcsContainers,
			jobContainers: jobContainers,
			exported:      []string{"process-isolated", "hostprocess", "hyperv"},
			hyperv:        []string{"hyperv"},
		},
		// Hyper-V isolated containers belong to the hcs collector.
		"hostprocess only": {
			jobContainers: jobContainers,
			exported:      []string{"hostprocess"},
		},
		// The test containers have no job object, so they are taken as Hyper-V isolated.
		"hcs only": {
			hcsContainers: hcsContainers,
			exported:      []string{"process-isolated", "hostprocess", "hyperv"},
			hyperv:        []string{"hostprocess", "hyperv"},
		},
		// Without HCS, Hyper-V isolated containers can't be told apart from process-isolated ones.
		"HCS not available": {
			jobContainers: jobContainers,
			exported:      []string{"hostprocess"},
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			exported, hyperv := selectCRIStatsContainers(kubernetes, tc.hcsContainers, tc.jobContainers)
			require.ElementsMatch(t, tc.exported, slices.Collect(maps.Keys(exported)))
			require.ElementsMatch(t, tc.hyperv, slices.Collect(maps.Keys(hyperv)))
		})
	}
}

func TestCollectCRIContainers(t *testing.T) {
	t.Parallel()

	c := New(&Config{
		CollectorsEnabled: []string{subCollectorHostprocess},
		CRIEndpoint:       "npipe:////./pipe/windows_exporter-container-test-does-not-exist",
	})
	require.NoError(t, c.Build(slog.New(slog.DiscardHandler), nil))

	t.Cleanup(func() { _ = c.Close() })

	ptr := func(v uint64) *uint64 { return &v }

	kubernetes := kubernetesContainers{
		containers: map[string]containerInfo{
			"hyperv":           {id: "containerd://hyperv", namespace: "default", pod: "pod", container: "hyperv", createdAt: time.Unix(1700000000, 0)},
			"process-isolated": {id: "containerd://process-isolated", namespace: "default", pod: "pod", container: "app"},
			"no-stats":         {id: "containerd://no-stats"},
		},
	}

	stats := []cri.ContainerStats{
		{
			ID:            "hyperv",
			CPU:           &cri.CPUUsage{Timestamp: 1, UsageCoreNanoSeconds: ptr(1500000000)},
			Memory:        &cri.MemoryUsage{Timestamp: 1, WorkingSetBytes: ptr(4096), UsageBytes: ptr(8192)},
			WritableLayer: &cri.FilesystemUsage{Timestamp: 1, UsedBytes: ptr(1024)},
		},
		{
			// HCS provides the CPU and memory usage of process-isolated containers.
			ID:            "process-isolated",
			CPU:           &cri.CPUUsage{Timestamp: 1, UsageCoreNanoSeconds: ptr(1)},
			WritableLayer: &cri.FilesystemUsage{Timestamp: 1, UsedBytes: ptr(2048)},
		},
		{
			// A container without task metrics, e.g. exited after listing, whose writable layer containerd
			// has not measured yet. Nothing is exported for it.
			ID:            "no-stats",
			WritableLayer: &cri.FilesystemUsage{UsedBytes: ptr(0)},
		},
		{
			// Containers not exported, e.g. pause containers, are skipped.
			ID:            "sandbox",
			WritableLayer: &cri.FilesystemUsage{Timestamp: 1, UsedBytes: ptr(1)},
		},
	}

	ch := make(chan prometheus.Metric, 100)

	c.collectCRIContainers(ch, stats, kubernetes,
		map[string]struct{}{"hyperv": {}, "process-isolated": {}, "no-stats": {}},
		map[string]struct{}{"hyperv": {}, "no-stats": {}},
	)
	close(ch)

	type sample struct {
		desc        *prometheus.Desc
		containerID string
		value       float64
	}

	var samples []sample

	for m := range ch {
		var metric dto.Metric

		require.NoError(t, m.Write(&metric))

		s := sample{desc: m.Desc()}

		for _, label := range metric.GetLabel() {
			if label.GetName() == "container_id" {
				s.containerID = label.GetValue()
			}
		}

		switch {
		case metric.GetGauge() != nil:
			s.value = metric.GetGauge().GetValue()
		case metric.GetCounter() != nil:
			s.value = metric.GetCounter().GetValue()
		}

		samples = append(samples, s)
	}

	require.ElementsMatch(t, []sample{
		{c.containerAvailable, "containerd://hyperv", 1},
		{c.startTime, "containerd://hyperv", 1700000000},
		{c.runtimeTotal, "containerd://hyperv", 1.5},
		{c.usagePrivateWorkingSetBytes, "containerd://hyperv", 4096},
		{c.usageCommitBytes, "containerd://hyperv", 8192},
		{c.writableLayerUsageBytes, "containerd://hyperv", 1024},
		{c.writableLayerUsageBytes, "containerd://process-isolated", 2048},
	}, samples)
}

func TestBuildInvalidCRIEndpoint(t *testing.T) {
	t.Parallel()

	c := New(&Config{
		CollectorsEnabled: []string{subCollectorHostprocess},
		CRIEndpoint:       "unix:///run/containerd/containerd.sock",
	})
	require.Error(t, c.Build(slog.New(slog.DiscardHandler), nil))
}
