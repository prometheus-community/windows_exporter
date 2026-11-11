// SPDX-License-Identifier: MIT
//
// Copyright The Prometheus Authors

//go:build windows

package utils

import (
	"fmt"
	"regexp"
	"slices"

	"github.com/alecthomas/kingpin/v2"
	"github.com/prometheus-community/windows_exporter/internal/types"
)

//nolint:gochecknoglobals
var (
	// matchAllExpressions are the expressions of an include filter that matches every name.
	matchAllExpressions = []string{anchor(".+"), anchor(".*"), types.RegExpAny.String()}
	// matchNoneExpressions are the expressions of an exclude filter that matches no name.
	matchNoneExpressions = []string{anchor(""), types.RegExpEmpty.String()}
)

// anchor wraps expr, so that it must match the whole input.
func anchor(expr string) string {
	return "^(?:" + expr + ")$"
}

// CompileAnchored compiles expr, so that it must match the whole input.
func CompileAnchored(expr string) (*regexp.Regexp, error) {
	return regexp.Compile(anchor(expr))
}

// MustCompileAnchored is like [CompileAnchored], but panics if expr can't be parsed.
func MustCompileAnchored(expr string) *regexp.Regexp {
	return regexp.MustCompile(anchor(expr))
}

// IsMatchAll reports whether re is an include filter that matches every name:
// the default of an include flag, [types.RegExpAny] or nil, which the New function of
// a collector replaces with the default.
func IsMatchAll(re *regexp.Regexp) bool {
	return re == nil || slices.Contains(matchAllExpressions, re.String())
}

// IsMatchNone reports whether re is an exclude filter that matches no name:
// the default of an exclude flag, [types.RegExpEmpty] or nil, which the New function
// of a collector replaces with the default.
func IsMatchNone(re *regexp.Regexp) bool {
	return re == nil || slices.Contains(matchNoneExpressions, re.String())
}

// anchoredRegexpValue is a [kingpin.Value] which compiles its value with [CompileAnchored].
type anchoredRegexpValue struct {
	name   string
	expr   string
	target **regexp.Regexp
}

func (v *anchoredRegexpValue) Set(expr string) error {
	re, err := CompileAnchored(expr)
	if err != nil {
		return fmt.Errorf("%s: %w", v.name, err)
	}

	v.expr = expr
	*v.target = re

	return nil
}

func (v *anchoredRegexpValue) String() string {
	return v.expr
}

// RegexpFlag registers the flag name, whose value is a regular expression that
// must match the whole name. The compiled expression is stored in target when
// the flags are parsed, including the default and a value from the
// configuration file. An invalid expression fails the parsing with an error
// that starts with the name of the flag.
func RegexpFlag(app *kingpin.Application, name, help, defaultValue string, target **regexp.Regexp) *kingpin.FlagClause {
	flag := app.Flag(name, help).Default(defaultValue)
	flag.SetValue(&anchoredRegexpValue{name: name, target: target})

	return flag
}
