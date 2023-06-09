//go:build windows
// +build windows

package main

import (
	"github.com/prometheus-community/windows_exporter/collector"
	"sort"
	"strings"
	"testing"
)

type expansionTestCase struct {
	input          string
	expectedOutput []string
}

func TestExpandEnabled(t *testing.T) {
	expansionTests := []expansionTestCase{
		{"", []string{}},
		// Default case
		{"cs,os", []string{"cs", "os"}},
		// Placeholder expansion
		{collector.DefaultCollectorsPlaceholder, strings.Split(collector.DefaultCollectors, ",")},
		// De-duplication
		{"cs,cs", []string{"cs"}},
		// De-duplicate placeholder
		{collector.DefaultCollectorsPlaceholder + "," + collector.DefaultCollectorsPlaceholder, strings.Split(collector.DefaultCollectors, ",")},
		// Composite case
		{"foo," + collector.DefaultCollectorsPlaceholder + ",bar", append(strings.Split(collector.DefaultCollectors, ","), "foo", "bar")},
	}

	for _, testCase := range expansionTests {
		output := collector.ExpandEnabledCollectors(testCase.input)
		sort.Strings(output)

		success := true
		if len(output) != len(testCase.expectedOutput) {
			success = false
		} else {
			sort.Strings(testCase.expectedOutput)
			for idx := range output {
				if output[idx] != testCase.expectedOutput[idx] {
					success = false
					break
				}
			}
		}
		if !success {
			t.Error("For", testCase.input, "expected", testCase.expectedOutput, "got", output)
		}
	}
}
