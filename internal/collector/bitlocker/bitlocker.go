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

package bitlocker

import (
	"errors"
	"fmt"
	"log/slog"

	"github.com/alecthomas/kingpin/v2"
	"github.com/prometheus-community/windows_exporter/internal/mi"
	"github.com/prometheus-community/windows_exporter/internal/types"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	Name = "bitlocker"
)

type Config struct{}

//nolint:gochecknoglobals
var ConfigDefaults = Config{}

// A Collector is a Prometheus Collector for WMI Win32_PerfRawData_DNS_DNS metrics.
type Collector struct {
	config Config

	miSession *mi.Session
	miQuery   mi.Query

	volumeInfo       *prometheus.Desc
	conversionStatus *prometheus.Desc
	encryptionMethod *prometheus.Desc
	protectionStatus *prometheus.Desc
}

func New(config *Config) *Collector {
	if config == nil {
		config = &ConfigDefaults
	}

	c := &Collector{
		config: *config,
	}

	return c
}

func NewWithFlags(_ *kingpin.Application) *Collector {
	c := &Collector{
		config: ConfigDefaults,
	}

	return c
}

func (c *Collector) GetName() string {
	return Name
}

func (c *Collector) Close() error {
	return nil
}

func (c *Collector) Build(_ *slog.Logger, miSession *mi.Session) error {
	if miSession == nil {
		return errors.New("miSession is nil")
	}

	query, err := mi.NewQuery("SELECT ConversionStatus, DeviceID, DriveLetter, EncryptionMethod, ProtectionStatus, VolumeType FROM Win32_EncryptableVolume")
	if err != nil {
		return fmt.Errorf("failed to create query: %w", err)
	}

	c.miSession = miSession
	c.miQuery = query

	c.volumeInfo = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "volume_info"),
		"Information about the encryptable volume.",
		[]string{"volume", "volume_path", "type"},
		nil,
	)

	c.conversionStatus = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "conversion_status"),
		"Encryption state of the volume.",
		[]string{"volume", "status"},
		nil,
	)

	c.encryptionMethod = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "encryption_method"),
		"Algorithm used to encrypt the volume.",
		[]string{"volume", "method"},
		nil,
	)

	c.protectionStatus = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "protection_status"),
		"Status of the volume, whether or not BitLocker is protecting the volume.",
		[]string{"volume", "status"},
		nil,
	)

	return nil
}

// Collect sends the metric values for each metric
// to the provided prometheus Metric channel.
func (c *Collector) Collect(ch chan<- prometheus.Metric) error {
	var encryptableVolume []EncryptableVolume
	if err := c.miSession.Query(&encryptableVolume, mi.NamespaceRootMicrosoftVolumeEncryption, c.miQuery); err != nil {
		return fmt.Errorf("failed to query MicrosoftVolumeEncryption: %w", err)
	}

	for _, volume := range encryptableVolume {
		ch <- prometheus.MustNewConstMetric(
			c.volumeInfo,
			prometheus.CounterValue,
			1,
			volume.DriveLetter,
			volume.DeviceID,
			volume.VolumeType.String(),
		)

		for i, status := range []string{
			"FULLY DECRYPTED",
			"FULLY ENCRYPTED",
			"ENCRYPTION IN PROGRESS",
			"DECRYPTION IN PROGRESS",
			"ENCRYPTION PAUSED",
			"DECRYPTION PAUSED",
		} {
			val := 0.0
			if volume.ConversionStatus == uint32(i) {
				val = 1.0
			}

			ch <- prometheus.MustNewConstMetric(
				c.conversionStatus,
				prometheus.GaugeValue,
				val,
				volume.DriveLetter,
				status,
			)
		}

		for i, status := range []string{
			"PROTECTION OFF",
			"PROTECTION ON",
			"PROTECTION UNKNOWN",
		} {
			val := 0.0
			if volume.ProtectionStatus == uint32(i) {
				val = 1.0
			}

			ch <- prometheus.MustNewConstMetric(
				c.protectionStatus,
				prometheus.GaugeValue,
				val,
				volume.DriveLetter,
				status,
			)
		}

		for i, status := range []string{
			"NOT ENCRYPTED",
			"AES 128 WITH DIFFUSER",
			"AES 256 WITH DIFFUSER",
			"AES 128",
			"AES 256",
			"HARDWARE ENCRYPTION",
			"XTS-AES 128",
			"XTS-AES 256 WITH DIFFUSER",
		} {
			val := 0.0
			if volume.EncryptionMethod == uint32(i) {
				val = 1.0
			}

			ch <- prometheus.MustNewConstMetric(
				c.encryptionMethod,
				prometheus.GaugeValue,
				val,
				volume.DriveLetter,
				status,
			)
		}
	}

	return nil
}
