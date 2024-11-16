//go:build windows

package secur32_test

import (
	"testing"

	"github.com/prometheus-community/windows_exporter/internal/headers/secur32"
	"github.com/stretchr/testify/require"
)

func TestGetLogonSessions(t *testing.T) {
	t.Parallel()

	sessionData, err := secur32.GetLogonSessions()
	require.NoError(t, err)
	require.NotEmpty(t, sessionData)
}
