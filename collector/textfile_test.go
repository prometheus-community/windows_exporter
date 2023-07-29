package collector

import (
	"fmt"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/dimchansky/utfbom"
	"github.com/go-kit/log"

	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
)

var baseDir = "../tools/textfile-test"

func TestCRFilter(t *testing.T) {
	sr := strings.NewReader("line 1\r\nline 2")
	cr := carriageReturnFilteringReader{r: sr}
	b, err := io.ReadAll(cr)
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

func TestDuplicateMetricEntry(t *testing.T) {
	metric_name := "windows_sometest"
	metric_help := "This is a Test."
	metric_type := dto.MetricType_GAUGE

	gauge_value := 1.0

	gauge := dto.Gauge{
		Value: &gauge_value,
	}

	label1_name := "display_name"
	label1_value := "foobar"

	label1 := dto.LabelPair{
		Name:  &label1_name,
		Value: &label1_value,
	}

	label2_name := "display_version"
	label2_value := "13.4.0"

	label2 := dto.LabelPair{
		Name:  &label2_name,
		Value: &label2_value,
	}

	metric1 := dto.Metric{
		Label: []*dto.LabelPair{&label1, &label2},
		Gauge: &gauge,
	}

	metric2 := dto.Metric{
		Label: []*dto.LabelPair{&label1, &label2},
		Gauge: &gauge,
	}

	duplicate := dto.MetricFamily{
		Name:   &metric_name,
		Help:   &metric_help,
		Type:   &metric_type,
		Metric: []*dto.Metric{&metric1, &metric2},
	}

	duplicateFamily := []*dto.MetricFamily{}
	duplicateFamily = append(duplicateFamily, &duplicate)

	// Ensure detection for duplicate metrics
	if !duplicateMetricEntry(duplicateFamily) {
		t.Errorf("Duplicate not found in duplicateFamily")
	}

	label3_name := "test"
	label3_value := "1.0"

	label3 := dto.LabelPair{
		Name:  &label3_name,
		Value: &label3_value,
	}
	metric3 := dto.Metric{
		Label: []*dto.LabelPair{&label1, &label2, &label3},
		Gauge: &gauge,
	}

	differentLabels := dto.MetricFamily{
		Name:   &metric_name,
		Help:   &metric_help,
		Type:   &metric_type,
		Metric: []*dto.Metric{&metric1, &metric3},
	}

	duplicateFamily = []*dto.MetricFamily{}
	duplicateFamily = append(duplicateFamily, &differentLabels)

	// Additional label on second metric should not be cause for duplicate detection
	if duplicateMetricEntry(duplicateFamily) {
		t.Errorf("Unexpected duplicate found in differentLabels")
	}

	label4_value := "2.0"

	label4 := dto.LabelPair{
		Name:  &label3_name,
		Value: &label4_value,
	}
	metric4 := dto.Metric{
		Label: []*dto.LabelPair{&label1, &label2, &label4},
		Gauge: &gauge,
	}

	differentValues := dto.MetricFamily{
		Name:   &metric_name,
		Help:   &metric_help,
		Type:   &metric_type,
		Metric: []*dto.Metric{&metric3, &metric4},
	}
	duplicateFamily = []*dto.MetricFamily{}
	duplicateFamily = append(duplicateFamily, &differentValues)

	// Additional label with different values metric should not be cause for duplicate detection
	if duplicateMetricEntry(duplicateFamily) {
		t.Errorf("Unexpected duplicate found in differentValues")
	}
}

func TestMultipleDirectories(t *testing.T) {
	testDir := baseDir + "/multiple-dirs"
	testDirs := fmt.Sprintf("%[1]s/dir1,%[1]s/dir2,%[1]s/dir3", testDir)
	collector := &textFileCollector{
		logger:      log.NewLogfmtLogger(os.Stdout),
		directories: testDirs,
	}
	scrapeContext, err := PrepareScrapeContext([]string{"textfile_test"})
	if err != nil {
		t.Errorf("Unexpected error %s", err)
	}
	metrics := make(chan prometheus.Metric)
	got := ""
	go func() {
		for {
			var metric dto.Metric
			val := <-metrics
			err := val.Write(&metric)
			if err != nil {
				t.Errorf("Unexpected error %s", err)
			}
			got += metric.String()
		}
	}()
	err = collector.Collect(scrapeContext, metrics)
	if err != nil {
		t.Errorf("Unexpected error %s", err)
	}
	for _, f := range []string{"dir1", "dir2", "dir3", "dir3sub"} {
		if !strings.Contains(got, f) {
			t.Errorf("Unexpected output %s: %q", f, got)
		}
	}
}

func TestDuplicateFileName(t *testing.T) {
	testDir := baseDir + "/duplicate-filename"
	collector := &textFileCollector{
		logger:      log.NewLogfmtLogger(os.Stdout),
		directories: testDir,
	}
	scrapeContext, err := PrepareScrapeContext([]string{"textfile_test"})
	if err != nil {
		t.Errorf("Unexpected error %s", err)
	}
	metrics := make(chan prometheus.Metric)
	got := ""
	go func() {
		for {
			var metric dto.Metric
			val := <-metrics
			err := val.Write(&metric)
			if err != nil {
				t.Errorf("Unexpected error %s", err)
			}
			got += metric.String()
		}
	}()
	err = collector.Collect(scrapeContext, metrics)
	if err != nil {
		t.Errorf("Unexpected error %s", err)
	}
	if !strings.Contains(got, "file") {
		t.Errorf("Unexpected output  %q", got)
	}
	if strings.Contains(got, "sub_file") {
		t.Errorf("Unexpected output  %q", got)
	}
}
