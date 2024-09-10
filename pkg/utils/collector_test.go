package utils_test

import (
	"sort"
	"strings"
	"testing"

	"github.com/prometheus-community/windows_exporter/pkg/types"
	"github.com/prometheus-community/windows_exporter/pkg/utils"
)

func TestExpandEnabled(t *testing.T) {
	t.Parallel()

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
