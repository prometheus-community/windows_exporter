package perflib

import (
	"bytes"
	"fmt"
	"strconv"
)

// Initialize global name tables
// TODO: profiling, add option to disable name tables if necessary
// Not sure if we should resolve the names at all or just have the caller do it on demand
// (for many use cases the index is sufficient)

var CounterNameTable = *QueryNameTable("Counter 009")
var HelpNameTable = *QueryNameTable("Help 009")

func (p *perfObjectType) LookupName() string {
	return CounterNameTable.LookupString(p.ObjectNameTitleIndex)
}

func (p *perfObjectType) LookupHelp() string {
	return HelpNameTable.LookupString(p.ObjectHelpTitleIndex)
}

type NameTable struct {
	byIndex  map[uint32]string
	byString map[string]uint32
}

func (t *NameTable) LookupString(index uint32) string {
	return t.byIndex[index]
}

func (t *NameTable) LookupIndex(str string) uint32 {
	return t.byString[str]
}

// QueryNameTable Query a perflib name table from the registry. Specify the type and the language
// code (i.e. "Counter 009" or "Help 009") for English language.
func QueryNameTable(tableName string) *NameTable {
	nameTable := new(NameTable)
	nameTable.byIndex = make(map[uint32]string)

	buffer, err := queryRawData(tableName)
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

		if err != nil {
			panic(fmt.Sprint("Invalid index ", index))
		}

		nameTable.byIndex[uint32(indexInt)] = desc
	}

	nameTable.byString = make(map[string]uint32)

	for k, v := range nameTable.byIndex {
		nameTable.byString[v] = k
	}

	return nameTable
}
