package hyperv

import (
	"errors"
	"fmt"

	"github.com/prometheus-community/windows_exporter/internal/perfdata"
	"github.com/prometheus-community/windows_exporter/internal/types"
	"github.com/prometheus/client_golang/prometheus"
)

// collectorDataStore Hyper-V DataStore metrics
type collectorDataStore struct {
	perfDataCollectorDataStore *perfdata.Collector

	dataStoreFragmentationRatio          *prometheus.Desc // \Hyper-V DataStore(*)\Fragmentation ratio
	dataStoreSectorSize                  *prometheus.Desc // \Hyper-V DataStore(*)\Sector size
	dataStoreDataAlignment               *prometheus.Desc // \Hyper-V DataStore(*)\Data alignment
	dataStoreCurrentReplayLogSize        *prometheus.Desc // \Hyper-V DataStore(*)\Current replay logSize
	dataStoreAvailableEntries            *prometheus.Desc // \Hyper-V DataStore(*)\Number of available entries inside object tables
	dataStoreEmptyEntries                *prometheus.Desc // \Hyper-V DataStore(*)\Number of empty entries inside object tables
	dataStoreFreeBytes                   *prometheus.Desc // \Hyper-V DataStore(*)\Number of free bytes inside key tables
	dataStoreDataEnd                     *prometheus.Desc // \Hyper-V DataStore(*)\Data end
	dataStoreFileObjects                 *prometheus.Desc // \Hyper-V DataStore(*)\Number of file objects
	dataStoreObjectTables                *prometheus.Desc // \Hyper-V DataStore(*)\Number of object tables
	dataStoreKeyTables                   *prometheus.Desc // \Hyper-V DataStore(*)\Number of key tables
	dataStoreFileDataSize                *prometheus.Desc // \Hyper-V DataStore(*)\File data size in bytes
	dataStoreTableDataSize               *prometheus.Desc // \Hyper-V DataStore(*)\Table data size in bytes
	dataStoreNamesSize                   *prometheus.Desc // \Hyper-V DataStore(*)\Names size in bytes
	dataStoreNumberOfKeys                *prometheus.Desc // \Hyper-V DataStore(*)\Number of keys
	dataStoreReconnectLatencyMicro       *prometheus.Desc // \Hyper-V DataStore(*)\Reconnect latency microseconds
	dataStoreDisconnectCount             *prometheus.Desc // \Hyper-V DataStore(*)\Disconnect count
	dataStoreWriteToFileByteLatency      *prometheus.Desc // \Hyper-V DataStore(*)\Write to file byte latency microseconds
	dataStoreWriteToFileByteCount        *prometheus.Desc // \Hyper-V DataStore(*)\Write to file byte count
	dataStoreWriteToFileCount            *prometheus.Desc // \Hyper-V DataStore(*)\Write to file count
	dataStoreReadFromFileByteLatency     *prometheus.Desc // \Hyper-V DataStore(*)\Read from file byte latency microseconds
	dataStoreReadFromFileByteCount       *prometheus.Desc // \Hyper-V DataStore(*)\Read from file byte count
	dataStoreReadFromFileCount           *prometheus.Desc // \Hyper-V DataStore(*)\Read from file count
	dataStoreWriteToStorageByteLatency   *prometheus.Desc // \Hyper-V DataStore(*)\Write to storage byte latency microseconds
	dataStoreWriteToStorageByteCount     *prometheus.Desc // \Hyper-V DataStore(*)\Write to storage byte count
	dataStoreWriteToStorageCount         *prometheus.Desc // \Hyper-V DataStore(*)\Write to storage count
	dataStoreReadFromStorageByteLatency  *prometheus.Desc // \Hyper-V DataStore(*)\Read from storage byte latency microseconds
	dataStoreReadFromStorageByteCount    *prometheus.Desc // \Hyper-V DataStore(*)\Read from storage byte count
	dataStoreReadFromStorageCount        *prometheus.Desc // \Hyper-V DataStore(*)\Read from storage count
	dataStoreCommitByteLatency           *prometheus.Desc // \Hyper-V DataStore(*)\Commit byte latency microseconds
	dataStoreCommitByteCount             *prometheus.Desc // \Hyper-V DataStore(*)\Commit byte count
	dataStoreCommitCount                 *prometheus.Desc // \Hyper-V DataStore(*)\Commit count
	dataStoreCacheUpdateOperationLatency *prometheus.Desc // \Hyper-V DataStore(*)\Cache update operation latency microseconds
	dataStoreCacheUpdateOperationCount   *prometheus.Desc // \Hyper-V DataStore(*)\Cache update operation count
	dataStoreCommitOperationLatency      *prometheus.Desc // \Hyper-V DataStore(*)\Commit operation latency microseconds
	dataStoreCommitOperationCount        *prometheus.Desc // \Hyper-V DataStore(*)\Commit operation count
	dataStoreCompactOperationLatency     *prometheus.Desc // \Hyper-V DataStore(*)\Compact operation latency microseconds
	dataStoreCompactOperationCount       *prometheus.Desc // \Hyper-V DataStore(*)\Compact operation count
	dataStoreLoadFileOperationLatency    *prometheus.Desc // \Hyper-V DataStore(*)\Load file operation latency microseconds
	dataStoreLoadFileOperationCount      *prometheus.Desc // \Hyper-V DataStore(*)\Load file operation count
	dataStoreRemoveOperationLatency      *prometheus.Desc // \Hyper-V DataStore(*)\Remove operation latency microseconds
	dataStoreRemoveOperationCount        *prometheus.Desc // \Hyper-V DataStore(*)\Remove operation count
	dataStoreQuerySizeOperationLatency   *prometheus.Desc // \Hyper-V DataStore(*)\Query size operation latency microseconds
	dataStoreQuerySizeOperationCount     *prometheus.Desc // \Hyper-V DataStore(*)\Query size operation count
	dataStoreSetOperationLatencyMicro    *prometheus.Desc // \Hyper-V DataStore(*)\Set operation latency microseconds
	dataStoreSetOperationCount           *prometheus.Desc // \Hyper-V DataStore(*)\Set operation count
}

const (
	// Hyper-V DataStore metrics
	dataStoreFragmentationRatio          = "Fragmentation ratio"
	dataStoreSectorSize                  = "Sector size"
	dataStoreDataAlignment               = "Data alignment"
	dataStoreCurrentReplayLogSize        = "Current replay logSize"
	dataStoreAvailableEntries            = "Number of available entries inside object tables"
	dataStoreEmptyEntries                = "Number of empty entries inside object tables"
	dataStoreFreeBytes                   = "Number of free bytes inside key tables"
	dataStoreDataEnd                     = "Data end"
	dataStoreFileObjects                 = "Number of file objects"
	dataStoreObjectTables                = "Number of object tables"
	dataStoreKeyTables                   = "Number of key tables"
	dataStoreFileDataSize                = "File data size in bytes"
	dataStoreTableDataSize               = "Table data size in bytes"
	dataStoreNamesSize                   = "Names size in bytes"
	dataStoreNumberOfKeys                = "Number of keys"
	dataStoreReconnectLatencyMicro       = "Reconnect latency microseconds"
	dataStoreDisconnectCount             = "Disconnect count"
	dataStoreWriteToFileByteLatency      = "Write to file byte latency microseconds"
	dataStoreWriteToFileByteCount        = "Write to file byte count"
	dataStoreWriteToFileCount            = "Write to file count"
	dataStoreReadFromFileByteLatency     = "Read from file byte latency microseconds"
	dataStoreReadFromFileByteCount       = "Read from file byte count"
	dataStoreReadFromFileCount           = "Read from file count"
	dataStoreWriteToStorageByteLatency   = "Write to storage byte latency microseconds"
	dataStoreWriteToStorageByteCount     = "Write to storage byte count"
	dataStoreWriteToStorageCount         = "Write to storage count"
	dataStoreReadFromStorageByteLatency  = "Read from storage byte latency microseconds"
	dataStoreReadFromStorageByteCount    = "Read from storage byte count"
	dataStoreReadFromStorageCount        = "Read from storage count"
	dataStoreCommitByteLatency           = "Commit byte latency microseconds"
	dataStoreCommitByteCount             = "Commit byte count"
	dataStoreCommitCount                 = "Commit count"
	dataStoreCacheUpdateOperationLatency = "Cache update operation latency microseconds"
	dataStoreCacheUpdateOperationCount   = "Cache update operation count"
	dataStoreCommitOperationLatency      = "Commit operation latency microseconds"
	dataStoreCommitOperationCount        = "Commit operation count"
	dataStoreCompactOperationLatency     = "Compact operation latency microseconds"
	dataStoreCompactOperationCount       = "Compact operation count"
	dataStoreLoadFileOperationLatency    = "Load file operation latency microseconds"
	dataStoreLoadFileOperationCount      = "Load file operation count"
	dataStoreRemoveOperationLatency      = "Remove operation latency microseconds"
	dataStoreRemoveOperationCount        = "Remove operation count"
	dataStoreQuerySizeOperationLatency   = "Query size operation latency microseconds"
	dataStoreQuerySizeOperationCount     = "Query size operation count"
	dataStoreSetOperationLatencyMicro    = "Set operation latency microseconds"
	dataStoreSetOperationCount           = "Set operation count"
)

func (c *Collector) buildDataStore() error {
	var err error

	c.perfDataCollectorDataStore, err = perfdata.NewCollector("Hyper-V DataStore", perfdata.InstanceAll, []string{
		dataStoreFragmentationRatio,
		dataStoreSectorSize,
		dataStoreDataAlignment,
		dataStoreCurrentReplayLogSize,
		dataStoreAvailableEntries,
		dataStoreEmptyEntries,
		dataStoreFreeBytes,
		dataStoreDataEnd,
		dataStoreFileObjects,
		dataStoreObjectTables,
		dataStoreKeyTables,
		dataStoreFileDataSize,
		dataStoreTableDataSize,
		dataStoreNamesSize,
		dataStoreNumberOfKeys,
		dataStoreReconnectLatencyMicro,
		dataStoreDisconnectCount,
		dataStoreWriteToFileByteLatency,
		dataStoreWriteToFileByteCount,
		dataStoreWriteToFileCount,
		dataStoreReadFromFileByteLatency,
		dataStoreReadFromFileByteCount,
		dataStoreReadFromFileCount,
		dataStoreWriteToStorageByteLatency,
		dataStoreWriteToStorageByteCount,
		dataStoreWriteToStorageCount,
		dataStoreReadFromStorageByteLatency,
		dataStoreReadFromStorageByteCount,
		dataStoreReadFromStorageCount,
		dataStoreCommitByteLatency,
		dataStoreCommitByteCount,
		dataStoreCommitCount,
		dataStoreCacheUpdateOperationLatency,
		dataStoreCacheUpdateOperationCount,
		dataStoreCommitOperationLatency,
		dataStoreCommitOperationCount,
		dataStoreCompactOperationLatency,
		dataStoreCompactOperationCount,
		dataStoreLoadFileOperationLatency,
		dataStoreLoadFileOperationCount,
		dataStoreRemoveOperationLatency,
		dataStoreRemoveOperationCount,
		dataStoreQuerySizeOperationLatency,
		dataStoreQuerySizeOperationCount,
		dataStoreSetOperationLatencyMicro,
		dataStoreSetOperationCount,
	})
	if err != nil && !errors.Is(err, perfdata.ErrNoData) {
		return fmt.Errorf("failed to create Hyper-V DataStore collector: %w", err)
	}

	c.dataStoreFragmentationRatio = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "datastore_fragmentation_ratio"),
		"Represents the fragmentation ratio of the DataStore.",
		[]string{"datastore"},
		nil,
	)
	c.dataStoreSectorSize = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "datastore_sector_size_bytes"),
		"Represents the sector size of the DataStore.",
		[]string{"datastore"},
		nil,
	)
	c.dataStoreDataAlignment = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "datastore_data_alignment_bytes"),
		"Represents the data alignment of the DataStore.",
		[]string{"datastore"},
		nil,
	)
	c.dataStoreCurrentReplayLogSize = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "datastore_current_replay_log_size_bytes"),
		"Represents the current replay log size of the DataStore.",
		[]string{"datastore"},
		nil,
	)
	c.dataStoreAvailableEntries = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "datastore_available_entries"),
		"Represents the number of available entries inside object tables.",
		[]string{"datastore"},
		nil,
	)
	c.dataStoreEmptyEntries = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "datastore_empty_entries"),
		"Represents the number of empty entries inside object tables.",
		[]string{"datastore"},
		nil,
	)
	c.dataStoreFreeBytes = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "datastore_free_bytes"),
		"Represents the number of free bytes inside key tables.",
		[]string{"datastore"},
		nil,
	)
	c.dataStoreDataEnd = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "datastore_data_end_bytes"),
		"Represents the data end of the DataStore.",
		[]string{"datastore"},
		nil,
	)
	c.dataStoreFileObjects = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "datastore_file_objects"),
		"Represents the number of file objects in the DataStore.",
		[]string{"datastore"},
		nil,
	)
	c.dataStoreObjectTables = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "datastore_object_tables"),
		"Represents the number of object tables in the DataStore.",
		[]string{"datastore"},
		nil,
	)
	c.dataStoreKeyTables = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "datastore_key_tables"),
		"Represents the number of key tables in the DataStore.",
		[]string{"datastore"},
		nil,
	)
	c.dataStoreFileDataSize = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "datastore_file_data_size_bytes"),
		"Represents the file data size in bytes of the DataStore.",
		[]string{"datastore"},
		nil,
	)
	c.dataStoreTableDataSize = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "datastore_table_data_size_bytes"),
		"Represents the table data size in bytes of the DataStore.",
		[]string{"datastore"},
		nil,
	)
	c.dataStoreNamesSize = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "datastore_names_size_bytes"),
		"Represents the names size in bytes of the DataStore.",
		[]string{"datastore"},
		nil,
	)
	c.dataStoreNumberOfKeys = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "datastore_number_of_keys"),
		"Represents the number of keys in the DataStore.",
		[]string{"datastore"},
		nil,
	)
	c.dataStoreReconnectLatencyMicro = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "datastore_reconnect_latency_microseconds"),
		"Represents the reconnect latency in microseconds of the DataStore.",
		[]string{"datastore"},
		nil,
	)
	c.dataStoreDisconnectCount = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "datastore_disconnect_count"),
		"Represents the disconnect count of the DataStore.",
		[]string{"datastore"},
		nil,
	)
	c.dataStoreWriteToFileByteLatency = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "datastore_write_to_file_byte_latency_microseconds"),
		"Represents the write to file byte latency in microseconds of the DataStore.",
		[]string{"datastore"},
		nil,
	)
	c.dataStoreWriteToFileByteCount = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "datastore_write_to_file_byte_count"),
		"Represents the write to file byte count of the DataStore.",
		[]string{"datastore"},
		nil,
	)
	c.dataStoreWriteToFileCount = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "datastore_write_to_file_count"),
		"Represents the write to file count of the DataStore.",
		[]string{"datastore"},
		nil,
	)
	c.dataStoreReadFromFileByteLatency = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "datastore_read_from_file_byte_latency_microseconds"),
		"Represents the read from file byte latency in microseconds of the DataStore.",
		[]string{"datastore"},
		nil,
	)
	c.dataStoreReadFromFileByteCount = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "datastore_read_from_file_byte_count"),
		"Represents the read from file byte count of the DataStore.",
		[]string{"datastore"},
		nil,
	)
	c.dataStoreReadFromFileCount = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "datastore_read_from_file_count"),
		"Represents the read from file count of the DataStore.",
		[]string{"datastore"},
		nil,
	)
	c.dataStoreWriteToStorageByteLatency = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "datastore_write_to_storage_byte_latency_microseconds"),
		"Represents the write to storage byte latency in microseconds of the DataStore.",
		[]string{"datastore"},
		nil,
	)
	c.dataStoreWriteToStorageByteCount = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "datastore_write_to_storage_byte_count"),
		"Represents the write to storage byte count of the DataStore.",
		[]string{"datastore"},
		nil,
	)
	c.dataStoreWriteToStorageCount = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "datastore_write_to_storage_count"),
		"Represents the write to storage count of the DataStore.",
		[]string{"datastore"},
		nil,
	)
	c.dataStoreReadFromStorageByteLatency = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "datastore_read_from_storage_byte_latency_microseconds"),
		"Represents the read from storage byte latency in microseconds of the DataStore.",
		[]string{"datastore"},
		nil,
	)
	c.dataStoreReadFromStorageByteCount = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "datastore_read_from_storage_byte_count"),
		"Represents the read from storage byte count of the DataStore.",
		[]string{"datastore"},
		nil,
	)
	c.dataStoreReadFromStorageCount = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "datastore_read_from_storage_count"),
		"Represents the read from storage count of the DataStore.",
		[]string{"datastore"},
		nil,
	)
	c.dataStoreCommitByteLatency = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "datastore_commit_byte_latency_microseconds"),
		"Represents the commit byte latency in microseconds of the DataStore.",
		[]string{"datastore"},
		nil,
	)
	c.dataStoreCommitByteCount = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "datastore_commit_byte_count"),
		"Represents the commit byte count of the DataStore.",
		[]string{"datastore"},
		nil,
	)
	c.dataStoreCommitCount = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "datastore_commit_count"),
		"Represents the commit count of the DataStore.",
		[]string{"datastore"},
		nil,
	)
	c.dataStoreCacheUpdateOperationLatency = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "datastore_cache_update_operation_latency_microseconds"),
		"Represents the cache update operation latency in microseconds of the DataStore.",
		[]string{"datastore"},
		nil,
	)
	c.dataStoreCacheUpdateOperationCount = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "datastore_cache_update_operation_count"),
		"Represents the cache update operation count of the DataStore.",
		[]string{"datastore"},
		nil,
	)
	c.dataStoreCommitOperationLatency = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "datastore_commit_operation_latency_microseconds"),
		"Represents the commit operation latency in microseconds of the DataStore.",
		[]string{"datastore"},
		nil,
	)
	c.dataStoreCommitOperationCount = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "datastore_commit_operation_count"),
		"Represents the commit operation count of the DataStore.",
		[]string{"datastore"},
		nil,
	)
	c.dataStoreCompactOperationLatency = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "datastore_compact_operation_latency_microseconds"),
		"Represents the compact operation latency in microseconds of the DataStore.",
		[]string{"datastore"},
		nil,
	)
	c.dataStoreCompactOperationCount = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "datastore_compact_operation_count"),
		"Represents the compact operation count of the DataStore.",
		[]string{"datastore"},
		nil,
	)
	c.dataStoreLoadFileOperationLatency = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "datastore_load_file_operation_latency_microseconds"),
		"Represents the load file operation latency in microseconds of the DataStore.",
		[]string{"datastore"},
		nil,
	)
	c.dataStoreLoadFileOperationCount = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "datastore_load_file_operation_count"),
		"Represents the load file operation count of the DataStore.",
		[]string{"datastore"},
		nil,
	)
	c.dataStoreRemoveOperationLatency = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "datastore_remove_operation_latency_microseconds"),
		"Represents the remove operation latency in microseconds of the DataStore.",
		[]string{"datastore"},
		nil,
	)
	c.dataStoreRemoveOperationCount = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "datastore_remove_operation_count"),
		"Represents the remove operation count of the DataStore.",
		[]string{"datastore"},
		nil,
	)
	c.dataStoreQuerySizeOperationLatency = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "datastore_query_size_operation_latency_microseconds"),
		"Represents the query size operation latency in microseconds of the DataStore.",
		[]string{"datastore"},
		nil,
	)
	c.dataStoreQuerySizeOperationCount = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "datastore_query_size_operation_count"),
		"Represents the query size operation count of the DataStore.",
		[]string{"datastore"},
		nil,
	)
	c.dataStoreSetOperationLatencyMicro = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "datastore_set_operation_latency_microseconds"),
		"Represents the set operation latency in microseconds of the DataStore.",
		[]string{"datastore"},
		nil,
	)
	c.dataStoreSetOperationCount = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "datastore_set_operation_count"),
		"Represents the set operation count of the DataStore.",
		[]string{"datastore"},
		nil,
	)

	return nil
}

func (c *Collector) collectDataStore(ch chan<- prometheus.Metric) error {
	data, err := c.perfDataCollectorDataStore.Collect()
	if err != nil && !errors.Is(err, perfdata.ErrNoData) {
		return fmt.Errorf("failed to collect Hyper-V DataStore metrics: %w", err)
	}

	for name, page := range data {
		ch <- prometheus.MustNewConstMetric(
			c.dataStoreFragmentationRatio,
			prometheus.GaugeValue,
			page[dataStoreFragmentationRatio].FirstValue,
			name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.dataStoreSectorSize,
			prometheus.GaugeValue,
			page[dataStoreSectorSize].FirstValue,
			name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.dataStoreDataAlignment,
			prometheus.GaugeValue,
			page[dataStoreDataAlignment].FirstValue,
			name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.dataStoreCurrentReplayLogSize,
			prometheus.GaugeValue,
			page[dataStoreCurrentReplayLogSize].FirstValue,
			name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.dataStoreAvailableEntries,
			prometheus.GaugeValue,
			page[dataStoreAvailableEntries].FirstValue,
			name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.dataStoreEmptyEntries,
			prometheus.GaugeValue,
			page[dataStoreEmptyEntries].FirstValue,
			name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.dataStoreFreeBytes,
			prometheus.GaugeValue,
			page[dataStoreFreeBytes].FirstValue,
			name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.dataStoreDataEnd,
			prometheus.GaugeValue,
			page[dataStoreDataEnd].FirstValue,
			name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.dataStoreFileObjects,
			prometheus.GaugeValue,
			page[dataStoreFileObjects].FirstValue,
			name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.dataStoreObjectTables,
			prometheus.GaugeValue,
			page[dataStoreObjectTables].FirstValue,
			name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.dataStoreKeyTables,
			prometheus.GaugeValue,
			page[dataStoreKeyTables].FirstValue,
			name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.dataStoreFileDataSize,
			prometheus.GaugeValue,
			page[dataStoreFileDataSize].FirstValue,
			name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.dataStoreTableDataSize,
			prometheus.GaugeValue,
			page[dataStoreTableDataSize].FirstValue,
			name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.dataStoreNamesSize,
			prometheus.GaugeValue,
			page[dataStoreNamesSize].FirstValue,
			name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.dataStoreNumberOfKeys,
			prometheus.GaugeValue,
			page[dataStoreNumberOfKeys].FirstValue,
			name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.dataStoreReconnectLatencyMicro,
			prometheus.GaugeValue,
			page[dataStoreReconnectLatencyMicro].FirstValue,
			name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.dataStoreDisconnectCount,
			prometheus.CounterValue,
			page[dataStoreDisconnectCount].FirstValue,
			name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.dataStoreWriteToFileByteLatency,
			prometheus.GaugeValue,
			page[dataStoreWriteToFileByteLatency].FirstValue,
			name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.dataStoreWriteToFileByteCount,
			prometheus.CounterValue,
			page[dataStoreWriteToFileByteCount].FirstValue,
			name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.dataStoreWriteToFileCount,
			prometheus.CounterValue,
			page[dataStoreWriteToFileCount].FirstValue,
			name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.dataStoreReadFromFileByteLatency,
			prometheus.GaugeValue,
			page[dataStoreReadFromFileByteLatency].FirstValue,
			name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.dataStoreReadFromFileByteCount,
			prometheus.CounterValue,
			page[dataStoreReadFromFileByteCount].FirstValue,
			name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.dataStoreReadFromFileCount,
			prometheus.CounterValue,
			page[dataStoreReadFromFileCount].FirstValue,
			name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.dataStoreWriteToStorageByteLatency,
			prometheus.GaugeValue,
			page[dataStoreWriteToStorageByteLatency].FirstValue,
			name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.dataStoreWriteToStorageByteCount,
			prometheus.CounterValue,
			page[dataStoreWriteToStorageByteCount].FirstValue,
			name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.dataStoreWriteToStorageCount,
			prometheus.CounterValue,
			page[dataStoreWriteToStorageCount].FirstValue,
			name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.dataStoreReadFromStorageByteLatency,
			prometheus.GaugeValue,
			page[dataStoreReadFromStorageByteLatency].FirstValue,
			name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.dataStoreReadFromStorageByteCount,
			prometheus.CounterValue,
			page[dataStoreReadFromStorageByteCount].FirstValue,
			name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.dataStoreReadFromStorageCount,
			prometheus.CounterValue,
			page[dataStoreReadFromStorageCount].FirstValue,
			name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.dataStoreCommitByteLatency,
			prometheus.GaugeValue,
			page[dataStoreCommitByteLatency].FirstValue,
			name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.dataStoreCommitByteCount,
			prometheus.CounterValue,
			page[dataStoreCommitByteCount].FirstValue,
			name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.dataStoreCommitCount,
			prometheus.CounterValue,
			page[dataStoreCommitCount].FirstValue,
			name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.dataStoreCacheUpdateOperationLatency,
			prometheus.GaugeValue,
			page[dataStoreCacheUpdateOperationLatency].FirstValue,
			name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.dataStoreCacheUpdateOperationCount,
			prometheus.CounterValue,
			page[dataStoreCacheUpdateOperationCount].FirstValue,
			name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.dataStoreCommitOperationLatency,
			prometheus.GaugeValue,
			page[dataStoreCommitOperationLatency].FirstValue,
			name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.dataStoreCommitOperationCount,
			prometheus.CounterValue,
			page[dataStoreCommitOperationCount].FirstValue,
			name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.dataStoreCompactOperationLatency,
			prometheus.GaugeValue,
			page[dataStoreCompactOperationLatency].FirstValue,
			name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.dataStoreCompactOperationCount,
			prometheus.CounterValue,
			page[dataStoreCompactOperationCount].FirstValue,
			name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.dataStoreLoadFileOperationLatency,
			prometheus.GaugeValue,
			page[dataStoreLoadFileOperationLatency].FirstValue,
			name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.dataStoreLoadFileOperationCount,
			prometheus.CounterValue,
			page[dataStoreLoadFileOperationCount].FirstValue,
			name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.dataStoreRemoveOperationLatency,
			prometheus.GaugeValue,
			page[dataStoreRemoveOperationLatency].FirstValue,
			name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.dataStoreRemoveOperationCount,
			prometheus.CounterValue,
			page[dataStoreRemoveOperationCount].FirstValue,
			name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.dataStoreQuerySizeOperationLatency,
			prometheus.GaugeValue,
			page[dataStoreQuerySizeOperationLatency].FirstValue,
			name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.dataStoreQuerySizeOperationCount,
			prometheus.CounterValue,
			page[dataStoreQuerySizeOperationCount].FirstValue,
			name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.dataStoreSetOperationLatencyMicro,
			prometheus.GaugeValue,
			page[dataStoreSetOperationLatencyMicro].FirstValue,
			name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.dataStoreSetOperationCount,
			prometheus.CounterValue,
			page[dataStoreSetOperationCount].FirstValue,
			name,
		)
	}

	return nil
}
