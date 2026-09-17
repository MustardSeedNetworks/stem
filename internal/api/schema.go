// SPDX-License-Identifier: BUSL-1.1

package api

import "github.com/invopop/jsonschema"

// The generated JSON Schemas (docs/schemas/api, and the TypeScript the UI
// reads through ui/src/types/generated) come from reflection over these
// structs, which loses two things Go cannot express in a field's type: a
// string that is really a closed set of constants, and a pointer that the
// encoder writes as JSON null. Both are narrowed here, from the constants
// themselves, so adding a status constant cannot leave a TypeScript union
// silently short.

// nullable widens a property to accept the JSON null a nil pointer encodes to.
func nullable(schema *jsonschema.Schema, property, jsonType string) {
	prop, ok := schema.Properties.Get(property)
	if !ok {
		return
	}
	prop.Type = ""
	prop.OneOf = []*jsonschema.Schema{{Type: jsonType}, {Type: "null"}}
}

// oneOfConstants restricts a property to the given values.
func oneOfConstants(schema *jsonschema.Schema, property string, values ...string) {
	prop, ok := schema.Properties.Get(property)
	if !ok {
		return
	}
	prop.Enum = make([]any, 0, len(values))
	for _, value := range values {
		prop.Enum = append(prop.Enum, value)
	}
}

// JSONSchemaExtend narrows the reflected schema for Stats.
func (Stats) JSONSchemaExtend(schema *jsonschema.Schema) {
	oneOfConstants(schema, "testStatus",
		statusIdle, statusStarting, statusRunning, statusCompleted,
		statusError, statusStopped, statusCancelled,
	)
	nullable(schema, "currentTest", "string")
	nullable(schema, "estimatedRemainingSeconds", "integer")
}

// JSONSchemaExtend narrows the reflected schema for RunPlanStep.
func (RunPlanStep) JSONSchemaExtend(schema *jsonschema.Schema) {
	oneOfConstants(schema, "status",
		stepPending, stepRunning, stepPassed, stepFailed, stepSkipped, statusCancelled,
	)
}
