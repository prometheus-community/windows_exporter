package collector

import (
	"strconv"

	"github.com/leoluk/perflib_exporter/perflib"
)

var labelPromotionLabels = map[uint][]string{
	230: {
		"process_id",
		"creating_process_id",
	},
}
var labelPromotionValues = map[uint][]uint{
	230: {
		784,  // process_id
		1410, // creating_process_id
	},
}

// Get a list of promoted labels for an object
func PromotedLabelsForObject(index uint) []string {
	return labelPromotionLabels[index]
}

// Get a list of label values for a given object and instance
func PromotedLabelValuesForInstance(index uint, instance *perflib.PerfInstance) []string {
	values := make([]string, len(labelPromotionValues[index]))

	for _, c := range instance.Counters {
		for i, v := range labelPromotionValues[index] {
			if c.Def.NameIndex == v {
				values[i] = strconv.Itoa(int(c.Value))
			}
		}
	}

	return values
}

// Return if a given object has label promotion definitions
func HasPromotedLabels(index uint) bool {
	_, ok := labelPromotionLabels[index]
	return ok
}

// Return if a given definition is a promoted label for an object
func IsDefPromotedLabel(objIndex uint, def uint) bool {
	for _, v := range labelPromotionValues[objIndex] {
		if v == def {
			return true
		}
	}
	return false
}
