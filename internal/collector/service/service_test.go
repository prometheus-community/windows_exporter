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

package service_test

import (
	"testing"

	"github.com/alecthomas/kingpin/v2"
	"github.com/prometheus-community/windows_exporter/internal/collector/service"
	"github.com/prometheus-community/windows_exporter/internal/utils/testutils"
	"github.com/stretchr/testify/require"
)

func BenchmarkCollector(b *testing.B) {
	testutils.FuncBenchmarkCollector(b, service.Name, service.NewWithFlags)
}

func TestCollector(t *testing.T) {
	testutils.TestCollector(t, service.New, nil)
}

func TestNewWithFlagsReportsRegexpFlagName(t *testing.T) {
	t.Parallel()

	for _, flagName := range []string{"collector.service.exclude", "collector.service.include"} {
		t.Run(flagName, func(t *testing.T) {
			t.Parallel()

			app := kingpin.New("test", "")
			service.NewWithFlags(app)

			_, err := app.Parse([]string{"--" + flagName + "=("})
			require.ErrorContains(t, err, flagName+":")
		})
	}
}
