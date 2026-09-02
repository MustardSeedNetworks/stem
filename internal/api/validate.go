package api

import (
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
)

// dtoValidator is the package-level validator instance used to enforce
// struct-tag rules on HTTP request DTOs. Registered with a json-tag-name
// function so error messages reference the keys clients sent on the wire.
// Matches the package-level helper pattern used by ratelimit caches.
//
//nolint:gochecknoglobals // process-wide validator; init once, immutable thereafter
var dtoValidator = newDTOValidator()

func newDTOValidator() *validator.Validate {
	v := validator.New(validator.WithRequiredStructEnabled())
	v.RegisterTagNameFunc(jsonFieldName)
	return v
}

func jsonFieldName(field reflect.StructField) string {
	tag := field.Tag.Get("json")
	if tag == "" || tag == "-" {
		return ""
	}
	name, _, _ := strings.Cut(tag, ",")
	return name
}

// validateStruct runs the struct-tag validator against dto and, on failure,
// writes a 400 response with a single human-readable line listing the
// failing fields, then returns false. The caller should return immediately.
//
// Pair this with decodeJSONStrict: decode for shape, validate for
// semantics. The two helpers together close the loop on boundary input.
//
// Stem doesn't carry a localizer, so the failure message is English
// only — matching the rest of the API.
func validateStruct(w http.ResponseWriter, dto any) bool {
	err := dtoValidator.Struct(dto)
	if err == nil {
		return true
	}
	if verrs, ok := errors.AsType[validator.ValidationErrors](err); ok {
		WriteInvalidRequest(w, formatValidationErrors(verrs))
		return false
	}
	WriteInvalidRequest(w, "validation failed: "+err.Error())
	return false
}

// formatValidationErrors collapses a ValidationErrors slice into a single
// line like `username: is required; mode: must be one of [reflector
// test_master]`. Each entry uses the json-tag name (configured via
// jsonFieldName above) so the client sees the field they sent, not the Go
// struct field.
func formatValidationErrors(verrs validator.ValidationErrors) string {
	parts := make([]string, 0, len(verrs))
	for _, fe := range verrs {
		// Strip the struct-name prefix validator/v10 prepends:
		// "AuthLoginRequest.username" → "username".
		ns := fe.Namespace()
		if idx := strings.IndexByte(ns, '.'); idx >= 0 {
			ns = ns[idx+1:]
		}
		parts = append(parts, fmt.Sprintf("%s: %s", ns, describeValidationTag(fe)))
	}
	return "validation failed: " + strings.Join(parts, "; ")
}

// describeValidationTag renders a failing rule as something an operator can
// act on. Without this the response says `mode: oneof`, which names the
// library's rule rather than the mistake — a downgrade from the hand-written
// checks this validator replaced, and the reason those checks were worth
// keeping until now.
//
// Unknown tags fall through to the tag name, which is still better than
// nothing and keeps new rules from silently rendering as an empty string.
func describeValidationTag(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return "is required"
	case "oneof":
		return "must be one of [" + fe.Param() + "]"
	case "min":
		return "must be at least " + fe.Param() + lengthOrValueSuffix(fe)
	case "max":
		return "must be at most " + fe.Param() + lengthOrValueSuffix(fe)
	case "gte":
		return "must be >= " + fe.Param()
	case "lte":
		return "must be <= " + fe.Param()
	case "email":
		return "must be a valid email address"
	case "url":
		return "must be a valid URL"
	default:
		return fe.Tag()
	}
}

// lengthOrValueSuffix distinguishes "at least 8 characters" from "at least 8",
// because min/max mean length on strings and slices but magnitude on numbers.
func lengthOrValueSuffix(fe validator.FieldError) string {
	// An if-chain rather than a switch on purpose: a tagged switch trips
	// exhaustive (it wants all ~25 reflect.Kind cases enumerated, which would
	// say nothing) and an untagged one trips staticcheck's QF1002. Every kind
	// not named here is a number, where min/max mean magnitude and no suffix
	// reads correctly.
	kind := fe.Kind()
	if kind == reflect.String {
		return " characters"
	}
	if kind == reflect.Slice || kind == reflect.Array || kind == reflect.Map {
		return " items"
	}
	return ""
}
