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

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"maps"
	"net/http"
	"strings"
)

// buildUID is the UID of the temporary dashboard that Grafana converts.
const buildUID = "windows-exporter-build"

type grafanaClient struct {
	url       string
	namespace string
	user      string
	password  string
	token     string
	client    *http.Client
}

func (g grafanaClient) call(ctx context.Context, method, path string, body []byte) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, method, strings.TrimSuffix(g.url, "/")+path, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	if g.token != "" {
		req.Header.Set("Authorization", "Bearer "+g.token)
	} else {
		req.SetBasicAuth(g.user, g.password)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := g.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("%s %s: %s: %s", method, path, resp.Status, respBody)
	}

	return respBody, nil
}

// convertToV2 lets Grafana convert the v1 dashboard to the v2 schema and groups the rows into tabs.
// The dashboard is stored in Grafana under buildUID during the conversion and deleted afterward.
func (g grafanaClient) convertToV2(ctx context.Context, v1 M) (M, error) {
	uid, _ := v1["uid"].(string)
	title := v1["title"]

	build := make(M, len(v1))
	maps.Copy(build, v1)

	build["uid"] = buildUID
	build["title"] = buildUID

	payload, err := json.Marshal(M{"overwrite": true, "dashboard": build})
	if err != nil {
		return nil, err
	}

	if _, err = g.call(ctx, http.MethodPost, "/api/dashboards/db", payload); err != nil {
		return nil, fmt.Errorf("upload dashboard: %w", err)
	}

	raw, err := g.call(ctx, http.MethodGet, "/apis/dashboard.grafana.app/v2/namespaces/"+g.namespace+"/dashboards/"+buildUID, nil)

	if _, delErr := g.call(ctx, http.MethodDelete, "/api/dashboards/uid/"+buildUID, nil); delErr != nil {
		err = errors.Join(err, fmt.Errorf("delete temporary dashboard: %w", delErr))
	}

	if err != nil {
		return nil, err
	}

	var v2 M
	if err = json.Unmarshal(raw, &v2); err != nil {
		return nil, err
	}

	if status, ok := v2["status"].(M); ok {
		if conv, ok := status["conversion"].(M); ok && conv["failed"] == true {
			return nil, fmt.Errorf("conversion failed: %v", conv)
		}
	}

	spec, ok := v2["spec"].(M)
	if !ok {
		return nil, errors.New("v2 dashboard has no spec")
	}

	spec["title"] = title

	if err = regroupTabs(spec); err != nil {
		return nil, err
	}

	return M{
		"apiVersion": v2["apiVersion"],
		"kind":       "Dashboard",
		"metadata":   M{"name": uid},
		"spec":       spec,
	}, nil
}

// regroupTabs turns the rows of the converted dashboard into tabs. A row titled "Tab" starts a tab
// without a row title, "Tab / Row" adds a row to the tab. Tabs keep the order of their first row.
// A tab with a single untitled row gets the layout of that row; other tabs get a RowsLayout.
func regroupTabs(spec M) error {
	layout, _ := spec["layout"].(M)
	layoutSpec, _ := layout["spec"].(M)

	rows, ok := layoutSpec["rows"].([]any)
	if !ok {
		return errors.New("converted dashboard has no RowsLayout")
	}

	type tabRow struct {
		spec M
		row  any
	}

	var order []string

	rowsByTab := map[string][]tabRow{}

	for _, r := range rows {
		row, _ := r.(M)

		rowSpec, ok := row["spec"].(M)
		if !ok {
			return fmt.Errorf("row without spec: %v", r)
		}

		title, _ := rowSpec["title"].(string)
		tab, sub, _ := strings.Cut(title, " / ")

		if _, seen := rowsByTab[tab]; !seen {
			order = append(order, tab)
		}

		rowSpec["title"] = sub
		rowsByTab[tab] = append(rowsByTab[tab], tabRow{spec: rowSpec, row: row})
	}

	tabs := make([]any, 0, len(order))

	for _, tab := range order {
		tabRows := rowsByTab[tab]

		var tabLayout any

		if len(tabRows) == 1 && tabRows[0].spec["title"] == "" {
			tabLayout = tabRows[0].spec["layout"]
		} else {
			rowList := make([]any, 0, len(tabRows))

			for _, r := range tabRows {
				r.spec["collapse"] = false
				rowList = append(rowList, r.row)
			}

			tabLayout = M{"kind": "RowsLayout", "spec": M{"rows": rowList}}
		}

		tabs = append(tabs, M{"kind": "TabsLayoutTab", "spec": M{"title": tab, "layout": tabLayout}})
	}

	spec["layout"] = M{"kind": "TabsLayout", "spec": M{"tabs": tabs}}

	return nil
}
