package collector

import (
	"bytes"
	"reflect"

	"github.com/go-kit/kit/log/level"
)

func className(src interface{}) string {
	s := reflect.Indirect(reflect.ValueOf(src))
	t := s.Type()
	if s.Kind() == reflect.Slice {
		t = t.Elem()
	}
	return t.Name()
}

func queryAll(src interface{}) string {
	var b bytes.Buffer
	b.WriteString("SELECT * FROM ")
	b.WriteString(className(src))

	level.Debug(logger).Log("msg", "Generated WMI query", "query", b.String())
	return b.String()
}

func queryAllForClass(src interface{}, class string) string {
	var b bytes.Buffer
	b.WriteString("SELECT * FROM ")
	b.WriteString(class)

	level.Debug(logger).Log("msg", "Generated WMI query", "query", b.String())
	return b.String()
}

func queryAllWhere(src interface{}, where string) string {
	var b bytes.Buffer
	b.WriteString("SELECT * FROM ")
	b.WriteString(className(src))

	if where != "" {
		b.WriteString(" WHERE ")
		b.WriteString(where)
	}

	level.Debug(logger).Log("msg", "Generated WMI query", "query", b.String())
	return b.String()
}

func queryAllForClassWhere(src interface{}, class string, where string) string {
	var b bytes.Buffer
	b.WriteString("SELECT * FROM ")
	b.WriteString(class)

	if where != "" {
		b.WriteString(" WHERE ")
		b.WriteString(where)
	}

	level.Debug(logger).Log("msg", "Generated WMI query", "query", b.String())
	return b.String()
}
