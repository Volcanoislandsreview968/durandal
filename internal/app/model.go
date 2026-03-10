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
type glitchTickMsg time.Time

// Model is the root Bubble Tea model.
type Model struct {
	Width  int
	Height int

	Header        components.Header
	CPU           components.CPU
	GPU           components.GPU
	Memory        components.Memory
	Processes     components.Processes
	Network       components.Network
	Disk          components.Disk
	Inspector     components.Inspector
	InspectorOpen bool

	ProcY    int // Y-coordinate where process list starts
	ready    bool
	quitting bool
}

func NewModel() Model {
	return Model{
		CPU:       components.NewCPU(),
		GPU:       components.NewGPU(),
		Memory:    components.NewMemory(),
		Processes: components.NewProcesses(),
		Network:   components.NewNetwork(),
		Disk:      components.NewDisk(),
		Header:    components.NewHeader(),
		Inspector: components.NewInspector(),
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(tickCmd(), collectCmd(), glitchTickCmd())
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Inspector view keys
		if m.InspectorOpen {
			switch msg.String() {
			case "esc", "enter", "q":
				if msg.String() == "q" {
					m.quitting = true
					return m, tea.Quit
				}
				m.InspectorOpen = false
			case "d":
				styles.Dimmed = !styles.Dimmed
			}
			return m, nil
		}

		// If process list is filtering, route keys to textinput
		if m.Processes.IsFiltering {
			switch msg.String() {
			case "esc", "enter":
				m.Processes.IsFiltering = false
			default:
				var cmd tea.Cmd
				m.Processes.FilterInput, cmd = m.Processes.FilterInput.Update(msg)
				m.Processes.SetFilter(m.Processes.FilterInput.Value())
				return m, cmd
			}
			return m, nil
		}

		// If process list is in kill confirmation mode, intercept specific keys
		if m.Processes.KillConfirm {
			switch msg.String() {
			case "y", "Y":
				m.Processes.ConfirmKill()
			case "n", "N", "esc", "q", "ctrl+c":
				m.Processes.CancelKill()
			}
			return m, nil
		}

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
		case "/":
			m.Processes.IsFiltering = true
			m.Processes.FilterInput.Focus()
		case "enter":
			if len(m.Processes.List) > 0 && m.Processes.Cursor >= 0 {
				m.InspectorOpen = true
				m.Inspector.Update(m.Processes.List[m.Processes.Cursor])
			}
		case "K": // Shift+K to kill
			m.Processes.RequestKill()
		case "d":
			styles.Dimmed = !styles.Dimmed
		}

	case tea.MouseMsg:
		if m.InspectorOpen {
			return m, nil // Block process interactions behind inspector
		}
		switch msg.Type {
		case tea.MouseWheelUp:
			m.Processes.ScrollUp()
		case tea.MouseWheelDown:
			m.Processes.ScrollDown()
		case tea.MouseLeft:
			contentStartY := m.ProcY + 3
			contentEndY := m.ProcY + m.Processes.Height - 1
			if msg.Y >= contentStartY && msg.Y < contentEndY {
				clickedIdx := msg.Y - contentStartY + m.Processes.Offset
				if clickedIdx >= 0 && clickedIdx < len(m.Processes.List) {
					m.Processes.Cursor = clickedIdx
					if m.Processes.KillConfirm {
						m.Processes.CancelKill()
					}
				}
			}
		}

	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
		m.ready = true
		m = calculateLayout(m)

	case glitchTickMsg:
		// Fast loop (100ms) for aesthetic UI updates like glitches
		return m, glitchTickCmd()

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

		// Update Inspector data if open
		if m.InspectorOpen && len(m.Processes.List) > 0 {
			m.Inspector.Update(m.Processes.List[m.Processes.Cursor])
		}

		m.GPU.Update(snap.GPUs)

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

func glitchTickCmd() tea.Cmd {
	return tea.Tick(time.Millisecond*150, func(t time.Time) tea.Msg {
		return glitchTickMsg(t)
	})
}

func collectCmd() tea.Cmd {
	return func() tea.Msg {
		snap, _ := metrics.CollectSnapshot()
		return snapshotMsg(snap)
	}
}
