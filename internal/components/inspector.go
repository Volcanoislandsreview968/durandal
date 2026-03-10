package components

import (
	"fmt"
	"strings"

	"github.com/blumenwagen/durandal/internal/metrics"
	"github.com/blumenwagen/durandal/internal/styles"
	"github.com/charmbracelet/lipgloss"
	"github.com/shirou/gopsutil/v4/process"
)

// Inspector provides a detailed view of a single process.
type Inspector struct {
	Width  int
	Height int
	Proc   metrics.ProcessInfo
}

func NewInspector() Inspector {
	return Inspector{}
}

func (i *Inspector) Update(p metrics.ProcessInfo) {
	i.Proc = p
}

func (i Inspector) View() string {
	if i.Width <= 0 || i.Height <= 0 {
		return ""
	}

	iw := i.Width - 2
	if iw < 20 {
		iw = 20
	}

	var lines []string

	// Title / Banner
	header := lipgloss.NewStyle().Foreground(styles.DeepBlack).Background(styles.Secondary()).Bold(true).Render(fmt.Sprintf(" PROCESS %d: %s ", i.Proc.PID, i.Proc.Name))
	lines = append(lines, header)
	lines = append(lines, "")

	// Status Line
	lines = append(lines, styles.Dim("USER:   ")+styles.Bright(i.Proc.User))
	lines = append(lines, styles.Dim("STATUS: ")+styles.Pink(i.Proc.Status))
	lines = append(lines, "")

	// CPU & MEM
	lines = append(lines, styles.Dim("CPU:    ")+styles.Accent(fmt.Sprintf("%.1f%%", i.Proc.CPU)))
	lines = append(lines, styles.Dim("MEM:    ")+styles.Teal(fmt.Sprintf("%.1f%%", i.Proc.Memory))+styles.Dim(" (RSS: "+styles.FormatBytes(i.Proc.MemRSS)+")"))
	lines = append(lines, "")

	// Detailed Fetching via gopsutil
	p, err := process.NewProcess(i.Proc.PID)
	if err == nil {
		threads, _ := p.NumThreads()
		lines = append(lines, styles.Dim("THREADS:       ")+styles.Bright(fmt.Sprintf("%d", threads)))

		cwd, _ := p.Cwd()
		if cwd != "" {
			lines = append(lines, styles.Dim("CWD:           ")+styles.Bright(cwd))
		}

		nice, _ := p.Nice()
		lines = append(lines, styles.Dim("NICE:          ")+styles.Bright(fmt.Sprintf("%d", nice)))

		io, err := p.IOCounters()
		if err == nil {
			lines = append(lines, "")
			lines = append(lines, styles.Dim("DISK READ:     ")+styles.Bright(styles.FormatBytes(io.ReadBytes)))
			lines = append(lines, styles.Dim("DISK WRITE:    ")+styles.Bright(styles.FormatBytes(io.WriteBytes)))
		}
		lines = append(lines, "")
	}

	// Command Wrapping
	lines = append(lines, styles.Dim("COMMAND LINE:"))
	cmd := i.Proc.Command
	if cmd == "" {
		cmd = i.Proc.Name
	}

	// Poor man's text wrapping for the command to fit 'iw'
	cmdLines := wrapString(cmd, iw)
	for _, cl := range cmdLines {
		lines = append(lines, "  "+styles.Bright(cl))
	}

	// Bottom instructions
	lines = append(lines, "")
	lines = append(lines, styles.Dim("Press ")+styles.Pink("[ESC]")+styles.Dim(" or ")+styles.Pink("[ENTER]")+styles.Dim(" to close inspector"))

	return styles.Panel("INSPECTOR", strings.Join(lines, "\n"), i.Width, i.Height)
}

func wrapString(s string, w int) []string {
	if w <= 0 {
		return []string{s}
	}
	var res []string
	for len(s) > w {
		res = append(res, s[:w])
		s = s[w:]
	}
	if len(s) > 0 {
		res = append(res, s)
	}
	return res
}
