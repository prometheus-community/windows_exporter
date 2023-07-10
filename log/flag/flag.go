// Copyright 2017 The Prometheus Authors
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package flag

import (
	"github.com/alecthomas/kingpin/v2"
	"github.com/prometheus-community/windows_exporter/log"
	"github.com/prometheus/common/promlog"
	promlogflag "github.com/prometheus/common/promlog/flag"
)

// FileFlagName is the canonical flag name to configure the log file
const FileFlagName = "log.file"

// FileFlagHelp is the help description for the log.file flag.
const FileFlagHelp = "Output file of log messages. One of [stdout, stderr, eventlog, <path to log file>]"

// AddFlags adds the flags used by this package to the Kingpin application.
// To use the default Kingpin application, call AddFlags(kingpin.CommandLine)
func AddFlags(a *kingpin.Application, config *log.Config) {
	config.Level = &promlog.AllowedLevel{}
	a.Flag(promlogflag.LevelFlagName, promlogflag.LevelFlagHelp).
		Default("info").SetValue(config.Level)

	config.File = &log.AllowedFile{}
	a.Flag(FileFlagName, FileFlagHelp).
		Default("stderr").SetValue(config.File)

	config.Format = &promlog.AllowedFormat{}
	a.Flag(promlogflag.FormatFlagName, promlogflag.FormatFlagHelp).
		Default("logfmt").SetValue(config.Format)
}
