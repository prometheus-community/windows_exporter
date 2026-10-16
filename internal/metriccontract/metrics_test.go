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

// Package metriccontract_test checks that the metric names used by the
// documentation, the mixin and the dashboards are emitted by the exporter.
//
// The collectors only build on Windows and most create their descriptors in
// Build, after probing the system, so the test cannot enumerate descriptors at
// run time on every platform. Instead it reads the Go sources: it finds the
// names passed to prometheus.NewDesc and evaluates them statically, following
// constants, string concatenation, fmt.Sprintf, prometheus.BuildFQName, local
// variables, struct fields and the arguments of the functions that wrap
// NewDesc. A part it cannot evaluate, like a name from the user configuration,
// matches any name characters. The test runs on every platform.
package metriccontract_test

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

const (
	modulePath = "github.com/prometheus-community/windows_exporter/"
	repoRoot   = "../.."

	// dynamic matches a part of a metric name the test cannot evaluate.
	dynamic = `[a-zA-Z0-9_]*`
	// maxPatterns bounds the alternatives of an expression.
	maxPatterns = 256
	// maxDepth bounds the call sites followed for a function argument.
	maxDepth = 6
)

// suffixes are added to the name of a histogram or summary by its series.
//
//nolint:gochecknoglobals
var suffixes = []string{"_bucket", "_sum", "_count"}

// referencePattern finds metric names in documentation and queries. %s is a
// placeholder of a Jsonnet format string. The name must not continue a
// recording rule name, like windows:cpu_time:ratio_rate.
var referencePattern = regexp.MustCompile(`(?:^|[^a-zA-Z0-9_:%])(windows_(?:[a-zA-Z0-9_]|%s)+)`)

// configPattern finds the lines of configuration examples in the
// documentation that name user-defined metrics or objects, like the metric of
// a performancecounter counter or the name of a registry group.
var configPattern = regexp.MustCompile(`^\s*(?:-\s*)?"?(?:name|metric)"?\s*:`)

// dynamicPrefixes are the prefixes of the metric names that are defined by
// the user configuration.
//
//nolint:gochecknoglobals
var dynamicPrefixes = []string{
	"windows_performancecounter_",
	"windows_registry_",
	"windows_wmi_",
}

// historicalReferences are metric names the exporter no longer emits, which
// the documentation names to describe a rename or removal. The reason names
// where.
//
//nolint:gochecknoglobals
var historicalReferences = map[string]string{
	"windows_logical_disk_avg_read_requests_queued":  "docs/collector.logical_disk.md: removed in 0.32, use the read/write *_seconds_total rate",
	"windows_logical_disk_avg_write_requests_queued": "docs/collector.logical_disk.md: removed in 0.32, use the read/write *_seconds_total rate",
	"windows_net_route_info":                         "docs/collector.net.md: documented but never emitted, removed from the docs",
}

// ignoredReferences are windows_ names that are not metrics of the exporter.
//
//nolint:gochecknoglobals
var ignoredReferences = map[string]string{
	"windows_exporter":         "the name of the project",
	"windows_rules":            "a rule file of the mixin tests",
	"windows_alerts":           "an alert file of the mixin tests",
	"windows_hyperv_datastore": "the prefix of the datastore metrics in docs/collector.hyperv.md",
}

func TestMetricReferences(t *testing.T) {
	t.Parallel()

	emitted := emittedMetricPatterns(t)
	require.NotEmpty(t, emitted)

	// A partly dynamic name matches many references, so only the collectors
	// with user-defined metric names may create one. A new one means that the
	// evaluation of the sources missed a name.
	for _, pattern := range emitted {
		if !strings.Contains(pattern, dynamic) {
			continue
		}

		name := strings.ReplaceAll(pattern, dynamic, "*")
		if !slices.ContainsFunc(dynamicPrefixes, func(prefix string) bool { return strings.HasPrefix(pattern, prefix) }) {
			t.Errorf("metric name %s is partly dynamic. Build it from constants, or add its collector to dynamicPrefixes "+
				"if its names are user-defined", name)
		}
	}

	matcher := regexp.MustCompile("^(?:" + strings.Join(emitted, "|") + ")$")

	// samples are emitted names, with x for the dynamic parts.
	samples := make([]string, len(emitted))
	for i, pattern := range emitted {
		samples[i] = strings.ReplaceAll(strings.ReplaceAll(pattern, dynamic, "x"), `\`, "")
	}

	var unresolved []string

	for _, ref := range metricReferences(t) {
		if !resolves(matcher, samples, ref.name) {
			unresolved = append(unresolved, fmt.Sprintf("%s: %s", ref.location, ref.name))
		}
	}

	if len(unresolved) > 0 {
		t.Errorf("%d metric references match no metric the exporter emits. "+
			"Rename them, or update the collector if the metric is missing:\n%s",
			len(unresolved), strings.Join(unresolved, "\n"))
	}
}

// resolves reports whether name, or the base name of a histogram or summary
// series, is emitted.
//
// A name with %s placeholders, from a Jsonnet format string, or ending with
// an underscore, a prefix in a string concatenation, is a template. It
// resolves if it and an emitted name can match the same name: an emitted
// sample, with x for its dynamic parts, matches the template, or the
// template, with x for its placeholders, matches an emitted name.
func resolves(matcher *regexp.Regexp, samples []string, name string) bool {
	if strings.Contains(name, "%s") || strings.HasSuffix(name, "_") {
		pattern := strings.ReplaceAll(regexp.QuoteMeta(name), "%s", "[a-zA-Z0-9_]+")
		sample := strings.ReplaceAll(name, "%s", "x")

		if strings.HasSuffix(name, "_") {
			pattern += "[a-zA-Z0-9_]+"
			sample += "x"
		}

		return matcher.MatchString(sample) ||
			slices.ContainsFunc(samples, regexp.MustCompile("^"+pattern+"$").MatchString)
	}

	if matcher.MatchString(name) {
		return true
	}

	for _, suffix := range suffixes {
		if base, ok := strings.CutSuffix(name, suffix); ok && matcher.MatchString(base) {
			return true
		}
	}

	return false
}

type reference struct {
	location string
	name     string
}

// metricReferences returns the windows_ metric names in the collector
// documentation, the mixin sources and the generated dashboards.
func metricReferences(t *testing.T) []reference {
	t.Helper()

	files := []string{
		filepath.Join(repoRoot, "README.md"),
		filepath.Join(repoRoot, "docs", "README.md"),
		filepath.Join(repoRoot, "contrib", "mixin", "README.md"),
	}

	docs, err := filepath.Glob(filepath.Join(repoRoot, "docs", "collector.*.md"))
	require.NoError(t, err)

	files = append(files, docs...)

	runbooks, err := filepath.Glob(filepath.Join(repoRoot, "docs", "runbooks", "*.md"))
	require.NoError(t, err)

	files = append(files, runbooks...)

	dashboards, err := filepath.Glob(filepath.Join(repoRoot, "dashboard", "*.json"))
	require.NoError(t, err)

	files = append(files, dashboards...)

	err = filepath.WalkDir(filepath.Join(repoRoot, "contrib", "mixin"), func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if entry.IsDir() && entry.Name() == "vendor" {
			return filepath.SkipDir
		}

		switch filepath.Ext(path) {
		case ".jsonnet", ".libsonnet", ".yaml", ".yml":
			files = append(files, path)
		}

		return nil
	})
	require.NoError(t, err)
	require.NotEmpty(t, docs)
	require.NotEmpty(t, dashboards)

	var refs []reference

	for _, path := range files {
		data, err := os.ReadFile(path)
		require.NoError(t, err)

		rel, err := filepath.Rel(repoRoot, path)
		require.NoError(t, err)

		seen := map[string]bool{}

		for i, line := range strings.Split(string(data), "\n") {
			if filepath.Ext(path) == ".md" && configPattern.MatchString(line) {
				continue
			}

			for _, match := range referencePattern.FindAllStringSubmatchIndex(line, -1) {
				name := line[match[2]:match[3]]

				// A label value, like name="windows_nt", is not a metric.
				before := line[:match[2]]
				if strings.HasSuffix(before, `="`) || strings.HasSuffix(before, `~"`) ||
					strings.HasSuffix(before, `='`) || strings.HasSuffix(before, `~'`) {
					continue
				}

				if _, ok := ignoredReferences[name]; ok || seen[name] {
					continue
				}

				if _, ok := historicalReferences[name]; ok {
					continue
				}

				seen[name] = true

				refs = append(refs, reference{location: fmt.Sprintf("%s:%d", filepath.ToSlash(rel), i+1), name: name})
			}
		}
	}

	return refs
}

// emittedMetricPatterns returns regular expressions for the names of the
// metrics the exporter emits, read from the Go sources.
func emittedMetricPatterns(t *testing.T) []string {
	t.Helper()

	loader := newLoader(t)

	var patterns []string

	for _, dir := range []string{"internal", "pkg", "cmd"} {
		err := filepath.WalkDir(filepath.Join(repoRoot, dir), func(path string, entry fs.DirEntry, err error) error {
			if err != nil || !entry.IsDir() {
				return err
			}

			pkg := loader.load(path)
			if pkg == nil {
				return nil
			}

			for _, file := range pkg.files {
				ast.Inspect(file, func(node ast.Node) bool {
					call, ok := node.(*ast.CallExpr)
					if !ok || len(call.Args) == 0 {
						return true
					}

					var name []string

					switch {
					case isCall(call, "prometheus", "NewDesc"):
						name = pkg.eval(call.Args[0], file, 0)
					case isCall(call, "version", "NewCollector"):
						// The build_info metric of client_golang.
						name = concat(pkg.eval(call.Args[0], file, 0), []string{"_build_info"})
					default:
						return true
					}

					for _, pattern := range name {
						// A name without a literal windows_ prefix is not
						// part of the contract, like the user-defined
						// performancecounter metrics.
						if strings.HasPrefix(pattern, "windows_") {
							patterns = append(patterns, pattern)
						}
					}

					return true
				})
			}

			return nil
		})
		require.NoError(t, err)
	}

	slices.Sort(patterns)

	return slices.Compact(patterns)
}

func isCall(call *ast.CallExpr, pkg, name string) bool {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok || sel.Sel.Name != name {
		return false
	}

	ident, ok := sel.X.(*ast.Ident)

	return ok && ident.Name == pkg
}

// loader parses the packages of the module by directory.
type loader struct {
	t        *testing.T
	fset     *token.FileSet
	packages map[string]*pkgSource
}

func newLoader(t *testing.T) *loader {
	t.Helper()

	return &loader{t: t, fset: token.NewFileSet(), packages: map[string]*pkgSource{}}
}

// pkgSource holds the non-test Go files of a package, for all platforms.
type pkgSource struct {
	loader *loader
	dir    string
	files  []*ast.File
	// parents maps every node to its parent.
	parents map[ast.Node]ast.Node
	// values maps package-level constants and variables to their values.
	values map[string]ast.Expr
}

func (l *loader) load(dir string) *pkgSource {
	dir = filepath.Clean(dir)
	if pkg, ok := l.packages[dir]; ok {
		return pkg
	}

	paths, err := filepath.Glob(filepath.Join(dir, "*.go"))
	require.NoError(l.t, err)

	pkg := &pkgSource{loader: l, dir: dir, parents: map[ast.Node]ast.Node{}, values: map[string]ast.Expr{}}

	for _, path := range paths {
		if strings.HasSuffix(path, "_test.go") {
			continue
		}

		file, err := parser.ParseFile(l.fset, path, nil, parser.SkipObjectResolution)
		require.NoError(l.t, err)

		pkg.files = append(pkg.files, file)

		var stack []ast.Node

		ast.Inspect(file, func(node ast.Node) bool {
			if node == nil {
				stack = stack[:len(stack)-1]

				return false
			}

			if len(stack) > 0 {
				pkg.parents[node] = stack[len(stack)-1]
			}

			stack = append(stack, node)

			return true
		})

		for _, decl := range file.Decls {
			gen, ok := decl.(*ast.GenDecl)
			if !ok || (gen.Tok != token.CONST && gen.Tok != token.VAR) {
				continue
			}

			for _, spec := range gen.Specs {
				valueSpec := spec.(*ast.ValueSpec) //nolint:forcetypeassert // const and var declarations hold value specs.
				for i, name := range valueSpec.Names {
					if i < len(valueSpec.Values) {
						pkg.values[name.Name] = valueSpec.Values[i]
					}
				}
			}
		}
	}

	if len(pkg.files) == 0 {
		pkg = nil
	}

	l.packages[dir] = pkg

	return pkg
}

// imported returns the package that file imports as name.
func (p *pkgSource) imported(file *ast.File, name string) *pkgSource {
	for _, spec := range file.Imports {
		path, err := strconv.Unquote(spec.Path.Value)
		require.NoError(p.loader.t, err)

		rel, ok := strings.CutPrefix(path, modulePath)
		if !ok {
			continue
		}

		importName := filepath.Base(rel)
		if spec.Name != nil {
			importName = spec.Name.Name
		}

		if importName == name {
			return p.loader.load(filepath.Join(repoRoot, filepath.FromSlash(rel)))
		}
	}

	return nil
}

// fileOf returns the file that contains node.
func (p *pkgSource) fileOf(node ast.Node) *ast.File {
	for node != nil {
		if file, ok := node.(*ast.File); ok {
			return file
		}

		node = p.parents[node]
	}

	return nil
}

// eval returns regular expressions for the string values expr can have.
func (p *pkgSource) eval(expr ast.Expr, file *ast.File, depth int) []string {
	if depth > maxDepth {
		return []string{dynamic}
	}

	switch e := expr.(type) {
	case *ast.BasicLit:
		if e.Kind != token.STRING {
			return []string{regexp.QuoteMeta(e.Value)}
		}

		value, err := strconv.Unquote(e.Value)
		require.NoError(p.loader.t, err)

		return []string{regexp.QuoteMeta(value)}
	case *ast.ParenExpr:
		return p.eval(e.X, file, depth)
	case *ast.BinaryExpr:
		if e.Op != token.ADD {
			return []string{dynamic}
		}

		return concat(p.eval(e.X, file, depth), p.eval(e.Y, file, depth))
	case *ast.Ident:
		return p.evalIdent(e, depth)
	case *ast.SelectorExpr:
		if ident, ok := e.X.(*ast.Ident); ok {
			if imported := p.imported(file, ident.Name); imported != nil {
				if value, ok := imported.values[e.Sel.Name]; ok {
					return imported.eval(value, imported.fileOf(value), depth+1)
				}

				return []string{dynamic}
			}
		}

		return p.evalField(e.Sel.Name, depth)
	case *ast.CallExpr:
		return p.evalCall(e, file, depth)
	default:
		return []string{dynamic}
	}
}

// evalCall evaluates the calls that build metric names.
func (p *pkgSource) evalCall(call *ast.CallExpr, file *ast.File, depth int) []string {
	switch {
	case isCall(call, "prometheus", "BuildFQName") && len(call.Args) == 3:
		var parts [][]string

		for _, arg := range call.Args {
			if lit, ok := arg.(*ast.BasicLit); ok && lit.Value == `""` {
				continue
			}

			parts = append(parts, p.eval(arg, file, depth))
		}

		result := []string{""}

		for i, part := range parts {
			if i > 0 {
				result = concat(result, []string{"_"})
			}

			result = concat(result, part)
		}

		return result
	case isCall(call, "fmt", "Sprintf") && len(call.Args) > 0:
		lit, ok := call.Args[0].(*ast.BasicLit)
		if !ok {
			return []string{dynamic}
		}

		format, err := strconv.Unquote(lit.Value)
		require.NoError(p.loader.t, err)

		result := []string{""}
		args := call.Args[1:]

		for {
			before, after, found := strings.Cut(format, "%")
			result = concat(result, []string{regexp.QuoteMeta(before)})

			if !found || after == "" {
				return result
			}

			verb := after[0]
			format = after[1:]

			switch {
			case verb == '%':
				result = concat(result, []string{"%"})
			case len(args) == 0:
				return []string{dynamic}
			case verb == 's' || verb == 'v':
				result = concat(result, p.eval(args[0], file, depth))
				args = args[1:]
			default:
				result = concat(result, []string{dynamic})
				args = args[1:]
			}
		}
	case isCall(call, "strings", "ToLower") && len(call.Args) == 1:
		return p.eval(call.Args[0], file, depth)
	}

	// A sanitizer keeps the valid name characters of its argument.
	if ident, ok := call.Fun.(*ast.Ident); ok && strings.HasPrefix(ident.Name, "sanitize") && len(call.Args) == 1 {
		return p.eval(call.Args[0], file, depth)
	}

	return []string{dynamic}
}

// evalIdent evaluates a variable, constant or function parameter.
func (p *pkgSource) evalIdent(ident *ast.Ident, depth int) []string {
	// Look for a declaration in the enclosing functions, innermost first.
	for node := p.parents[ident]; node != nil; node = p.parents[node] {
		var (
			funcType *ast.FuncType
			body     *ast.BlockStmt
		)

		switch fn := node.(type) {
		case *ast.FuncDecl:
			funcType, body = fn.Type, fn.Body
		case *ast.FuncLit:
			funcType, body = fn.Type, fn.Body
		default:
			continue
		}

		if index, ok := paramIndex(funcType, ident.Name); ok {
			return p.evalParam(node, index, depth)
		}

		if values, ok := p.localValues(body, ident.Name, depth); ok {
			return values
		}
	}

	if value, ok := p.values[ident.Name]; ok {
		return p.eval(value, p.fileOf(value), depth+1)
	}

	return []string{dynamic}
}

// paramIndex returns the position of the parameter name of a function.
func paramIndex(funcType *ast.FuncType, name string) (int, bool) {
	index := 0

	for _, field := range funcType.Params.List {
		if len(field.Names) == 0 {
			index++

			continue
		}

		for _, fieldName := range field.Names {
			if fieldName.Name == name {
				return index, true
			}

			index++
		}
	}

	return 0, false
}

// evalParam evaluates the argument index of every call of the function fn
// in the package.
func (p *pkgSource) evalParam(fn ast.Node, index int, depth int) []string {
	var (
		name  string
		scope ast.Node
	)

	switch f := fn.(type) {
	case *ast.FuncDecl:
		name = f.Name.Name
	case *ast.FuncLit:
		// A function literal assigned to a variable, called in the
		// enclosing function.
		assign, ok := p.parents[f].(*ast.AssignStmt)
		if !ok || len(assign.Lhs) != 1 {
			return []string{dynamic}
		}

		ident, ok := assign.Lhs[0].(*ast.Ident)
		if !ok {
			return []string{dynamic}
		}

		name = ident.Name

		for scope = p.parents[assign]; scope != nil; scope = p.parents[scope] {
			if _, ok := scope.(*ast.BlockStmt); ok {
				break
			}
		}
	}

	var (
		result []string
		calls  int
	)

	visit := func(node ast.Node) bool {
		call, ok := node.(*ast.CallExpr)
		if !ok || len(call.Args) <= index || call.Ellipsis.IsValid() {
			return true
		}

		var callee string

		switch fun := call.Fun.(type) {
		case *ast.Ident:
			callee = fun.Name
		case *ast.SelectorExpr:
			callee = fun.Sel.Name
		}

		if callee != name {
			return true
		}

		calls++

		result = append(result, p.eval(call.Args[index], p.fileOf(call), depth+1)...)

		return true
	}

	if scope != nil {
		ast.Inspect(scope, visit)
	} else {
		for _, file := range p.files {
			ast.Inspect(file, visit)
		}
	}

	if calls == 0 {
		return []string{dynamic}
	}

	return limit(result)
}

// localValues evaluates the values assigned to the local variable or
// constant name in body.
func (p *pkgSource) localValues(body *ast.BlockStmt, name string, depth int) ([]string, bool) {
	if body == nil {
		return nil, false
	}

	var (
		result []string
		found  bool
	)

	ast.Inspect(body, func(node ast.Node) bool {
		switch n := node.(type) {
		case *ast.FuncLit:
			// The variables of nested functions are their own.
			return false
		case *ast.AssignStmt:
			for i, lhs := range n.Lhs {
				if ident, ok := lhs.(*ast.Ident); ok && ident.Name == name {
					found = true

					if len(n.Lhs) == len(n.Rhs) {
						result = append(result, p.eval(n.Rhs[i], p.fileOf(n), depth+1)...)
					} else {
						result = append(result, dynamic)
					}
				}
			}
		case *ast.ValueSpec:
			for i, ident := range n.Names {
				if ident.Name == name {
					found = true

					if i < len(n.Values) {
						result = append(result, p.eval(n.Values[i], p.fileOf(n), depth+1)...)
					} else {
						result = append(result, dynamic)
					}
				}
			}
		case *ast.RangeStmt:
			for _, expr := range []ast.Expr{n.Key, n.Value} {
				if ident, ok := expr.(*ast.Ident); ok && ident.Name == name {
					found = true

					result = append(result, dynamic)
				}
			}
		}

		return true
	})

	return limit(result), found
}

// evalField evaluates the values assigned to struct fields called name in
// the package, by assignment or in composite literals.
//
// It over-approximates: it ignores the struct type and the receiver, so the
// values of every field with that name in the package count. That can make a
// stale reference resolve when another struct has a field with the same
// name, but never makes an emitted name unresolved.
func (p *pkgSource) evalField(name string, depth int) []string {
	var result []string

	for _, file := range p.files {
		ast.Inspect(file, func(node ast.Node) bool {
			switch n := node.(type) {
			case *ast.AssignStmt:
				for i, lhs := range n.Lhs {
					if sel, ok := lhs.(*ast.SelectorExpr); ok && sel.Sel.Name == name && len(n.Lhs) == len(n.Rhs) {
						result = append(result, p.eval(n.Rhs[i], file, depth+1)...)
					}
				}
			case *ast.KeyValueExpr:
				if key, ok := n.Key.(*ast.Ident); ok && key.Name == name {
					result = append(result, p.eval(n.Value, file, depth+1)...)
				}
			}

			return true
		})
	}

	if len(result) == 0 {
		return []string{dynamic}
	}

	return limit(result)
}

// concat returns every concatenation of a value of a and a value of b.
func concat(a, b []string) []string {
	result := make([]string, 0, len(a)*len(b))

	for _, x := range a {
		for _, y := range b {
			result = append(result, x+y)
		}
	}

	return limit(result)
}

// limit removes duplicates and collapses too many alternatives into a
// dynamic value.
func limit(values []string) []string {
	slices.Sort(values)
	values = slices.Compact(values)

	if len(values) > maxPatterns {
		return []string{dynamic}
	}

	for i, value := range values {
		// Adjacent dynamic parts are one dynamic part.
		for strings.Contains(value, dynamic+dynamic) {
			value = strings.ReplaceAll(value, dynamic+dynamic, dynamic)
		}

		values[i] = value
	}

	return values
}
