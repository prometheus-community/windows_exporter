// SPDX-License-Identifier: MIT
//
// Copyright The Prometheus Authors

//go:build windows

package process

type WorkerProcess struct {
	AppPoolName string `mi:"AppPoolName"`
	ProcessId   uint64 `mi:"ProcessId"`
}
