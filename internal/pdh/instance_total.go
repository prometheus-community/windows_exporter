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

package pdh

import (
	"strconv"
	"strings"
)

// IsTotalInstance identifies aggregate instances without discarding ordinary
// application names ending in _Total. Processor Information also publishes
// processor-group aggregates named <group>,_Total.
func IsTotalInstance(object, instance string) bool {
	if instance == InstanceTotal {
		return true
	}

	if !strings.EqualFold(object, "Processor Information") {
		return false
	}

	group, ok := strings.CutSuffix(instance, ","+InstanceTotal)
	if !ok {
		return false
	}

	_, err := strconv.ParseUint(group, 10, 32)

	return err == nil
}
