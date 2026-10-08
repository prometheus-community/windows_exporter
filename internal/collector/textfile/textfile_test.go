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

package textfile

import (
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"

	dto "github.com/prometheus/client_model/go"
)

func TestCRFilter(t *testing.T) {
	t.Parallel()

	sr := strings.NewReader("line 1\r\nline 2")
	cr := carriageReturnFilteringReader{r: sr}

	b, err := io.ReadAll(cr)
	if err != nil {
		t.Error(err)
	}

	if string(b) != "line 1\nline 2" {
		t.Errorf("Unexpected output %q", b)
	}
}

func TestScrapeFileBOM(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name   string
		prefix string
		err    string
	}{
		{name: "plain UTF-8"},
		{name: "UTF-8 BOM", prefix: "\xef\xbb\xbf"},
		{name: "UTF-16 BE", prefix: "\xfe\xff", err: "UTF16BigEndian"},
		{name: "UTF-16 LE", prefix: "\xff\xfe", err: "UTF16LittleEndian"},
		{name: "UTF-32 BE", prefix: "\x00\x00\xfe\xff", err: "UTF32BigEndian"},
		{name: "UTF-32 LE", prefix: "\xff\xfe\x00\x00", err: "UTF32LittleEndian"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			path := filepath.Join(t.TempDir(), "metric.prom")
			content := tc.prefix + "# TYPE metric gauge\r\nmetric 1\r\n"

			if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
				t.Fatal(err)
			}

			families, err := scrapeFile(path, slog.New(slog.DiscardHandler))
			if tc.err != "" {
				if err == nil || err.Error() != tc.err {
					t.Fatalf("got error %v, want %s", err, tc.err)
				}

				return
			}

			if err != nil {
				t.Fatal(err)
			}

			if len(families) != 1 || families[0].GetName() != "metric" {
				t.Fatalf("unexpected metric families: %v", families)
			}

			metrics := families[0].GetMetric()
			if len(metrics) != 1 || metrics[0].GetGauge().GetValue() != 1 {
				t.Errorf("unexpected metrics: %v", metrics)
			}
		})
	}
}

func TestDuplicateMetricEntry(t *testing.T) {
	t.Parallel()

	metric_name := "windows_sometest"
	metric_help := "This is a Test."
	metric_type := dto.MetricType_GAUGE

	gauge_value := 1.0

	gauge := dto.Gauge{
		Value: &gauge_value,
	}

	label1_name := "display_name"
	label1_value := "foobar"

	label1 := dto.LabelPair{
		Name:  &label1_name,
		Value: &label1_value,
	}

	label2_name := "display_version"
	label2_value := "13.4.0"

	label2 := dto.LabelPair{
		Name:  &label2_name,
		Value: &label2_value,
	}

	metric1 := dto.Metric{
		Label: []*dto.LabelPair{&label1, &label2},
		Gauge: &gauge,
	}

	metric2 := dto.Metric{
		Label: []*dto.LabelPair{&label1, &label2},
		Gauge: &gauge,
	}

	duplicate := dto.MetricFamily{
		Name:   &metric_name,
		Help:   &metric_help,
		Type:   &metric_type,
		Metric: []*dto.Metric{&metric1, &metric2},
	}

	duplicateFamily := make([]*dto.MetricFamily, 0, 1)

	duplicateFamily = append(duplicateFamily, &duplicate)

	// Ensure detection for duplicate metrics
	if !duplicateMetricEntry(duplicateFamily) {
		t.Errorf("Duplicate not found in duplicateFamily")
	}

	label3_name := "test"
	label3_value := "1.0"

	label3 := dto.LabelPair{
		Name:  &label3_name,
		Value: &label3_value,
	}
	metric3 := dto.Metric{
		Label: []*dto.LabelPair{&label1, &label2, &label3},
		Gauge: &gauge,
	}

	differentLabels := dto.MetricFamily{
		Name:   &metric_name,
		Help:   &metric_help,
		Type:   &metric_type,
		Metric: []*dto.Metric{&metric1, &metric3},
	}

	duplicateFamily = make([]*dto.MetricFamily, 0, 1)
	duplicateFamily = append(duplicateFamily, &differentLabels)

	// Additional label on second metric should not be cause for duplicate detection
	if duplicateMetricEntry(duplicateFamily) {
		t.Errorf("Unexpected duplicate found in differentLabels")
	}

	label4_value := "2.0"

	label4 := dto.LabelPair{
		Name:  &label3_name,
		Value: &label4_value,
	}
	metric4 := dto.Metric{
		Label: []*dto.LabelPair{&label1, &label2, &label4},
		Gauge: &gauge,
	}

	differentValues := dto.MetricFamily{
		Name:   &metric_name,
		Help:   &metric_help,
		Type:   &metric_type,
		Metric: []*dto.Metric{&metric3, &metric4},
	}
	duplicateFamily = make([]*dto.MetricFamily, 0, 1)
	duplicateFamily = append(duplicateFamily, &differentValues)

	// Additional label with different values metric should not be cause for duplicate detection
	if duplicateMetricEntry(duplicateFamily) {
		t.Errorf("Unexpected duplicate found in differentValues")
	}
}

func TestDuplicateMetricEntryAnyPosition(t *testing.T) {
	t.Parallel()

	type pair [2]string

	metric := func(labels ...pair) *dto.Metric {
		m := &dto.Metric{}

		for _, label := range labels {
			m.Label = append(m.Label, &dto.LabelPair{Name: &label[0], Value: &label[1]})
		}

		return m
	}

	family := func(name string, metrics ...*dto.Metric) *dto.MetricFamily {
		return &dto.MetricFamily{Name: &name, Metric: metrics}
	}

	for _, tc := range []struct {
		name      string
		families  []*dto.MetricFamily
		duplicate bool
	}{
		{
			name:      "duplicate is not adjacent",
			families:  []*dto.MetricFamily{family("m", metric(pair{"a", "1"}), metric(pair{"a", "2"}), metric(pair{"a", "1"}))},
			duplicate: true,
		},
		{
			name:      "duplicate in another family of the same name",
			families:  []*dto.MetricFamily{family("m", metric(pair{"a", "1"}), metric(pair{"a", "2"})), family("m", metric(pair{"a", "1"}))},
			duplicate: true,
		},
		{
			name:      "different label order",
			families:  []*dto.MetricFamily{family("m", metric(pair{"a", "1"}, pair{"b", "2"}), metric(pair{"b", "2"}, pair{"a", "1"}))},
			duplicate: true,
		},
		{
			name:      "empty label value is absent",
			families:  []*dto.MetricFamily{family("m", metric(pair{"a", "1"}), metric(pair{"a", "1"}, pair{"b", ""}))},
			duplicate: true,
		},
		{
			name:     "same labels in different metrics",
			families: []*dto.MetricFamily{family("m", metric(pair{"a", "1"})), family("n", metric(pair{"a", "1"}))},
		},
		{
			name:     "label value shifted between labels",
			families: []*dto.MetricFamily{family("m", metric(pair{"a", "1"}, pair{"b", "2"}), metric(pair{"a", "12"}, pair{"b", ""}))},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if got := duplicateMetricEntry(tc.families); got != tc.duplicate {
				t.Errorf("duplicateMetricEntry() = %t, want %t", got, tc.duplicate)
			}
		})
	}
}
