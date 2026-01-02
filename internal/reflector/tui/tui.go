// Copyright (c) 2025 Mustard Seed Networks. All rights reserved.

// Package tui provides the Terminal User Interface for the Reflector.
//
// Uses tview/tcell for real-time dashboard rendering with live packet
// statistics, interface status, and signature filter status.
package tui

import (
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/krisarmstrong/stem/internal/reflector/dataplane"
)

// Constants for TUI configuration and formatting.
const (
	statsFlexWeight     = 2 // Weight for stats panel in flex layout.
	tickerIntervalMs    = 500
	bitsPerByte         = 8.0
	megabitsPerSecDenom = 1000000.0
	billion             = 1000000000
	million             = 1000000
	thousand            = 1000
	terabyte            = 1099511627776
	gigabyte            = 1073741824
	megabyte            = 1048576
	kilobyte            = 1024
	secondsPerMinute    = 60
)

// App holds the TUI application state.
type App struct {
	dp        *dataplane.Dataplane
	app       *tview.Application
	statsView *tview.TextView
	sigView   *tview.TextView
	latView   *tview.TextView
	helpView  *tview.TextView
	startTime time.Time
	stopChan  chan struct{}
	stopOnce  sync.Once // Prevent double-close panic
}

// New creates a new TUI application.
func New(dp *dataplane.Dataplane) *App {
	return &App{
		dp:        dp,
		app:       tview.NewApplication(),
		statsView: nil,
		sigView:   nil,
		latView:   nil,
		helpView:  nil,
		startTime: time.Now(),
		stopChan:  make(chan struct{}),
		stopOnce:  sync.Once{},
	}
}

// Run starts the TUI.
func (a *App) Run() error {
	// Create main stats panel.
	a.statsView = tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignLeft)
	a.statsView.SetBorder(true).SetTitle(" Statistics ")

	// Create signature breakdown panel.
	a.sigView = tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignLeft)
	a.sigView.SetBorder(true).SetTitle(" Signatures ")

	// Create latency panel.
	a.latView = tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignLeft)
	a.latView.SetBorder(true).SetTitle(" Latency ")

	// Create help panel.
	a.helpView = tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignCenter).
		SetText("[yellow]q[white] quit  [yellow]r[white] reset  [yellow]p[white] pause")
	a.helpView.SetBorder(false)

	// Create header with MSN branding.
	header := tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignCenter).
		SetText(fmt.Sprintf("[#2d7a3e]MSN Reflector[white] | [yellow]Mustard Seed Networks[white] | Interface: [cyan]%s[white] | Status: [#2d7a3e]● RUNNING",
			a.dp.Interface()))

	// Layout.
	statsRow := tview.NewFlex().
		AddItem(a.statsView, 0, statsFlexWeight, false).
		AddItem(a.sigView, 0, 1, false).
		AddItem(a.latView, 0, 1, false)

	mainFlex := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(header, 1, 0, false).
		AddItem(statsRow, 0, 1, false).
		AddItem(a.helpView, 1, 0, false)

	// Key bindings.
	a.app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Rune() {
		case 'q', 'Q':
			a.Stop()
			return nil
		case 'r', 'R':
			// Reset stats - would need to add this to dataplane.
			return nil
		case 'p', 'P':
			// Pause - would toggle stats updates.
			return nil
		}
		return event
	})

	// Start stats update goroutine.
	go a.updateLoop()

	// Run the app.
	err := a.app.SetRoot(mainFlex, true).EnableMouse(false).Run()
	if err != nil {
		return fmt.Errorf("TUI app run failed: %w", err)
	}
	return nil
}

// Stop signals the TUI to exit.
func (a *App) Stop() {
	a.stopOnce.Do(func() {
		close(a.stopChan)
		a.app.Stop()
	})
}

// updateLoop periodically refreshes the display.
func (a *App) updateLoop() {
	ticker := time.NewTicker(tickerIntervalMs * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-a.stopChan:
			return
		case <-ticker.C:
			a.updateStats()
		}
	}
}

// updateStats refreshes all stat panels.
func (a *App) updateStats() {
	stats := a.dp.GetStats()
	elapsed := time.Since(a.startTime).Seconds()

	// Calculate rates.
	pps := float64(0)
	mbps := float64(0)
	if elapsed > 0 {
		pps = float64(stats.PacketsReflected) / elapsed
		mbps = float64(stats.BytesReflected) * bitsPerByte / (elapsed * megabitsPerSecDenom)
	}

	// Main stats with MSN branding colors.
	statsText := fmt.Sprintf(
		"[#d4a017]RX Packets:[white]  %s\n"+
			"[#d4a017]TX Packets:[white]  %s\n"+
			"[#d4a017]RX Bytes:[white]    %s\n"+
			"[#d4a017]TX Bytes:[white]    %s\n"+
			"\n"+
			"[#2d7a3e]Rate:[white]        %.0f pps\n"+
			"[#2d7a3e]Throughput:[white]  %.2f Mbps\n"+
			"\n"+
			"[cyan]Uptime:[white]      %s",
		formatNumber(stats.PacketsReceived),
		formatNumber(stats.PacketsReflected),
		formatBytes(stats.BytesReceived),
		formatBytes(stats.BytesReflected),
		pps, mbps,
		formatDuration(time.Since(a.startTime)),
	)

	// Signature breakdown - ITO and Custom.
	sigText := fmt.Sprintf(
		"[cyan]ITO Signatures:[white]\n"+
			"  PROBEOT:  %s\n"+
			"  DATA:OT:  %s\n"+
			"  LATENCY:  %s\n"+
			"\n"+
			"[#d4a017]Custom Signatures:[white]\n"+
			"  RFC2544:  %s\n"+
			"  Y.1564:   %s\n"+
			"  MSN:      %s",
		formatNumber(stats.SigProbeOT),
		formatNumber(stats.SigDataOT),
		formatNumber(stats.SigLatency),
		formatNumber(stats.SigRFC2544),
		formatNumber(stats.SigY1564),
		formatNumber(stats.SigMSN),
	)

	// Latency stats.
	latText := ""
	if stats.LatencyCount > 0 {
		latText = fmt.Sprintf(
			"[#2d7a3e]Min:[white]   %.2f µs\n"+
				"[#2d7a3e]Avg:[white]   %.2f µs\n"+
				"[#2d7a3e]Max:[white]   %.2f µs\n"+
				"[#2d7a3e]Count:[white] %s",
			stats.LatencyMin,
			stats.LatencyAvg,
			stats.LatencyMax,
			formatNumber(stats.LatencyCount),
		)
	} else {
		latText = "[gray]No latency data\n(use --latency)"
	}

	// Update views on main thread.
	a.app.QueueUpdateDraw(func() {
		a.statsView.SetText(statsText)
		a.sigView.SetText(sigText)
		a.latView.SetText(latText)
	})
}

// Helper functions for formatting.

func formatNumber(n uint64) string {
	if n >= billion {
		return fmt.Sprintf("%.2fB", float64(n)/billion)
	}
	if n >= million {
		return fmt.Sprintf("%.2fM", float64(n)/million)
	}
	if n >= thousand {
		return fmt.Sprintf("%.2fK", float64(n)/thousand)
	}
	return strconv.FormatUint(n, 10)
}

func formatBytes(n uint64) string {
	if n >= terabyte {
		return fmt.Sprintf("%.2f TB", float64(n)/terabyte)
	}
	if n >= gigabyte {
		return fmt.Sprintf("%.2f GB", float64(n)/gigabyte)
	}
	if n >= megabyte {
		return fmt.Sprintf("%.2f MB", float64(n)/megabyte)
	}
	if n >= kilobyte {
		return fmt.Sprintf("%.2f KB", float64(n)/kilobyte)
	}
	return strconv.FormatUint(n, 10) + " B"
}

func formatDuration(d time.Duration) string {
	hours := int(d.Hours())
	minutes := int(d.Minutes()) % secondsPerMinute
	seconds := int(d.Seconds()) % secondsPerMinute

	if hours > 0 {
		return fmt.Sprintf("%dh %dm %ds", hours, minutes, seconds)
	}
	if minutes > 0 {
		return fmt.Sprintf("%dm %ds", minutes, seconds)
	}
	return strconv.Itoa(seconds) + "s"
}
