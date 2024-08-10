package perflib

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
)

// Conversion factors.
const (
	TicksToSecondScaleFactor = 1 / 1e7
	WindowsEpoch             = 116444736000000000
)

func UnmarshalObject(obj *PerfObject, vs interface{}, logger log.Logger) error {
	if obj == nil {
		return errors.New("counter not found")
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

		counters := make(map[string]*PerfCounter, len(instance.Counters))
		for _, ctr := range instance.Counters {
			if ctr.Def.IsBaseValue && !ctr.Def.IsNanosecondCounter {
				counters[ctr.Def.Name+"_Base"] = ctr
			} else {
				counters[ctr.Def.Name] = ctr
			}
		}

		for i := range target.NumField() {
			f := rt.Field(i)
			tag := f.Tag.Get("perflib")
			if tag == "" {
				continue
			}
			secondValue := false

			st := strings.Split(tag, ",")
			tag = st[0]

			for _, t := range st {
				if t == "secondvalue" {
					secondValue = true
				}
			}

			ctr, found := counters[tag]
			if !found {
				_ = level.Debug(logger).Log("msg", fmt.Sprintf("missing counter %q, have %v", tag, counterMapKeys(counters)))
				continue
			}
			if !target.Field(i).CanSet() {
				return fmt.Errorf("tagged field %v cannot be written to", f.Name)
			}
			if fieldType := target.Field(i).Type(); fieldType != reflect.TypeOf((*float64)(nil)).Elem() {
				return fmt.Errorf("tagged field %v has wrong type %v, must be float64", f.Name, fieldType)
			}

			if secondValue {
				if !ctr.Def.HasSecondValue {
					return fmt.Errorf("tagged field %v expected a SecondValue, which was not present", f.Name)
				}
				target.Field(i).SetFloat(float64(ctr.SecondValue))
				continue
			}

			switch ctr.Def.CounterType {
			case PERF_ELAPSED_TIME:
				target.Field(i).SetFloat(float64(ctr.Value-WindowsEpoch) / float64(obj.Frequency))
			case PERF_100NSEC_TIMER, PERF_PRECISION_100NS_TIMER:
				target.Field(i).SetFloat(float64(ctr.Value) * TicksToSecondScaleFactor)
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

func counterMapKeys(m map[string]*PerfCounter) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
