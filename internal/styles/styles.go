package styles

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// ── Marathon Color Palette ──────────────────────────────────────────────────
// "Graphic Retro Futurism" — Y2K Cyberpunk × The Designers Republic

var (
	NeonLime  = lipgloss.Color("#BFFF00")
	HotPink   = lipgloss.Color("#FF007F")
	Cyan      = lipgloss.Color("#00F0FF")
	DeepBlack = lipgloss.Color("#0A0A0A")
	DarkNavy  = lipgloss.Color("#12121F")
	Charcoal  = lipgloss.Color("#1A1A2E")
	MutedGrey = lipgloss.Color("#3E3E50")
	DimGrey   = lipgloss.Color("#2A2A3A")
	OffWhite  = lipgloss.Color("#D8D8E8")
	BrightWht = lipgloss.Color("#FFFFFF")
	Amber     = lipgloss.Color("#FFB800")
	Red       = lipgloss.Color("#FF2244")
	DimCyan   = lipgloss.Color("#0A3F3F")
	DimLime   = lipgloss.Color("#1A2A00")
	DimPink   = lipgloss.Color("#2A0015")
)

// Dimmed mode variants
var (
	DimmedLime = lipgloss.Color("#6F8F00")
	DimmedPink = lipgloss.Color("#8F003F")
	DimmedCyan = lipgloss.Color("#007F8F")
)

// Dimmed toggles the color intensity
var Dimmed = false

func Primary() lipgloss.Color {
	if Dimmed {
		return DimmedLime
	}
	return NeonLime
}

func Secondary() lipgloss.Color {
	if Dimmed {
		return DimmedPink
	}
	return HotPink
}

func Tertiary() lipgloss.Color {
	if Dimmed {
		return DimmedCyan
	}
	return Cyan
}

// ── Panel Rendering ─────────────────────────────────────────────────────────

// Panel renders content inside a Marathon-style bordered box.
// The width/height are the OUTER dimensions including the border.
func Panel(title string, content string, width, height int) string {
	innerW := width - 2 // left + right border chars
	if innerW < 1 {
		innerW = 1
	}
	innerH := height - 2 // top + bottom border lines
	if innerH < 1 {
		innerH = 1
	}

	// Build top border with title
	topBorder := buildTopBorder(title, innerW)

	// Build content area — pad or truncate to exact innerH lines
	contentLines := strings.Split(content, "\n")

	var body strings.Builder
	for i := 0; i < innerH; i++ {
		line := ""
		if i < len(contentLines) {
			line = contentLines[i]
		}

		// Trim or pad to exact width
		lineW := lipgloss.Width(line)
		if lineW > innerW {
			// Truncate — find a safe cut point
			line = truncateToWidth(line, innerW)
		} else if lineW < innerW {
			line = line + strings.Repeat(" ", innerW-lineW)
		}

		borderChar := lipgloss.NewStyle().Foreground(MutedGrey).Render("│")
		body.WriteString(borderChar + line + borderChar + "\n")
	}

	// Bottom border
	bottomBorder := lipgloss.NewStyle().Foreground(MutedGrey).Render(
		"└" + strings.Repeat("─", innerW) + "┘")

	return topBorder + "\n" + body.String() + bottomBorder
}

func buildTopBorder(title string, innerW int) string {
	if title == "" {
		return lipgloss.NewStyle().Foreground(Tertiary()).Render(
			"┌" + strings.Repeat("─", innerW) + "┐")
	}

	titleRendered := lipgloss.NewStyle().Foreground(Primary()).Bold(true).Render("┤") +
		lipgloss.NewStyle().Foreground(BrightWht).Bold(true).Render(" "+title+" ") +
		lipgloss.NewStyle().Foreground(Primary()).Bold(true).Render("├")

	titleVisualWidth := lipgloss.Width(titleRendered)

	leftDash := 1
	rightDash := innerW - titleVisualWidth - leftDash
	if rightDash < 0 {
		rightDash = 0
	}

	left := lipgloss.NewStyle().Foreground(Tertiary()).Render("┌" + strings.Repeat("─", leftDash))
	right := lipgloss.NewStyle().Foreground(Tertiary()).Render(strings.Repeat("─", rightDash) + "┐")

	return left + titleRendered + right
}

// truncateToWidth cuts a string to fit within a visual width.
func truncateToWidth(s string, maxW int) string {
	if maxW <= 0 {
		return ""
	}
	runes := []rune(s)
	// Simple byte-level truncation — good enough for monospace
	for len(runes) > 0 && lipgloss.Width(string(runes)) > maxW {
		runes = runes[:len(runes)-1]
	}
	return string(runes)
}

// ── Bar Rendering ───────────────────────────────────────────────────────────

// Bar draws a compact horizontal usage bar with percentage.
func Bar(percent float64, width int, fg lipgloss.Color) string {
	if width < 6 {
		return fmt.Sprintf("%3.0f%%", percent)
	}

	barW := width - 5 // " 100%"
	if barW < 1 {
		barW = 1
	}

	filled := int(percent / 100.0 * float64(barW))
	if filled > barW {
		filled = barW
	}
	if filled < 0 {
		filled = 0
	}
	empty := barW - filled

	bar := lipgloss.NewStyle().Foreground(fg).Render(strings.Repeat("█", filled)) +
		lipgloss.NewStyle().Foreground(DimGrey).Render(strings.Repeat("░", empty))

	pct := lipgloss.NewStyle().Foreground(OffWhite).Bold(true).Render(fmt.Sprintf(" %3.0f%%", percent))

	return bar + pct
}

// GradientBar renders a bar that changes color based on value.
func GradientBar(percent float64, width int) string {
	return Bar(percent, width, UsageColor(percent))
}

// UsageColor picks a color based on utilization thresholds
func UsageColor(pct float64) lipgloss.Color {
	switch {
	case pct >= 90:
		return Red
	case pct >= 70:
		return Amber
	case pct >= 40:
		return Primary()
	default:
		return Tertiary()
	}
}

// ── Sparkline ───────────────────────────────────────────────────────────────

var sparkChars = []rune{'▁', '▂', '▃', '▄', '▅', '▆', '▇', '█'}

func Sparkline(values []float64, width int, color lipgloss.Color) string {
	if width <= 0 {
		return ""
	}

	// Take last `width` values
	start := 0
	if len(values) > width {
		start = len(values) - width
	}
	visible := values[start:]

	var sb strings.Builder
	for _, v := range visible {
		if v < 0 {
			v = 0
		}
		if v > 100 {
			v = 100
		}
		idx := int(v / 100.0 * float64(len(sparkChars)-1))
		if idx >= len(sparkChars) {
			idx = len(sparkChars) - 1
		}
		sb.WriteRune(sparkChars[idx])
	}
	// Pad
	for i := len(visible); i < width; i++ {
		sb.WriteRune(' ')
	}

	return lipgloss.NewStyle().Foreground(color).Render(sb.String())
}

// ── Utility ─────────────────────────────────────────────────────────────────

func Dim(s string) string {
	return lipgloss.NewStyle().Foreground(MutedGrey).Render(s)
}

func Bright(s string) string {
	return lipgloss.NewStyle().Foreground(OffWhite).Bold(true).Render(s)
}

func Accent(s string) string {
	return lipgloss.NewStyle().Foreground(Primary()).Render(s)
}

func Warn(s string) string {
	return lipgloss.NewStyle().Foreground(Amber).Render(s)
}

func Crit(s string) string {
	return lipgloss.NewStyle().Foreground(Red).Bold(true).Render(s)
}

func Pink(s string) string {
	return lipgloss.NewStyle().Foreground(Secondary()).Render(s)
}

func Teal(s string) string {
	return lipgloss.NewStyle().Foreground(Tertiary()).Render(s)
}

// FormatBytes converts bytes to human-readable format
func FormatBytes(bytes uint64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
		TB = GB * 1024
	)
	switch {
	case bytes >= TB:
		return fmt.Sprintf("%.1fT", float64(bytes)/float64(TB))
	case bytes >= GB:
		return fmt.Sprintf("%.1fG", float64(bytes)/float64(GB))
	case bytes >= MB:
		return fmt.Sprintf("%.1fM", float64(bytes)/float64(MB))
	case bytes >= KB:
		return fmt.Sprintf("%.1fK", float64(bytes)/float64(KB))
	default:
		return fmt.Sprintf("%dB", bytes)
	}
}

func FormatBytesRate(bytes uint64) string {
	return FormatBytes(bytes) + "/s"
}

// Pad right-pads a string to a given visual width
func Pad(s string, width int) string {
	w := lipgloss.Width(s)
	if w >= width {
		return s
	}
	return s + strings.Repeat(" ", width-w)
}
