package collector

import (
	"testing"
)

func BenchmarkSmtpCollector(b *testing.B) {
	benchmarkCollector(b, "smtp", NewSMTPCollector)
}
