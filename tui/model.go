package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/DisplaceTech/tss-ceremony/tui/scenes"
)

// Scene represents a scene in the TUI
type Scene interface {
	tea.Model
	Render() string
	Narrator() string
}

// Model represents the main TUI model
type Model struct {
	currentScene int
	scenes       []Scene
	width        int
	height       int
}

// NewModel creates a new TUI model with all scenes
func NewModel() *Model {
	return &Model{
		currentScene: 0,
		scenes: []Scene{
			// Bonus scenes (15-19)
			scenes.NewRevealScene(),              // Scene 15: The Reveal
			scenes.NewSchnorrCompareScene(),      // Scene 16: Schnorr vs ECDSA
			scenes.NewScene(),                    // Scene 17: FROST Side-by-Side
			scenes.NewFrostAnimatedScene(),       // Scene 18: Animated FROST
			scenes.NewWhyBothScene(),             // Scene 19: Why Both Exist
		},
	}
}

// Init initializes the model
func (m *Model) Init() tea.Cmd {
	if len(m.scenes) > 0 {
		return m.scenes[m.currentScene].Init()
	}
	return nil
}

// Update handles events in the model
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Handle global navigation
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "n", "j", "down":
			// Next scene
			if m.currentScene < len(m.scenes)-1 {
				m.currentScene++
				return m, m.scenes[m.currentScene].Init()
			}
		case "p", "k", "up":
			// Previous scene
			if m.currentScene > 0 {
				m.currentScene--
				return m, m.scenes[m.currentScene].Init()
			}
		case "1":
			m.currentScene = 0
			return m, m.scenes[m.currentScene].Init()
		case "2":
			m.currentScene = 1
			return m, m.scenes[m.currentScene].Init()
		case "3":
			m.currentScene = 2
			return m, m.scenes[m.currentScene].Init()
		case "4":
			m.currentScene = 3
			return m, m.scenes[m.currentScene].Init()
		case "5":
			m.currentScene = 4
			return m, m.scenes[m.currentScene].Init()
		}
	}

	// Pass message to current scene
	if len(m.scenes) > 0 {
		scene, cmd := m.scenes[m.currentScene].Update(msg)
		m.scenes[m.currentScene] = scene.(Scene)
		return m, cmd
	}

	return m, nil
}

// View renders the model view
func (m *Model) View() string {
	if len(m.scenes) == 0 {
		return "No scenes available"
	}

	scene := m.scenes[m.currentScene]

	// Define styles
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		MarginBottom(1).
		Foreground(lipgloss.Color("226")) // Yellow

	narratorStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("243")).
		Bold(true).
		MarginTop(1).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("241")).
		Padding(0, 1)

	navigationStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("243"))

	// Build the view
	var builder string

	// Header with scene info
	sceneNames := []string{
		"Scene 15: The Reveal",
		"Scene 16: Schnorr vs ECDSA",
		"Scene 17: FROST Side-by-Side",
		"Scene 18: Animated FROST",
		"Scene 19: Why Both Exist",
	}

	builder += headerStyle.Render(sceneNames[m.currentScene]) + "\n"

	// Scene content
	builder += scene.Render() + "\n"

	// Narrator panel
	builder += "\n" + narratorStyle.Render("Narrator: "+scene.Narrator()) + "\n"

	// Navigation hint
	builder += "\n" + navigationStyle.Render("[n/p or j/k or ↑/↓ to navigate] [1-5 to jump] [q to quit]")

	return builder
}

// GetCurrentScene returns the current scene index
func (m *Model) GetCurrentScene() int {
	return m.currentScene
}

// GetSceneCount returns the total number of scenes
func (m *Model) GetSceneCount() int {
	return len(m.scenes)
}
