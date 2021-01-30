package collector

import (
	"fmt"
	"reflect"
	"strconv"

	perflibCollector "github.com/leoluk/perflib_exporter/collector"
	"github.com/leoluk/perflib_exporter/perflib"
	"github.com/prometheus-community/windows_exporter/log"
)

var nametable = perflib.QueryNameTable("Counter 009") // Reads the names in English TODO: validate that the English names are always present

func MapCounterToIndex(name string) string {
	return strconv.Itoa(int(nametable.LookupIndex(name)))
}

func getPerflibSnapshot(objNames string) (map[string]*perflib.PerfObject, error) {
	objects, err := perflib.QueryPerformanceData(objNames)
	if err != nil {
		return nil, err
	}

	indexed := make(map[string]*perflib.PerfObject)
	for _, obj := range objects {
		indexed[obj.Name] = obj
	}
	return indexed, nil
}

func unmarshalObject(obj *perflib.PerfObject, vs interface{}) error {
	if obj == nil {
		return fmt.Errorf("counter not found")
	}
	rv := reflect.ValueOf(vs)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return fmt.Errorf("%v is nil or not a pointer to slice", reflect.TypeOf(vs))
	}
	ev := rv.Elem()
	if ev.Kind() != reflect.Slice {
		return fmt.Errorf("%v is not slice", reflect.TypeOf(vs))
	}

	// Ensure sufficient length
	if ev.Cap() < len(obj.Instances) {
		nvs := reflect.MakeSlice(ev.Type(), len(obj.Instances), len(obj.Instances))
		ev.Set(nvs)
	}

	for idx, instance := range obj.Instances {
		target := ev.Index(idx)
		rt := target.Type()

		counters := make(map[string]*perflib.PerfCounter, len(instance.Counters))
		for _, ctr := range instance.Counters {
			if ctr.Def.IsBaseValue && !ctr.Def.IsNanosecondCounter {
				counters[ctr.Def.Name+"_Base"] = ctr
			} else {
				counters[ctr.Def.Name] = ctr
			}
		}

		for i := 0; i < target.NumField(); i++ {
			f := rt.Field(i)
			tag := f.Tag.Get("perflib")
			if tag == "" {
				continue
			}

			ctr, found := counters[tag]
			if !found {
				log.Debugf("missing counter %q, have %v", tag, counterMapKeys(counters))
				continue
			}
			if !target.Field(i).CanSet() {
				return fmt.Errorf("tagged field %v cannot be written to", f.Name)
			}
			if fieldType := target.Field(i).Type(); fieldType != reflect.TypeOf((*float64)(nil)).Elem() {
				return fmt.Errorf("tagged field %v has wrong type %v, must be float64", f.Name, fieldType)
			}

			switch ctr.Def.CounterType {
			case perflibCollector.PERF_ELAPSED_TIME:
				target.Field(i).SetFloat(float64(ctr.Value-windowsEpoch) / float64(obj.Frequency))
			case perflibCollector.PERF_100NSEC_TIMER, perflibCollector.PERF_PRECISION_100NS_TIMER:
				target.Field(i).SetFloat(float64(ctr.Value) * ticksToSecondsScaleFactor)
			default:
				target.Field(i).SetFloat(float64(ctr.Value))
			}
		}

		if instance.Name != "" && target.FieldByName("Name").CanSet() {
			target.FieldByName("Name").SetString(instance.Name)
		}
	}

	return nil
}

func counterMapKeys(m map[string]*perflib.PerfCounter) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
