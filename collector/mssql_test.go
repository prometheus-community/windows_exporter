package collector

import (
	"testing"
)

func BenchmarkMSSQLCollector(b *testing.B) {
	benchmarkCollector(b, "mssql", NewMSSQLCollector)
}
