package perflib

import (
	"strconv"
)

func MapCounterToIndex(name string) string {
	return strconv.Itoa(int(CounterNameTable.LookupIndex(name)))
}

func GetPerflibSnapshot(objNames string) (map[string]*PerfObject, error) {
	objects, err := QueryPerformanceData(objNames)
	if err != nil {
		return nil, err
	}

	indexed := make(map[string]*PerfObject)
	for _, obj := range objects {
		indexed[obj.Name] = obj
	}
	return indexed, nil
}
