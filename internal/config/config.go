// SPDX-License-Identifier: Apache-2.0
//
// Copyright 2025 The Prometheus Authors
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

package config

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/alecthomas/kingpin/v2"
	"github.com/prometheus-community/windows_exporter/pkg/collector"
	"gopkg.in/yaml.v3"
)

// configFile represents the structure of the windows_exporter configuration file,
// including configuration from the collector and web packages.
type configFile struct {
	Debug struct {
		Enabled bool `yaml:"enabled"`
	} `yaml:"debug"`
	Collectors struct {
		Enabled string `yaml:"enabled"`
	} `yaml:"collectors"`
	Collector collector.Config `yaml:"collector"`
	Log       struct {
		Level  string `yaml:"level"`
		Format string `yaml:"format"`
		File   string `yaml:"file"`
	} `yaml:"log"`
	Process struct {
		Priority    string `yaml:"priority"`
		MemoryLimit string `yaml:"memory-limit"`
	} `yaml:"process"`
	Scrape struct {
		TimeoutMargin string `yaml:"timeout-margin"`
	} `yaml:"scrape"`
	Telemetry struct {
		Path string `yaml:"path"`
	} `yaml:"telemetry"`
	Web struct {
		DisableExporterMetrics bool     `yaml:"disable-exporter-metrics"`
		ListenAddresses        []string `yaml:"listen-address"`
		Config                 struct {
			File string `yaml:"file"`
		} `yaml:"config"`
	} `yaml:"web"`
}

type getFlagger interface {
	GetFlag(name string) *kingpin.FlagClause
}

// Resolver represents a configuration file resolver for kingpin.
type Resolver struct {
	flags map[string]string
}

// Parse parses the command line arguments and configuration files.
func Parse(app *kingpin.Application, args []string) error {
	configFile := ParseConfigFile(args)
	if configFile != "" {
		resolver, err := NewConfigFileResolver(configFile)
		if err != nil {
			return fmt.Errorf("failed to load configuration file: %w", err)
		}

		if err = resolver.Bind(app, args); err != nil {
			return fmt.Errorf("failed to bind configuration: %w", err)
		}
	}

	if _, err := app.Parse(args); err != nil {
		return fmt.Errorf("failed to parse flags: %w", err)
	}

	return nil
}

// ParseConfigFile manually parses the configuration file from the command line arguments.
func ParseConfigFile(args []string) string {
	for i, cliFlag := range args {
		if strings.HasPrefix(cliFlag, "--config.file=") {
			return strings.TrimPrefix(cliFlag, "--config.file=")
		}

		if strings.HasPrefix(cliFlag, "-config.file=") {
			return strings.TrimPrefix(cliFlag, "-config.file=")
		}

		if strings.HasSuffix(cliFlag, "-config.file") {
			if len(os.Args) <= i+1 {
				return ""
			}

			return os.Args[i+1]
		}
	}

	return ""
}

// NewConfigFileResolver returns a Resolver structure.
func NewConfigFileResolver(filePath string) (*Resolver, error) {
	flags := map[string]string{}

	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open configuration file: %w", err)
	}

	defer func() {
		_ = file.Close()
	}()

	var configFileStructure configFile

	decoder := yaml.NewDecoder(file)
	decoder.KnownFields(true)

	if err = decoder.Decode(&configFileStructure); err != nil {
		return nil, fmt.Errorf("configuration file validation error: %w", err)
	}

	_, err = file.Seek(0, io.SeekStart)
	if err != nil {
		return nil, fmt.Errorf("failed to rewind file: %w", err)
	}

	var rawValues map[string]interface{}

	decoder = yaml.NewDecoder(file)
	if err = decoder.Decode(&rawValues); err != nil {
		return nil, fmt.Errorf("failed to parse configuration file: %w", err)
	}

	// Flatten nested YAML values
	flattenedValues := flatten(rawValues)
	for k, v := range flattenedValues {
		if _, ok := flags[k]; !ok {
			flags[k] = v
		}
	}

	return &Resolver{flags: flags}, nil
}

func (c *Resolver) setDefault(v getFlagger) {
	for name, value := range c.flags {
		if f := v.GetFlag(name); f != nil {
			f.Default(value)
		}
	}
}

// Bind sets active flags with their default values from the configuration file(s).
func (c *Resolver) Bind(app *kingpin.Application, args []string) error {
	// Parse the command line arguments to get the selected command.
	pc, err := app.ParseContext(args)
	if err != nil {
		return err
	}

	c.setDefault(app)

	if pc.SelectedCommand != nil {
		c.setDefault(pc.SelectedCommand)
	}

	return nil
}
