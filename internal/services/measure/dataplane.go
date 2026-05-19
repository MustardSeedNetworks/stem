// SPDX-License-Identifier: BUSL-1.1

package measure

import "github.com/krisarmstrong/stem/internal/services/orchestrator/dataplane"

// Y1731Dataplane is the narrow interface the Measure executor depends on.
//
// It captures only the dataplane methods the executor actually calls — the
// four Y.1731 OAM test runners. Consumers depend on this interface rather
// than on the concrete *dataplane.Context, which allows unit tests to
// substitute an in-memory mock and avoid driving the real C dataplane code
// paths (which require raw-socket capabilities and SIGSEGV in CI runners
// without them).
//
// The concrete *dataplane.Context (both the cgo-linked Linux build and the
// stub build for other platforms) already satisfies this interface — see
// internal/services/orchestrator/dataplane.
type Y1731Dataplane interface {
	RunY1731DelayTest(cfg *dataplane.Y1731Config) (*dataplane.Y1731DelayResult, error)
	RunY1731LossTest(cfg *dataplane.Y1731Config) (*dataplane.Y1731LossResult, error)
	RunY1731SyntheticLossTest(cfg *dataplane.Y1731Config) (*dataplane.Y1731LossResult, error)
	RunY1731LoopbackTest(cfg *dataplane.Y1731Config) (*dataplane.Y1731LoopbackResult, error)
	Close()
}
