package perflib

import (
	"bytes"
	"strconv"
	"fmt"
)

type nameTableLookuper interface {
	LookupName() string
	LookupHelp() string
}

func (p *perfObjectType) LookupName() string {
	return counterNameTable.LookupString(p.ObjectNameTitleIndex)
}

func (p *perfObjectType) LookupHelp() string {
	return helpNameTable.LookupString(p.ObjectHelpTitleIndex)
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

// Query a perflib name table from the registry. Specify the type and the language
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
