package components

import (
	"strings"

	"github.com/blumenwagen/durandal/internal/metrics"
	"github.com/blumenwagen/durandal/internal/styles"
)

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

	// Left: title
	title := styles.Accent("《") +
		styles.Bright(" DURANDAL ") +
		styles.Accent("》") +
		" " + styles.Teal("SYSTEM MONITOR")

	// Right: system info
	host := styles.Bright(h.Host.Hostname)
	kern := styles.Dim(h.Host.Kernel)
	up := styles.Pink("▲ " + h.Host.Uptime)

	dimTag := ""
	if styles.Dimmed {
		dimTag = styles.Dim(" [DIM]")
	}

	right := host + " " + kern + " " + up + dimTag

	// Fill gap between left and right
	titleW := lipglossWidth(title)
	rightW := lipglossWidth(right)
	gap := w - titleW - rightW - 2
	if gap < 1 {
		gap = 1
	}

	fill := styles.Dim(strings.Repeat("─", gap))

	return title + " " + fill + " " + right
}

// HelpBar renders the bottom keybinding reference bar.
func HelpBar(w int, dimmed bool) string {
	keys := []struct{ key, desc string }{
		{"↑/k", "up"},
		{"↓/j", "down"},
		{"s", "sort"},
		{"K", "kill"},
		{"d", "dim"},
		{"q", "quit"},
	}

	var parts []string
	for _, k := range keys {
		parts = append(parts, styles.Accent(k.key)+styles.Dim(" "+k.desc))
	}
	bar := strings.Join(parts, styles.Dim("  │  "))

	barW := lipglossWidth(bar)
	if barW >= w {
		return bar
	}

	// Center the bar
	pad := (w - barW) / 2
	return strings.Repeat(" ", pad) + bar
}

// lipglossWidth is a helper to avoid importing lipgloss everywhere
func lipglossWidth(s string) int {
	// Count visible width — iterate runes, skip ANSI escapes
	w := 0
	inEscape := false
	for _, r := range s {
		if r == '\033' {
			inEscape = true
			continue
		}
		if inEscape {
			if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') {
				inEscape = false
			}
			continue
		}
		w++
	}
	return w
}
