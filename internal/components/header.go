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

func (h Header) View() string {
	if h.Width < 30 {
		return ""
	}
	w := h.Width

	// Glitched texts
	durandalTxt := glitchString(" DURANDAL ", 0.08, 3)
	sysMonTxt := glitchString("SYSTEM MONITOR", 0.03, 2)

	// Left: title
	title := styles.Accent(glitchString("《", 0.01, 1)) +
		styles.Bright(durandalTxt) +
		styles.Accent(glitchString("》", 0.01, 1)) +
		" " + styles.Teal(sysMonTxt)

	// Right: system info
	host := styles.Bright(glitchString(h.Host.Hostname, 0.02, 2))
	kern := styles.Dim(h.Host.Kernel)
	up := styles.Pink("▲ " + glitchString(h.Host.Uptime, 0.02, 1))

	dimTag := ""
	if styles.Dimmed {
		dimTag = styles.Dim(" [DIM]")
	}

	right := host + " " + kern + " " + up + dimTag

	// Fill gap between left and right
	titleW := lipgloss.Width(title)
	rightW := lipgloss.Width(right)
	gap := w - titleW - rightW - 2
	if gap < 1 {
		gap = 1
	}

	fill := styles.Dim(strings.Repeat("─", gap))
	// apply occasional heavy glitch to the scanline itself
	if rand.Float64() < 0.05 {
		fill = styles.Dim(glitchString(strings.Repeat("─", gap), 1.0, gap/4))
	}

	return title + " " + fill + " " + right
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
