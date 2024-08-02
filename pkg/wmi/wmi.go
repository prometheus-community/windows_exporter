package wmi

import (
	"bytes"
	"reflect"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/yusufpapurcu/wmi"
)

func InitWbem(logger log.Logger) error {
	// This initialization prevents a memory leak on WMF 5+. See
	// https://github.com/prometheus-community/windows_exporter/issues/77 and
	// linked issues for details.
	_ = level.Debug(logger).Log("msg", "Initializing SWbemServices")
	s, err := wmi.InitializeSWbemServices(wmi.DefaultClient)
	if err != nil {
		return err
	}
	wmi.DefaultClient.AllowMissingFields = true
	wmi.DefaultClient.SWbemServicesClient = s

	return nil
}

func className(src interface{}) string {
	s := reflect.Indirect(reflect.ValueOf(src))
	t := s.Type()
	if s.Kind() == reflect.Slice {
		t = t.Elem()
	}
	return t.Name()
}

func Query(query string, dst interface{}, connectServerArgs ...interface{}) error {
	return wmi.Query(query, dst, connectServerArgs...)
}

func QueryNamespace(query string, dst interface{}, namespace string) error {
	return wmi.QueryNamespace(query, dst, namespace)
}

func QueryAll(src interface{}, logger log.Logger) string {
	var b bytes.Buffer
	b.WriteString("SELECT * FROM ")
	b.WriteString(className(src))

	_ = level.Debug(logger).Log("msg", "Generated WMI query "+b.String())
	return b.String()
}

func QueryAllForClass(_ interface{}, class string, logger log.Logger) string {
	var b bytes.Buffer
	b.WriteString("SELECT * FROM ")
	b.WriteString(class)

	_ = level.Debug(logger).Log("msg", "Generated WMI query "+b.String())
	return b.String()
}

func QueryAllWhere(src interface{}, where string, logger log.Logger) string {
	var b bytes.Buffer
	b.WriteString("SELECT * FROM ")
	b.WriteString(className(src))

	if where != "" {
		b.WriteString(" WHERE ")
		b.WriteString(where)
	}

	_ = level.Debug(logger).Log("msg", "Generated WMI query "+b.String())
	return b.String()
}

func QueryAllForClassWhere(_ interface{}, class string, where string, logger log.Logger) string {
	var b bytes.Buffer
	b.WriteString("SELECT * FROM ")
	b.WriteString(class)

	if where != "" {
		b.WriteString(" WHERE ")
		b.WriteString(where)
	}

	_ = level.Debug(logger).Log("msg", "Generated WMI query "+b.String())
	return b.String()
}
