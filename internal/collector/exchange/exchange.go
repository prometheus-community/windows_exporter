//go:build windows

package exchange

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/alecthomas/kingpin/v2"
	"github.com/prometheus-community/windows_exporter/internal/mi"
	"github.com/prometheus-community/windows_exporter/internal/perfdata"
	"github.com/prometheus/client_golang/prometheus"
)

const Name = "exchange"

const (
	adAccessProcesses   = "ADAccessProcesses"
	transportQueues     = "TransportQueues"
	httpProxy           = "HttpProxy"
	activeSync          = "ActiveSync"
	availabilityService = "AvailabilityService"
	outlookWebAccess    = "OutlookWebAccess"
	autoDiscover        = "Autodiscover"
	workloadManagement  = "WorkloadManagement"
	rpcClientAccess     = "RpcClientAccess"
	mapiHttpEmsmdb      = "MapiHttpEmsmdb"
)

type Config struct {
	CollectorsEnabled []string `yaml:"collectors_enabled"`
}

var ConfigDefaults = Config{
	CollectorsEnabled: []string{
		adAccessProcesses,
		transportQueues,
		httpProxy,
		activeSync,
		availabilityService,
		outlookWebAccess,
		autoDiscover,
		workloadManagement,
		rpcClientAccess,
		mapiHttpEmsmdb,
	},
}

type Collector struct {
	config Config

	perfDataCollectorADAccessProcesses           *perfdata.Collector
	perfDataCollectorTransportQueues             *perfdata.Collector
	perfDataCollectorHttpProxy                   *perfdata.Collector
	perfDataCollectorActiveSync                  *perfdata.Collector
	perfDataCollectorAvailabilityService         *perfdata.Collector
	perfDataCollectorOWA                         *perfdata.Collector
	perfDataCollectorAutoDiscover                *perfdata.Collector
	perfDataCollectorWorkloadManagementWorkloads *perfdata.Collector
	perfDataCollectorRpcClientAccess             *perfdata.Collector
	perfDataCollectorMapiHttpEmsmdb              *perfdata.Collector

	activeMailboxDeliveryQueueLength        *prometheus.Desc
	activeSyncRequestsPerSec                *prometheus.Desc
	activeTasks                             *prometheus.Desc
	activeUserCount                         *prometheus.Desc
	activeUserCountMapiHttpEmsMDB           *prometheus.Desc
	autoDiscoverRequestsPerSec              *prometheus.Desc
	availabilityRequestsSec                 *prometheus.Desc
	averageAuthenticationLatency            *prometheus.Desc
	averageCASProcessingLatency             *prometheus.Desc
	completedTasks                          *prometheus.Desc
	connectionCount                         *prometheus.Desc
	currentUniqueUsers                      *prometheus.Desc
	externalActiveRemoteDeliveryQueueLength *prometheus.Desc
	externalLargestDeliveryQueueLength      *prometheus.Desc
	internalActiveRemoteDeliveryQueueLength *prometheus.Desc
	internalLargestDeliveryQueueLength      *prometheus.Desc
	isActive                                *prometheus.Desc
	ldapReadTime                            *prometheus.Desc
	ldapSearchTime                          *prometheus.Desc
	ldapTimeoutErrorsPerSec                 *prometheus.Desc
	ldapWriteTime                           *prometheus.Desc
	longRunningLDAPOperationsPerMin         *prometheus.Desc
	mailboxServerLocatorAverageLatency      *prometheus.Desc
	mailboxServerProxyFailureRate           *prometheus.Desc
	outstandingProxyRequests                *prometheus.Desc
	owaRequestsPerSec                       *prometheus.Desc
	pingCommandsPending                     *prometheus.Desc
	poisonQueueLength                       *prometheus.Desc
	proxyRequestsPerSec                     *prometheus.Desc
	queuedTasks                             *prometheus.Desc
	retryMailboxDeliveryQueueLength         *prometheus.Desc
	rpcAveragedLatency                      *prometheus.Desc
	rpcOperationsPerSec                     *prometheus.Desc
	rpcRequests                             *prometheus.Desc
	syncCommandsPerSec                      *prometheus.Desc
	unreachableQueueLength                  *prometheus.Desc
	userCount                               *prometheus.Desc
	yieldedTasks                            *prometheus.Desc
	messagesQueuedForDeliveryTotal          *prometheus.Desc
	messagesSubmittedTotal                  *prometheus.Desc
	messagesDelayedTotal                    *prometheus.Desc
	messagesCompletedDeliveryTotal          *prometheus.Desc
	shadowQueueLength                       *prometheus.Desc
	submissionQueueLength                   *prometheus.Desc
	delayQueueLength                        *prometheus.Desc
	itemsCompletedDeliveryTotal             *prometheus.Desc
	itemsQueuedForDeliveryExpiredTotal      *prometheus.Desc
	itemsQueuedForDeliveryTotal             *prometheus.Desc
	itemsResubmittedTotal                   *prometheus.Desc
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

	var listAllCollectors bool

	var collectorsEnabled string

	app.Flag(
		"collector.exchange.list",
		"List the collectors along with their perflib object name/ids",
	).BoolVar(&listAllCollectors)

	app.Flag(
		"collector.exchange.enabled",
		"Comma-separated list of collectors to use. Defaults to all, if not specified.",
	).Default(strings.Join(ConfigDefaults.CollectorsEnabled, ",")).StringVar(&collectorsEnabled)

	app.PreAction(func(*kingpin.ParseContext) error {
		if listAllCollectors {
			collectorDesc := map[string]string{
				adAccessProcesses:   "[19108] MSExchange ADAccess Processes",
				transportQueues:     "[20524] MSExchangeTransport Queues",
				httpProxy:           "[36934] MSExchange HttpProxy",
				activeSync:          "[25138] MSExchange ActiveSync",
				availabilityService: "[24914] MSExchange Availability Service",
				outlookWebAccess:    "[24618] MSExchange OWA",
				autoDiscover:        "[29240] MSExchange Autodiscover",
				workloadManagement:  "[19430] MSExchange WorkloadManagement Workloads",
				rpcClientAccess:     "[29336] MSExchange RpcClientAccess",
				mapiHttpEmsmdb:      "[26463] MSExchange MapiHttp Emsmdb",
			}

			sb := strings.Builder{}
			sb.WriteString(fmt.Sprintf("%-32s %-32s\n", "Collector Name", "[PerfID] Perflib Object"))

			for _, cname := range ConfigDefaults.CollectorsEnabled {
				sb.WriteString(fmt.Sprintf("%-32s %-32s\n", cname, collectorDesc[cname]))
			}

			app.UsageTemplate(sb.String()).Usage(nil)

			os.Exit(0)
		}

		return nil
	})

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
	return nil
}

func (c *Collector) Build(_ *slog.Logger, _ *mi.Session) error {
	collectorFuncs := map[string]func() error{
		adAccessProcesses:   c.buildADAccessProcesses,
		transportQueues:     c.buildTransportQueues,
		httpProxy:           c.buildHTTPProxy,
		activeSync:          c.buildActiveSync,
		availabilityService: c.buildAvailabilityService,
		outlookWebAccess:    c.buildOWA,
		autoDiscover:        c.buildAutoDiscover,
		workloadManagement:  c.buildWorkloadManagementWorkloads,
		rpcClientAccess:     c.buildRPC,
		mapiHttpEmsmdb:      c.buildMapiHttpEmsmdb,
	}

	for _, collectorName := range c.config.CollectorsEnabled {
		if err := collectorFuncs[collectorName](); err != nil {
			return err
		}
	}

	return nil
}

// Collect collects exchange metrics and sends them to prometheus.
func (c *Collector) Collect(ch chan<- prometheus.Metric) error {
	collectorFuncs := map[string]func(ch chan<- prometheus.Metric) error{
		adAccessProcesses:   c.collectADAccessProcesses,
		transportQueues:     c.collectTransportQueues,
		httpProxy:           c.collectHTTPProxy,
		activeSync:          c.collectActiveSync,
		availabilityService: c.collectAvailabilityService,
		outlookWebAccess:    c.collectOWA,
		autoDiscover:        c.collectAutoDiscover,
		workloadManagement:  c.collectWorkloadManagementWorkloads,
		rpcClientAccess:     c.collectRPC,
		mapiHttpEmsmdb:      c.collectMapiHttpEmsmdb,
	}

	errs := make([]error, len(c.config.CollectorsEnabled))

	for i, collectorName := range c.config.CollectorsEnabled {
		errs[i] = collectorFuncs[collectorName](ch)
	}

	return errors.Join(errs...)
}

// toLabelName converts strings to lowercase and replaces all whitespaces and dots with underscores.
func (c *Collector) toLabelName(name string) string {
	s := strings.ReplaceAll(strings.Join(strings.Fields(strings.ToLower(name)), "_"), ".", "_")
	s = strings.ReplaceAll(s, "__", "_")

	return s
}

// msToSec converts from ms to seconds.
func (c *Collector) msToSec(t float64) float64 {
	return t / 1000
}
