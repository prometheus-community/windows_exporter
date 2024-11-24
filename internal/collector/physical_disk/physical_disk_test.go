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

package physical_disk_test

import (
	"testing"

	"github.com/prometheus-community/windows_exporter/internal/collector/physical_disk"
	"github.com/prometheus-community/windows_exporter/internal/utils/testutils"
	"github.com/prometheus-community/windows_exporter/pkg/types"
)

func BenchmarkCollector(b *testing.B) {
	testutils.FuncBenchmarkCollector(b, physical_disk.Name, physical_disk.NewWithFlags)
}

func TestCollector(t *testing.T) {
	testutils.TestCollector(t, physical_disk.New, &physical_disk.Config{
		DiskInclude: types.RegExpAny,
	})
}
