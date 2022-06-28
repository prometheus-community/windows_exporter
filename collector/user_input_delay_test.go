package collector

import (
	"reflect"
	"testing"
)

func BenchmarkUserInputDelayCollector(b *testing.B) {
	benchmarkCollector(b, "user_input_delay", NewUserInputCollector)
}

func TestSplitProcessLabel(t *testing.T) {
	data := "1:1268 <svchost.exe>"
	expectedOutput := processLabels{
		sessionID:   "1",
		processID:   "1268",
		processName: "svchost.exe",
	}

	result, err := splitProcessLabel(data)
	if err != nil {
		t.Error(err)
	}

	if !reflect.DeepEqual(result, &expectedOutput) {
		t.Errorf("Output mismatch, expected %+v, got %+v", expectedOutput, result)
	}
}
