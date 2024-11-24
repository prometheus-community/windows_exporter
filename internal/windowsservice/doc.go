// Copyright 2024 The Prometheus Authors
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

//go:build windows

// Package windowsservice allows initiating time-sensitive components like registering the Windows service
// as early as possible in the startup process.
// init functions are called in the order they are declared, so this package should be imported first.
// Declare imports on this package should be avoided where possible.
//
// Ref: https://github.com/prometheus-community/windows_exporter/issues/551#issuecomment-1220774835

package windowsservice
