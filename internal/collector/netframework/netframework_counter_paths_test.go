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

package netframework

import (
	"reflect"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// These paths are independently listed by Netdata's native Windows perflib
// collector and Mono's CLR counter definitions. Remoting is absent from the
// current Microsoft .NET Framework performance-counter documentation.
// https://github.com/netdata/netdata/blob/master/src/collectors/windows.plugin/perflib-netframework.c
// https://github.com/mono/mono/blob/main/mono/metadata/mono-perfcounters-def.h
func TestClrRemotingCounterPaths(t *testing.T) {
	expected := map[string]string{
		"Channels":                       "Channels",
		"ContextBoundClassesLoaded":      "Context-Bound Classes Loaded",
		"ContextBoundObjectsAllocPersec": "Context-Bound Objects Alloc / sec",
		"ContextProxies":                 "Context Proxies",
		"Contexts":                       "Contexts",
		"TotalRemoteCalls":               "Total Remote Calls",
	}
	rowType := reflect.TypeFor[perfDataClrRemoting]()
	require.Equal(t, len(expected)+1, rowType.NumField())

	for field, counter := range expected {
		f, found := rowType.FieldByName(field)
		require.True(t, found, field)
		require.Equal(t, counter, f.Tag.Get("perfdata"), field)
	}
}

// Counter names also appear in the Windows counter inventory captured by:
// https://gist.github.com/hdansou/040207527657de0225b68912ff94400e
// Memory names are corroborated by Mono's primary CLR counter definitions.
func TestClrCounterPaths(t *testing.T) {
	cases := []struct {
		rowType  reflect.Type
		counters []string
	}{
		{reflect.TypeFor[perfDataClrExceptions](), []string{"# of Exceps Thrown", "# of Filters / sec", "# of Finallys / sec", "Throw To Catch Depth / sec"}},
		{reflect.TypeFor[perfDataClrInterop](), []string{"# of CCWs", "# of Stubs", "# of marshalling"}},
		{reflect.TypeFor[perfDataClrJIT](), []string{"# of Methods Jitted", "% Time in Jit", "Standard Jit Failures", "Total # of IL Bytes Jitted"}},
		{reflect.TypeFor[perfDataClrLoading](), []string{"Bytes in Loader Heap", "Current appdomains", "Current Assemblies", "Current Classes Loaded", "Total Appdomains", "Total appdomains unloaded", "Total Assemblies", "Total Classes Loaded", "Total # of Load Failures"}},
		{reflect.TypeFor[perfDataClrLocksAndThreads](), []string{"Current Queue Length", "# of current logical Threads", "# of current physical Threads", "# of current recognized threads", "# of total recognized threads", "Queue Length Peak", "Total # of Contentions"}},
		{reflect.TypeFor[perfDataClrMemory](), []string{"Allocated Bytes/sec", "Finalization Survivors", "Gen 0 heap size", "Gen 0 Promoted Bytes/Sec", "Gen 1 heap size", "Gen 1 Promoted Bytes/Sec", "Gen 2 heap size", "Large Object Heap size", "# GC Handles", "# Gen 0 Collections", "# Gen 1 Collections", "# Gen 2 Collections", "# Induced GC", "# of Pinned Objects", "# of Sink Blocks in use", "# Total committed Bytes", "# Total reserved Bytes", "% Time in GC", "% Time in GC,secondvalue", "Process ID"}},
		{reflect.TypeFor[perfDataClrRemoting](), []string{"Channels", "Context-Bound Classes Loaded", "Context-Bound Objects Alloc / sec", "Context Proxies", "Contexts", "Total Remote Calls"}},
		{reflect.TypeFor[perfDataClrSecurity](), []string{"# Link Time Checks", "% Time in RT checks", "Stack Walk Depth", "Total Runtime Checks"}},
	}
	for _, tc := range cases {
		t.Run(tc.rowType.Name(), func(t *testing.T) {
			actual := make([]string, 0, tc.rowType.NumField()-1)
			for field := range tc.rowType.Fields() {
				if tag, found := field.Tag.Lookup("perfdata"); found {
					actual = append(actual, strings.ToLower(tag))
				}
			}

			expected := make([]string, len(tc.counters))
			for i, counter := range tc.counters {
				expected[i] = strings.ToLower(counter)
			}

			require.ElementsMatch(t, expected, actual)
		})
	}
}
