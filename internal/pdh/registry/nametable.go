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
	"strconv"
	"sync"
)

// CounterNameTable Initialize global name tables
// profiling, add option to disable name tables if necessary
// Not sure if we should resolve the names at all or just have the caller do it on demand
// (for many use cases the index is sufficient)
//
//nolint:gochecknoglobals
var CounterNameTable = *QueryNameTable("Counter 009")

func (p *perfObjectType) LookupName() string {
	return CounterNameTable.LookupString(p.ObjectNameTitleIndex)
}

type NameTable struct {
	once sync.Once

	name string

	table struct {
		index  map[uint32]string
		string map[string]uint32
	}
}

func (t *NameTable) LookupString(index uint32) string {
	t.initialize()

	return t.table.index[index]
}

func (t *NameTable) LookupIndex(str string) uint32 {
	t.initialize()

	return t.table.string[str]
}

// QueryNameTable Query a perflib name table from the v1. Specify the type and the language
// code (i.e. "Counter 009" or "Help 009") for English language.
func QueryNameTable(tableName string) *NameTable {
	return &NameTable{
		name: tableName,
	}
}

func (t *NameTable) initialize() {
	t.once.Do(func() {
		t.table.index = make(map[uint32]string)
		t.table.string = make(map[string]uint32)

		buffer, err := queryRawData(t.name)
		if err != nil {
			panic(err)
		}

		r := bytes.NewReader(buffer)

		for {
			index, err := readUTF16String(r)
			if err != nil {
				break
			}

			desc, err := readUTF16String(r)
			if err != nil {
				break
			}

			indexInt, _ := strconv.Atoi(index)

			t.table.index[uint32(indexInt)] = desc
			t.table.string[desc] = uint32(indexInt)
		}
	})
}
