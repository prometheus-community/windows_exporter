// Copyright 2024 The Prometheus Authors
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

package performancecounter_test

import (
	"testing"

	"github.com/alecthomas/kingpin/v2"
	"github.com/prometheus-community/windows_exporter/internal/collector/performancecounter"
	"github.com/prometheus-community/windows_exporter/internal/utils/testutils"
)

func BenchmarkCollector(b *testing.B) {
	perfDataObjects := `[{"object":"Processor Information","instances":["*"],"instance_label":"core","counters":[{"name":"% Processor Time","metric":"windows_performancecounter_processor_information_processor_time","labels":{"state":"active"}},{"name":"% Idle Time","metric":"windows_performancecounter_processor_information_processor_time","labels":{"state":"idle"}}]},{"object":"Memory","counters":[{"name":"Cache Faults/sec","type":"counter"}]}]`

	testutils.FuncBenchmarkCollector(b, performancecounter.Name, performancecounter.NewWithFlags, func(app *kingpin.Application) {
		app.GetFlag("collector.perfdata.objects").StringVar(&perfDataObjects)
	})
}
