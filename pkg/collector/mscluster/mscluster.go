package mscluster

import (
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/alecthomas/kingpin/v2"
	"github.com/go-kit/log"
	"github.com/prometheus-community/windows_exporter/pkg/types"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/yusufpapurcu/wmi"
)

const Name = "mscluster"

type Config struct {
	CollectorsEnabled []string `yaml:"collectors_enabled"`
}

var ConfigDefaults = Config{
	CollectorsEnabled: []string{
		"cluster",
		"network",
		"node",
		"resource",
		"resourcegroup",
	},
}

// A Collector is a Prometheus Collector for WMI MSCluster_Cluster metrics.
type Collector struct {
	config    Config
	wmiClient *wmi.Client

	// cluster
	clusterAddEvictDelay                           *prometheus.Desc
	clusterAdminAccessPoint                        *prometheus.Desc
	clusterAutoAssignNodeSite                      *prometheus.Desc
	clusterAutoBalancerLevel                       *prometheus.Desc
	clusterAutoBalancerMode                        *prometheus.Desc
	clusterBackupInProgress                        *prometheus.Desc
	clusterBlockCacheSize                          *prometheus.Desc
	clusterClusSvcHangTimeout                      *prometheus.Desc
	clusterClusSvcRegroupOpeningTimeout            *prometheus.Desc
	clusterClusSvcRegroupPruningTimeout            *prometheus.Desc
	clusterClusSvcRegroupStageTimeout              *prometheus.Desc
	clusterClusSvcRegroupTickInMilliseconds        *prometheus.Desc
	clusterClusterEnforcedAntiAffinity             *prometheus.Desc
	clusterClusterFunctionalLevel                  *prometheus.Desc
	clusterClusterGroupWaitDelay                   *prometheus.Desc
	clusterClusterLogLevel                         *prometheus.Desc
	clusterClusterLogSize                          *prometheus.Desc
	clusterClusterUpgradeVersion                   *prometheus.Desc
	clusterCrossSiteDelay                          *prometheus.Desc
	clusterCrossSiteThreshold                      *prometheus.Desc
	clusterCrossSubnetDelay                        *prometheus.Desc
	clusterCrossSubnetThreshold                    *prometheus.Desc
	clusterCsvBalancer                             *prometheus.Desc
	clusterDatabaseReadWriteMode                   *prometheus.Desc
	clusterDefaultNetworkRole                      *prometheus.Desc
	clusterDetectedCloudPlatform                   *prometheus.Desc
	clusterDetectManagedEvents                     *prometheus.Desc
	clusterDetectManagedEventsThreshold            *prometheus.Desc
	clusterDisableGroupPreferredOwnerRandomization *prometheus.Desc
	clusterDrainOnShutdown                         *prometheus.Desc
	clusterDynamicQuorumEnabled                    *prometheus.Desc
	clusterEnableSharedVolumes                     *prometheus.Desc
	clusterFixQuorum                               *prometheus.Desc
	clusterGracePeriodEnabled                      *prometheus.Desc
	clusterGracePeriodTimeout                      *prometheus.Desc
	clusterGroupDependencyTimeout                  *prometheus.Desc
	clusterHangRecoveryAction                      *prometheus.Desc
	clusterIgnorePersistentStateOnStartup          *prometheus.Desc
	clusterLogResourceControls                     *prometheus.Desc
	clusterLowerQuorumPriorityNodeId               *prometheus.Desc
	clusterMaxNumberOfNodes                        *prometheus.Desc
	clusterMessageBufferLength                     *prometheus.Desc
	clusterMinimumNeverPreemptPriority             *prometheus.Desc
	clusterMinimumPreemptorPriority                *prometheus.Desc
	clusterNetftIPSecEnabled                       *prometheus.Desc
	clusterPlacementOptions                        *prometheus.Desc
	clusterPlumbAllCrossSubnetRoutes               *prometheus.Desc
	clusterPreventQuorum                           *prometheus.Desc
	clusterQuarantineDuration                      *prometheus.Desc
	clusterQuarantineThreshold                     *prometheus.Desc
	clusterQuorumArbitrationTimeMax                *prometheus.Desc
	clusterQuorumArbitrationTimeMin                *prometheus.Desc
	clusterQuorumLogFileSize                       *prometheus.Desc
	clusterQuorumTypeValue                         *prometheus.Desc
	clusterRequestReplyTimeout                     *prometheus.Desc
	clusterResiliencyDefaultPeriod                 *prometheus.Desc
	clusterResiliencyLevel                         *prometheus.Desc
	clusterResourceDllDeadlockPeriod               *prometheus.Desc
	clusterRootMemoryReserved                      *prometheus.Desc
	clusterRouteHistoryLength                      *prometheus.Desc
	clusterS2DBusTypes                             *prometheus.Desc
	clusterS2DCacheDesiredState                    *prometheus.Desc
	clusterS2DCacheFlashReservePercent             *prometheus.Desc
	clusterS2DCachePageSizeKBytes                  *prometheus.Desc
	clusterS2DEnabled                              *prometheus.Desc
	clusterS2DIOLatencyThreshold                   *prometheus.Desc
	clusterS2DOptimizations                        *prometheus.Desc
	clusterSameSubnetDelay                         *prometheus.Desc
	clusterSameSubnetThreshold                     *prometheus.Desc
	clusterSecurityLevel                           *prometheus.Desc
	clusterSecurityLevelForStorage                 *prometheus.Desc
	clusterSharedVolumeVssWriterOperationTimeout   *prometheus.Desc
	clusterShutdownTimeoutInMinutes                *prometheus.Desc
	clusterUseClientAccessNetworksForSharedVolumes *prometheus.Desc
	clusterWitnessDatabaseWriteTimeout             *prometheus.Desc
	clusterWitnessDynamicWeight                    *prometheus.Desc
	clusterWitnessRestartInterval                  *prometheus.Desc

	// network
	networkCharacteristics *prometheus.Desc
	networkFlags           *prometheus.Desc
	networkMetric          *prometheus.Desc
	networkRole            *prometheus.Desc
	networkState           *prometheus.Desc

	// node
	nodeBuildNumber           *prometheus.Desc
	nodeCharacteristics       *prometheus.Desc
	nodeDetectedCloudPlatform *prometheus.Desc
	nodeDynamicWeight         *prometheus.Desc
	nodeFlags                 *prometheus.Desc
	nodeMajorVersion          *prometheus.Desc
	nodeMinorVersion          *prometheus.Desc
	nodeNeedsPreventQuorum    *prometheus.Desc
	nodeNodeDrainStatus       *prometheus.Desc
	nodeNodeHighestVersion    *prometheus.Desc
	nodeNodeLowestVersion     *prometheus.Desc
	nodeNodeWeight            *prometheus.Desc
	nodeState                 *prometheus.Desc
	nodeStatusInformation     *prometheus.Desc

	resourceCharacteristics        *prometheus.Desc
	resourceDeadlockTimeout        *prometheus.Desc
	resourceEmbeddedFailureAction  *prometheus.Desc
	resourceFlags                  *prometheus.Desc
	resourceIsAlivePollInterval    *prometheus.Desc
	resourceLooksAlivePollInterval *prometheus.Desc
	resourceMonitorProcessId       *prometheus.Desc
	resourceOwnerNode              *prometheus.Desc
	resourcePendingTimeout         *prometheus.Desc
	resourceResourceClass          *prometheus.Desc
	resourceRestartAction          *prometheus.Desc
	resourceRestartDelay           *prometheus.Desc
	resourceRestartPeriod          *prometheus.Desc
	resourceRestartThreshold       *prometheus.Desc
	resourceRetryPeriodOnFailure   *prometheus.Desc
	resourceState                  *prometheus.Desc
	resourceSubClass               *prometheus.Desc

	// ResourceGroup
	resourceGroupAutoFailbackType    *prometheus.Desc
	resourceGroupCharacteristics     *prometheus.Desc
	resourceGroupColdStartSetting    *prometheus.Desc
	resourceGroupDefaultOwner        *prometheus.Desc
	resourceGroupFailbackWindowEnd   *prometheus.Desc
	resourceGroupFailbackWindowStart *prometheus.Desc
	resourceGroupFailOverPeriod      *prometheus.Desc
	resourceGroupFailOverThreshold   *prometheus.Desc
	resourceGroupFlags               *prometheus.Desc
	resourceGroupGroupType           *prometheus.Desc
	resourceGroupOwnerNode           *prometheus.Desc
	resourceGroupPriority            *prometheus.Desc
	resourceGroupResiliencyPeriod    *prometheus.Desc
	resourceGroupState               *prometheus.Desc
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
		"collectors.mscluster.enabled",
		"Comma-separated list of collectors to use.",
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

func (c *Collector) GetPerfCounter(_ log.Logger) ([]string, error) {
	return []string{"Memory"}, nil
}

func (c *Collector) Close() error {
	return nil
}

func (c *Collector) Build(_ log.Logger, wmiClient *wmi.Client) error {
	if len(c.config.CollectorsEnabled) == 0 {
		return nil
	}

	if wmiClient == nil || wmiClient.SWbemServicesClient == nil {
		return errors.New("wmiClient or SWbemServicesClient is nil")
	}

	c.wmiClient = wmiClient

	if slices.Contains(c.config.CollectorsEnabled, "cluster") {
		c.buildCluster()
	}

	if slices.Contains(c.config.CollectorsEnabled, "network") {
		c.buildNetwork()
	}

	if slices.Contains(c.config.CollectorsEnabled, "node") {
		c.buildNode()
	}

	if slices.Contains(c.config.CollectorsEnabled, "resource") {
		c.buildResource()
	}

	if slices.Contains(c.config.CollectorsEnabled, "resourcegroup") {
		c.buildResourceGroup()
	}

	return nil
}

// Collect sends the metric values for each metric
// to the provided prometheus Metric channel.
func (c *Collector) Collect(_ *types.ScrapeContext, _ log.Logger, ch chan<- prometheus.Metric) error {
	if len(c.config.CollectorsEnabled) == 0 {
		return nil
	}

	var (
		err       error
		errs      []error
		nodeNames []string
	)

	if slices.Contains(c.config.CollectorsEnabled, "cluster") {
		if err = c.collectCluster(ch); err != nil {
			errs = append(errs, fmt.Errorf("failed to collect cluster metrics: %w", err))
		}
	}

	if slices.Contains(c.config.CollectorsEnabled, "network") {
		if err = c.collectNetwork(ch); err != nil {
			errs = append(errs, fmt.Errorf("failed to collect network metrics: %w", err))
		}
	}

	if slices.Contains(c.config.CollectorsEnabled, "node") {
		if nodeNames, err = c.collectNode(ch); err != nil {
			errs = append(errs, fmt.Errorf("failed to collect node metrics: %w", err))
		}
	}

	if slices.Contains(c.config.CollectorsEnabled, "resource") {
		if err = c.collectResource(ch, nodeNames); err != nil {
			errs = append(errs, fmt.Errorf("failed to collect resource metrics: %w", err))
		}
	}

	if slices.Contains(c.config.CollectorsEnabled, "resourcegroup") {
		if err = c.collectResourceGroup(ch, nodeNames); err != nil {
			errs = append(errs, fmt.Errorf("failed to collect resource group metrics: %w", err))
		}
	}

	if len(errs) > 0 {
		return errors.Join(errs...)
	}

	return nil
}
