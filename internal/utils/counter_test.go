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

package utils_test

import (
	"math"
	"testing"

	"github.com/prometheus-community/windows_exporter/internal/utils"
	"github.com/stretchr/testify/assert"
)

func TestCounter(t *testing.T) {
	t.Parallel()

	c := utils.NewCounter(0)
	assert.Equal(t, 0.0, c.Value()) //nolint:testifylint

	c.AddValue(1)

	assert.Equal(t, 1.0, c.Value()) //nolint:testifylint

	c.AddValue(math.MaxUint32)

	assert.Equal(t, float64(math.MaxUint32)+1.0, c.Value()) //nolint:testifylint
}
