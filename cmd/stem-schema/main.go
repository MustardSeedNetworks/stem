// stem-schema generates JSON Schemas for stem's HTTP API request DTOs
// directly from the Go types, so the wire contract documented in
// docs/schemas/ stays in sync with internal/api/.
//
// Mirrors the pattern documented in ADR 0001 of MustardSeedNetworks/niac-go
// (docs/adr/0001-schema-generation-from-go-structs.md). Each top-level
// DTO becomes its own schema file under docs/schemas/api/; clients and
// future TypeScript codegen can consume them as a stable contract.
//
// Usage:
//
//	stem-schema -o docs/schemas/api
//
// The output directory is created if it doesn't exist. Existing files
// are overwritten — this command is the source of truth.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/invopop/jsonschema"

	"github.com/MustardSeedNetworks/stem/internal/api"
	"github.com/MustardSeedNetworks/stem/internal/netif"
)

// schemaTarget pairs a Go DTO with the on-disk filename it should be
// written to. Adding a DTO to this list is the only step required to
// generate a schema for it; the generator handles the rest.
type schemaTarget struct {
	value    any    // pointer to a zero-value of the DTO
	filename string // filename without directory (e.g., "login.schema.json")
	title    string // human-readable schema title
}

// schemaTargets returns the DTOs we currently publish schemas for.
//
// Requests: the auth + license + mode + recovery DTOs — the surface that
// carries `validate:` tags (#268). Responses: the DTOs the UI reads on
// every poll. Those were hand-transcribed into ui/src/types/api.ts and
// compared to nothing, which is how a TypeScript-only Stats.errorMessage
// survived (#1251); generating them is what makes a Go field added
// without its TS counterpart fail check-types-drift.sh.
//
// Function rather than package-level var to keep gochecknoglobals
// happy and to make the list lazily constructed.
func schemaTargets() []schemaTarget {
	return []schemaTarget{
		{
			value:    &api.AuthLoginRequest{},
			filename: "auth-login.schema.json",
			title:    "AuthLoginRequest",
		},
		{
			value:    &api.AuthRefreshRequest{},
			filename: "auth-refresh.schema.json",
			title:    "AuthRefreshRequest",
		},
		{
			value:    &api.ModeRequest{},
			filename: "mode.schema.json",
			title:    "ModeRequest",
		},
		{
			value:    &api.LicenseActivateRequest{},
			filename: "license-activate.schema.json",
			title:    "LicenseActivateRequest",
		},
		{
			value:    &api.RecoveryCompleteRequest{},
			filename: "recovery-complete.schema.json",
			title:    "RecoveryCompleteRequest",
		},
		{
			value:    &api.AuthLoginResponse{},
			filename: "auth-login-response.schema.json",
			title:    "AuthLoginResponse",
		},
		{
			value:    &api.Stats{},
			filename: "stats.schema.json",
			title:    "Stats",
		},
		{
			value:    &api.TestResultResponse{},
			filename: "test-result.schema.json",
			title:    "TestResultResponse",
		},
		{
			// The /api/v1/interfaces handler writes netif's type straight
			// to the wire, so that is the DTO, package boundary or not.
			value:    &netif.InterfaceInfo{},
			filename: "interface-info.schema.json",
			title:    "InterfaceInfo",
		},
	}
}

func main() {
	outDir := flag.String("o", "docs/schemas/api", "output directory")
	flag.Parse()

	// 0o750 is the strictest mode that still lets the operator's group
	// list the dir; gosec G301 flags anything looser.
	if err := os.MkdirAll(*outDir, 0o750); err != nil {
		fmt.Fprintf(os.Stderr, "mkdir %s: %v\n", *outDir, err)
		os.Exit(1)
	}

	reflector := newReflector()

	for _, target := range schemaTargets() {
		schema := reflector.Reflect(target.value)
		schema.Title = target.title
		schema.Description = fmt.Sprintf(
			"%s — generated from the Go struct in internal/api; refresh with `make schema` after struct changes.",
			target.title,
		)
		schema.ID = jsonschema.ID(fmt.Sprintf(
			"https://raw.githubusercontent.com/MustardSeedNetworks/stem/main/docs/schemas/api/%s",
			target.filename,
		))

		data, err := json.MarshalIndent(schema, "", "  ")
		if err != nil {
			fmt.Fprintf(os.Stderr, "marshal %s: %v\n", target.filename, err)
			os.Exit(1)
		}
		data = append(data, '\n')

		path := filepath.Join(*outDir, target.filename)
		if writeErr := os.WriteFile(path, data, 0o600); writeErr != nil {
			fmt.Fprintf(os.Stderr, "write %s: %v\n", path, writeErr)
			os.Exit(1)
		}
		// Routed through stderr so stdout stays reserved for generator
		// output downstream tooling might consume; the trailing summary
		// is operator feedback, not data.
		fmt.Fprintf(os.Stderr, "wrote %s\n", path)
	}
}

// newReflector returns a jsonschema.Reflector configured for HTTP API
// DTOs:
//
//   - FieldNameTag: "json" — schemas reflect the wire format clients see,
//     not the Go field casing
//   - AllowAdditionalProperties: false — schemas match the
//     DisallowUnknownFields posture in decodeJSONStrict
//   - Anonymous: true — nested types are inlined rather than producing
//     $ref indirection, which makes schemas easier to consume by tools
//     that don't resolve refs across files
func newReflector() *jsonschema.Reflector {
	r := &jsonschema.Reflector{
		ExpandedStruct:            false,
		Anonymous:                 true,
		AllowAdditionalProperties: false,
	}
	r.KeyNamer = func(s string) string { return s }
	r.FieldNameTag = "json"
	return r
}
