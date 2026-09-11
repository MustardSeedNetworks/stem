// SPDX-License-Identifier: BUSL-1.1

package main

import (
	"encoding/json"
	"strings"
	"testing"
)

func decodeData(t *testing.T, raw string) any {
	t.Helper()
	var v any
	if err := json.Unmarshal([]byte(raw), &v); err != nil {
		t.Fatalf("decode fixture: %v", err)
	}
	return v
}

// The daemon's result arrives as JSON, so the renderer must read what is
// actually there. The old printers type-switched on in-process dataplane
// structs and would have fallen through to a %+v dump for every run.
func TestRenderResultShowsMeasuredValues(t *testing.T) {
	data := decodeData(t, `{
		"maxRatePct": 98.5,
		"maxRateMbps": 985.25,
		"iterations": 7,
		"latency": {"avgNs": 12500, "maxNs": 40000}
	}`)

	out := captureStdout(t, func() { renderResult(data) })

	for _, want := range []string{"maxRatePct", "98.5", "maxRateMbps", "985.25", "iterations", "7", "latency.avgNs", "12500"} {
		if !strings.Contains(out, want) {
			t.Errorf("rendered result is missing %q:\n%s", want, out)
		}
	}
}

// Keys come out in a stable order; a map's iteration order would make the
// CLI's output differ run to run for the same result.
func TestRenderResultIsOrdered(t *testing.T) {
	data := decodeData(t, `{"zeta": 1, "alpha": 2, "mu": 3}`)

	first := captureStdout(t, func() { renderResult(data) })
	second := captureStdout(t, func() { renderResult(data) })
	if first != second {
		t.Errorf("two renders of one result differ:\n%s\n---\n%s", first, second)
	}
	alpha, mu, zeta := strings.Index(first, "alpha"), strings.Index(first, "mu"), strings.Index(first, "zeta")
	if alpha >= mu || mu >= zeta {
		t.Errorf("keys are not in sorted order:\n%s", first)
	}
}

func TestRenderResultHandlesNothing(t *testing.T) {
	if out := captureStdout(t, func() { renderResult(nil) }); strings.TrimSpace(out) != "" {
		t.Errorf("renderResult(nil) printed %q, want nothing", out)
	}
}

// A non-object result (a bare list of per-load rows, say) still has to
// reach the operator rather than being dropped for not being a map.
func TestRenderResultHandlesAList(t *testing.T) {
	data := decodeData(t, `[{"loadPct": 50, "avgNs": 900}, {"loadPct": 100, "avgNs": 1800}]`)

	out := captureStdout(t, func() { renderResult(data) })
	for _, want := range []string{"loadPct", "50", "avgNs", "1800"} {
		if !strings.Contains(out, want) {
			t.Errorf("rendered list is missing %q:\n%s", want, out)
		}
	}
}

// CSV is consumed by scripts, so its columns are derived from the result
// rather than from a schema the daemon no longer guarantees.
func TestRenderCSVEmitsAHeaderAndARow(t *testing.T) {
	data := decodeData(t, `{"frameSize": 64, "maxRatePct": 98.5, "maxRateMbps": 985.25}`)

	out := captureStdout(t, func() { renderCSV(data) })
	lines := strings.Split(strings.TrimSpace(out), "\n")
	if len(lines) != 2 {
		t.Fatalf("renderCSV emitted %d lines, want a header and one row:\n%s", len(lines), out)
	}
	if lines[0] != "frameSize,maxRateMbps,maxRatePct" {
		t.Errorf("header = %q, want sorted keys", lines[0])
	}
	if lines[1] != "64,985.25,98.5" {
		t.Errorf("row = %q", lines[1])
	}
}

// A list of rows is the shape a multi-frame-size run produces.
func TestRenderCSVEmitsOneRowPerEntry(t *testing.T) {
	data := decodeData(t, `[{"frameSize": 64, "maxRatePct": 98.5}, {"frameSize": 1518, "maxRatePct": 100}]`)

	out := captureStdout(t, func() { renderCSV(data) })
	lines := strings.Split(strings.TrimSpace(out), "\n")
	if len(lines) != 3 {
		t.Fatalf("renderCSV emitted %d lines, want a header and two rows:\n%s", len(lines), out)
	}
	if lines[1] != "64,98.5" || lines[2] != "1518,100" {
		t.Errorf("rows = %q, %q", lines[1], lines[2])
	}
}
