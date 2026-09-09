// SPDX-License-Identifier: BUSL-1.1

package api

func (s *Server) planResult(success bool, message string) *TestResultResponse {
	return &TestResultResponse{
		Status:   s.testStatus,
		TestType: "run_plan",
		Module:   "orchestrator",
		Success:  success,
		Message:  message,
		SuiteID:  s.runPlan.ID,
		Steps:    append([]RunPlanStep(nil), s.runPlan.Steps...),
	}
}

func estimateStepSeconds(step RunPlanStep) *int64 {
	if step.Config == nil || step.Config.RFC2544 == nil {
		return nil
	}
	config := step.Config.RFC2544
	seconds := int64((config.Duration + config.Warmup) * config.Trials * len(config.FrameSizes))
	return &seconds
}
