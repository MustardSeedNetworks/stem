// SPDX-License-Identifier: BUSL-1.1

package help

// Standard / acronym strings repeated across help text, glossary, and
// test descriptions. Centralised here so future renames stay consistent
// (e.g. updating to RFC 2544bis would be a single-line edit).
const (
	StandardRFC2544  = "RFC 2544"
	StandardRFC2889  = "RFC 2889"
	StandardRFC6349  = "RFC 6349"
	StandardY1564    = "Y.1564"
	StandardITUY1564 = "ITU-T Y.1564"
	StandardITUY1731 = "ITU-T Y.1731"
	StandardMEF      = "MEF"
	StandardTSN      = "TSN"
	TermCIR          = "CIR"
	TermCIRFull      = "Committed Information Rate"
	UnitMbps         = "Mbps"
)

// CLI flag strings repeated across help text and examples.
const (
	FlagDuration      = "--duration"
	FlagFrameSizes    = "--frame-sizes"
	FlagCIR           = "--cir"
	ExampleDuration60 = "--duration 60"
	ExampleCIR100     = "--cir 100"
	ExampleFrameSizes = "--frame-sizes 64,512,1518"
	DefaultFrameSizes = "64,128,256,512,1024,1280,1518"
	LabelFrameSizes   = "Frame Sizes"
	LabelDuration     = "Duration"
)

// Type/value strings repeated in help options (parameter types + boolean values).
const (
	TypeString         = "string"
	TypeInteger        = "integer"
	TypeIntegerSeconds = "integer (seconds)"
	TypeBoolean        = "boolean"
	ValueFalse         = "false"
	ValueAuto          = "auto"
	ValueAll           = "all"
	ValueBaseline      = "baseline"
)

// Module name strings (matches internal/services/*/module.go ModuleName consts).
// Duplicated here to avoid an import cycle from help → services.
const (
	ModuleBenchmark = "benchmark"
	ModuleReflector = "reflector"
)

// Category / test-type wire identifiers used as lookup keys in the help
// content tables (glossary cross-references, test-by-category dispatch,
// tutorial routing). Lowercase to match the JSON/CLI representation.
const (
	CatRFC2544 = "rfc2544"
	CatRFC2889 = "rfc2889"
	CatRFC6349 = "rfc6349"
	CatY1564   = "y1564"
	CatY1731   = "y1731"
	CatMEF     = "mef"
	CatTSN     = "tsn"

	TestTypeThroughput = "throughput"
	TestTypeLatency    = "latency"
	TestTypeFrameLoss  = "frame_loss"
)
