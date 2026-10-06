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

package httphandler_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/prometheus-community/windows_exporter/internal/httphandler"
	"github.com/prometheus/common/version"
	"github.com/stretchr/testify/require"
)

func TestVersionHandler(t *testing.T) {
	t.Parallel()

	response := httptest.NewRecorder()
	httphandler.NewVersionHandler().ServeHTTP(response, httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/version", nil))
	require.Equal(t, http.StatusOK, response.Code)

	var got map[string]string
	require.NoError(t, json.Unmarshal(response.Body.Bytes(), &got))
	require.Equal(t, map[string]string{
		"version": version.Version, "revision": version.Revision,
		"branch": version.Branch, "buildUser": version.BuildUser,
		"buildDate": version.BuildDate, "goVersion": version.GoVersion,
	}, got)
}
