package main

import (
	"fmt"
	"os"
	"sort"

	"github.com/go-kit/log/level"
	"github.com/prometheus-community/windows_exporter/pkg/collector"
	"github.com/prometheus-community/windows_exporter/pkg/exporter"
	"github.com/prometheus-community/windows_exporter/pkg/windows_service"
	"golang.org/x/sys/windows/svc"
)

func main() {
	exporter := exporter.New()
	if exporter.PrintCollectors() {
		collectorNames := collector.Available()
		sort.Strings(collectorNames)
		fmt.Printf("Available collectors:\n")
		for _, n := range collectorNames {
			fmt.Printf(" - %s\n", n)
		}
		os.Exit(0)
	}
	isWinService, err := svc.IsWindowsService()
	if err != nil {
		_ = level.Error(exporter.GetLogger()).Log("Failed to detect [IsWindowsService]: ", "err", err)
		os.Exit(1)
	}
	if isWinService {
		windows_service.Run(exporter)
	} else {
		exporter.RunAsCli()
	}
}
