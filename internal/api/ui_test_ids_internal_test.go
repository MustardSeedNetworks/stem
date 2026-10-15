// SPDX-License-Identifier: BUSL-1.1

package api

import (
	"os"
	"path/filepath"
	"regexp"
	"testing"
)

// uiTestSectionsDir holds the web UI's test-selection sections, each of which
// lists the test ids it offers in a testKeyMap.
const uiTestSectionsDir = "../../ui/src/components/settings/tests"

// uiOfferedTestIDs returns every test id a test-selection section offers, by
// the file that offers it.
func uiOfferedTestIDs(t *testing.T) map[string]string {
	t.Helper()

	files, err := filepath.Glob(filepath.Join(uiTestSectionsDir, "*Section.tsx"))
	if err != nil {
		t.Fatal(err)
	}
	entry := regexp.MustCompile(`\[\s*'([a-z0-9_]+)',\s*'[A-Za-z0-9]+'\s*\]`)

	ids := make(map[string]string)
	for _, file := range files {
		data, readErr := os.ReadFile(file)
		if readErr != nil {
			t.Fatal(readErr)
		}
		for _, m := range entry.FindAllStringSubmatch(string(data), -1) {
			ids[m[1]] = filepath.Base(file)
		}
	}
	return ids
}

// TestUIOfferedTestsStartARun holds every test the web UI offers to the run
// plan the daemon builds from it: a checkbox whose id no module accepts fails
// the whole run with "unknown test type" (#1708).
func TestUIOfferedTestsStartARun(t *testing.T) {
	ids := uiOfferedTestIDs(t)
	if len(ids) < 20 {
		t.Fatalf("found %d UI test ids; the extraction is broken, not the UI", len(ids))
	}

	for id, file := range ids {
		request := TestStartRequest{Tests: []TestStepRequest{{TestType: id}}}
		if _, err := newRunPlan("ui-"+id, request); err != nil {
			t.Errorf("%s offers %q, which the daemon refuses: %v", file, id, err)
		}
	}
}
