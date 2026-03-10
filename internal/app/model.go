package app

import (
	"sort"
	"time"

	"github.com/blumenwagen/durandal/internal/components"
	"github.com/blumenwagen/durandal/internal/metrics"
	"github.com/blumenwagen/durandal/internal/styles"
	tea "github.com/charmbracelet/bubbletea"
)

type tickMsg time.Time
type snapshotMsg metrics.Snapshot

// Model is the root Bubble Tea model.
type Model struct {
	Width  int
	Height int

	Header    components.Header
	CPU       components.CPU
	Memory    components.Memory
	Processes components.Processes
	Network   components.Network
	Disk      components.Disk

	ready    bool
	quitting bool
}

func NewModel() Model {
	return Model{
		CPU:       components.NewCPU(),
		Memory:    components.NewMemory(),
		Processes: components.NewProcesses(),
		Network:   components.NewNetwork(),
		Disk:      components.NewDisk(),
		Header:    components.NewHeader(),
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(tickCmd(), collectCmd())
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			m.quitting = true
			return m, tea.Quit
		case "j", "down":
			m.Processes.ScrollDown()
		case "k", "up":
			m.Processes.ScrollUp()
		case "s", "tab":
			m.Processes.ToggleSort()
		case "K": // Shift+K to kill
			_ = m.Processes.KillSelected()
		case "d":
			styles.Dimmed = !styles.Dimmed
		}

	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
		m.ready = true
		m = calculateLayout(m)

	case tickMsg:
		return m, tea.Batch(tickCmd(), collectCmd())

	case snapshotMsg:
		snap := metrics.Snapshot(msg)
		m.Header.Host = snap.Host
		m.CPU.Update(snap.CPU)
		m.Memory.Update(snap.Memory, snap.Sensors)

		procs := snap.Processes
		if !m.Processes.SortByCPU {
			sort.Slice(procs, func(i, j int) bool {
				return procs[i].Memory > procs[j].Memory
			})
		}
		m.Processes.Update(procs)
		m.Network.Update(snap.Network)
		m.Disk.Update(snap.Disks)
	}

	return m, nil
}

func (m Model) View() string {
	if m.quitting {
		return ""
	}
	if !m.ready {
		return "\n  " + styles.Accent("《") + styles.Bright(" DURANDAL ") + styles.Accent("》") + " initializing…"
	}
	return renderLayout(m)
}

func tickCmd() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func collectCmd() tea.Cmd {
	return func() tea.Msg {
		snap, _ := metrics.CollectSnapshot()
		return snapshotMsg(snap)
	}
}
