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

package clusapi

import (
	"encoding/hex"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// CLUSPROP_FORMAT values that the collectors do not decode.
const (
	formatBinary        = 1
	formatMultiString   = 5
	formatULargeInteger = 6
	formatFileTime      = 12
)

// Property lists in testdata were captured with TestDumpPropertyLists from the
// CI cluster (Windows Server 2022). Unlike lists built from the documentation,
// they end with the extra CLUSPROP_SYNTAX_ENDMARK of the cluster service.
func readPropertyFixture(tb testing.TB, name string) []byte {
	tb.Helper()

	text, err := os.ReadFile(filepath.Join("testdata", name))
	if err != nil {
		tb.Fatal(err)
	}

	data, err := hex.DecodeString(strings.Join(strings.Fields(string(text)), ""))
	if err != nil {
		tb.Fatal(err)
	}

	return data
}

type fixtureValue struct {
	format uint16
	length int
	dword  *uint32
	text   *string
}

func dwordValue(value uint32) fixtureValue {
	return fixtureValue{format: formatDWORD, length: 4, dword: &value}
}

func stringValue(value string, length int) fixtureValue {
	return fixtureValue{format: formatString, length: length, text: &value}
}

func checkFixtureProperties(t *testing.T, properties map[string][]Property, count int, want map[string]fixtureValue) {
	t.Helper()

	if len(properties) != count {
		t.Errorf("properties = %d, want %d", len(properties), count)
	}

	for name, value := range want {
		values := properties[name]
		if len(values) != 1 || values[0].Format != value.format || len(values[0].Data) != value.length {
			t.Errorf("%s = %+v, want format %d with %d bytes", name, values, value.format, value.length)

			continue
		}

		if value.dword != nil {
			if got, err := values[0].DWORD(); err != nil || got != *value.dword {
				t.Errorf("%s = %d, %v, want %d", name, got, err, *value.dword)
			}
		}

		if value.text != nil {
			if got, err := values[0].String(); err != nil || got != *value.text {
				t.Errorf("%s = %q, %v, want %q", name, got, err, *value.text)
			}
		}
	}
}

func TestParsePropertiesCapturedResourceLists(t *testing.T) {
	// Read-only common properties: an SZ whose 38 bytes need 2 bytes of
	// padding, ULARGE_INTEGER values and a DWORD.
	properties, err := ParseProperties(readPropertyFixture(t, "resource_ro_common_generic_service.hex"))
	if err != nil {
		t.Fatal(err)
	}

	checkFixtureProperties(t, properties, 6, map[string]fixtureValue{
		"Name":                    stringValue("CI Generic Service", 38),
		"MonitorProcessId":        dwordValue(9092),
		"StatusInformation":       {format: formatULargeInteger, length: 8},
		"LastOperationStatusCode": {format: formatULargeInteger, length: 8},
		"ResourceSpecificData1":   {format: formatULargeInteger, length: 8},
		"ResourceSpecificData2":   {format: formatULargeInteger, length: 8},
	})

	// Common properties: the resource type name the collector publishes, empty
	// strings (2 bytes, padded) and the DWORD values.
	properties, err = ParseProperties(readPropertyFixture(t, "resource_common_generic_service.hex"))
	if err != nil {
		t.Fatal(err)
	}

	checkFixtureProperties(t, properties, 15, map[string]fixtureValue{
		"Type":                   stringValue("Generic Service", 32),
		"Description":            stringValue("", 2),
		"ResourceSpecificStatus": stringValue("", 2),
		"DeadlockTimeout":        dwordValue(300000),
		"EmbeddedFailureAction":  dwordValue(2),
		"IsAlivePollInterval":    dwordValue(^uint32(0)),
		"LooksAlivePollInterval": dwordValue(^uint32(0)),
		"PendingTimeout":         dwordValue(180000),
		"PersistentState":        dwordValue(1),
		"RestartAction":          dwordValue(2),
		"RestartDelay":           dwordValue(500),
		"RestartPeriod":          dwordValue(600000),
		"RestartThreshold":       dwordValue(1),
		"RetryPeriodOnFailure":   dwordValue(600000),
		"SeparateMonitor":        dwordValue(0),
	})

	// Read-only private properties of an IP Address: FILETIME values.
	properties, err = ParseProperties(readPropertyFixture(t, "resource_ro_private_ip_address.hex"))
	if err != nil {
		t.Fatal(err)
	}

	checkFixtureProperties(t, properties, 5, map[string]fixtureValue{
		"LeaseObtainedTime": {format: formatFileTime, length: 8},
		"LeaseExpiresTime":  {format: formatFileTime, length: 8},
		"DhcpServer":        stringValue("255.255.255.255", 32),
		"DhcpAddress":       stringValue("0.0.0.0", 16),
		"DhcpSubnetMask":    stringValue("255.0.0.0", 20),
	})
}

func TestParsePropertiesCapturedClusterList(t *testing.T) {
	// The cluster common list is larger than the 4 KiB first buffer and holds
	// BINARY security descriptors, MULTI_SZ lists, ULARGE_INTEGER and FILETIME
	// values next to 64 DWORDs.
	data := readPropertyFixture(t, "cluster_common.hex")
	if len(data) <= propertyListBufferSize {
		t.Fatalf("fixture has %d bytes, want more than the first buffer", len(data))
	}

	properties, err := ParseProperties(data)
	if err != nil {
		t.Fatal(err)
	}

	checkFixtureProperties(t, properties, 74, map[string]fixtureValue{
		"Description":                     stringValue("", 2),
		"PreferredSite":                   stringValue("", 2),
		"Security Descriptor":             {format: formatBinary, length: 396},
		"SharedVolumeSecurityDescriptor":  {format: formatBinary, length: 280},
		"SharedVolumeCompatibleFilters":   {format: formatMultiString, length: 2},
		"SharedVolumeIncompatibleFilters": {format: formatMultiString, length: 2},
		"EnabledEventLogs":                {format: formatMultiString, length: 354},
		"DumpPolicy":                      {format: formatULargeInteger, length: 8},
		"S2DCacheBehavior":                {format: formatULargeInteger, length: 8},
		"RecentEventsResetTime":           {format: formatFileTime, length: 8},
		"ClusterEnforcedAntiaffinity":     {format: formatDWORD, length: 4},
	})

	dwords := 0

	for _, values := range properties {
		if _, err := values[0].Value32(); err == nil {
			dwords++
		}
	}

	if dwords != 64 {
		t.Errorf("32-bit values = %d, want 64", dwords)
	}

	// Every truncation of a real list must be rejected without a panic.
	for _, name := range []string{"resource_ro_common_generic_service.hex", "resource_common_generic_service.hex"} {
		data := readPropertyFixture(t, name)
		for index := range len(data) - 4 {
			if _, err := ParseProperties(data[:index]); err == nil {
				t.Errorf("%s: accepted truncation at %d", name, index)
			}
		}
	}
}

func FuzzParsePropertiesCaptured(f *testing.F) {
	for _, name := range []string{"resource_ro_common_generic_service.hex", "resource_common_generic_service.hex", "resource_ro_private_ip_address.hex", "cluster_common.hex"} {
		f.Add(readPropertyFixture(f, name))
	}

	f.Fuzz(func(t *testing.T, data []byte) {
		properties, err := ParseProperties(data)
		if err != nil {
			return
		}

		for _, values := range properties {
			for _, value := range values {
				_, _ = value.String()
				_, _ = value.Value32()

				if (value.Format == formatDWORD || value.Format == formatLong) && len(value.Data) != 4 {
					t.Fatalf("32-bit value with %d bytes", len(value.Data))
				}
			}
		}
	})
}
