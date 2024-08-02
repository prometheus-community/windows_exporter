// Copyright 2018 Prometheus Team
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

package config

import (
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/alecthomas/kingpin/v2"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"gopkg.in/yaml.v3"
)

type getFlagger interface {
	GetFlag(name string) *kingpin.FlagClause
}

// Resolver represents a configuration file resolver for kingpin.
type Resolver struct {
	flags map[string]string
}

// NewResolver returns a Resolver structure.
func NewResolver(file string, logger log.Logger, insecureSkipVerify bool) (*Resolver, error) {
	flags := map[string]string{}
	var fileBytes []byte
	var err error
	if strings.HasPrefix(file, "http://") || strings.HasPrefix(file, "https://") {
		_ = level.Info(logger).Log("msg", fmt.Sprintf("Loading configuration file from URL: %v", file))
		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: insecureSkipVerify}, //nolint:gosec
		}
		if insecureSkipVerify {
			_ = level.Warn(logger).Log("msg", "Loading configuration file with TLS verification disabled")
		}
		client := &http.Client{Transport: tr}
		resp, err := client.Get(file)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()
		fileBytes, err = io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
	} else {
		_ = level.Info(logger).Log("msg", fmt.Sprintf("Loading configuration file: %v", file))
		if _, err := os.Stat(file); err != nil {
			return nil, err
		}
		fileBytes, err = os.ReadFile(file)
		if err != nil {
			return nil, err
		}
	}

	var rawValues map[string]interface{}
	err = yaml.Unmarshal(fileBytes, &rawValues)
	if err != nil {
		return nil, err
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
