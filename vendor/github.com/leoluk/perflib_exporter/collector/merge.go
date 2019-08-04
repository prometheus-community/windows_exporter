package collector

import "fmt"

var mergedDefinitions = map[uint]map[string]map[uint]string{
	230: {
		"processor_time_total": {
			0: "mode",
			6:   "",           // Processor Time (drop)
			142: "user",       // User Time
			144: "privileged", // Privileged Time
		},
	},
}

// Return if a given object has merge definitions
func HasMergedLabels(index uint) bool {
	_, ok := mergedDefinitions[index]
	return ok
}

// Return a list of merged label names for an instance
func MergedLabelsForInstance(objIndex uint, def uint) (name string, labelName string) {
	return MergedMetricForInstance(objIndex, 0)
}

// Return merged metric name and label value for an instance
func MergedMetricForInstance(objIndex uint, def uint) (name string, label string) {
	for k, v := range mergedDefinitions[objIndex] {
		for n := range v {
			if def == n {
				return k, v[n]
			}
		}
	}

	panic(fmt.Sprintf("No merge definition for obj %d, inst %d", objIndex, def))
}
