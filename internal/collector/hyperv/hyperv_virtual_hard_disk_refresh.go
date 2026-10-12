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
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/prometheus-community/windows_exporter/internal/utils/recovery"
)

// virtualHardDiskSnapshot is the result of the last refresh.
type virtualHardDiskSnapshot struct {
	results []virtualHardDiskResult
	// readErrors is the number of disks of the last refresh that could not be read.
	readErrors int
	// refreshed is the time of the last successful refresh, zero before the first one.
	refreshed time.Time
	// err is the error of the last disk listing. The results of the last
	// successful refresh are kept.
	err error
}

// virtualHardDiskRefresher reads the virtual hard disks in the background, so
// a scrape only publishes the last values and a slow or hanging disk can't
// block the hyperv collector.
type virtualHardDiskRefresher struct {
	logger   *slog.Logger
	interval time.Duration
	list     func() ([]attachedVirtualHardDisk, error)
	// instances returns the "Hyper-V Virtual Storage Device" counter instances.
	instances func() ([]string, error)
	reader    virtualHardDiskReader
	now       func() time.Time

	// listMu is held while list and instances run. stop waits for it, because
	// they use the MI session and the PDH query, which are released after Close.
	listMu sync.Mutex

	mu       sync.RWMutex
	snapshot virtualHardDiskSnapshot

	cancel context.CancelFunc
}

func newVirtualHardDiskRefresher(
	logger *slog.Logger,
	interval time.Duration,
	list func() ([]attachedVirtualHardDisk, error),
	instances func() ([]string, error),
	reader virtualHardDiskReader,
) *virtualHardDiskRefresher {
	return &virtualHardDiskRefresher{
		logger:    logger,
		interval:  interval,
		list:      list,
		instances: instances,
		reader:    reader,
		now:       time.Now,
	}
}

// start refreshes in the background and waits up to firstRefreshTimeout for
// the first refresh, so the first scrape after a start already has data.
func (r *virtualHardDiskRefresher) start(firstRefreshTimeout time.Duration) {
	ctx, cancel := context.WithCancel(context.Background())
	r.cancel = cancel

	firstRefresh := make(chan struct{})

	go func() {
		for first := true; ; first = false {
			if err := recovery.Run(func() error {
				r.refresh(ctx)

				return nil
			}); err != nil {
				r.logger.Error("virtual hard disk refresh failed", slog.Any("err", err))
			}

			if first {
				close(firstRefresh)
			}

			select {
			case <-ctx.Done():
				return
			case <-time.After(r.interval):
			}
		}
	}()

	select {
	case <-firstRefresh:
	case <-time.After(firstRefreshTimeout):
		r.logger.Warn("first virtual hard disk refresh did not finish in time, continuing in the background",
			slog.Duration("timeout", firstRefreshTimeout),
		)
	}
}

// stop ends the background refresh. It waits for a running disk listing, but
// not for disk reads, which don't use shared resources.
func (r *virtualHardDiskRefresher) stop() {
	if r.cancel != nil {
		r.cancel()
	}

	r.listMu.Lock()
	defer r.listMu.Unlock()
}

func (r *virtualHardDiskRefresher) get() virtualHardDiskSnapshot {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.snapshot
}

func (r *virtualHardDiskRefresher) refresh(ctx context.Context) {
	disks, instances, err := r.listDisks(ctx)
	if ctx.Err() != nil {
		return
	}

	if err != nil {
		r.logger.WarnContext(ctx, "failed to list virtual hard disks", slog.Any("err", err))

		r.mu.Lock()
		r.snapshot.err = err
		r.mu.Unlock()

		return
	}

	results, err := readVirtualHardDisks(disks, virtualHardDiskWorkers, r.reader)
	if err != nil {
		r.logger.WarnContext(ctx, "failed to read virtual hard disks", slog.Any("err", err))
	}

	matcher := newVirtualStorageDeviceMatcher(instances)
	for i := range results {
		results[i].device = matcher.match(results[i].chain)
	}

	r.mu.Lock()
	r.snapshot = virtualHardDiskSnapshot{
		results:    results,
		readErrors: len(disks) - len(results),
		refreshed:  r.now(),
	}
	r.mu.Unlock()
}

// listDisks returns the attached disks and the counter instances. A failure to
// read the instances is logged; the disks are then published without device.
func (r *virtualHardDiskRefresher) listDisks(ctx context.Context) ([]attachedVirtualHardDisk, []string, error) {
	r.listMu.Lock()
	defer r.listMu.Unlock()

	if ctx.Err() != nil {
		return nil, nil, ctx.Err()
	}

	disks, err := r.list()
	if err != nil {
		return nil, nil, err
	}

	instances, err := r.instances()
	if err != nil {
		r.logger.WarnContext(ctx, "failed to read Hyper-V Virtual Storage Device instances", slog.Any("err", err))
	}

	return disks, instances, nil
}
