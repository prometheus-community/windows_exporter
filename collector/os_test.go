package collector

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"
)

func BenchmarkOsCollect(b *testing.B) {
	o, err := NewOSCollector()
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
		o.Collect(&ScrapeContext{}, metrics)
	}
}
