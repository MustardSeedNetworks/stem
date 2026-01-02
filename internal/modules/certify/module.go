// Copyright (c) 2025 Mustard Seed Networks. All rights reserved.

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

// testTypes lists all test types owned by the Certify module.
//
//nolint:gochecknoglobals // Static module definition.
var testTypes = []string{
	// RFC 2889 LAN Switch
	"rfc2889_forwarding",
	"rfc2889_caching",
	"rfc2889_learning",
	"rfc2889_broadcast",
	"rfc2889_congestion",
	// RFC 6349 TCP
	"rfc6349_throughput",
	"rfc6349_path",
	// TSN 802.1Qbv
	"tsn_timing",
	"tsn_isolation",
	"tsn_latency",
	"tsn",
}

// testDescriptions provides descriptions for each test type.
//
//nolint:gochecknoglobals // Static module definition.
var testDescriptions = map[string]string{
	// RFC 2889
	"rfc2889_forwarding": "RFC 2889 Forwarding rate test",
	"rfc2889_caching":    "RFC 2889 Address caching capacity",
	"rfc2889_learning":   "RFC 2889 Address learning rate",
	"rfc2889_broadcast":  "RFC 2889 Broadcast forwarding",
	"rfc2889_congestion": "RFC 2889 Congestion control",
	// RFC 6349
	"rfc6349_throughput": "RFC 6349 TCP throughput (BDP analysis)",
	"rfc6349_path":       "RFC 6349 Path analysis (RTT/bandwidth)",
	// TSN
	"tsn_timing":    "IEEE 802.1Qbv Gate timing accuracy",
	"tsn_isolation": "IEEE 802.1Qbv Traffic class isolation",
	"tsn_latency":   "IEEE 802.1Qbv Scheduled latency",
	"tsn":           "IEEE 802.1Qbv Full TSN test suite",
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
	return testTypes
}

// CanRun returns true if this module can execute the given test type.
func (m *Module) CanRun(testType string) bool {
	return slices.Contains(testTypes, testType)
}

// TestDescription returns the description for a given test type.
func (m *Module) TestDescription(testType string) string {
	return testDescriptions[testType]
}
