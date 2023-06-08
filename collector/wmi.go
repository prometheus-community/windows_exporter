package collector

import (
	"bytes"
	"fmt"
	"reflect"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
)

func className(src interface{}) string {
	s := reflect.Indirect(reflect.ValueOf(src))
	t := s.Type()
	if s.Kind() == reflect.Slice {
		t = t.Elem()
	}
	return t.Name()
}

func queryAll(src interface{}, logger log.Logger) string {
	var b bytes.Buffer
	b.WriteString("SELECT * FROM ")
	b.WriteString(className(src))

	_ = level.Debug(logger).Log("msg", fmt.Sprintf("Generated WMI query %s", b.String()))
	return b.String()
}

func queryAllForClass(src interface{}, class string, logger log.Logger) string {
	var b bytes.Buffer
	b.WriteString("SELECT * FROM ")
	b.WriteString(class)

	_ = level.Debug(logger).Log("msg", fmt.Sprintf("Generated WMI query %s", b.String()))
	return b.String()
}

func queryAllWhere(src interface{}, where string, logger log.Logger) string {
	var b bytes.Buffer
	b.WriteString("SELECT * FROM ")
	b.WriteString(className(src))

	if where != "" {
		b.WriteString(" WHERE ")
		b.WriteString(where)
	}

	_ = level.Debug(logger).Log("msg", fmt.Sprintf("Generated WMI query %s", b.String()))
	return b.String()
}

func queryAllForClassWhere(src interface{}, class string, where string, logger log.Logger) string {
	var b bytes.Buffer
	b.WriteString("SELECT * FROM ")
	b.WriteString(class)

	if where != "" {
		b.WriteString(" WHERE ")
		b.WriteString(where)
	}

	_ = level.Debug(logger).Log("msg", fmt.Sprintf("Generated WMI query %s", b.String()))
	return b.String()
}
