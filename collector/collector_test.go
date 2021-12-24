package collector

import (
	"reflect"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
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

func benchmarkCollector(b *testing.B, name string, collectFunc func() (Collector, error)) {
	// Create perflib scrape context. Some perflib collectors required a correct context,
	// or will fail during benchmark.
	scrapeContext, err := PrepareScrapeContext([]string{name})
	if err != nil {
		b.Error(err)
	}
	c, err := collectFunc()
	if err != nil {
		b.Error(err)
	}

	metrics := make(chan prometheus.Metric)
	go func() {
		for {
			<-metrics
		}
	}()

	for i := 0; i < b.N; i++ {
		c.Collect(scrapeContext, metrics) //nolint:errcheck
	}
}
