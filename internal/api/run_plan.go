// SPDX-License-Identifier: BUSL-1.1

package api

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/MustardSeedNetworks/stem/internal/logging"
	"github.com/MustardSeedNetworks/stem/internal/services"
	"github.com/MustardSeedNetworks/stem/internal/services/modtypes"
)

const (
	stepPending = "pending"
	stepRunning = "running"
	stepPassed  = "passed"
	stepFailed  = "failed"
	stepSkipped = "skipped"
)

type runPlan struct {
	ID              string
	Peer            string
	PeerPort        uint16
	Steps           []RunPlanStep
	Current         int
	Complete        int
	StepStarted     time.Time
	StepElapsedSec  int64
	StepEstimateSec *int64
}

func newRunPlan(id string, request TestStartRequest) (*runPlan, error) {
	if len(request.Tests) == 0 {
		return nil, errors.New("select at least one test")
	}
	steps := make([]RunPlanStep, len(request.Tests))
	for i, stepRequest := range request.Tests {
		module := services.GetModuleForTest(stepRequest.TestType)
		if module == nil || !module.CanRun(stepRequest.TestType) {
			return nil, fmt.Errorf("unknown test type: %s", stepRequest.TestType)
		}
		steps[i] = RunPlanStep{
			TestType: stepRequest.TestType,
			Module:   module.Name(),
			Status:   stepPending,
			Config:   stepRequest.Config,
		}
	}
	peerPort := request.PeerPort
	if peerPort == 0 {
		peerPort = DefaultPortFilter
	}
	return &runPlan{ID: id, Peer: strings.TrimSpace(request.Peer), PeerPort: peerPort, Steps: steps, Current: -1}, nil
}

func (s *Server) runTestPlan(runID uint64, iface string) {
	s.statsMu.Lock()
	plan := s.runPlan
	s.statsMu.Unlock()
	if plan == nil {
		return
	}

	for index := range plan.Steps {
		step, started := s.startPlanStep(runID, index)
		if !started {
			return
		}
		result, err := s.executePlanStep(runID, iface, step)
		if !s.finishPlanStep(runID, index, result, err) {
			return
		}
	}
	s.finishRunPlan(runID)
}

func (s *Server) startPlanStep(runID uint64, index int) (RunPlanStep, bool) {
	s.statsMu.Lock()
	defer s.statsMu.Unlock()
	if s.testRunID != runID || s.runPlan == nil {
		return RunPlanStep{}, false
	}
	step := &s.runPlan.Steps[index]
	step.Status = stepRunning
	s.runPlan.Current = index
	s.runPlan.StepStarted = time.Now()
	s.runPlan.StepElapsedSec = 0
	s.runPlan.StepEstimateSec = estimateStepSeconds(*step)
	s.currentTest = step.TestType
	s.currentModule = step.Module
	s.testStatus = statusRunning
	return *step, true
}

func (s *Server) executePlanStep(
	runID uint64,
	iface string,
	step RunPlanStep,
) (*TestResultResponse, error) {
	resolver := s.testExecutorResolver()
	if resolver == nil {
		resolver = services.Factory
	}
	factory, ok := resolver(step.Module)
	if !ok {
		return nil, fmt.Errorf("executor not implemented for module: %s", step.Module)
	}
	exec, err := factory(iface)
	if err != nil {
		return nil, fmt.Errorf("create %s executor: %w", step.Module, err)
	}
	defer exec.Close()

	s.statsMu.Lock()
	if s.testRunID != runID {
		s.statsMu.Unlock()
		return nil, errors.New("run plan cancelled")
	}
	plan := s.runPlan
	s.activeTestExec = exec
	s.statsMu.Unlock()

	result, execErr := exec.Execute(
		step.TestType,
		plan.moduleConfig(iface, step),
	)
	if result == nil {
		return nil, execErr
	}
	response := &TestResultResponse{
		Status:   statusCompleted,
		TestType: step.TestType,
		Module:   step.Module,
		Success:  result.Success,
		Error:    result.Error,
		Data:     result.Data,
	}
	return response, execErr
}

func (p *runPlan) moduleConfig(iface string, step RunPlanStep) *modtypes.TestConfig {
	cfg := convertToModuleConfig(iface, step.TestType, step.Config)
	cfg.Peer = p.Peer
	cfg.PeerPort = p.PeerPort
	return cfg
}

func (s *Server) finishPlanStep(
	runID uint64,
	index int,
	result *TestResultResponse,
	err error,
) bool {
	s.statsMu.Lock()
	defer s.statsMu.Unlock()
	if s.testRunID != runID || s.runPlan == nil {
		return false
	}
	s.activeTestExec = nil
	step := &s.runPlan.Steps[index]
	s.runPlan.StepElapsedSec = int64(time.Since(s.runPlan.StepStarted).Seconds())
	if err != nil || result == nil || !result.Success {
		step.Status = stepFailed
		logCause := "test returned an unsuccessful result"
		if err != nil {
			logCause = err.Error()
			step.Error = "Test execution failed"
		} else if result != nil {
			step.Error = result.Error
		}
		step.Result = result
		for later := index + 1; later < len(s.runPlan.Steps); later++ {
			s.runPlan.Steps[later].Status = stepSkipped
		}
		s.testStatus = statusError
		s.currentTest = ""
		s.currentRunID = ""
		s.currentModule = ""
		s.testResult = s.planResult(false, "Run plan failed")
		logging.Error(
			"Run plan step failed",
			"suiteId",
			s.runPlan.ID,
			"testType",
			step.TestType,
			"error",
			logCause,
		)
		return false
	}
	step.Status = stepPassed
	step.Result = result
	s.runPlan.Complete++
	s.testResult = s.planResult(false, "Run plan in progress")
	return true
}

func (s *Server) cancelRunPlanLocked(message string) {
	if s.runPlan == nil {
		return
	}
	if s.runPlan.Current >= 0 && s.runPlan.Current < len(s.runPlan.Steps) {
		step := &s.runPlan.Steps[s.runPlan.Current]
		if step.Status == stepRunning {
			step.Status = statusCancelled
			s.runPlan.StepElapsedSec = int64(time.Since(s.runPlan.StepStarted).Seconds())
		}
	}
	for i := s.runPlan.Current + 1; i < len(s.runPlan.Steps); i++ {
		s.runPlan.Steps[i].Status = stepSkipped
	}
	s.testResult = s.planResult(false, message)
}

func (s *Server) finishRunPlan(runID uint64) {
	s.statsMu.Lock()
	defer s.statsMu.Unlock()
	if s.testRunID != runID || s.runPlan == nil {
		return
	}
	s.testStatus = statusCompleted
	s.currentTest = ""
	s.currentRunID = ""
	s.currentModule = ""
	s.activeTestExec = nil
	s.testResult = s.planResult(true, "Run plan completed")
	logging.Info("Run plan completed", "suiteId", s.runPlan.ID, "steps", len(s.runPlan.Steps))
}

// describe projects the plan onto a stats snapshot. A nil receiver is the
// no-plan case, so the caller does not repeat the check.
func (p *runPlan) describe(stats *Stats) {
	if p == nil {
		return
	}
	stats.SuiteID = p.ID
	stats.Steps = append([]RunPlanStep(nil), p.Steps...)
	stats.StepsTotal = len(p.Steps)
	stats.StepsComplete = p.Complete
	if p.Current < 0 {
		return
	}
	stats.CurrentStep = p.Current + 1
	stats.ElapsedSeconds = p.StepElapsedSec
	if p.Steps[p.Current].Status != stepRunning {
		return
	}
	stats.Phase = "Executing " + p.Steps[p.Current].TestType
	stats.ElapsedSeconds = int64(time.Since(p.StepStarted).Seconds())
	if p.StepEstimateSec != nil {
		remaining := max(0, *p.StepEstimateSec-stats.ElapsedSeconds)
		stats.EstimatedRemainingSeconds = &remaining
	}
}
