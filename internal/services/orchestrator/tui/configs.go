// SPDX-License-Identifier: BUSL-1.1

package tui

import (
	"fmt"
	"math"
	"strconv"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// safeIntToUint32 safely converts int to uint32 with bounds checking.
// Returns 0 if the value is negative or exceeds uint32 max.
func safeIntToUint32(v int) uint32 {
	if v < 0 || v > math.MaxUint32 {
		return 0
	}
	return uint32(v)
}

// safeIntToUint8 safely converts int to uint8 with bounds checking.
// Returns 0 if the value is negative or exceeds uint8 max.
func safeIntToUint8(v int) uint8 {
	if v < 0 || v > math.MaxUint8 {
		return 0
	}
	return uint8(v)
}

// RFC2889Config holds RFC 2889 LAN Switch test configuration.
type RFC2889Config struct {
	FrameSize      uint32  // Frame size in bytes.
	Duration       uint32  // Test duration in seconds.
	Warmup         uint32  // Warmup duration in seconds.
	AddressCount   uint32  // Number of MAC addresses.
	AcceptableLoss float64 // Maximum acceptable loss %.
	PortCount      uint32  // Number of ports.
	Pattern        uint32  // Traffic pattern: 0=mesh, 1=pair, 2=broadcast.
}

// RFC6349Config holds RFC 6349 TCP throughput test configuration.
type RFC6349Config struct {
	TargetRateMbps  float64 // Target rate in Mbps.
	MinRTTMs        float64 // Minimum RTT in ms.
	MaxRTTMs        float64 // Maximum RTT in ms.
	RWNDSize        uint32  // Receive window size.
	Duration        uint32  // Test duration in seconds.
	ParallelStreams uint32  // Number of parallel streams.
	MSS             uint32  // Maximum segment size.
	Mode            uint32  // 0=bidirectional, 1=upstream, 2=downstream.
}

// Y1731Config holds Y.1731 OAM test configuration.
type Y1731Config struct {
	MEPID          uint32 // Maintenance End Point ID.
	MEGLevel       uint32 // MEG Level (0-7).
	MEGID          string // MEG ID string.
	CCMInterval    uint32 // CCM interval in ms.
	Priority       uint8  // 802.1p priority.
	Duration       uint32 // Test duration in seconds.
	IntervalMs     uint32 // Measurement interval in ms.
	Count          uint32 // Frames per interval.
	FrameSize      uint32 // Frame size in bytes.
	PriorityTagged bool   // Use VLAN tagging.
}

// TSNConfig holds TSN 802.1Qbv test configuration.
type TSNConfig struct {
	Duration          uint32 // Test duration in seconds.
	Warmup            uint32 // Warmup in seconds.
	FrameSize         uint32 // Frame size in bytes.
	MaxLatencyNs      uint32 // Max acceptable latency in ns.
	MaxJitterNs       uint32 // Max acceptable jitter in ns.
	RequirePTPSync    bool   // Require PTP sync.
	MaxSyncOffsetNs   uint32 // Max PTP offset in ns.
	PTPEnabled        bool   // Enable PTP timestamping.
	PreemptionEnabled bool   // Enable frame preemption.
	NumTrafficClasses uint32 // Number of traffic classes.
	BaseTimeNs        uint64 // Base time for gate schedule (ns since epoch).
	CycleTimeNs       uint32 // TAS cycle time in ns.
	TrafficClass      uint32 // Traffic class for test.
}

// TrafficGenConfig holds custom traffic generator configuration.
type TrafficGenConfig struct {
	FrameSize       uint32  // Frame size in bytes.
	RatePct         float64 // Rate as % of line rate.
	Duration        uint32  // Duration in seconds.
	Warmup          uint32  // Warmup in seconds.
	StreamID        uint32  // Stream identifier.
	BurstMode       bool    // Enable burst mode.
	BurstSize       uint32  // Frames per burst.
	InterBurstGapUs uint32  // Gap between bursts in µs.
	SrcMac          string  // Source MAC address (empty=auto).
	DstMac          string  // Destination MAC address (empty=broadcast).
	VlanID          uint16  // VLAN ID (0=untagged).
	VlanPriority    uint8   // VLAN priority.
}

// Default configurations.
const (
	// RFC 2889 defaults.
	defaultRFC2889FrameSize    = 64
	defaultRFC2889Duration     = 60
	defaultRFC2889Warmup       = 2
	defaultRFC2889AddressCount = 8192
	defaultRFC2889PortCount    = 2
	defaultRFC2889Pattern      = 0

	// RFC 6349 defaults.
	defaultRFC6349Rate     = 100.0
	defaultRFC6349MinRTT   = 1.0
	defaultRFC6349MaxRTT   = 100.0
	defaultRFC6349RWND     = 65535
	defaultRFC6349Duration = 30
	defaultRFC6349Streams  = 1
	defaultRFC6349MSS      = 1460
	defaultRFC6349Mode     = 0

	// Y.1731 defaults.
	defaultY1731MEPID       = 1
	defaultY1731MEGLevel    = 4
	defaultY1731CCMInterval = 1000
	defaultY1731Priority    = 6
	defaultY1731Duration    = 60
	defaultY1731IntervalMs  = 100
	defaultY1731Count       = 10
	defaultY1731FrameSize   = 64

	// TSN defaults.
	defaultTSNDuration      = 60
	defaultTSNWarmup        = 5
	defaultTSNFrameSize     = 64
	defaultTSNMaxLatencyNs  = 1000000
	defaultTSNMaxJitterNs   = 100000
	defaultTSNMaxSyncOffset = 1000
	defaultTSNNumClasses    = 8
	defaultTSNCycleTimeNs   = 1000000
	defaultTSNTrafficClass  = 7

	// TrafficGen defaults.
	defaultTGenFrameSize       = 64
	defaultTGenRatePct         = 100.0
	defaultTGenDuration        = 60
	defaultTGenWarmup          = 2
	defaultTGenStreamID        = 1
	defaultTGenBurstSize       = 100
	defaultTGenInterBurstGapUs = 1000

	// Form layout constants.
	configFormWidth = 15
	megIDExtraWidth = 5

	// Conversion constants.
	nsToMicroseconds = 1000

	// Dropdown default indices.
	defaultCCMIntervalIndex = 3 // Default 1s in CCM interval list.
	defaultCycleTimeIndex   = 3 // Default 1ms in cycle time list.
)

// DefaultRFC2889Config returns default RFC 2889 configuration.
func DefaultRFC2889Config() RFC2889Config {
	return RFC2889Config{
		FrameSize:      defaultRFC2889FrameSize,
		Duration:       defaultRFC2889Duration,
		Warmup:         defaultRFC2889Warmup,
		AddressCount:   defaultRFC2889AddressCount,
		AcceptableLoss: 0.0,
		PortCount:      defaultRFC2889PortCount,
		Pattern:        defaultRFC2889Pattern,
	}
}

// DefaultRFC6349Config returns default RFC 6349 configuration.
func DefaultRFC6349Config() RFC6349Config {
	return RFC6349Config{
		TargetRateMbps:  defaultRFC6349Rate,
		MinRTTMs:        defaultRFC6349MinRTT,
		MaxRTTMs:        defaultRFC6349MaxRTT,
		RWNDSize:        defaultRFC6349RWND,
		Duration:        defaultRFC6349Duration,
		ParallelStreams: defaultRFC6349Streams,
		MSS:             defaultRFC6349MSS,
		Mode:            defaultRFC6349Mode,
	}
}

// DefaultY1731Config returns default Y.1731 configuration.
func DefaultY1731Config() Y1731Config {
	return Y1731Config{
		MEPID:          defaultY1731MEPID,
		MEGLevel:       defaultY1731MEGLevel,
		MEGID:          "MSN-MEG-01",
		CCMInterval:    defaultY1731CCMInterval,
		Priority:       defaultY1731Priority,
		Duration:       defaultY1731Duration,
		IntervalMs:     defaultY1731IntervalMs,
		Count:          defaultY1731Count,
		FrameSize:      defaultY1731FrameSize,
		PriorityTagged: true,
	}
}

// DefaultTSNConfig returns default TSN configuration.
func DefaultTSNConfig() TSNConfig {
	return TSNConfig{
		Duration:          defaultTSNDuration,
		Warmup:            defaultTSNWarmup,
		FrameSize:         defaultTSNFrameSize,
		MaxLatencyNs:      defaultTSNMaxLatencyNs,
		MaxJitterNs:       defaultTSNMaxJitterNs,
		RequirePTPSync:    true,
		MaxSyncOffsetNs:   defaultTSNMaxSyncOffset,
		PTPEnabled:        true,
		PreemptionEnabled: false,
		NumTrafficClasses: defaultTSNNumClasses,
		BaseTimeNs:        0, // Use current time if 0.
		CycleTimeNs:       defaultTSNCycleTimeNs,
		TrafficClass:      defaultTSNTrafficClass,
	}
}

// DefaultTrafficGenConfig returns default traffic generator configuration.
func DefaultTrafficGenConfig() TrafficGenConfig {
	return TrafficGenConfig{
		FrameSize:       defaultTGenFrameSize,
		RatePct:         defaultTGenRatePct,
		Duration:        defaultTGenDuration,
		Warmup:          defaultTGenWarmup,
		StreamID:        defaultTGenStreamID,
		BurstMode:       false,
		BurstSize:       defaultTGenBurstSize,
		InterBurstGapUs: defaultTGenInterBurstGapUs,
		SrcMac:          "", // Empty = auto-generated.
		DstMac:          "", // Empty = broadcast.
		VlanID:          0,
		VlanPriority:    0,
	}
}

// RFC2889ConfigEditor provides a form for RFC 2889 configuration.
type RFC2889ConfigEditor struct {
	app      *App
	form     *tview.Form
	config   *RFC2889Config
	onSave   func(RFC2889Config)
	onCancel func()
}

// NewRFC2889ConfigEditor creates a new RFC 2889 configuration editor.
func (a *App) NewRFC2889ConfigEditor(
	config *RFC2889Config,
	onSave func(RFC2889Config),
	onCancel func(),
) *RFC2889ConfigEditor {
	if config == nil {
		def := DefaultRFC2889Config()
		config = &def
	}

	editor := &RFC2889ConfigEditor{
		app:      a,
		form:     tview.NewForm(),
		config:   config,
		onSave:   onSave,
		onCancel: onCancel,
	}

	editor.build()
	return editor
}

func (e *RFC2889ConfigEditor) build() {
	e.form.SetTitle(" RFC 2889 Configuration ").SetBorder(true)

	// Frame Size.
	frameSizes := []string{"64", "128", "256", "512", "1024", "1280", "1518"}
	fsIdx := 0
	currentFS := strconv.FormatUint(uint64(e.config.FrameSize), 10)
	for i, s := range frameSizes {
		if s == currentFS {
			fsIdx = i
			break
		}
	}
	e.form.AddDropDown("Frame Size (bytes)", frameSizes, fsIdx, func(option string, _ int) {
		v, _ := strconv.ParseUint(option, 10, 32)
		e.config.FrameSize = uint32(v)
	})

	// Duration.
	e.form.AddInputField("Duration (s)", strconv.FormatUint(uint64(e.config.Duration), 10),
		configFormWidth, nil, func(text string) {
			v, err := strconv.ParseUint(text, 10, 32)
			if err == nil {
				e.config.Duration = uint32(v)
			}
		})

	// Warmup.
	e.form.AddInputField("Warmup (s)", strconv.FormatUint(uint64(e.config.Warmup), 10),
		configFormWidth, nil, func(text string) {
			v, err := strconv.ParseUint(text, 10, 32)
			if err == nil {
				e.config.Warmup = uint32(v)
			}
		})

	// Address Count.
	e.form.AddInputField("Address Count", strconv.FormatUint(uint64(e.config.AddressCount), 10),
		configFormWidth, nil, func(text string) {
			v, err := strconv.ParseUint(text, 10, 32)
			if err == nil {
				e.config.AddressCount = uint32(v)
			}
		})

	// Port Count.
	e.form.AddInputField("Port Count", strconv.FormatUint(uint64(e.config.PortCount), 10),
		configFormWidth, nil, func(text string) {
			v, err := strconv.ParseUint(text, 10, 32)
			if err == nil {
				e.config.PortCount = uint32(v)
			}
		})

	// Pattern.
	patterns := []string{"Full Mesh", "Pair", "Broadcast"}
	patternIdx := 0
	if e.config.Pattern < safeIntToUint32(len(patterns)) {
		patternIdx = int(e.config.Pattern)
	}
	e.form.AddDropDown("Traffic Pattern", patterns, patternIdx, func(_ string, idx int) {
		if idx >= 0 && idx < len(patterns) {
			e.config.Pattern = safeIntToUint32(idx)
		}
	})

	// Acceptable Loss.
	e.form.AddInputField("Acceptable Loss (%)", fmt.Sprintf("%.4f", e.config.AcceptableLoss),
		configFormWidth, nil, func(text string) {
			v, err := strconv.ParseFloat(text, 64)
			if err == nil {
				e.config.AcceptableLoss = v
			}
		})

	// Help text.
	helpText := tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignCenter).
		SetText("[yellow]F5[white] Save | [yellow]Esc[white] Cancel")

	container := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(e.form, 0, 1, true).
		AddItem(helpText, 1, 0, false)

	e.app.pages.AddPage("rfc2889config", container, true, false)
}

// Show displays the RFC 2889 configuration editor.
func (e *RFC2889ConfigEditor) Show() {
	e.app.pages.SwitchToPage("rfc2889config")

	e.app.app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		//exhaustive:ignore
		switch event.Key() {
		case tcell.KeyF5:
			if e.onSave != nil {
				e.onSave(*e.config)
			}
			e.Hide()
			return nil
		case tcell.KeyEscape:
			if e.onCancel != nil {
				e.onCancel()
			}
			e.Hide()
			return nil
		default:
			return event
		}
	})
}

// Hide hides the editor and returns to main view.
func (e *RFC2889ConfigEditor) Hide() {
	e.app.pages.SwitchToPage("main")
	e.app.build()
}

// ShowRFC2889Config shows the RFC 2889 configuration editor.
func (a *App) ShowRFC2889Config(config *RFC2889Config, onSave func(RFC2889Config)) {
	editor := a.NewRFC2889ConfigEditor(config, onSave, nil)
	editor.Show()
}

// RFC6349ConfigEditor provides a form for RFC 6349 configuration.
type RFC6349ConfigEditor struct {
	app      *App
	form     *tview.Form
	config   *RFC6349Config
	onSave   func(RFC6349Config)
	onCancel func()
}

// NewRFC6349ConfigEditor creates a new RFC 6349 configuration editor.
func (a *App) NewRFC6349ConfigEditor(
	config *RFC6349Config,
	onSave func(RFC6349Config),
	onCancel func(),
) *RFC6349ConfigEditor {
	if config == nil {
		def := DefaultRFC6349Config()
		config = &def
	}

	editor := &RFC6349ConfigEditor{
		app:      a,
		form:     tview.NewForm(),
		config:   config,
		onSave:   onSave,
		onCancel: onCancel,
	}

	editor.build()
	return editor
}

func (e *RFC6349ConfigEditor) build() {
	e.form.SetTitle(" RFC 6349 TCP Configuration ").SetBorder(true)

	// Target Rate.
	e.form.AddInputField("Target Rate (Mbps)", fmt.Sprintf("%.2f", e.config.TargetRateMbps),
		configFormWidth, nil, func(text string) {
			v, err := strconv.ParseFloat(text, 64)
			if err == nil {
				e.config.TargetRateMbps = v
			}
		})

	// Min RTT.
	e.form.AddInputField("Min RTT (ms)", fmt.Sprintf("%.1f", e.config.MinRTTMs),
		configFormWidth, nil, func(text string) {
			v, err := strconv.ParseFloat(text, 64)
			if err == nil {
				e.config.MinRTTMs = v
			}
		})

	// Max RTT.
	e.form.AddInputField("Max RTT (ms)", fmt.Sprintf("%.1f", e.config.MaxRTTMs),
		configFormWidth, nil, func(text string) {
			v, err := strconv.ParseFloat(text, 64)
			if err == nil {
				e.config.MaxRTTMs = v
			}
		})

	// RWND Size.
	e.form.AddInputField("RWND Size (bytes)", strconv.FormatUint(uint64(e.config.RWNDSize), 10),
		configFormWidth, nil, func(text string) {
			v, err := strconv.ParseUint(text, 10, 32)
			if err == nil {
				e.config.RWNDSize = uint32(v)
			}
		})

	// Duration.
	e.form.AddInputField("Duration (s)", strconv.FormatUint(uint64(e.config.Duration), 10),
		configFormWidth, nil, func(text string) {
			v, err := strconv.ParseUint(text, 10, 32)
			if err == nil {
				e.config.Duration = uint32(v)
			}
		})

	// Parallel Streams.
	e.form.AddInputField("Parallel Streams", strconv.FormatUint(uint64(e.config.ParallelStreams), 10),
		configFormWidth, nil, func(text string) {
			v, err := strconv.ParseUint(text, 10, 32)
			if err == nil {
				e.config.ParallelStreams = uint32(v)
			}
		})

	// MSS.
	mssOptions := []string{"536", "1220", "1460", "8960"}
	mssIdx := 2 // Default to 1460.
	currentMSS := strconv.FormatUint(uint64(e.config.MSS), 10)
	for i, s := range mssOptions {
		if s == currentMSS {
			mssIdx = i
			break
		}
	}
	e.form.AddDropDown("MSS (bytes)", mssOptions, mssIdx, func(option string, _ int) {
		v, _ := strconv.ParseUint(option, 10, 32)
		e.config.MSS = uint32(v)
	})

	// Mode.
	modes := []string{"Bidirectional", "Upstream", "Downstream"}
	modeIdx := 0
	if e.config.Mode < safeIntToUint32(len(modes)) {
		modeIdx = int(e.config.Mode)
	}
	e.form.AddDropDown("Test Mode", modes, modeIdx, func(_ string, idx int) {
		if idx >= 0 && idx < len(modes) {
			e.config.Mode = safeIntToUint32(idx)
		}
	})

	helpText := tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignCenter).
		SetText("[yellow]F5[white] Save | [yellow]Esc[white] Cancel")

	container := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(e.form, 0, 1, true).
		AddItem(helpText, 1, 0, false)

	e.app.pages.AddPage("rfc6349config", container, true, false)
}

// Show displays the RFC 6349 configuration editor.
func (e *RFC6349ConfigEditor) Show() {
	e.app.pages.SwitchToPage("rfc6349config")

	e.app.app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		//exhaustive:ignore
		switch event.Key() {
		case tcell.KeyF5:
			if e.onSave != nil {
				e.onSave(*e.config)
			}
			e.Hide()
			return nil
		case tcell.KeyEscape:
			if e.onCancel != nil {
				e.onCancel()
			}
			e.Hide()
			return nil
		default:
			return event
		}
	})
}

// Hide hides the editor.
func (e *RFC6349ConfigEditor) Hide() {
	e.app.pages.SwitchToPage("main")
	e.app.build()
}

// ShowRFC6349Config shows the RFC 6349 configuration editor.
func (a *App) ShowRFC6349Config(config *RFC6349Config, onSave func(RFC6349Config)) {
	editor := a.NewRFC6349ConfigEditor(config, onSave, nil)
	editor.Show()
}

// Y1731ConfigEditor provides a form for Y.1731 OAM configuration.
type Y1731ConfigEditor struct {
	app      *App
	form     *tview.Form
	config   *Y1731Config
	onSave   func(Y1731Config)
	onCancel func()
}

// NewY1731ConfigEditor creates a new Y.1731 configuration editor.
func (a *App) NewY1731ConfigEditor(
	config *Y1731Config,
	onSave func(Y1731Config),
	onCancel func(),
) *Y1731ConfigEditor {
	if config == nil {
		def := DefaultY1731Config()
		config = &def
	}

	editor := &Y1731ConfigEditor{
		app:      a,
		form:     tview.NewForm(),
		config:   config,
		onSave:   onSave,
		onCancel: onCancel,
	}

	editor.build()
	return editor
}

func (e *Y1731ConfigEditor) build() {
	e.form.SetTitle(" Y.1731 OAM Configuration ").SetBorder(true)
	e.addMEPFields()
	e.addTimingFields()
	e.addFrameFields()
	e.buildY1731Container()
}

func (e *Y1731ConfigEditor) addMEPFields() {
	// MEP ID.
	e.form.AddInputField("MEP ID", strconv.FormatUint(uint64(e.config.MEPID), 10),
		configFormWidth, nil, func(text string) {
			v, err := strconv.ParseUint(text, 10, 32)
			if err == nil {
				e.config.MEPID = uint32(v)
			}
		})

	// MEG Level.
	levels := []string{"0", "1", "2", "3", "4", "5", "6", "7"}
	levelIdx := 0
	if e.config.MEGLevel < safeIntToUint32(len(levels)) {
		levelIdx = int(e.config.MEGLevel)
	}
	e.form.AddDropDown("MEG Level", levels, levelIdx, func(_ string, idx int) {
		if idx >= 0 && idx < len(levels) {
			e.config.MEGLevel = safeIntToUint32(idx)
		}
	})

	// MEG ID.
	e.form.AddInputField("MEG ID", e.config.MEGID, configFormWidth+megIDExtraWidth, nil, func(text string) {
		e.config.MEGID = text
	})

	// CCM Interval.
	ccmOptions := []string{"3.33ms", "10ms", "100ms", "1s", "10s", "1min", "10min"}
	ccmValues := []uint32{3, 10, 100, 1000, 10000, 60000, 600000}
	ccmIdx := findCCMIndex(e.config.CCMInterval, ccmValues)
	e.form.AddDropDown("CCM Interval", ccmOptions, ccmIdx, func(_ string, idx int) {
		if idx >= 0 && idx < len(ccmValues) {
			e.config.CCMInterval = ccmValues[idx]
		}
	})

	// Priority.
	priorities := []string{"0", "1", "2", "3", "4", "5", "6", "7"}
	priorityIdx := 0
	if int(e.config.Priority) < len(priorities) {
		priorityIdx = int(e.config.Priority)
	}
	e.form.AddDropDown("Priority", priorities, priorityIdx, func(_ string, idx int) {
		if idx >= 0 && idx < len(priorities) {
			e.config.Priority = safeIntToUint8(idx)
		}
	})
}

func findCCMIndex(interval uint32, values []uint32) int {
	for i, v := range values {
		if v == interval {
			return i
		}
	}
	return defaultCCMIntervalIndex
}

func (e *Y1731ConfigEditor) addTimingFields() {
	// Duration.
	e.form.AddInputField("Duration (s)", strconv.FormatUint(uint64(e.config.Duration), 10),
		configFormWidth, nil, func(text string) {
			v, err := strconv.ParseUint(text, 10, 32)
			if err == nil {
				e.config.Duration = uint32(v)
			}
		})

	// Interval Ms.
	e.form.AddInputField("Measurement Interval (ms)", strconv.FormatUint(uint64(e.config.IntervalMs), 10),
		configFormWidth, nil, func(text string) {
			v, err := strconv.ParseUint(text, 10, 32)
			if err == nil {
				e.config.IntervalMs = uint32(v)
			}
		})

	// Count.
	e.form.AddInputField("Frames per Interval", strconv.FormatUint(uint64(e.config.Count), 10),
		configFormWidth, nil, func(text string) {
			v, err := strconv.ParseUint(text, 10, 32)
			if err == nil {
				e.config.Count = uint32(v)
			}
		})
}

func (e *Y1731ConfigEditor) addFrameFields() {
	// Frame Size.
	frameSizes := []string{"64", "128", "256", "512", "1024", "1518"}
	fsIdx := findFrameSizeIndex(e.config.FrameSize, frameSizes)
	e.form.AddDropDown("Frame Size", frameSizes, fsIdx, func(option string, _ int) {
		v, _ := strconv.ParseUint(option, 10, 32)
		e.config.FrameSize = uint32(v)
	})

	// Priority Tagged.
	e.form.AddCheckbox("Priority Tagged (802.1Q)", e.config.PriorityTagged, func(checked bool) {
		e.config.PriorityTagged = checked
	})
}

func findFrameSizeIndex(frameSize uint32, sizes []string) int {
	currentFS := strconv.FormatUint(uint64(frameSize), 10)
	for i, s := range sizes {
		if s == currentFS {
			return i
		}
	}
	return 0
}

func (e *Y1731ConfigEditor) buildY1731Container() {
	helpText := tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignCenter).
		SetText("[yellow]F5[white] Save | [yellow]Esc[white] Cancel")

	container := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(e.form, 0, 1, true).
		AddItem(helpText, 1, 0, false)

	e.app.pages.AddPage("y1731config", container, true, false)
}

// Show displays the Y.1731 configuration editor.
func (e *Y1731ConfigEditor) Show() {
	e.app.pages.SwitchToPage("y1731config")

	e.app.app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		//exhaustive:ignore
		switch event.Key() {
		case tcell.KeyF5:
			if e.onSave != nil {
				e.onSave(*e.config)
			}
			e.Hide()
			return nil
		case tcell.KeyEscape:
			if e.onCancel != nil {
				e.onCancel()
			}
			e.Hide()
			return nil
		default:
			return event
		}
	})
}

// Hide hides the editor.
func (e *Y1731ConfigEditor) Hide() {
	e.app.pages.SwitchToPage("main")
	e.app.build()
}

// ShowY1731Config shows the Y.1731 configuration editor.
func (a *App) ShowY1731Config(config *Y1731Config, onSave func(Y1731Config)) {
	editor := a.NewY1731ConfigEditor(config, onSave, nil)
	editor.Show()
}

// TSNConfigEditor provides a form for TSN configuration.
type TSNConfigEditor struct {
	app      *App
	form     *tview.Form
	config   *TSNConfig
	onSave   func(TSNConfig)
	onCancel func()
}

// NewTSNConfigEditor creates a new TSN configuration editor.
func (a *App) NewTSNConfigEditor(
	config *TSNConfig,
	onSave func(TSNConfig),
	onCancel func(),
) *TSNConfigEditor {
	if config == nil {
		def := DefaultTSNConfig()
		config = &def
	}

	editor := &TSNConfigEditor{
		app:      a,
		form:     tview.NewForm(),
		config:   config,
		onSave:   onSave,
		onCancel: onCancel,
	}

	editor.build()
	return editor
}

func (e *TSNConfigEditor) build() {
	e.form.SetTitle(" TSN 802.1Qbv Configuration ").SetBorder(true)
	e.addTSNBasicFields()
	e.addTSNPTPFields()
	e.addTSNScheduleFields()
	e.buildTSNContainer()
}

func (e *TSNConfigEditor) addTSNBasicFields() {
	// Duration.
	e.form.AddInputField("Duration (s)", strconv.FormatUint(uint64(e.config.Duration), 10),
		configFormWidth, nil, func(text string) {
			v, err := strconv.ParseUint(text, 10, 32)
			if err == nil {
				e.config.Duration = uint32(v)
			}
		})

	// Warmup.
	e.form.AddInputField("Warmup (s)", strconv.FormatUint(uint64(e.config.Warmup), 10),
		configFormWidth, nil, func(text string) {
			v, err := strconv.ParseUint(text, 10, 32)
			if err == nil {
				e.config.Warmup = uint32(v)
			}
		})

	// Frame Size.
	frameSizes := []string{"64", "128", "256", "512", "1024", "1518"}
	fsIdx := findFrameSizeIndex(e.config.FrameSize, frameSizes)
	e.form.AddDropDown("Frame Size", frameSizes, fsIdx, func(option string, _ int) {
		v, _ := strconv.ParseUint(option, 10, 32)
		e.config.FrameSize = uint32(v)
	})

	// Max Latency (µs).
	e.form.AddInputField("Max Latency (µs)", strconv.FormatUint(uint64(e.config.MaxLatencyNs/nsToMicroseconds), 10),
		configFormWidth, nil, func(text string) {
			v, err := strconv.ParseUint(text, 10, 32)
			if err == nil {
				e.config.MaxLatencyNs = uint32(v) * nsToMicroseconds
			}
		})

	// Max Jitter (µs).
	e.form.AddInputField("Max Jitter (µs)", strconv.FormatUint(uint64(e.config.MaxJitterNs/nsToMicroseconds), 10),
		configFormWidth, nil, func(text string) {
			v, err := strconv.ParseUint(text, 10, 32)
			if err == nil {
				e.config.MaxJitterNs = uint32(v) * nsToMicroseconds
			}
		})
}

func (e *TSNConfigEditor) addTSNPTPFields() {
	// PTP Enabled.
	e.form.AddCheckbox("Enable PTP Timestamping", e.config.PTPEnabled, func(checked bool) {
		e.config.PTPEnabled = checked
	})

	// Require PTP Sync.
	e.form.AddCheckbox("Require PTP Sync", e.config.RequirePTPSync, func(checked bool) {
		e.config.RequirePTPSync = checked
	})

	// Max Sync Offset (ns).
	e.form.AddInputField("Max Sync Offset (ns)", strconv.FormatUint(uint64(e.config.MaxSyncOffsetNs), 10),
		configFormWidth, nil, func(text string) {
			v, err := strconv.ParseUint(text, 10, 32)
			if err == nil {
				e.config.MaxSyncOffsetNs = uint32(v)
			}
		})

	// Preemption Enabled.
	e.form.AddCheckbox("Enable Frame Preemption", e.config.PreemptionEnabled, func(checked bool) {
		e.config.PreemptionEnabled = checked
	})
}

func (e *TSNConfigEditor) addTSNScheduleFields() {
	// Cycle Time.
	cycleOptions := []string{"125µs", "250µs", "500µs", "1ms", "2ms", "4ms"}
	cycleValues := []uint32{125000, 250000, 500000, 1000000, 2000000, 4000000}
	cycleIdx := findCycleTimeIndex(e.config.CycleTimeNs, cycleValues)
	e.form.AddDropDown("Cycle Time", cycleOptions, cycleIdx, func(_ string, idx int) {
		if idx >= 0 && idx < len(cycleValues) {
			e.config.CycleTimeNs = cycleValues[idx]
		}
	})

	// Traffic Class.
	tcOptions := []string{"0", "1", "2", "3", "4", "5", "6", "7"}
	tcIdx := 0
	if e.config.TrafficClass < safeIntToUint32(len(tcOptions)) {
		tcIdx = int(e.config.TrafficClass)
	}
	e.form.AddDropDown("Traffic Class", tcOptions, tcIdx, func(_ string, idx int) {
		if idx >= 0 && idx < len(tcOptions) {
			e.config.TrafficClass = safeIntToUint32(idx)
		}
	})

	// Num Traffic Classes.
	e.form.AddInputField("Num Traffic Classes", strconv.FormatUint(uint64(e.config.NumTrafficClasses), 10),
		configFormWidth, nil, func(text string) {
			v, err := strconv.ParseUint(text, 10, 32)
			if err == nil && v >= 1 && v <= 8 {
				e.config.NumTrafficClasses = uint32(v)
			}
		})

	// Base Time (ns since epoch).
	e.form.AddInputField("Base Time (ns)", strconv.FormatUint(e.config.BaseTimeNs, 10),
		configFormWidth+megIDExtraWidth, nil, func(text string) {
			v, err := strconv.ParseUint(text, 10, 64)
			if err == nil {
				e.config.BaseTimeNs = v
			}
		})
}

func findCycleTimeIndex(cycleTime uint32, values []uint32) int {
	for i, v := range values {
		if v == cycleTime {
			return i
		}
	}
	return defaultCycleTimeIndex
}

func (e *TSNConfigEditor) buildTSNContainer() {
	helpText := tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignCenter).
		SetText("[yellow]F5[white] Save | [yellow]Esc[white] Cancel")

	container := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(e.form, 0, 1, true).
		AddItem(helpText, 1, 0, false)

	e.app.pages.AddPage("tsnconfig", container, true, false)
}

// Show displays the TSN configuration editor.
func (e *TSNConfigEditor) Show() {
	e.app.pages.SwitchToPage("tsnconfig")

	e.app.app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		//exhaustive:ignore
		switch event.Key() {
		case tcell.KeyF5:
			if e.onSave != nil {
				e.onSave(*e.config)
			}
			e.Hide()
			return nil
		case tcell.KeyEscape:
			if e.onCancel != nil {
				e.onCancel()
			}
			e.Hide()
			return nil
		default:
			return event
		}
	})
}

// Hide hides the editor.
func (e *TSNConfigEditor) Hide() {
	e.app.pages.SwitchToPage("main")
	e.app.build()
}

// ShowTSNConfig shows the TSN configuration editor.
func (a *App) ShowTSNConfig(config *TSNConfig, onSave func(TSNConfig)) {
	editor := a.NewTSNConfigEditor(config, onSave, nil)
	editor.Show()
}

// TrafficGenConfigEditor provides a form for traffic generator configuration.
type TrafficGenConfigEditor struct {
	app      *App
	form     *tview.Form
	config   *TrafficGenConfig
	onSave   func(TrafficGenConfig)
	onCancel func()
}

// NewTrafficGenConfigEditor creates a new traffic generator configuration editor.
func (a *App) NewTrafficGenConfigEditor(
	config *TrafficGenConfig,
	onSave func(TrafficGenConfig),
	onCancel func(),
) *TrafficGenConfigEditor {
	if config == nil {
		def := DefaultTrafficGenConfig()
		config = &def
	}

	editor := &TrafficGenConfigEditor{
		app:      a,
		form:     tview.NewForm(),
		config:   config,
		onSave:   onSave,
		onCancel: onCancel,
	}

	editor.build()
	return editor
}

func (e *TrafficGenConfigEditor) build() {
	e.form.SetTitle(" Traffic Generator Configuration ").SetBorder(true)
	e.addTrafficGenBasicFields()
	e.addTrafficGenBurstFields()
	e.addTrafficGenVLANFields()
	e.buildTrafficGenContainer()
}

func (e *TrafficGenConfigEditor) addTrafficGenBasicFields() {
	// Frame Size.
	frameSizes := []string{"64", "128", "256", "512", "1024", "1280", "1518", "9000"}
	fsIdx := findFrameSizeIndex(e.config.FrameSize, frameSizes)
	e.form.AddDropDown("Frame Size", frameSizes, fsIdx, func(option string, _ int) {
		v, _ := strconv.ParseUint(option, 10, 32)
		e.config.FrameSize = uint32(v)
	})

	// Rate %.
	e.form.AddInputField("Rate (% of line)", fmt.Sprintf("%.2f", e.config.RatePct),
		configFormWidth, nil, func(text string) {
			v, err := strconv.ParseFloat(text, 64)
			if err == nil {
				e.config.RatePct = v
			}
		})

	// Duration.
	e.form.AddInputField("Duration (s)", strconv.FormatUint(uint64(e.config.Duration), 10),
		configFormWidth, nil, func(text string) {
			v, err := strconv.ParseUint(text, 10, 32)
			if err == nil {
				e.config.Duration = uint32(v)
			}
		})

	// Warmup.
	e.form.AddInputField("Warmup (s)", strconv.FormatUint(uint64(e.config.Warmup), 10),
		configFormWidth, nil, func(text string) {
			v, err := strconv.ParseUint(text, 10, 32)
			if err == nil {
				e.config.Warmup = uint32(v)
			}
		})

	// Stream ID.
	e.form.AddInputField("Stream ID", strconv.FormatUint(uint64(e.config.StreamID), 10),
		configFormWidth, nil, func(text string) {
			v, err := strconv.ParseUint(text, 10, 32)
			if err == nil {
				e.config.StreamID = uint32(v)
			}
		})
}

func (e *TrafficGenConfigEditor) addTrafficGenBurstFields() {
	// Burst Mode.
	e.form.AddCheckbox("Burst Mode", e.config.BurstMode, func(checked bool) {
		e.config.BurstMode = checked
	})

	// Burst Size.
	e.form.AddInputField("Burst Size (frames)", strconv.FormatUint(uint64(e.config.BurstSize), 10),
		configFormWidth, nil, func(text string) {
			v, err := strconv.ParseUint(text, 10, 32)
			if err == nil {
				e.config.BurstSize = uint32(v)
			}
		})

	// Inter-Burst Gap.
	e.form.AddInputField("Inter-Burst Gap (µs)", strconv.FormatUint(uint64(e.config.InterBurstGapUs), 10),
		configFormWidth, nil, func(text string) {
			v, err := strconv.ParseUint(text, 10, 32)
			if err == nil {
				e.config.InterBurstGapUs = uint32(v)
			}
		})
}

func (e *TrafficGenConfigEditor) addTrafficGenVLANFields() {
	// VLAN ID.
	e.form.AddInputField("VLAN ID (0=untagged)", strconv.FormatUint(uint64(e.config.VlanID), 10),
		configFormWidth, nil, func(text string) {
			v, err := strconv.ParseUint(text, 10, 16)
			if err == nil {
				e.config.VlanID = uint16(v)
			}
		})

	// VLAN Priority.
	vlanPriorities := []string{"0", "1", "2", "3", "4", "5", "6", "7"}
	vlanPriorityIdx := 0
	if int(e.config.VlanPriority) < len(vlanPriorities) {
		vlanPriorityIdx = int(e.config.VlanPriority)
	}
	e.form.AddDropDown("VLAN Priority", vlanPriorities, vlanPriorityIdx, func(_ string, idx int) {
		if idx >= 0 && idx < len(vlanPriorities) {
			e.config.VlanPriority = safeIntToUint8(idx)
		}
	})

	// Source MAC.
	e.form.AddInputField("Source MAC (empty=auto)", e.config.SrcMac,
		configFormWidth+megIDExtraWidth, nil, func(text string) {
			e.config.SrcMac = text
		})

	// Destination MAC.
	e.form.AddInputField("Dest MAC (empty=bcast)", e.config.DstMac,
		configFormWidth+megIDExtraWidth, nil, func(text string) {
			e.config.DstMac = text
		})
}

func (e *TrafficGenConfigEditor) buildTrafficGenContainer() {
	helpText := tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignCenter).
		SetText("[yellow]F5[white] Save | [yellow]Esc[white] Cancel")

	container := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(e.form, 0, 1, true).
		AddItem(helpText, 1, 0, false)

	e.app.pages.AddPage("trafficgenconfig", container, true, false)
}

// Show displays the traffic generator configuration editor.
func (e *TrafficGenConfigEditor) Show() {
	e.app.pages.SwitchToPage("trafficgenconfig")

	e.app.app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		//exhaustive:ignore
		switch event.Key() {
		case tcell.KeyF5:
			if e.onSave != nil {
				e.onSave(*e.config)
			}
			e.Hide()
			return nil
		case tcell.KeyEscape:
			if e.onCancel != nil {
				e.onCancel()
			}
			e.Hide()
			return nil
		default:
			return event
		}
	})
}

// Hide hides the editor.
func (e *TrafficGenConfigEditor) Hide() {
	e.app.pages.SwitchToPage("main")
	e.app.build()
}

// ShowTrafficGenConfig shows the traffic generator configuration editor.
func (a *App) ShowTrafficGenConfig(config *TrafficGenConfig, onSave func(TrafficGenConfig)) {
	editor := a.NewTrafficGenConfigEditor(config, onSave, nil)
	editor.Show()
}
