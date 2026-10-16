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

package metriccontract_test

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

const (
	// upgradeGuideFile is the upgrade guide of 0.32. Its rename and instance
	// label tables are the source of the compatibility rules.
	upgradeGuideFile = "docs/upgrade/0.32.md"
	// compatibilityRulesFile holds recording rules that publish the metrics
	// of 0.32 under their 0.31 names and labels. Set
	// UPDATE_COMPATIBILITY_RULES=1 to regenerate it.
	compatibilityRulesFile = "docs/upgrade/0.32.prom"
)

// renameRow matches a row of a rename table in the upgrade guide: the 0.31
// name, the 0.32 name and a note.
var renameRow = regexp.MustCompile("^\\|\\s*`(windows_[A-Za-z0-9_]+)`\\s*\\|\\s*`(windows_[A-Za-z0-9_]+)`\\s*\\|\\s*(.*?)\\s*\\|\\s*$")

// labelRow matches a row of the instance label table in the upgrade guide.
var labelRow = regexp.MustCompile("^\\|\\s*`([a-z_]+)`\\s*\\|([^|]*)\\|([^|]*)\\|([^|]*)\\|\\s*$")

var backticked = regexp.MustCompile("`([^`]+)`")

// conversion turns the value of a 0.32 metric back into its 0.31 value.
type conversion struct {
	// match is a lowercase part of the note of a rename.
	match string
	// expr formats the 0.31 value of the metric %s.
	expr string
}

// conversions are tried in order. The first match wins.
//
//nolint:gochecknoglobals
var conversions = []conversion{
	{"seconds instead of milliseconds", "%s * 1000"},
	{"now in seconds, was milliseconds", "%s * 1000"},
	{"seconds instead of microseconds", "%s * 1e6"},
	{"seconds instead of minutes", "%s / 60"},
	{"seconds instead of hours", "%s / 3600"},
	{"bytes instead of kilobytes", "%s / 1024"},
	{"was kilobytes, now bytes", "%s / 1024"},
	{"bytes instead of megabytes", "%s / 1048576"},
	{"hertz instead of megahertz", "%s / 1e6"},
	{"instead of percent", "%s * 100"},
	{"instead of parts per billion", "%s * 1e9"},
	{"instead of the per-minute value", "%s * 3600"},
	{"same value", "%s"},
	{"unit suffix added", "%s"},
	{"now a counter", "%s"},
	{"now a gauge", "%s"},
	{"was a running count", "%s"},
}

// unconvertible are lowercase parts of notes whose 0.31 value can't be
// computed from the 0.32 metric. They are checked before conversions.
//
//nolint:gochecknoglobals
var unconvertible = []string{
	"removed",
	"see [",
	"see below",
	"0.31 divided",
	"raw millisecond",
	"divide its `rate()` by that",
	"label `",
}

type labelRename struct {
	metrics *regexp.Regexp
	// prefix is the longest literal prefix of the metric names, which
	// decides between overlapping rows like windows_mscluster_resource_* and
	// windows_mscluster_resourcegroup_*.
	prefix string
	old    []string
	new    []string
}

// labelRenameOf returns the most specific label rename of the metric name.
func labelRenameOf(renames []labelRename, name string) (labelRename, bool) {
	var (
		best  labelRename
		found bool
	)

	for _, rename := range renames {
		if rename.metrics.MatchString(name) && (!found || len(rename.prefix) > len(best.prefix)) {
			best, found = rename, true
		}
	}

	return best, found
}

type compatibilityRule struct {
	record string
	expr   string
}

// TestCompatibilityRules checks that the compatibility recording rules are
// generated from the current upgrade guide, read only metrics the exporter
// emits and record no name it emits itself.
func TestCompatibilityRules(t *testing.T) {
	t.Parallel()

	upgrade, err := os.ReadFile(filepath.Join(repoRoot, filepath.FromSlash(upgradeGuideFile)))
	require.NoError(t, err)

	rules := compatibilityRules(t, strings.Split(strings.ReplaceAll(string(upgrade), "\r\n", "\n"), "\n"))
	require.NotEmpty(t, rules)

	generated := formatCompatibilityRules(rules)
	path := filepath.Join(repoRoot, filepath.FromSlash(compatibilityRulesFile))

	if os.Getenv("UPDATE_COMPATIBILITY_RULES") != "" {
		require.NoError(t, os.WriteFile(path, []byte(generated), 0o644)) //nolint:gosec // a documentation file
	}

	committed, err := os.ReadFile(path)
	require.NoError(t, err)
	require.Equal(t, generated, strings.ReplaceAll(string(committed), "\r\n", "\n"),
		"%s is out of date with %s. Run: UPDATE_COMPATIBILITY_RULES=1 go test ./internal/metriccontract/", compatibilityRulesFile, upgradeGuideFile)

	emitted := emittedMetricPatterns(t)
	matcher := regexp.MustCompile("^(?:" + strings.Join(emitted, "|") + ")$")

	for _, rules := range rules {
		for _, rule := range rules {
			source := referencePattern.FindStringSubmatch(rule.expr)
			require.NotNil(t, source, rule.expr)
			require.True(t, matcher.MatchString(source[1]), "%s reads %s, which the exporter does not emit", rule.record, source[1])
			require.False(t, matcher.MatchString(rule.record), "%s is emitted by the exporter, a recording rule would duplicate it", rule.record)
		}
	}
}

// compatibilityRules returns the rules by collector.
func compatibilityRules(t *testing.T, upgrade []string) map[string][]compatibilityRule {
	t.Helper()

	labels := labelRenames(t, upgrade)
	rules := map[string][]compatibilityRule{}
	seen := map[string]bool{}

	for _, line := range upgrade {
		m := renameRow.FindStringSubmatch(line)
		if m == nil {
			continue
		}

		oldName, newName, note := m[1], m[2], strings.ToLower(m[3])

		if seen[oldName] || slices.ContainsFunc(unconvertible, func(s string) bool { return strings.Contains(note, s) }) {
			continue
		}

		i := slices.IndexFunc(conversions, func(c conversion) bool { return strings.Contains(note, c.match) })
		if i < 0 {
			t.Logf("%s: no conversion for the note %q, left out", oldName, m[3])

			continue
		}

		seen[oldName] = true

		expr := fmt.Sprintf(conversions[i].expr, newName)

		var dropped []string

		if rename, ok := labelRenameOf(labels, newName); ok {
			for j, oldLabel := range rename.old {
				expr = fmt.Sprintf(`label_replace(%s, "%s", "$1", "%s", "(.+)")`, expr, oldLabel, rename.new[j])
				dropped = append(dropped, rename.new[j])
			}
		}

		if len(dropped) > 0 {
			expr = fmt.Sprintf("max without (%s) (%s)", strings.Join(dropped, ", "), expr)
		}

		collector := collectorOf(t, newName)
		rules[collector] = append(rules[collector], compatibilityRule{record: oldName, expr: expr})
	}

	return rules
}

// labelRenames returns the instance label renames of the upgrade guide. A label
// named instance in 0.31 is restored as exported_instance, the name under
// which Prometheus stored it next to the target label.
func labelRenames(t *testing.T, upgrade []string) []labelRename {
	t.Helper()

	var renames []labelRename

	for _, line := range upgrade {
		m := labelRow.FindStringSubmatch(line)
		if m == nil {
			continue
		}

		var (
			patterns []string
			prefix   string
		)

		for _, metric := range backticked.FindAllStringSubmatch(m[2], -1) {
			if strings.HasPrefix(metric[1], "windows_") {
				patterns = append(patterns, strings.ReplaceAll(regexp.QuoteMeta(metric[1]), `\*`, "[a-zA-Z0-9_]*"))
				prefix, _, _ = strings.Cut(metric[1], "*")
			}
		}

		if len(patterns) == 0 {
			continue
		}

		var oldLabels, newLabels []string

		for _, label := range backticked.FindAllStringSubmatch(m[3], -1) {
			name := label[1]
			if name == "instance" {
				name = "exported_instance"
			}

			oldLabels = append(oldLabels, name)
		}

		for _, label := range backticked.FindAllStringSubmatch(m[4], -1) {
			newLabels = append(newLabels, label[1])
		}

		require.GreaterOrEqual(t, len(newLabels), len(oldLabels), line)

		renames = append(renames, labelRename{
			metrics: regexp.MustCompile("^(?:" + strings.Join(patterns, "|") + ")$"),
			prefix:  prefix,
			old:     oldLabels,
			new:     newLabels[:len(oldLabels)],
		})
	}

	require.NotEmpty(t, renames, "the instance label table of %s was not found", upgradeGuideFile)

	return renames
}

func formatCompatibilityRules(rules map[string][]compatibilityRule) string {
	var b strings.Builder

	fmt.Fprintf(&b, `# Recording rules that publish the metrics of windows_exporter 0.32 under
# their 0.31 names, units and instance labels, for dashboards and alerts that
# have not been migrated yet. Load them into Prometheus next to the scrape of
# the new exporter and remove them once the queries use the new names.
#
# Generated from %s by internal/metriccontract; do not edit.
#
# Not covered:
# - Metrics that only renamed their instance label, e.g. windows_service_*.
#   Recording them under the same name would duplicate every series.
# - Metrics whose 0.31 value can't be computed from 0.32, e.g. removed
#   metrics, split metrics and ratios that 0.31 computed wrongly.
# - Series that 0.32 leaves out, e.g. mscluster settings that are zero.
groups:
`, upgradeGuideFile)

	collectors := make([]string, 0, len(rules))
	for collector := range rules {
		collectors = append(collectors, collector)
	}

	slices.Sort(collectors)

	for _, collector := range collectors {
		fmt.Fprintf(&b, "  - name: windows_exporter-0.32-%s\n    rules:\n", collector)

		for _, rule := range rules[collector] {
			fmt.Fprintf(&b, "      - record: %s\n        expr: %s\n", rule.record, rule.expr)
		}
	}

	return b.String()
}

// collectorOf returns the collector that emits the metric name: the longest
// directory name in internal/collector that prefixes the name, e.g. container
// for windows_containers.
func collectorOf(t *testing.T, name string) string {
	t.Helper()

	entries, err := os.ReadDir(filepath.Join(repoRoot, "internal", "collector"))
	require.NoError(t, err)

	var collector string

	for _, entry := range entries {
		if entry.IsDir() && len(entry.Name()) > len(collector) && strings.HasPrefix(name, "windows_"+entry.Name()) {
			collector = entry.Name()
		}
	}

	require.NotEmpty(t, collector, "no collector for %s", name)

	return collector
}
