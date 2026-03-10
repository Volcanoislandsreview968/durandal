package app

import (
	"github.com/blumenwagen/durandal/internal/components"
	"github.com/charmbracelet/lipgloss"
)

// calculateLayout distributes terminal dimensions strictly so nothing overflows.
// Every pixel is accounted for. The total of all heights MUST equal m.Height.
func calculateLayout(m Model) Model {
	w := m.Width
	h := m.Height

	if w < 40 || h < 15 {
		return m
	}

	// Reserve fixed rows: header (1) + help bar (1) = 2
	usableH := h - 2

	// Vertical split: top panels | process list | bottom panels
	// Proportions: 35% / 40% / 25% — clamped
	topH := usableH * 35 / 100
	bottomH := usableH * 25 / 100
	procH := usableH - topH - bottomH

	// Minimums
	if topH < 6 {
		topH = 6
	}
	if bottomH < 6 {
		bottomH = 6
	}
	if procH < 5 {
		procH = 5
	}

	// Rebalance if we overshoot
	total := topH + procH + bottomH
	if total > usableH {
		excess := total - usableH
		// Shrink process list first, then bottom, then top
		if procH-excess >= 5 {
			procH -= excess
		} else {
			procH = 5
			excess = topH + 5 + bottomH - usableH
			if bottomH-excess >= 4 {
				bottomH -= excess
			} else {
				bottomH = 4
				topH = usableH - procH - bottomH
			}
		}
	}

	// Header — uses full width, 1 line
	m.Header.Width = w

	// Top row: CPU (60%) | MEM (40%)
	cpuW := w * 60 / 100
	memW := w - cpuW
	m.CPU.Width = cpuW
	m.CPU.Height = topH
	m.Memory.Width = memW
	m.Memory.Height = topH

	// Middle: process list — full width
	m.Processes.Width = w
	m.Processes.Height = procH

	// Bottom: NET (50%) | DISK (50%)
	netW := w / 2
	diskW := w - netW
	m.Network.Width = netW
	m.Network.Height = bottomH
	m.Disk.Width = diskW
	m.Disk.Height = bottomH

	return m
}

// renderLayout composes all views into the final screen.
func renderLayout(m Model) string {
	header := m.Header.View()

	topRow := lipgloss.JoinHorizontal(lipgloss.Top,
		m.CPU.View(),
		m.Memory.View(),
	)

	procs := m.Processes.View()

	bottomRow := lipgloss.JoinHorizontal(lipgloss.Top,
		m.Network.View(),
		m.Disk.View(),
	)

	helpBar := components.HelpBar(m.Width, false)

	return lipgloss.JoinVertical(lipgloss.Left,
		header,
		topRow,
		procs,
		bottomRow,
		helpBar,
	)
}
