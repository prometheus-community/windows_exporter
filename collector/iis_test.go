package collector

import (
	"reflect"
	"testing"
)

func BenchmarkIISCollector(b *testing.B) {
	benchmarkCollector(b, "iis", newIISCollector)
}

func TestIISDeduplication(t *testing.T) {
	start := []perflibAPP_POOL_WAS{
		{
			Name:             "foo",
			Frequency_Object: 1,
		},
		{
			Name:             "foo1#999",
			Frequency_Object: 2,
		},
		{
			Name:             "foo#2",
			Frequency_Object: 3,
		},
		{
			Name:             "bar$2",
			Frequency_Object: 4,
		},
		{
			Name:             "bar_2",
			Frequency_Object: 5,
		},
	}
	var expected = make(map[string]perflibAPP_POOL_WAS)
	// Should be deduplicated from "foo#2"
	expected["foo"] = perflibAPP_POOL_WAS{Name: "foo#2", Frequency_Object: 3}
	// Map key should have suffix stripped, but struct name field should be unchanged
	expected["foo1"] = perflibAPP_POOL_WAS{Name: "foo1#999", Frequency_Object: 2}
	// Map key and Name should be identical, as there is no suffix starting with "#"
	expected["bar$2"] = perflibAPP_POOL_WAS{Name: "bar$2", Frequency_Object: 4}
	// Map key and Name should be identical, as there is no suffix starting with "#"
	expected["bar_2"] = perflibAPP_POOL_WAS{Name: "bar_2", Frequency_Object: 5}

	deduplicated := dedupIISNames(start)
	if !reflect.DeepEqual(expected, deduplicated) {
		t.Errorf("Flattened values do not match!\nExpected result: %+v\nActual result: %+v", expected, deduplicated)
	}
}
