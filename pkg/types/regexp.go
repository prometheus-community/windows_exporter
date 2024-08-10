package types

import "regexp"

var RegExpEmpty = regexp.MustCompile("^$")
var RegExpAny = regexp.MustCompile(".+")
