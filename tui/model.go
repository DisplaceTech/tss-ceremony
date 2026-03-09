package tui

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// Model represents the main TUI model that manages scenes
type Model struct {
	config       *Config
	currentScene int
	scenes       []Scene
	quit         bool
	speedDelay   time.Duration
	styles       *Styles
}

// Config holds the TUI configuration
type Config struct {
	FixedMode bool
	Message   string
	Speed     string
	NoColor   bool
}

// Scene represents a single scene in the ceremony
type Scene interface {
	Update(msg tea.Msg) tea.Cmd
	View() string
}

// NewModel creates a new TUI model with all scenes wired up
func NewModel(config *Config) Model {
	m := Model{
		config:     config,
		quit:       false,
		speedDelay: getSpeedDelay(config.Speed),
		styles:     NewStyles(config.NoColor),
	}

	// Wire up all scenes
	m.scenes = m.createScenes()

	return m
}

// getSpeedDelay returns the delay duration based on speed setting
func getSpeedDelay(speed string) time.Duration {
	switch speed {
	case "slow":
		return 200 * time.Millisecond
	case "fast":
		return 50 * time.Millisecond
	default:
		return 100 * time.Millisecond
	}
}

// createScenes creates all ceremony scenes
func (m Model) createScenes() []Scene {
	// For now, create placeholder scenes
	// In a full implementation, each scene would be a separate file
	scenes := make([]Scene, 20) // Scenes 0-19

	for i := range scenes {
		scenes[i] = &PlaceholderScene{SceneNum: i, Config: m.config, Styles: m.styles}
	}

	return scenes
}

// Init initializes the model
func (m Model) Init() tea.Cmd {
	// Return a tick command based on speed setting for animation timing
	return tea.Tick(m.speedDelay, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

// tickMsg is a message type for animation ticks
type tickMsg time.Time

// Update handles messages and updates the model
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c", "esc":
			m.quit = true
			return m, tea.Quit
		case "right", "l", " ":
			// Advance to next scene
			if m.currentScene < len(m.scenes)-1 {
				m.currentScene++
			}
		case "left", "h":
			// Go to previous scene
			if m.currentScene > 0 {
				m.currentScene--
			}
		}

	case tea.WindowSizeMsg:
		// Handle window resize if needed

	case tickMsg:
		// Handle animation tick - schedule next tick
		return m, tea.Tick(m.speedDelay, func(t time.Time) tea.Msg {
			return tickMsg(t)
		})
	}

	// Update current scene
	if m.currentScene < len(m.scenes) {
		cmd := m.scenes[m.currentScene].Update(msg)
		return m, cmd
	}

	return m, nil
}

// View renders the current scene
func (m Model) View() string {
	if m.quit {
		return "Goodbye!\n"
	}

	var view string

	// Show fixed mode banner if applicable
	if m.config.FixedMode {
		view += fixedModeBanner()
		view += "\n"
	}

	// Show current scene
	if m.currentScene < len(m.scenes) {
		view += fmt.Sprintf("Scene %d/%d\n", m.currentScene, len(m.scenes)-1)
		view += m.scenes[m.currentScene].View()
	}

	view += "\nPress q to quit, ←/→ to navigate scenes\n"

	return view
}

// fixedModeBanner returns a banner for fixed mode
func fixedModeBanner() string {
	return "=== FIXED MODE ==="
}

// PlaceholderScene is a placeholder for actual scene implementations
type PlaceholderScene struct {
	SceneNum int
	Config   *Config
	Styles   *Styles
}

// Update handles messages for the placeholder scene
func (s *PlaceholderScene) Update(msg tea.Msg) tea.Cmd {
	return nil
}

// View renders the placeholder scene
func (s *PlaceholderScene) View() string {
	if s.Config.NoColor {
		return fmt.Sprintf("Scene %d placeholder - implement me!", s.SceneNum)
	}
	return fmt.Sprintf("%sScene %d placeholder - implement me!%s", "\033[36m", s.SceneNum, "\033[0m")
}
