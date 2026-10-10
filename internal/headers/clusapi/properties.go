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

package clusapi

import (
	"encoding/binary"
	"errors"
	"fmt"
	"unicode/utf16"
)

const (
	propertyName         = 0x00040003
	propertyValue        = 0x00010000
	formatDWORD          = 2
	formatString         = 3
	formatExpandString   = 4
	formatLong           = 7
	formatExpandedString = 8
)

// Property is a single CLUSPROP_LIST_VALUE. Values retain their native format;
// unsupported formats can be skipped without confusing an absent DWORD with zero.
type Property struct {
	Format uint16
	Data   []byte
}

func (p Property) DWORD() (uint32, error) {
	if p.Format != formatDWORD || len(p.Data) != 4 {
		return 0, fmt.Errorf("expected DWORD, format %d, length %d", p.Format, len(p.Data))
	}

	return binary.LittleEndian.Uint32(p.Data), nil
}

// String decodes an SZ, EXPAND_SZ or EXPANDED_SZ value like WMI does: the value
// ends at the first NUL and invalid UTF-16 is replaced with U+FFFD.
func (p Property) String() (string, error) {
	switch p.Format {
	case formatString, formatExpandString, formatExpandedString:
		return decodeString(p.Data), nil
	default:
		return "", fmt.Errorf("expected string, format %d", p.Format)
	}
}

// Value32 returns the bit pattern of a DWORD or LONG value. Callers decide the
// signedness, matching the WMI type of the property they publish.
func (p Property) Value32() (uint32, error) {
	if (p.Format != formatDWORD && p.Format != formatLong) || len(p.Data) != 4 {
		return 0, fmt.Errorf("expected DWORD or LONG, format %d, length %d", p.Format, len(p.Data))
	}

	return binary.LittleEndian.Uint32(p.Data), nil
}

// ParseProperties decodes CLUSPROP_LIST without casting untrusted bytes to native
// structs. Headers and payloads are DWORD aligned; cbLength is measured in bytes.
// Only the structure is validated: string contents are decoded when used, so one
// unusual value does not discard the whole list.
// https://learn.microsoft.com/en-us/previous-versions/windows/desktop/mscs/property-lists
func ParseProperties(data []byte) (map[string][]Property, error) {
	if len(data) < 4 {
		return nil, errors.New("property list is missing its count")
	}

	count := binary.LittleEndian.Uint32(data)

	data = data[4:]
	if uint64(count) > uint64(len(data))/16 {
		return nil, errors.New("property count exceeds buffer")
	}

	properties := make(map[string][]Property, count)

	for range count {
		if len(data) < 8 || binary.LittleEndian.Uint32(data) != propertyName {
			return nil, errors.New("invalid property name header")
		}

		nameBytes, rest, err := propertyPayload(data)
		if err != nil {
			return nil, err
		}

		name := decodeString(nameBytes)
		data = rest
		values := make([]Property, 0, 1)

		for {
			if len(data) < 4 {
				return nil, fmt.Errorf("missing endmark for %q", name)
			}

			syntax := binary.LittleEndian.Uint32(data)
			if syntax == 0 {
				data = data[4:]

				break
			}

			if syntax&0xffff0000 != propertyValue {
				return nil, fmt.Errorf("invalid value syntax %#x for %q", syntax, name)
			}

			payload, rest, err := propertyPayload(data)
			if err != nil {
				return nil, fmt.Errorf("property %q: %w", name, err)
			}

			format := uint16(syntax)
			if (format == formatDWORD || format == formatLong) && len(payload) != 4 {
				return nil, fmt.Errorf("invalid DWORD length for %q", name)
			}

			values = append(values, Property{Format: format, Data: payload})
			data = rest
		}

		// Keep the first occurrence; a repeated name does not affect the structure.
		if _, exists := properties[name]; !exists {
			properties[name] = values
		}
	}

	// Lists returned by the cluster service end with a CLUSPROP_SYNTAX_ENDMARK
	// after the last value list. Accept zero padding, reject anything else.
	for _, b := range data {
		if b != 0 {
			return nil, fmt.Errorf("%d bytes of trailing property list data starting with % x", len(data), data[:min(len(data), 16)])
		}
	}

	return properties, nil
}

func propertyPayload(data []byte) ([]byte, []byte, error) {
	if len(data) < 8 {
		return nil, nil, errors.New("truncated property header")
	}

	length := uint64(binary.LittleEndian.Uint32(data[4:]))

	aligned := (length + 3) &^ uint64(3)
	if aligned > uint64(len(data)-8) {
		return nil, nil, errors.New("property payload exceeds buffer")
	}

	return data[8 : 8+length], data[8+aligned:], nil
}

// decodeString decodes little-endian UTF-16 bytes. A trailing odd byte is ignored.
func decodeString(data []byte) string {
	units := make([]uint16, len(data)/2)
	for i := range units {
		units[i] = binary.LittleEndian.Uint16(data[i*2:])
	}

	return decodeUnits(units)
}

// decodeUnits ends the string at the first NUL and replaces invalid UTF-16 with
// U+FFFD, matching the strings WMI returned for the same names.
func decodeUnits(units []uint16) string {
	for i, unit := range units {
		if unit == 0 {
			units = units[:i]

			break
		}
	}

	return string(utf16.Decode(units))
}
