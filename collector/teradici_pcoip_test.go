package collector

import (
	"testing"
)

func benchmarkTeradiciPcoipCollector(b *testing.B) {
	benchmarkCollector(b, "teradici_pcoip", newTeradiciPcoipCollector)
}
