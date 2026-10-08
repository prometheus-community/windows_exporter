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

// Command generator writes dashboard/windows-exporter-dashboard.json.
//
// The dashboard is defined as a v1 (classic) Grafana dashboard in dashboard.go. A running Grafana 13
// converts it to the v2 schema, and the generator groups the rows into tabs. See dashboard/README.md.
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"time"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func run() error {
	grafana := grafanaClient{
		password: os.Getenv("GRAFANA_PASSWORD"),
		token:    os.Getenv("GRAFANA_TOKEN"),
		client:   &http.Client{Timeout: time.Minute},
	}

	output := flag.String("output", "dashboard/windows-exporter-dashboard.json", "Path of the v2 dashboard. Empty skips the conversion.")
	v1Output := flag.String("v1-output", "", "Also write the v1 dashboard before the conversion to this path.")

	flag.StringVar(&grafana.url, "grafana-url", envOr("GRAFANA_URL", "http://localhost:3000"), "Grafana 13 that converts the dashboard. Env: GRAFANA_URL.")
	flag.StringVar(&grafana.user, "grafana-user", envOr("GRAFANA_USER", "admin"), "Grafana user for basic auth. The password is read from GRAFANA_PASSWORD. Env: GRAFANA_USER.")
	flag.StringVar(&grafana.namespace, "grafana-namespace", "default", "Grafana API namespace of the organization.")

	check := flag.Bool("check", false, "Run every panel query of the generated dashboard against Prometheus.")
	promURL := flag.String("prometheus-url", envOr("PROMETHEUS_URL", "http://localhost:9090"), "Prometheus for -check. Env: PROMETHEUS_URL.")
	instance := flag.String("check-instance", "127.0.0.1:9182", "Value of the instance variable for -check.")

	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage: go run ./dashboard/generator [flags]\n\n"+
			"Authenticates with GRAFANA_TOKEN (service account token) if set, otherwise with -grafana-user and GRAFANA_PASSWORD.\n\n")
		flag.PrintDefaults()
	}

	flag.Parse()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	v1 := buildDashboard()

	if *v1Output != "" {
		if err := writeJSON(*v1Output, v1); err != nil {
			return err
		}
	}

	if *output == "" {
		return nil
	}

	if grafana.token == "" && grafana.password == "" {
		return errors.New("set GRAFANA_TOKEN or GRAFANA_PASSWORD to convert the dashboard")
	}

	v2, err := grafana.convertToV2(ctx, v1)
	if err != nil {
		return err
	}

	if err = writeJSON(*output, v2); err != nil {
		return err
	}

	fmt.Fprintln(os.Stderr, "wrote", *output)

	if *check {
		return checkQueries(ctx, grafana.client, *promURL, *instance, v2, os.Stdout)
	}

	return nil
}

func envOr(name, fallback string) string {
	if v, ok := os.LookupEnv(name); ok {
		return v
	}

	return fallback
}

// writeJSON writes v with four-space indentation and without HTML escaping, like the committed dashboard.
func writeJSON(path string, v any) error {
	var buf bytes.Buffer

	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "    ")

	if err := enc.Encode(v); err != nil {
		return err
	}

	return os.WriteFile(path, buf.Bytes(), 0o600)
}
