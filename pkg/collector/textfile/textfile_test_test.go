package textfile_test

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/prometheus-community/windows_exporter/pkg/collector"
	"github.com/prometheus-community/windows_exporter/pkg/collector/textfile"
	"github.com/prometheus-community/windows_exporter/pkg/types"

	"github.com/go-kit/log"
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
)

var baseDir = "../../../tools/textfile-test"

func TestMultipleDirectories(t *testing.T) {
	testDir := baseDir + "/multiple-dirs"
	testDirs := fmt.Sprintf("%[1]s/dir1,%[1]s/dir2,%[1]s/dir3", testDir)

	textfileCollector := textfile.New(log.NewLogfmtLogger(os.Stdout), &textfile.Config{
		TextFileDirectories: testDirs,
	})

	collectors := collector.New(map[string]types.Collector{textfile.Name: textfileCollector})
	require.NoError(t, collectors.Build())

	scrapeContext, err := collectors.PrepareScrapeContext()
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

	err = textfileCollector.Collect(scrapeContext, metrics)
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
	textfileCollector := textfile.New(log.NewLogfmtLogger(os.Stdout), &textfile.Config{
		TextFileDirectories: testDir,
	})

	collectors := collector.New(map[string]types.Collector{textfile.Name: textfileCollector})
	require.NoError(t, collectors.Build())

	scrapeContext, err := collectors.PrepareScrapeContext()
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
	err = textfileCollector.Collect(scrapeContext, metrics)
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
