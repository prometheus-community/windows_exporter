package collector

import (
	"testing"
)

type fakeWmiClass struct {
	Name         string
	SomeProperty int
}

var (
	mapQueryAll = func(src interface{}, class string, where string) string {
		return queryAll(src)
	}
	mapQueryAllWhere = func(src interface{}, class string, where string) string {
		return queryAllWhere(src, where)
	}
	mapQueryAllForClass = func(src interface{}, class string, where string) string {
		return queryAllForClass(src, class)
	}
	mapQueryAllForClassWhere = func(src interface{}, class string, where string) string {
		return queryAllForClassWhere(src, class, where)
	}
)

type queryFunc func(src interface{}, class string, where string) string

func TestCreateQuery(t *testing.T) {
	cases := []struct {
		desc      string
		dst       interface{}
		class     string
		where     string
		queryFunc queryFunc
		expected  string
	}{
		{
			desc:      "queryAll on single instance",
			dst:       fakeWmiClass{},
			queryFunc: mapQueryAll,
			expected:  "SELECT * FROM fakeWmiClass",
		},
		{
			desc:      "queryAll on slice",
			dst:       []fakeWmiClass{},
			queryFunc: mapQueryAll,
			expected:  "SELECT * FROM fakeWmiClass",
		},
		{
			desc:      "queryAllWhere on single instance",
			dst:       fakeWmiClass{},
			where:     "foo = bar",
			queryFunc: mapQueryAllWhere,
			expected:  "SELECT * FROM fakeWmiClass WHERE foo = bar",
		},
		{
			desc:      "queryAllWhere on slice",
			dst:       []fakeWmiClass{},
			where:     "foo = bar",
			queryFunc: mapQueryAllWhere,
			expected:  "SELECT * FROM fakeWmiClass WHERE foo = bar",
		},
		{
			desc:      "queryAllWhere on single instance with empty where",
			dst:       fakeWmiClass{},
			queryFunc: mapQueryAllWhere,
			expected:  "SELECT * FROM fakeWmiClass",
		},
		{
			desc:      "queryAllForClass on single instance",
			dst:       fakeWmiClass{},
			class:     "someClass",
			queryFunc: mapQueryAllForClass,
			expected:  "SELECT * FROM someClass",
		},
		{
			desc:      "queryAllForClass on slice",
			dst:       []fakeWmiClass{},
			class:     "someClass",
			queryFunc: mapQueryAllForClass,
			expected:  "SELECT * FROM someClass",
		},
		{
			desc:      "queryAllForClassWhere on single instance",
			dst:       fakeWmiClass{},
			class:     "someClass",
			where:     "foo = bar",
			queryFunc: mapQueryAllForClassWhere,
			expected:  "SELECT * FROM someClass WHERE foo = bar",
		},
		{
			desc:      "queryAllForClassWhere on slice",
			dst:       []fakeWmiClass{},
			class:     "someClass",
			where:     "foo = bar",
			queryFunc: mapQueryAllForClassWhere,
			expected:  "SELECT * FROM someClass WHERE foo = bar",
		},
		{
			desc:      "queryAllForClassWhere on single instance with empty where",
			dst:       fakeWmiClass{},
			class:     "someClass",
			queryFunc: mapQueryAllForClassWhere,
			expected:  "SELECT * FROM someClass",
		},
	}
	for _, c := range cases {
		t.Run(c.desc, func(t *testing.T) {
			if q := c.queryFunc(c.dst, c.class, c.where); q != c.expected {
				t.Errorf("Case %q failed: Expected %q, got %q", c.desc, c.expected, q)
			}
		})
	}
}
