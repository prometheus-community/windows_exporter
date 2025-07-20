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

package process

import (
	"context"
	"fmt"
	"log/slog"
	"runtime/debug"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/prometheus-community/windows_exporter/internal/mi"
	"github.com/prometheus/client_golang/prometheus"
)

type processWorkerRequest struct {
	ch                       chan<- prometheus.Metric
	name                     string
	performanceCounterValues perfDataCounterValues
	waitGroup                *sync.WaitGroup
	workerProcesses          []WorkerProcess
}

func (c *Collector) collect(ch chan<- prometheus.Metric) error {
	err := c.perfDataCollector.Collect(&c.perfDataObject)
	if err != nil {
		return fmt.Errorf("failed to collect metrics: %w", err)
	}

	err = nil

	var workerProcesses []WorkerProcess
	if c.config.EnableWorkerProcess {
		if err = c.miSession.Query(&workerProcesses, mi.NamespaceRootWebAdministration, c.workerProcessMIQueryQuery); err != nil {
			err = fmt.Errorf("WMI query for collector.process.iis failed: %w", err)
		}
	}

	wg := &sync.WaitGroup{}

	for _, process := range c.perfDataObject {
		// Duplicate processes are suffixed #, and an index number. Remove those.
		name, _, _ := strings.Cut(process.Name, ":") // Process V2

		// Duplicate processes are suffixed #, and an index number. Remove those.
		name, _, _ = strings.Cut(name, "#") // Process V1

		if c.config.ProcessExclude.MatchString(name) || !c.config.ProcessInclude.MatchString(name) {
			continue
		}

		if process.ProcessID == 0 && name != "Idle" {
			c.logger.LogAttrs(context.Background(), slog.LevelDebug, "Skipping process with PID 0",
				slog.String("name", name),
				slog.String("process_name", process.Name),
				slog.Any("process", fmt.Sprintf("%+v", process)),
			)

			continue
		}

		wg.Add(1)

		c.workerCh <- processWorkerRequest{
			ch:                       ch,
			name:                     name,
			performanceCounterValues: process,
			workerProcesses:          workerProcesses,
			waitGroup:                wg,
		}
	}

	wg.Wait()

	return err
}

func (c *Collector) collectWorker() {
	defer func() {
		if r := recover(); r != nil {
			c.logger.Error("Worker panic",
				slog.Any("panic", r),
				slog.String("stack", string(debug.Stack())),
			)

			// Restart the collectWorker
			go c.collectWorker()
		}
	}()

	for req := range c.workerCh {
		(func() {
			defer req.waitGroup.Done()

			ch := req.ch
			name := req.name
			data := req.performanceCounterValues

			pid := uint64(data.ProcessID)
			parentPID := strconv.FormatUint(uint64(data.CreatingProcessID), 10)

			if c.config.EnableWorkerProcess {
				for _, wp := range req.workerProcesses {
					if wp.ProcessId == pid {
						name = strings.Join([]string{name, wp.AppPoolName}, "_")

						break
					}
				}
			}

			cmdLine, processOwner, processGroupID, err := c.getProcessInformation(uint32(pid))
			if err != nil {
				slog.LogAttrs(context.Background(), slog.LevelDebug, "Failed to get process information",
					slog.Uint64("pid", pid),
					slog.Any("err", err),
				)
			}

			pidString := strconv.FormatUint(pid, 10)

			ch <- prometheus.MustNewConstMetric(
				c.info,
				prometheus.GaugeValue,
				1.0,
				name, pidString, parentPID, strconv.Itoa(int(processGroupID)), processOwner, cmdLine,
			)

			startTime := data.ElapsedTime
			if c.config.CounterVersion == 2 {
				// For V2, the ElapsedTime is in seconds, so we need to convert it to a timestamp.
				// The start time is the current time minus the elapsed time.
				startTime = float64(time.Now().UnixMicro())/1e6 - data.ElapsedTime
			}

			ch <- prometheus.MustNewConstMetric(
				c.startTime,
				prometheus.GaugeValue,
				startTime,
				name, pidString,
			)

			ch <- prometheus.MustNewConstMetric(
				c.handleCount,
				prometheus.GaugeValue,
				data.HandleCount,
				name, pidString,
			)

			ch <- prometheus.MustNewConstMetric(
				c.cpuTimeTotal,
				prometheus.CounterValue,
				data.PercentPrivilegedTime,
				name, pidString, "privileged",
			)

			ch <- prometheus.MustNewConstMetric(
				c.cpuTimeTotal,
				prometheus.CounterValue,
				data.PercentUserTime,
				name, pidString, "user",
			)

			ch <- prometheus.MustNewConstMetric(
				c.ioBytesTotal,
				prometheus.CounterValue,
				data.IoOtherBytesPerSec,
				name, pidString, "other",
			)

			ch <- prometheus.MustNewConstMetric(
				c.ioOperationsTotal,
				prometheus.CounterValue,
				data.IoOtherOperationsPerSec,
				name, pidString, "other",
			)

			ch <- prometheus.MustNewConstMetric(
				c.ioBytesTotal,
				prometheus.CounterValue,
				data.IoReadBytesPerSec,
				name, pidString, "read",
			)

			ch <- prometheus.MustNewConstMetric(
				c.ioOperationsTotal,
				prometheus.CounterValue,
				data.IoReadOperationsPerSec,
				name, pidString, "read",
			)

			ch <- prometheus.MustNewConstMetric(
				c.ioBytesTotal,
				prometheus.CounterValue,
				data.IoWriteBytesPerSec,
				name, pidString, "write",
			)

			ch <- prometheus.MustNewConstMetric(
				c.ioOperationsTotal,
				prometheus.CounterValue,
				data.IoWriteOperationsPerSec,
				name, pidString, "write",
			)

			ch <- prometheus.MustNewConstMetric(
				c.pageFaultsTotal,
				prometheus.CounterValue,
				data.PageFaultsPerSec,
				name, pidString,
			)

			ch <- prometheus.MustNewConstMetric(
				c.pageFileBytes,
				prometheus.GaugeValue,
				data.PageFileBytes,
				name, pidString,
			)

			ch <- prometheus.MustNewConstMetric(
				c.poolBytes,
				prometheus.GaugeValue,
				data.PoolNonPagedBytes,
				name, pidString, "nonpaged",
			)

			ch <- prometheus.MustNewConstMetric(
				c.poolBytes,
				prometheus.GaugeValue,
				data.PoolPagedBytes,
				name, pidString, "paged",
			)

			ch <- prometheus.MustNewConstMetric(
				c.priorityBase,
				prometheus.GaugeValue,
				data.PriorityBase,
				name, pidString,
			)

			ch <- prometheus.MustNewConstMetric(
				c.privateBytes,
				prometheus.GaugeValue,
				data.PrivateBytes,
				name, pidString,
			)

			ch <- prometheus.MustNewConstMetric(
				c.threadCount,
				prometheus.GaugeValue,
				data.ThreadCount,
				name, pidString,
			)

			ch <- prometheus.MustNewConstMetric(
				c.virtualBytes,
				prometheus.GaugeValue,
				data.VirtualBytes,
				name, pidString,
			)

			ch <- prometheus.MustNewConstMetric(
				c.workingSetPrivate,
				prometheus.GaugeValue,
				data.WorkingSetPrivate,
				name, pidString,
			)

			ch <- prometheus.MustNewConstMetric(
				c.workingSetPeak,
				prometheus.GaugeValue,
				data.WorkingSetPeak,
				name, pidString,
			)

			ch <- prometheus.MustNewConstMetric(
				c.workingSet,
				prometheus.GaugeValue,
				data.WorkingSet,
				name, pidString,
			)
		})()
	}
}
