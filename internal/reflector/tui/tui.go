// SPDX-License-Identifier: BUSL-1.1

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

// FilterAll is the canonical "no filter" value for the reflector filter-active
// dropdown and the predefined "all" profile.
const FilterAll = "all"

// Constants for TUI configuration and formatting.
const (
	statsFlexWeight      = 2 // Weight for stats panel in flex layout.
	tickerIntervalMs     = 500
	bitsPerByte          = 8.0
	megabitsPerSecDenom  = 1000000.0
	billion              = 1000000000
	million              = 1000000
	thousand             = 1000
	terabyte             = 1099511627776
	gigabyte             = 1073741824
	megabyte             = 1048576
	kilobyte             = 1024
	secondsPerMinute     = 60
	profileListHeight    = 12 // Height for profile selector list.
	profileSelectorWidth = 50 // Width for profile selector modal.
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

// GetPredefinedProfiles returns the built-in filter profiles.
func GetPredefinedProfiles() []FilterProfile {
	return []FilterProfile{
		{Name: FilterAll, Description: "All signatures (no filter)", ITO: true, RFC2544: true, Y1564: true, MSN: true},
		{Name: "ito", Description: "ITO signatures only", ITO: true, RFC2544: false, Y1564: false, MSN: false},
		{Name: "rfc2544", Description: "RFC 2544 signatures only", ITO: false, RFC2544: true, Y1564: false, MSN: false},
		{Name: "y1564", Description: "Y.1564 signatures only", ITO: false, RFC2544: false, Y1564: true, MSN: false},
		{Name: "msn", Description: "MSN custom signatures only", ITO: false, RFC2544: false, Y1564: false, MSN: true},
		{Name: "standards", Description: "RFC 2544 + Y.1564", ITO: false, RFC2544: true, Y1564: true, MSN: false},
	}
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
		filterActive:   FilterAll,
		currentProfile: GetPredefinedProfiles()[0], // Default to "all".
		showExtHelp:    false,
	}
}

// NewWithFilter creates a new TUI application with a specific filter profile.
func NewWithFilter(dp *dataplane.Dataplane, filterProfile string) *App {
	a := New(dp)
	a.filterActive = filterProfile
	// Find and set the matching profile.
	for _, p := range GetPredefinedProfiles() {
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

// KeyAction represents the action to take for a key press.
type KeyAction int

// Key action constants.
const (
	KeyActionNone         KeyAction = iota // No action / unhandled key
	KeyActionQuit                          // Quit the application
	KeyActionReset                         // Reset statistics
	KeyActionTogglePause                   // Toggle pause state
	KeyActionShowProfiles                  // Show profile selector
	KeyActionToggleHelp                    // Toggle extended help
	KeyActionSetProfile1                   // Set profile 1
	KeyActionSetProfile2                   // Set profile 2
	KeyActionSetProfile3                   // Set profile 3
	KeyActionSetProfile4                   // Set profile 4
	KeyActionSetProfile5                   // Set profile 5
	KeyActionSetProfile6                   // Set profile 6
)

// ParseKeyAction determines what action to take for a given key rune.
// This is a pure function that can be tested independently.
func ParseKeyAction(r rune) KeyAction {
	switch r {
	case 'q', 'Q':
		return KeyActionQuit
	case 'r', 'R':
		return KeyActionReset
	case 'p', 'P':
		return KeyActionTogglePause
	case 'f', 'F':
		return KeyActionShowProfiles
	case 'h', 'H', '?':
		return KeyActionToggleHelp
	case '1':
		return KeyActionSetProfile1
	case '2':
		return KeyActionSetProfile2
	case '3':
		return KeyActionSetProfile3
	case '4':
		return KeyActionSetProfile4
	case '5':
		return KeyActionSetProfile5
	case '6':
		return KeyActionSetProfile6
	default:
		return KeyActionNone
	}
}

// handleKeyEvent handles keyboard input.
func (a *App) handleKeyEvent(event *tcell.EventKey) *tcell.EventKey {
	action := ParseKeyAction(event.Rune())

	switch action {
	case KeyActionQuit:
		a.Stop()
		return nil
	case KeyActionReset:
		a.resetStats()
		return nil
	case KeyActionTogglePause:
		a.togglePause()
		return nil
	case KeyActionShowProfiles:
		a.showProfileSelector()
		return nil
	case KeyActionToggleHelp:
		a.toggleExtendedHelp()
		return nil
	case KeyActionSetProfile1, KeyActionSetProfile2, KeyActionSetProfile3,
		KeyActionSetProfile4, KeyActionSetProfile5, KeyActionSetProfile6:
		profiles := GetPredefinedProfiles()
		idx := int(action - KeyActionSetProfile1)
		if idx < len(profiles) {
			a.setProfile(profiles[idx])
		}
		return nil
	case KeyActionNone:
		return event
	}

	return event
}

// createProfileSelector creates the filter profile selection modal.
func (a *App) createProfileSelector() {
	list := tview.NewList().
		ShowSecondaryText(true)

	for i, p := range GetPredefinedProfiles() {
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
			AddItem(list, profileListHeight, 0, true).
			AddItem(nil, 0, 1, false), profileSelectorWidth, 0, true).
		AddItem(nil, 0, 1, false)

	a.pages.AddPage("profiles", modal, true, false)
}

// showProfileSelector shows the profile selection modal.
func (a *App) showProfileSelector() {
	a.pages.SwitchToPage("profiles")
}

// setProfileState updates the filter profile state without UI updates.
// This is the testable core logic of setProfile.
func (a *App) setProfileState(p FilterProfile) {
	a.filterActive = p.Name
	a.currentProfile = p
}

// setProfile sets the active filter profile.
func (a *App) setProfile(p FilterProfile) {
	a.setProfileState(p)
	a.app.QueueUpdateDraw(func() {
		a.updateHeaderStatus()
	})
}

// toggleExtendedHelpState toggles the extended help state without UI updates.
// This is the testable core logic of toggleExtendedHelp.
func (a *App) toggleExtendedHelpState() {
	a.showExtHelp = !a.showExtHelp
}

// toggleExtendedHelp toggles extended help display.
func (a *App) toggleExtendedHelp() {
	a.toggleExtendedHelpState()
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

// togglePauseState toggles the paused state without UI updates.
// This is the testable core logic of togglePause.
func (a *App) togglePauseState() {
	a.pauseMu.Lock()
	a.paused = !a.paused
	a.pauseMu.Unlock()
}

// togglePause toggles the paused state.
func (a *App) togglePause() {
	a.togglePauseState()

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

// GenerateHeaderText generates the header text for display.
// This is a pure function that can be tested independently.
func GenerateHeaderText(interfaceName string, filterActive string, paused bool) string {
	status := "[#2d7a3e]● RUNNING"
	if paused {
		status = "[yellow]● PAUSED"
	}

	filterText := ""
	if filterActive != FilterAll && filterActive != "" {
		filterText = fmt.Sprintf(" | Filter: [cyan]%s[white]", filterActive)
	}

	return fmt.Sprintf(
		"[#2d7a3e]MSN Reflector[white] | [yellow]Mustard Seed Networks[white] | "+
			"Interface: [cyan]%s[white]%s | Status: %s",
		interfaceName,
		filterText,
		status,
	)
}

// updateHeaderStatus updates the header with current status.
func (a *App) updateHeaderStatus() {
	interfaceName := ""
	if a.dp != nil {
		interfaceName = a.dp.Interface()
	}
	a.headerView.SetText(GenerateHeaderText(interfaceName, a.filterActive, a.isPaused()))
}

// GenerateHelpText generates the help bar text for display.
// This is a pure function that can be tested independently.
func GenerateHelpText(paused bool, extendedHelp bool) string {
	pauseAction := "pause"
	if paused {
		pauseAction = "resume"
	}

	if extendedHelp {
		// Extended help with all keyboard shortcuts.
		return fmt.Sprintf(
			"[yellow]q[white] quit | [yellow]r[white] reset | [yellow]p[white] %s | "+
				"[yellow]f[white] filter | [yellow]1-6[white] quick filter | "+
				"[yellow]h/?[white] toggle help",
			pauseAction,
		)
	}
	// Compact help.
	return fmt.Sprintf(
		"[yellow]q[white] quit  [yellow]r[white] reset  [yellow]p[white] %s  "+
			"[yellow]f[white] filter  [yellow]?[white] help",
		pauseAction,
	)
}

// updateHelpText updates the help bar based on current state.
func (a *App) updateHelpText() {
	a.helpView.SetText(GenerateHelpText(a.isPaused(), a.showExtHelp))
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

// StatsInput holds all the data needed to generate stats text.
// This allows testing the stats text generation without dataplane dependency.
type StatsInput struct {
	PacketsReceived  uint64
	PacketsReflected uint64
	BytesReceived    uint64
	BytesReflected   uint64
	SigProbeOT       uint64
	SigDataOT        uint64
	SigLatency       uint64
	SigRFC2544       uint64
	SigY1564         uint64
	SigMSN           uint64
	LatencyMin       float64
	LatencyAvg       float64
	LatencyMax       float64
	LatencyCount     uint64
	Elapsed          float64
	Uptime           time.Duration
}

// GenerateStatsText generates the main statistics text panel.
func GenerateStatsText(s StatsInput) string {
	// Calculate rates.
	pps := float64(0)
	mbps := float64(0)
	if s.Elapsed > 0 {
		pps = float64(s.PacketsReflected) / s.Elapsed
		mbps = float64(s.BytesReflected) * bitsPerByte / (s.Elapsed * megabitsPerSecDenom)
	}

	return fmt.Sprintf(
		"[#d4a017]RX Packets:[white]  %s\n"+
			"[#d4a017]TX Packets:[white]  %s\n"+
			"[#d4a017]RX Bytes:[white]    %s\n"+
			"[#d4a017]TX Bytes:[white]    %s\n"+
			"\n"+
			"[#2d7a3e]Rate:[white]        %.0f pps\n"+
			"[#2d7a3e]Throughput:[white]  %.2f Mbps\n"+
			"\n"+
			"[cyan]Uptime:[white]      %s",
		FormatNumber(s.PacketsReceived),
		FormatNumber(s.PacketsReflected),
		FormatBytes(s.BytesReceived),
		FormatBytes(s.BytesReflected),
		pps, mbps,
		FormatDuration(s.Uptime),
	)
}

// GenerateSignatureText generates the signature breakdown text panel.
func GenerateSignatureText(s StatsInput) string {
	return fmt.Sprintf(
		"[cyan]ITO Signatures:[white]\n"+
			"  PROBEOT:  %s\n"+
			"  DATA:OT:  %s\n"+
			"  LATENCY:  %s\n"+
			"\n"+
			"[#d4a017]Custom Signatures:[white]\n"+
			"  RFC2544:  %s\n"+
			"  Y.1564:   %s\n"+
			"  MSN:      %s",
		FormatNumber(s.SigProbeOT),
		FormatNumber(s.SigDataOT),
		FormatNumber(s.SigLatency),
		FormatNumber(s.SigRFC2544),
		FormatNumber(s.SigY1564),
		FormatNumber(s.SigMSN),
	)
}

// GenerateLatencyText generates the latency statistics text panel.
func GenerateLatencyText(s StatsInput) string {
	if s.LatencyCount > 0 {
		return fmt.Sprintf(
			"[#2d7a3e]Min:[white]   %.2f µs\n"+
				"[#2d7a3e]Avg:[white]   %.2f µs\n"+
				"[#2d7a3e]Max:[white]   %.2f µs\n"+
				"[#2d7a3e]Count:[white] %s",
			s.LatencyMin,
			s.LatencyAvg,
			s.LatencyMax,
			FormatNumber(s.LatencyCount),
		)
	}
	return "[gray]No latency data\n(use --latency)"
}

// updateStats refreshes all stat panels.
func (a *App) updateStats() {
	stats := a.dp.GetStats()
	elapsed := time.Since(a.startTime).Seconds()

	input := StatsInput{
		PacketsReceived:  stats.PacketsReceived,
		PacketsReflected: stats.PacketsReflected,
		BytesReceived:    stats.BytesReceived,
		BytesReflected:   stats.BytesReflected,
		SigProbeOT:       stats.SigProbeOT,
		SigDataOT:        stats.SigDataOT,
		SigLatency:       stats.SigLatency,
		SigRFC2544:       stats.SigRFC2544,
		SigY1564:         stats.SigY1564,
		SigMSN:           stats.SigMSN,
		LatencyMin:       stats.LatencyMin,
		LatencyAvg:       stats.LatencyAvg,
		LatencyMax:       stats.LatencyMax,
		LatencyCount:     stats.LatencyCount,
		Elapsed:          elapsed,
		Uptime:           time.Since(a.startTime),
	}

	statsText := GenerateStatsText(input)
	sigText := GenerateSignatureText(input)
	latText := GenerateLatencyText(input)

	// Update views on main thread.
	a.app.QueueUpdateDraw(func() {
		a.statsView.SetText(statsText)
		a.sigView.SetText(sigText)
		a.latView.SetText(latText)
	})
}

// FormatNumber formats large numbers with K/M/B suffixes for readability.
// Used for packet counts and similar metrics.
func FormatNumber(n uint64) string {
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

// FormatBytes formats byte counts with KB/MB/GB/TB suffixes for readability.
func FormatBytes(n uint64) string {
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

// FormatDuration formats a [time.Duration] as a human-readable string (e.g., "1h 5m 30s").
func FormatDuration(d time.Duration) string {
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
