package components

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/blumenwagen/durandal/internal/metrics"
	"github.com/blumenwagen/durandal/internal/styles"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/lipgloss"
)

// Processes shows a scrollable sorted process list with kill support.
type Processes struct {
	Width     int
	Height    int
	List      []metrics.ProcessInfo
	Cursor    int
	SortByCPU bool
	Offset    int

	// Search state
	FilterInput textinput.Model
	IsFiltering bool
	filterTerm  string

	// Kill state
	KillConfirm    bool   // waiting for Y/N confirmation
	KillErrorPopup string // status message for permission denied popup
	KillResult     string // status message after attempt
	KillTime       time.Time
}

func NewProcesses() Processes {
	ti := textinput.New()
	ti.Placeholder = "Process name or PID..."
	ti.Prompt = "/"
	ti.PromptStyle = lipgloss.NewStyle().Foreground(styles.Tertiary()).Bold(true)
	ti.TextStyle = lipgloss.NewStyle().Foreground(styles.BrightWht)

	return Processes{
		SortByCPU:   true,
		FilterInput: ti,
	}
}

func (p *Processes) Update(procs []metrics.ProcessInfo) {
	// Apply filter
	var filtered []metrics.ProcessInfo
	term := strings.ToLower(p.filterTerm)

	if term == "" {
		filtered = procs
	} else {
		for _, pr := range procs {
			if strings.Contains(strings.ToLower(pr.Name), term) ||
				strings.Contains(strings.ToLower(pr.Command), term) ||
				strings.Contains(fmt.Sprintf("%d", pr.PID), term) {
				filtered = append(filtered, pr)
			}
		}
	}

	p.List = filtered
	if p.Cursor >= len(p.List) {
		p.Cursor = len(p.List) - 1
	}
	if p.Cursor < 0 {
		p.Cursor = 0
	}

	if p.KillResult != "" && time.Since(p.KillTime) > 3*time.Second {
		p.KillResult = ""
	}
}

func (p *Processes) SetFilter(term string) {
	p.filterTerm = term
	// Re-applying filter immediately isn't deeply necessary because snapshotMsg ticks every second,
	// but it makes UI snappier to at least reset cursor
	p.Cursor = 0
}

func (p *Processes) ScrollUp() {
	if p.KillConfirm {
		return
	}
	if p.Cursor > 0 {
		p.Cursor--
	}
}

func (p *Processes) ScrollDown() {
	if p.KillConfirm {
		return
	}
	if p.Cursor < len(p.List)-1 {
		p.Cursor++
	}
}

func (p *Processes) ToggleSort() {
	if p.KillConfirm {
		return
	}
	p.SortByCPU = !p.SortByCPU
}

func (p *Processes) RequestKill() {
	if p.Cursor < 0 || p.Cursor >= len(p.List) {
		return
	}
	proc := p.List[p.Cursor]
	if proc.User == "root" && !styles.IsRoot {
		p.KillErrorPopup = fmt.Sprintf("Cannot kill root process %d (%s) as non-root user", proc.PID, proc.Name)
		return
	}
	p.KillConfirm = true
	p.KillResult = ""
}

func (p *Processes) ConfirmKill() {
	if !p.KillConfirm || p.Cursor < 0 || p.Cursor >= len(p.List) {
		p.KillConfirm = false
		return
	}
	proc := p.List[p.Cursor]

	// Use os-level process handling for cross-platform compat
	osProc, err := os.FindProcess(int(proc.PID))
	if err == nil {
		err = osProc.Kill()
	}

	if err != nil {
		p.KillResult = fmt.Sprintf("✗ kill %d (%s): %v", proc.PID, proc.Name, err)
	} else {
		p.KillResult = fmt.Sprintf("✓ KILL → %d (%s)", proc.PID, proc.Name)
	}
	p.KillTime = time.Now()
	p.KillConfirm = false
}

func (p *Processes) CancelKill() {
	p.KillConfirm = false
	p.KillErrorPopup = ""
}

func (p Processes) View() string {
	if p.KillErrorPopup != "" {
		popup := lipgloss.NewStyle().Foreground(styles.Red).Bold(true).Render("PERMISSION DENIED") + "\n\n" +
			lipgloss.NewStyle().Foreground(styles.BrightWht).Render(p.KillErrorPopup) + "\n\n" +
			styles.Dim("[Press Esc or Enter to close]")

		content := lipgloss.Place(p.Width-2, p.Height-2, lipgloss.Center, lipgloss.Center, popup)
		return styles.Panel("ERROR", content, p.Width, p.Height)
	}

	iw := p.Width - 2
	if iw < 20 {
		iw = 20
	}

	var lines []string

	// Sort indicator
	sortStr := styles.Accent("CPU▼")
	if !p.SortByCPU {
		sortStr = styles.Pink("MEM▼")
	}

	statusText := styles.Dim("sort:") + sortStr + styles.Dim(fmt.Sprintf("  %d procs", len(p.List)))

	if p.filterTerm != "" && !p.IsFiltering {
		statusText += styles.Dim("  filter:") + styles.Teal("/"+p.filterTerm)
	}

	// Status Line: Kill confirm OR filter bar OR standard sort
	statusLine := statusText

	if p.KillConfirm && p.Cursor >= 0 && p.Cursor < len(p.List) {
		proc := p.List[p.Cursor]
		statusLine = styles.Crit(fmt.Sprintf("  KILL %d (%s)? ", proc.PID, proc.Name)) +
			styles.Accent("[y]") + styles.Dim("es ") +
			styles.Accent("[n]") + styles.Dim("o")
	} else if p.KillResult != "" {
		if strings.HasPrefix(p.KillResult, "✓") {
			statusLine = styles.Accent(p.KillResult)
		} else {
			statusLine = styles.Crit(p.KillResult)
		}
	} else if p.IsFiltering {
		statusLine = p.FilterInput.View()
	}

	lines = append(lines, statusLine)

	// Table header
	hdr := fmtProcRow("PID", "COMMAND", "CPU%", "MEM%", "RSS", "USER", iw)
	lines = append(lines, lipgloss.NewStyle().Foreground(styles.Tertiary()).Bold(true).Render(hdr))

	visibleRows := p.Height - 4
	if visibleRows < 1 {
		visibleRows = 1
	}

	if p.Cursor < p.Offset {
		p.Offset = p.Cursor
	}
	if p.Cursor >= p.Offset+visibleRows {
		p.Offset = p.Cursor - visibleRows + 1
	}

	for i := p.Offset; i < len(p.List) && i < p.Offset+visibleRows; i++ {
		proc := p.List[i]
		isSelected := i == p.Cursor

		user := proc.User
		if len(user) > 8 {
			user = user[:8]
		}

		cmd := proc.Command
		if cmd == "" {
			cmd = proc.Name
		}

		row := fmtProcRow(
			fmt.Sprintf("%d", proc.PID), cmd,
			fmt.Sprintf("%.1f", proc.CPU), fmt.Sprintf("%.1f", proc.Memory),
			styles.FormatBytes(proc.MemRSS), user, iw,
		)

		if isSelected {
			bg := styles.Primary()
			if p.KillConfirm {
				bg = styles.Red
			}
			row = lipgloss.NewStyle().Foreground(styles.DeepBlack).Background(bg).Bold(true).Render(row)
		} else if proc.CPU > 50 {
			row = lipgloss.NewStyle().Foreground(styles.Secondary()).Render(row)
		} else if proc.CPU > 20 {
			row = lipgloss.NewStyle().Foreground(styles.Amber).Render(row)
		} else {
			row = lipgloss.NewStyle().Foreground(styles.OffWhite).Render(row)
		}

		lines = append(lines, row)
	}

	return styles.Panel("PROCESSES", strings.Join(lines, "\n"), p.Width, p.Height)
}

func fmtProcRow(pid, name, cpu, mem, rss, user string, maxW int) string {
	fixedW := 41
	nameW := maxW - fixedW
	if nameW < 5 {
		nameW = 5
	}

	dispName := name
	if len(dispName) > nameW {
		dispName = dispName[:nameW-1] + "…"
	} else if len(dispName) < nameW {
		dispName = dispName + strings.Repeat(" ", nameW-len(dispName))
	}

	s := fmt.Sprintf(" %-7s %s %6s %6s %8s %-8s", pid, dispName, cpu, mem, rss, user)
	if len(s) > maxW {
		s = s[:maxW]
	}
	return s
}
