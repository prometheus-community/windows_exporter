// SPDX-License-Identifier: MIT
//
// Copyright The Prometheus Authors

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
