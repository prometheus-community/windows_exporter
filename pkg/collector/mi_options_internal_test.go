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

package collector

import (
	"testing"
	"time"

	"github.com/prometheus-community/windows_exporter/internal/mi"
	"github.com/stretchr/testify/require"
)

func TestMISessionAfterCallerOptionsDeleted(t *testing.T) {
	t.Parallel()

	app, err := mi.ApplicationInitialize()
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, app.Close()) })

	options, err := app.NewDestinationOptions()
	require.NoError(t, err)
	// newMISession releases the original options before returning.
	session, err := newMISession(app, options)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, session.Close()) })

	query, err := mi.NewQuery("SELECT Caption FROM Win32_OperatingSystem")
	require.NoError(t, err)

	var operatingSystems []struct{ Caption string }
	require.NoError(t, session.Query(&operatingSystems, mi.NamespaceRootCIMv2, query, 5*time.Second))
	require.NotEmpty(t, operatingSystems)
}
