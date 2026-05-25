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
	var verrs validator.ValidationErrors
	if errors.As(err, &verrs) {
		WriteInvalidRequest(w, formatValidationErrors(verrs))
		return false
	}
	WriteInvalidRequest(w, "validation failed: "+err.Error())
	return false
}

// formatValidationErrors collapses a ValidationErrors slice into a single
// line like `username: required; port: gte`. Each entry uses the json-tag
// name (configured via jsonFieldName above) so the client sees the field
// they sent, not the Go struct field.
func formatValidationErrors(verrs validator.ValidationErrors) string {
	parts := make([]string, 0, len(verrs))
	for _, fe := range verrs {
		// Strip the struct-name prefix validator/v10 prepends:
		// "AuthLoginRequest.username" → "username".
		ns := fe.Namespace()
		if idx := strings.IndexByte(ns, '.'); idx >= 0 {
			ns = ns[idx+1:]
		}
		parts = append(parts, fmt.Sprintf("%s: %s", ns, fe.Tag()))
	}
	return "validation failed: " + strings.Join(parts, "; ")
}
