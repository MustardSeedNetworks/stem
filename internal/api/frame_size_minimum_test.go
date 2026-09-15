// SPDX-License-Identifier: BUSL-1.1

package api_test

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/MustardSeedNetworks/stem/internal/api"
)

// The extended payload formats (TSN, TrafficGen, Y.1564) carry a 24-byte
// payload where the RFC 2544 measurement format carries 18, so their packet
// buffer needs 66 bytes and the standard 64-byte Ethernet frame cannot hold
// one (ADR 0008). The dataplane returns a bare -22 for a frame below that;
// these tests pin the 400 that names the minimum instead.
func TestExtendedFormatsRejectFramesBelowTheirMinimum(t *testing.T) {
	tooSmall := api.MinExtendedFrameSize - 1

	cases := []struct {
		name     string
		testType string
		config   string
	}{
		{"tsn", "tsn_isolation", fmt.Sprintf(`{"tsn":{"frameSize":%d}}`, tooSmall)},
		{"trafficGen", "custom_stream", fmt.Sprintf(`{"trafficGen":{"frameSize":%d}}`, tooSmall)},
		{"y1564", "y1564", fmt.Sprintf(`{"y1564":{"frameSizes":[%d]}}`, tooSmall)},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			s := setupTestingTestServer(t)
			s.UseLicenseForTest(proLicense(t))

			body := fmt.Sprintf(
				`{"interface":"lo","peer":"192.0.2.1","tests":[{"testType":%q,"config":%s}]}`,
				tc.testType, tc.config)
			w := startTestRaw(t, s, getTestingAuthToken(t, s), body)

			if w.Code != http.StatusBadRequest {
				t.Fatalf("status = %d, want 400: %s", w.Code, w.Body.String())
			}
			msg := responseMessage(t, w)
			if !strings.Contains(msg, strconv.Itoa(api.MinExtendedFrameSize)) {
				t.Errorf("message = %q, want it to name the minimum %d", msg, api.MinExtendedFrameSize)
			}
		})
	}
}

// A frame at the minimum is not rejected for being too small. The run may fail
// later for want of a peer or a dataplane; what it must not do is 400 over the
// frame size.
func TestExtendedFormatsAcceptTheMinimumFrame(t *testing.T) {
	s := setupTestingTestServer(t)
	s.UseLicenseForTest(proLicense(t))

	config := fmt.Sprintf(`{"tsn":{"frameSize":%d}}`, api.MinExtendedFrameSize)
	body := fmt.Sprintf(
		`{"interface":"lo","peer":"192.0.2.1","tests":[{"testType":"tsn_isolation","config":%s}]}`,
		config)
	w := startTestRaw(t, s, getTestingAuthToken(t, s), body)

	if msg := responseMessage(t, w); strings.Contains(msg, "frame size") {
		t.Errorf("frame size %d was rejected: %q", api.MinExtendedFrameSize, msg)
	}
}

// The minimum has one definition, in include/rfc2544.h. Go mirrors it because
// the Mac build has no CGO and the cgo preamble redeclares the header rather
// than including it (stem#1240), so nothing else would catch the two drifting.
func TestExtendedFrameMinimumMatchesTheHeader(t *testing.T) {
	header, err := os.ReadFile(filepath.Join("..", "..", "include", "rfc2544.h"))
	if err != nil {
		t.Fatalf("read header: %v", err)
	}

	packetMin := headerConstant(t, string(header), "EXTENDED_MIN_PACKET_SIZE")
	fcs := headerConstant(t, string(header), "RFC2544_FCS_SIZE")

	if want := packetMin + fcs; want != api.MinExtendedFrameSize {
		t.Errorf("header says %d (%d packet + %d FCS), Go says %d", want, packetMin, fcs,
			api.MinExtendedFrameSize)
	}
}

// headerConstant reads a plain integer #define out of the C header.
func headerConstant(t *testing.T, header, name string) int {
	t.Helper()
	re := regexp.MustCompile(`(?m)^#define\s+` + name + `\s+(\d+)\s*$`)
	match := re.FindStringSubmatch(header)
	if match == nil {
		t.Fatalf("%s not found in include/rfc2544.h", name)
	}
	value, err := strconv.Atoi(match[1])
	if err != nil {
		t.Fatalf("%s = %q: %v", name, match[1], err)
	}
	return value
}
