package collector

import (
	"bytes"
	"reflect"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
)

// ...
const (
	Namespace = "wmi"

	// Conversion factors
	ticksToSecondsScaleFactor = 1 / 1e7
)

// Factories ...
var Factories = make(map[string]func() (Collector, error))

// Collector is the interface a collector has to implement.
type Collector interface {
	// Get new metrics and expose them via prometheus registry.
	Collect(ch chan<- prometheus.Metric) (err error)
}

// This is adapted from StackExchange/wmi/wmi.go, and lets us change the class
// name being queried for:
// CreateQuery returns a WQL query string that queries all columns of src. where
// is an optional string that is appended to the query, to be used with WHERE
// clauses. In such a case, the "WHERE" string should appear at the beginning.
func createQuery(src interface{}, class, where string) string {
	var b bytes.Buffer
	b.WriteString("SELECT ")
	s := reflect.Indirect(reflect.ValueOf(src))
	t := s.Type()
	if s.Kind() == reflect.Slice {
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct {
		return ""
	}
	var fields []string
	for i := 0; i < t.NumField(); i++ {
		fields = append(fields, t.Field(i).Name)
	}
	b.WriteString(strings.Join(fields, ", "))
	b.WriteString(" FROM ")
	b.WriteString(class)
	b.WriteString(" " + where)
	return b.String()
}
