package perflib

import (
	"fmt"
	"testing"
)

func ExampleQueryPerformanceData() {
	objects, err := QueryPerformanceData("2")

	if err != nil {
		panic(err)
	}

	for _, object := range objects {
		fmt.Printf("%d %s [%d counters, %d instances]\n",
			object.NameIndex, object.Name, len(object.CounterDefs), len(object.Instances))

		for _, instance := range object.Instances {
			if !((instance.Name == "_Total") || (instance.Name == "")) {
				continue
			}

			if instance.Name == "" {
				fmt.Println("No instance.", instance.Name)
			} else {
				fmt.Println("Instance:", instance.Name)
			}

			for _, counter := range instance.Counters {
				fmt.Printf("  -> %s %d\n", counter.Def.Name, counter.Def.NameIndex)
			}
		}
	}
}

func BenchmarkQueryPerformanceData(b *testing.B) {
	for n := 0; n < b.N; n++ {
		_, _ = QueryPerformanceData("Global")
	}
}
