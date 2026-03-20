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

package registry

import (
	"testing"
)

// TestNewCollectorStructTypeParam guards against a regression where
// reflect.TypeFor[T]().Elem() panicked when T is a plain struct (not a pointer).
// See https://github.com/prometheus-community/windows_exporter/issues/2365
func TestNewCollectorStructTypeParam(t *testing.T) {
	type systemCounterValues struct {
		Name string

		ProcessorQueueLength float64 `perfdata:"Processor Queue Length"`
	}

	_, err := NewCollector[systemCounterValues]("System", nil)
	if err != nil {
		t.Skipf("skipping: failed to create collector: %v", err)
	}
}

func BenchmarkQueryPerformanceData(b *testing.B) {
	for b.Loop() {
		_, _ = QueryPerformanceData("Global", "")
	}
}
