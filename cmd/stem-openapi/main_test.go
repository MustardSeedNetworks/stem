// SPDX-License-Identifier: BUSL-1.1

package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"

	"github.com/MustardSeedNetworks/stem/internal/api"
)

// generated renders the committed source against the live registry, which is
// what `make openapi` writes.
func generated(t *testing.T) map[string]map[string]any {
	t.Helper()
	src, err := os.ReadFile("../../docs/openapi-source.yaml")
	if err != nil {
		t.Fatalf("reading source: %v", err)
	}
	out, err := generate(src, api.RouteManifest())
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	var doc struct {
		Paths map[string]map[string]any `yaml:"paths"`
	}
	if unmarshalErr := yaml.Unmarshal(out, &doc); unmarshalErr != nil {
		t.Fatalf("generated document is not valid YAML: %v", unmarshalErr)
	}
	return doc.Paths
}

// TestEveryRegistryRouteIsDocumented: no registered route may be absent, and
// each carries exactly the methods the registry gates.
func TestEveryRegistryRouteIsDocumented(t *testing.T) {
	paths := generated(t)
	byPrefix := map[string]map[string]any{}
	for docPath, item := range paths {
		key := docPath
		if i := strings.Index(docPath, "/{"); i >= 0 {
			key = docPath[:i+1]
		}
		byPrefix[key] = item
	}

	for _, rt := range api.RouteManifest() {
		if rt.Hidden {
			continue // the embedded UI's catch-all
		}
		item, ok := byPrefix[rt.Path]
		if !ok {
			t.Errorf("registered route %s is missing from the generated document", rt.Path)
			continue
		}
		for _, m := range rt.Methods {
			if _, documented := item[strings.ToLower(m)]; !documented {
				t.Errorf("%s: method %s is registered but not documented", rt.Path, m)
			}
		}
		documented := 0
		for key := range item {
			if key != "parameters" {
				documented++
			}
		}
		if documented != len(rt.Methods) {
			t.Errorf("%s documents %d methods, the registry gates %v", rt.Path, documented, rt.Methods)
		}
	}
}

// TestPolicyIsDocumented checks the half a hand-written spec cannot keep
// true: a CSRF-protected mutation names the CSRF scheme and its 403, and a
// pre-session route asks for no credential.
func TestPolicyIsDocumented(t *testing.T) {
	paths := generated(t)
	op := func(path, method string) map[string]any {
		t.Helper()
		o, ok := paths[path][method].(map[string]any)
		if !ok {
			t.Fatalf("%s %s is not documented", method, path)
		}
		return o
	}

	start := op("/api/v1/test/start", "post")
	if _, has403 := start["responses"].(map[string]any)["403"]; !has403 {
		t.Error("POST /api/v1/test/start documents no CSRF 403")
	}
	security, _ := start["security"].([]any)
	if len(security) != 1 || security[0].(map[string]any)["CsrfToken"] == nil {
		t.Errorf("POST /api/v1/test/start security = %v, want BearerAuth + CsrfToken", start["security"])
	}

	login := op("/api/v1/auth/login", "post")
	if loginSecurity, _ := login["security"].([]any); loginSecurity == nil || len(loginSecurity) != 0 {
		t.Errorf("POST /api/v1/auth/login security = %v, want [] (pre-session)", login["security"])
	}
}

// TestRunWritesTheCommittedDocument: what `make openapi` writes is what is
// committed, so a route change that skips `make openapi` fails `go test` as
// well as the CI drift step.
func TestRunWritesTheCommittedDocument(t *testing.T) {
	out := filepath.Join(t.TempDir(), "openapi.yaml")
	if err := run("../../docs/openapi-source.yaml", out); err != nil {
		t.Fatalf("run: %v", err)
	}
	got, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	committed, err := os.ReadFile("../../docs/openapi.yaml")
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(committed) {
		t.Error("docs/openapi.yaml is stale; run `make openapi`")
	}
}
