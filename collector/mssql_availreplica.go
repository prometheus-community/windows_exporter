// returns data points from the following classes:
// - Win32_PerfRawData_MSSQLSERVER_SQLServerAvailabilityReplica
//   https://docs.microsoft.com/en-us/sql/relational-databases/performance-monitor/sql-server-availability-replica

package collector

import (
	"fmt"

	"github.com/StackExchange/wmi"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
)

func init() {
	Factories["mssql_availreplica"] = NewMSSQLAvailReplicaCollector
}

// MSSQLAvailReplicaCollector is a Prometheus collector for Win32_PerfRawData_{instance}_SQLServerAvailabilityReplica metrics
type MSSQLAvailReplicaCollector struct {
	BytesReceivedfromReplicaPersec *prometheus.Desc
	BytesSenttoReplicaPersec       *prometheus.Desc
	BytesSenttoTransportPersec     *prometheus.Desc
	FlowControlPersec              *prometheus.Desc
	FlowControlTimemsPersec        *prometheus.Desc
	ReceivesfromReplicaPersec      *prometheus.Desc
	ResentMessagesPersec           *prometheus.Desc
	SendstoReplicaPersec           *prometheus.Desc
	SendstoTransportPersec         *prometheus.Desc

	sqlInstances sqlInstancesType
}

// NewMSSQLAvailReplicaCollector ...
func NewMSSQLAvailReplicaCollector() (Collector, error) {

	const subsystem = "mssql_availreplica"
	return &MSSQLAvailReplicaCollector{

		// Win32_PerfRawData_{instance}_SQLServerAvailabilityReplica
		BytesReceivedfromReplicaPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "received_from_replica_bytes"),
			"(AvailabilityReplica.BytesReceivedfromReplica)",
			[]string{"instance", "replica"},
			nil,
		),
		BytesSenttoReplicaPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "sent_to_replica_bytes"),
			"(AvailabilityReplica.BytesSenttoReplica)",
			[]string{"instance", "replica"},
			nil,
		),
		BytesSenttoTransportPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "sent_to_transport_bytes"),
			"(AvailabilityReplica.BytesSenttoTransport)",
			[]string{"instance", "replica"},
			nil,
		),
		FlowControlPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "initiated_flow_controls"),
			"(AvailabilityReplica.FlowControl)",
			[]string{"instance", "replica"},
			nil,
		),
		FlowControlTimemsPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "flow_control_wait_seconds"),
			"(AvailabilityReplica.FlowControlTimems)",
			[]string{"instance", "replica"},
			nil,
		),
		ReceivesfromReplicaPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "receives_from_replica"),
			"(AvailabilityReplica.ReceivesfromReplica)",
			[]string{"instance", "replica"},
			nil,
		),
		ResentMessagesPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "resent_messages"),
			"(AvailabilityReplica.ResentMessages)",
			[]string{"instance", "replica"},
			nil,
		),
		SendstoReplicaPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "sends_to_replica"),
			"(AvailabilityReplica.SendstoReplica)",
			[]string{"instance", "replica"},
			nil,
		),
		SendstoTransportPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "sends_to_transport"),
			"(AvailabilityReplica.SendstoTransport)",
			[]string{"instance", "replica"},
			nil,
		),

		sqlInstances: getMSSQLInstances(),
	}, nil
}

// Collect sends the metric values for each metric
// to the provided prometheus Metric channel.
func (c *MSSQLAvailReplicaCollector) Collect(ch chan<- prometheus.Metric) error {
	for instance := range c.sqlInstances {
		log.Debugf("mssql_availreplica collector iterating sql instance %s.", instance)

		// Win32_PerfRawData_{instance}_SQLServerAvailabilityReplica
		if desc, err := c.collectAvailabilityReplica(ch, instance); err != nil {
			log.Error("failed collecting MSSQL GeneralStatistics metrics:", desc, err)
			return err
		}
	}

	return nil
}

type win32PerfRawDataSQLServerAvailabilityReplica struct {
	Name                           string
	BytesReceivedfromReplicaPersec uint64
	BytesSenttoReplicaPersec       uint64
	BytesSenttoTransportPersec     uint64
	FlowControlPersec              uint64
	FlowControlTimemsPersec        uint64
	ReceivesfromReplicaPersec      uint64
	ResentMessagesPersec           uint64
	SendstoReplicaPersec           uint64
	SendstoTransportPersec         uint64
}

func (c *MSSQLAvailReplicaCollector) collectAvailabilityReplica(ch chan<- prometheus.Metric, sqlInstance string) (*prometheus.Desc, error) {
	var dst []win32PerfRawDataSQLServerAvailabilityReplica
	class := fmt.Sprintf("Win32_PerfRawData_%s_SQLServerAvailabilityReplica", sqlInstance)
	q := queryAllForClassWhere(&dst, class, `Name <> '_Total'`)
	if err := wmi.Query(q, &dst); err != nil {
		return nil, err
	}

	for _, v := range dst {
		replicaName := v.Name

		ch <- prometheus.MustNewConstMetric(
			c.BytesReceivedfromReplicaPersec,
			prometheus.CounterValue,
			float64(v.BytesReceivedfromReplicaPersec),
			sqlInstance, replicaName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.BytesSenttoReplicaPersec,
			prometheus.CounterValue,
			float64(v.BytesSenttoReplicaPersec),
			sqlInstance, replicaName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.BytesSenttoTransportPersec,
			prometheus.CounterValue,
			float64(v.BytesSenttoTransportPersec),
			sqlInstance, replicaName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.FlowControlPersec,
			prometheus.CounterValue,
			float64(v.FlowControlPersec),
			sqlInstance, replicaName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.FlowControlTimemsPersec,
			prometheus.CounterValue,
			float64(v.FlowControlTimemsPersec)/1000.0,
			sqlInstance, replicaName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.ReceivesfromReplicaPersec,
			prometheus.CounterValue,
			float64(v.ReceivesfromReplicaPersec),
			sqlInstance, replicaName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.ResentMessagesPersec,
			prometheus.CounterValue,
			float64(v.ResentMessagesPersec),
			sqlInstance, replicaName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.SendstoReplicaPersec,
			prometheus.CounterValue,
			float64(v.SendstoReplicaPersec),
			sqlInstance, replicaName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.SendstoTransportPersec,
			prometheus.CounterValue,
			float64(v.SendstoTransportPersec),
			sqlInstance, replicaName,
		)
	}

	return nil, nil
}
