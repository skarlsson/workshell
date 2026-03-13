package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/skarlsson/ws-manager/internal/config"
	"github.com/skarlsson/ws-manager/internal/state"
	"github.com/spf13/cobra"
)

var dashboardCmd = &cobra.Command{
	Use:   "dashboard",
	Short: "Interactive TUI dashboard for managing workspaces",
	RunE: func(cmd *cobra.Command, args []string) error {
		p := tea.NewProgram(newDashboardModel(), tea.WithAltScreen())
		_, err := p.Run()
		return err
	},
}

func init() {
	rootCmd.AddCommand(dashboardCmd)
}

// Styles
var (
	titleStyle    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205")).Padding(0, 1)
	selectedStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("229")).Background(lipgloss.Color("57"))
	normalStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("252"))
	activeStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("46"))
	inactiveStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	helpStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	headerStyle   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("75")).Padding(0, 1)
)

type workspaceEntry struct {
	ws     config.Workspace
	state  state.WorkspaceState
}

type dashboardModel struct {
	entries  []workspaceEntry
	cursor   int
	width    int
	height   int
	message  string
	quitting bool
}

type refreshMsg struct{}

func newDashboardModel() dashboardModel {
	m := dashboardModel{}
	m.refresh()
	return m
}

func (m *dashboardModel) refresh() {
	workspaces, _ := config.ListWorkspaces()
	m.entries = make([]workspaceEntry, len(workspaces))
	for i, ws := range workspaces {
		st, _ := state.Load(ws.Name)
		m.entries[i] = workspaceEntry{ws: ws, state: st}
	}
	if m.cursor >= len(m.entries) && len(m.entries) > 0 {
		m.cursor = len(m.entries) - 1
	}
}

func (m dashboardModel) Init() tea.Cmd {
	return nil
}

func (m dashboardModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, key.NewBinding(key.WithKeys("q", "ctrl+c"))):
			m.quitting = true
			return m, tea.Quit
		case key.Matches(msg, key.NewBinding(key.WithKeys("j", "down"))):
			if m.cursor < len(m.entries)-1 {
				m.cursor++
			}
		case key.Matches(msg, key.NewBinding(key.WithKeys("k", "up"))):
			if m.cursor > 0 {
				m.cursor--
			}
		case key.Matches(msg, key.NewBinding(key.WithKeys("r"))):
			m.refresh()
			m.message = "Refreshed"
		case key.Matches(msg, key.NewBinding(key.WithKeys("enter", "o"))):
			if len(m.entries) > 0 {
				e := m.entries[m.cursor]
				if !e.state.Active {
					m.message = fmt.Sprintf("Use 'ws open %s' to open workspace", e.ws.Name)
				} else {
					m.message = fmt.Sprintf("Workspace %s is already active", e.ws.Name)
				}
			}
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case refreshMsg:
		m.refresh()
	}
	return m, nil
}

func (m dashboardModel) View() string {
	if m.quitting {
		return ""
	}

	var b strings.Builder

	b.WriteString(titleStyle.Render("ws-manager dashboard"))
	b.WriteString("\n\n")

	if len(m.entries) == 0 {
		b.WriteString(normalStyle.Render("  No workspaces configured. Use 'ws new' to create one."))
		b.WriteString("\n")
	} else {
		// Header
		b.WriteString(headerStyle.Render(fmt.Sprintf("  %-15s %-30s %-20s %-12s %s", "NAME", "DIR", "BRANCH", "TASK", "STATUS")))
		b.WriteString("\n")

		for i, e := range m.entries {
			status := inactiveStyle.Render("inactive")
			if e.state.Active {
				status = activeStyle.Render("active")
			}
			task := e.ws.CurrentTask
			if task == "" {
				task = "-"
			}

			// Truncate dir for display
			dir := e.ws.Dir
			if len(dir) > 28 {
				dir = "..." + dir[len(dir)-25:]
			}

			line := fmt.Sprintf("  %-15s %-30s %-20s %-12s %s", e.ws.Name, dir, e.ws.CurrentBranch, task, status)
			if i == m.cursor {
				b.WriteString(selectedStyle.Render(line))
			} else {
				b.WriteString(normalStyle.Render(line))
			}
			b.WriteString("\n")
		}
	}

	b.WriteString("\n")
	if m.message != "" {
		b.WriteString("  " + m.message + "\n")
	}
	b.WriteString("\n")
	b.WriteString(helpStyle.Render("  j/k: navigate  o: open  r: refresh  q: quit"))
	b.WriteString("\n")

	return b.String()
}

// RunDashboardIfNoArgs returns true if dashboard should be shown (no args)
func RunDashboardIfNoArgs() bool {
	if len(os.Args) <= 1 {
		return true
	}
	return false
}
