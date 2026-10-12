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

package hyperv

import (
	"errors"
	"log/slog"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/prometheus-community/windows_exporter/internal/headers/virtdisk"
	"github.com/stretchr/testify/require"
)

func noVirtualStorageDevices() ([]string, error) { return nil, nil }

func noParents(string) ([]string, error) { return nil, nil }

func testDisks(paths ...string) []attachedVirtualHardDisk {
	disks := make([]attachedVirtualHardDisk, 0, len(paths))
	for _, path := range paths {
		disks = append(disks, attachedVirtualHardDisk{VMID: "id", VMName: "vm", Path: path})
	}

	return disks
}

func TestVirtualHardDiskRefresherFirstRefresh(t *testing.T) {
	t.Parallel()

	errUnreadable := errors.New("unreadable")
	refreshed := time.Date(2026, 10, 12, 8, 0, 0, 0, time.UTC)

	refresher := newVirtualHardDiskRefresher(slog.New(slog.DiscardHandler), time.Hour,
		func() ([]attachedVirtualHardDisk, error) { return testDisks(`C:\a.vhdx`, `C:\b.vhdx`), nil },
		noVirtualStorageDevices,
		virtualHardDiskReader{getInfo: func(path string) (virtdisk.Info, error) {
			if path == `C:\b.vhdx` {
				return virtdisk.Info{}, errUnreadable
			}

			return virtdisk.Info{VirtualSize: 10, PhysicalSize: 5, ProviderSubtype: virtdisk.ProviderSubtypeDynamic}, nil
		}, parentLocations: noParents},
	)
	refresher.now = func() time.Time { return refreshed }

	require.Equal(t, virtualHardDiskSnapshot{}, refresher.get(), "no data before the first refresh")

	refresher.start(time.Second)
	t.Cleanup(refresher.stop)

	snapshot := refresher.get()
	require.NoError(t, snapshot.err, "per-disk errors are logged, not returned")
	require.Equal(t, refreshed, snapshot.refreshed)
	require.Equal(t, 1, snapshot.readErrors)
	require.Len(t, snapshot.results, 1)
	require.Equal(t, `C:\a.vhdx`, snapshot.results[0].disk.Path)
	require.Equal(t, uint64(10), snapshot.results[0].info.VirtualSize)
}

func TestVirtualHardDiskRefresherListErrorKeepsResults(t *testing.T) {
	t.Parallel()

	errWMI := errors.New("wmi is broken")

	var calls atomic.Int32

	refreshes := make(chan struct{}, 10)
	done := make(chan struct{})

	refresher := newVirtualHardDiskRefresher(slog.New(slog.DiscardHandler), 10*time.Millisecond,
		func() ([]attachedVirtualHardDisk, error) {
			defer func() { refreshes <- struct{}{} }()

			switch calls.Add(1) {
			case 1:
				return testDisks(`C:\a.vhdx`), nil
			case 2:
				return nil, errWMI
			default:
				// Block further refreshes, so the test sees the state after the error.
				<-done

				return nil, errWMI
			}
		},
		noVirtualStorageDevices,
		virtualHardDiskReader{getInfo: func(string) (virtdisk.Info, error) { return virtdisk.Info{VirtualSize: 10}, nil }, parentLocations: noParents},
	)

	refresher.start(time.Second)
	t.Cleanup(func() {
		close(done)
		refresher.stop()
	})

	<-refreshes

	first := refresher.get()

	<-refreshes
	require.Eventually(t, func() bool { return refresher.get().err != nil }, time.Second, time.Millisecond)

	snapshot := refresher.get()
	require.ErrorIs(t, snapshot.err, errWMI)
	require.Len(t, snapshot.results, 1, "the last good disks stay published")
	require.Equal(t, first.refreshed, snapshot.refreshed, "the refresh time stays at the last good refresh")
}

func TestVirtualHardDiskRefresherStopWaitsForList(t *testing.T) {
	t.Parallel()

	listing := make(chan struct{})
	release := make(chan struct{})

	var calls atomic.Int32

	refresher := newVirtualHardDiskRefresher(slog.New(slog.DiscardHandler), time.Millisecond,
		func() ([]attachedVirtualHardDisk, error) {
			if calls.Add(1) == 2 {
				close(listing)
				<-release
			}

			return nil, nil
		},
		noVirtualStorageDevices,
		virtualHardDiskReader{getInfo: func(string) (virtdisk.Info, error) { return virtdisk.Info{}, nil }, parentLocations: noParents},
	)

	refresher.start(time.Second)
	<-listing

	var stopped atomic.Bool

	var wg sync.WaitGroup

	wg.Go(func() {
		refresher.stop()
		stopped.Store(true)
	})

	time.Sleep(20 * time.Millisecond)
	require.False(t, stopped.Load(), "stop returned while the WMI query was still running")

	close(release)
	wg.Wait()

	callsAtStop := calls.Load()

	time.Sleep(20 * time.Millisecond)
	require.Equal(t, callsAtStop, calls.Load(), "the refresher kept running after stop")
}

func TestVirtualHardDiskRefresherStartIsBounded(t *testing.T) {
	t.Parallel()

	release := make(chan struct{})

	refresher := newVirtualHardDiskRefresher(slog.New(slog.DiscardHandler), time.Hour,
		func() ([]attachedVirtualHardDisk, error) {
			<-release

			return nil, nil
		},
		noVirtualStorageDevices,
		virtualHardDiskReader{getInfo: func(string) (virtdisk.Info, error) { return virtdisk.Info{}, nil }, parentLocations: noParents},
	)

	start := time.Now()
	refresher.start(20 * time.Millisecond)
	require.Less(t, time.Since(start), time.Second, "start waited for a hanging first refresh")

	close(release)
	refresher.stop()
}

func TestVirtualHardDiskRefresherDevice(t *testing.T) {
	t.Parallel()

	const (
		vhdx       = `C:\CSV01\VM01\Disk(1).vhdx`
		avhd       = `C:\CSV06\VM02\vm02_os_8C9A696D-1374-4DD6-B15C-B74E2CA47C29.avhd`
		avhdParent = `C:\CSV06\VM02\vm02_os.vhd`
		off        = `C:\CSV09\VM09\off.vhdx`
	)

	reader := virtualHardDiskReader{
		getInfo: func(path string) (virtdisk.Info, error) {
			if path == avhd {
				return virtdisk.Info{ProviderSubtype: virtdisk.ProviderSubtypeDifferencing}, nil
			}

			return virtdisk.Info{ProviderSubtype: virtdisk.ProviderSubtypeDynamic}, nil
		},
		parentLocations: func(path string) ([]string, error) {
			if path == avhd {
				return []string{avhdParent}, nil
			}

			return nil, nil
		},
	}

	for name, tc := range map[string]struct {
		instances func() ([]string, error)
		want      map[string]string
	}{
		"matched": {
			instances: func() ([]string, error) {
				return []string{`C:-CSV01-VM01-Disk[1].vhdx`, `C:-CSV06-VM02-vm02_os.vhd`, `--?-C:-CSV01-VM01-Virtual Machines-x.vmgs`}, nil
			},
			want: map[string]string{vhdx: `C:-CSV01-VM01-Disk[1].vhdx`, avhd: `C:-CSV06-VM02-vm02_os.vhd`, off: ""},
		},
		"counters fail": {
			instances: func() ([]string, error) { return nil, errors.New("pdh failed") },
			want:      map[string]string{vhdx: "", avhd: "", off: ""},
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			refresher := newVirtualHardDiskRefresher(slog.New(slog.DiscardHandler), time.Hour,
				func() ([]attachedVirtualHardDisk, error) { return testDisks(vhdx, avhd, off), nil },
				tc.instances,
				reader,
			)

			refresher.start(time.Second)
			t.Cleanup(refresher.stop)

			snapshot := refresher.get()
			require.NoError(t, snapshot.err, "a failed counter read must not fail the refresh")

			got := make(map[string]string)
			for _, result := range snapshot.results {
				got[result.disk.Path] = result.device
			}

			require.Equal(t, tc.want, got)
		})
	}
}
