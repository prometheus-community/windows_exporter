package collector

import (
	"testing"
)

func BenchmarkRemoteFXCollector(b *testing.B) {
	benchmarkCollector(b, "remote_fx", NewRemoteFx)
}
