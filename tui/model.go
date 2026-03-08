// Package tui provides the terminal user interface for the TSS ceremony.
// It uses bubbletea to create interactive scenes that animate the DKLS23
// 2-of-2 threshold ECDSA signature ceremony.
package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Scene represents a single view in the TUI application.
type Scene interface {
	// Title returns the title of the scene.
	Title() string
	// Description returns a brief description of the scene.
	Description() string
}

// Style holds the styling configuration for TUI elements.
type Style struct {
	Title       lipgloss.Style
	Description lipgloss.Style
	Success     lipgloss.Style
	Error       lipgloss.Style
	Info        lipgloss.Style
	Warning     lipgloss.Style
	Code        lipgloss.Style
}

// Narrator represents the entity explaining the ceremony steps.
type Narrator struct {
	Name        string
	Avatar      string
	Message     string
	IsSpeaking  bool
}

// Model is the main TUI model that implements bubbletea.Model.
type Model struct {
	CurrentScene Scene
	Styles       Style
	Narrator     Narrator
	Quitting     bool
}

// Init implements bubbletea.Model.
func (m Model) Init() tea.Cmd {
	return nil
}

// Update implements bubbletea.Model.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			m.Quitting = true
			return m, tea.Quit
		}
	}
	return m, nil
}

// View implements bubbletea.Model.
func (m Model) View() string {
	if m.Quitting {
		return "Goodbye!\n"
	}

	var sb string
	sb += m.Styles.Title.Render(m.CurrentScene.Title()) + "\n"
	sb += m.Styles.Description.Render(m.CurrentScene.Description()) + "\n"
	sb += "\n"
	sb += m.Styles.Info.Render("Press q or ctrl+c to quit")

	return sb
}
