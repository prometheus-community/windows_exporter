// SPDX-License-Identifier: MIT
//
// Copyright The Prometheus Authors

//go:build windows

package msmq

type perfDataCounterValues struct {
	Name string

	BytesInJournalQueue    float64 `perfdata:"Bytes in Journal Queue"`
	BytesInQueue           float64 `perfdata:"Bytes in Queue"`
	MessagesInJournalQueue float64 `perfdata:"Messages in Journal Queue"`
	MessagesInQueue        float64 `perfdata:"Messages in Queue"`
}
