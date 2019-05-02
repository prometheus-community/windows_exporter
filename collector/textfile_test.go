package collector

import (
	"github.com/dimchansky/utfbom"
	"io/ioutil"
	"strings"
	"testing"
)

func TestCRFilter(t *testing.T) {
	sr := strings.NewReader("line 1\r\nline 2")
	cr := carriageReturnFilteringReader{r: sr}
	b, err := ioutil.ReadAll(cr)
	if err != nil {
		t.Error(err)
	}

	if string(b) != "line 1\nline 2" {
		t.Errorf("Unexpected output %q", b)
	}
}

func TestCheckBOM(t *testing.T) {
	testdata := []struct {
		encoding utfbom.Encoding
		err      string
	}{
		{utfbom.Unknown, ""},
		{utfbom.UTF8, ""},
		{utfbom.UTF16BigEndian, "UTF16BigEndian"},
		{utfbom.UTF16LittleEndian, "UTF16LittleEndian"},
		{utfbom.UTF32BigEndian, "UTF32BigEndian"},
		{utfbom.UTF32LittleEndian, "UTF32LittleEndian"},
	}
	for _, d := range testdata {
		err := checkBOM(d.encoding)
		if d.err == "" && err != nil {
			t.Error(err)
		}
		if d.err != "" && err == nil {
			t.Errorf("Missing expected error %s", d.err)
		}
		if err != nil && !strings.Contains(err.Error(), d.err) {
			t.Error(err)
		}
	}
}
