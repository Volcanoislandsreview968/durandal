package components

import (
	"fmt"
	"strings"
	"syscall"

	"github.com/blumenwagen/durandal/internal/metrics"
	"github.com/blumenwagen/durandal/internal/styles"
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
}

func NewProcesses() Processes {
	return Processes{SortByCPU: true}
}

func (p *Processes) Update(procs []metrics.ProcessInfo) {
	p.List = procs
	if p.Cursor >= len(p.List) {
		p.Cursor = len(p.List) - 1
	}
	if p.Cursor < 0 {
		p.Cursor = 0
	}
}

func (p *Processes) ScrollUp() {
	if p.Cursor > 0 {
		p.Cursor--
	}
}

func (p *Processes) ScrollDown() {
	if p.Cursor < len(p.List)-1 {
		p.Cursor++
	}
}

func (p *Processes) ToggleSort() {
	p.SortByCPU = !p.SortByCPU
}

// KillSelected sends SIGTERM to the selected process.
func (p *Processes) KillSelected() error {
	if p.Cursor < 0 || p.Cursor >= len(p.List) {
		return nil
	}
	pid := p.List[p.Cursor].PID
	return syscall.Kill(int(pid), syscall.SIGTERM)
}

func (p Processes) View() string {
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
	lines = append(lines, styles.Dim("sort:")+sortStr+
		styles.Dim(fmt.Sprintf("  %d procs", len(p.List))))

	// Table header
	hdr := fmtProcRow("PID", "NAME", "CPU%", "MEM%", "RSS", "USER", iw)
	lines = append(lines, lipgloss.NewStyle().
		Foreground(styles.Tertiary()).
		Bold(true).
		Render(hdr))

	// Visible rows
	visibleRows := p.Height - 4 // title border + header + sort + bottom border
	if visibleRows < 1 {
		visibleRows = 1
	}

	// Keep cursor in view
	if p.Cursor < p.Offset {
		p.Offset = p.Cursor
	}
	if p.Cursor >= p.Offset+visibleRows {
		p.Offset = p.Cursor - visibleRows + 1
	}

	for i := p.Offset; i < len(p.List) && i < p.Offset+visibleRows; i++ {
		proc := p.List[i]
		isSelected := i == p.Cursor

		name := proc.Name
		if len(name) > 15 {
			name = name[:12] + "…"
		}
		user := proc.User
		if len(user) > 8 {
			user = user[:8]
		}

		row := fmtProcRow(
			fmt.Sprintf("%d", proc.PID),
			name,
			fmt.Sprintf("%.1f", proc.CPU),
			fmt.Sprintf("%.1f", proc.Memory),
			styles.FormatBytes(proc.MemRSS),
			user,
			iw,
		)

		if isSelected {
			row = lipgloss.NewStyle().
				Foreground(styles.DeepBlack).
				Background(styles.Primary()).
				Bold(true).
				Render(row)
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
	s := fmt.Sprintf(" %-7s %-15s %6s %6s %8s %-8s", pid, name, cpu, mem, rss, user)
	if len(s) > maxW {
		s = s[:maxW]
	}
	return s
}
