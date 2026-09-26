// SPDX-License-Identifier: BUSL-1.1

package api

func (s *Server) planResult(success bool, message string) *TestResultResponse {
	result := &TestResultResponse{
		Status:   s.testStatus,
		TestType: "run_plan",
		Module:   "orchestrator",
		Success:  success,
		Message:  message,
		SuiteID:  s.runPlan.ID,
		Steps:    append([]RunPlanStep(nil), s.runPlan.Steps...),
	}
	s.stampRunTimingLocked(result)
	return result
}

func estimateStepSeconds(step RunPlanStep) *int64 {
	if step.Config == nil || step.Config.RFC2544 == nil {
		return nil
	}
	config := step.Config.RFC2544
	runs := max(1, len(stepFrameSizes(step)))
	seconds := int64((config.Duration + config.Warmup) * config.Trials * runs)
	return &seconds
}
