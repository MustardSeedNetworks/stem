// SPDX-License-Identifier: BUSL-1.1

package measure_test

import (
	"sync/atomic"

	"github.com/MustardSeedNetworks/stem/internal/services/measure"
	"github.com/MustardSeedNetworks/stem/internal/services/orchestrator/dataplane"
)

// mockY1731Dataplane is an in-memory implementation of
// measure.Y1731Dataplane used by unit tests so the executor can be
// exercised without invoking the real cgo dataplane.
//
// Each Run* method returns a preconfigured result/error pair. Call
// counts are recorded atomically so tests can verify dispatch.
type mockY1731Dataplane struct {
	delayResult    *dataplane.Y1731DelayResult
	delayErr       error
	lossResult     *dataplane.Y1731LossResult
	lossErr        error
	slmResult      *dataplane.Y1731LossResult
	slmErr         error
	loopbackResult *dataplane.Y1731LoopbackResult
	loopbackErr    error

	delayCalls    atomic.Uint32
	lossCalls     atomic.Uint32
	slmCalls      atomic.Uint32
	loopbackCalls atomic.Uint32
	closeCalls    atomic.Uint32

	lastConfig atomic.Pointer[dataplane.Y1731Config]
}

// Ensure mockY1731Dataplane satisfies the interface.
var _ measure.Y1731Dataplane = (*mockY1731Dataplane)(nil)

func newMockY1731Dataplane() *mockY1731Dataplane {
	return &mockY1731Dataplane{
		delayResult: &dataplane.Y1731DelayResult{
			FramesSent:       10,
			FramesReceived:   10,
			FramesLost:       0,
			DelayMinUs:       1.0,
			DelayAvgUs:       2.5,
			DelayMaxUs:       5.0,
			DelayVariationUs: 0.5,
		},
		lossResult: &dataplane.Y1731LossResult{
			FramesTx:         100,
			FramesRx:         99,
			NearEndLoss:      1,
			FarEndLoss:       0,
			NearEndLossRatio: 0.01,
			FarEndLossRatio:  0.0,
			AvailabilityPct:  99.0,
		},
		slmResult: &dataplane.Y1731LossResult{
			FramesTx:         200,
			FramesRx:         198,
			NearEndLoss:      2,
			FarEndLoss:       0,
			NearEndLossRatio: 0.01,
			FarEndLossRatio:  0.0,
			AvailabilityPct:  99.0,
		},
		loopbackResult: &dataplane.Y1731LoopbackResult{
			LBMSent:     50,
			LBRReceived: 50,
			RTTMinMs:    0.5,
			RTTAvgMs:    1.0,
			RTTMaxMs:    2.0,
		},
	}
}

func (m *mockY1731Dataplane) RunY1731DelayTest(
	cfg *dataplane.Y1731Config,
) (*dataplane.Y1731DelayResult, error) {
	m.delayCalls.Add(1)
	m.lastConfig.Store(cfg)
	if m.delayErr != nil {
		return nil, m.delayErr
	}
	return m.delayResult, nil
}

func (m *mockY1731Dataplane) RunY1731LossTest(
	cfg *dataplane.Y1731Config,
) (*dataplane.Y1731LossResult, error) {
	m.lossCalls.Add(1)
	m.lastConfig.Store(cfg)
	if m.lossErr != nil {
		return nil, m.lossErr
	}
	return m.lossResult, nil
}

func (m *mockY1731Dataplane) RunY1731SyntheticLossTest(
	cfg *dataplane.Y1731Config,
) (*dataplane.Y1731LossResult, error) {
	m.slmCalls.Add(1)
	m.lastConfig.Store(cfg)
	if m.slmErr != nil {
		return nil, m.slmErr
	}
	return m.slmResult, nil
}

func (m *mockY1731Dataplane) RunY1731LoopbackTest(
	cfg *dataplane.Y1731Config,
) (*dataplane.Y1731LoopbackResult, error) {
	m.loopbackCalls.Add(1)
	m.lastConfig.Store(cfg)
	if m.loopbackErr != nil {
		return nil, m.loopbackErr
	}
	return m.loopbackResult, nil
}

func (m *mockY1731Dataplane) Close() {
	m.closeCalls.Add(1)
}
