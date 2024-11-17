//go:build windows

package httphandler

import (
	"net/http"
)

type HealthHandler struct{}

// Interface guard.
var _ http.Handler = (*HealthHandler)(nil)

func NewHealthHandler() HealthHandler {
	return HealthHandler{}
}

func (h HealthHandler) ServeHTTP(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write([]byte(`{"status":"ok"}`))
}
