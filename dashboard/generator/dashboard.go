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

package main

import "fmt"

// buildDashboard returns the v1 dashboard. Rows titled "Tab" or "Tab / Row" become tabs and rows in the v2 dashboard.
func buildDashboard() M {
	panels := []any{}
	add := func(ps ...M) {
		for _, p := range ps {
			// Panel IDs are numbered in order. The v2 conversion turns them into the element names (panel-<id>).
			p["id"] = len(panels) + 1
			panels = append(panels, p)
		}
	}
	// y is the top of the next panel row in grid units.
	y := 0
	// tab starts a group of panels. The title "Tab / Row" starts a row inside the tab; see regroupTabs.
	tab := func(title string) {
		add(row(title, y))
		y++
	}

	cpuUtil := func(sel string) string {
		return fmt.Sprintf(`100 * (1 - avg by (instance) (clamp_max(rate(windows_cpu_time_total{%s, mode="idle"}[$__rate_interval]), 1)))`, sel)
	}
	memUtil := func(sel string) string {
		return fmt.Sprintf(`100 * (1 - windows_memory_physical_free_bytes{%s} / windows_memory_physical_total_bytes{%s})`, sel, sel)
	}
	volSelFleet := F + `, volume!~"HarddiskVolume.*"`
	volSel := H + `, volume=~"$volume"`
	nicSel := H + `, nic=~"$nic"`

	// Fleet graphs show only the busiest hosts so they stay readable for large fleets.
	top := func(expr string) string { return "topk(25, " + expr + ")" }
	topDesc := "Shows the 25 highest hosts at each point in time."

	// ================================================================ Fleet
	tab("Fleet")
	add(table(grid(0, y, 24, 10), TB{
		Title: "Hosts",
		Desc:  "One row per scraped host. Click a hostname to open its details. Services up counts running services; Auto stopped counts services with start mode auto that are not running.",
		Targets: []T{
			{Expr: `max by (instance, hostname) (windows_os_hostname{` + F + `, hostname=~"$hostname"})`},
			{Expr: fleetFilter(`max by (instance, product, version) (windows_os_info{` + F + `})`)},
			{Expr: fleetFilter(`max by (instance) (time() - windows_system_boot_time_timestamp{` + F + `})`)},
			{Expr: fleetFilter(`max by (instance) (windows_cpu_logical_processor{` + F + `})`)},
			{Expr: fleetFilter(cpuUtil(F))},
			{Expr: fleetFilter(`max by (instance) (windows_memory_physical_total_bytes{` + F + `})`)},
			{Expr: fleetFilter(`max by (instance) (` + memUtil(F) + `)`)},
			{Expr: fleetFilter(`max by (instance) (100 * windows_memory_committed_bytes{` + F + `} / windows_memory_commit_limit{` + F + `})`)},
			{Expr: fleetFilter(`max by (instance) (100 * (1 - windows_logical_disk_free_bytes{` + volSelFleet + `} / windows_logical_disk_size_bytes{` + volSelFleet + `}))`)},
			{Expr: fleetFilter(`sum by (instance) (windows_service_state{` + F + `, state="running"})`)},
			{Expr: fleetFilter(`count by (instance) (windows_service_state{` + F + `, state="running"} == 0 and on (instance, name) windows_service_start_mode{` + F + `, start_mode="auto"} == 1)`)},
			{Expr: fleetFilter(`max by (instance) (windows_system_processes{` + F + `})`)},
		},
		JoinBy: "instance",
		Order: M{
			"hostname": 0, "instance": 1, "product": 2, "version": 3, "Value #C": 4, "Value #D": 5, "Value #E": 6, "Value #F": 7,
			"Value #G": 8, "Value #H": 9, "Value #I": 10, "Value #J": 11, "Value #K": 12, "Value #L": 13,
		},
		Exclude: []string{"Value #A", "Value #B"},
		Rename: M{
			"hostname": "Hostname", "instance": "Instance", "product": "OS", "version": "Version",
			"Value #C": "Uptime", "Value #D": "CPUs", "Value #E": "CPU", "Value #F": "Memory",
			"Value #G": "Memory used", "Value #H": "Commit used", "Value #I": "Fullest volume",
			"Value #J": "Services up", "Value #K": "Auto stopped", "Value #L": "Processes",
		},
		SortBy: []any{M{"desc": false, "displayName": "Hostname"}},
		Footer: []string{"CPUs", "Memory", "Processes"},
		Overrides: []any{
			byName("hostname", prop("links", []any{M{
				"title": "Show details for ${__data.fields.hostname}",
				"url":   "/d/${__dashboard.uid}?${__url_time_range}&${datasource:queryparam}&${job:queryparam}&var-hostname=$__all&var-instance=${__data.fields.instance}&dtab=Overview",
			}}), prop("custom.width", 180), prop("custom.filterable", true)),
			byName("Value #C", prop("unit", "dtdurations"), prop("custom.width", 90)),
			byName("instance", prop("custom.width", 140)),
			byName("product", prop("custom.width", 160), prop("custom.filterable", true)),
			byName("version", prop("custom.width", 100), prop("custom.filterable", true)),
			byName("Value #D", prop("unit", "none"), prop("custom.width", 60)),
			byName("Value #E", gaugeCell("percent", 100)...),
			byName("Value #F", prop("unit", "bytes"), prop("custom.width", 90)),
			byName("Value #G", gaugeCell("percent", 100)...),
			byName("Value #H", gaugeCell("percent", 100)...),
			byName("Value #I", gaugeCell("percent", 100)...),
			byName("Value #J", unitCell("none")...),
			byName("Value #K", prop("unit", "none"), prop("thresholds", steps("green", nil, "orange", 1)),
				prop("custom.cellOptions", M{"type": "color-text"}), prop("noValue", "0")),
			byName("Value #L", prop("unit", "none"), prop("custom.width", 90)),
		},
	}))
	y += 10
	add(
		timeseries(grid(0, y, 8, 8), TS{
			Title: "CPU utilization", Desc: topDesc, Unit: "percent", Min: 0, Max: 100, Legend: "list",
			Targets: []T{{Expr: top(fleetJoin(cpuUtil(F))), Legend: "{{hostname}}"}},
		}),
		timeseries(grid(8, y, 8, 8), TS{
			Title: "Memory utilization", Desc: topDesc, Unit: "percent", Min: 0, Max: 100, Legend: "list",
			Targets: []T{{Expr: top(fleetJoin(`max by (instance) (` + memUtil(F) + `)`)), Legend: "{{hostname}}"}},
		}),
		timeseries(grid(16, y, 8, 8), TS{
			Title: "Fullest volume", Desc: "Usage of the fullest volume per host. " + topDesc, Unit: "percent", Min: 0, Max: 100, Legend: "list",
			Steps: pctSteps(), Thresholds: true,
			Targets: []T{{Expr: top(fleetJoin(`max by (instance) (100 * (1 - windows_logical_disk_free_bytes{` + volSelFleet + `} / windows_logical_disk_size_bytes{` + volSelFleet + `}))`)), Legend: "{{hostname}}"}},
		}),
	)
	y += 8
	add(
		timeseries(grid(0, y, 8, 8), TS{
			Title: "Network throughput", Desc: "Sum over all interfaces. Received is drawn below the axis. " + topDesc, Unit: "bps", Legend: "list",
			Targets: []T{
				{Expr: top(fleetJoin(`sum by (instance) (rate(windows_net_bytes_sent_total{` + F + `}[$__rate_interval]) * 8)`)), Legend: "{{hostname}} sent"},
				{Expr: top(fleetJoin(`sum by (instance) (rate(windows_net_bytes_received_total{` + F + `}[$__rate_interval]) * 8)`)), Legend: "{{hostname}} received"},
			},
			Overrides: []any{negativeY(".* received$")},
		}),
		timeseries(grid(8, y, 8, 8), TS{
			Title: "Disk throughput", Desc: "Sum over all volumes. Read is drawn below the axis. " + topDesc, Unit: "Bps", Legend: "list",
			Targets: []T{
				{Expr: top(fleetJoin(`sum by (instance) (rate(windows_logical_disk_write_bytes_total{` + volSelFleet + `}[$__rate_interval]))`)), Legend: "{{hostname}} write"},
				{Expr: top(fleetJoin(`sum by (instance) (rate(windows_logical_disk_read_bytes_total{` + volSelFleet + `}[$__rate_interval]))`)), Legend: "{{hostname}} read"},
			},
			Overrides: []any{negativeY(".* read$")},
		}),
		timeseries(grid(16, y, 8, 8), TS{
			Title: "Network errors and discards", Desc: "Sum over all interfaces of received and outbound errors and discards. " + topDesc, Unit: "pps", Min: 0, Legend: "list",
			Targets: []T{{Expr: top(fleetJoin(`sum by (instance) (rate(windows_net_packets_received_errors_total{` + F + `}[$__rate_interval]) + rate(windows_net_packets_received_discarded_total{` + F + `}[$__rate_interval]) + rate(windows_net_packets_outbound_errors_total{` + F + `}[$__rate_interval]) + rate(windows_net_packets_outbound_discarded_total{` + F + `}[$__rate_interval]))`)), Legend: "{{hostname}}"}},
		}),
	)
	y += 8

	// ================================================================ Overview
	tab("Overview")
	add(
		stat(grid(0, y, 4, 4), ST{
			Title: "Operating system", Expr: `windows_os_info{` + H + `}`, Text: "product", Color: "none",
		}),
		stat(grid(4, y, 3, 4), ST{
			Title: "Uptime", Unit: "dtdurations", Desc: "Time since the last boot.",
			Expr:  `time() - windows_system_boot_time_timestamp{` + H + `}`,
			Steps: steps("orange", nil, "text", 3600),
		}),
		stat(grid(7, y, 2, 4), ST{
			Title: "CPUs", Unit: "none",
			Expr: `windows_cpu_logical_processor{` + H + `}`,
		}),
		stat(grid(9, y, 2, 4), ST{
			Title: "RAM", Unit: "bytes", Decimals: 1,
			Expr: `windows_memory_physical_total_bytes{` + H + `}`,
		}),
		stat(grid(11, y, 3, 4), ST{
			Title: "CPU busy", Unit: "percent", Min: 0, Max: 100, Graph: true, Steps: pctSteps(), Decimals: 1,
			Expr: cpuUtil(H),
		}),
		stat(grid(14, y, 3, 4), ST{
			Title: "Memory used", Unit: "percent", Min: 0, Max: 100, Graph: true, Steps: pctSteps(), Decimals: 1,
			Expr: memUtil(H),
		}),
		stat(grid(17, y, 3, 4), ST{
			Title: "Commit used", Desc: "Committed virtual memory as a share of the commit limit (physical memory + page files). Allocations fail at 100%.",
			Unit: "percent", Min: 0, Max: 100, Graph: true, Steps: pctSteps(), Decimals: 1,
			Expr: `100 * windows_memory_committed_bytes{` + H + `} / windows_memory_commit_limit{` + H + `}`,
		}),
		stat(grid(20, y, 4, 4), ST{
			Title: "Stopped auto services", Desc: "Services with start mode \"auto\" that are not running. See the Services tab for names.",
			Unit: "none", Steps: steps("green", nil, "orange", 1),
			Expr: `count(windows_service_state{` + H + `, state="running"} == 0 and on (name) windows_service_start_mode{` + H + `, start_mode="auto"} == 1) or vector(0)`,
		}),
	)
	y += 4
	add(
		timeseries(grid(0, y, 12, 8), TS{
			Title: "CPU utilization", Unit: "percent", Min: 0, Max: 100, Legend: "list", Steps: pctSteps(),
			Targets: []T{{Expr: cpuUtil(H), Legend: "CPU busy"}},
		}),
		timeseries(grid(12, y, 12, 8), TS{
			Title: "Physical memory", Unit: "bytes", Min: 0, Stack: true, Fill: 40, Legend: "list",
			Desc: "Available memory is the standby (cache), free and zero page lists.",
			Targets: []T{
				{Expr: `windows_memory_physical_total_bytes{` + H + `} - windows_memory_physical_free_bytes{` + H + `}`, Legend: "used"},
				{Expr: `windows_memory_physical_free_bytes{` + H + `}`, Legend: "available"},
			},
			Overrides: []any{byName("used", fixedColor("orange")), byName("available", fixedColor("green"))},
		}),
	)
	y += 8
	add(
		table(grid(0, y, 8, 8), TB{
			Title: "Volumes",
			Targets: []T{
				{Expr: `windows_logical_disk_size_bytes{` + volSel + `}`},
				{Expr: `windows_logical_disk_free_bytes{` + volSel + `}`},
				{Expr: `100 * (1 - windows_logical_disk_free_bytes{` + volSel + `} / windows_logical_disk_size_bytes{` + volSel + `})`},
			},
			JoinBy:  "volume",
			Exclude: []string{"instance", "job", "port", "__name__", "instance 1", "instance 2", "instance 3", "job 1", "job 2", "job 3", "port 1", "port 2", "port 3"},
			Order:   M{"volume": 0, "Value #A": 1, "Value #B": 2, "Value #C": 3},
			Rename:  M{"volume": "Volume", "Value #A": "Size", "Value #B": "Free", "Value #C": "Used"},
			SortBy:  []any{M{"desc": true, "displayName": "Used"}},
			Overrides: []any{
				byName("Value #A", unitCell("bytes")...),
				byName("Value #B", unitCell("bytes")...),
				byName("Value #C", gaugeCell("percent", 100)...),
			},
		}),
		timeseries(grid(8, y, 8, 8), TS{
			Title: "Disk throughput", Unit: "Bps", Desc: "Sum over the selected volumes. Reads are drawn below the axis.", Legend: "list",
			Targets: []T{
				{Expr: `sum(rate(windows_logical_disk_write_bytes_total{` + volSel + `}[$__rate_interval]))`, Legend: "write"},
				{Expr: `sum(rate(windows_logical_disk_read_bytes_total{` + volSel + `}[$__rate_interval]))`, Legend: "read"},
			},
			Overrides: []any{negativeY("read")},
		}),
		timeseries(grid(16, y, 8, 8), TS{
			Title: "Network throughput", Unit: "bps", Desc: "Sum over the selected interfaces. Received traffic is drawn below the axis.", Legend: "list",
			Targets: []T{
				{Expr: `sum(rate(windows_net_bytes_sent_total{` + nicSel + `}[$__rate_interval])) * 8`, Legend: "sent"},
				{Expr: `sum(rate(windows_net_bytes_received_total{` + nicSel + `}[$__rate_interval])) * 8`, Legend: "received"},
			},
			Overrides: []any{negativeY("received")},
		}),
	)
	y += 8

	// ================================================================ CPU and system
	tab("CPU")
	add(
		timeseries(grid(0, y, 12, 10), TS{
			Title: "CPU utilization by mode", Desc: "Share of total CPU time across all logical processors.",
			Unit: "percentunit", Min: 0, Max: 1, Stack: true, Fill: 40,
			Targets: []T{{
				Expr:   `sum by (mode) (rate(windows_cpu_time_total{` + H + `, mode!="idle"}[$__rate_interval])) / scalar(count(windows_cpu_time_total{` + H + `, mode="idle"}))`,
				Legend: "{{mode}}",
			}},
		}),
		// Single-digit core numbers are padded ("0,2" -> "0,02") so the rows sort numerically.
		stateTimeline(grid(12, y, 12, 10), "CPU utilization per core",
			"Busy time of each logical processor, labelled processor group,core. Green is idle, red is fully busy.",
			"percentunit", 1, T{
				Expr:   `label_replace(clamp(1 - rate(windows_cpu_time_total{` + H + `, mode="idle"}[$__rate_interval]), 0, 1), "core", "$1,0$2", "core", "(\\d+),(\\d)")`,
				Legend: "{{core}}",
			}),
	)
	y += 10
	add(
		timeseries(grid(0, y, 8, 8), TS{
			Title: "Processor queue length", Unit: "short", Min: 0, Legend: "list",
			Desc:    "Threads that are ready to run but waiting for a CPU. A sustained value above 2 per core indicates CPU contention.",
			Targets: []T{{Expr: `windows_system_processor_queue_length{` + H + `}`, Legend: "queue length"}},
		}),
		timeseries(grid(8, y, 8, 8), TS{
			Title: "Context switches and interrupts", Unit: "ops", Min: 0,
			Targets: []T{
				{Expr: `rate(windows_system_context_switches_total{` + H + `}[$__rate_interval])`, Legend: "context switches"},
				{Expr: `sum(rate(windows_cpu_interrupts_total{` + H + `}[$__rate_interval]))`, Legend: "interrupts"},
				{Expr: `sum(rate(windows_cpu_dpcs_total{` + H + `}[$__rate_interval]))`, Legend: "DPCs"},
			},
		}),
		timeseries(grid(16, y, 8, 8), TS{
			Title: "Processes and threads", Unit: "short", Min: 0,
			Targets: []T{
				{Expr: `windows_system_processes{` + H + `}`, Legend: "processes"},
				{Expr: `windows_system_threads{` + H + `}`, Legend: "threads"},
			},
			Overrides: []any{byName("threads", prop("custom.axisPlacement", "right"))},
		}),
	)
	y += 8
	add(
		timeseries(grid(0, y, 12, 8), TS{
			Title: "System calls and exceptions", Unit: "ops", Min: 0,
			Targets: []T{
				{Expr: `rate(windows_system_system_calls_total{` + H + `}[$__rate_interval])`, Legend: "system calls"},
				{Expr: `rate(windows_system_exception_dispatches_total{` + H + `}[$__rate_interval])`, Legend: "exception dispatches"},
			},
			Overrides: []any{byName("exception dispatches", prop("custom.axisPlacement", "right"))},
		}),
		timeseries(grid(12, y, 12, 8), TS{
			Title: "CPU frequency", Unit: "hertz", Min: 0, Legend: "list",
			Desc: "Effective frequency from the APERF/MPERF counters, averaged and maximum across logical processors, against the nominal frequency. Values above nominal mean turbo boost.",
			Targets: []T{
				{Expr: `avg(1e4 * windows_cpu_core_frequency_mhz{` + H + `} * rate(windows_cpu_processor_performance_total{` + H + `}[$__rate_interval]) / rate(windows_cpu_processor_mperf_total{` + H + `}[$__rate_interval]))`, Legend: "average"},
				{Expr: `max(1e4 * windows_cpu_core_frequency_mhz{` + H + `} * rate(windows_cpu_processor_performance_total{` + H + `}[$__rate_interval]) / rate(windows_cpu_processor_mperf_total{` + H + `}[$__rate_interval]))`, Legend: "maximum"},
				{Expr: `avg(windows_cpu_core_frequency_mhz{` + H + `}) * 1e6`, Legend: "nominal"},
			},
			Overrides: []any{byName("nominal", fixedColor("text"), prop("custom.fillOpacity", 0), prop("custom.lineStyle", M{"dash": []int{10, 10}, "fill": "dash"}))},
		}),
	)
	y += 8

	// ================================================================ Memory
	tab("Memory")
	add(
		timeseries(grid(0, y, 12, 8), TS{
			Title: "Physical memory", Unit: "bytes", Min: 0, Stack: true, Fill: 40,
			Desc: "Available memory is the standby (cache), free and zero page lists.",
			Targets: []T{
				{Expr: `windows_memory_physical_total_bytes{` + H + `} - windows_memory_physical_free_bytes{` + H + `}`, Legend: "used"},
				{Expr: `windows_memory_physical_free_bytes{` + H + `}`, Legend: "available"},
			},
			Overrides: []any{byName("used", fixedColor("orange")), byName("available", fixedColor("green"))},
		}),
		timeseries(grid(12, y, 12, 8), TS{
			Title: "Commit charge", Unit: "bytes", Min: 0,
			Desc: "Committed virtual memory against the commit limit (physical memory + page files).",
			Targets: []T{
				{Expr: `windows_memory_committed_bytes{` + H + `}`, Legend: "committed"},
				{Expr: `windows_memory_commit_limit{` + H + `}`, Legend: "limit"},
			},
			Overrides: []any{byName("limit", fixedColor("red"), prop("custom.fillOpacity", 0), prop("custom.lineStyle", M{"dash": []int{10, 10}, "fill": "dash"}))},
		}),
	)
	y += 8
	add(
		timeseries(grid(0, y, 12, 8), TS{
			Title: "Paging", Unit: "ops", Min: 0,
			Desc: "Hard page faults read from or written to disk. Sustained page reads indicate memory pressure.",
			Targets: []T{
				{Expr: `rate(windows_memory_swap_page_reads_total{` + H + `}[$__rate_interval])`, Legend: "page reads"},
				{Expr: `rate(windows_memory_swap_page_writes_total{` + H + `}[$__rate_interval])`, Legend: "page writes"},
			},
		}),
		timeseries(grid(12, y, 12, 8), TS{
			Title: "Kernel pools and cache", Unit: "bytes", Min: 0,
			Desc: "A steadily growing nonpaged pool usually means a driver leaks memory.",
			Targets: []T{
				{Expr: `windows_memory_pool_nonpaged_bytes{` + H + `}`, Legend: "nonpaged pool"},
				{Expr: `windows_memory_pool_paged_bytes{` + H + `}`, Legend: "paged pool"},
				{Expr: `windows_memory_cache_bytes{` + H + `}`, Legend: "system cache"},
			},
		}),
	)
	y += 8

	// ================================================================ Disk
	tab("Disk")
	add(
		table(grid(0, y, 8, 8), TB{
			Title: "Volumes",
			Targets: []T{
				{Expr: `windows_logical_disk_size_bytes{` + volSel + `}`},
				{Expr: `windows_logical_disk_free_bytes{` + volSel + `}`},
				{Expr: `100 * (1 - windows_logical_disk_free_bytes{` + volSel + `} / windows_logical_disk_size_bytes{` + volSel + `})`},
			},
			JoinBy:  "volume",
			Exclude: []string{"instance", "job", "port", "__name__", "instance 1", "instance 2", "instance 3", "job 1", "job 2", "job 3", "port 1", "port 2", "port 3"},
			Order:   M{"volume": 0, "Value #A": 1, "Value #B": 2, "Value #C": 3},
			Rename:  M{"volume": "Volume", "Value #A": "Size", "Value #B": "Free", "Value #C": "Used"},
			SortBy:  []any{M{"desc": true, "displayName": "Used"}},
			Overrides: []any{
				byName("Value #A", unitCell("bytes")...),
				byName("Value #B", unitCell("bytes")...),
				byName("Value #C", gaugeCell("percent", 100)...),
			},
		}),
		timeseries(grid(8, y, 8, 8), TS{
			Title: "Volume usage", Unit: "percent", Min: 0, Max: 100, Steps: pctSteps(), Thresholds: true,
			Targets: []T{{Expr: `100 * (1 - windows_logical_disk_free_bytes{` + volSel + `} / windows_logical_disk_size_bytes{` + volSel + `})`, Legend: "{{volume}}"}},
			Calcs:   []string{"min", "max", "lastNotNull"},
		}),
		timeseries(grid(16, y, 8, 8), TS{
			Title: "Volume busy time", Unit: "percentunit", Min: 0, Max: 1,
			Desc:    "Share of time the volume was servicing requests (1 - idle time).",
			Targets: []T{{Expr: `1 - clamp_max(rate(windows_logical_disk_idle_seconds_total{` + volSel + `}[$__rate_interval]), 1)`, Legend: "{{volume}}"}},
		}),
	)
	y += 8
	add(
		timeseries(grid(0, y, 12, 8), TS{
			Title: "Disk throughput", Unit: "Bps", Desc: "Reads are drawn below the axis.",
			Targets: []T{
				{Expr: `rate(windows_logical_disk_write_bytes_total{` + volSel + `}[$__rate_interval])`, Legend: "{{volume}} write"},
				{Expr: `rate(windows_logical_disk_read_bytes_total{` + volSel + `}[$__rate_interval])`, Legend: "{{volume}} read"},
			},
			Overrides: []any{negativeY(".* read$")},
		}),
		timeseries(grid(12, y, 12, 8), TS{
			Title: "Disk IOPS", Unit: "iops", Desc: "Reads are drawn below the axis.",
			Targets: []T{
				{Expr: `rate(windows_logical_disk_writes_total{` + volSel + `}[$__rate_interval])`, Legend: "{{volume}} write"},
				{Expr: `rate(windows_logical_disk_reads_total{` + volSel + `}[$__rate_interval])`, Legend: "{{volume}} read"},
			},
			Overrides: []any{negativeY(".* read$")},
		}),
	)
	y += 8
	add(
		timeseries(grid(0, y, 12, 8), TS{
			Title: "Disk latency", Unit: "s",
			Desc: "Average time per read and write operation. Reads are drawn below the axis.",
			Targets: []T{
				{Expr: `rate(windows_logical_disk_write_seconds_total{` + volSel + `}[$__rate_interval]) / rate(windows_logical_disk_writes_total{` + volSel + `}[$__rate_interval])`, Legend: "{{volume}} write"},
				{Expr: `rate(windows_logical_disk_read_seconds_total{` + volSel + `}[$__rate_interval]) / rate(windows_logical_disk_reads_total{` + volSel + `}[$__rate_interval])`, Legend: "{{volume}} read"},
			},
			Overrides: []any{negativeY(".* read$")},
		}),
		timeseries(grid(12, y, 12, 8), TS{
			Title: "Disk queue length", Unit: "short",
			Desc: "Average number of read and write requests queued for the volume. Reads are drawn below the axis. A queue that stays high while latency rises indicates a disk bottleneck. The avg_*_requests_queued metrics grow like counters, so the panel takes their rate.",
			Targets: []T{
				{Expr: `rate(windows_logical_disk_avg_write_requests_queued{` + volSel + `}[$__rate_interval])`, Legend: "{{volume}} write"},
				{Expr: `rate(windows_logical_disk_avg_read_requests_queued{` + volSel + `}[$__rate_interval])`, Legend: "{{volume}} read"},
			},
			Overrides: []any{negativeY(".* read$")},
		}),
	)
	y += 8

	// ================================================================ Network
	tab("Network")
	add(
		timeseries(grid(0, y, 12, 8), TS{
			Title: "Network throughput", Unit: "bps", Desc: "Received traffic is drawn below the axis.",
			Targets: []T{
				{Expr: `rate(windows_net_bytes_sent_total{` + nicSel + `}[$__rate_interval]) * 8`, Legend: "{{nic}} sent"},
				{Expr: `rate(windows_net_bytes_received_total{` + nicSel + `}[$__rate_interval]) * 8`, Legend: "{{nic}} received"},
			},
			Overrides: []any{negativeY(".* received$")},
		}),
		timeseries(grid(12, y, 12, 8), TS{
			Title: "Network utilization", Unit: "percentunit", Min: 0, Steps: pctSteps(),
			Desc: "Sent plus received traffic relative to the link speed reported by the adapter. Interfaces without a link speed are not shown.",
			Targets: []T{{
				Expr:   `rate(windows_net_bytes_total{` + nicSel + `}[$__rate_interval]) / (windows_net_current_bandwidth_bytes{` + nicSel + `} > 0)`,
				Legend: "{{nic}}",
			}},
		}),
	)
	y += 8
	add(
		timeseries(grid(0, y, 12, 8), TS{
			Title: "Packets", Unit: "pps", Desc: "Received packets are drawn below the axis.",
			Targets: []T{
				{Expr: `rate(windows_net_packets_sent_total{` + nicSel + `}[$__rate_interval])`, Legend: "{{nic}} sent"},
				{Expr: `rate(windows_net_packets_received_total{` + nicSel + `}[$__rate_interval])`, Legend: "{{nic}} received"},
			},
			Overrides: []any{negativeY(".* received$")},
		}),
		timeseries(grid(12, y, 12, 8), TS{
			Title: "Network errors and discards", Unit: "pps", Min: 0,
			Targets: []T{
				{Expr: `rate(windows_net_packets_received_errors_total{` + nicSel + `}[$__rate_interval])`, Legend: "{{nic}} received errors"},
				{Expr: `rate(windows_net_packets_received_discarded_total{` + nicSel + `}[$__rate_interval])`, Legend: "{{nic}} received discarded"},
				{Expr: `rate(windows_net_packets_outbound_errors_total{` + nicSel + `}[$__rate_interval])`, Legend: "{{nic}} outbound errors"},
				{Expr: `rate(windows_net_packets_outbound_discarded_total{` + nicSel + `}[$__rate_interval])`, Legend: "{{nic}} outbound discarded"},
				{Expr: `rate(windows_net_packets_received_unknown_total{` + nicSel + `}[$__rate_interval])`, Legend: "{{nic}} received unknown protocol"},
			},
		}),
	)
	y += 8

	// ================================================================ Services
	tab("Services")
	add(
		stat(grid(0, y, 6, 4), ST{
			Title: "Running services", Unit: "none",
			Expr: `sum(windows_service_state{` + H + `, state="running"})`,
		}),
		stat(grid(6, y, 6, 4), ST{
			Title: "Stopped auto services", Desc: "Services with start mode \"auto\" that are not running.",
			Unit: "none", Steps: steps("green", nil, "orange", 1),
			Expr: `count(windows_service_state{` + H + `, state="running"} == 0 and on (name) windows_service_start_mode{` + H + `, start_mode="auto"} == 1) or vector(0)`,
		}),
		stat(grid(12, y, 6, 4), ST{
			Title: "Pending services", Desc: "Services in a start, stop, pause or continue pending state. A service that stays pending is hung.",
			Unit: "none", Steps: steps("green", nil, "orange", 1),
			Expr: `sum(windows_service_state{` + H + `, state=~".* pending"}) or vector(0)`,
		}),
		stat(grid(18, y, 6, 4), ST{
			Title: "Disabled services", Unit: "none",
			Expr: `sum(windows_service_start_mode{` + H + `, start_mode="disabled"})`,
		}),
	)
	y += 4
	add(
		table(grid(0, y, 12, 10), TB{
			Title: "Automatic services not running",
			Desc:  "Services configured to start automatically that are not in the running state. Delayed-start and trigger-start services may show up here until they start.",
			Targets: []T{{
				Expr: `max by (name, state) (windows_service_state{` + H + `} == 1) and on (name) (windows_service_state{` + H + `, state="running"} == 0) and on (name) (windows_service_start_mode{` + H + `, start_mode="auto"} == 1)`,
			}},
			Exclude: []string{"Value"},
			Order:   M{"name": 0, "state": 1},
			Rename:  M{"name": "Service", "state": "State"},
			SortBy:  []any{M{"desc": false, "displayName": "Service"}},
		}),
		statusTimeline(grid(12, y, 12, 10), "Service state changes",
			"Services that started or stopped in the selected time range.",
			T{
				Expr:   `windows_service_state{` + H + `, state="running"} and on (name) (changes(windows_service_state{` + H + `, state="running"}[$__range] @ end()) > 0)`,
				Legend: "{{name}}",
			},
			"running", "not running"),
	)
	y += 10

	// ================================================================ GPU (optional collector)
	// Engine time is per process and engine. Like Task Manager, utilization of an engine type is
	// the busiest engine of that type, and the GPU utilization is the busiest engine overall.
	gpuEngine := `sum by (luid, eng, engtype) (rate(windows_gpu_engine_time_seconds{` + H + `}[$__rate_interval]))`
	gpuName := `max by (luid, name) (windows_gpu_info{` + H + `})`
	gpuProc := `max by (process_id) (sum by (process_id, luid, eng) (rate(windows_gpu_engine_time_seconds{` + H + `}[$__rate_interval])))`
	gpuProcMem := `sum by (process_id) (windows_gpu_process_memory_dedicated_bytes{` + H + `})`
	// withProcessName adds the process name when the process collector is enabled and keeps the PID otherwise.
	withProcessName := func(expr string) string {
		return `(` + expr + ` * on (process_id) group_left (process) max by (process_id, process) (windows_process_info{` + H + `})) or on (process_id) ` + expr
	}

	tab("GPU")
	add(table(grid(0, y, 24, 5), TB{
		Title: "GPUs",
		Desc:  "Needs the gpu collector, which is not enabled by default. Utilization is the busiest engine of the GPU.",
		Targets: []T{
			{Expr: gpuName},
			{Expr: `clamp_max(max by (luid) (` + gpuEngine + `), 1)`},
			{Expr: `max by (luid) (windows_gpu_adapter_memory_dedicated_bytes{` + H + `})`},
			{Expr: `max by (luid) (windows_gpu_dedicated_video_memory_size_bytes{` + H + `})`},
			{Expr: `max by (luid) (windows_gpu_adapter_memory_shared_bytes{` + H + `})`},
			{Expr: `max by (luid) (windows_gpu_shared_system_memory_size_bytes{` + H + `})`},
		},
		JoinBy:  "luid",
		Exclude: []string{"luid", "Value #A"},
		Order:   M{"name": 0, "Value #B": 1, "Value #C": 2, "Value #D": 3, "Value #E": 4, "Value #F": 5},
		Rename: M{
			"name": "GPU", "Value #B": "Utilization", "Value #C": "Dedicated memory used", "Value #D": "Dedicated memory",
			"Value #E": "Shared memory used", "Value #F": "Shared memory",
		},
		Overrides: []any{
			byName("Value #B", gaugeCell("percentunit", 1)...),
			byName("Value #C", unitCell("bytes")...),
			byName("Value #D", unitCell("bytes")...),
			byName("Value #E", unitCell("bytes")...),
			byName("Value #F", unitCell("bytes")...),
		},
	}))
	y += 5
	add(
		timeseries(grid(0, y, 12, 8), TS{
			Title: "GPU utilization by engine type", Unit: "percentunit", Min: 0, Max: 1,
			Desc: "Busiest engine of each engine type, as in Task Manager.",
			Targets: []T{{
				Expr:   `clamp_max(max by (luid, engtype) (` + gpuEngine + `), 1) * on (luid) group_left (name) ` + gpuName,
				Legend: "{{name}} {{engtype}}",
			}},
		}),
		timeseries(grid(12, y, 12, 8), TS{
			Title: "GPU memory", Unit: "bytes", Min: 0,
			Desc: "Dedicated (video) and shared (system) memory in use, against the size of each pool.",
			Targets: []T{
				{Expr: `max by (luid) (windows_gpu_adapter_memory_dedicated_bytes{` + H + `}) * on (luid) group_left (name) ` + gpuName, Legend: "{{name}} dedicated used"},
				{Expr: `max by (luid) (windows_gpu_dedicated_video_memory_size_bytes{` + H + `}) * on (luid) group_left (name) ` + gpuName, Legend: "{{name}} dedicated size"},
				{Expr: `max by (luid) (windows_gpu_adapter_memory_shared_bytes{` + H + `}) * on (luid) group_left (name) ` + gpuName, Legend: "{{name}} shared used"},
				{Expr: `max by (luid) (windows_gpu_shared_system_memory_size_bytes{` + H + `}) * on (luid) group_left (name) ` + gpuName, Legend: "{{name}} shared size"},
			},
			Overrides: []any{byRegexp(".* size", prop("custom.fillOpacity", 0), prop("custom.lineStyle", M{"dash": []int{10, 10}, "fill": "dash"}))},
		}),
	)
	y += 8
	add(
		timeseries(grid(0, y, 12, 8), TS{
			Title: "Top 10 processes by GPU utilization", Unit: "percentunit", Min: 0, SortBy: "Mean",
			Desc:    "Busiest GPU engine used by each process. Process names need the process collector; otherwise only the PID is shown.",
			Targets: []T{{Expr: `topk(10, ` + withProcessName(gpuProc) + `)`, Legend: "{{process}} ({{process_id}})"}},
		}),
		timeseries(grid(12, y, 12, 8), TS{
			Title: "Top 10 processes by dedicated GPU memory", Unit: "bytes", Min: 0, SortBy: "Mean",
			Desc:    "Process names need the process collector; otherwise only the PID is shown.",
			Targets: []T{{Expr: `topk(10, ` + withProcessName(gpuProcMem) + `)`, Legend: "{{process}} ({{process_id}})"}},
		}),
	)
	y += 8

	// ================================================================ Hyper-V (optional collector)
	tab("Hyper-V / Host")
	add(
		stat(grid(0, y, 4, 4), ST{
			Title: "Virtual machines", Unit: "none", Desc: "Needs the hyperv collector, which is not enabled by default.",
			Expr: `sum(windows_hyperv_virtual_machine_health_total_count{` + H + `})`,
		}),
		stat(grid(4, y, 4, 4), ST{
			Title: "VMs in critical health", Unit: "none", Steps: steps("green", nil, "red", 1),
			Expr: `sum(windows_hyperv_virtual_machine_health_total_count{` + H + `, state="critical"})`,
		}),
		stat(grid(8, y, 4, 4), ST{
			Title: "Hyper-V WMI", Steps: steps("red", nil, "green", 1),
			Desc:     "Whether the Hyper-V WMI namespace answers. When it stops answering, Hyper-V Manager and Failover Cluster Manager usually cannot manage the host.",
			Expr:     `windows_hyperv_wmi_health{` + H + `}`,
			Mappings: []any{M{"type": "value", "options": M{"0": M{"text": "not responding", "index": 0}, "1": M{"text": "ok", "index": 1}}}},
		}),
		stat(grid(12, y, 4, 4), ST{
			Title: "Logical processors", Unit: "none",
			Expr: `windows_hyperv_host_logical_processor_count{` + H + `}`,
		}),
		stat(grid(16, y, 4, 4), ST{
			Title: "Virtual processors", Unit: "none", Desc: "Virtual processors assigned to all VMs.",
			Expr: `windows_hyperv_total_vm_processor_count{` + H + `}`,
		}),
		stat(grid(20, y, 4, 4), ST{
			Title: "vCPU per logical CPU", Unit: "none", Decimals: 2,
			Desc: "Virtual processors assigned to VMs per logical processor of the host.",
			Expr: `windows_hyperv_total_vm_processor_count{` + H + `} / windows_hyperv_host_logical_processor_count{` + H + `}`,
		}),
	)
	y += 4
	add(
		timeseries(grid(0, y, 12, 8), TS{
			Title: "Host CPU time by state", Unit: "percentunit", Min: 0, Max: 1, Stack: true, Fill: 40,
			Desc: "Share of all logical processors spent running guest code (VMs and the root partition) and in the hypervisor. With Hyper-V enabled, the CPU tab only sees the root partition.",
			Targets: []T{{
				Expr:   `sum by (state) (rate(windows_hyperv_hypervisor_logical_processor_time_total{` + H + `, state!="idle"}[$__rate_interval])) / scalar(count(windows_hyperv_hypervisor_logical_processor_time_total{` + H + `, state="idle"}))`,
				Legend: "{{state}}",
			}},
		}),
		timeseries(grid(12, y, 12, 8), TS{
			Title: "CPU usage per VM", Unit: "percentunit", Min: 0, Stack: true, Fill: 40, SortBy: "Mean",
			Desc: "Virtual processor run time of each VM as a share of all logical processors of the host.",
			Targets: []T{{
				Expr:   `sum by (vm) (rate(windows_hyperv_hypervisor_virtual_processor_run_time_total{` + H + `}[$__rate_interval])) / scalar(max(windows_hyperv_host_logical_processor_count{` + H + `}))`,
				Legend: "{{vm}}",
			}},
		}),
	)
	y += 8

	tab("Hyper-V / Memory")
	add(
		timeseries(grid(0, y, 8, 8), TS{
			Title: "Memory assigned to VMs", Unit: "bytes", Min: 0, Stack: true, Fill: 40, SortBy: "Mean",
			Desc:    "Physical memory currently assigned to each VM with dynamic memory.",
			Targets: []T{{Expr: `windows_hyperv_dynamic_memory_vm_physical_bytes{` + H + `}`, Legend: "{{vm}}"}},
		}),
		timeseries(grid(8, y, 8, 8), TS{
			Title: "VM memory pressure", Unit: "percentunit", Min: 0, Steps: steps("green", nil, "red", 1), Thresholds: true,
			Desc:    "Memory a VM needs as a share of the memory it has. Above 100% the VM needs more memory than it is assigned.",
			Targets: []T{{Expr: `windows_hyperv_dynamic_memory_vm_pressure_current_ratio{` + H + `}`, Legend: "{{vm}}"}},
		}),
		timeseries(grid(16, y, 8, 8), TS{
			Title: "Memory available for VMs", Unit: "bytes", Min: 0, Legend: "list",
			Desc:    "Memory the dynamic memory balancer can still hand out to VMs.",
			Targets: []T{{Expr: `windows_hyperv_dynamic_memory_balancer_available_memory_bytes{` + H + `}`, Legend: "{{balancer}}"}},
		}),
	)
	y += 8

	tab("Hyper-V / Storage")
	add(
		timeseries(grid(0, y, 8, 8), TS{
			Title: "Virtual disk throughput", Unit: "Bps", Desc: "Reads are drawn below the axis.",
			Targets: []T{
				{Expr: `rate(windows_hyperv_virtual_storage_device_bytes_written{` + H + `}[$__rate_interval])`, Legend: "{{device}} write"},
				{Expr: `rate(windows_hyperv_virtual_storage_device_bytes_read{` + H + `}[$__rate_interval])`, Legend: "{{device}} read"},
			},
			Overrides: []any{negativeY(".* read$")},
		}),
		timeseries(grid(8, y, 8, 8), TS{
			Title: "Virtual disk IOPS", Unit: "iops", Desc: "Reads are drawn below the axis.",
			Targets: []T{
				{Expr: `rate(windows_hyperv_virtual_storage_device_operations_written_total{` + H + `}[$__rate_interval])`, Legend: "{{device}} write"},
				{Expr: `rate(windows_hyperv_virtual_storage_device_operations_read_total{` + H + `}[$__rate_interval])`, Legend: "{{device}} read"},
			},
			Overrides: []any{negativeY(".* read$")},
		}),
		timeseries(grid(16, y, 8, 8), TS{
			Title: "Virtual disk latency and errors", Unit: "s", Min: 0,
			Desc: "Average I/O latency of each virtual disk. windows_hyperv_virtual_storage_device_latency_seconds holds the raw, growing counter in 100 ns ticks, so the panel divides its rate by the rate of I/O operations. Errors use the right axis.",
			Targets: []T{
				{Expr: `rate(windows_hyperv_virtual_storage_device_latency_seconds{` + H + `}[$__rate_interval]) / (rate(windows_hyperv_virtual_storage_device_operations_read_total{` + H + `}[$__rate_interval]) + rate(windows_hyperv_virtual_storage_device_operations_written_total{` + H + `}[$__rate_interval])) / 1e7`, Legend: "{{device}} latency"},
				{Expr: `rate(windows_hyperv_virtual_storage_device_error_count_total{` + H + `}[$__rate_interval]) > 0`, Legend: "{{device}} errors"},
			},
			Overrides: []any{byRegexp(".* errors", prop("unit", "ops"), prop("custom.axisPlacement", "right"), fixedColor("red"))},
		}),
	)
	y += 8

	tab("Hyper-V / Network")
	add(
		timeseries(grid(0, y, 8, 8), TS{
			Title: "Virtual switch throughput", Unit: "bps", Desc: "Received traffic is drawn below the axis.",
			Targets: []T{
				{Expr: `rate(windows_hyperv_vswitch_bytes_sent_total{` + H + `}[$__rate_interval]) * 8`, Legend: "{{vswitch}} sent"},
				{Expr: `rate(windows_hyperv_vswitch_bytes_received_total{` + H + `}[$__rate_interval]) * 8`, Legend: "{{vswitch}} received"},
			},
			Overrides: []any{negativeY(".* received$")},
		}),
		timeseries(grid(8, y, 8, 8), TS{
			Title: "Virtual switch dropped packets", Unit: "pps", Min: 0,
			Targets: []T{
				{Expr: `rate(windows_hyperv_vswitch_dropped_packets_incoming_total{` + H + `}[$__rate_interval])`, Legend: "{{vswitch}} incoming"},
				{Expr: `rate(windows_hyperv_vswitch_dropped_packets_outcoming_total{` + H + `}[$__rate_interval])`, Legend: "{{vswitch}} outgoing"},
			},
		}),
		timeseries(grid(16, y, 8, 8), TS{
			Title: "VM network adapter throughput", Unit: "bps", Desc: "Received traffic is drawn below the axis.",
			Targets: []T{
				{Expr: `rate(windows_hyperv_virtual_network_adapter_sent_bytes_total{` + H + `}[$__rate_interval]) * 8`, Legend: "{{adapter}} sent"},
				{Expr: `rate(windows_hyperv_virtual_network_adapter_received_bytes_total{` + H + `}[$__rate_interval]) * 8`, Legend: "{{adapter}} received"},
			},
			Overrides: []any{negativeY(".* received$")},
		}),
	)
	y += 8

	// ================================================================ Time (optional collector)
	tab("Time")
	add(
		stat(grid(0, y, 6, 4), ST{
			Title: "Time zone", Desc: "Needs the time collector, which is not enabled by default.", Expr: `windows_time_timezone{` + H + `}`, Text: "timezone", Color: "none",
		}),
		stat(grid(6, y, 6, 4), ST{
			Title: "Clock sync source", Expr: `windows_time_clock_sync_source{` + H + `} == 1`, Text: "type", Color: "none",
		}),
		stat(grid(12, y, 6, 4), ST{
			Title: "NTP time sources", Desc: "Number of time sources the NTP client uses. 0 means the clock is not synchronized.",
			Unit: "none", Steps: steps("red", nil, "green", 1),
			Expr: `windows_time_ntp_client_time_sources{` + H + `}`,
		}),
		stat(grid(18, y, 6, 4), ST{
			Title: "Clock offset", Desc: "Absolute offset between the system clock and the selected time source.",
			Unit: "s", Graph: true, Steps: steps("green", nil, "orange", 0.5, "red", 1),
			Expr: `abs(windows_time_computed_time_offset_seconds{` + H + `})`,
		}),
	)
	y += 4
	add(timeseries(grid(0, y, 24, 8), TS{
		Title: "Clock offset and NTP round-trip delay", Unit: "s",
		Targets: []T{
			{Expr: `windows_time_computed_time_offset_seconds{` + H + `}`, Legend: "clock offset"},
			{Expr: `windows_time_ntp_round_trip_delay_seconds{` + H + `}`, Legend: "NTP round-trip delay"},
		},
	}))
	y += 8

	// ================================================================ Exporter
	tab("Exporter / Scrape")
	add(
		stat(grid(0, y, 4, 4), ST{
			Title: "Scrape status", Desc: "Result of the last scrape by Prometheus.",
			Expr: `up{` + H + `}`, Steps: steps("red", nil, "green", 1),
			Mappings: []any{M{"type": "value", "options": M{"0": M{"text": "down", "index": 0}, "1": M{"text": "up", "index": 1}}}},
		}),
		stat(grid(4, y, 4, 4), ST{
			Title: "Exporter version", Desc: "Builds without a version show the first seven characters of the Git revision.", Text: "version", Color: "none",
			// Development builds have no version label, so fall back to the short revision.
			Expr: `windows_exporter_build_info{` + H + `, version!=""} or on (instance) label_replace(windows_exporter_build_info{` + H + `}, "version", "$1", "revision", "(.{7}).*")`,
		}),
		stat(grid(8, y, 4, 4), ST{
			Title: "Exporter uptime", Unit: "dtdurations", Desc: "Time since the exporter process started. A short uptime on a host with a long uptime means the exporter restarted.",
			Expr: `time() - process_start_time_seconds{` + H + `}`, Steps: steps("orange", nil, "text", 3600),
		}),
		stat(grid(12, y, 4, 4), ST{
			Title: "Failed collectors", Unit: "none", Steps: steps("green", nil, "red", 1),
			Desc: "Collectors that failed at least once in the selected time range.",
			Expr: `count(min_over_time(windows_exporter_collector_success{` + H + `}[$__range]) == 0) or vector(0)`,
		}),
		stat(grid(16, y, 4, 4), ST{
			Title: "Timed-out collectors", Unit: "none", Steps: steps("green", nil, "red", 1),
			Desc: "Collectors that hit the collector timeout at least once in the selected time range.",
			Expr: `count(max_over_time(windows_exporter_collector_timeout{` + H + `}[$__range]) == 1) or vector(0)`,
		}),
		stat(grid(20, y, 4, 4), ST{
			Title: "Samples per scrape", Unit: "short", Graph: true,
			Expr: `scrape_samples_scraped{` + H + `}`,
		}),
	)
	y += 4
	add(
		timeseries(grid(0, y, 12, 8), TS{
			Title: "Scrape duration", Unit: "s", Min: 0,
			Desc: "Scrape duration measured by Prometheus and by the exporter. It must stay below the scrape timeout of the job.",
			Targets: []T{
				{Expr: `scrape_duration_seconds{` + H + `}`, Legend: "Prometheus"},
				{Expr: `windows_exporter_scrape_duration_seconds{` + H + `}`, Legend: "exporter"},
			},
		}),
		timeseries(grid(12, y, 12, 8), TS{
			Title: "HTTP requests to /metrics", Unit: "reqps", Min: 0,
			Desc: "Scrape requests by HTTP status code, and errors while gathering or encoding metrics. More requests than one per scrape interval mean several jobs or Prometheus servers scrape this exporter.",
			Targets: []T{
				{Expr: `sum by (code) (rate(promhttp_metric_handler_requests_total{` + H + `}[$__rate_interval])) > 0`, Legend: "HTTP {{code}}"},
				{Expr: `sum by (cause) (rate(promhttp_metric_handler_errors_total{` + H + `}[$__rate_interval])) > 0`, Legend: "{{cause}} errors"},
			},
			Overrides: []any{byRegexp(".* errors", fixedColor("red"))},
		}),
	)
	y += 8

	tab("Exporter / Collectors")
	add(
		table(grid(0, y, 8, 8), TB{
			Title: "Collectors",
			Targets: []T{
				{Expr: `windows_exporter_collector_success{` + H + `}`},
				{Expr: `windows_exporter_collector_timeout{` + H + `}`},
				{Expr: `windows_exporter_collector_duration_seconds{` + H + `}`},
			},
			JoinBy:  "collector",
			Exclude: []string{"instance", "job", "port", "__name__", "instance 1", "instance 2", "instance 3", "job 1", "job 2", "job 3", "port 1", "port 2", "port 3", "__name__ 1", "__name__ 2", "__name__ 3"},
			Order:   M{"collector": 0, "Value #A": 1, "Value #B": 2, "Value #C": 3},
			Rename:  M{"collector": "Collector", "Value #A": "Status", "Value #B": "Timeout", "Value #C": "Duration"},
			SortBy:  []any{M{"desc": true, "displayName": "Duration"}},
			Overrides: []any{
				byName("Value #A",
					prop("mappings", []any{M{"type": "value", "options": M{
						"0": M{"color": "red", "index": 0, "text": "failed"},
						"1": M{"color": "green", "index": 1, "text": "ok"},
					}}}),
					prop("custom.cellOptions", M{"type": "color-text"}),
				),
				byName("Value #B",
					prop("mappings", []any{M{"type": "value", "options": M{
						"0": M{"color": "green", "index": 0, "text": "no"},
						"1": M{"color": "red", "index": 1, "text": "yes"},
					}}}),
					prop("custom.cellOptions", M{"type": "color-text"}),
				),
				byName("Value #C", unitCell("s")...),
			},
		}),
		timeseries(grid(8, y, 16, 8), TS{
			Title: "Collector duration", Unit: "s", Min: 0, Stack: true, Fill: 40, SortBy: "Mean",
			Desc:    "Time each collector spent per scrape. Compare against the scrape timeout.",
			Targets: []T{{Expr: `windows_exporter_collector_duration_seconds{` + H + `}`, Legend: "{{collector}}"}},
		}),
	)
	y += 8
	add(
		statusTimeline(grid(0, y, 12, 9), "Collector failures",
			"Collectors that failed at least once in the selected time range.",
			T{
				Expr:   `windows_exporter_collector_success{` + H + `} and on (collector) (min_over_time(windows_exporter_collector_success{` + H + `}[$__range] @ end()) == 0)`,
				Legend: "{{collector}}",
			},
			"ok", "failed"),
		statusTimeline(grid(12, y, 12, 9), "Collector timeouts",
			"Collectors that hit the collector timeout at least once in the selected time range.",
			T{
				Expr:   `1 - windows_exporter_collector_timeout{` + H + `} and on (collector) (max_over_time(windows_exporter_collector_timeout{` + H + `}[$__range] @ end()) == 1)`,
				Legend: "{{collector}}",
			},
			"in time", "timed out"),
	)
	y += 9

	tab("Exporter / Process and Go runtime")
	add(
		timeseries(grid(0, y, 8, 8), TS{
			Title: "Exporter CPU usage", Unit: "percentunit", Min: 0, Legend: "list",
			Desc: "CPU time used by the exporter process, as a share of one core. CPU usage grows with the number of scrapes, for example when two Prometheus jobs scrape the same exporter; CPU per scrape (right axis) does not.",
			Targets: []T{
				{Expr: `rate(process_cpu_seconds_total{` + H + `}[$__rate_interval])`, Legend: "CPU"},
				{Expr: `rate(process_cpu_seconds_total{` + H + `}[$__rate_interval]) / on (instance) sum by (instance) (rate(promhttp_metric_handler_requests_total{` + H + `}[$__rate_interval]))`, Legend: "CPU per scrape"},
			},
			Overrides: []any{byName("CPU per scrape", prop("unit", "s"), prop("custom.axisPlacement", "right"))},
		}),
		timeseries(grid(8, y, 8, 8), TS{
			Title: "Exporter memory", Unit: "bytes", Min: 0,
			Desc: "Working set and private bytes of the exporter process, and the Go heap in use. Steady growth points to a leak.",
			Targets: []T{
				{Expr: `process_resident_memory_bytes{` + H + `}`, Legend: "working set"},
				{Expr: `process_virtual_memory_bytes{` + H + `}`, Legend: "private bytes"},
				{Expr: `go_memstats_heap_inuse_bytes{` + H + `}`, Legend: "Go heap in use"},
			},
		}),
		timeseries(grid(16, y, 8, 8), TS{
			Title: "Exporter handles", Unit: "short", Min: 0, Legend: "list",
			Desc:    "Open handles of the exporter process (process_open_fds). Steady growth points to a handle leak.",
			Targets: []T{{Expr: `process_open_fds{` + H + `}`, Legend: "open handles"}},
		}),
	)
	y += 8
	// The go_sched_* metrics need an exporter that enables the Go scheduler metrics; older builds show no data there.
	add(
		timeseries(grid(0, y, 8, 8), TS{
			Title: "Goroutines by state", Unit: "short", Min: 0, Stack: true, Fill: 40,
			Desc: "Steady growth of goroutines points to a goroutine leak. Goroutines that stay \"not in Go\" sit in a system call, for example a collector waiting on WMI or PDH. The states need an exporter build with the Go scheduler metrics; the total line works with every build.",
			Targets: []T{
				{Expr: `go_sched_goroutines_running_goroutines{` + H + `}`, Legend: "running"},
				{Expr: `go_sched_goroutines_runnable_goroutines{` + H + `}`, Legend: "runnable"},
				{Expr: `go_sched_goroutines_waiting_goroutines{` + H + `}`, Legend: "waiting"},
				{Expr: `go_sched_goroutines_not_in_go_goroutines{` + H + `}`, Legend: "not in Go"},
				{Expr: `go_goroutines{` + H + `}`, Legend: "total"},
			},
			Overrides: []any{byName("total", fixedColor("text"), prop("custom.stacking", M{"group": "B", "mode": "none"}), prop("custom.fillOpacity", 0), prop("custom.lineStyle", M{"dash": []int{10, 10}, "fill": "dash"}))},
		}),
		timeseries(grid(8, y, 8, 8), TS{
			Title: "Goroutines created and threads", Unit: "short", Min: 0,
			Desc: "Goroutines started per second, and OS threads owned by the Go runtime. Goroutines created needs an exporter build with the Go scheduler metrics.",
			Targets: []T{
				{Expr: `rate(go_sched_goroutines_created_goroutines_total{` + H + `}[$__rate_interval])`, Legend: "goroutines created/s"},
				{Expr: `go_threads{` + H + `}`, Legend: "OS threads"},
			},
			Overrides: []any{byName("OS threads", prop("custom.axisPlacement", "right"))},
		}),
		timeseries(grid(16, y, 8, 8), TS{
			Title: "Scheduler latency", Unit: "s", Min: 0,
			Desc: "Time goroutines wait in the runnable state before they run. High values mean the exporter does not get enough CPU. Needs an exporter build with the Go scheduler metrics.",
			Targets: []T{
				{Expr: `histogram_quantile(0.99, sum by (le) (rate(go_sched_latencies_seconds_bucket{` + H + `}[$__rate_interval])))`, Legend: "p99"},
				{Expr: `histogram_quantile(0.5, sum by (le) (rate(go_sched_latencies_seconds_bucket{` + H + `}[$__rate_interval])))`, Legend: "p50"},
			},
		}),
	)
	y += 8
	add(
		timeseries(grid(0, y, 12, 8), TS{
			Title: "Garbage collection", Unit: "s", Min: 0,
			Desc: "Average GC pause and GC runs per second.",
			Targets: []T{
				{Expr: `rate(go_gc_duration_seconds_sum{` + H + `}[$__rate_interval]) / rate(go_gc_duration_seconds_count{` + H + `}[$__rate_interval])`, Legend: "average pause"},
				{Expr: `rate(go_gc_duration_seconds_count{` + H + `}[$__rate_interval])`, Legend: "GC runs"},
			},
			Overrides: []any{byName("GC runs", prop("unit", "ops"), prop("custom.axisPlacement", "right"))},
		}),
		timeseries(grid(12, y, 12, 8), TS{
			Title: "Allocation rate", Unit: "Bps", Min: 0, Legend: "list",
			Desc:    "Bytes allocated on the Go heap per second. Spikes line up with expensive collectors.",
			Targets: []T{{Expr: `rate(go_memstats_alloc_bytes_total{` + H + `}[$__rate_interval])`, Legend: "allocated"}},
		}),
	)

	return M{
		"__inputs": []any{},
		"__requires": []any{
			M{"type": "grafana", "id": "grafana", "name": "Grafana", "version": "11.0.0"},
			M{"type": "datasource", "id": "prometheus", "name": "Prometheus", "version": "1.0.0"},
			M{"type": "panel", "id": "stat", "name": "Stat", "version": ""},
			M{"type": "panel", "id": "table", "name": "Table", "version": ""},
			M{"type": "panel", "id": "timeseries", "name": "Time series", "version": ""},
		},
		"annotations": M{"list": []any{M{
			"builtIn": 1, "datasource": M{"type": "grafana", "uid": "-- Grafana --"}, "enable": true, "hide": true,
			"iconColor": "rgba(0, 211, 255, 1)", "name": "Annotations & Alerts", "type": "dashboard",
		}, M{
			// The boot timestamp only changes on reboot, so each distinct value marks one reboot.
			"datasource":      datasource(),
			"enable":          true,
			"hide":            false,
			"iconColor":       "orange",
			"name":            "Reboots",
			"expr":            `windows_system_boot_time_timestamp{` + H + `} * 1000 > $__from < $__to`,
			"step":            "1m",
			"tagKeys":         "instance",
			"textFormat":      "",
			"titleFormat":     "Reboot",
			"useValueForTime": "on",
		}}},
		"description":          "Fleet overview and per-host details for Windows hosts monitored by windows_exporter.",
		"editable":             true,
		"fiscalYearStartMonth": 0,
		"graphTooltip":         1,
		"id":                   nil,
		"links": []any{M{
			"asDropdown": false, "icon": "doc", "includeVars": false, "keepTime": false, "tags": []any{},
			"targetBlank": true, "title": "windows_exporter", "tooltip": "", "type": "link",
			"url": "https://github.com/prometheus-community/windows_exporter",
		}},
		"panels":        panels,
		"refresh":       "1m",
		"schemaVersion": 41,
		"tags":          []string{"prometheus", "windows", "windows_exporter"},
		"templating": M{"list": []any{
			M{
				"type": "datasource", "name": "datasource", "label": "Data source", "query": "prometheus",
				"current": M{}, "hide": 0, "includeAll": false, "multi": false, "options": []any{}, "refresh": 1, "regex": "",
			},
			variable("job", "Job", `label_values(windows_os_hostname, job)`, true, true, false, ""),
			variable("hostname", "Hostname", `label_values(windows_os_hostname{job=~"$job"}, hostname)`, true, true, false, ""),
			variable("instance", "Instance", `label_values(windows_os_hostname{job=~"$job", hostname=~"$hostname"}, instance)`, false, false, false, ""),
			variable("show_hostname", "", `label_values(windows_os_hostname{job=~"$job", instance="$instance"}, hostname)`, false, false, true, ""),
			variable("volume", "Volume", `label_values(windows_logical_disk_size_bytes{job=~"$job", instance="$instance"}, volume)`, true, true, false, `/^(?!HarddiskVolume).+$/`),
			variable("nic", "Network interface", `label_values(windows_net_bytes_total{job=~"$job", instance="$instance"}, nic)`, true, true, false, `/^(?!isatap|Teredo|6to4).+$/`),
		}},
		"time":       M{"from": "now-3h", "to": "now"},
		"timepicker": M{},
		"timezone":   "",
		"title":      "Windows Exporter",
		"uid":        "Kdaassddw",
		"version":    1,
		"weekStart":  "",
	}
}
