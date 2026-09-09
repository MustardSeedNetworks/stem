// SPDX-License-Identifier: BUSL-1.1

package truststore

import (
	"strings"
	"testing"
)

// maxTrimmedLen mirrors the maxLen constant inside trimError. Operator-facing
// error strings carry a command's combined output, which for a failing
// update-ca-certificates run can be kilobytes.
const maxTrimmedLen = 400

func TestTrimError_TrimsSurroundingWhitespace(t *testing.T) {
	t.Parallel()
	if got := trimError([]byte("\n  certutil: access denied \t\n")); got != "certutil: access denied" {
		t.Errorf("trimError = %q, want %q", got, "certutil: access denied")
	}
}

func TestTrimError_KeepsShortOutputVerbatim(t *testing.T) {
	t.Parallel()
	in := strings.Repeat("x", maxTrimmedLen)
	got := trimError([]byte(in))
	if got != in {
		t.Errorf("output of exactly %d bytes was altered: len(got) = %d", maxTrimmedLen, len(got))
	}
	if strings.HasSuffix(got, "…") {
		t.Error("output of exactly maxLen bytes must not be marked as truncated")
	}
}

func TestTrimError_TruncatesLongOutput(t *testing.T) {
	t.Parallel()
	got := trimError([]byte(strings.Repeat("x", maxTrimmedLen+1)))
	if !strings.HasSuffix(got, "…") {
		t.Errorf("truncated output missing ellipsis marker: %q", got[max(0, len(got)-10):])
	}
	if want := strings.Repeat("x", maxTrimmedLen) + "…"; got != want {
		t.Errorf("trimError truncated to %d runes of payload, want %d", len(strings.TrimSuffix(got, "…")), maxTrimmedLen)
	}
}

func TestTrimError_Empty(t *testing.T) {
	t.Parallel()
	if got := trimError(nil); got != "" {
		t.Errorf("trimError(nil) = %q, want empty", got)
	}
}
