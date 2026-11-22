// SPDX-License-Identifier: MIT
//
// Copyright The Prometheus Authors

//go:build windows

package ad

import (
	"errors"
	"fmt"
	"strings"

	"github.com/prometheus-community/windows_exporter/internal/headers/ntdsapi"
	"github.com/prometheus-community/windows_exporter/internal/types"
	"github.com/prometheus/client_golang/prometheus"
)

// buildReplicationPartnerDescriptors creates the descriptors of the
// replication_partners sub-collector.
func (c *Collector) buildReplicationPartnerDescriptors() {
	labels := []string{"naming_context", "partner", "partner_site"}

	c.replicationPartnerConsecutiveFailures = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "replication_partner_consecutive_failures"),
		"Number of failed inbound replication attempts of the naming context from the partner since the last successful one (repadmin /showrepl)",
		labels,
		nil,
	)
	c.replicationPartnerLastResult = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "replication_partner_last_result"),
		"Win32 error code of the last inbound replication attempt of the naming context from the partner, 0 on success",
		labels,
		nil,
	)
	c.replicationPartnerLastSuccess = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "replication_partner_last_success_timestamp_seconds"),
		"Time of the last successful inbound replication of the naming context from the partner, as a Unix timestamp. Absent if it never succeeded",
		labels,
		nil,
	)
	c.replicationPartnerLastAttempt = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "replication_partner_last_attempt_timestamp_seconds"),
		"Time of the last inbound replication attempt of the naming context from the partner, as a Unix timestamp. Absent if it was never attempted",
		labels,
		nil,
	)
}

// collectReplicationPartners publishes the inbound replication state of every
// naming context and source domain controller, as repadmin /showrepl shows it.
func (c *Collector) collectReplicationPartners(ch chan<- prometheus.Metric) error {
	if c.dsHandle == 0 {
		handle, err := ntdsapi.Bind(c.dcName)
		if err != nil {
			return fmt.Errorf("failed to bind to the directory service on %s: %w", c.dcName, err)
		}

		c.dsHandle = handle
	}

	neighbors, err := ntdsapi.InboundNeighbors(c.dsHandle)
	if err != nil {
		// A binding does not survive a restart of the directory service, so
		// bind again on the next scrape.
		unbindErr := ntdsapi.Unbind(c.dsHandle)
		c.dsHandle = 0

		return errors.Join(err, unbindErr)
	}

	c.writeReplicationPartners(ch, neighbors)

	return nil
}

func (c *Collector) writeReplicationPartners(ch chan<- prometheus.Metric, neighbors []ntdsapi.Neighbor) {
	type key struct{ namingContext, partner, site string }

	seen := make(map[key]struct{}, len(neighbors))

	for _, neighbor := range neighbors {
		partner, site := partnerFromDsaDN(neighbor.SourceDsaDN)

		// A domain controller that was demoted and promoted again under the same
		// name can show up twice, once as a deleted DSA. Fall back to the full
		// distinguished name to keep the series apart.
		k := key{neighbor.NamingContext, partner, site}
		if _, ok := seen[k]; ok {
			partner, site = neighbor.SourceDsaDN, ""
			k = key{neighbor.NamingContext, partner, site}

			if _, ok := seen[k]; ok {
				continue
			}
		}

		seen[k] = struct{}{}

		ch <- prometheus.MustNewConstMetric(
			c.replicationPartnerConsecutiveFailures,
			prometheus.GaugeValue,
			float64(neighbor.ConsecutiveSyncFailures),
			neighbor.NamingContext, partner, site,
		)

		ch <- prometheus.MustNewConstMetric(
			c.replicationPartnerLastResult,
			prometheus.GaugeValue,
			float64(neighbor.LastSyncResult),
			neighbor.NamingContext, partner, site,
		)

		if !neighbor.LastSyncSuccess.IsZero() {
			ch <- prometheus.MustNewConstMetric(
				c.replicationPartnerLastSuccess,
				prometheus.GaugeValue,
				float64(neighbor.LastSyncSuccess.UnixNano())/1e9,
				neighbor.NamingContext, partner, site,
			)
		}

		if !neighbor.LastSyncAttempt.IsZero() {
			ch <- prometheus.MustNewConstMetric(
				c.replicationPartnerLastAttempt,
				prometheus.GaugeValue,
				float64(neighbor.LastSyncAttempt.UnixNano())/1e9,
				neighbor.NamingContext, partner, site,
			)
		}
	}
}

// partnerFromDsaDN returns the server and site name of a source domain
// controller from the distinguished name of its NTDS Settings object,
// CN=NTDS Settings,CN=<server>,CN=Servers,CN=<site>,CN=Sites,CN=Configuration,...
// For any other shape it returns the distinguished name and an empty site.
func partnerFromDsaDN(dn string) (string, string) {
	rdns := splitDN(dn)
	if len(rdns) < 5 ||
		!strings.EqualFold(rdns[2], "CN=Servers") ||
		!strings.EqualFold(rdns[4], "CN=Sites") ||
		!hasCNPrefix(rdns[1]) || !hasCNPrefix(rdns[3]) {
		return dn, ""
	}

	return unescapeRDNValue(rdns[1][3:]), unescapeRDNValue(rdns[3][3:])
}

func hasCNPrefix(rdn string) bool {
	return len(rdn) > 3 && strings.EqualFold(rdn[:3], "CN=")
}

// splitDN splits a distinguished name at the commas that are not escaped with
// a backslash (RFC 4514). The parts keep their escapes.
func splitDN(dn string) []string {
	var (
		parts   []string
		start   int
		escaped bool
	)

	for i := range len(dn) {
		switch {
		case escaped:
			escaped = false
		case dn[i] == '\\':
			escaped = true
		case dn[i] == ',':
			parts = append(parts, strings.TrimSpace(dn[start:i]))
			start = i + 1
		}
	}

	return append(parts, strings.TrimSpace(dn[start:]))
}

// unescapeRDNValue removes the backslash before an escaped special character.
// Hex escapes such as \0A stay as they are.
func unescapeRDNValue(value string) string {
	if !strings.Contains(value, `\`) {
		return value
	}

	var b strings.Builder

	for i := 0; i < len(value); i++ {
		if value[i] == '\\' && i+1 < len(value) && strings.IndexByte(`,+"\<>;=# `, value[i+1]) >= 0 {
			i++
		}

		b.WriteByte(value[i])
	}

	return b.String()
}
