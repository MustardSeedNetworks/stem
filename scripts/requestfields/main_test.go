// SPDX-License-Identifier: BUSL-1.1

package main

import (
	"os"
	"path/filepath"
	"slices"
	"testing"
)

func TestUnreadRequestFields(t *testing.T) {
	dir := t.TempDir()
	source := `package fixture
type ExampleRequest struct {
	Read string ` + "`json:\"read\"`" + `
	WriteOnly string ` + "`json:\"writeOnly\"`" + `
	Unused string ` + "`json:\"unused\"`" + `
}
func handle() {
	var req ExampleRequest
	req.WriteOnly = "default"
	_ = req.Read
}
`
	writeFixture(t, dir, source)

	got, err := unreadRequestFields(dir)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"ExampleRequest.Unused", "ExampleRequest.WriteOnly"}
	if !slices.Equal(got, want) {
		t.Fatalf("unread fields = %v, want %v", got, want)
	}
}

func TestUnreadRequestFieldsRecognizesPointerCompositeLiteral(t *testing.T) {
	dir := t.TempDir()
	source := `package fixture
type ExampleRequest struct {
	Used string ` + "`json:\"used\"`" + `
}
func handle() {
	req := &ExampleRequest{}
	_ = req.Used
}
`
	writeFixture(t, dir, source)

	got, err := unreadRequestFields(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 0 {
		t.Fatalf("unread fields = %v, want none", got)
	}
}

func TestUnreadRequestFieldsRespectsLexicalShadowing(t *testing.T) {
	dir := t.TempDir()
	source := `package fixture
type ExampleRequest struct {
	Unread string ` + "`json:\"unread\"`" + `
}
type other struct { Unread string }
func handle() {
	var req ExampleRequest
	{
		req := other{}
		_ = req.Unread
	}
	_ = req
}
`
	writeFixture(t, dir, source)

	got, err := unreadRequestFields(dir)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"ExampleRequest.Unread"}
	if !slices.Equal(got, want) {
		t.Fatalf("unread fields = %v, want %v", got, want)
	}
}

func writeFixture(t *testing.T, dir, source string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, "fixture.go"), []byte(source), 0o600); err != nil {
		t.Fatal(err)
	}
}
