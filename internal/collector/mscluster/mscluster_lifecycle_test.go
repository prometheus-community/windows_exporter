// SPDX-License-Identifier: MIT
//
// Copyright The Prometheus Authors

//go:build windows

package mscluster

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/prometheus-community/windows_exporter/internal/headers/clusapi"
	"github.com/prometheus/client_golang/prometheus"
)

// A ClusAPI RPC that hangs (for example in a resource DLL until its deadlock
// timeout) must not hold Collect: the other sub-collectors publish, Collect
// returns at the budget, and Close does not wait for the stuck read.
func TestCollectDoesNotWaitForStuckSource(t *testing.T) {
	stuck := &resourceFixtureSource{started: make(chan struct{}), release: make(chan struct{})}
	defer close(stuck.release)

	networks := &objectFixtureSource{objects: []clusapi.Object{networkFixture("Cluster Network 1")}}

	c := New(&Config{CollectorsEnabled: []string{subCollectorNetwork, subCollectorResource}})
	c.resourceSource, c.networkSource = stuck, networks
	c.buildResourceDescriptors()
	c.buildNetworkDescriptors()

	ch := make(chan prometheus.Metric, 100)
	start := time.Now()

	err := c.Collect(ch, 50*time.Millisecond)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("stuck resource read: %v", err)
	}

	if elapsed := time.Since(start); elapsed > 5*time.Second {
		t.Fatalf("Collect waited %s for the stuck read", elapsed)
	}

	close(ch)

	published := 0
	for range ch {
		published++
	}

	if published == 0 {
		t.Fatal("network metrics were not published")
	}

	closeDone := make(chan error, 1)

	go func() { closeDone <- c.Close() }()

	select {
	case err := <-closeDone:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("Close waited for the stuck read")
	}
}

// Library consumers may call Collect after a failed Build.
func TestCollectWithoutSources(t *testing.T) {
	c := New(&Config{CollectorsEnabled: []string{subCollectorCluster, subCollectorNetwork, subCollectorNode, subCollectorResource, subCollectorResourceGroup, subCollectorSharedVolumes}})

	err := c.Collect(make(chan prometheus.Metric, 10), time.Second)
	if err == nil || !strings.Contains(err.Error(), "not built") {
		t.Fatalf("Collect without sources: %v", err)
	}
}
