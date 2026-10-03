// SPDX-License-Identifier: BUSL-1.1

package api

import (
	"strings"

	"github.com/MustardSeedNetworks/stem/internal/services/orchestrator/dataplane"
)

// Causes a failed run reports on /api/v1/stats. The operator needs to know
// which of these happened — each has a different next step — and nothing more:
// the raw error names interfaces, socket paths and peer addresses, which are
// the daemon host's business and not a stable API. The daemon log keeps the
// full text for support.
const (
	causeInterfaceBusy    = "The interface or port is already in use by another process."
	causeNotPermitted     = "The daemon does not have permission to open the interface."
	causeInterfaceMissing = "The selected interface is not available."
	causeUnreachable      = "The peer did not answer."
	causeGeneric          = "The test failed. See the daemon log for the cause."
	// causeInternalFault is the run that never produced a measurement
	// because the daemon itself faulted — a supervised panic in the
	// dataplane (#1336). It is deliberately distinct from causeGeneric:
	// "the test failed" points the operator at the network, and this one
	// points at the product.
	causeInternalFault = "The run stopped on an internal fault. See the daemon log."
	// causeCriteriaNotMet is a run that measured and whose measurement missed
	// its acceptance criteria (#1463): nothing faulted, and the step's own
	// result says which criterion failed, so the daemon log has nothing to add.
	causeCriteriaNotMet = "The test ran and did not meet its acceptance criteria. See the step results."
	// causeNothingSent is a run that failed because a step put no frame on the
	// wire, so it measured nothing (#1482). A frame larger than the interface
	// MTU is the usual reason, which is what the operator should check first.
	causeNothingSent = "The test transmitted no frames. Check the frame size against the interface MTU."
)

// classifyRunCause maps a failure's own wording onto the closed set above.
// It takes the text rather than an error because a run-plan step can fail with
// an unsuccessful result and no Go error, and that case needs a cause too.
// Matching is on substrings: the text arrives from three layers (Go net, the
// cgo dataplane, module executors) that word the same condition differently.
// An unrecognised cause falls back to the vague message rather than echoing
// the original.
func classifyRunCause(cause string) string {
	msg := strings.ToLower(cause)
	switch {
	case strings.Contains(msg, "address already in use"),
		strings.Contains(msg, "address in use"),
		strings.Contains(msg, "device or resource busy"):
		return causeInterfaceBusy
	case strings.Contains(msg, "operation not permitted"),
		strings.Contains(msg, "permission denied"),
		strings.Contains(msg, "not permitted"):
		return causeNotPermitted
	case strings.Contains(msg, "no such device"),
		strings.Contains(msg, "no such interface"),
		strings.Contains(msg, "interface not found"),
		strings.Contains(msg, "cannot assign requested address"):
		return causeInterfaceMissing
	case strings.Contains(msg, "no route to host"),
		strings.Contains(msg, "connection refused"),
		strings.Contains(msg, "i/o timeout"),
		strings.Contains(msg, "timed out"),
		strings.Contains(msg, dataplane.ErrNoFramesReturned.Error()):
		return causeUnreachable
	case strings.Contains(msg, dataplane.ErrThroughputNoPassingRate.Error()):
		return causeCriteriaNotMet
	default:
		return causeGeneric
	}
}

// classifyResultCause is the cause of a run whose step measured and returned
// an unsuccessful result, from the result's own failure text.
func classifyResultCause(failure string) string {
	if strings.Contains(failure, dataplane.ErrY1564NothingTransmitted.Error()) {
		return causeNothingSent
	}
	return causeCriteriaNotMet
}
