//go:build windows

package textfile_test

import (
	"fmt"
	"io"
	"log/slog"
	"strings"
	"testing"

	"github.com/prometheus-community/windows_exporter/internal/collector/textfile"
	"github.com/prometheus-community/windows_exporter/pkg/collector"
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"github.com/stretchr/testify/require"
)

var baseDir = "../../../tools/textfile-test"

//nolint:paralleltest
func TestMultipleDirectories(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	testDir := baseDir + "/multiple-dirs"
	testDirs := fmt.Sprintf("%[1]s/dir1,%[1]s/dir2,%[1]s/dir3", testDir)

	textFileCollector := textfile.New(&textfile.Config{
		TextFileDirectories: strings.Split(testDirs, ","),
	})

	collectors := collector.New(map[string]collector.Collector{textfile.Name: textFileCollector})
	require.NoError(t, collectors.Build(logger))

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

	err := textFileCollector.Collect(metrics)
	if err != nil {
		t.Errorf("Unexpected error %s", err)
	}

	for _, f := range []string{"dir1", "dir2", "dir3", "dir3sub"} {
		if !strings.Contains(got, f) {
			t.Errorf("Unexpected output %s: %q", f, got)
		}
	}
}

//nolint:paralleltest
func TestDuplicateFileName(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	testDir := baseDir + "/duplicate-filename"
	textFileCollector := textfile.New(&textfile.Config{
		TextFileDirectories: []string{testDir},
	})

	collectors := collector.New(map[string]collector.Collector{textfile.Name: textFileCollector})
	require.NoError(t, collectors.Build(logger))

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

	err := textFileCollector.Collect(metrics)
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
