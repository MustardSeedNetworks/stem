//go:build cgo && linux

package dataplane

import "fmt"

// New creates a new RFC2544 context with configuration
func New(cfg Config) (*Context, error) {
	ctx, err := NewContext(cfg.Interface)
	if err != nil {
		return nil, err
	}

	if err := ctx.Configure(&cfg); err != nil {
		ctx.Close()
		return nil, err
	}

	// Store config in context for later use
	ctx.config = cfg

	return ctx, nil
}

// SetFrameSize sets the frame size for subsequent tests
func (c *Context) SetFrameSize(frameSize uint32) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.frameSize = frameSize
}

// RunThroughputTestCLI runs throughput test and returns CLI-friendly result
func (c *Context) RunThroughputTest() (*ThroughputResultCLI, error) {
	results, err := c.runThroughputTestInternal(c.frameSize)
	if err != nil {
		return nil, err
	}
	if len(results) == 0 {
		return nil, fmt.Errorf("no results")
	}

	r := results[0]
	return &ThroughputResultCLI{
		FrameSize:   r.FrameSize,
		MaxRatePct:  r.MaxRatePct,
		MaxRateMbps: r.MaxRateMbps,
		MaxRatePPS:  r.MaxRatePps,
		Iterations:  r.Iterations,
		Latency:     r.Latency,
	}, nil
}

// RunLatencyTestCLI runs latency test at multiple load levels
func (c *Context) RunLatencyTest(loadLevels []float64) ([]LatencyResultCLI, error) {
	var results []LatencyResultCLI

	for _, load := range loadLevels {
		result, err := c.runLatencyTestInternal(c.frameSize, load)
		if err != nil {
			continue
		}
		results = append(results, LatencyResultCLI{
			FrameSize: c.frameSize,
			LoadPct:   load,
			Latency:   result.Latency,
		})
	}

	if len(results) == 0 {
		return nil, fmt.Errorf("no latency results")
	}

	return results, nil
}

// RunFrameLossTestCLI runs frame loss test with stepped load
func (c *Context) RunFrameLossTest(startPct, endPct, stepPct float64) ([]FrameLossResultCLI, error) {
	results, err := c.runFrameLossTestInternal(c.frameSize)
	if err != nil {
		return nil, err
	}

	var cliResults []FrameLossResultCLI
	for _, r := range results {
		cliResults = append(cliResults, FrameLossResultCLI{
			FrameSize:  c.frameSize,
			OfferedPct: r.OfferedRatePct,
			FramesTx:   r.FramesSent,
			FramesRx:   r.FramesRecv,
			LossPct:    r.LossPct,
		})
	}

	return cliResults, nil
}

// RunBackToBackTestCLI runs back-to-back burst test
func (c *Context) RunBackToBackTest(initialBurst uint64, trials uint32) (*BackToBackResultCLI, error) {
	result, err := c.runBackToBackTestInternal(c.frameSize)
	if err != nil {
		return nil, err
	}

	return &BackToBackResultCLI{
		FrameSize:       c.frameSize,
		MaxBurstFrames:  result.MaxBurst,
		BurstDurationUs: uint64(result.BurstDuration),
		Trials:          result.Trials,
	}, nil
}
