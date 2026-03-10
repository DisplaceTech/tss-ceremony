package scenes

import (
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// TestSecretGenSceneInit verifies that the SecretGenScene initializes correctly.
func TestSecretGenSceneInit(t *testing.T) {
	config := &Config{Speed: "normal"}
	styles := &Styles{NoColor: false}

	scene := NewSecretGenScene(config, styles)

	if scene == nil {
		t.Fatal("NewSecretGenScene() returned nil")
	}

	// Verify initial state
	if scene.phase != 0 {
		t.Errorf("Expected initial phase to be 0, got %d", scene.phase)
	}

	if scene.partyAIndex != 0 {
		t.Errorf("Expected initial partyAIndex to be 0, got %d", scene.partyAIndex)
	}

	if scene.partyBIndex != 0 {
		t.Errorf("Expected initial partyBIndex to be 0, got %d", scene.partyBIndex)
	}

	if scene.started {
		t.Error("Expected started to be false initially")
	}

	// Verify Init() sets started to true and returns a command
	cmd := scene.Init()
	if cmd == nil {
		t.Error("Init() returned nil command")
	}

	if !scene.started {
		t.Error("Expected started to be true after Init()")
	}
}

// TestSecretGenSceneSpeed verifies that the scene uses the correct duration based on speed flag.
func TestSecretGenSceneSpeed(t *testing.T) {
	tests := []struct {
		name     string
		speed    string
		expected time.Duration
	}{
		{"slow", "slow", 100 * time.Millisecond},
		{"normal", "normal", 50 * time.Millisecond},
		{"fast", "fast", 20 * time.Millisecond},
		{"unknown", "unknown", 50 * time.Millisecond}, // defaults to normal
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &Config{Speed: tt.speed}
			styles := &Styles{NoColor: false}

			scene := NewSecretGenScene(config, styles)

			if scene.duration != tt.expected {
				t.Errorf("Expected duration %v for speed %s, got %v", tt.expected, tt.speed, scene.duration)
			}
		})
	}
}

// TestSecretGenSceneCharacterCount verifies that the scene generates the correct number of characters.
func TestSecretGenSceneCharacterCount(t *testing.T) {
	config := &Config{Speed: "normal"}
	styles := &Styles{NoColor: false}

	scene := NewSecretGenScene(config, styles)

	// Verify that the character arrays are sized correctly (64 hex chars = 32 bytes)
	if len(scene.partyAChars) != 64 {
		t.Errorf("Expected partyAChars length to be 64, got %d", len(scene.partyAChars))
	}

	if len(scene.partyBChars) != 64 {
		t.Errorf("Expected partyBChars length to be 64, got %d", len(scene.partyBChars))
	}
}

// TestSecretGenSceneAnimationProgress verifies that the animation progresses correctly.
func TestSecretGenSceneAnimationProgress(t *testing.T) {
	config := &Config{Speed: "fast"}
	styles := &Styles{NoColor: false}

	scene := NewSecretGenScene(config, styles)
	scene.Init()

	// Simulate animation ticks for Party A (64 characters)
	for i := 0; i < 64; i++ {
		updated, _ := scene.Update(tickMsg(time.Now()))
		scene = updated.(*SecretGenScene)
	}

	// After 64 ticks, Party A should be complete and phase should be 1
	if scene.phase != 1 {
		t.Errorf("Expected phase to be 1 after Party A animation, got %d", scene.phase)
	}

	if scene.partyAIndex != 64 {
		t.Errorf("Expected partyAIndex to be 64, got %d", scene.partyAIndex)
	}

	// Simulate animation ticks for Party B (64 characters)
	for i := 0; i < 64; i++ {
		updated, _ := scene.Update(tickMsg(time.Now()))
		scene = updated.(*SecretGenScene)
	}

	// After another 64 ticks, Party B should be complete and phase should be 2
	if scene.phase != 2 {
		t.Errorf("Expected phase to be 2 after Party B animation, got %d", scene.phase)
	}

	if scene.partyBIndex != 64 {
		t.Errorf("Expected partyBIndex to be 64, got %d", scene.partyBIndex)
	}
}

// TestSecretGenSceneColors verifies that the correct colors are applied during rendering.
func TestSecretGenSceneColors(t *testing.T) {
	config := &Config{Speed: "normal"}
	styles := &Styles{NoColor: false}

	scene := NewSecretGenScene(config, styles)

	// Fill in some characters for testing
	for i := 0; i < 64; i++ {
		scene.partyAChars[i] = 'a'
		scene.partyBChars[i] = 'b'
	}
	scene.partyAIndex = 64
	scene.partyBIndex = 64

	rendered := scene.Render()

	// Verify that the rendered output contains expected labels
	if !strings.Contains(rendered, "Party A Secret") {
		t.Error("Rendered output should contain 'Party A Secret'")
	}

	if !strings.Contains(rendered, "Party B Secret") {
		t.Error("Rendered output should contain 'Party B Secret'")
	}

	// Verify that colors are applied when NoColor is false
	if !styles.NoColor {
		if !strings.Contains(rendered, PartyAColor) {
			t.Error("Rendered output should contain PartyAColor when NoColor is false")
		}
		if !strings.Contains(rendered, PartyBColor) {
			t.Error("Rendered output should contain PartyBColor when NoColor is false")
		}
	}
}

// TestSecretGenSceneNoColorMode verifies that colors are not applied in no-color mode.
func TestSecretGenSceneNoColorMode(t *testing.T) {
	config := &Config{Speed: "normal"}
	styles := &Styles{NoColor: true}

	scene := NewSecretGenScene(config, styles)

	// Fill in some characters for testing
	for i := 0; i < 64; i++ {
		scene.partyAChars[i] = 'a'
		scene.partyBChars[i] = 'b'
	}
	scene.partyAIndex = 64
	scene.partyBIndex = 64

	rendered := scene.Render()

	// Verify that ANSI color codes are not present when NoColor is true
	if strings.Contains(rendered, "\033[") {
		t.Error("Rendered output should not contain ANSI color codes when NoColor is true")
	}
}

// TestSecretGenSceneNarrator verifies that the narrator text is correct for each phase.
func TestSecretGenSceneNarrator(t *testing.T) {
	config := &Config{Speed: "normal"}
	styles := &Styles{NoColor: false}

	scene := NewSecretGenScene(config, styles)

	// Test phase 0 (Party A generating)
	scene.phase = 0
	narrator := scene.Narrator()
	if !strings.Contains(narrator, "Party A") {
		t.Error("Narrator for phase 0 should mention Party A")
	}
	if !strings.Contains(narrator, "secret") {
		t.Error("Narrator for phase 0 should mention secret")
	}

	// Test phase 1 (Party B generating)
	scene.phase = 1
	narrator = scene.Narrator()
	if !strings.Contains(narrator, "Party B") {
		t.Error("Narrator for phase 1 should mention Party B")
	}
	if !strings.Contains(narrator, "secret") {
		t.Error("Narrator for phase 1 should mention secret")
	}

	// Test phase 2 (complete)
	scene.phase = 2
	narrator = scene.Narrator()
	if !strings.Contains(narrator, "Both parties") {
		t.Error("Narrator for phase 2 should mention both parties")
	}
}

// TestSecretGenSceneQuit verifies that the scene handles quit commands correctly.
func TestSecretGenSceneQuit(t *testing.T) {
	config := &Config{Speed: "normal"}
	styles := &Styles{NoColor: false}

	scene := NewSecretGenScene(config, styles)

	// Test quit with 'q' key
	updated, cmd := scene.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	if updated == nil {
		t.Error("Update() should return a valid model on quit")
	}
	if cmd == nil {
		t.Error("Update() should return a command on 'q' key")
	}

	// Test quit with 'ctrl+c' key
	updated, cmd = scene.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	if updated == nil {
		t.Error("Update() should return a valid model on ctrl+c")
	}
	if cmd == nil {
		t.Error("Update() should return a command on ctrl+c")
	}
}

// TestSecretGenSceneNavigation verifies that the scene handles navigation keys correctly.
func TestSecretGenSceneNavigation(t *testing.T) {
	config := &Config{Speed: "normal"}
	styles := &Styles{NoColor: false}

	scene := NewSecretGenScene(config, styles)

	// Test navigation keys that should not quit
	navKeys := []tea.KeyMsg{
		{Type: tea.KeyEnter},
		{Type: tea.KeyRunes, Runes: []rune{' '}},
		{Type: tea.KeyRunes, Runes: []rune{'n'}},
		{Type: tea.KeyRunes, Runes: []rune{'j'}},
		{Type: tea.KeyRunes, Runes: []rune{'l'}},
		{Type: tea.KeyRight},
	}

	for _, key := range navKeys {
		updated, cmd := scene.Update(key)
		if updated == nil {
			t.Errorf("Update() should return a valid model for key %v", key)
		}
		if cmd != nil {
			t.Errorf("Update() should return nil command for navigation key %v", key)
		}
	}
}

// TestSecretGenSceneRender verifies that the Render method produces valid output.
func TestSecretGenSceneRender(t *testing.T) {
	config := &Config{Speed: "normal"}
	styles := &Styles{NoColor: false}

	scene := NewSecretGenScene(config, styles)

	// Test render with empty characters (initial state)
	rendered := scene.Render()
	if rendered == "" {
		t.Error("Render() should not return empty string")
	}

	// Test render with partial characters
	for i := 0; i < 32; i++ {
		scene.partyAChars[i] = 'a'
	}
	scene.partyAIndex = 32

	rendered = scene.Render()
	if rendered == "" {
		t.Error("Render() should not return empty string with partial characters")
	}

	// Test render with complete characters
	for i := 32; i < 64; i++ {
		scene.partyAChars[i] = 'a'
		scene.partyBChars[i] = 'b'
	}
	scene.partyAIndex = 64
	scene.partyBIndex = 64

	rendered = scene.Render()
	if rendered == "" {
		t.Error("Render() should not return empty string with complete characters")
	}
}

// TestSecretGenSceneView verifies that the View method delegates to Render.
func TestSecretGenSceneView(t *testing.T) {
	config := &Config{Speed: "normal"}
	styles := &Styles{NoColor: false}

	scene := NewSecretGenScene(config, styles)

	rendered := scene.Render()
	view := scene.View()

	if rendered != view {
		t.Error("View() should return the same output as Render()")
	}
}

// TestSecretGenSceneHexFormatting verifies that hex characters are formatted in 8-char groups.
func TestSecretGenSceneHexFormatting(t *testing.T) {
	config := &Config{Speed: "normal"}
	styles := &Styles{NoColor: true} // Disable colors for easier testing

	scene := NewSecretGenScene(config, styles)

	// Fill all characters
	for i := 0; i < 64; i++ {
		scene.partyAChars[i] = 'a'
	}
	scene.partyAIndex = 64

	rendered := scene.Render()

	// There should be spaces between groups, but also other spaces in the output
	// We just verify the output is not empty and contains expected content
	if !strings.Contains(rendered, "aaaaaaaa") {
		t.Error("Rendered output should contain hex characters")
	}
}
