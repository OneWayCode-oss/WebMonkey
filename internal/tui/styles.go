package tui

import "github.com/charmbracelet/lipgloss"

var (
	// Colors
	primaryColor = lipgloss.Color("#00D7FF") // Bright Cyan
	secondaryColor = lipgloss.Color("#7D56F4") // Purple
	accentColor    = lipgloss.Color("#FF00D7") // Pink
	successColor   = lipgloss.Color("#04B575") // Green
	warningColor   = lipgloss.Color("#FF9D00") // Orange
	errorColor     = lipgloss.Color("#FF4141") // Red
	grayColor      = lipgloss.Color("#626262")
	bgColor        = lipgloss.Color("#1A1A1A")

	// Styles
	MainStyle = lipgloss.NewStyle().
			Padding(1, 2)

	HeaderStyle = lipgloss.NewStyle().
			Foreground(primaryColor).
			Bold(true).
			PaddingBottom(1)

	BannerStyle = lipgloss.NewStyle().
			Foreground(secondaryColor).
			Bold(true)

	StatusStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(secondaryColor).
			Padding(0, 1).
			MarginRight(1)

	HelpStyle = lipgloss.NewStyle().
			Foreground(grayColor).
			Italic(true)

	TabStyle = lipgloss.NewStyle().
			Padding(0, 2).
			Foreground(grayColor)

	ActiveTabStyle = TabStyle.Copy().
			Foreground(primaryColor).
			Bold(true).
			Underline(true)

	TableHeadStyle = lipgloss.NewStyle().
			Foreground(primaryColor).
			Bold(true).
			BorderStyle(lipgloss.NormalBorder()).
			BorderBottom(true)

	OnlineStyle = lipgloss.NewStyle().Foreground(successColor)
	OfflineStyle = lipgloss.NewStyle().Foreground(grayColor)
	NewStyle     = lipgloss.NewStyle().Foreground(accentColor).Bold(true)

	BoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(grayColor).
			Padding(0, 1)

	TitleStyle = lipgloss.NewStyle().
			Foreground(primaryColor).
			Bold(true)
)
