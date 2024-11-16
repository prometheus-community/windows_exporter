//go:build windows

package config

import (
	"fmt"
	"strconv"
)

// flatten flattens the nested struct.
//
// All keys will be joined by dot
// e.g. {"a": {"b":"c"}} => {"a.b":"c"}
// or {"a": {"b":[1,2]}} => {"a.b.0":1, "a.b.1": 2}.
func flatten(data map[string]interface{}) map[string]string {
	ret := make(map[string]string)

	for k, v := range data {
		switch typed := v.(type) {
		case map[interface{}]interface{}:
			for fk, fv := range flatten(convertMap(typed)) {
				ret[fmt.Sprintf("%s.%s", k, fk)] = fv
			}
		case map[string]interface{}:
			for fk, fv := range flatten(typed) {
				ret[fmt.Sprintf("%s.%s", k, fk)] = fv
			}
		case []interface{}:
			for fk, fv := range flattenSlice(typed) {
				ret[fmt.Sprintf("%s.%s", k, fk)] = fv
			}
		default:
			ret[k] = fmt.Sprint(typed)
		}
	}

	return ret
}

func flattenSlice(data []interface{}) map[string]string {
	ret := make(map[string]string)

	for idx, v := range data {
		switch typed := v.(type) {
		case map[interface{}]interface{}:
			for fk, fv := range flatten(convertMap(typed)) {
				ret[fmt.Sprintf("%d,%s", idx, fk)] = fv
			}
		case map[string]interface{}:
			for fk, fv := range flatten(typed) {
				ret[fmt.Sprintf("%d,%s", idx, fk)] = fv
			}
		case []interface{}:
			for fk, fv := range flattenSlice(typed) {
				ret[fmt.Sprintf("%d,%s", idx, fk)] = fv
			}
		default:
			ret[strconv.Itoa(idx)] = fmt.Sprint(typed)
		}
	}

	return ret
}

func convertMap(originalMap map[interface{}]interface{}) map[string]interface{} {
	convertedMap := map[string]interface{}{}

	for key, value := range originalMap {
		if keyString, ok := key.(string); ok {
			convertedMap[keyString] = value
		}
	}

	return convertedMap
}
