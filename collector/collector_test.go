package collector

import (
	"reflect"
	"testing"
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
			output := expandEnabledChildCollectors(c.input)
			if !reflect.DeepEqual(output, c.expectedOutput) {
				t.Errorf("Output mismatch, expected %+v, got %+v", c.expectedOutput, output)
			}
		})
	}
}
