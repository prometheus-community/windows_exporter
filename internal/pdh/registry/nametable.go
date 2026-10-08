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

package registry

import (
	"bytes"
	"fmt"
	"strconv"
	"sync"
)

// CounterNameTable Initialize global name tables
// profiling, add option to disable name tables if necessary
// Not sure if we should resolve the names at all or just have the caller do it on demand
// (for many use cases the index is sufficient)
//
//nolint:gochecknoglobals
var CounterNameTable = QueryNameTable("Counter 009")

func (p *perfObjectType) LookupName() string {
	return CounterNameTable.LookupString(p.ObjectNameTitleIndex)
}

type NameTable struct {
	mu sync.Mutex

	name string

	table struct {
		index  map[uint32]string
		string map[string]uint32
	}
}

// LookupString returns the name for index. It returns an empty string if the
// index is unknown or the name table could not be loaded.
func (t *NameTable) LookupString(index uint32) string {
	if t.load() != nil {
		return ""
	}

	return t.table.index[index]
}

// LookupIndex returns the index for str. It returns 0 if the name is unknown
// or the name table could not be loaded.
func (t *NameTable) LookupIndex(str string) uint32 {
	if t.load() != nil {
		return 0
	}

	return t.table.string[str]
}

// QueryNameTable Query a perflib name table from the v1. Specify the type and the language
// code (i.e. "Counter 009" or "Help 009") for English language.
func QueryNameTable(tableName string) *NameTable {
	return &NameTable{
		name: tableName,
	}
}

// load reads the name table from the registry. A failed load is retried on the
// next call, so a transient error does not disable the table for good. The
// maps are not modified once loaded.
func (t *NameTable) load() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.table.index != nil {
		return nil
	}

	buffer, err := queryRawData(t.name)
	if err != nil {
		return fmt.Errorf("failed to query perflib name table %q: %w", t.name, err)
	}

	t.table.index, t.table.string = parseNameTable(buffer)

	return nil
}

// parseNameTable parses a name table, a REG_MULTI_SZ of alternating index and
// name strings. Parsing stops at the first incomplete pair.
func parseNameTable(buffer []byte) (map[uint32]string, map[string]uint32) {
	index := make(map[uint32]string)
	names := make(map[string]uint32)

	r := bytes.NewReader(buffer)

	for {
		indexStr, err := readUTF16String(r)
		if err != nil {
			break
		}

		desc, err := readUTF16String(r)
		if err != nil {
			break
		}

		indexInt, _ := strconv.ParseUint(indexStr, 10, 32)

		index[uint32(indexInt)] = desc
		names[desc] = uint32(indexInt)
	}

	return index, names
}
