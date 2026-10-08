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

package textfile

import (
	"bufio"
	"bytes"
	"errors"
	"io"
)

// skipUTF8BOM removes one leading UTF-8 BOM and rejects UTF-16/32 BOMs.
// The buffered reader is also reused by the metrics parser.
func skipUTF8BOM(r *bufio.Reader) error {
	prefix, err := r.Peek(4)
	if err != nil && !errors.Is(err, io.EOF) {
		return err
	}

	switch {
	case bytes.HasPrefix(prefix, []byte("\x00\x00\xfe\xff")):
		return errors.New("UTF32BigEndian")
	case bytes.HasPrefix(prefix, []byte("\xff\xfe\x00\x00")):
		return errors.New("UTF32LittleEndian")
	case bytes.HasPrefix(prefix, []byte("\xfe\xff")):
		return errors.New("UTF16BigEndian")
	case bytes.HasPrefix(prefix, []byte("\xff\xfe")):
		return errors.New("UTF16LittleEndian")
	case bytes.HasPrefix(prefix, []byte("\xef\xbb\xbf")):
		_, err = r.Discard(3)

		return err
	default:
		return nil
	}
}
