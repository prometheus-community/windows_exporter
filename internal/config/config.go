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

package config

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strings"

	"github.com/alecthomas/kingpin/v2"
	"gopkg.in/yaml.v3"
)

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
func NewConfigFileResolver(file string) (*Resolver, error) {
	flags := map[string]string{}

	var (
		err       error
		fileBytes []byte
	)

	if strings.HasPrefix(file, "http://") || strings.HasPrefix(file, "https://") {
		//nolint:sloglint // we do not have an logger yet
		slog.Warn("Loading configuration file from URL is deprecated and will be removed in 0.31.0. Use a local file instead.")

		fileBytes, err = readFromURL(file)
		if err != nil {
			return nil, err
		}
	} else {
		fileBytes, err = readFromFile(file)
		if err != nil {
			return nil, err
		}
	}

	var rawValues map[string]interface{}

	err = yaml.Unmarshal(fileBytes, &rawValues)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal configuration file: %w", err)
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

func readFromFile(file string) ([]byte, error) {
	if _, err := os.Stat(file); err != nil {
		return nil, fmt.Errorf("failed to read configuration file: %w", err)
	}

	fileBytes, err := os.ReadFile(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read configuration file: %w", err)
	}

	return fileBytes, nil
}

func readFromURL(file string) ([]byte, error) {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, //nolint:gosec
	}

	client := &http.Client{Transport: tr}

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, file, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to read configuration file from URL: %w", err)
	}

	defer resp.Body.Close()

	fileBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return fileBytes, nil
}

func (c *Resolver) setDefault(v getFlagger) {
	for name, value := range c.flags {
		f := v.GetFlag(name)
		if f != nil {
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
