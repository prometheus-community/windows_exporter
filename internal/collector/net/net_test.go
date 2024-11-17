//go:build windows

package net_test

import (
	"testing"

	"github.com/prometheus-community/windows_exporter/internal/collector/net"
	"github.com/prometheus-community/windows_exporter/internal/utils/testutils"
)

func TestCollector(t *testing.T) {
	testutils.TestCollector(t, net.New, nil)
}
