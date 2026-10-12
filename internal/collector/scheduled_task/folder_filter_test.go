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

package scheduled_task

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFolderFilter(t *testing.T) {
	t.Parallel()

	// The flags anchor the expressions like this.
	flag := func(expr string) *regexp.Regexp {
		return regexp.MustCompile(fmt.Sprintf("^(?:%s)$", expr))
	}

	for _, tc := range []struct {
		name     string
		include  *regexp.Regexp
		exclude  *regexp.Regexp
		included []string
		skipped  []string
	}{
		{
			name:     "defaults",
			include:  ConfigDefaults.TaskInclude,
			exclude:  ConfigDefaults.TaskExclude,
			included: []string{"/Microsoft", "/Microsoft/Windows", "/Other"},
		},
		{
			name:     "flag defaults",
			include:  flag(".+"),
			exclude:  flag(""),
			included: []string{"/Microsoft", "/Other"},
		},
		{
			name:     "nil filters",
			included: []string{"/Microsoft"},
		},
		{
			name:     "exclude subtree",
			include:  flag(".+"),
			exclude:  flag("/Microsoft/.+"),
			included: []string{"/Other", "/MicrosoftEdge", "/Microsofx"},
			skipped:  []string{"/Microsoft", "/Microsoft/Windows", "/Microsoft/Windows/Defrag"},
		},
		{
			name:     "exclude deeper subtree",
			include:  flag(".+"),
			exclude:  flag("/Microsoft/Windows/.*"),
			included: []string{"/Microsoft", "/Microsoft/Office", "/Microsoft/WindowsUpdate"},
			skipped:  []string{"/Microsoft/Windows", "/Microsoft/Windows/Defrag"},
		},
		{
			name:     "exclude alternatives",
			include:  flag(".+"),
			exclude:  flag("/Microsoft/.*|/Mozilla/.+"),
			included: []string{"/Other", "/Mozillas"},
			skipped:  []string{"/Microsoft/Windows", "/Mozilla", "/Mozilla/Firefox"},
		},
		{
			name:     "exclude capture group",
			exclude:  regexp.MustCompile(`^(/Microsoft/.*)$`),
			included: []string{"/Other"},
			skipped:  []string{"/Microsoft/Windows"},
		},
		{
			name:     "unanchored exclude",
			exclude:  regexp.MustCompile(`/Microsoft/.*`),
			included: []string{"/Other", "/Vendor"},
			skipped:  []string{"/Microsoft"},
		},
		{
			name:    "exclude every task",
			exclude: flag(".*"),
			skipped: []string{"/Microsoft", "/Other"},
		},
		{
			name: "exclude without a full subtree",
			exclude: flag(
				"/Microsoft/[^/]+|/Microsoft/Windows/.+/Task|/Microsoft/Windows|(?i)/microsoft/.+|/Other/.+x|/Other/\n.+|/Other/(?:^.+)",
			),
			included: []string{"/Microsoft", "/Microsoft/Windows", "/Other"},
		},
		{
			name:     "include subtree",
			include:  flag("/Vendor/App/.+"),
			included: []string{"/Vendor", "/Vendor/App", "/Vendor/App/Sub"},
			skipped:  []string{"/Microsoft", "/Vendor/Other", "/Vendo"},
		},
		{
			name:     "include alternatives",
			include:  flag("/Vendor/.+|/Other/Task"),
			included: []string{"/Vendor", "/Other"},
			skipped:  []string{"/Microsoft", "/Other/Sub"},
		},
		{
			name:     "include with folder prefix",
			include:  flag("/Vend.*"),
			included: []string{"/Vendor", "/Vendor/App", "/Vend"},
			skipped:  []string{"/Microsoft"},
		},
		{
			name:     "unanchored include",
			include:  regexp.MustCompile(`Firefox`),
			included: []string{"/Microsoft", "/Mozilla"},
		},
		{
			name:     "case-insensitive include",
			include:  flag("(?i)/vendor/.+"),
			included: []string{"/Vendor", "/Microsoft"},
		},
		{
			name:     "include and exclude",
			include:  flag("/Microsoft/.+"),
			exclude:  flag("/Microsoft/Windows/.+"),
			included: []string{"/Microsoft", "/Microsoft/Office"},
			skipped:  []string{"/Other", "/Microsoft/Windows", "/Microsoft/Windows/Defrag"},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			filter := newFolderFilter(tc.include, tc.exclude)

			for _, path := range tc.included {
				require.True(t, filter.includeFolder(path), path)
			}

			for _, path := range tc.skipped {
				require.False(t, filter.includeFolder(path), path)
			}
		})
	}
}

// TestFolderFilterSkipsOnlyFilteredTasks checks the folder decision against the
// task filters for task paths below each folder.
func TestFolderFilterSkipsOnlyFilteredTasks(t *testing.T) {
	t.Parallel()

	expressions := []string{
		"", ".+", ".*", "/Microsoft/.+", "/Microsoft/.*", "/Microsoft/Windows/.+", "/Microsoft/.+|/Other/.*",
		"/Microsoft/Windows", "/Microsoft/[^/]+", "/Micro.*", "(?i)/microsoft/.+", "/Other/Task",
	}
	folders := []string{"/Microsoft", "/Microsoft/Windows", "/Microsoft/Windows/Defrag", "/MicrosoftEdge", "/Other"}
	names := []string{"Task", "Sub/Task", "x"}

	for _, include := range expressions {
		for _, exclude := range expressions {
			c := New(&Config{
				TaskInclude: regexp.MustCompile(fmt.Sprintf("^(?:%s)$", include)),
				TaskExclude: regexp.MustCompile(fmt.Sprintf("^(?:%s)$", exclude)),
			})
			filter := newFolderFilter(c.config.TaskInclude, c.config.TaskExclude)

			for _, folder := range folders {
				if filter.includeFolder(folder) {
					continue
				}

				for _, name := range names {
					path := folder + "/" + name
					require.False(t, c.includeTask(path), "include %q exclude %q path %q", include, exclude, path)
				}
			}
		}
	}
}
