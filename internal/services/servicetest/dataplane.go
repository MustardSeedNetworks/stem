// SPDX-License-Identifier: BUSL-1.1

package servicetest

import "github.com/MustardSeedNetworks/stem/internal/services/orchestrator/dataplane"

// ServiceDataplane is the narrow interface the ServiceTest executor depends
// on: the Y.1564 and MEF runners plus the context lifecycle. The concrete
// *dataplane.Context satisfies it on both the cgo and the stub build, and
// unit tests substitute a fake to drive the executor with a chosen verdict.
type ServiceDataplane interface {
	Configure(cfg *dataplane.Config) error
	RunY1564ConfigTest(service *dataplane.Y1564Service) (*dataplane.Y1564ConfigResult, error)
	RunY1564PerfTest(service *dataplane.Y1564Service, durationSec uint32) (*dataplane.Y1564PerfResult, error)
	RunMEFConfigTest(cfg *dataplane.MEFConfig) (*dataplane.MEFConfigResult, error)
	RunMEFPerfTest(cfg *dataplane.MEFConfig) (*dataplane.MEFPerfResult, error)
	RunMEFFullTest(cfg *dataplane.MEFConfig) (*dataplane.MEFConfigResult, *dataplane.MEFPerfResult, error)
	Cancel()
	Close()
}
