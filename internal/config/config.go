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

package config

import (
	"errors"
	"fmt"
	"io"
	"maps"
	"os"
	"reflect"
	"slices"
	"strings"

	"github.com/alecthomas/kingpin/v2"
	"github.com/prometheus-community/windows_exporter/pkg/collector"
	"go.yaml.in/yaml/v3"
)

// configFile represents the structure of the windows_exporter configuration file,
// including configuration from the collector and web packages. Each leaf is a
// flag; the field types document the value of the flag, which parses and
// validates it. [validateKeys] checks the keys of the file against this schema.
type configFile struct {
	Debug struct {
		Enabled bool `yaml:"enabled"`
	} `yaml:"debug"`
	Collectors struct {
		Enabled  string `yaml:"enabled"`
		Disabled string `yaml:"disabled"`
	} `yaml:"collectors"`
	Collector collector.Config `yaml:"collector"`
	Log       struct {
		Level    string `yaml:"level"`
		Format   string `yaml:"format"`
		File     string `yaml:"file"`
		Rotation struct {
			MaxSize    int  `yaml:"max-size"`
			MaxBackups int  `yaml:"max-backups"`
			Compress   bool `yaml:"compress"`
		} `yaml:"rotation"`
	} `yaml:"log"`
	Process struct {
		Priority    string `yaml:"priority"`
		MemoryLimit int64  `yaml:"memory-limit"`
	} `yaml:"process"`
	Scrape struct {
		TimeoutMargin float64 `yaml:"timeout-margin"`
	} `yaml:"scrape"`
	Telemetry struct {
		Path string `yaml:"path"`
	} `yaml:"telemetry"`
	Web struct {
		DisableExporterMetrics bool `yaml:"disable-exporter-metrics"`
		ListenAddresses        any  `yaml:"listen-address"`
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
	flags map[string][]string
	// lines maps the keys of the configuration file to their lines.
	lines map[string]int
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
		if configFile, ok := strings.CutPrefix(cliFlag, "--config.file="); ok {
			return configFile
		}

		if configFile, ok := strings.CutPrefix(cliFlag, "-config.file="); ok {
			return configFile
		}

		if cliFlag == "--config.file" || cliFlag == "-config.file" {
			if len(args) <= i+1 {
				return ""
			}

			return args[i+1]
		}
	}

	return ""
}

// NewConfigFileResolver returns a Resolver structure.
func NewConfigFileResolver(filePath string) (*Resolver, error) {
	flags := map[string][]string{}

	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open configuration file: %w", err)
	}

	defer func() {
		_ = file.Close()
	}()

	var document yaml.Node

	if err = yaml.NewDecoder(file).Decode(&document); err != nil {
		// Handle EOF error gracefully, indicating no configuration was found.
		if errors.Is(err, io.EOF) {
			return &Resolver{flags: flags}, nil
		}

		return nil, fmt.Errorf("configuration file validation error: %w", err)
	}

	// Decoding first rejects duplicate keys and expands aliases within the
	// limits of the decoder before the schema walks the document.
	var rawValues map[string]any

	if err = document.Decode(&rawValues); err != nil {
		return nil, fmt.Errorf("failed to parse configuration file: %w", err)
	}

	if err = validateKeys(&document, "", reflect.TypeFor[configFile]()); err != nil {
		return nil, fmt.Errorf("configuration file validation error: %w", err)
	}

	// Flatten nested YAML values
	flattenedValues := flatten(rawValues)
	for k, v := range flattenedValues {
		if _, ok := flags[k]; !ok {
			flags[k] = v
		}
	}

	lines := map[string]int{}
	keyLines(&document, "", lines)

	return &Resolver{flags: flags, lines: lines}, nil
}

func (c *Resolver) setDefault(v getFlagger) error {
	for _, name := range slices.Sorted(maps.Keys(c.flags)) {
		values := c.flags[name]

		f := v.GetFlag(name)
		if f == nil {
			continue
		}

		if err := checkFlagValue(f, values); err != nil {
			// Some flag values, like those of utils.RegexpFlag, name their flag
			// in their errors, because kingpin doesn't for command line values.
			if strings.HasPrefix(err.Error(), name+": ") {
				return fmt.Errorf("line %d: %w", c.lines[name], err)
			}

			return fmt.Errorf("line %d: %s: %w", c.lines[name], name, err)
		}

		// A list sets a repeatable flag, like web.listen-address, once per element.
		// Other flags take the list as a comma-separated value.
		if repeatable, ok := f.Model().Value.(interface{ IsCumulative() bool }); ok && repeatable.IsCumulative() {
			f.Default(values...)
		} else {
			f.Default(strings.Join(values, ","))
		}
	}

	return nil
}

// Bind sets active flags with their default values from the configuration file(s).
func (c *Resolver) Bind(app *kingpin.Application, args []string) error {
	// Parse the command line arguments to get the selected command.
	pc, err := app.ParseContext(args)
	if err != nil {
		return err
	}

	if err = c.setDefault(app); err != nil {
		return err
	}

	if pc.SelectedCommand != nil {
		return c.setDefault(pc.SelectedCommand)
	}

	return nil
}
