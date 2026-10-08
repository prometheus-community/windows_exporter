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
	"cmp"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"slices"
	"strings"
)

type checkResult struct {
	Status string `json:"status"`
	Error  string `json:"error"`
	Data   struct {
		Result []json.RawMessage `json:"result"`
	} `json:"data"`
}

// checkQueries runs every panel query of the v2 dashboard as an instant query against Prometheus and
// prints the number of series per query. Dashboard variables are replaced by values that match the
// given instance. Queries of optional collectors return no data unless those collectors are enabled.
// It returns an error if Prometheus rejects any query.
func checkQueries(ctx context.Context, client *http.Client, promURL, instance string, v2 M, w io.Writer) error {
	replacer := strings.NewReplacer(
		`"$job"`, `".*"`,
		`"$hostname"`, `".*"`,
		`"$instance"`, `"`+instance+`"`,
		`"$volume"`, `".*"`,
		`"$nic"`, `".*"`,
		`$__rate_interval`, `1m`,
		`$__range`, `1h`,
	)

	spec, _ := v2["spec"].(M)
	elements, _ := spec["elements"].(M)

	names := make([]string, 0, len(elements))
	for name := range elements {
		names = append(names, name)
	}

	// panel-2 before panel-10
	slices.SortFunc(names, func(a, b string) int {
		return cmp.Or(cmp.Compare(len(a), len(b)), cmp.Compare(a, b))
	})

	var failed, empty, total int

	for _, name := range names {
		element, _ := elements[name].(M)
		elementSpec, _ := element["spec"].(M)
		data, _ := elementSpec["data"].(M)
		dataSpec, _ := data["spec"].(M)
		queries, _ := dataSpec["queries"].([]any)

		for _, q := range queries {
			query, _ := q.(M)
			querySpec, _ := query["spec"].(M)
			inner, _ := querySpec["query"].(M)
			innerSpec, _ := inner["spec"].(M)
			expr, _ := innerSpec["expr"].(string)

			status, err := runQuery(ctx, client, promURL, replacer.Replace(expr))
			if err != nil {
				return err
			}

			total++

			switch {
			case status.Status != "success":
				failed++

				fmt.Fprintf(w, "%-50.50s %s: ERROR %s\n", elementSpec["title"], querySpec["refId"], status.Error)
			case len(status.Data.Result) == 0:
				empty++

				fmt.Fprintf(w, "%-50.50s %s: EMPTY\n", elementSpec["title"], querySpec["refId"])
			default:
				fmt.Fprintf(w, "%-50.50s %s: %d series\n", elementSpec["title"], querySpec["refId"], len(status.Data.Result))
			}
		}
	}

	fmt.Fprintf(w, "%d queries, %d without data, %d failed\n", total, empty, failed)

	if failed > 0 {
		return fmt.Errorf("%d queries failed", failed)
	}

	return nil
}

func runQuery(ctx context.Context, client *http.Client, promURL, expr string) (checkResult, error) {
	var result checkResult

	form := url.Values{"query": {expr}}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, strings.TrimSuffix(promURL, "/")+"/api/v1/query", strings.NewReader(form.Encode()))
	if err != nil {
		return result, err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := client.Do(req)
	if err != nil {
		return result, err
	}
	defer resp.Body.Close()

	// Prometheus answers rejected queries with status 400 and an error in the body.
	if err = json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return result, fmt.Errorf("query %q: %s: %w", expr, resp.Status, err)
	}

	return result, nil
}
