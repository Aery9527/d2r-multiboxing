package editor

import "github.com/charmbracelet/lipgloss"

var (
	// Title bar
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("230")).
			Background(lipgloss.Color("62")).
			Padding(0, 1)

	// Tab styles
	activeTabStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("230")).
			Background(lipgloss.Color("62")).
			Padding(0, 1)

	inactiveTabStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("252")).
				Background(lipgloss.Color("236")).
				Padding(0, 1)

	// List item styles
	selectedItemStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("230")).
				Background(lipgloss.Color("62"))

	normalItemStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("252"))

	keyStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("110")).
			Width(24)

	fileTagStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("244")).
			Italic(true)

	// Search
	searchPromptStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("205"))

	// Editor
	labelStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("110"))

	previewBorderStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("62")).
				Padding(0, 1)

	// Status bar
	statusBarStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			Background(lipgloss.Color("236")).
			Padding(0, 1)

	successStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("78"))

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("196"))

	// Help
	helpKeyStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("110"))

	helpDescStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("244"))
)
