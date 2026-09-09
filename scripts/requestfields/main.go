// SPDX-License-Identifier: BUSL-1.1

// Package main implements the request-field read ratchet used by CI.
package main

import (
	"bufio"
	"flag"
	"fmt"
	"go/ast"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strconv"
	"strings"
)

const (
	requiredArguments = 2
	exitFailure       = 1
	exitUsage         = 2
)

type fieldKey struct {
	typeName  string
	fieldName string
}

func (k fieldKey) String() string { return k.typeName + "." + k.fieldName }

func main() {
	flag.Parse()
	if flag.NArg() != requiredArguments {
		fmt.Fprintln(os.Stderr, "usage: requestfields <api-directory> <baseline-file>")
		os.Exit(exitUsage)
	}

	findings, err := unreadRequestFields(flag.Arg(0))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(exitFailure)
	}
	baseline, err := readBaseline(flag.Arg(1))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(exitFailure)
	}

	if !slices.Equal(findings, baseline) {
		fmt.Fprintln(os.Stderr, "request-field read gate failed")
		printDifference("new unread fields", findings, baseline)
		printDifference("stale baseline entries", baseline, findings)
		os.Exit(exitFailure)
	}

	fmt.Fprintln(os.Stdout, "OK: "+strconv.Itoa(len(findings))+" unread request fields match the ratchet baseline.")
}

func printDifference(label string, left, right []string) {
	var difference []string
	for _, value := range left {
		if !slices.Contains(right, value) {
			difference = append(difference, value)
		}
	}
	if len(difference) == 0 {
		return
	}
	fmt.Fprintln(os.Stderr, label+":")
	for _, value := range difference {
		fmt.Fprintln(os.Stderr, "  "+value)
	}
}

func readBaseline(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open baseline: %w", err)
	}
	defer file.Close()

	var entries []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" && !strings.HasPrefix(line, "#") {
			entries = append(entries, line)
		}
	}
	scanErr := scanner.Err()
	if scanErr != nil {
		return nil, fmt.Errorf("read baseline: %w", scanErr)
	}
	slices.Sort(entries)
	return entries, nil
}

func unreadRequestFields(dir string) ([]string, error) {
	fset, files, err := parsePackage(dir)
	if err != nil {
		return nil, err
	}

	declared := requestFields(files)
	read := fieldReads(dir, fset, files, declared)
	findings := make([]string, 0, len(declared))
	for key := range declared {
		if _, found := read[key]; !found {
			findings = append(findings, key.String())
		}
	}
	slices.Sort(findings)
	return findings, nil
}

func parsePackage(dir string) (*token.FileSet, []*ast.File, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, nil, fmt.Errorf("read API directory: %w", err)
	}
	fset := token.NewFileSet()
	var files []*ast.File
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") || strings.HasSuffix(entry.Name(), "_test.go") {
			continue
		}
		file, parseErr := parser.ParseFile(fset, filepath.Join(dir, entry.Name()), nil, 0)
		if parseErr != nil {
			return nil, nil, fmt.Errorf("parse API package: %w", parseErr)
		}
		files = append(files, file)
	}
	return fset, files, nil
}

func requestFields(files []*ast.File) map[fieldKey]struct{} {
	declared := make(map[fieldKey]struct{})
	for _, file := range files {
		ast.Inspect(file, func(node ast.Node) bool {
			typeSpec, isType := node.(*ast.TypeSpec)
			if !isType || !strings.HasSuffix(typeSpec.Name.Name, "Request") {
				return true
			}
			collectStructFields(typeSpec, declared)
			return false
		})
	}
	return declared
}

func collectStructFields(typeSpec *ast.TypeSpec, declared map[fieldKey]struct{}) {
	structType, isStruct := typeSpec.Type.(*ast.StructType)
	if !isStruct {
		return
	}
	for _, field := range structType.Fields.List {
		if !hasJSONTag(field) {
			continue
		}
		for _, name := range field.Names {
			declared[fieldKey{typeName: typeSpec.Name.Name, fieldName: name.Name}] = struct{}{}
		}
	}
}

func hasJSONTag(field *ast.Field) bool {
	return field.Tag != nil && reflect.StructTag(strings.Trim(field.Tag.Value, "`")).Get("json") != ""
}

func fieldReads(
	dir string,
	fset *token.FileSet,
	files []*ast.File,
	declared map[fieldKey]struct{},
) map[fieldKey]struct{} {
	info := &types.Info{Types: make(map[ast.Expr]types.TypeAndValue)}
	config := types.Config{
		Importer: importer.Default(),
		Error:    func(error) {},
	}
	_, _ = config.Check(dir, fset, files, info)

	writes := assignmentWrites(files)
	read := make(map[fieldKey]struct{})
	for _, file := range files {
		ast.Inspect(file, func(node ast.Node) bool {
			selector, isSelector := node.(*ast.SelectorExpr)
			if !isSelector {
				return true
			}
			if _, writeOnly := writes[selector.Pos()]; writeOnly {
				return true
			}
			key := selectorField(info, selector)
			if _, tracked := declared[key]; tracked {
				read[key] = struct{}{}
			}
			return true
		})
	}
	return read
}

func assignmentWrites(files []*ast.File) map[token.Pos]struct{} {
	writes := make(map[token.Pos]struct{})
	for _, file := range files {
		ast.Inspect(file, func(node ast.Node) bool {
			assignment, isAssignment := node.(*ast.AssignStmt)
			if !isAssignment || assignment.Tok != token.ASSIGN {
				return true
			}
			for _, expression := range assignment.Lhs {
				if selector, isSelector := expression.(*ast.SelectorExpr); isSelector {
					writes[selector.Pos()] = struct{}{}
				}
			}
			return true
		})
	}
	return writes
}

func selectorField(info *types.Info, selector *ast.SelectorExpr) fieldKey {
	typeName := namedType(info.TypeOf(selector.X))
	return fieldKey{typeName: typeName, fieldName: selector.Sel.Name}
}

func namedType(value types.Type) string {
	if pointer, isPointer := value.(*types.Pointer); isPointer {
		value = pointer.Elem()
	}
	named, isNamed := value.(*types.Named)
	if !isNamed || !strings.HasSuffix(named.Obj().Name(), "Request") {
		return ""
	}
	return named.Obj().Name()
}
