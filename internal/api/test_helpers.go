// SPDX-License-Identifier: BUSL-1.1

package api

import (
	"github.com/MustardSeedNetworks/stem/internal/auth"
	"github.com/MustardSeedNetworks/stem/internal/license"
)

// TestExecutor is the public alias for the unexported testExecutor
// interface. Tests in api_test can implement this interface and inject
// instances via Server.UseTestExecutorResolver to swap out the real
// dataplane-backed executors.
type TestExecutor = testExecutor

// TestExecutorFactory creates a TestExecutor for the given interface name.
type TestExecutorFactory func(iface string) (TestExecutor, error)

// TestExecutorResolver maps a module name to a TestExecutorFactory.
type TestExecutorResolver func(moduleName string) (TestExecutorFactory, bool)

// UseTestExecutorResolver injects a custom executor resolver into the
// Server. Tests use this to substitute fake executors so the server's
// test-start path can be exercised end-to-end without invoking the real
// cgo dataplane (which requires raw-socket capabilities and is unsafe in
// CI runners).
//
// Passing nil restores the default factory.
func (s *Server) UseTestExecutorResolver(resolver TestExecutorResolver) {
	s.statsMu.Lock()
	defer s.statsMu.Unlock()
	if resolver == nil {
		s.executorResolver = nil
		return
	}
	s.executorResolver = func(moduleName string) (executorFactory, bool) {
		factory, ok := resolver(moduleName)
		if !ok {
			return nil, false
		}
		return func(iface string) (testExecutor, error) {
			exec, err := factory(iface)
			if err != nil {
				return nil, err
			}
			return exec, nil
		}, true
	}
}

// ResetTestStateForTest clears the server's transient test execution
// state (status, current test, current module, last result). Tests that
// drive multiple sequential test-start requests against the same Server
// instance use this to return to a clean baseline between requests.
func (s *Server) ResetTestStateForTest() {
	s.statsMu.Lock()
	exec := s.activeTestExec
	s.activeTestExec = nil
	s.testRunID++
	s.testStatus = ""
	s.currentTest = ""
	s.currentModule = ""
	s.testResult = nil
	s.runPlan = nil
	s.statsMu.Unlock()
	if cancellable, ok := exec.(interface{ Cancel() }); ok {
		cancellable.Cancel()
	}
}

// CSRFManagerForTest exposes the server's CSRFManager for tests that
// need to assert directly on token-lifecycle state (e.g. rotation on
// login, expiry handling).
func (s *Server) CSRFManagerForTest() *auth.CSRFManager {
	return s.csrfManager
}

// SessionIDFromJWTForTest is the test-side wrapper around the package-
// private sessionIDFromJWT helper used by handleAuthLogin. Exposing it
// here keeps the implementation private to api/ while still allowing
// CSRF-rotation tests to compare session IDs.
func SessionIDFromJWTForTest(token string) string {
	return sessionIDFromJWT(token)
}

// UseReflectorAvailabilityForTest swaps the platform-capability probe
// used by POST /api/v1/mode. The fn returns (available, reason);
// reason is ignored when available is true and surfaced verbatim in
// the 403 response message when available is false. Passing nil
// restores the default probe (real reflector dataplane availability).
//
// Tests use this to exercise the unsupported-platform branch of the
// mode handler without rebuilding under different cgo tags — the same
// pattern the executor resolver uses to avoid touching the real
// dataplane.
func (s *Server) UseReflectorAvailabilityForTest(fn func() (bool, string)) {
	if fn == nil {
		s.reflectorAvailability = nil
		return
	}
	s.reflectorAvailability = func() (bool, string) {
		return fn()
	}
}

// UseLicenseForTest swaps the server's license manager so a test can decide
// what the deployment is entitled to. Passing nil is the deployment whose
// manager could not be built at all, which grants the Free features and
// nothing else; a test exercising a paid capability installs a manager rather
// than depending on whatever activation state the developer's ~/.config/stem
// holds.
func (s *Server) UseLicenseForTest(mgr *license.Manager) {
	s.licenseManager = mgr
}

// FeatureForTestTypeForTest exposes the entitlement map so the catalog
// completeness test can assert, from outside the package, that every priced
// capability has a gate and every gate names a real catalog string.
func FeatureForTestTypeForTest(testType string) (string, bool) {
	return featureForTestType(testType)
}
