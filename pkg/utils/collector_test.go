package utils_test

import (
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/prometheus-community/windows_exporter/pkg/types"
	"github.com/prometheus-community/windows_exporter/pkg/utils"
)

func TestExpandChildCollectors(t *testing.T) {
	cases := []struct {
		name           string
		input          string
		expectedOutput []string
	}{
		{
			name:           "simple",
			input:          "testing1,testing2,testing3",
			expectedOutput: []string{"testing1", "testing2", "testing3"},
		},
		{
			name:           "duplicate",
			input:          "testing1,testing2,testing2,testing3",
			expectedOutput: []string{"testing1", "testing2", "testing3"},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			output := utils.ExpandEnabledChildCollectors(c.input)
			if !reflect.DeepEqual(output, c.expectedOutput) {
				t.Errorf("Output mismatch, expected %+v, got %+v", c.expectedOutput, output)
			}
		})
	}
}

func TestExpandEnabled(t *testing.T) {
	expansionTests := []struct {
		input          string
		expectedOutput []string
	}{
		{"", []string{}},
		// Default case
		{"cs,os", []string{"cs", "os"}},
		// Placeholder expansion
		{types.DefaultCollectorsPlaceholder, strings.Split(types.DefaultCollectors, ",")},
		// De-duplication
		{"cs,cs", []string{"cs"}},
		// De-duplicate placeholder
		{types.DefaultCollectorsPlaceholder + "," + types.DefaultCollectorsPlaceholder, strings.Split(types.DefaultCollectors, ",")},
		// Composite case
		{"foo," + types.DefaultCollectorsPlaceholder + ",bar", append(strings.Split(types.DefaultCollectors, ","), "foo", "bar")},
	}

	for _, testCase := range expansionTests {
		output := utils.ExpandEnabledCollectors(testCase.input)
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
