/*
 * Stem - Test Documentation
 *
 * Comprehensive documentation for all 27 test types across 7 categories.
 */

package help

// GetAllTests returns help content for all tests.
func GetAllTests() map[string]TestHelp {
	return map[string]TestHelp{
		// RFC 2544 Tests
		TestTypeThroughput:        rfc2544Throughput(),
		TestTypeLatency:           rfc2544Latency(),
		TestTypeFrameLoss:         rfc2544FrameLoss(),
		"rfc2544_back_to_back":    rfc2544BackToBack(),
		"rfc2544_system_recovery": rfc2544SystemRecovery(),
		"rfc2544_reset":           rfc2544Reset(),

		// Y.1564 Tests
		"y1564_config": y1564Config(),
		"y1564_perf":   y1564Performance(),
		"y1564":        y1564Full(),

		// RFC 2889 Tests
		"rfc2889_forwarding": rfc2889Forwarding(),
		"rfc2889_caching":    rfc2889AddressCache(),
		"rfc2889_learning":   rfc2889LearningRate(),
		"rfc2889_broadcast":  rfc2889Broadcast(),
		"rfc2889_congestion": rfc2889Congestion(),

		// RFC 6349 Tests
		"rfc6349_throughput": rfc6349TCPThroughput(),
		"rfc6349_path":       rfc6349PathAnalysis(),

		// Y.1731 Tests
		"y1731_delay":    y1731FrameDelay(),
		"y1731_loss":     y1731FrameLoss(),
		"y1731_slm":      y1731SyntheticLoss(),
		"y1731_loopback": y1731Loopback(),

		// MEF Tests
		"mef_config": mefConfig(),
		"mef_perf":   mefPerformance(),
		"mef":        mefFull(),

		// TSN Tests
		"tsn_timing":    tsnGateTiming(),
		"tsn_isolation": tsnTrafficIsolation(),
		"tsn_latency":   tsnScheduledLatency(),
		"tsn":           tsnFull(),
	}
}
