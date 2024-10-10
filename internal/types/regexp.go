package types

import "regexp"

var (
	RegExpAny   = regexp.MustCompile("^.+$")
	RegExpEmpty = regexp.MustCompile("^$")
)
