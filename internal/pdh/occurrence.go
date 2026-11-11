// SPDX-License-Identifier: MIT
//
// Copyright The Prometheus Authors

//go:build windows

package pdh

import "strconv"

// SuffixDuplicates makes the names of rows unique within a sample for
// collectors that publish every occurrence of a duplicate instance name.
//
// name returns a pointer to the name of a row, occurrence its Occurrence (see
// [NewCollector]). The first occurrence keeps its name. Later occurrences get
// the suffix "#N", where N counts from 1 and skips every N for which
// "<name>#N" is the name of another row, for example a suffix that Windows
// added itself. Only the names of rows reserve suffixes: an instance that
// Windows reported without valid data has no row and reserves nothing. N
// depends on the order of the sample and is not a stable identity.
//
// Collectors can map names, e.g. an empty name to a placeholder, before they
// call SuffixDuplicates, so that the suffix is added to the published name.
func SuffixDuplicates[T any](rows []T, name func(*T) *string, occurrence func(*T) int) {
	var names map[string]struct{}

	for i := range rows {
		if occurrence(&rows[i]) == 0 {
			continue
		}

		if names == nil {
			names = make(map[string]struct{}, len(rows))

			for j := range rows {
				names[*name(&rows[j])] = struct{}{}
			}
		}

		instanceName := name(&rows[i])
		base := *instanceName
		remaining := occurrence(&rows[i])

		for suffix := 1; ; suffix++ {
			candidate := base + "#" + strconv.Itoa(suffix)
			if _, literal := names[candidate]; literal {
				continue
			}

			remaining--
			if remaining == 0 {
				*instanceName = candidate

				break
			}
		}
	}
}
