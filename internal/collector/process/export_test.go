// SPDX-License-Identifier: MIT
//
// Copyright The Prometheus Authors

//go:build windows

package process

// WorkerProcessAppPool exposes the command line source to the IIS tests, which
// must live in process_test because testutils imports the collector registry.
//
//nolint:gochecknoglobals
var WorkerProcessAppPool = workerProcessAppPool
