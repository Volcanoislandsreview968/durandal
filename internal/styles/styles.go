package styles

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var IsRoot = false

func init() {
	if os.Geteuid() == 0 {
		IsRoot = true
	}
}

// ── Marathon Color Palette ──────────────────────────────────────────────────

var (
	NeonLime  = lipgloss.Color("#BFFF00")
	HotPink   = lipgloss.Color("#FF007F")
	Cyan      = lipgloss.Color("#00F0FF")
	DeepBlack = lipgloss.Color("#0A0A0A")
	DarkNavy  = lipgloss.Color("#12121F")
	Charcoal  = lipgloss.Color("#1A1A2E")
	MutedGrey = lipgloss.Color("#3E3E50")
	DimGrey   = lipgloss.Color("#1F1F2E")
	OffWhite  = lipgloss.Color("#D8D8E8")
	BrightWht = lipgloss.Color("#FFFFFF")
	Amber     = lipgloss.Color("#FFB800")
	Red       = lipgloss.Color("#FF2244")
	DimCyan   = lipgloss.Color("#053333")
	DimLime   = lipgloss.Color("#1A2A00")
	DimPink   = lipgloss.Color("#2A0015")
)

var (
	DimmedLime = lipgloss.Color("#6F8F00")
	DimmedPink = lipgloss.Color("#8F003F")
	DimmedCyan = lipgloss.Color("#007F8F")

	RootLime = lipgloss.Color("#FF3300")
	RootPink = lipgloss.Color("#FF0033")
	RootCyan = lipgloss.Color("#FF8800")

	RootDimmedLime = lipgloss.Color("#8F1100")
	RootDimmedPink = lipgloss.Color("#8F0011")
	RootDimmedCyan = lipgloss.Color("#8F4400")
)

var Dimmed = false

func Primary() lipgloss.Color {
	if IsRoot {
		if Dimmed {
			return RootDimmedLime
		}
		return RootLime
	}
	if Dimmed {
		return DimmedLime
	}
	return NeonLime
}

func Secondary() lipgloss.Color {
	if IsRoot {
		if Dimmed {
			return RootDimmedPink
		}
		return RootPink
	}
	if Dimmed {
		return DimmedPink
	}
	return HotPink
}

func Tertiary() lipgloss.Color {
	if IsRoot {
		if Dimmed {
			return RootDimmedCyan
		}
		return RootCyan
	}
	if Dimmed {
		return DimmedCyan
	}
	return Cyan
}

// ── Panel Rendering ─────────────────────────────────────────────────────────

func TechPanel(title string, content string, width, height int, accent lipgloss.Color) string {
	innerW := width - 2
	if innerW < 1 {
		innerW = 1
	}
	innerH := height - 2
	if innerH < 1 {
		innerH = 1
	}

	topBorder := buildSciFiBorder(title, innerW, accent)
	contentLines := strings.Split(content, "\n")

	var body strings.Builder

	// Create rigid vertical side borders that match the accent color
	borderStyle := lipgloss.NewStyle().Foreground(accent)

	for i := 0; i < innerH; i++ {
		line := ""
		if i < len(contentLines) {
			line = contentLines[i]
		}

		// We add a literal space margin inside the text area for breathability
		paddedLine := " " + line
		lineW := lipgloss.Width(paddedLine)

		if lineW > innerW-1 { // -1 to preserve right margin space
			paddedLine = truncateToWidth(paddedLine, innerW-1) + " "
		} else if lineW < innerW {
			paddedLine = paddedLine + strings.Repeat(" ", innerW-lineW)
		}

		borderChar := borderStyle.Render("▌")
		body.WriteString(borderChar + paddedLine + borderStyle.Render("▐") + "\n")
	}

	bottomBorder := borderStyle.Render("▙" + strings.Repeat("▄", innerW) + "▟")

	return topBorder + "\n" + body.String() + bottomBorder
}

func buildSciFiBorder(title string, innerW int, accent lipgloss.Color) string {
	accentStyle := lipgloss.NewStyle().Foreground(accent)

	if title == "" {
		return accentStyle.Render("▛" + strings.Repeat("▀", innerW) + "▜")
	}

	// ▛▀▀ [ T I T L E ] ▀▀▀▀▀▀▀▀▀▀ /// ▀▀▜
	trackedTitle := TextTracking(strings.ToUpper(title))

	titleBracketL := lipgloss.NewStyle().Foreground(BrightWht).Bold(true).Render("[ ")
	titleText := lipgloss.NewStyle().Foreground(DeepBlack).Background(accent).Bold(true).Render(trackedTitle)
	titleBracketR := lipgloss.NewStyle().Foreground(BrightWht).Bold(true).Render(" ]")

	// Assembly the title portion
	fullTitle := titleBracketL + titleText + titleBracketR
	titleVisualWidth := lipgloss.Width(fullTitle)

	leftDashCount := 2

	// Calculate the remaining space on the right side
	rightDashCount := innerW - titleVisualWidth - leftDashCount
	rightSide := ""

	if rightDashCount >= 7 {
		// Embed the visual /// decor directly into the structural border
		mainDashCount := rightDashCount - 6
		rightSide = accentStyle.Render(strings.Repeat("▀", mainDashCount)) +
			lipgloss.NewStyle().Foreground(BrightWht).Bold(true).Render(" /// ") +
			accentStyle.Render("▀▜")
	} else if rightDashCount >= 0 {
		rightSide = accentStyle.Render(strings.Repeat("▀", rightDashCount) + "▜")
	}

	left := accentStyle.Render("▛" + strings.Repeat("▀", leftDashCount))

	return left + fullTitle + rightSide
}

func TextTracking(s string) string {
	res := ""
	for i, c := range s {
		res += string(c)
		if i < len(s)-1 {
			res += " "
		}
	}
	return res
}

// ── Bar Rendering ───────────────────────────────────────────────────────────

func Bar(percent float64, width int, fg lipgloss.Color) string {
	if width < 6 {
		return fmt.Sprintf("%3.0f%%", percent)
	}

	barW := width - 5
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

	// visually richer blocks
	filledStr := strings.Repeat("█", filled)
	emptyStr := strings.Repeat("⡀", empty) // subtle dot/line instead of heavy block for empty

	bar := lipgloss.NewStyle().Foreground(fg).Render(filledStr) +
		lipgloss.NewStyle().Foreground(DimGrey).Render(emptyStr)

	pct := lipgloss.NewStyle().Foreground(OffWhite).Bold(true).Render(fmt.Sprintf(" %3.0f%%", percent))

	return bar + pct
}

func GradientBar(percent float64, width int) string {
	return Bar(percent, width, UsageColor(percent))
}

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

var sparkChars = []rune{' ', '▂', '▃', '▄', '▅', '▆', '▇', '█'}

func Sparkline(values []float64, width int, color lipgloss.Color) string {
	return MultiSparkline(values, width, 1, color)
}

func MultiSparkline(values []float64, width, height int, color lipgloss.Color) string {
	if width <= 0 || height <= 0 {
		return ""
	}

	start := 0
	if len(values) > width {
		start = len(values) - width
	}
	visible := values[start:]

	// Pad with zeros to fill width
	data := make([]float64, width)
	offset := width - len(visible)
	for i, v := range visible {
		data[offset+i] = v
	}

	var lines []string
	for r := height - 1; r >= 0; r-- {
		var sb strings.Builder
		rowBottom := float64(r) / float64(height) * 100.0
		rowTop := float64(r+1) / float64(height) * 100.0
		rowRange := rowTop - rowBottom

		for _, v := range data {
			if v > 100 {
				v = 100
			}
			if v >= rowTop {
				sb.WriteRune('█')
			} else if v <= rowBottom {
				sb.WriteRune(' ')
			} else {
				// Partial fill
				frac := (v - rowBottom) / rowRange
				idx := 1 + int(frac*float64(len(sparkChars)-2))
				if idx >= len(sparkChars) {
					idx = len(sparkChars) - 1
				}
				if idx < 0 {
					idx = 0
				}
				sb.WriteRune(sparkChars[idx])
			}
		}
		lines = append(lines, lipgloss.NewStyle().Foreground(color).Render(sb.String()))
	}

	return strings.Join(lines, "\n")
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

func Pad(s string, width int) string {
	w := lipgloss.Width(s)
	if w >= width {
		return s
	}
	return s + strings.Repeat(" ", width-w)
}

func truncateToWidth(s string, maxW int) string {
	if maxW <= 0 {
		return ""
	}
	runes := []rune(s)
	for len(runes) > 0 && lipgloss.Width(string(runes)) > maxW {
		runes = runes[:len(runes)-1]
	}
	return string(runes)
}
