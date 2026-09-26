// SPDX-License-Identifier: BUSL-1.1

package api

import (
	"errors"
	"reflect"
	"sync"
	"testing"

	"github.com/MustardSeedNetworks/stem/internal/services/modtypes"
)

// sizeRecorder records the frame size of every call. Its verdict per size is
// an input, so a test can fail one size and see the rest still measured.
type sizeRecorder struct {
	mu     sync.Mutex
	sizes  []uint32
	failAt uint32
	errAt  uint32
	after  func()
}

func (*sizeRecorder) Close() {}

func (r *sizeRecorder) Execute(testType string, cfg *modtypes.TestConfig) (*modtypes.Result, error) {
	r.mu.Lock()
	r.sizes = append(r.sizes, cfg.FrameSize)
	r.mu.Unlock()
	if r.after != nil {
		defer r.after()
	}
	if cfg.FrameSize == r.errAt {
		return nil, errors.New("dataplane fault")
	}
	result := &modtypes.Result{
		TestType: testType,
		Success:  cfg.FrameSize != r.failAt,
		Data:     map[string]any{"frameSize": cfg.FrameSize},
	}
	if !result.Success {
		result.Error = "the service did not meet its acceptance criteria"
	}
	return result, nil
}

func (r *sizeRecorder) measured() []uint32 {
	r.mu.Lock()
	defer r.mu.Unlock()
	return append([]uint32(nil), r.sizes...)
}

// Before D-STEM-29 (stem#1464) a step measured only its first selected size:
// the list went into a params key no module read, while the estimate already
// multiplied by every size.
func TestRunPlanMeasuresEverySelectedFrameSize(t *testing.T) {
	sizes := []uint32{64, 512, 1518}
	tests := []struct {
		name       string
		testType   string
		config     *TestConfig
		recorder   *sizeRecorder
		wantSizes  []uint32
		wantStatus string
		wantStep   string
		wantData   []uint32
	}{
		{
			name:       "RFC 2544 throughput",
			testType:   "rfc2544_throughput",
			config:     &TestConfig{RFC2544: &RFC2544TestConfig{Duration: 1, Trials: 1, FrameSizes: sizes}},
			recorder:   &sizeRecorder{},
			wantSizes:  sizes,
			wantStatus: statusCompleted,
			wantStep:   stepPassed,
			wantData:   sizes,
		},
		{
			name:       "Y.1564 configuration",
			testType:   "y1564_config",
			config:     &TestConfig{Y1564: &Y1564TestConfig{FrameSizes: []uint32{128, 1518}}},
			recorder:   &sizeRecorder{},
			wantSizes:  []uint32{128, 1518},
			wantStatus: statusCompleted,
			wantStep:   stepPassed,
			wantData:   []uint32{128, 1518},
		},
		{
			name:       "a failed verdict fails the step and the other sizes are still measured",
			testType:   "y1564_config",
			config:     &TestConfig{Y1564: &Y1564TestConfig{FrameSizes: []uint32{128, 512, 1518}}},
			recorder:   &sizeRecorder{failAt: 512},
			wantSizes:  []uint32{128, 512, 1518},
			wantStatus: statusError,
			wantStep:   stepFailed,
			wantData:   []uint32{128, 512, 1518},
		},
		{
			name:       "an execution error ends the step and keeps the sizes already measured",
			testType:   "rfc2544_throughput",
			config:     &TestConfig{RFC2544: &RFC2544TestConfig{Duration: 1, Trials: 1, FrameSizes: sizes}},
			recorder:   &sizeRecorder{errAt: 512},
			wantSizes:  []uint32{64, 512},
			wantStatus: statusError,
			wantStep:   stepFailed,
			wantData:   []uint32{64},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			step := runFrameSizePlan(t, tc.recorder, tc.testType, tc.config, tc.wantStatus)
			if got := tc.recorder.measured(); !reflect.DeepEqual(got, tc.wantSizes) {
				t.Errorf("measured sizes = %v, want %v", got, tc.wantSizes)
			}
			if step.Status != tc.wantStep {
				t.Errorf("step status = %q, want %q", step.Status, tc.wantStep)
			}
			if got := reportedSizes(t, step); !reflect.DeepEqual(got, tc.wantData) {
				t.Errorf("reported sizes = %v, want %v", got, tc.wantData)
			}
		})
	}
}

// runFrameSizePlan runs a one-step plan against the recorder until the run
// reaches wantStatus and returns the step as the plan recorded it.
func runFrameSizePlan(
	t *testing.T,
	recorder *sizeRecorder,
	testType string,
	config *TestConfig,
	wantStatus string,
) RunPlanStep {
	t.Helper()
	s := newTestServer(t)
	s.executorResolver = func(string) (executorFactory, bool) {
		return func(string) (testExecutor, error) { return recorder, nil }, true
	}
	plan, err := newRunPlan("", TestStartRequest{
		Peer:  "198.51.100.9",
		Tests: []TestStepRequest{{TestType: testType, Config: config}},
	})
	if err != nil {
		t.Fatalf("newRunPlan: %v", err)
	}
	runID, beginErr := s.beginRunPlan(plan)
	if beginErr != nil {
		t.Fatalf("beginRunPlan: %v", beginErr)
	}
	s.startRunPlan(runID, "lo", plan.ID)
	waitForTestStatus(t, s, wantStatus)
	s.statsMu.RLock()
	defer s.statsMu.RUnlock()
	return s.runPlan.Steps[0]
}

func reportedSizes(t *testing.T, step RunPlanStep) []uint32 {
	t.Helper()
	if step.Result == nil {
		t.Fatal("the step has no result")
	}
	perSize, ok := step.Result.Data.([]frameSizeResult)
	if !ok {
		t.Fatalf("step data = %T, want one result per frame size", step.Result.Data)
	}
	sizes := make([]uint32, len(perSize))
	for i, r := range perSize {
		sizes[i] = r.FrameSize
	}
	return sizes
}

// A stop between two sizes ends the run there: the next size is a new
// measurement the operator no longer wants.
func TestRunPlanStopBetweenFrameSizesMeasuresNoMore(t *testing.T) {
	s := newTestServer(t)
	recorder := &sizeRecorder{}
	recorder.after = func() {
		s.statsMu.Lock()
		s.testRunID++
		s.testStatus = statusStopped
		s.statsMu.Unlock()
	}
	s.executorResolver = func(string) (executorFactory, bool) {
		return func(string) (testExecutor, error) { return recorder, nil }, true
	}
	plan, err := newRunPlan("", TestStartRequest{
		Peer: "198.51.100.9",
		Tests: []TestStepRequest{{TestType: "rfc2544_throughput", Config: &TestConfig{
			RFC2544: &RFC2544TestConfig{Duration: 1, Trials: 1, FrameSizes: []uint32{64, 512, 1518}},
		}}},
	})
	if err != nil {
		t.Fatalf("newRunPlan: %v", err)
	}
	runID, beginErr := s.beginRunPlan(plan)
	if beginErr != nil {
		t.Fatalf("beginRunPlan: %v", beginErr)
	}
	s.runTestPlan(runID, "lo")

	if got := recorder.measured(); !reflect.DeepEqual(got, []uint32{64}) {
		t.Errorf("measured sizes = %v, want [64]", got)
	}
}

// The estimate and the run count the same executions: one per selected size,
// and one when the step names no sizes.
func TestEstimateStepSecondsCountsEverySize(t *testing.T) {
	tests := []struct {
		name  string
		sizes []uint32
		want  int64
	}{
		{name: "seven sizes", sizes: []uint32{64, 128, 256, 512, 1024, 1280, 1518}, want: 7 * 12 * 2},
		{name: "one size", sizes: []uint32{1518}, want: 12 * 2},
		{name: "no sizes runs once at the module default", sizes: nil, want: 12 * 2},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			step := RunPlanStep{
				TestType: "rfc2544_throughput",
				Config: &TestConfig{RFC2544: &RFC2544TestConfig{
					Duration: 10, Warmup: 2, Trials: 2, FrameSizes: tc.sizes,
				}},
			}
			got := estimateStepSeconds(step)
			if got == nil || *got != tc.want {
				t.Fatalf("estimate = %v, want %d", got, tc.want)
			}
		})
	}
}
