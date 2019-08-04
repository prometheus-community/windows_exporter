package collector

import (
	"fmt"
	"strings"

	"github.com/leoluk/perflib_exporter/perflib"
	"github.com/prometheus/client_golang/prometheus"
)

func manglePerflibName(s string) string {
	s = strings.ToLower(s)
	s = strings.Replace(s, " ", "_", -1)
	s = strings.Replace(s, ".", "", -1)
	s = strings.Replace(s, "(", "", -1)
	s = strings.Replace(s, ")", "", -1)
	s = strings.Replace(s, "+", "", -1)
	s = strings.Replace(s, "-", "", -1)
	s = strings.Replace(s, ",", "", -1)

	return s
}

func manglePerflibCounterName(s string) string {
	s = manglePerflibName(s)

	s = strings.Replace(s, "total_", "", -1)
	s = strings.Replace(s, "_total", "", -1)
	s = strings.Replace(s, "/second", "", -1)
	s = strings.Replace(s, "/sec", "", -1)
	s = strings.Replace(s, "_%", "", -1)
	s = strings.Replace(s, "%_", "", -1)
	s = strings.Replace(s, "/", "_per_", -1)
	s = strings.Replace(s, "&", "and", -1)
	s = strings.Replace(s, "#_of_", "", -1)
	s = strings.Replace(s, ":", "", -1)
	s = strings.Replace(s, "__", "_", -1)

	s = strings.Trim(s, " _")

	return s
}

func MakePrometheusLabel(def *perflib.PerfCounterDef) (s string) {
	s = manglePerflibCounterName(def.Name)

	if len(s) > 0 {
		if IsCounter(def.CounterType) {
			s += "_total"
		} else if IsBaseValue(def.CounterType) && !strings.HasSuffix(s, "_base") {
			s += "_max"
		}
	}

	return
}

func pdhNameFromCounterDef(obj perflib.PerfObject, def perflib.PerfCounterDef) string {
	return fmt.Sprintf(`\%s(*)\%s`, obj.Name, def.Name)
}

func descFromCounterDef(obj perflib.PerfObject, def perflib.PerfCounterDef) *prometheus.Desc {
	subsystem := manglePerflibName(obj.Name)
	counterName := MakePrometheusLabel(&def)

	labels := []string{"name"}

	if len(obj.Instances) == 1 {
		labels = []string{}
	}

	if HasPromotedLabels(obj.NameIndex) {
		labels = append(labels, PromotedLabelsForObject(obj.NameIndex)...)
	}

	// TODO - Label merging needs to be fixed for [230] Process
	//if HasMergedLabels(obj.NameIndex) {
	//	s, labelsForObject := MergedLabelsForInstance(obj.NameIndex, def.NameIndex)
	//	counterName = s
	//	labels = append(labels, labelsForObject)
	//}

	return prometheus.NewDesc(
		prometheus.BuildFQName(Namespace, subsystem, counterName),
		fmt.Sprintf("perflib metric: %s (see /dump for docs) [%d]",
			pdhNameFromCounterDef(obj, def), def.NameIndex),
		labels,
		nil,
	)
}
