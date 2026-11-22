// SPDX-License-Identifier: MIT
//
// Copyright The Prometheus Authors

//go:build windows

package ad

import (
	"testing"
	"time"

	"github.com/prometheus-community/windows_exporter/internal/headers/ntdsapi"
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"github.com/stretchr/testify/require"
)

func TestPartnerFromDsaDN(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		dn, partner, site string
	}{
		{
			dn:      "CN=NTDS Settings,CN=DC02,CN=Servers,CN=Default-First-Site-Name,CN=Sites,CN=Configuration,DC=example,DC=com",
			partner: "DC02",
			site:    "Default-First-Site-Name",
		},
		{
			// A deleted DSA keeps the server and site of its successor, so it
			// keeps the full DN.
			dn:      `CN=NTDS Settings\0ADEL:5b3c1d2e-0000-0000-0000-000000000000,CN=DC03,CN=Servers,CN=Branch,CN=Sites,CN=Configuration,DC=example,DC=com`,
			partner: `CN=NTDS Settings\0ADEL:5b3c1d2e-0000-0000-0000-000000000000,CN=DC03,CN=Servers,CN=Branch,CN=Sites,CN=Configuration,DC=example,DC=com`,
		},
		{
			dn:      `CN=NTDS Settings,CN=DC04,CN=Servers,CN=Berlin\, Mitte,CN=Sites,CN=Configuration,DC=example,DC=com`,
			partner: "DC04",
			site:    "Berlin, Mitte",
		},
		{
			dn:      "cn=NTDS Settings,cn=dc05,cn=servers,cn=HQ,cn=sites,cn=Configuration,dc=example,dc=com",
			partner: "dc05",
			site:    "HQ",
		},
		{dn: "CN=Unexpected,DC=example,DC=com", partner: "CN=Unexpected,DC=example,DC=com"},
		{dn: "", partner: ""},
	} {
		partner, site := partnerFromDsaDN(tc.dn)
		require.Equal(t, tc.partner, partner, tc.dn)
		require.Equal(t, tc.site, site, tc.dn)
	}
}

func TestWriteReplicationPartners(t *testing.T) {
	t.Parallel()

	const (
		domainNC = "DC=example,DC=com"
		configNC = "CN=Configuration,DC=example,DC=com"
		dc02     = "CN=NTDS Settings,CN=DC02,CN=Servers,CN=HQ,CN=Sites,CN=Configuration,DC=example,DC=com"
		dc02Old  = `CN=NTDS Settings\0ADEL:5b3c1d2e-0000-0000-0000-000000000000,CN=DC02,CN=Servers,CN=HQ,CN=Sites,CN=Configuration,DC=example,DC=com`
	)

	success := time.Date(2026, 11, 22, 21, 0, 0, 500_000_000, time.UTC)
	attempt := time.Date(2026, 11, 22, 22, 0, 0, 0, time.UTC)

	neighbors := []ntdsapi.Neighbor{
		{
			// Same server and site as the live DC02, from a deleted DSA. It
			// comes first, so the live entry must not depend on the order.
			NamingContext:  domainNC,
			SourceDsaDN:    dc02Old,
			LastSyncResult: 8419, // ERROR_DS_CANT_FIND_DSA_OBJ.
		},
		{
			NamingContext:           domainNC,
			SourceDsaDN:             dc02,
			LastSyncSuccess:         success,
			LastSyncAttempt:         attempt,
			LastSyncResult:          1722, // RPC_S_SERVER_UNAVAILABLE.
			ConsecutiveSyncFailures: 3,
		},
		{
			// Never replicated: no timestamps.
			NamingContext: configNC,
			SourceDsaDN:   dc02,
		},
	}

	c := New(nil)
	c.buildDescriptors()

	registry := prometheus.NewPedanticRegistry()
	require.NoError(t, registry.Register(replicationCollector{collector: c, neighbors: neighbors}))

	families, err := registry.Gather()
	require.NoError(t, err)

	byName := make(map[string]*dto.MetricFamily, len(families))
	for _, family := range families {
		require.NotEmpty(t, family.GetHelp(), "metric %s has no HELP", family.GetName())
		require.Equal(t, dto.MetricType_GAUGE, family.GetType(), family.GetName())
		byName[family.GetName()] = family
	}

	dc02Labels := map[string]string{"naming_context": domainNC, "partner": "DC02", "partner_site": "HQ"}
	configLabels := map[string]string{"naming_context": configNC, "partner": "DC02", "partner_site": "HQ"}
	oldLabels := map[string]string{"naming_context": domainNC, "partner": dc02Old, "partner_site": ""}

	for _, tc := range []struct {
		name   string
		labels map[string]string
		value  float64
	}{
		{"windows_ad_replication_partner_consecutive_failures", dc02Labels, 3},
		{"windows_ad_replication_partner_last_result", dc02Labels, 1722},
		{"windows_ad_replication_partner_last_success_timestamp_seconds", dc02Labels, 1795381200.5},
		{"windows_ad_replication_partner_last_attempt_timestamp_seconds", dc02Labels, 1795384800},
		{"windows_ad_replication_partner_consecutive_failures", configLabels, 0},
		{"windows_ad_replication_partner_last_result", configLabels, 0},
		{"windows_ad_replication_partner_last_result", oldLabels, 8419},
	} {
		family, ok := byName[tc.name]
		require.True(t, ok, "metric %s not published", tc.name)
		require.InDelta(t, tc.value, findMetric(t, family, tc.labels).GetGauge().GetValue(), 1e-6, "%s%v", tc.name, tc.labels)
	}

	require.Len(t, byName["windows_ad_replication_partner_consecutive_failures"].GetMetric(), 3)
	// Timestamps are absent for the naming context that never replicated.
	require.Len(t, byName["windows_ad_replication_partner_last_success_timestamp_seconds"].GetMetric(), 1)
	require.Len(t, byName["windows_ad_replication_partner_last_attempt_timestamp_seconds"].GetMetric(), 1)
}

func TestBuildRejectsUnknownSubCollector(t *testing.T) {
	t.Parallel()

	c := New(&Config{CollectorsEnabled: []string{"metrics", "replication"}})
	require.ErrorContains(t, c.Build(nil, nil), "unknown sub collector: replication")
}

type replicationCollector struct {
	collector *Collector
	neighbors []ntdsapi.Neighbor
}

func (c replicationCollector) Describe(ch chan<- *prometheus.Desc) {
	prometheus.DescribeByCollect(c, ch)
}

func (c replicationCollector) Collect(ch chan<- prometheus.Metric) {
	c.collector.writeReplicationPartners(ch, c.neighbors)
}
