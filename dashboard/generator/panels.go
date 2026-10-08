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
	"fmt"
	"strconv"
)

// M is a JSON object.
type M = map[string]any

func datasource() M { return M{"type": "prometheus", "uid": "${datasource}"} }

const (
	// H selects the host chosen in the Instance variable.
	H = `job=~"$job", instance="$instance"`
	// F selects all hosts of the chosen jobs. Fleet queries add fleetJoin or fleetFilter for the Hostname variable.
	F = `job=~"$job"`
)

// fleetJoin adds the hostname label and restricts to the selected hostnames.
func fleetJoin(expr string) string {
	return fmt.Sprintf(`(%s) * on (instance) group_left (hostname) max by (instance, hostname) (windows_os_hostname{%s, hostname=~"$hostname"})`, expr, F)
}

// fleetFilter restricts to the selected hostnames without adding labels (for tables).
func fleetFilter(expr string) string {
	return fmt.Sprintf(`(%s) and on (instance) windows_os_hostname{%s, hostname=~"$hostname"}`, expr, F)
}

func grid(x, y, w, h int) M { return M{"x": x, "y": y, "w": w, "h": h} }

// T is a Prometheus query of a panel. An empty Legend uses Grafana's automatic legend.
type T struct {
	Expr, Legend string
	Instant      bool // instant query instead of a range query
	Table        bool // table format, needed to show labels as values
}

func targets(ts []T) []any {
	out := make([]any, 0, len(ts))

	for i, t := range ts {
		m := M{
			"datasource":   datasource(),
			"editorMode":   "code",
			"expr":         t.Expr,
			"legendFormat": t.Legend,
			"range":        !t.Instant,
			"instant":      t.Instant,
			"refId":        string(rune('A' + i)),
		}
		if t.Legend == "" {
			m["legendFormat"] = "__auto"
		}

		if t.Table {
			m["format"] = "table"
		}

		out = append(out, m)
	}

	return out
}

// steps builds absolute thresholds from color, value pairs. The first value is nil (base).
func steps(s ...any) M {
	st := make([]any, 0, len(s)/2)
	for i := 0; i+1 < len(s); i += 2 {
		st = append(st, M{"color": s[i], "value": s[i+1]})
	}

	return M{"mode": "absolute", "steps": st}
}

func pctSteps() M   { return steps("green", nil, "orange", 80, "red", 90) }
func plainSteps() M { return steps("green", nil) }
func textSteps() M  { return steps("text", nil) }

func row(title string, y int) M {
	return M{
		"type":      "row",
		"title":     title,
		"gridPos":   grid(0, y, 24, 1),
		"collapsed": false,
		"panels":    []any{},
	}
}

// TS configures a time series panel. Zero values use the defaults of timeseries.
type TS struct {
	Title, Desc, Unit string
	Targets           []T
	Stack             bool
	Min, Max          any
	Decimals          any
	Legend            string // "table" (default), "list", "hidden"
	Calcs             []string
	Overrides         []any
	Fill              int
	Steps             M
	Thresholds        bool
	SortBy            string // legend column to sort by, descending
}

func timeseries(g M, o TS) M {
	if o.Legend == "" {
		o.Legend = "table"
	}

	if o.Calcs == nil {
		o.Calcs = []string{"mean", "max", "lastNotNull"}
	}

	if o.Fill == 0 {
		o.Fill = 10
	}

	stack := M{"group": "A", "mode": "none"}
	if o.Stack {
		stack["mode"] = "normal"
	}

	custom := M{
		"axisBorderShow":    false,
		"axisCenteredZero":  false,
		"axisColorMode":     "text",
		"axisLabel":         "",
		"axisPlacement":     "auto",
		"barAlignment":      0,
		"drawStyle":         "line",
		"fillOpacity":       o.Fill,
		"gradientMode":      "opacity",
		"hideFrom":          M{"legend": false, "tooltip": false, "viz": false},
		"insertNulls":       false,
		"lineInterpolation": "linear",
		"lineWidth":         1,
		"pointSize":         4,
		"scaleDistribution": M{"type": "linear"},
		"showPoints":        "never",
		"spanNulls":         false,
		"stacking":          stack,
		"thresholdsStyle":   M{"mode": "off"},
	}

	th := plainSteps()
	if o.Steps != nil {
		th = o.Steps
	}

	if o.Thresholds {
		custom["thresholdsStyle"] = M{"mode": "dashed"}
	}

	defaults := M{
		"color":      M{"mode": "palette-classic"},
		"custom":     custom,
		"mappings":   []any{},
		"thresholds": th,
		"unit":       o.Unit,
	}
	if o.Min != nil {
		defaults["min"] = o.Min
	}

	if o.Max != nil {
		defaults["max"] = o.Max
	}

	if o.Decimals != nil {
		defaults["decimals"] = o.Decimals
	}

	legend := M{"calcs": o.Calcs, "displayMode": "table", "placement": "bottom", "showLegend": true}
	if o.SortBy != "" {
		legend["sortBy"] = o.SortBy
		legend["sortDesc"] = true
	}

	switch o.Legend {
	case "list":
		legend = M{"calcs": []string{}, "displayMode": "list", "placement": "bottom", "showLegend": true}
	case "hidden":
		legend = M{"calcs": []string{}, "displayMode": "list", "placement": "bottom", "showLegend": false}
	}

	if o.Overrides == nil {
		o.Overrides = []any{}
	}

	return M{
		"type":        "timeseries",
		"title":       o.Title,
		"description": o.Desc,
		"datasource":  datasource(),
		"gridPos":     g,
		"fieldConfig": M{"defaults": defaults, "overrides": o.Overrides},
		"options": M{
			"legend":  legend,
			"tooltip": M{"hideZeros": false, "mode": "multi", "sort": "desc"},
		},
		"targets": targets(o.Targets),
	}
}

// negativeY mirrors series whose display name matches the regex below the X axis.
func negativeY(regex string) M { return byRegexp(regex, prop("custom.transform", "negative-Y")) }

func byName(name string, props ...M) M { return override("byName", name, props) }

func byRegexp(re string, props ...M) M { return override("byRegexp", re, props) }

func override(matcher, options string, props []M) M {
	ps := make([]any, 0, len(props))
	for _, p := range props {
		ps = append(ps, p)
	}

	return M{"matcher": M{"id": matcher, "options": options}, "properties": ps}
}

func prop(id string, v any) M { return M{"id": id, "value": v} }

func fixedColor(c string) M { return prop("color", M{"fixedColor": c, "mode": "fixed"}) }

// ST configures a stat panel. Zero values use the defaults of stat.
type ST struct {
	Title, Desc, Unit string
	Expr              string
	Steps             M
	Graph             bool
	Color             string // value, background, none
	Decimals          any
	Min, Max          any
	Text              string // reduce field: show a label instead of the value
	Mappings          []any
	Legend            string
}

func stat(g M, o ST) M {
	if o.Steps == nil {
		o.Steps = textSteps()
	}

	if o.Color == "" {
		o.Color = "value"
	}

	graph := "none"
	if o.Graph {
		graph = "area"
	}

	defaults := M{
		"color":      M{"mode": "thresholds"},
		"mappings":   []any{},
		"thresholds": o.Steps,
		"unit":       o.Unit,
	}
	if o.Mappings != nil {
		defaults["mappings"] = o.Mappings
	}

	if o.Decimals != nil {
		defaults["decimals"] = o.Decimals
	}

	if o.Min != nil {
		defaults["min"] = o.Min
	}

	if o.Max != nil {
		defaults["max"] = o.Max
	}

	reduce := M{"calcs": []string{"lastNotNull"}, "fields": "", "values": false}
	textMode := "auto"

	if o.Text != "" {
		reduce["fields"] = "/^" + o.Text + "$/"
		textMode = "value"
	}
	// A label shown instead of the value needs the query result as a table.
	t := T{Expr: o.Expr, Legend: o.Legend, Instant: !o.Graph, Table: o.Text != ""}

	return M{
		"type":        "stat",
		"title":       o.Title,
		"description": o.Desc,
		"datasource":  datasource(),
		"gridPos":     g,
		"fieldConfig": M{"defaults": defaults, "overrides": []any{}},
		"options": M{
			"colorMode":              o.Color,
			"graphMode":              graph,
			"justifyMode":            "auto",
			"orientation":            "auto",
			"percentChangeColorMode": "standard",
			"reduceOptions":          reduce,
			"showPercentChange":      false,
			"textMode":               textMode,
			"wideLayout":             true,
		},
		"targets": targets([]T{t}),
	}
}

// TB configures a table panel. Every query is an instant table query. Columns are named
// "Value #A", "Value #B" and so on by query; Rename and Order refer to those names.
type TB struct {
	Title, Desc string
	Targets     []T
	JoinBy      string
	Exclude     []string
	Rename      M
	Order       M
	Overrides   []any
	SortBy      []any
	Transform   []any
	Footer      []string // fields summed in the table footer
}

func table(g M, o TB) M {
	ex := M{"Time": true}
	for _, e := range o.Exclude {
		ex[e] = true
	}

	for i := range o.Targets {
		o.Targets[i].Instant = true
		o.Targets[i].Table = true
		ex["Time "+strconv.Itoa(i+1)] = true
	}

	if o.Rename == nil {
		o.Rename = M{}
	}

	if o.Order == nil {
		o.Order = M{}
	}

	if o.SortBy == nil {
		o.SortBy = []any{}
	}

	tr := []any{}
	if o.JoinBy != "" {
		tr = append(tr, M{"id": "joinByField", "options": M{"byField": o.JoinBy, "mode": "outer"}})
	}

	tr = append(tr, o.Transform...)

	tr = append(tr, M{"id": "organize", "options": M{
		"excludeByName": ex, "includeByName": M{}, "indexByName": o.Order, "renameByName": o.Rename,
	}})
	if o.Overrides == nil {
		o.Overrides = []any{}
	}

	footer := M{"countRows": false, "fields": "", "reducer": []string{"sum"}, "show": false}
	if o.Footer != nil {
		footer["fields"] = o.Footer
		footer["show"] = true
	}

	return M{
		"type":        "table",
		"title":       o.Title,
		"description": o.Desc,
		"datasource":  datasource(),
		"gridPos":     g,
		"fieldConfig": M{
			"defaults": M{
				"color":      M{"mode": "thresholds"},
				"custom":     M{"align": "auto", "cellOptions": M{"type": "auto"}, "filterable": false, "inspect": false, "minWidth": 60},
				"mappings":   []any{},
				"thresholds": plainSteps(),
			},
			"overrides": o.Overrides,
		},
		"options": M{
			"cellHeight":       "sm",
			"footer":           footer,
			"showHeader":       true,
			"sortBy":           o.SortBy,
			"enablePagination": false,
		},
		"targets":         targets(o.Targets),
		"transformations": tr,
	}
}

func gaugeCell(unit string, maxValue any) []M {
	return []M{
		prop("unit", unit),
		prop("min", 0),
		prop("max", maxValue),
		prop("thresholds", pctSteps()),
		prop("color", M{"mode": "continuous-GrYlRd"}),
		prop("custom.cellOptions", M{"type": "gauge", "mode": "basic", "valueDisplayMode": "text"}),
		prop("decimals", 1),
	}
}

func unitCell(unit string) []M { return []M{prop("unit", unit)} }

func variable(name, label, query string, multi, all, hide bool, regex string) M {
	h := 0
	if hide {
		h = 2
	}

	v := M{
		"type":       "query",
		"name":       name,
		"label":      label,
		"datasource": datasource(),
		"definition": query,
		"query":      M{"qryType": 1, "query": query, "refId": "PrometheusVariableQueryEditor-VariableQuery"},
		"refresh":    2,
		"regex":      regex,
		"sort":       1,
		"multi":      multi,
		"includeAll": all,
		"hide":       h,
		"current":    M{},
		"options":    []any{},
	}
	if all {
		v["current"] = M{"text": []string{"All"}, "value": []string{"$__all"}}
	}

	return v
}

// stateTimeline draws one row per series, colored by value. It is used for per-core CPU load,
// where a line per core becomes unreadable.
func stateTimeline(g M, title, desc, unit string, maxValue any, t T) M {
	return stateTimelinePanel(g, title, desc, t, M{
		"color":      M{"mode": "continuous-GrYlRd"},
		"custom":     stateTimelineCustom(),
		"mappings":   []any{},
		"min":        0,
		"max":        maxValue,
		"thresholds": plainSteps(),
		"unit":       unit,
	})
}

// statusTimeline draws a 0/1 metric per series as a red/green timeline.
func statusTimeline(g M, title, desc string, t T, okText, failText string) M {
	return stateTimelinePanel(g, title, desc, t, M{
		"color":  M{"mode": "thresholds"},
		"custom": stateTimelineCustom(),
		"mappings": []any{M{"type": "value", "options": M{
			"0": M{"color": "red", "index": 0, "text": failText},
			"1": M{"color": "green", "index": 1, "text": okText},
		}}},
		"min":        0,
		"max":        1,
		"thresholds": steps("red", nil, "green", 1),
		"unit":       "none",
	})
}

func stateTimelineCustom() M {
	return M{
		"fillOpacity": 90, "hideFrom": M{"legend": false, "tooltip": false, "viz": false},
		"insertNulls": false, "lineWidth": 0, "spanNulls": false,
	}
}

func stateTimelinePanel(g M, title, desc string, t T, defaults M) M {
	return M{
		"type":        "state-timeline",
		"title":       title,
		"description": desc,
		"datasource":  datasource(),
		"gridPos":     g,
		"fieldConfig": M{"defaults": defaults, "overrides": []any{}},
		"options": M{
			"alignValue":  "left",
			"legend":      M{"displayMode": "list", "placement": "bottom", "showLegend": false},
			"mergeValues": false,
			"rowHeight":   1,
			"showValue":   "never",
			"tooltip":     M{"hideZeros": false, "mode": "single", "sort": "none"},
		},
		"targets": targets([]T{t}),
	}
}
