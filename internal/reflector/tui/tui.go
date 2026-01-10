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

// FilterProfile defines a signature filter configuration.
type FilterProfile struct {
	Name        string // Profile name.
	Description string // Profile description.
	ITO         bool   // Enable ITO signatures (PROBEOT, DATA:OT, LATENCY).
	RFC2544     bool   // Enable RFC 2544 signatures.
	Y1564       bool   // Enable Y.1564 signatures.
	MSN         bool   // Enable MSN custom signatures.
}

// PredefinedProfiles are the built-in filter profiles.
var PredefinedProfiles = []FilterProfile{
	{Name: "all", Description: "All signatures (no filter)", ITO: true, RFC2544: true, Y1564: true, MSN: true},
	{Name: "ito", Description: "ITO signatures only", ITO: true, RFC2544: false, Y1564: false, MSN: false},
	{Name: "rfc2544", Description: "RFC 2544 signatures only", ITO: false, RFC2544: true, Y1564: false, MSN: false},
	{Name: "y1564", Description: "Y.1564 signatures only", ITO: false, RFC2544: false, Y1564: true, MSN: false},
	{Name: "msn", Description: "MSN custom signatures only", ITO: false, RFC2544: false, Y1564: false, MSN: true},
	{Name: "standards", Description: "RFC 2544 + Y.1564", ITO: false, RFC2544: true, Y1564: true, MSN: false},
}

// App holds the TUI application state.
type App struct {
	dp             *dataplane.Dataplane
	app            *tview.Application
	pages          *tview.Pages
	statsView      *tview.TextView
	sigView        *tview.TextView
	latView        *tview.TextView
	helpView       *tview.TextView
	headerView     *tview.TextView
	startTime      time.Time
	stopChan       chan struct{}
	stopOnce       sync.Once // Prevent double-close panic
	paused         bool
	pauseMu        sync.Mutex
	filterActive   string        // Current filter profile name.
	currentProfile FilterProfile // Current filter profile settings.
	showExtHelp    bool          // Show extended help.
}

// New creates a new TUI application.
func New(dp *dataplane.Dataplane) *App {
	return &App{
		dp:             dp,
		app:            tview.NewApplication(),
		pages:          tview.NewPages(),
		statsView:      nil,
		sigView:        nil,
		latView:        nil,
		helpView:       nil,
		headerView:     nil,
		startTime:      time.Now(),
		stopChan:       make(chan struct{}),
		stopOnce:       sync.Once{},
		paused:         false,
		pauseMu:        sync.Mutex{},
		filterActive:   "all",
		currentProfile: PredefinedProfiles[0], // Default to "all".
		showExtHelp:    false,
	}
}

// NewWithFilter creates a new TUI application with a specific filter profile.
func NewWithFilter(dp *dataplane.Dataplane, filterProfile string) *App {
	a := New(dp)
	a.filterActive = filterProfile
	// Find and set the matching profile.
	for _, p := range PredefinedProfiles {
		if p.Name == filterProfile {
			a.currentProfile = p
			break
		}
	}
	return a
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
		SetTextAlign(tview.AlignCenter)
	a.updateHelpText()
	a.helpView.SetBorder(false)

	// Create header with MSN branding.
	a.headerView = tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignCenter)
	a.updateHeaderStatus()

	// Layout.
	statsRow := tview.NewFlex().
		AddItem(a.statsView, 0, statsFlexWeight, false).
		AddItem(a.sigView, 0, 1, false).
		AddItem(a.latView, 0, 1, false)

	mainFlex := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(a.headerView, 1, 0, false).
		AddItem(statsRow, 0, 1, false).
		AddItem(a.helpView, 1, 0, false)

	// Add main page.
	a.pages.AddPage("main", mainFlex, true, true)

	// Create profile selector page.
	a.createProfileSelector()

	// Key bindings.
	a.app.SetInputCapture(a.handleKeyEvent)

	// Start stats update goroutine.
	go a.updateLoop()

	// Run the app.
	err := a.app.SetRoot(a.pages, true).EnableMouse(false).Run()
	if err != nil {
		return fmt.Errorf("TUI app run failed: %w", err)
	}
	return nil
}

// handleKeyEvent handles keyboard input.
func (a *App) handleKeyEvent(event *tcell.EventKey) *tcell.EventKey {
	switch event.Rune() {
	case 'q', 'Q':
		a.Stop()
		return nil
	case 'r', 'R':
		a.resetStats()
		return nil
	case 'p', 'P':
		a.togglePause()
		return nil
	case 'f', 'F':
		a.showProfileSelector()
		return nil
	case 'h', 'H':
		a.toggleExtendedHelp()
		return nil
	case '?':
		a.toggleExtendedHelp()
		return nil
	}

	// Number keys 1-6 for quick profile selection.
	if event.Rune() >= '1' && event.Rune() <= '6' {
		idx := int(event.Rune() - '1')
		if idx < len(PredefinedProfiles) {
			a.setProfile(PredefinedProfiles[idx])
		}
		return nil
	}

	return event
}

// createProfileSelector creates the filter profile selection modal.
func (a *App) createProfileSelector() {
	list := tview.NewList().
		ShowSecondaryText(true)

	for i, p := range PredefinedProfiles {
		shortcut := rune('1' + i)
		profile := p // Capture for closure.
		list.AddItem(p.Name, p.Description, shortcut, func() {
			a.setProfile(profile)
			a.pages.SwitchToPage("main")
		})
	}

	list.SetBorder(true).SetTitle(" Select Filter Profile ")

	// Center the list in a modal-like frame.
	modal := tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(nil, 0, 1, false).
			AddItem(list, 12, 0, true).
			AddItem(nil, 0, 1, false), 50, 0, true).
		AddItem(nil, 0, 1, false)

	a.pages.AddPage("profiles", modal, true, false)
}

// showProfileSelector shows the profile selection modal.
func (a *App) showProfileSelector() {
	a.pages.SwitchToPage("profiles")
}

// setProfile sets the active filter profile.
func (a *App) setProfile(p FilterProfile) {
	a.filterActive = p.Name
	a.currentProfile = p
	a.app.QueueUpdateDraw(func() {
		a.updateHeaderStatus()
	})
}

// toggleExtendedHelp toggles extended help display.
func (a *App) toggleExtendedHelp() {
	a.showExtHelp = !a.showExtHelp
	a.app.QueueUpdateDraw(func() {
		a.updateHelpText()
	})
}

// Stop signals the TUI to exit.
func (a *App) Stop() {
	a.stopOnce.Do(func() {
		close(a.stopChan)
		a.app.Stop()
	})
}

// togglePause toggles the paused state.
func (a *App) togglePause() {
	a.pauseMu.Lock()
	a.paused = !a.paused
	a.pauseMu.Unlock()

	a.app.QueueUpdateDraw(func() {
		a.updateHeaderStatus()
		a.updateHelpText()
	})
}

// isPaused returns the current paused state.
func (a *App) isPaused() bool {
	a.pauseMu.Lock()
	defer a.pauseMu.Unlock()
	return a.paused
}

// resetStats resets the dataplane statistics and TUI timer.
func (a *App) resetStats() {
	a.dp.ResetStats()
	a.startTime = time.Now()

	// Force an immediate update to show zeroed stats.
	a.updateStats()
}

// updateHeaderStatus updates the header with current status.
func (a *App) updateHeaderStatus() {
	status := "[#2d7a3e]● RUNNING"
	if a.isPaused() {
		status = "[yellow]● PAUSED"
	}

	filterText := ""
	if a.filterActive != "all" && a.filterActive != "" {
		filterText = fmt.Sprintf(" | Filter: [cyan]%s[white]", a.filterActive)
	}

	a.headerView.SetText(fmt.Sprintf(
		"[#2d7a3e]MSN Reflector[white] | [yellow]Mustard Seed Networks[white] | "+
			"Interface: [cyan]%s[white]%s | Status: %s",
		a.dp.Interface(),
		filterText,
		status,
	))
}

// updateHelpText updates the help bar based on current state.
func (a *App) updateHelpText() {
	pauseAction := "pause"
	if a.isPaused() {
		pauseAction = "resume"
	}

	if a.showExtHelp {
		// Extended help with all keyboard shortcuts.
		a.helpView.SetText(fmt.Sprintf(
			"[yellow]q[white] quit | [yellow]r[white] reset | [yellow]p[white] %s | "+
				"[yellow]f[white] filter | [yellow]1-6[white] quick filter | "+
				"[yellow]h/?[white] toggle help",
			pauseAction,
		))
	} else {
		// Compact help.
		a.helpView.SetText(fmt.Sprintf(
			"[yellow]q[white] quit  [yellow]r[white] reset  [yellow]p[white] %s  "+
				"[yellow]f[white] filter  [yellow]?[white] help",
			pauseAction,
		))
	}
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
			if !a.isPaused() {
				a.updateStats()
			}
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
