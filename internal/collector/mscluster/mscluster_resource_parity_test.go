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

package mscluster

import (
	"errors"
	"os"
	"reflect"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/prometheus-community/windows_exporter/internal/headers/clusapi"
	"github.com/prometheus-community/windows_exporter/internal/mi"
	"google.golang.org/protobuf/proto"
)

// This parity test runs against the CI failover-cluster fixture. Optional local
// hosts may lack the role; fixture-required hosts must never skip native failures.
func TestResourceNativeWMIParity(t *testing.T) {
	native, session, required := resourceComparisonSources(t)

	query, err := mi.NewQuery("SELECT Name,Type,OwnerGroup,OwnerNode,Characteristics,DeadlockTimeout,EmbeddedFailureAction,Flags,IsAlivePollInterval,LooksAlivePollInterval,MonitorProcessId,PendingTimeout,ResourceClass,RestartAction,RestartDelay,RestartPeriod,RestartThreshold,RetryPeriodOnFailure,State,Subclass FROM MSCluster_Resource")
	if err != nil {
		t.Fatal(err)
	}

	var rows []msClusterResource

	start := time.Now()

	if err := session.Query(&rows, mi.NamespaceRootMSCluster, query, time.Minute); err != nil {
		t.Fatal(err)
	}

	t.Logf("WMI resource query duration: %s", time.Since(start))

	// The CI cluster provisions online, offline and failed resources. Comparing
	// two empty sets proves nothing, so the fixture must provide resources.
	if len(rows) == 0 {
		if required {
			t.Fatal("cluster resource fixture has no resources")
		}

		t.Log("cluster has no resources; comparing empty sources")
	}

	expected := make([]clusapi.Resource, 0, len(rows))
	for _, row := range rows {
		resource := clusapi.Resource{Name: row.Name, Type: row.Type, OwnerGroup: row.OwnerGroup, OwnerNode: row.OwnerNode, IdentityValid: true, Values: make(map[string]uint32)}

		value := reflect.ValueOf(row)
		for index := 4; index < value.NumField(); index++ {
			resource.Values[value.Type().Field(index).Name] = uint32(value.Field(index).Uint())
		}

		expected = append(expected, resource)
	}

	start = time.Now()

	resources, err := native.Resources(time.Now().Add(time.Minute))
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("ClusAPI resource query duration: %s", time.Since(start))

	nodeQuery, err := mi.NewQuery("SELECT Name FROM MSCluster_Node")
	if err != nil {
		t.Fatal(err)
	}

	var nodes []struct {
		Name string `mi:"Name"`
	}
	if err := session.Query(&nodes, mi.NamespaceRootMSCluster, nodeQuery, time.Minute); err != nil {
		t.Fatal(err)
	}

	nodeNames := make([]string, 0, len(nodes))
	for _, node := range nodes {
		nodeNames = append(nodeNames, node.Name)
	}

	if len(nodeNames) == 0 {
		t.Fatal("cluster fixture has no nodes")
	}

	gotFamilies, err := gatherResources(t, resources, nil, nodeNames...)
	if err != nil {
		t.Fatal(err)
	}

	wantFamilies, err := gatherResources(t, expected, nil, nodeNames...)
	if err != nil {
		t.Fatal(err)
	}

	if len(gotFamilies) != len(wantFamilies) {
		t.Fatalf("families = %d, want %d", len(gotFamilies), len(wantFamilies))
	}

	for name, want := range wantFamilies {
		if !proto.Equal(gotFamilies[name], want) {
			t.Errorf("native/WMI metric mismatch: %s\nnative: %v\nWMI: %v", name, gotFamilies[name], want)
		}
	}

	if t.Failed() {
		logResourceWMINulls(t, session, query)
	}
}

// logResourceWMINulls shows which WMI properties are NULL. The WMI collector
// published NULL as 0, so a native value can only match if it is 0 as well.
func logResourceWMINulls(t *testing.T, session *mi.Session, query mi.Query) {
	t.Helper()

	err := session.QueryFunc(mi.NamespaceRootMSCluster, query, time.Minute, func(instance *mi.Instance) error {
		element, err := instance.GetElement("Name")
		if err != nil {
			return err
		}

		name, err := element.String()
		if err != nil {
			return err
		}

		var nulls []string

		for _, property := range []string{"Characteristics", "DeadlockTimeout", "EmbeddedFailureAction", "Flags", "IsAlivePollInterval", "LooksAlivePollInterval", "MonitorProcessId", "PendingTimeout", "ResourceClass", "RestartAction", "RestartDelay", "RestartPeriod", "RestartThreshold", "RetryPeriodOnFailure", "State", "Subclass", "Type", "OwnerGroup", "OwnerNode"} {
			element, err := instance.GetElement(property)
			if err != nil {
				return err
			}

			if element.IsNull() {
				nulls = append(nulls, property)
			}
		}

		t.Logf("WMI resource %q NULL properties: %v", name, nulls)

		return nil
	})
	if err != nil {
		t.Logf("WMI NULL diagnostics: %v", err)
	}
}

// Resource source comparisons require a live cluster; missing CI fixtures fail.
func resourceComparisonSources(tb testing.TB) (*clusapi.Cluster, *mi.Session, bool) {
	tb.Helper()

	required := slices.Contains(strings.Split(os.Getenv("WINDOWS_EXPORTER_TEST_COLLECTORS"), ","), Name)

	native, err := clusapi.Open()
	if err != nil {
		if !required && errors.Is(err, errors.ErrUnsupported) {
			tb.Skipf("failover cluster role unavailable: %v", err)
		}

		tb.Fatal(err)
	}

	tb.Cleanup(func() {
		if err := native.Close(); err != nil {
			tb.Error(err)
		}
	})

	app, err := mi.ApplicationInitialize()
	if err != nil {
		tb.Fatal(err)
	}

	tb.Cleanup(func() {
		if err := app.Close(); err != nil {
			tb.Error(err)
		}
	})

	session, err := app.NewSession(nil)
	if err != nil {
		tb.Fatal(err)
	}

	tb.Cleanup(func() {
		if err := session.Close(); err != nil {
			tb.Error(err)
		}
	})

	return native, session, required
}

func BenchmarkResourceSources(b *testing.B) {
	native, session, _ := resourceComparisonSources(b)

	query, err := mi.NewQuery("SELECT Name,Type,OwnerGroup,OwnerNode,Characteristics,DeadlockTimeout,EmbeddedFailureAction,Flags,IsAlivePollInterval,LooksAlivePollInterval,MonitorProcessId,PendingTimeout,ResourceClass,RestartAction,RestartDelay,RestartPeriod,RestartThreshold,RetryPeriodOnFailure,State,Subclass FROM MSCluster_Resource")
	if err != nil {
		b.Fatal(err)
	}

	var rows []msClusterResource
	if err := session.Query(&rows, mi.NamespaceRootMSCluster, query, time.Minute); err != nil {
		b.Fatal(err)
	}

	// Timing two empty enumerations says nothing about resource collection.
	if len(rows) == 0 {
		b.Skip("cluster has no resources")
	}

	b.Run("WMI", func(b *testing.B) {
		b.ReportAllocs()

		for b.Loop() {
			var rows []msClusterResource
			if err := session.Query(&rows, mi.NamespaceRootMSCluster, query, time.Minute); err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("ClusAPI", func(b *testing.B) {
		b.ReportAllocs()

		for b.Loop() {
			if _, err := native.Resources(time.Now().Add(time.Minute)); err != nil {
				b.Fatal(err)
			}
		}
	})
}
