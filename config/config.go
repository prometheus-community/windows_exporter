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
	"os"

	"github.com/alecthomas/kingpin/v2"
	"github.com/prometheus-community/windows_exporter/log"
	"gopkg.in/yaml.v3"
)

type getFlagger interface {
	GetFlag(name string) *kingpin.FlagClause
}

// Resolver represents a configuration file resolver for kingpin.
type Resolver struct {
	flags map[string]string
}

type HookFunc func(interface{}) map[string]string

type ConfigHook struct {
	ConfigAttrs []string
	Hook        HookFunc
}

type ConfigHooks map[string]ConfigHook

func (c *ConfigHook) match(key string, val interface{}, level int) (bool, interface{}) {
	var ok bool = false
	var res interface{}
	if level < len(c.ConfigAttrs) {
		if c.ConfigAttrs[level] == key {
			level++
			if level < len(c.ConfigAttrs) {

				switch typed := val.(type) {
				case map[interface{}]interface{}:
					for fk, fv := range convertMap(typed) {
						ok, res = c.match(fk, fv, level)
						if ok {
							break
						}
					}
				case map[string]interface{}:
					for fk, fv := range typed {
						ok, res = c.match(fk, fv, level)
						if ok {
							break
						}
					}
				default:
				}
			} else {
				ok = true
				res = val
			}

		}
	}
	return ok, res
}

func (c *ConfigHooks) Match(key string, val interface{}) (map[string]interface{}, bool) {
	var ok bool = false
	var value interface{}
	params := make(map[string]interface{})

	for varname, hook := range *c {
		ok, value = hook.match(key, val, 0)
		if ok {
			params[varname] = hook.Hook(value)
		}
	}
	return params, ok
}

// NewResolver returns a Resolver structure.
func NewResolver(file string, hooks ConfigHooks) (*Resolver, error) {
	flags := map[string]string{}
	log.Infof("Loading configuration file: %v", file)
	if _, err := os.Stat(file); err != nil {
		return nil, err
	}
	b, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}

	var rawValues map[string]interface{}
	err = yaml.Unmarshal(b, &rawValues)
	if err != nil {
		return nil, err
	}

	if hooks != nil {
		for k, v := range rawValues {
			hooks.Match(k, v)
		}
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
