// SPDX-License-Identifier: BUSL-1.1

// Package certify implements the Certify module for compliance certification.
// This module owns RFC 2889 switch tests, RFC 6349 TCP tests, and TSN 802.1Qbv tests.
package certify

import "slices"

const (
	// ModuleName is the unique identifier for the Certify module.
	ModuleName = "certify"

	// DisplayName is the human-readable name.
	DisplayName = "Certify"

	// ColorHex is the module's UI color (Green).
	ColorHex = "#16a34a"

	// StandardRef is the primary standard this module implements.
	StandardRef = "RFC 2889/6349/TSN"
)

// Test type identifiers. Used as the canonical wire name in the API,
// the registry, and the dispatcher.
const (
	TestRFC2889Forwarding = "rfc2889_forwarding"
	TestRFC2889Caching    = "rfc2889_caching"
	TestRFC2889Learning   = "rfc2889_learning"
	TestRFC2889Broadcast  = "rfc2889_broadcast"
	TestRFC2889Congestion = "rfc2889_congestion"
	TestRFC6349Throughput = "rfc6349_throughput"
	TestRFC6349Path       = "rfc6349_path"
	TestTSNTiming         = "tsn_timing"
	TestTSNIsolation      = "tsn_isolation"
	TestTSNLatency        = "tsn_latency"
	TestTSN               = "tsn"
)

func testTypes() []string {
	return []string{
		// RFC 2889 LAN Switch
		TestRFC2889Forwarding,
		TestRFC2889Caching,
		TestRFC2889Learning,
		TestRFC2889Broadcast,
		TestRFC2889Congestion,
		// RFC 6349 TCP
		TestRFC6349Throughput,
		TestRFC6349Path,
		// TSN 802.1Qbv
		TestTSNTiming,
		TestTSNIsolation,
		TestTSNLatency,
		TestTSN,
	}
}

func testDescriptions() map[string]string {
	return map[string]string{
		// RFC 2889
		TestRFC2889Forwarding: "RFC 2889 Forwarding rate test",
		TestRFC2889Caching:    "RFC 2889 Address caching capacity",
		TestRFC2889Learning:   "RFC 2889 Address learning rate",
		TestRFC2889Broadcast:  "RFC 2889 Broadcast forwarding",
		TestRFC2889Congestion: "RFC 2889 Congestion control",
		// RFC 6349
		TestRFC6349Throughput: "RFC 6349 TCP throughput (BDP analysis)",
		TestRFC6349Path:       "RFC 6349 Path analysis (RTT/bandwidth)",
		// TSN
		TestTSNTiming:    "IEEE 802.1Qbv Gate timing accuracy",
		TestTSNIsolation: "IEEE 802.1Qbv Traffic class isolation",
		TestTSNLatency:   "IEEE 802.1Qbv Scheduled latency",
		TestTSN:          "IEEE 802.1Qbv Full TSN test suite",
	}
}

// Module implements the modules.Module interface for compliance certification.
type Module struct{}

// New creates a new Certify module instance.
func New() *Module {
	return &Module{}
}

// Name returns the module's unique identifier.
func (m *Module) Name() string {
	return ModuleName
}

// DisplayName returns the human-readable name.
func (m *Module) DisplayName() string {
	return DisplayName
}

// Description returns a brief description of the module's purpose.
func (m *Module) Description() string {
	return "Compliance certification - RFC 2889 switch, RFC 6349 TCP, and TSN 802.1Qbv tests"
}

// Color returns the module's UI color in hex format.
func (m *Module) Color() string {
	return ColorHex
}

// Standard returns the primary standard this module implements.
func (m *Module) Standard() string {
	return StandardRef
}

// TestTypes returns the list of test types this module can execute.
func (m *Module) TestTypes() []string {
	return testTypes()
}

// CanRun returns true if this module can execute the given test type.
func (m *Module) CanRun(testType string) bool {
	return slices.Contains(testTypes(), testType)
}

// TestDescription returns the description for a given test type.
func (m *Module) TestDescription(testType string) string {
	return testDescriptions()[testType]
}
