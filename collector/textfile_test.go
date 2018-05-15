package collector

import (
	"testing"
	"strings"
	"io/ioutil"
)

func TestCRFilter(t *testing.T) {
	sr := strings.NewReader("line 1\r\nline 2")
	cr := carriageReturnFilteringReader{ r: sr }
	b, err := ioutil.ReadAll(cr)
	if err != nil {
		t.Error(err)
	}

	if string(b) != "line 1\nline 2" {
		t.Errorf("Unexpected output %q", b)
	}
}