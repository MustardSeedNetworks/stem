// SPDX-License-Identifier: BUSL-1.1

package api

import (
	"testing"
	"time"
)

func TestRunPlanDescribe(t *testing.T) {
	twoSteps := []RunPlanStep{
		{TestType: "rfc2544_throughput", Module: "benchmark", Status: stepPassed},
		{TestType: "y1731_delay", Module: "servicetest", Status: stepRunning},
	}
	estimate := int64(120)
	ninetyRemaining := int64(90)

	tests := []struct {
		name string
		plan *runPlan
		want Stats
	}{
		{
			name: "no plan leaves the snapshot untouched",
			plan: nil,
			want: Stats{},
		},
		{
			name: "accepted but not yet started reports totals only",
			plan: &runPlan{ID: "stem-1", Steps: twoSteps, Current: -1},
			want: Stats{SuiteID: "stem-1", StepsTotal: 2},
		},
		{
			name: "a finished step reports the recorded elapsed time and no phase",
			plan: &runPlan{
				ID: "stem-2", Steps: []RunPlanStep{{TestType: "reflect", Status: stepPassed}},
				Current: 0, Complete: 1, StepElapsedSec: 7,
			},
			want: Stats{
				SuiteID: "stem-2", StepsTotal: 1, StepsComplete: 1,
				CurrentStep: 1, ElapsedSeconds: 7,
			},
		},
		{
			name: "a running step without an estimate is indeterminate",
			plan: &runPlan{
				ID: "stem-3", Steps: twoSteps, Current: 1, Complete: 1,
				StepStarted: time.Now().Add(-30 * time.Second),
			},
			want: Stats{
				SuiteID: "stem-3", StepsTotal: 2, StepsComplete: 1, CurrentStep: 2,
				Phase: "Executing y1731_delay", ElapsedSeconds: 30,
			},
		},
		{
			name: "a running step with an estimate reports the remainder",
			plan: &runPlan{
				ID: "stem-4", Steps: twoSteps, Current: 1, Complete: 1,
				StepStarted: time.Now().Add(-30 * time.Second), StepEstimateSec: &estimate,
			},
			want: Stats{
				SuiteID: "stem-4", StepsTotal: 2, StepsComplete: 1, CurrentStep: 2,
				Phase: "Executing y1731_delay", ElapsedSeconds: 30,
				EstimatedRemainingSeconds: &ninetyRemaining,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var got Stats
			tc.plan.describe(&got)
			assertStatsProjection(t, got, tc.want)
			if tc.plan != nil && len(got.Steps) != len(tc.plan.Steps) {
				t.Errorf("steps = %d, want %d", len(got.Steps), len(tc.plan.Steps))
			}
		})
	}
}

// TestRunPlanDescribeClampsOverrunEstimate pins the floor: a step that outlives
// its estimate reports zero remaining, never a negative countdown.
func TestRunPlanDescribeClampsOverrunEstimate(t *testing.T) {
	estimate := int64(5)
	plan := &runPlan{
		ID:      "stem-5",
		Steps:   []RunPlanStep{{TestType: "rfc2544_throughput", Status: stepRunning}},
		Current: 0, StepStarted: time.Now().Add(-60 * time.Second), StepEstimateSec: &estimate,
	}
	var got Stats
	plan.describe(&got)
	if got.EstimatedRemainingSeconds == nil || *got.EstimatedRemainingSeconds != 0 {
		t.Fatalf("estimatedRemainingSeconds = %v, want 0", got.EstimatedRemainingSeconds)
	}
}

func assertStatsProjection(t *testing.T, got, want Stats) {
	t.Helper()
	if got.SuiteID != want.SuiteID {
		t.Errorf("suiteId = %q, want %q", got.SuiteID, want.SuiteID)
	}
	if got.StepsTotal != want.StepsTotal || got.StepsComplete != want.StepsComplete {
		t.Errorf(
			"stepsTotal/stepsComplete = %d/%d, want %d/%d",
			got.StepsTotal, got.StepsComplete, want.StepsTotal, want.StepsComplete,
		)
	}
	if got.CurrentStep != want.CurrentStep {
		t.Errorf("currentStep = %d, want %d", got.CurrentStep, want.CurrentStep)
	}
	if got.Phase != want.Phase {
		t.Errorf("phase = %q, want %q", got.Phase, want.Phase)
	}
	if got.ElapsedSeconds != want.ElapsedSeconds {
		t.Errorf("elapsedSeconds = %d, want %d", got.ElapsedSeconds, want.ElapsedSeconds)
	}
	switch {
	case want.EstimatedRemainingSeconds == nil && got.EstimatedRemainingSeconds != nil:
		t.Errorf("estimatedRemainingSeconds = %d, want nil", *got.EstimatedRemainingSeconds)
	case want.EstimatedRemainingSeconds != nil && got.EstimatedRemainingSeconds == nil:
		t.Errorf("estimatedRemainingSeconds = nil, want %d", *want.EstimatedRemainingSeconds)
	case want.EstimatedRemainingSeconds != nil &&
		*got.EstimatedRemainingSeconds != *want.EstimatedRemainingSeconds:
		t.Errorf(
			"estimatedRemainingSeconds = %d, want %d",
			*got.EstimatedRemainingSeconds, *want.EstimatedRemainingSeconds,
		)
	}
}
