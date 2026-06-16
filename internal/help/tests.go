/*
 * The Stem - Test Documentation
 *
 * Comprehensive documentation for all 27 test types across 7 categories.
 */

package help

// GetAllTests returns help content for all tests.
func GetAllTests() map[string]TestHelp {
	return map[string]TestHelp{
		// RFC 2544 Tests
		TestTypeThroughput: rfc2544Throughput(),
		TestTypeLatency:    rfc2544Latency(),
		TestTypeFrameLoss:  rfc2544FrameLoss(),
		"back_to_back":     rfc2544BackToBack(),
		"system_recovery":  rfc2544SystemRecovery(),
		"reset":            rfc2544Reset(),

		// Y.1564 Tests
		"y1564_config":      y1564Config(),
		"y1564_performance": y1564Performance(),
		"y1564_full":        y1564Full(),

		// RFC 2889 Tests
		"forwarding":    rfc2889Forwarding(),
		"address_cache": rfc2889AddressCache(),
		"learning_rate": rfc2889LearningRate(),
		"broadcast":     rfc2889Broadcast(),
		"congestion":    rfc2889Congestion(),

		// RFC 6349 Tests
		"tcp_throughput": rfc6349TCPThroughput(),
		"path_analysis":  rfc6349PathAnalysis(),

		// Y.1731 Tests
		"frame_delay":      y1731FrameDelay(),
		"y1731_frame_loss": y1731FrameLoss(),
		"synthetic_loss":   y1731SyntheticLoss(),
		"loopback":         y1731Loopback(),

		// MEF Tests
		"mef_config":      mefConfig(),
		"mef_performance": mefPerformance(),
		"mef_full":        mefFull(),

		// TSN Tests
		"gate_timing":       tsnGateTiming(),
		"traffic_isolation": tsnTrafficIsolation(),
		"scheduled_latency": tsnScheduledLatency(),
		"tsn_full":          tsnFull(),
	}
}
