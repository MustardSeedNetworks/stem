// Copyright (c) 2025 Mustard Seed Networks. All rights reserved.

// Package tui provides a terminal user interface for the Test Master.
//
// Uses tview/tcell for real-time test progress display, showing current
// test status, throughput metrics, and test results.
package tui

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// UI layout constants.
const (
	progressBarWidth      = 50
	nsToUsConversion      = 1000
	topRowFlexProportion  = 3
	progressBarFlexHeight = 3
	statusBarFlexHeight   = 1
	resultsFlexProportion = 2
	fullPercent           = 100.0
	y1564DefaultStepCount = 4
	percentMultiplier     = 100.0
)

// Table column indices for RFC 2544 results.
const (
	colRFC2544FrameSize = 0
	colRFC2544RatePct   = 1
	colRFC2544RateMbps  = 2
	colRFC2544LossPct   = 3
	colRFC2544Latency   = 4
)

// Table column indices for Y.1564 results.
const (
	colY1564Service = 0
	colY1564Phase   = 1
	colY1564CIR     = 2
	colY1564FLR     = 3
	colY1564FD      = 4
	colY1564FDV     = 5
	colY1564Result  = 6
)

// TestType for display.
type TestType string

// Test type display name constants.
const (
	// TestThroughput is RFC 2544 throughput test display name.
	TestThroughput TestType = "Throughput"
	// TestLatency is RFC 2544 latency test display name.
	TestLatency TestType = "Latency"
	// TestFrameLoss is RFC 2544 frame loss test display name.
	TestFrameLoss TestType = "Frame Loss"
	// TestBackToBack is RFC 2544 back-to-back test display name.
	TestBackToBack TestType = "Back-to-Back"
	// TestSystemRecovery is RFC 2544 system recovery test display name.
	TestSystemRecovery TestType = "System Recovery"
	// TestReset is RFC 2544 reset test display name.
	TestReset TestType = "Reset"

	// TestY1564Config is Y.1564 config test display name.
	TestY1564Config TestType = "Y.1564 Config"
	// TestY1564Perf is Y.1564 performance test display name.
	TestY1564Perf TestType = "Y.1564 Perf"
	// TestY1564Full is Y.1564 full test display name.
	TestY1564Full TestType = "Y.1564 Full"

	// TestRFC2889Forwarding is RFC 2889 forwarding test display name.
	TestRFC2889Forwarding TestType = "RFC2889 Forwarding"
	// TestRFC2889Caching is RFC 2889 caching test display name.
	TestRFC2889Caching TestType = "RFC2889 Caching"
	// TestRFC2889Learning is RFC 2889 learning test display name.
	TestRFC2889Learning TestType = "RFC2889 Learning"
	// TestRFC2889Broadcast is RFC 2889 broadcast test display name.
	TestRFC2889Broadcast TestType = "RFC2889 Broadcast"
	// TestRFC2889Congestion is RFC 2889 congestion test display name.
	TestRFC2889Congestion TestType = "RFC2889 Congestion"

	// TestRFC6349Throughput is RFC 6349 throughput test display name.
	TestRFC6349Throughput TestType = "RFC6349 Throughput"
	// TestRFC6349Path is RFC 6349 path test display name.
	TestRFC6349Path TestType = "RFC6349 Path"

	// TestY1731Delay is Y.1731 delay test display name.
	TestY1731Delay TestType = "Y.1731 Delay"
	// TestY1731Loss is Y.1731 loss test display name.
	TestY1731Loss TestType = "Y.1731 Loss"
	// TestY1731SLM is Y.1731 SLM test display name.
	TestY1731SLM TestType = "Y.1731 SLM"
	// TestY1731Loopback is Y.1731 loopback test display name.
	TestY1731Loopback TestType = "Y.1731 Loopback"

	// TestMEFConfig is MEF config test display name.
	TestMEFConfig TestType = "MEF Config"
	// TestMEFPerf is MEF performance test display name.
	TestMEFPerf TestType = "MEF Perf"
	// TestMEFFull is MEF full test display name.
	TestMEFFull TestType = "MEF Full"

	// TestTSNTiming is TSN timing test display name.
	TestTSNTiming TestType = "TSN Timing"
	// TestTSNIsolation is TSN isolation test display name.
	TestTSNIsolation TestType = "TSN Isolation"
	// TestTSNLatency is TSN latency test display name.
	TestTSNLatency TestType = "TSN Latency"
	// TestTSNFull is TSN full test display name.
	TestTSNFull TestType = "TSN Full"
)

// Stats represents real-time test statistics.
type Stats struct {
	// Current test info.
	TestType  TestType
	FrameSize uint32
	Progress  float64
	State     string
	Iteration int
	MaxIter   int

	// Packet counters.
	TxPackets uint64
	TxBytes   uint64
	RxPackets uint64
	RxBytes   uint64

	// Rates.
	TxRate float64 // Mbps.
	RxRate float64 // Mbps.
	TxPPS  float64 // Packets per second.
	RxPPS  float64 // Packets per second.

	// Current trial.
	OfferedRate float64 // % of line rate.
	LossPct     float64

	// Latency (nanoseconds).
	LatencyMin float64
	LatencyMax float64
	LatencyAvg float64
	LatencyP99 float64

	// Uptime.
	StartTime time.Time
	Duration  time.Duration

	// Y.1564 specific fields.
	ServiceID    uint32  // Current service being tested.
	ServiceName  string  // Service name.
	CurrentStep  int     // Current step (1-4 for config test).
	TotalSteps   int     // Total steps.
	CIRMbps      float64 // Target CIR.
	FDMs         float64 // Frame Delay (ms).
	FDVMs        float64 // Frame Delay Variation (ms).
	FLRPct       float64 // Frame Loss Ratio (%).
	FDThreshold  float64 // FD SLA threshold.
	FDVThreshold float64 // FDV SLA threshold.
	FLRThreshold float64 // FLR SLA threshold.
	FDPass       bool    // FD within SLA.
	FDVPass      bool    // FDV within SLA.
	FLRPass      bool    // FLR within SLA.
}

// Result represents a completed test result.
type Result struct {
	FrameSize    uint32
	MaxRatePct   float64
	MaxRateMbps  float64
	LossPct      float64
	LatencyAvgNs float64
	Timestamp    time.Time
}

// Y1564StepResult represents a Y.1564 configuration test step result.
type Y1564StepResult struct {
	Step           int
	OfferedRatePct float64
	FLRPct         float64
	FDMs           float64
	FDVMs          float64
	FLRPass        bool
	FDPass         bool
	FDVPass        bool
	StepPass       bool
}

// Y1564Result represents a completed Y.1564 test result.
type Y1564Result struct {
	ServiceID   uint32
	ServiceName string
	TestPhase   string // "Config" or "Perf".
	FrameSize   uint32
	CIRMbps     float64
	FLRPct      float64
	FDMs        float64
	FDVMs       float64
	FLRPass     bool
	FDPass      bool
	FDVPass     bool
	ServicePass bool
	Steps       []Y1564StepResult // For config test.
	Timestamp   time.Time
}

// App represents the TUI application.
type App struct {
	app         *tview.Application
	pages       *tview.Pages
	statsView   *tview.Table
	resultsView *tview.Table
	logView     *tview.TextView
	progressBar *tview.TextView
	statusBar   *tview.TextView

	stats        Stats
	results      []Result
	y1564Results []Y1564Result

	// Callbacks.
	OnStart  func()
	OnStop   func()
	OnCancel func()
	OnQuit   func()
}

// New creates a new TUI application.
func New() *App {
	a := &App{
		app:          tview.NewApplication(),
		pages:        tview.NewPages(),
		statsView:    nil,
		resultsView:  nil,
		logView:      nil,
		progressBar:  nil,
		statusBar:    nil,
		stats:        Stats{}, //nolint:exhaustruct // Stats fields default to zero values.
		results:      make([]Result, 0),
		y1564Results: make([]Y1564Result, 0),
		OnStart:      nil,
		OnStop:       nil,
		OnCancel:     nil,
		OnQuit:       nil,
	}
	a.build()
	return a
}

func (a *App) build() {
	// Stats panel (left side).
	a.statsView = tview.NewTable().
		SetBorders(false).
		SetSelectable(false, false)
	a.statsView.SetTitle(" Statistics ").SetBorder(true)
	a.initStatsView()

	// Results panel (right side).
	a.resultsView = tview.NewTable().
		SetBorders(true).
		SetSelectable(true, false)
	a.resultsView.SetTitle(" Results ").SetBorder(true)
	a.initResultsView()

	// Progress bar.
	a.progressBar = tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignCenter)
	a.progressBar.SetTitle(" Progress ").SetBorder(true)
	a.updateProgressBar(0)

	// Log view (bottom).
	a.logView = tview.NewTextView().
		SetDynamicColors(true).
		SetScrollable(true).
		SetChangedFunc(func() {
			a.app.Draw()
		})
	a.logView.SetTitle(" Log ").SetBorder(true)

	// Status bar.
	a.statusBar = tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignCenter)
	statusText := "[yellow]RFC2544 Test Master[white] | " +
		"[green]F1[white] Start | [red]F2[white] Stop | [blue]F10[white] Quit"
	a.statusBar.SetText(statusText)

	// Layout.
	topRow := tview.NewFlex().
		AddItem(a.statsView, 0, 1, false).
		AddItem(a.resultsView, 0, resultsFlexProportion, false)

	mainFlex := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(topRow, 0, topRowFlexProportion, false).
		AddItem(a.progressBar, progressBarFlexHeight, 0, false).
		AddItem(a.logView, 0, 1, false).
		AddItem(a.statusBar, statusBarFlexHeight, 0, false)

	a.pages.AddPage("main", mainFlex, true, true)

	// Key bindings.
	a.app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() { //nolint:exhaustive // Only handle specific keys.
		case tcell.KeyF1:
			if a.OnStart != nil {
				go a.OnStart()
			}
			return nil
		case tcell.KeyF2:
			if a.OnStop != nil {
				go a.OnStop()
			}
			return nil
		case tcell.KeyF10, tcell.KeyEscape:
			if a.OnQuit != nil {
				a.OnQuit()
			}
			a.app.Stop()
			return nil
		case tcell.KeyCtrlC:
			if a.OnCancel != nil {
				a.OnCancel()
			}
			return nil
		default:
			return event
		}
	})

	a.app.SetRoot(a.pages, true)
}

func (a *App) initStatsView() {
	labels := []string{
		"Test Type:",
		"Frame Size:",
		"State:",
		"Progress:",
		"Iteration:",
		"",
		"TX Packets:",
		"TX Rate:",
		"RX Packets:",
		"RX Rate:",
		"",
		"Offered Rate:",
		"Frame Loss:",
		"",
		"Latency Min:",
		"Latency Avg:",
		"Latency Max:",
		"Latency P99:",
		"",
		"Duration:",
	}

	for i, label := range labels {
		a.statsView.SetCell(i, 0, tview.NewTableCell(label).
			SetTextColor(tcell.ColorYellow).
			SetAlign(tview.AlignRight))
		a.statsView.SetCell(i, 1, tview.NewTableCell("-").
			SetTextColor(tcell.ColorWhite).
			SetAlign(tview.AlignLeft))
	}
}

func (a *App) initResultsView() {
	headers := []string{"Frame Size", "Max Rate %", "Rate Mbps", "Loss %", "Latency Avg"}
	for i, h := range headers {
		a.resultsView.SetCell(0, i, tview.NewTableCell(h).
			SetTextColor(tcell.ColorYellow).
			SetAlign(tview.AlignCenter).
			SetSelectable(false))
	}
}

// UpdateStats updates the statistics display.
func (a *App) UpdateStats(s Stats) {
	a.stats = s
	a.app.QueueUpdateDraw(func() {
		// Check if this is a Y.1564 or MEF test type (SLA-based display).
		isSLATest := s.TestType == TestY1564Config || s.TestType == TestY1564Perf || s.TestType == TestY1564Full ||
			s.TestType == TestMEFConfig || s.TestType == TestMEFPerf || s.TestType == TestMEFFull

		if isSLATest {
			a.updateY1564Stats(s)
		} else {
			a.updateRFC2544Stats(s)
		}

		a.updateProgressBar(s.Progress)
	})
}

// updateRFC2544Stats updates the display for RFC 2544 tests.
func (a *App) updateRFC2544Stats(s Stats) {
	values := []string{
		string(s.TestType),
		fmt.Sprintf("%d bytes", s.FrameSize),
		s.State,
		fmt.Sprintf("%.1f%%", s.Progress),
		fmt.Sprintf("%d / %d", s.Iteration, s.MaxIter),
		"",
		strconv.FormatUint(s.TxPackets, 10),
		fmt.Sprintf("%.2f Mbps (%.0f pps)", s.TxRate, s.TxPPS),
		strconv.FormatUint(s.RxPackets, 10),
		fmt.Sprintf("%.2f Mbps (%.0f pps)", s.RxRate, s.RxPPS),
		"",
		fmt.Sprintf("%.2f%%", s.OfferedRate),
		fmt.Sprintf("%.4f%%", s.LossPct),
		"",
		fmt.Sprintf("%.2f us", s.LatencyMin/nsToUsConversion),
		fmt.Sprintf("%.2f us", s.LatencyAvg/nsToUsConversion),
		fmt.Sprintf("%.2f us", s.LatencyMax/nsToUsConversion),
		fmt.Sprintf("%.2f us", s.LatencyP99/nsToUsConversion),
		"",
		s.Duration.Round(time.Second).String(),
	}

	for i, v := range values {
		a.statsView.SetCell(i, 1, tview.NewTableCell(v).
			SetTextColor(tcell.ColorWhite).
			SetAlign(tview.AlignLeft))
	}
}

// updateY1564Stats updates the display for Y.1564 tests.
func (a *App) updateY1564Stats(s Stats) {
	// Update labels for Y.1564.
	y1564Labels := []string{
		"Test Type:",
		"Service:",
		"State:",
		"Progress:",
		"Step:",
		"",
		"TX Packets:",
		"TX Rate:",
		"RX Packets:",
		"RX Rate:",
		"",
		"CIR Target:",
		"Offered Rate:",
		"",
		"Frame Loss:",
		"Frame Delay:",
		"Frame Delay Var:",
		"",
		"SLA Status:",
		"Duration:",
	}

	for i, label := range y1564Labels {
		a.statsView.SetCell(i, 0, tview.NewTableCell(label).
			SetTextColor(tcell.ColorYellow).
			SetAlign(tview.AlignRight))
	}

	// Format pass/fail indicators.
	flrStatus := formatPassFail(s.FLRPass, fmt.Sprintf("%.4f%% (<=%.4f%%)", s.FLRPct, s.FLRThreshold))
	fdStatus := formatPassFail(s.FDPass, fmt.Sprintf("%.2f ms (<=%.2f ms)", s.FDMs, s.FDThreshold))
	fdvStatus := formatPassFail(s.FDVPass, fmt.Sprintf("%.2f ms (<=%.2f ms)", s.FDVMs, s.FDVThreshold))

	// Overall SLA status.
	slaStatus := "[green]PASS"
	if !s.FLRPass || !s.FDPass || !s.FDVPass {
		slaStatus = "[red]FAIL"
	}

	serviceName := s.ServiceName
	if serviceName == "" {
		serviceName = fmt.Sprintf("Service %d", s.ServiceID)
	}

	stepInfo := "-"
	if s.TotalSteps > 0 {
		stepInfo = fmt.Sprintf("%d / %d", s.CurrentStep, s.TotalSteps)
	}

	values := []string{
		string(s.TestType),
		serviceName,
		s.State,
		fmt.Sprintf("%.1f%%", s.Progress),
		stepInfo,
		"",
		strconv.FormatUint(s.TxPackets, 10),
		fmt.Sprintf("%.2f Mbps (%.0f pps)", s.TxRate, s.TxPPS),
		strconv.FormatUint(s.RxPackets, 10),
		fmt.Sprintf("%.2f Mbps (%.0f pps)", s.RxRate, s.RxPPS),
		"",
		fmt.Sprintf("%.2f Mbps", s.CIRMbps),
		fmt.Sprintf("%.2f%%", s.OfferedRate),
		"",
		flrStatus,
		fdStatus,
		fdvStatus,
		"",
		slaStatus,
		s.Duration.Round(time.Second).String(),
	}

	for i, v := range values {
		a.statsView.SetCell(i, 1, tview.NewTableCell(v).
			SetTextColor(tcell.ColorWhite).
			SetAlign(tview.AlignLeft))
	}
}

// formatPassFail returns a colored string based on pass/fail status.
func formatPassFail(pass bool, value string) string {
	if pass {
		return fmt.Sprintf("[green]%s [OK]", value)
	}
	return fmt.Sprintf("[red]%s [X]", value)
}

// AddResult adds a test result to the results table.
func (a *App) AddResult(r Result) {
	a.results = append(a.results, r)
	a.app.QueueUpdateDraw(func() {
		row := len(a.results)
		a.resultsView.SetCell(row, colRFC2544FrameSize,
			tview.NewTableCell(strconv.FormatUint(uint64(r.FrameSize), 10)).SetAlign(tview.AlignCenter))
		a.resultsView.SetCell(row, colRFC2544RatePct,
			tview.NewTableCell(fmt.Sprintf("%.2f", r.MaxRatePct)).SetAlign(tview.AlignCenter))
		a.resultsView.SetCell(row, colRFC2544RateMbps,
			tview.NewTableCell(fmt.Sprintf("%.2f", r.MaxRateMbps)).SetAlign(tview.AlignCenter))
		a.resultsView.SetCell(row, colRFC2544LossPct,
			tview.NewTableCell(fmt.Sprintf("%.4f", r.LossPct)).SetAlign(tview.AlignCenter))
		a.resultsView.SetCell(row, colRFC2544Latency,
			tview.NewTableCell(fmt.Sprintf("%.2f us", r.LatencyAvgNs/nsToUsConversion)).SetAlign(tview.AlignCenter))
	})
}

// AddY1564Result adds a Y.1564 test result to the results table.
func (a *App) AddY1564Result(r Y1564Result) {
	a.y1564Results = append(a.y1564Results, r)
	a.app.QueueUpdateDraw(func() {
		// If this is the first Y.1564 result, reinitialize the results view with Y.1564 headers.
		if len(a.y1564Results) == 1 {
			a.resultsView.Clear()
			a.initY1564ResultsView()
		}

		row := len(a.y1564Results)

		// Service name.
		serviceName := r.ServiceName
		if serviceName == "" {
			serviceName = fmt.Sprintf("Svc %d", r.ServiceID)
		}

		// Pass/Fail with color.
		passText := "[green]PASS"
		if !r.ServicePass {
			passText = "[red]FAIL"
		}

		a.resultsView.SetCell(row, colY1564Service,
			tview.NewTableCell(serviceName).SetAlign(tview.AlignCenter))
		a.resultsView.SetCell(row, colY1564Phase,
			tview.NewTableCell(r.TestPhase).SetAlign(tview.AlignCenter))
		a.resultsView.SetCell(row, colY1564CIR,
			tview.NewTableCell(fmt.Sprintf("%.2f", r.CIRMbps)).SetAlign(tview.AlignCenter))
		a.resultsView.SetCell(row, colY1564FLR,
			tview.NewTableCell(fmt.Sprintf("%.4f%%", r.FLRPct)).SetAlign(tview.AlignCenter))
		a.resultsView.SetCell(row, colY1564FD,
			tview.NewTableCell(fmt.Sprintf("%.2f ms", r.FDMs)).SetAlign(tview.AlignCenter))
		a.resultsView.SetCell(row, colY1564FDV,
			tview.NewTableCell(fmt.Sprintf("%.2f ms", r.FDVMs)).SetAlign(tview.AlignCenter))
		a.resultsView.SetCell(row, colY1564Result,
			tview.NewTableCell(passText).SetAlign(tview.AlignCenter))
	})
}

// initY1564ResultsView initializes the results view with Y.1564 headers.
func (a *App) initY1564ResultsView() {
	headers := []string{"Service", "Phase", "CIR Mbps", "FLR %", "FD ms", "FDV ms", "Result"}
	for i, h := range headers {
		a.resultsView.SetCell(0, i, tview.NewTableCell(h).
			SetTextColor(tcell.ColorYellow).
			SetAlign(tview.AlignCenter).
			SetSelectable(false))
	}
}

// Logf adds a message to the log view.
func (a *App) Logf(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	timestamp := time.Now().Format("15:04:05")
	a.app.QueueUpdateDraw(func() {
		_, _ = fmt.Fprintf(a.logView, "[gray]%s[white] %s\n", timestamp, msg)
		a.logView.ScrollToEnd()
	})
}

// LogInfof logs an info message.
func (a *App) LogInfof(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	timestamp := time.Now().Format("15:04:05")
	a.app.QueueUpdateDraw(func() {
		_, _ = fmt.Fprintf(a.logView, "[gray]%s [green][INFO][white] %s\n", timestamp, msg)
		a.logView.ScrollToEnd()
	})
}

// LogWarnf logs a warning message.
func (a *App) LogWarnf(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	timestamp := time.Now().Format("15:04:05")
	a.app.QueueUpdateDraw(func() {
		_, _ = fmt.Fprintf(a.logView, "[gray]%s [yellow][WARN][white] %s\n", timestamp, msg)
		a.logView.ScrollToEnd()
	})
}

// LogErrorf logs an error message.
func (a *App) LogErrorf(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	timestamp := time.Now().Format("15:04:05")
	a.app.QueueUpdateDraw(func() {
		_, _ = fmt.Fprintf(a.logView, "[gray]%s [red][ERROR][white] %s\n", timestamp, msg)
		a.logView.ScrollToEnd()
	})
}

func (a *App) updateProgressBar(pct float64) {
	width := progressBarWidth
	filled := int(pct / fullPercent * float64(width))
	filled = min(filled, width)

	var bar strings.Builder
	for i := range width {
		if i < filled {
			bar.WriteString("[green]#")
		} else {
			bar.WriteString("[gray]-")
		}
	}
	a.progressBar.SetText(fmt.Sprintf("%s[white] %.1f%%", bar.String(), pct))
}

// SetStatus updates the status bar.
func (a *App) SetStatus(msg string) {
	a.app.QueueUpdateDraw(func() {
		a.statusBar.SetText(msg)
	})
}

// Run starts the TUI application.
func (a *App) Run() error {
	err := a.app.Run()
	if err != nil {
		return fmt.Errorf("TUI app run failed: %w", err)
	}
	return nil
}

// Stop stops the TUI application.
func (a *App) Stop() {
	a.app.Stop()
}

// ClearResults clears the results table.
func (a *App) ClearResults() {
	a.results = a.results[:0]
	a.y1564Results = a.y1564Results[:0]
	a.app.QueueUpdateDraw(func() {
		a.resultsView.Clear()
		a.initResultsView()
	})
}

// SwitchToY1564View switches the results view to Y.1564 format.
func (a *App) SwitchToY1564View() {
	a.y1564Results = a.y1564Results[:0]
	a.app.QueueUpdateDraw(func() {
		a.resultsView.Clear()
		a.initY1564ResultsView()
	})
}

// SwitchToRFC2544View switches the results view to RFC 2544 format.
func (a *App) SwitchToRFC2544View() {
	a.results = a.results[:0]
	a.app.QueueUpdateDraw(func() {
		a.resultsView.Clear()
		a.initResultsView()
	})
}
