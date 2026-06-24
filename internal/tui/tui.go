package tui

import (
	"context"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/username/webmonkey/internal/config"
	"github.com/username/webmonkey/internal/domain"
	"github.com/username/webmonkey/internal/scanner"
	"github.com/username/webmonkey/internal/service"
)

// Screen types for navigation
type screen int

const (
	screenDashboard screen = iota
	screenDevices
	screenHistory
	screenLogs
	screenHelp
)

// Msg types for async updates
type scanProgressMsg scanner.ProgressUpdate
type scanFinishedMsg struct {
	scan *domain.Scan
	err  error
}

// Model is the main Bubble Tea model for WebMonkey
type Model struct {
	manager  *service.Manager
	cfg      *config.Config
	current  screen
	width    int
	height   int

	// Scan state
	isScanning bool
	progress   scanner.ProgressUpdate
	spinner    spinner.Model

	// View state
	filter string
	cursor int
}

func NewModel(cfg *config.Config, mgr *service.Manager) Model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(primaryColor)

	return Model{
		manager: mgr,
		cfg:     cfg,
		current: screenDashboard,
		spinner: s,
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "tab":
			m.current = (m.current + 1) % 5
		case "1":
			m.current = screenDashboard
		case "2":
			m.current = screenDevices
		case "3":
			m.current = screenHistory
		case "4":
			m.current = screenLogs
		case "5":
			m.current = screenHelp
		case "s": // Start scan
			if !m.isScanning {
				m.isScanning = true
				return m, m.startScanCmd()
			}
		case "/": // Filter in devices
			if m.current == screenDevices {
				// Simple toggle for filter mode could be here
			}
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case scanProgressMsg:
		m.progress = scanner.ProgressUpdate(msg)
		return m, m.spinner.Tick

	case scanFinishedMsg:
		m.isScanning = false
		if msg.err != nil {
			// Handle error
		}
		return m, nil
	}

	if m.isScanning {
		m.spinner, cmd = m.spinner.Update(msg)
	}

	return m, cmd
}

func (m Model) View() string {
	if m.width == 0 {
		return "Initializing..."
	}

	var content string
	switch m.current {
	case screenDashboard:
		content = m.renderDashboard()
	case screenDevices:
		content = m.renderDevices()
	case screenHistory:
		content = m.renderHistory()
	case screenLogs:
		content = m.renderLogs()
	case screenHelp:
		content = m.renderHelp()
	}

	// Wrap in main style
	return MainStyle.Render(
		HeaderStyle.Render("🐵 WebMonkey - See your network. Know your network.") + "\n" +
			ActiveTabStyle.Render(m.currentTabName()) + " " +
			TabStyle.Render(m.otherTabs()) + "\n\n" +
			content + "\n\n" +
			HelpStyle.Render(" [q] Quit | [tab] Switch Screen | [s] Start Scan | [1-5] Quick Nav "),
	)
}

func (m Model) currentTabName() string {
	switch m.current {
	case screenDashboard:
		return "Dashboard"
	case screenDevices:
		return "Devices"
	case screenHistory:
		return "History"
	case screenLogs:
		return "Logs"
	case screenHelp:
		return "Help"
	}
	return ""
}

func (m Model) otherTabs() string {
	tabs := []string{"Dashboard", "Devices", "History", "Logs", "Help"}
	var result string
	for i, t := range tabs {
		if screen(i) != m.current {
			result += "  " + t + " "
		}
	}
	return result
}

func (m Model) startScanCmd() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		progressChan := make(chan scanner.ProgressUpdate)
		go func() {
			for range progressChan {
				// We can't send to tea.Program directly, we must return it as a Msg.
				// In Bubble Tea, the usual way is to have a separate Cmd that listens to a chan.
				// But for simplicity in this snippet, we'll use a custom Cmd.
			}
		}()

		scan, err := m.manager.PerformScan(ctx, progressChan)
		return scanFinishedMsg{scan: scan, err: err}
	}
}
