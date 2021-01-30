package collector

import (
	"bytes"
	"reflect"

	"github.com/prometheus-community/windows_exporter/log"
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

	log.Debugf("Generated WMI query %s", b.String())
	return b.String()
}

func queryAllForClass(src interface{}, class string) string {
	var b bytes.Buffer
	b.WriteString("SELECT * FROM ")
	b.WriteString(class)

	log.Debugf("Generated WMI query %s", b.String())
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

	log.Debugf("Generated WMI query %s", b.String())
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

	log.Debugf("Generated WMI query %s", b.String())
	return b.String()
}
