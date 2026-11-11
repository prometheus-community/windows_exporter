// SPDX-License-Identifier: MIT
//
// Copyright The Prometheus Authors

//go:build windows

package httphandler_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/prometheus-community/windows_exporter/internal/httphandler"
	"github.com/stretchr/testify/require"
)

func TestHealthHandler(t *testing.T) {
	t.Parallel()

	response := httptest.NewRecorder()
	httphandler.NewHealthHandler().ServeHTTP(response, httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/health", nil))
	require.Equal(t, http.StatusOK, response.Code)
	require.Equal(t, "application/json", response.Header().Get("Content-Type"))
	require.JSONEq(t, `{"status":"ok"}`, response.Body.String())
}
