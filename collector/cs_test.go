package collector

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"
)

func BenchmarkCsCollect(b *testing.B) {
	c, err := NewCSCollector()
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
		c.Collect(&ScrapeContext{}, metrics)
	}
}
