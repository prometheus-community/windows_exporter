//go:build windows
// +build windows

package main

import (
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
		{defaultCollectorsPlaceholder, strings.Split(defaultCollectors, ",")},
		// De-duplication
		{"cs,cs", []string{"cs"}},
		// De-duplicate placeholder
		{defaultCollectorsPlaceholder + "," + defaultCollectorsPlaceholder, strings.Split(defaultCollectors, ",")},
		// Composite case
		{"foo," + defaultCollectorsPlaceholder + ",bar", append(strings.Split(defaultCollectors, ","), "foo", "bar")},
	}

	for _, testCase := range expansionTests {
		output := expandEnabledCollectors(testCase.input)
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
