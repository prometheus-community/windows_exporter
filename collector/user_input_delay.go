//go:build windows
// +build windows

package collector

import (
	"errors"
	"fmt"
	"regexp"

	"github.com/prometheus-community/windows_exporter/log"
	"github.com/prometheus/client_golang/prometheus"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	userInputDelayWhitelist = kingpin.Flag(
		"collector.user_input_delay.whitelist",
		"Regexp of processes to include. Process name must both match whitelist and not match blacklist to be included.",
	).Default(".*").String()
	userInputDelayBlacklist = kingpin.Flag(
		"collector.user_input_delay.blacklist",
		"Regexp of processes to exclude. Process name must both match whitelist and not match blacklist to be included.",
	).Default("").String()
)

func init() {
	registerCollector("user_input_delay", NewUserInputCollector, "User Input Delay per Process", "User Input Delay per Session")
}

type UserInputCollector struct {
	SessionMaxInputDelay *prometheus.Desc
	ProcessMaxInputDelay *prometheus.Desc

	inputWhitelistPattern *regexp.Regexp
	inputBlacklistPattern *regexp.Regexp
}

func NewUserInputCollector() (Collector, error) {
	const subsystem = "user_input_delay"

	if *userInputDelayWhitelist == ".*" && *userInputDelayBlacklist == "" {
		log.Warn("No filters specified for user_input_delay collector. This will generate a very large number of metrics!")
	}

	return &UserInputCollector{
		SessionMaxInputDelay: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "session"),
			"Maximum value for queuing delay across all user input waiting to be picked-up by any process in the session during a target time interval",
			[]string{"session_id"},
			nil,
		),
		ProcessMaxInputDelay: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "process"),
			"Maximum value for queuing delay across all user input waiting to be picked-up by the process during a target time interval",
			[]string{"session_id", "process_id", "process_name"},
			nil,
		),
		inputWhitelistPattern: regexp.MustCompile(fmt.Sprintf("^(?:%s)$", *userInputDelayWhitelist)),
		inputBlacklistPattern: regexp.MustCompile(fmt.Sprintf("^(?:%s)$", *userInputDelayBlacklist)),
	}, nil
}

// Collect sends the metric values for each metric
// to the provided prometheus Metric channel.
func (c *UserInputCollector) Collect(ctx *ScrapeContext, ch chan<- prometheus.Metric) error {
	if desc, err := c.collect(ctx, ch); err != nil {
		log.Error("failed collecting user_input_delay metrics:", desc, err)
		return err
	}
	return nil
}

// Perflib: "User Input Delay per Session"
type SessionMaxInputDelay struct {
	Name string

	MaxInputDelay float64 `perflib:"MaxInputDelay"`
}

// Perflib: "User Input Delay per Process"
type ProcessMaxInputDelay struct {
	Name string

	MaxInputDelay float64 `perflib:"MaxInputDelay"`
}

type processLabels struct {
	sessionID   string
	processID   string
	processName string
}

// Process name provided by "User Input Delay per Process" ListSet can be split to multiple labels.
// E.G. "1:7232 <TextInputHost.exe>" could be split to:
//  sessionID: 1
//  processID: 7232
//  processName: TextInputHost.exe
func splitProcessLabel(process_label string) (*processLabels, error) {
	// Use regex in place of multiple splits
	query := regexp.MustCompile(`(\d):(\d+)\s+<(.*)>`)
	if !query.MatchString(process_label) {
		return nil, errors.New(fmt.Sprint("Unexpected process label structure: ", process_label))
	}
	result := query.FindStringSubmatch(process_label)

	labels := processLabels{
		sessionID:   result[1],
		processID:   result[2],
		processName: result[3],
	}

	return &labels, nil
}

func (c *UserInputCollector) collect(ctx *ScrapeContext, ch chan<- prometheus.Metric) (*prometheus.Desc, error) {
	var session_dst []SessionMaxInputDelay
	// Note that collector will exit early if Session CounterSet cannot be queried, resulting in the Process Couterset being skipped.
	if err := unmarshalObject(ctx.perfObjects["User Input Delay per Session"], &session_dst); err != nil {
		return nil, err
	}

	var process_dst []ProcessMaxInputDelay
	if err := unmarshalObject(ctx.perfObjects["User Input Delay per Process"], &process_dst); err != nil {
		return nil, err
	}

	for _, session := range session_dst {
		// Skip Average and Max session IDs, these can be computed with Prometheus functions.
		if session.Name == "Average" || session.Name == "Max" {
			continue
		}
		ch <- prometheus.MustNewConstMetric(
			c.SessionMaxInputDelay,
			prometheus.GaugeValue,
			session.MaxInputDelay,
			session.Name,
		)
	}

	for _, process := range process_dst {
		labels, err := splitProcessLabel(process.Name)
		if err != nil {
			return nil, err
		}

		// Filter *after* splitting labels, so only the process name is matched against the regex
		if c.inputBlacklistPattern.MatchString(labels.processName) || !c.inputWhitelistPattern.MatchString(labels.processName) {
			continue
		}

		ch <- prometheus.MustNewConstMetric(
			c.ProcessMaxInputDelay,
			prometheus.GaugeValue,
			process.MaxInputDelay,
			labels.sessionID,
			labels.processID,
			labels.processName,
		)
	}
	return nil, nil
}
