// Copyright 2024 The Prometheus Authors
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

package dhcp

import (
	"errors"
	"fmt"
	"log/slog"
	"slices"
	"strconv"
	"strings"

	"github.com/alecthomas/kingpin/v2"
	"github.com/prometheus-community/windows_exporter/internal/headers/dhcpsapi"
	"github.com/prometheus-community/windows_exporter/internal/mi"
	"github.com/prometheus-community/windows_exporter/internal/pdh"
	"github.com/prometheus-community/windows_exporter/internal/types"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	Name = "dhcp"

	subCollectorServerMetrics = "server_metrics"
	subCollectorScopeMetrics  = "scope_metrics"
)

type Config struct {
	CollectorsEnabled []string `yaml:"collectors_enabled"`
}

//nolint:gochecknoglobals
var ConfigDefaults = Config{
	CollectorsEnabled: []string{
		subCollectorServerMetrics,
		subCollectorScopeMetrics,
	},
}

// A Collector is a Prometheus Collector perflib DHCP metrics.
type Collector struct {
	config Config

	perfDataCollector *pdh.Collector
	perfDataObject    []perfDataCounterValues

	acksTotal                                        *prometheus.Desc
	activeQueueLength                                *prometheus.Desc
	conflictCheckQueueLength                         *prometheus.Desc
	declinesTotal                                    *prometheus.Desc
	deniedDueToMatch                                 *prometheus.Desc
	deniedDueToNonMatch                              *prometheus.Desc
	discoversTotal                                   *prometheus.Desc
	duplicatesDroppedTotal                           *prometheus.Desc
	failoverBndAckReceivedTotal                      *prometheus.Desc
	failoverBndAckSentTotal                          *prometheus.Desc
	failoverBndUpdDropped                            *prometheus.Desc
	failoverBndUpdPendingOutboundQueue               *prometheus.Desc
	failoverBndUpdReceivedTotal                      *prometheus.Desc
	failoverBndUpdSentTotal                          *prometheus.Desc
	failoverTransitionsCommunicationInterruptedState *prometheus.Desc
	failoverTransitionsPartnerDownState              *prometheus.Desc
	failoverTransitionsRecoverState                  *prometheus.Desc
	informsTotal                                     *prometheus.Desc
	nACKsTotal                                       *prometheus.Desc
	offerQueueLength                                 *prometheus.Desc
	offersTotal                                      *prometheus.Desc
	packetsExpiredTotal                              *prometheus.Desc
	packetsReceivedTotal                             *prometheus.Desc
	releasesTotal                                    *prometheus.Desc
	requestsTotal                                    *prometheus.Desc

	scopeInfo                               *prometheus.Desc
	scopeState                              *prometheus.Desc
	scopeAddressesFreeTotal                 *prometheus.Desc
	scopeAddressesFreeOnPartnerServerTotal  *prometheus.Desc
	scopeAddressesFreeOnThisServerTotal     *prometheus.Desc
	scopeAddressesInUseTotal                *prometheus.Desc
	scopeAddressesInUseOnPartnerServerTotal *prometheus.Desc
	scopeAddressesInUseOnThisServerTotal    *prometheus.Desc
	scopePendingOffersTotal                 *prometheus.Desc
	scopeReservedAddressTotal               *prometheus.Desc
}

func New(config *Config) *Collector {
	if config == nil {
		config = &ConfigDefaults
	}

	if config.CollectorsEnabled == nil {
		config.CollectorsEnabled = ConfigDefaults.CollectorsEnabled
	}

	c := &Collector{
		config: *config,
	}

	return c
}

func NewWithFlags(app *kingpin.Application) *Collector {
	c := &Collector{
		config: ConfigDefaults,
	}
	c.config.CollectorsEnabled = make([]string, 0)

	var collectorsEnabled string

	app.Flag(
		"collector.dhcp.enabled",
		"Comma-separated list of collectors to use. Defaults to all, if not specified.",
	).Default(strings.Join(ConfigDefaults.CollectorsEnabled, ",")).StringVar(&collectorsEnabled)

	app.Action(func(*kingpin.ParseContext) error {
		c.config.CollectorsEnabled = strings.Split(collectorsEnabled, ",")

		return nil
	})

	return c
}

func (c *Collector) GetName() string {
	return Name
}

func (c *Collector) Close() error {
	if slices.Contains(c.config.CollectorsEnabled, subCollectorServerMetrics) {
		c.perfDataCollector.Close()
	}

	return nil
}

func (c *Collector) Build(_ *slog.Logger, _ *mi.Session) error {
	var err error

	if slices.Contains(c.config.CollectorsEnabled, subCollectorScopeMetrics) {
		c.scopeInfo = prometheus.NewDesc(
			prometheus.BuildFQName(types.Namespace, Name, "scope_info"),
			"DHCP Scope information",
			[]string{"name", "superscope_name", "superscope_id", "scope"},
			nil,
		)

		c.scopeState = prometheus.NewDesc(
			prometheus.BuildFQName(types.Namespace, Name, "scope_state"),
			"DHCP Scope state",
			[]string{"scope", "state"},
			nil,
		)

		c.scopeAddressesFreeTotal = prometheus.NewDesc(
			prometheus.BuildFQName(types.Namespace, Name, "scope_addresses_free"),
			"DHCP Scope free addresses",
			[]string{"scope"},
			nil,
		)

		c.scopeAddressesFreeOnPartnerServerTotal = prometheus.NewDesc(
			prometheus.BuildFQName(types.Namespace, Name, "scope_addresses_free_on_partner_server"),
			"DHCP Scope free addresses on partner server",
			[]string{"scope"},
			nil,
		)

		c.scopeAddressesFreeOnThisServerTotal = prometheus.NewDesc(
			prometheus.BuildFQName(types.Namespace, Name, "scope_addresses_free_on_this_server"),
			"DHCP Scope free addresses on this server",
			[]string{"scope"},
			nil,
		)

		c.scopeAddressesInUseTotal = prometheus.NewDesc(
			prometheus.BuildFQName(types.Namespace, Name, "scope_addresses_in_use"),
			"DHCP Scope addresses in use",
			[]string{"scope"},
			nil,
		)

		c.scopeAddressesInUseOnPartnerServerTotal = prometheus.NewDesc(
			prometheus.BuildFQName(types.Namespace, Name, "scope_addresses_in_use_on_partner_server"),
			"DHCP Scope addresses in use on partner server",
			[]string{"scope"},
			nil,
		)

		c.scopeAddressesInUseOnThisServerTotal = prometheus.NewDesc(
			prometheus.BuildFQName(types.Namespace, Name, "scope_addresses_in_use_on_this_server"),
			"DHCP Scope addresses in use on this server",
			[]string{"scope"},
			nil,
		)

		c.scopePendingOffersTotal = prometheus.NewDesc(
			prometheus.BuildFQName(types.Namespace, Name, "scope_pending_offers"),
			"DHCP Scope pending offers",
			[]string{"scope"},
			nil,
		)

		c.scopeReservedAddressTotal = prometheus.NewDesc(
			prometheus.BuildFQName(types.Namespace, Name, "scope_reserved_address"),
			"DHCP Scope reserved addresses",
			[]string{"scope"},
			nil,
		)
	}

	if slices.Contains(c.config.CollectorsEnabled, subCollectorServerMetrics) {
		c.packetsReceivedTotal = prometheus.NewDesc(
			prometheus.BuildFQName(types.Namespace, Name, "packets_received_total"),
			"Total number of packets received by the DHCP server (PacketsReceivedTotal)",
			nil,
			nil,
		)
		c.duplicatesDroppedTotal = prometheus.NewDesc(
			prometheus.BuildFQName(types.Namespace, Name, "duplicates_dropped_total"),
			"Total number of duplicate packets received by the DHCP server (DuplicatesDroppedTotal)",
			nil,
			nil,
		)
		c.packetsExpiredTotal = prometheus.NewDesc(
			prometheus.BuildFQName(types.Namespace, Name, "packets_expired_total"),
			"Total number of packets expired in the DHCP server message queue (PacketsExpiredTotal)",
			nil,
			nil,
		)
		c.activeQueueLength = prometheus.NewDesc(
			prometheus.BuildFQName(types.Namespace, Name, "active_queue_length"),
			"Number of packets in the processing queue of the DHCP server (ActiveQueueLength)",
			nil,
			nil,
		)
		c.conflictCheckQueueLength = prometheus.NewDesc(
			prometheus.BuildFQName(types.Namespace, Name, "conflict_check_queue_length"),
			"Number of packets in the DHCP server queue waiting on conflict detection (ping). (ConflictCheckQueueLength)",
			nil,
			nil,
		)
		c.discoversTotal = prometheus.NewDesc(
			prometheus.BuildFQName(types.Namespace, Name, "discovers_total"),
			"Total DHCP Discovers received by the DHCP server (DiscoversTotal)",
			nil,
			nil,
		)
		c.offersTotal = prometheus.NewDesc(
			prometheus.BuildFQName(types.Namespace, Name, "offers_total"),
			"Total DHCP Offers sent by the DHCP server (OffersTotal)",
			nil,
			nil,
		)
		c.requestsTotal = prometheus.NewDesc(
			prometheus.BuildFQName(types.Namespace, Name, "requests_total"),
			"Total DHCP Requests received by the DHCP server (RequestsTotal)",
			nil,
			nil,
		)
		c.informsTotal = prometheus.NewDesc(
			prometheus.BuildFQName(types.Namespace, Name, "informs_total"),
			"Total DHCP Informs received by the DHCP server (InformsTotal)",
			nil,
			nil,
		)
		c.acksTotal = prometheus.NewDesc(
			prometheus.BuildFQName(types.Namespace, Name, "acks_total"),
			"Total DHCP Acks sent by the DHCP server (AcksTotal)",
			nil,
			nil,
		)
		c.nACKsTotal = prometheus.NewDesc(
			prometheus.BuildFQName(types.Namespace, Name, "nacks_total"),
			"Total DHCP Nacks sent by the DHCP server (NacksTotal)",
			nil,
			nil,
		)
		c.declinesTotal = prometheus.NewDesc(
			prometheus.BuildFQName(types.Namespace, Name, "declines_total"),
			"Total DHCP Declines received by the DHCP server (DeclinesTotal)",
			nil,
			nil,
		)
		c.releasesTotal = prometheus.NewDesc(
			prometheus.BuildFQName(types.Namespace, Name, "releases_total"),
			"Total DHCP Releases received by the DHCP server (ReleasesTotal)",
			nil,
			nil,
		)
		c.offerQueueLength = prometheus.NewDesc(
			prometheus.BuildFQName(types.Namespace, Name, "offer_queue_length"),
			"Number of packets in the offer queue of the DHCP server (OfferQueueLength)",
			nil,
			nil,
		)
		c.deniedDueToMatch = prometheus.NewDesc(
			prometheus.BuildFQName(types.Namespace, Name, "denied_due_to_match_total"),
			"Total number of DHCP requests denied, based on matches from the Deny list (DeniedDueToMatch)",
			nil,
			nil,
		)
		c.deniedDueToNonMatch = prometheus.NewDesc(
			prometheus.BuildFQName(types.Namespace, Name, "denied_due_to_nonmatch_total"),
			"Total number of DHCP requests denied, based on non-matches from the Allow list (DeniedDueToNonMatch)",
			nil,
			nil,
		)
		c.failoverBndUpdSentTotal = prometheus.NewDesc(
			prometheus.BuildFQName(types.Namespace, Name, "failover_bndupd_sent_total"),
			"Number of DHCP fail over Binding Update messages sent (FailoverBndupdSentTotal)",
			nil,
			nil,
		)
		c.failoverBndUpdReceivedTotal = prometheus.NewDesc(
			prometheus.BuildFQName(types.Namespace, Name, "failover_bndupd_received_total"),
			"Number of DHCP fail over Binding Update messages received (FailoverBndupdReceivedTotal)",
			nil,
			nil,
		)
		c.failoverBndAckSentTotal = prometheus.NewDesc(
			prometheus.BuildFQName(types.Namespace, Name, "failover_bndack_sent_total"),
			"Number of DHCP fail over Binding Ack messages sent (FailoverBndackSentTotal)",
			nil,
			nil,
		)
		c.failoverBndAckReceivedTotal = prometheus.NewDesc(
			prometheus.BuildFQName(types.Namespace, Name, "failover_bndack_received_total"),
			"Number of DHCP fail over Binding Ack messages received (FailoverBndackReceivedTotal)",
			nil,
			nil,
		)
		c.failoverBndUpdPendingOutboundQueue = prometheus.NewDesc(
			prometheus.BuildFQName(types.Namespace, Name, "failover_bndupd_pending_in_outbound_queue"),
			"Number of pending outbound DHCP fail over Binding Update messages (FailoverBndupdPendingOutboundQueue)",
			nil,
			nil,
		)
		c.failoverTransitionsCommunicationInterruptedState = prometheus.NewDesc(
			prometheus.BuildFQName(types.Namespace, Name, "failover_transitions_communicationinterrupted_state_total"),
			"Total number of transitions into COMMUNICATION INTERRUPTED state (FailoverTransitionsCommunicationinterruptedState)",
			nil,
			nil,
		)
		c.failoverTransitionsPartnerDownState = prometheus.NewDesc(
			prometheus.BuildFQName(types.Namespace, Name, "failover_transitions_partnerdown_state_total"),
			"Total number of transitions into PARTNER DOWN state (FailoverTransitionsPartnerdownState)",
			nil,
			nil,
		)
		c.failoverTransitionsRecoverState = prometheus.NewDesc(
			prometheus.BuildFQName(types.Namespace, Name, "failover_transitions_recover_total"),
			"Total number of transitions into RECOVER state (FailoverTransitionsRecoverState)",
			nil,
			nil,
		)
		c.failoverBndUpdDropped = prometheus.NewDesc(
			prometheus.BuildFQName(types.Namespace, Name, "failover_bndupd_dropped_total"),
			"Total number of DHCP fail over Binding Updates dropped (FailoverBndupdDropped)",
			nil,
			nil,
		)

		c.perfDataCollector, err = pdh.NewCollector[perfDataCounterValues](pdh.CounterTypeRaw, "DHCP Server", nil)
		if err != nil {
			return fmt.Errorf("failed to create DHCP Server collector: %w", err)
		}
	}

	return nil
}

func (c *Collector) Collect(ch chan<- prometheus.Metric) error {
	var errs []error

	if slices.Contains(c.config.CollectorsEnabled, subCollectorServerMetrics) {
		if err := c.collectServerMetrics(ch); err != nil {
			errs = append(errs, err)
		}
	}

	if slices.Contains(c.config.CollectorsEnabled, subCollectorScopeMetrics) {
		if err := c.collectScopeMetrics(ch); err != nil {
			errs = append(errs, err)
		}
	}

	return errors.Join(errs...)
}

func (c *Collector) collectServerMetrics(ch chan<- prometheus.Metric) error {
	err := c.perfDataCollector.Collect(&c.perfDataObject)
	if err != nil {
		return fmt.Errorf("failed to collect DHCP Server metrics: %w", err)
	}

	ch <- prometheus.MustNewConstMetric(
		c.packetsReceivedTotal,
		prometheus.CounterValue,
		c.perfDataObject[0].PacketsReceivedTotal,
	)

	ch <- prometheus.MustNewConstMetric(
		c.duplicatesDroppedTotal,
		prometheus.CounterValue,
		c.perfDataObject[0].DuplicatesDroppedTotal,
	)

	ch <- prometheus.MustNewConstMetric(
		c.packetsExpiredTotal,
		prometheus.CounterValue,
		c.perfDataObject[0].PacketsExpiredTotal,
	)

	ch <- prometheus.MustNewConstMetric(
		c.activeQueueLength,
		prometheus.GaugeValue,
		c.perfDataObject[0].ActiveQueueLength,
	)

	ch <- prometheus.MustNewConstMetric(
		c.conflictCheckQueueLength,
		prometheus.GaugeValue,
		c.perfDataObject[0].ConflictCheckQueueLength,
	)

	ch <- prometheus.MustNewConstMetric(
		c.discoversTotal,
		prometheus.CounterValue,
		c.perfDataObject[0].DiscoversTotal,
	)

	ch <- prometheus.MustNewConstMetric(
		c.offersTotal,
		prometheus.CounterValue,
		c.perfDataObject[0].OffersTotal,
	)

	ch <- prometheus.MustNewConstMetric(
		c.requestsTotal,
		prometheus.CounterValue,
		c.perfDataObject[0].RequestsTotal,
	)

	ch <- prometheus.MustNewConstMetric(
		c.informsTotal,
		prometheus.CounterValue,
		c.perfDataObject[0].InformsTotal,
	)

	ch <- prometheus.MustNewConstMetric(
		c.acksTotal,
		prometheus.CounterValue,
		c.perfDataObject[0].AcksTotal,
	)

	ch <- prometheus.MustNewConstMetric(
		c.nACKsTotal,
		prometheus.CounterValue,
		c.perfDataObject[0].NacksTotal,
	)

	ch <- prometheus.MustNewConstMetric(
		c.declinesTotal,
		prometheus.CounterValue,
		c.perfDataObject[0].DeclinesTotal,
	)

	ch <- prometheus.MustNewConstMetric(
		c.releasesTotal,
		prometheus.CounterValue,
		c.perfDataObject[0].ReleasesTotal,
	)

	ch <- prometheus.MustNewConstMetric(
		c.offerQueueLength,
		prometheus.GaugeValue,
		c.perfDataObject[0].OfferQueueLength,
	)

	ch <- prometheus.MustNewConstMetric(
		c.deniedDueToMatch,
		prometheus.CounterValue,
		c.perfDataObject[0].DeniedDueToMatch,
	)

	ch <- prometheus.MustNewConstMetric(
		c.deniedDueToNonMatch,
		prometheus.CounterValue,
		c.perfDataObject[0].DeniedDueToNonMatch,
	)

	ch <- prometheus.MustNewConstMetric(
		c.failoverBndUpdSentTotal,
		prometheus.CounterValue,
		c.perfDataObject[0].FailoverBndUpdSentTotal,
	)

	ch <- prometheus.MustNewConstMetric(
		c.failoverBndUpdReceivedTotal,
		prometheus.CounterValue,
		c.perfDataObject[0].FailoverBndUpdReceivedTotal,
	)

	ch <- prometheus.MustNewConstMetric(
		c.failoverBndAckSentTotal,
		prometheus.CounterValue,
		c.perfDataObject[0].FailoverBndAckSentTotal,
	)

	ch <- prometheus.MustNewConstMetric(
		c.failoverBndAckReceivedTotal,
		prometheus.CounterValue,
		c.perfDataObject[0].FailoverBndAckReceivedTotal,
	)

	ch <- prometheus.MustNewConstMetric(
		c.failoverBndUpdPendingOutboundQueue,
		prometheus.GaugeValue,
		c.perfDataObject[0].FailoverBndUpdPendingOutboundQueue,
	)

	ch <- prometheus.MustNewConstMetric(
		c.failoverTransitionsCommunicationInterruptedState,
		prometheus.CounterValue,
		c.perfDataObject[0].FailoverTransitionsCommunicationInterruptedState,
	)

	ch <- prometheus.MustNewConstMetric(
		c.failoverTransitionsPartnerDownState,
		prometheus.CounterValue,
		c.perfDataObject[0].FailoverTransitionsPartnerDownState,
	)

	ch <- prometheus.MustNewConstMetric(
		c.failoverTransitionsRecoverState,
		prometheus.CounterValue,
		c.perfDataObject[0].FailoverTransitionsRecoverState,
	)

	ch <- prometheus.MustNewConstMetric(
		c.failoverBndUpdDropped,
		prometheus.CounterValue,
		c.perfDataObject[0].FailoverBndUpdDropped,
	)

	return nil
}

func (c *Collector) collectScopeMetrics(ch chan<- prometheus.Metric) error {
	dhcpScopes, err := dhcpsapi.GetDHCPV4ScopeStatistics()
	if err != nil {
		return fmt.Errorf("failed to get DHCP scopes: %w", err)
	}

	for _, scope := range dhcpScopes {
		scopeID := scope.ScopeIPAddress.String()

		ch <- prometheus.MustNewConstMetric(
			c.scopeInfo,
			prometheus.GaugeValue,
			1,
			scope.Name,
			scope.SuperScopeName,
			strconv.Itoa(int(scope.SuperScopeNumber)),
			scopeID,
		)

		for state, name := range dhcpsapi.DHCP_SUBNET_STATE_NAMES {
			metric := 0.0
			if state == scope.State {
				metric = 1.0
			}

			ch <- prometheus.MustNewConstMetric(
				c.scopeState,
				prometheus.GaugeValue,
				metric,
				scopeID,
				name,
			)
		}

		if scope.AddressesFree != -1 {
			ch <- prometheus.MustNewConstMetric(
				c.scopeAddressesFreeTotal,
				prometheus.GaugeValue,
				scope.AddressesFree,
				scopeID,
			)
		}

		if scope.AddressesFreeOnPartnerServer != -1 {
			ch <- prometheus.MustNewConstMetric(
				c.scopeAddressesFreeOnPartnerServerTotal,
				prometheus.GaugeValue,
				scope.AddressesFreeOnPartnerServer,
				scopeID,
			)
		}

		if scope.AddressesFreeOnThisServer != -1 {
			ch <- prometheus.MustNewConstMetric(
				c.scopeAddressesFreeOnThisServerTotal,
				prometheus.GaugeValue,
				scope.AddressesFreeOnThisServer,
				scopeID,
			)
		}

		if scope.AddressesInUse != -1 {
			ch <- prometheus.MustNewConstMetric(
				c.scopeAddressesInUseTotal,
				prometheus.GaugeValue,
				scope.AddressesInUse,
				scopeID,
			)
		}

		if scope.AddressesInUseOnPartnerServer != -1 {
			ch <- prometheus.MustNewConstMetric(
				c.scopeAddressesInUseOnPartnerServerTotal,
				prometheus.GaugeValue,
				scope.AddressesInUseOnPartnerServer,
				scopeID,
			)
		}

		if scope.AddressesInUseOnThisServer != -1 {
			ch <- prometheus.MustNewConstMetric(
				c.scopeAddressesInUseOnThisServerTotal,
				prometheus.GaugeValue,
				scope.AddressesInUseOnThisServer,
				scopeID,
			)
		}

		if scope.PendingOffers != -1 {
			ch <- prometheus.MustNewConstMetric(
				c.scopePendingOffersTotal,
				prometheus.GaugeValue,
				scope.PendingOffers,
				scopeID,
			)
		}

		if scope.ReservedAddress != -1 {
			ch <- prometheus.MustNewConstMetric(
				c.scopeReservedAddressTotal,
				prometheus.GaugeValue,
				scope.ReservedAddress,
				scopeID,
			)
		}
	}

	return nil
}
