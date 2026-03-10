package components

import (
	"math/rand"
	"strings"

	"github.com/blumenwagen/durandal/internal/metrics"
	"github.com/blumenwagen/durandal/internal/styles"
	"github.com/charmbracelet/lipgloss"
)

func glitchString(s string, chance float64, intensity int) string {
	if rand.Float64() > chance {
		return s
	}
	chars := []rune(s)
	if len(chars) == 0 {
		return s
	}
	if intensity <= 0 {
		intensity = 1
	}
	numGlitches := rand.Intn(intensity) + 1
	glitchChars := []rune("█▓▒░×÷±#@_/")
	for i := 0; i < numGlitches; i++ {
		idx := rand.Intn(len(chars))
		gIdx := rand.Intn(len(glitchChars))
		chars[idx] = glitchChars[gIdx]
	}
	return string(chars)
}

// Header renders the top HUD bar — no border box, just a single styled line.
type Header struct {
	Width int
	Host  metrics.HostInfo
}

func NewHeader() Header { return Header{} }

func (h *Header) Update(host metrics.HostInfo) {
	h.Host = host
}

func (h Header) View() string {
	if h.Width <= 0 {
		return ""
	}

	left := lipgloss.NewStyle().Foreground(styles.NeonLime).Bold(true).Render(styles.TextTracking("D U R A N D A L "))
	left += lipgloss.NewStyle().Foreground(styles.BrightWht).Render(" /// S Y S T E M S")

	// We safely extract username from host object
	username := h.Host.User
	if username == "" {
		username = "system"
	}

	right := styles.Dim("USER: ") + lipgloss.NewStyle().Foreground(styles.Primary()).Render(username)
	if h.Host.Hostname != "" {
		right += styles.Dim(" @ ") + styles.Bright(h.Host.Hostname)
	}

	// Uptime is pre-formatted as string inside metrics.HostInfo
	right += styles.Dim(" // UPTIME: ") + styles.Accent(h.Host.Uptime)

	leftW := lipgloss.Width(left)
	rightW := lipgloss.Width(right)

	fillAmt := h.Width - leftW - rightW
	if fillAmt < 0 {
		fillAmt = 0
	}

	// ▛▀▀▀▀▀▀▀   ▀▀▀▀▀▀▀▜
	// ▌ DURANDAL ...    ▐
	// ▙▄▄▄▄▄▄▄   ▄▄▄▄▄▄▄▟
	// Instead of a full TechPanel, the header is just a structured bar
	barText := left + strings.Repeat(" ", fillAmt) + right
	return lipgloss.NewStyle().Background(styles.DarkNavy).Render(barText)
}

// HelpBar renders the bottom keybinding reference bar.
func HelpBar(w int, dimmed bool) string {
	keys := []struct{ key, desc string }{
		{"↑/k", "up"},
		{"↓/j", "down"},
		{"s", "sort"},
		{"/", "search"},
		{"⏎", "inspect"},
		{"K", "kill"},
		{"d", "dim"},
		{"q", "quit"},
	}

	var parts []string
	for _, k := range keys {
		parts = append(parts, styles.Accent(k.key)+styles.Dim(" "+k.desc))
	}
	bar := strings.Join(parts, styles.Dim("  │  "))

	barW := lipgloss.Width(bar)
	if barW >= w {
		return bar
	}

	// Center the bar
	pad := (w - barW) / 2
	return strings.Repeat(" ", pad) + bar
}
