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
	"errors"
	"io"
	"strings"
	"testing"
	"testing/iotest"

	"github.com/prometheus/common/expfmt"
	"github.com/prometheus/common/model"
)

func TestSkipUTF8BOM(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name  string
		input string
		want  string
		err   string
	}{
		{name: "empty"},
		{name: "one byte", input: "a", want: "a"},
		{name: "two bytes", input: "ab", want: "ab"},
		{name: "three bytes", input: "abc", want: "abc"},
		{name: "plain metrics", input: "metric 1\n", want: "metric 1\n"},
		{name: "UTF-8", input: "\xef\xbb\xbfmetric 1\n", want: "metric 1\n"},
		{name: "UTF-8 BOM only", input: "\xef\xbb\xbf"},
		{name: "one partial BOM byte", input: "\xef", want: "\xef"},
		{name: "two partial BOM bytes", input: "\xef\xbb", want: "\xef\xbb"},
		{name: "mismatched BOM", input: "\xef\xbbx", want: "\xef\xbbx"},
		{name: "embedded BOM", input: "a\xef\xbb\xbf", want: "a\xef\xbb\xbf"},
		{name: "repeated BOM", input: "\xef\xbb\xbf\xef\xbb\xbf", want: "\xef\xbb\xbf"},
		{name: "UTF-16 BE", input: "\xfe\xff", err: "UTF16BigEndian"},
		{name: "UTF-16 LE", input: "\xff\xfe", err: "UTF16LittleEndian"},
		{name: "UTF-16 BE with data", input: "\xfe\xff\x00a", err: "UTF16BigEndian"},
		{name: "UTF-16 LE with data", input: "\xff\xfea\x00", err: "UTF16LittleEndian"},
		{name: "UTF-32 BE", input: "\x00\x00\xfe\xff", err: "UTF32BigEndian"},
		{name: "UTF-32 LE", input: "\xff\xfe\x00\x00", err: "UTF32LittleEndian"},
		{name: "partial UTF-32 BE", input: "\x00\x00\xfe", want: "\x00\x00\xfe"},
		{name: "partial UTF-32 LE", input: "\xff\xfe\x00", err: "UTF16LittleEndian"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			for _, fragmented := range []bool{false, true} {
				var input io.Reader = strings.NewReader(tc.input)
				if fragmented {
					input = iotest.OneByteReader(input)
				}

				r := bufio.NewReader(input)

				err := skipUTF8BOM(r)
				if tc.err != "" {
					if err == nil || err.Error() != tc.err {
						t.Fatalf("fragmented=%t: got error %v, want %s", fragmented, err, tc.err)
					}

					continue
				}

				if err != nil {
					t.Fatal(err)
				}

				got, err := io.ReadAll(r)
				if err != nil {
					t.Fatal(err)
				}

				if string(got) != tc.want {
					t.Errorf("fragmented=%t: got %q, want %q", fragmented, got, tc.want)
				}
			}
		})
	}
}

func TestSkipUTF8BOMReadError(t *testing.T) {
	t.Parallel()

	readErr := errors.New("read failed")

	for _, tc := range []struct {
		name  string
		input string
		want  string
	}{
		{name: "before prefix"},
		{name: "partial prefix", input: "\xef\xbb"},
		{name: "after plain prefix", input: "metric 1\n", want: "metric 1\n"},
		{name: "after UTF-8 prefix", input: "\xef\xbb\xbfmetric 1\n", want: "metric 1\n"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			input := io.MultiReader(strings.NewReader(tc.input), iotest.ErrReader(readErr))
			r := bufio.NewReader(input)

			err := skipUTF8BOM(r)
			if len(tc.input) < 4 {
				if !errors.Is(err, readErr) {
					t.Fatalf("got error %v, want %v", err, readErr)
				}

				return
			}

			if err != nil {
				t.Fatal(err)
			}

			got, err := io.ReadAll(r)
			if !errors.Is(err, readErr) {
				t.Errorf("got error %v, want %v", err, readErr)
			}

			if string(got) != tc.want {
				t.Errorf("got %q, want %q", got, tc.want)
			}
		})
	}
}

func BenchmarkSkipUTF8BOM(b *testing.B) {
	for _, tc := range []struct {
		name  string
		input string
	}{
		{name: "plain", input: "metric 1\n"},
		{name: "UTF-8 BOM", input: "\xef\xbb\xbfmetric 1\n"},
	} {
		b.Run(tc.name, func(b *testing.B) {
			input := strings.NewReader(tc.input)
			r := bufio.NewReader(input)

			b.ReportAllocs()

			for b.Loop() {
				input.Reset(tc.input)
				r.Reset(input)

				if err := skipUTF8BOM(r); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func BenchmarkBOMParser(b *testing.B) {
	for _, tc := range []struct {
		name   string
		prefix string
	}{
		{name: "plain"},
		{name: "UTF-8 BOM", prefix: "\xef\xbb\xbf"},
	} {
		b.Run(tc.name, func(b *testing.B) {
			input := tc.prefix + "# HELP metric Example metric.\n# TYPE metric gauge\nmetric{label=\"value\"} 1\n"

			b.ReportAllocs()

			for b.Loop() {
				r := bufio.NewReader(strings.NewReader(input))
				if err := skipUTF8BOM(r); err != nil {
					b.Fatal(err)
				}

				parser := expfmt.NewTextParser(model.UTF8Validation)
				if _, err := parser.TextToMetricFamilies(r); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
