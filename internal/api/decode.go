// SPDX-License-Identifier: BUSL-1.1

package api

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/MustardSeedNetworks/stem/internal/logging"
)

// decodeJSONStrict reads and validates JSON from r.Body into dst. It applies
// [http.MaxBytesReader] to cap memory, DisallowUnknownFields to reject typoed
// keys, and rejects trailing data after the JSON object.
//
// On any decode failure it writes a structured error response (413 for
// oversized bodies, 400 for everything else) and returns false; the caller
// should return immediately. On success it returns true and dst is
// populated.
//
// Pass maxSize as a per-call-site limit. Most handlers pass maxRequestBodySize
// from the package-level constant; auth/MFA endpoints may want a smaller cap.
func decodeJSONStrict(w http.ResponseWriter, r *http.Request, dst any, maxSize int64) bool {
	return decodeJSONStrictWith(w, r, dst, maxSize)
}

// decodeJSONStrictWith is the variant of decodeJSONStrict that carries extra
// structured log fields (e.g., client_ip on auth flows, username on MFA
// verifications) into the WARN line on decode failure.
//
// Same contract as decodeJSONStrict otherwise. Use this for auth, MFA, and
// recovery endpoints where preserving the audit-log breadcrumb matters — the
// security team relies on those fields to spot brute-force patterns.
func decodeJSONStrictWith(
	w http.ResponseWriter,
	r *http.Request,
	dst any,
	maxSize int64,
	extraAttrs ...any,
) bool {
	r.Body = http.MaxBytesReader(w, r.Body, maxSize)

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	logger := logging.FromContext(r.Context())

	if err := decoder.Decode(dst); err != nil {
		var maxBytesErr *http.MaxBytesError
		if errors.As(err, &maxBytesErr) {
			attrs := append([]any{"max_size", maxSize}, extraAttrs...)
			logger.WarnContext(r.Context(), "Request body too large", attrs...)
			WriteError(w, ErrRequestTooLarge)
			return false
		}
		attrs := append([]any{"error", err}, extraAttrs...)
		logger.WarnContext(r.Context(), "Invalid request body", attrs...)
		// Don't leak json parser internals to clients.
		WriteInvalidRequest(w, "Invalid JSON in request body")
		return false
	}

	// Reject "smuggled" data after the top-level JSON object — multiple
	// concatenated payloads, trailing garbage, etc.
	if decoder.More() {
		logger.WarnContext(r.Context(), "Unexpected trailing data after JSON object", extraAttrs...)
		WriteInvalidRequest(w, "Unexpected trailing data after JSON object")
		return false
	}

	return true
}
