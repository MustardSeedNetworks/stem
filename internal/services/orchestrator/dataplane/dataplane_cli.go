//go:build cgo && linux

package dataplane

import "errors"

// New creates a new RFC2544 context with configuration.
func New(cfg Config) (*Context, error) {
	ctx, err := NewContext(cfg.Interface)
	if err != nil {
		return nil, err
	}

	if configureErr := ctx.Configure(&cfg); configureErr != nil {
		ctx.Close()
		return nil, configureErr
	}

	// Store config in context for later use
	ctx.config = cfg

	return ctx, nil
}

// SetFrameSize sets the frame size for subsequent tests.
func (c *Context) SetFrameSize(frameSize uint32) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.frameSize = frameSize
}

// RunThroughputTest runs a throughput test and returns a CLI-friendly result.
func (c *Context) RunThroughputTest() (*ThroughputResultCLI, error) {
	results, err := c.runThroughputTestInternal(c.frameSize)
	if err != nil {
		return nil, err
	}
	if len(results) == 0 {
		return nil, errors.New("no results")
	}

	r := results[0]
	if err = throughputVerdict(r); err != nil {
		return nil, err
	}

	return &ThroughputResultCLI{
		FrameSize:        r.FrameSize,
		MaxRatePct:       r.MaxRatePct,
		MaxRateMbps:      r.MaxRateMbps,
		MaxRatePPS:       r.MaxRatePps,
		OfferedRatePct:   r.OfferedRatePct,
		GeneratorLimited: r.GeneratorLimited,
		Iterations:       r.Iterations,
		Latency:          r.Latency,
	}, nil
}

// RunLatencyTest determines the frame size's throughput, then measures latency
// at each load level, a percentage of that throughput.
func (c *Context) RunLatencyTest(loadLevels []float64) ([]LatencyResultCLI, error) {
	throughput, err := c.runThroughputTestInternal(c.frameSize)
	if err != nil {
		return nil, err
	}
	if len(throughput) == 0 {
		return nil, errors.New("latency test needs a throughput rate: no throughput results")
	}
	rates, err := latencyOfferedRates(throughput[0], loadLevels)
	if err != nil {
		return nil, err
	}

	var results []LatencyResultCLI
	for i, load := range loadLevels {
		result, trialErr := c.runLatencyTestInternal(c.frameSize, rates[i])
		if trialErr != nil {
			continue
		}
		results = append(results, LatencyResultCLI{
			FrameSize:      c.frameSize,
			LoadPct:        load,
			OfferedRatePct: rates[i],
			Latency:        result.Latency,
		})
	}

	if len(results) == 0 {
		return nil, errors.New("no latency results")
	}
	if err = latencyVerdict(results); err != nil {
		return nil, err
	}

	return results, nil
}

// RunFrameLossTest runs a frame loss test with stepped load.
func (c *Context) RunFrameLossTest(_, _, _ float64) ([]FrameLossResultCLI, error) {
	results, err := c.runFrameLossTestInternal(c.frameSize)
	if err != nil {
		return nil, err
	}
	if err = frameLossVerdict(results); err != nil {
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

// RunBackToBackTest runs a back-to-back burst test.
func (c *Context) RunBackToBackTest(_ uint64, _ uint32) (*BackToBackResultCLI, error) {
	result, err := c.runBackToBackTestInternal(c.frameSize)
	if err != nil {
		return nil, err
	}
	if err = backToBackVerdict(*result); err != nil {
		return nil, err
	}

	return &BackToBackResultCLI{
		FrameSize:       c.frameSize,
		MaxBurstFrames:  result.MaxBurst,
		BurstDurationUs: uint64(result.BurstDuration),
		Trials:          result.Trials,
	}, nil
}
