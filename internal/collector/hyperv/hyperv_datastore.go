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

package hyperv

import (
	"fmt"

	"github.com/prometheus-community/windows_exporter/internal/pdh"
	"github.com/prometheus-community/windows_exporter/internal/types"
	"github.com/prometheus/client_golang/prometheus"
)

// collectorDataStore Hyper-V DataStore metrics
type collectorDataStore struct {
	perfDataCollectorDataStore *pdh.Collector
	perfDataObjectDataStore    []perfDataCounterValuesDataStore

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

type perfDataCounterValuesDataStore struct {
	Name string
	// Hyper-V DataStore metrics
	DataStoreFragmentationRatio          float64 `perfdata:"Fragmentation ratio"`
	DataStoreSectorSize                  float64 `perfdata:"Sector size"`
	DataStoreDataAlignment               float64 `perfdata:"Data alignment"`
	DataStoreCurrentReplayLogSize        float64 `perfdata:"Current replay logSize"`
	DataStoreAvailableEntries            float64 `perfdata:"Number of available entries inside object tables"`
	DataStoreEmptyEntries                float64 `perfdata:"Number of empty entries inside object tables"`
	DataStoreFreeBytes                   float64 `perfdata:"Number of free bytes inside key tables"`
	DataStoreDataEnd                     float64 `perfdata:"Data end"`
	DataStoreFileObjects                 float64 `perfdata:"Number of file objects"`
	DataStoreObjectTables                float64 `perfdata:"Number of object tables"`
	DataStoreKeyTables                   float64 `perfdata:"Number of key tables"`
	DataStoreFileDataSize                float64 `perfdata:"File data size in bytes"`
	DataStoreTableDataSize               float64 `perfdata:"Table data size in bytes"`
	DataStoreNamesSize                   float64 `perfdata:"Names size in bytes"`
	DataStoreNumberOfKeys                float64 `perfdata:"Number of keys"`
	DataStoreReconnectLatencyMicro       float64 `perfdata:"Reconnect latency microseconds"`
	DataStoreDisconnectCount             float64 `perfdata:"Disconnect count"`
	DataStoreWriteToFileByteLatency      float64 `perfdata:"Write to file byte latency microseconds"`
	DataStoreWriteToFileByteCount        float64 `perfdata:"Write to file byte count"`
	DataStoreWriteToFileCount            float64 `perfdata:"Write to file count"`
	DataStoreReadFromFileByteLatency     float64 `perfdata:"Read from file byte latency microseconds"`
	DataStoreReadFromFileByteCount       float64 `perfdata:"Read from file byte count"`
	DataStoreReadFromFileCount           float64 `perfdata:"Read from file count"`
	DataStoreWriteToStorageByteLatency   float64 `perfdata:"Write to storage byte latency microseconds"`
	DataStoreWriteToStorageByteCount     float64 `perfdata:"Write to storage byte count"`
	DataStoreWriteToStorageCount         float64 `perfdata:"Write to storage count"`
	DataStoreReadFromStorageByteLatency  float64 `perfdata:"Read from storage byte latency microseconds"`
	DataStoreReadFromStorageByteCount    float64 `perfdata:"Read from storage byte count"`
	DataStoreReadFromStorageCount        float64 `perfdata:"Read from storage count"`
	DataStoreCommitByteLatency           float64 `perfdata:"Commit byte latency microseconds"`
	DataStoreCommitByteCount             float64 `perfdata:"Commit byte count"`
	DataStoreCommitCount                 float64 `perfdata:"Commit count"`
	DataStoreCacheUpdateOperationLatency float64 `perfdata:"Cache update operation latency microseconds"`
	DataStoreCacheUpdateOperationCount   float64 `perfdata:"Cache update operation count"`
	DataStoreCommitOperationLatency      float64 `perfdata:"Commit operation latency microseconds"`
	DataStoreCommitOperationCount        float64 `perfdata:"Commit operation count"`
	DataStoreCompactOperationLatency     float64 `perfdata:"Compact operation latency microseconds"`
	DataStoreCompactOperationCount       float64 `perfdata:"Compact operation count"`
	DataStoreLoadFileOperationLatency    float64 `perfdata:"Load file operation latency microseconds"`
	DataStoreLoadFileOperationCount      float64 `perfdata:"Load file operation count"`
	DataStoreRemoveOperationLatency      float64 `perfdata:"Remove operation latency microseconds"`
	DataStoreRemoveOperationCount        float64 `perfdata:"Remove operation count"`
	DataStoreQuerySizeOperationLatency   float64 `perfdata:"Query size operation latency microseconds"`
	DataStoreQuerySizeOperationCount     float64 `perfdata:"Query size operation count"`
	DataStoreSetOperationLatencyMicro    float64 `perfdata:"Set operation latency microseconds"`
	DataStoreSetOperationCount           float64 `perfdata:"Set operation count"`
}

func (c *Collector) buildDataStore() error {
	var err error

	c.perfDataCollectorDataStore, err = pdh.NewCollector[perfDataCounterValuesDataStore](pdh.CounterTypeRaw, "Hyper-V DataStore", pdh.InstancesAll)
	if err != nil {
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
	err := c.perfDataCollectorDataStore.Collect(&c.perfDataObjectDataStore)
	if err != nil {
		return fmt.Errorf("failed to collect Hyper-V DataStore metrics: %w", err)
	}

	for _, data := range c.perfDataObjectDataStore {
		ch <- prometheus.MustNewConstMetric(
			c.dataStoreFragmentationRatio,
			prometheus.GaugeValue,
			data.DataStoreFragmentationRatio,
			data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.dataStoreSectorSize,
			prometheus.GaugeValue,
			data.DataStoreSectorSize,
			data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.dataStoreDataAlignment,
			prometheus.GaugeValue,
			data.DataStoreDataAlignment,
			data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.dataStoreCurrentReplayLogSize,
			prometheus.GaugeValue,
			data.DataStoreCurrentReplayLogSize,
			data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.dataStoreAvailableEntries,
			prometheus.GaugeValue,
			data.DataStoreAvailableEntries,
			data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.dataStoreEmptyEntries,
			prometheus.GaugeValue,
			data.DataStoreEmptyEntries,
			data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.dataStoreFreeBytes,
			prometheus.GaugeValue,
			data.DataStoreFreeBytes,
			data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.dataStoreDataEnd,
			prometheus.GaugeValue,
			data.DataStoreDataEnd,
			data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.dataStoreFileObjects,
			prometheus.GaugeValue,
			data.DataStoreFileObjects,
			data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.dataStoreObjectTables,
			prometheus.GaugeValue,
			data.DataStoreObjectTables,
			data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.dataStoreKeyTables,
			prometheus.GaugeValue,
			data.DataStoreKeyTables,
			data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.dataStoreFileDataSize,
			prometheus.GaugeValue,
			data.DataStoreFileDataSize,
			data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.dataStoreTableDataSize,
			prometheus.GaugeValue,
			data.DataStoreTableDataSize,
			data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.dataStoreNamesSize,
			prometheus.GaugeValue,
			data.DataStoreNamesSize,
			data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.dataStoreNumberOfKeys,
			prometheus.GaugeValue,
			data.DataStoreNumberOfKeys,
			data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.dataStoreReconnectLatencyMicro,
			prometheus.GaugeValue,
			data.DataStoreReconnectLatencyMicro,
			data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.dataStoreDisconnectCount,
			prometheus.CounterValue,
			data.DataStoreDisconnectCount,
			data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.dataStoreWriteToFileByteLatency,
			prometheus.GaugeValue,
			data.DataStoreWriteToFileByteLatency,
			data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.dataStoreWriteToFileByteCount,
			prometheus.CounterValue,
			data.DataStoreWriteToFileByteCount,
			data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.dataStoreWriteToFileCount,
			prometheus.CounterValue,
			data.DataStoreWriteToFileCount,
			data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.dataStoreReadFromFileByteLatency,
			prometheus.GaugeValue,
			data.DataStoreReadFromFileByteLatency,
			data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.dataStoreReadFromFileByteCount,
			prometheus.CounterValue,
			data.DataStoreReadFromFileByteCount,
			data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.dataStoreReadFromFileCount,
			prometheus.CounterValue,
			data.DataStoreReadFromFileCount,
			data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.dataStoreWriteToStorageByteLatency,
			prometheus.GaugeValue,
			data.DataStoreWriteToStorageByteLatency,
			data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.dataStoreWriteToStorageByteCount,
			prometheus.CounterValue,
			data.DataStoreWriteToStorageByteCount,
			data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.dataStoreWriteToStorageCount,
			prometheus.CounterValue,
			data.DataStoreWriteToStorageCount,
			data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.dataStoreReadFromStorageByteLatency,
			prometheus.GaugeValue,
			data.DataStoreReadFromStorageByteLatency,
			data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.dataStoreReadFromStorageByteCount,
			prometheus.CounterValue,
			data.DataStoreReadFromStorageByteCount,
			data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.dataStoreReadFromStorageCount,
			prometheus.CounterValue,
			data.DataStoreReadFromStorageCount,
			data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.dataStoreCommitByteLatency,
			prometheus.GaugeValue,
			data.DataStoreCommitByteLatency,
			data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.dataStoreCommitByteCount,
			prometheus.CounterValue,
			data.DataStoreCommitByteCount,
			data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.dataStoreCommitCount,
			prometheus.CounterValue,
			data.DataStoreCommitCount,
			data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.dataStoreCacheUpdateOperationLatency,
			prometheus.GaugeValue,
			data.DataStoreCacheUpdateOperationLatency,
			data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.dataStoreCacheUpdateOperationCount,
			prometheus.CounterValue,
			data.DataStoreCacheUpdateOperationCount,
			data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.dataStoreCommitOperationLatency,
			prometheus.GaugeValue,
			data.DataStoreCommitOperationLatency,
			data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.dataStoreCommitOperationCount,
			prometheus.CounterValue,
			data.DataStoreCommitOperationCount,
			data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.dataStoreCompactOperationLatency,
			prometheus.GaugeValue,
			data.DataStoreCompactOperationLatency,
			data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.dataStoreCompactOperationCount,
			prometheus.CounterValue,
			data.DataStoreCompactOperationCount,
			data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.dataStoreLoadFileOperationLatency,
			prometheus.GaugeValue,
			data.DataStoreLoadFileOperationLatency,
			data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.dataStoreLoadFileOperationCount,
			prometheus.CounterValue,
			data.DataStoreLoadFileOperationCount,
			data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.dataStoreRemoveOperationLatency,
			prometheus.GaugeValue,
			data.DataStoreRemoveOperationLatency,
			data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.dataStoreRemoveOperationCount,
			prometheus.CounterValue,
			data.DataStoreRemoveOperationCount,
			data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.dataStoreQuerySizeOperationLatency,
			prometheus.GaugeValue,
			data.DataStoreQuerySizeOperationLatency,
			data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.dataStoreQuerySizeOperationCount,
			prometheus.CounterValue,
			data.DataStoreQuerySizeOperationCount,
			data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.dataStoreSetOperationLatencyMicro,
			prometheus.GaugeValue,
			data.DataStoreSetOperationLatencyMicro,
			data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.dataStoreSetOperationCount,
			prometheus.CounterValue,
			data.DataStoreSetOperationCount,
			data.Name,
		)
	}

	return nil
}
