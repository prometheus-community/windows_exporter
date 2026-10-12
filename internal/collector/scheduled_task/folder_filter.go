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

package scheduled_task

import (
	"regexp"
	"regexp/syntax"
	"strings"
)

// folderFilter skips task folders whose tasks the include and exclude filters
// reject whatever their names are. Listing the tasks and subfolders of a folder
// costs two round trips to the Task Scheduler service, so a filter such as
// exclude="/Microsoft/.+" saves the walk of most folders of a Windows install.
//
// The analysis is conservative: it skips a folder only when the expression
// proves that no task below it passes, and keeps every folder otherwise. Task
// paths cannot contain line breaks, because tasks are stored as files below
// %SystemRoot%\System32\Tasks, so "." matches every character of a path.
type folderFilter struct {
	// include and exclude are nil when the expression gives no folder information.
	include *syntax.Regexp
	exclude *syntax.Regexp
	// override is nil when there is no include override.
	override *syntax.Regexp
}

func newFolderFilter(include, exclude, override *regexp.Regexp) folderFilter {
	f := folderFilter{include: parseFilter(include), exclude: parseFilter(exclude)}

	if override != nil {
		f.override = parseFilter(override)
		if f.override == nil {
			// Any folder may hold a task that passes an override without folder information.
			return folderFilter{}
		}
	}

	return f
}

func parseFilter(re *regexp.Regexp) *syntax.Regexp {
	if re == nil {
		return nil
	}

	// regexp.Compile parses with syntax.Perl.
	parsed, err := syntax.Parse(re.String(), syntax.Perl)
	if err != nil {
		return nil
	}

	return parsed.Simplify()
}

// includeFolder reports whether a task below the folder with the given
// slash-separated path may pass both filters.
func (f folderFilter) includeFolder(path string) bool {
	prefix := strings.TrimSuffix(path, "/") + "/"

	// Tasks that pass the override are included whatever include and exclude are.
	if f.override != nil && mayMatchPrefixed([]*syntax.Regexp{f.override}, prefix, false) {
		return true
	}

	if f.include != nil && !mayMatchPrefixed([]*syntax.Regexp{f.include}, prefix, false) {
		return false
	}

	return f.exclude == nil || !matchesAllPrefixed([]*syntax.Regexp{f.exclude}, prefix, true)
}

// mayMatchPrefixed reports false only if no string that starts with prefix
// matches the sequence of expressions. Without a start anchor a match can begin
// anywhere, so only anchored expressions rule out a prefix.
func mayMatchPrefixed(seq []*syntax.Regexp, prefix string, anchored bool) bool {
	for i, re := range seq {
		if prefix == "" {
			return true
		}

		switch re.Op {
		case syntax.OpBeginText, syntax.OpBeginLine:
			if i == 0 && !anchored {
				anchored = true

				continue
			}

			return true
		case syntax.OpEmptyMatch:
			continue
		case syntax.OpCapture, syntax.OpConcat:
			return mayMatchPrefixed(splice(re.Sub, seq[i+1:]), prefix, anchored)
		case syntax.OpAlternate:
			for _, sub := range re.Sub {
				if mayMatchPrefixed(splice([]*syntax.Regexp{sub}, seq[i+1:]), prefix, anchored) {
					return true
				}
			}

			return false
		case syntax.OpLiteral:
			if !anchored || re.Flags&syntax.FoldCase != 0 {
				return true
			}

			literal := string(re.Rune)
			if !strings.HasPrefix(prefix, literal) {
				return strings.HasPrefix(literal, prefix)
			}

			prefix = prefix[len(literal):]
		case syntax.OpEndText, syntax.OpEndLine:
			// Task paths continue after the folder prefix and have no line breaks.
			return !anchored
		default:
			return true
		}
	}

	return true
}

// matchesAllPrefixed reports true only if every string that starts with prefix
// and continues with at least one character matches the sequence of expressions.
// A full match implies an unanchored match, so start anchors are not required.
func matchesAllPrefixed(seq []*syntax.Regexp, prefix string, atStart bool) bool {
	for i, re := range seq {
		switch re.Op {
		case syntax.OpBeginText, syntax.OpBeginLine:
			if atStart {
				continue
			}

			return false
		case syntax.OpCapture, syntax.OpConcat:
			return matchesAllPrefixed(splice(re.Sub, seq[i+1:]), prefix, atStart)
		case syntax.OpAlternate:
			for _, sub := range re.Sub {
				if matchesAllPrefixed(splice([]*syntax.Regexp{sub}, seq[i+1:]), prefix, atStart) {
					return true
				}
			}

			return false
		case syntax.OpLiteral:
			literal := string(re.Rune)
			if re.Flags&syntax.FoldCase != 0 || !strings.HasPrefix(prefix, literal) {
				return false
			}

			prefix = prefix[len(literal):]
			atStart = atStart && literal == ""
		case syntax.OpStar, syntax.OpPlus:
			// ".*" or ".+" as the last expression matches the rest of the path,
			// which is never empty.
			op := re.Sub[0].Op

			return (op == syntax.OpAnyChar || op == syntax.OpAnyCharNotNL) && onlyEndAnchors(seq[i+1:])
		default:
			return false
		}
	}

	return false
}

func onlyEndAnchors(seq []*syntax.Regexp) bool {
	for _, re := range seq {
		if re.Op != syntax.OpEndText && re.Op != syntax.OpEndLine {
			return false
		}
	}

	return true
}

func splice(head, tail []*syntax.Regexp) []*syntax.Regexp {
	return append(append(make([]*syntax.Regexp, 0, len(head)+len(tail)), head...), tail...)
}
