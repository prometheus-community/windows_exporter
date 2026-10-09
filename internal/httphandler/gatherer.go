// SPDX-License-Identifier: Apache-2.0
//
// Copyright The Prometheus Authors
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

package httphandler

import (
	"errors"
	"fmt"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
)

// Interface guard.
var _ prometheus.Gatherer = mergedGatherer{}

// mergedGatherer combines the results of two registries.
//
// prometheus.Gatherers re-hashes and re-checks every metric that the registries
// already checked in their own Gather. mergedGatherer merges the sorted results
// without the second check, as long as no metric family name of one registry
// collides with the other. If a name collides, it hands both results to
// prometheus.Gatherers, so the output and errors stay the same as before.
// See https://github.com/prometheus/client_golang/pull/1702.
type mergedGatherer struct {
	first  prometheus.Gatherer
	second prometheus.Gatherer
}

func (g mergedGatherer) Gather() ([]*dto.MetricFamily, error) {
	firstMFs, firstErr := g.first.Gather()
	secondMFs, secondErr := g.second.Gather()

	if namesCollide(firstMFs, secondMFs) {
		return prometheus.Gatherers{
			gathered(firstMFs, firstErr),
			gathered(secondMFs, secondErr),
		}.Gather()
	}

	var errs prometheus.MultiError

	errs = appendGathererErrors(errs, 1, firstErr)
	errs = appendGathererErrors(errs, 2, secondErr)

	return mergeSorted(firstMFs, secondMFs), errs.MaybeUnwrap()
}

// gathered returns a Gatherer that returns an earlier Gather result.
func gathered(mfs []*dto.MetricFamily, err error) prometheus.GathererFunc {
	return prometheus.GathererFunc(func() ([]*dto.MetricFamily, error) {
		return mfs, err
	})
}

// appendGathererErrors adds err to errs like prometheus.Gatherers does.
func appendGathererErrors(errs prometheus.MultiError, gatherer int, err error) prometheus.MultiError {
	if err == nil {
		return errs
	}

	var multiErr prometheus.MultiError
	if !errors.As(err, &multiErr) {
		return append(errs, fmt.Errorf("[from Gatherer #%d] %w", gatherer, err))
	}

	for _, err := range multiErr {
		errs = append(errs, fmt.Errorf("[from Gatherer #%d] %w", gatherer, err))
	}

	return errs
}

// namesCollide reports whether a metric family name of a can clash with one of b.
// Besides equal names, it counts a name that equals the other name plus a
// summary or histogram suffix, regardless of the types. A false positive only
// means the scrape takes the slower prometheus.Gatherers path.
func namesCollide(a, b []*dto.MetricFamily) bool {
	if len(a) == 0 || len(b) == 0 {
		return false
	}

	// Index the smaller result. The exporter metrics registry has only a few families.
	if len(a) > len(b) {
		a, b = b, a
	}

	// names holds the names of a, bases the names of a without a summary or histogram suffix.
	names := make(map[string]struct{}, len(a))
	bases := make(map[string]struct{})

	for _, mf := range a {
		names[mf.GetName()] = struct{}{}

		if base, ok := trimHistogramSuffix(mf.GetName()); ok {
			bases[base] = struct{}{}
		}
	}

	for _, mf := range b {
		name := mf.GetName()

		if _, ok := names[name]; ok {
			return true
		}

		// A name of a is the name plus a suffix.
		if _, ok := bases[name]; ok {
			return true
		}

		// The name is a name of a plus a suffix.
		if base, ok := trimHistogramSuffix(name); ok {
			if _, ok := names[base]; ok {
				return true
			}
		}
	}

	return false
}

// trimHistogramSuffix removes the suffixes the text formats add to summaries and histograms.
// They are the suffixes client_golang checks for collisions.
func trimHistogramSuffix(name string) (string, bool) {
	for _, suffix := range [...]string{"_count", "_sum", "_bucket"} {
		if base, ok := strings.CutSuffix(name, suffix); ok {
			return base, true
		}
	}

	return name, false
}

// mergeSorted merges two metric family slices that are sorted by name and share no name.
func mergeSorted(a, b []*dto.MetricFamily) []*dto.MetricFamily {
	result := make([]*dto.MetricFamily, 0, len(a)+len(b))

	for len(a) > 0 && len(b) > 0 {
		if a[0].GetName() < b[0].GetName() {
			result = append(result, a[0])
			a = a[1:]
		} else {
			result = append(result, b[0])
			b = b[1:]
		}
	}

	result = append(result, a...)

	return append(result, b...)
}
