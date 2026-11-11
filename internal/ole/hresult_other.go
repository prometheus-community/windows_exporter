// SPDX-License-Identifier: MIT
//
// Copyright The Prometheus Authors

//go:build !windows

package ole

import "fmt"

func formatHRESULT(h HRESULT) string {
	return fmt.Sprintf("COM HRESULT 0x%08X", uint32(h))
}
