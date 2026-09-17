// SPDX-License-Identifier: BUSL-1.1

package api

import (
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"testing"
)

// committedProperty reads one property of one definition out of the schema
// that ships in docs/schemas/api. check-schema-drift.sh keeps that file equal
// to what these structs reflect to, so asserting on it asserts on the structs
// — and it is the artefact ui/src/types/generated is actually built from.
//
// The narrowing helpers in schema.go return silently when a property is
// absent, so a renamed `json` tag would widen the generated TypeScript back to
// a bare string with nothing else failing. This is what fails instead.
func committedProperty(t *testing.T, definition, property string) map[string]any {
	t.Helper()

	raw, err := os.ReadFile(filepath.Join("..", "..", "docs", "schemas", "api", "stats.schema.json"))
	if err != nil {
		t.Fatalf("read stats schema: %v", err)
	}
	// A property can be `true` (any value allowed), so the value type is any.
	var schema struct {
		Defs map[string]struct {
			Properties map[string]any `json:"properties"`
		} `json:"$defs"`
	}
	if unmarshalErr := json.Unmarshal(raw, &schema); unmarshalErr != nil {
		t.Fatalf("parse stats schema: %v", unmarshalErr)
	}
	def, ok := schema.Defs[definition]
	if !ok {
		t.Fatalf("stats schema has no definition %q", definition)
	}
	prop, ok := def.Properties[property].(map[string]any)
	if !ok {
		t.Fatalf("%s has no constrained property %q", definition, property)
	}
	return prop
}

func TestStatusPropertiesCarryTheirConstants(t *testing.T) {
	t.Parallel()

	cases := []struct {
		definition string
		property   string
		want       []any
	}{
		{
			definition: "Stats",
			property:   "testStatus",
			want: []any{
				statusIdle, statusStarting, statusRunning, statusCompleted,
				statusError, statusStopped, statusCancelled,
			},
		},
		{
			definition: "RunPlanStep",
			property:   "status",
			want: []any{
				stepPending, stepRunning, stepPassed, stepFailed, stepSkipped,
				statusCancelled,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.definition+"."+tc.property, func(t *testing.T) {
			t.Parallel()

			got, ok := committedProperty(t, tc.definition, tc.property)["enum"].([]any)
			if !ok {
				t.Fatalf("%s.%s carries no enum", tc.definition, tc.property)
			}
			if !slices.Equal(got, tc.want) {
				t.Errorf("enum = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestPointerPropertiesAcceptNull(t *testing.T) {
	t.Parallel()

	cases := []struct {
		property string
		jsonType string
	}{
		{property: "currentTest", jsonType: "string"},
		{property: "estimatedRemainingSeconds", jsonType: "integer"},
	}

	for _, tc := range cases {
		t.Run(tc.property, func(t *testing.T) {
			t.Parallel()

			branches, isList := committedProperty(t, "Stats", tc.property)["oneOf"].([]any)
			if !isList || len(branches) != 2 {
				t.Fatalf("oneOf = %v, want the %s and null branches", branches, tc.jsonType)
			}
			for i, want := range []string{tc.jsonType, "null"} {
				branch, isObject := branches[i].(map[string]any)
				if !isObject || branch["type"] != want {
					t.Errorf("oneOf[%d] = %v, want type %q", i, branches[i], want)
				}
			}
		})
	}
}
