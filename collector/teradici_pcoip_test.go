package collector

import (
	"testing"
)

func BenchmarkTeradiciPcoipCollector(b *testing.B) {
	benchmarkCollector(b, "teradici_pcoip", newTeradiciPcoipCollector)
}
