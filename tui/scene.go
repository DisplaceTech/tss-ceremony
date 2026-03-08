// Package tui provides the terminal user interface for the TSS ceremony.
package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// SceneType represents the type of scene in the TUI.
type SceneType int

const (
	SceneWelcome SceneType = iota
	SceneSetup
	SceneKeyGeneration
	SceneSigning
	SceneCompletion
)

// SceneManager handles scene navigation and state.
type SceneManager struct {
	Scenes       []Scene
	CurrentIndex int
}

// NewSceneManager creates a new SceneManager with the given scenes.
func NewSceneManager(scenes ...Scene) *SceneManager {
	return &SceneManager{
		Scenes:       scenes,
		CurrentIndex: 0,
	}
}

// Current returns the current scene.
func (sm *SceneManager) Current() Scene {
	if sm.CurrentIndex < 0 || sm.CurrentIndex >= len(sm.Scenes) {
		return nil
	}
	return sm.Scenes[sm.CurrentIndex]
}

// Next advances to the next scene.
func (sm *SceneManager) Next() bool {
	if sm.CurrentIndex < len(sm.Scenes)-1 {
		sm.CurrentIndex++
		return true
	}
	return false
}

// Previous goes back to the previous scene.
func (sm *SceneManager) Previous() bool {
	if sm.CurrentIndex > 0 {
		sm.CurrentIndex--
		return true
	}
	return false
}

// JumpTo sets the current scene by index.
func (sm *SceneManager) JumpTo(index int) bool {
	if index >= 0 && index < len(sm.Scenes) {
		sm.CurrentIndex = index
		return true
	}
	return false
}

// HasNext returns true if there is a next scene.
func (sm *SceneManager) HasNext() bool {
	return sm.CurrentIndex < len(sm.Scenes)-1
}

// HasPrevious returns true if there is a previous scene.
func (sm *SceneManager) HasPrevious() bool {
	return sm.CurrentIndex > 0
}

// SceneCount returns the total number of scenes.
func (sm *SceneManager) SceneCount() int {
	return len(sm.Scenes)
}

// WelcomeScene is the initial welcome scene.
type WelcomeScene struct{}

// Title returns the title of the welcome scene.
func (s WelcomeScene) Title() string {
	return "Welcome to TSS Ceremony"
}

// Description returns the description of the welcome scene.
func (s WelcomeScene) Description() string {
	return "This interactive guide will walk you through the DKLS23 2-of-2 threshold ECDSA signature ceremony."
}

// SetupScene represents the setup phase.
type SetupScene struct{}

// Title returns the title of the setup scene.
func (s SetupScene) Title() string {
	return "Setup Phase"
}

// Description returns the description of the setup scene.
func (s SetupScene) Description() string {
	return "Initialize the ceremony parameters and prepare for key generation."
}

// KeyGenerationScene represents the key generation phase.
type KeyGenerationScene struct{}

// Title returns the title of the key generation scene.
func (s KeyGenerationScene) Title() string {
	return "Key Generation Phase"
}

// Description returns the description of the key generation scene.
func (s KeyGenerationScene) Description() string {
	return "Generate the threshold key shares for both participants."
}

// SigningScene represents the signing phase.
type SigningScene struct{}

// Title returns the title of the signing scene.
func (s SigningScene) Title() string {
	return "Signing Phase"
}

// Description returns the description of the signing scene.
func (s SigningScene) Description() string {
	return "Perform the threshold signature operation using the generated key shares."
}

// CompletionScene represents the completion phase.
type CompletionScene struct{}

// Title returns the title of the completion scene.
func (s CompletionScene) Title() string {
	return "Ceremony Complete"
}

// Description returns the description of the completion scene.
func (s CompletionScene) Description() string {
	return "The TSS ceremony has been successfully completed."
}

// SceneCmd is a command for scene transitions.
type SceneCmd struct {
	Action string
	Index  int
}

// SceneTransitionCmd creates a command to transition to a scene.
func SceneTransitionCmd(index int) tea.Cmd {
	return func() tea.Msg {
		return SceneCmd{Action: "transition", Index: index}
	}
}

// SceneStyles defines styles for scene elements.
type SceneStyles struct {
	Header    lipgloss.Style
	Body      lipgloss.Style
	Footer    lipgloss.Style
	Progress  lipgloss.Style
	Highlight lipgloss.Style
}

// NewSceneStyles creates a new SceneStyles with default styling.
func NewSceneStyles() SceneStyles {
	return SceneStyles{
		Header: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("202")).
			MarginBottom(1),
		Body: lipgloss.NewStyle().
			MarginLeft(2).
			MarginRight(2),
		Footer: lipgloss.NewStyle().
			Italic(true).
			Foreground(lipgloss.Color("241")).
			MarginTop(1),
		Progress: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("45")),
		Highlight: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("212")),
	}
}

// RenderScene renders a scene with the given styles.
func RenderScene(scene Scene, styles SceneStyles, progress int, total int) string {
	var sb strings.Builder

	// Header
	sb.WriteString(styles.Header.Render(scene.Title()))
	sb.WriteString("\n")

	// Progress indicator
	if total > 0 {
		progressStr := fmt.Sprintf("Progress: %d/%d", progress, total)
		sb.WriteString(styles.Progress.Render(progressStr))
		sb.WriteString("\n")
	}

	// Body
	sb.WriteString(styles.Body.Render(scene.Description()))
	sb.WriteString("\n")

	return sb.String()
}

// CreateDefaultScenes returns a slice of default scenes for the ceremony.
func CreateDefaultScenes() []Scene {
	return []Scene{
		WelcomeScene{},
		SetupScene{},
		KeyGenerationScene{},
		SigningScene{},
		CompletionScene{},
	}
}
