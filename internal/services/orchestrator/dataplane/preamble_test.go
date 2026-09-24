package dataplane_test

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

var (
	cgoPreamble   = regexp.MustCompile(`(?s)/\*(.*?)\*/\s*import "C"`)
	typedefClose  = regexp.MustCompile(`}\s*(\w+)\s*;`)
	includeHeader = regexp.MustCompile(`(?m)^#include "rfc2544(_internal)?\.h"$`)
)

// TestCgoPreamblesIncludeTheHeader guards #1240. A preamble that redeclares a
// struct from include/rfc2544.h gives Go a second layout that nothing keeps in
// step with the one C compiles against. That is how the Y.1564 results came to
// lack service_name[32], leaving Go's result 32 bytes short of what C writes
// into it, and how rfc2889_cache_result_t came to be read field by field from
// a struct C never produced. The test reads source rather than calling cgo so
// it runs in every build, not only the Linux cgo one.
func TestCgoPreamblesIncludeTheHeader(t *testing.T) {
	headerTypes := readHeaderTypes(t)

	files, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatalf("glob: %v", err)
	}
	var preambles int
	for _, name := range files {
		preamble, ok := readPreamble(t, name)
		if !ok {
			continue
		}
		preambles++
		if !includeHeader.MatchString(preamble) {
			t.Errorf(`%s: cgo preamble includes neither "rfc2544.h" nor "rfc2544_internal.h"`, name)
		}
		if redeclared := redeclaredTypes(preamble, headerTypes); len(redeclared) > 0 {
			t.Errorf("%s: cgo preamble redeclares %d header types: %s",
				name, len(redeclared), strings.Join(redeclared, ", "))
		}
	}
	if preambles == 0 {
		t.Fatal("found no cgo preamble in the package; the preamble pattern no longer matches")
	}
}

func readHeaderTypes(t *testing.T) map[string]bool {
	t.Helper()
	types := map[string]bool{}
	for _, h := range []string{"rfc2544.h", "rfc2544_internal.h"} {
		header, err := os.ReadFile(filepath.Join("..", "..", "..", "..", "include", h))
		if err != nil {
			t.Fatalf("read header: %v", err)
		}
		for _, m := range typedefClose.FindAllStringSubmatch(string(header), -1) {
			types[m[1]] = true
		}
	}
	for _, want := range []string{"throughput_result_t", "trial_result_t"} {
		if !types[want] {
			t.Fatalf("found no %s in the headers; the typedef pattern no longer matches them", want)
		}
	}
	return types
}

func readPreamble(t *testing.T, name string) (string, bool) {
	t.Helper()
	src, err := os.ReadFile(name)
	if err != nil {
		t.Fatalf("read %s: %v", name, err)
	}
	m := cgoPreamble.FindSubmatch(src)
	if m == nil {
		return "", false
	}
	return string(m[1]), true
}

func redeclaredTypes(preamble string, headerTypes map[string]bool) []string {
	var redeclared []string
	for _, d := range typedefClose.FindAllStringSubmatch(preamble, -1) {
		if headerTypes[d[1]] {
			redeclared = append(redeclared, d[1])
		}
	}
	return redeclared
}
