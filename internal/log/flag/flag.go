// SPDX-License-Identifier: MIT
//
// Copyright The Prometheus Authors

//go:build windows

package flag

import (
	"github.com/alecthomas/kingpin/v2"
	"github.com/prometheus-community/windows_exporter/internal/log"
	"github.com/prometheus/common/promslog"
	"github.com/prometheus/common/promslog/flag"
)

// FileFlagName is the canonical flag name to configure the log file.
const FileFlagName = "log.file"

// FileFlagHelp is the help description for the log.file flag.
const FileFlagHelp = "Output file of log messages. One of [stdout, stderr, eventlog, <path to log file>]"

// AddFlags adds the flags used by this package to the Kingpin application.
// To use the default Kingpin application, call AddFlags(kingpin.CommandLine).
func AddFlags(a *kingpin.Application, config *log.Config) {
	config.Config = new(promslog.Config)
	flag.AddFlags(a, config.Config)

	if config.File == nil {
		config.File = &log.AllowedFile{}
	}

	a.Flag(FileFlagName, FileFlagHelp).Default(config.File.String()).SetValue(config.File)

	a.Flag(
		"log.rotation.max-size",
		"Size in megabytes at which the log file is rotated. 0 disables rotation. Requires log.file to be a file path.",
	).Default("0").IntVar(&config.Rotation.MaxSize)
	a.Flag(
		"log.rotation.max-backups",
		"Number of rotated log files to keep. 0 keeps all.",
	).Default("0").IntVar(&config.Rotation.MaxBackups)
	a.Flag(
		"log.rotation.compress",
		"Compress rotated log files with gzip.",
	).Default("false").BoolVar(&config.Rotation.Compress)
}
