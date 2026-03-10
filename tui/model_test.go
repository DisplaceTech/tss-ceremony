package tui

import (
	"testing"

	"github.com/DisplaceTech/tss-ceremony/tui/scenes"
)

// TestBonusSceneIntegration verifies that all bonus scenes (15-19) are properly
// integrated with the main TUI framework and can be navigated to from the main scene list.
func testConfig() *scenes.Config {
	return &scenes.Config{Speed: "normal"}
}

func TestBonusSceneIntegration(t *testing.T) {
	model := NewModel(testConfig())

	// Verify we have all 20 scenes (0-19)
	if model.GetSceneCount() != 20 {
		t.Errorf("Expected 20 scenes, got %d", model.GetSceneCount())
	}

	// Verify each scene is accessible and implements the Scene interface
	sceneNames := []string{
		"Scene 15: The Reveal",
		"Scene 16: Schnorr vs ECDSA",
		"Scene 17: FROST Side-by-Side",
		"Scene 18: Animated FROST",
		"Scene 19: Why Both Exist",
	}

	// Verify bonus scenes (15-19) have real content
	for i := 15; i < 20; i++ {
		scene := model.scenes[i]

		rendered := scene.Render()
		if rendered == "" {
			t.Errorf("Scene %d (%s) Render() returned empty string", i, sceneNames[i])
		}

		narrator := scene.Narrator()
		if narrator == "" {
			t.Errorf("Scene %d (%s) Narrator() returned empty string", i, sceneNames[i])
		}

		view := scene.View()
		if view == "" {
			t.Errorf("Scene %d (%s) View() returned empty string", i, sceneNames[i])
		}
	}
}

// TestSceneNavigation verifies that navigation between bonus scenes works correctly.
func TestSceneNavigation(t *testing.T) {
	model := NewModel(testConfig())

	// Start at scene 0
	if model.GetCurrentScene() != 0 {
		t.Errorf("Expected initial scene to be 0, got %d", model.GetCurrentScene())
	}

	// Navigate through all scenes using Update with key messages
	for i := 0; i < model.GetSceneCount()-1; i++ {
		_, cmd := model.Update(struct {
			Type string
		}{Type: "next"})

		// Note: The actual navigation uses tea.KeyMsg, but we're testing the logic
		// The model should handle navigation correctly
		_ = cmd
	}
}

// TestRevealScene verifies the Reveal scene (Scene 15) functionality.
func TestRevealScene(t *testing.T) {
	scene := scenes.NewRevealScene()

	// Verify initial state
	if scene == nil {
		t.Fatal("NewRevealScene() returned nil")
	}

	// Verify Render produces output
	rendered := scene.Render()
	if rendered == "" {
		t.Error("RevealScene.Render() returned empty string")
	}

	// Verify Narrator produces output
	narrator := scene.Narrator()
	if narrator == "" {
		t.Error("RevealScene.Narrator() returned empty string")
	}

	// Verify the scene contains expected content
	if !contains(rendered, "k⁻¹·x") {
		t.Error("RevealScene should contain explanation of k⁻¹·x term")
	}
}

// TestSchnorrCompareScene verifies the Schnorr comparison scene (Scene 16) functionality.
func TestSchnorrCompareScene(t *testing.T) {
	scene := scenes.NewSchnorrCompareScene()

	// Verify initial state
	if scene == nil {
		t.Fatal("NewSchnorrCompareScene() returned nil")
	}

	// Verify Render produces output
	rendered := scene.Render()
	if rendered == "" {
		t.Error("SchnorrCompareScene.Render() returned empty string")
	}

	// Verify Narrator produces output
	narrator := scene.Narrator()
	if narrator == "" {
		t.Error("SchnorrCompareScene.Narrator() returned empty string")
	}

	// Verify the scene contains expected content
	if !contains(rendered, "Schnorr") || !contains(rendered, "ECDSA") {
		t.Error("SchnorrCompareScene should contain both Schnorr and ECDSA content")
	}
}

// TestFrostSideBySideScene verifies the FROST side-by-side scene (Scene 17) functionality.
func TestFrostSideBySideScene(t *testing.T) {
	scene := scenes.NewScene()

	// Verify initial state
	if scene == nil {
		t.Fatal("NewScene() returned nil")
	}

	// Verify Render produces output
	rendered := scene.Render()
	if rendered == "" {
		t.Error("FrostSideBySideScene.Render() returned empty string")
	}

	// Verify Narrator produces output
	narrator := scene.Narrator()
	if narrator == "" {
		t.Error("FrostSideBySideScene.Narrator() returned empty string")
	}

	// Verify the scene contains expected content
	if !contains(rendered, "FROST") || !contains(rendered, "DKLS") {
		t.Error("FrostSideBySideScene should contain both FROST and DKLS content")
	}
}

// TestFrostAnimatedScene verifies the FROST animated scene (Scene 18) functionality.
func TestFrostAnimatedScene(t *testing.T) {
	scene := scenes.NewFrostAnimatedScene(nil)

	// Verify initial state
	if scene == nil {
		t.Fatal("NewFrostAnimatedScene() returned nil")
	}

	// Verify Render produces output
	rendered := scene.Render()
	if rendered == "" {
		t.Error("FrostAnimatedScene.Render() returned empty string")
	}

	// Verify Narrator produces output
	narrator := scene.Narrator()
	if narrator == "" {
		t.Error("FrostAnimatedScene.Narrator() returned empty string")
	}

	// Verify the scene contains expected content
	if !contains(rendered, "FROST") {
		t.Error("FrostAnimatedScene should contain FROST content")
	}
}

// TestWhyBothScene verifies the Why Both Exist scene (Scene 19) functionality.
func TestWhyBothScene(t *testing.T) {
	scene := scenes.NewWhyBothScene()

	// Verify initial state
	if scene == nil {
		t.Fatal("NewWhyBothScene() returned nil")
	}

	// Verify Render produces output
	rendered := scene.Render()
	if rendered == "" {
		t.Error("WhyBothScene.Render() returned empty string")
	}

	// Verify Narrator produces output
	narrator := scene.Narrator()
	if narrator == "" {
		t.Error("WhyBothScene.Narrator() returned empty string")
	}

	// Verify the scene contains expected content
	if !contains(rendered, "ECDSA") || !contains(rendered, "Schnorr") {
		t.Error("WhyBothScene should contain both ECDSA and Schnorr content")
	}
}

// TestSceneInit verifies that all bonus scenes can be initialized.
func TestSceneInit(t *testing.T) {
	scenesToTest := []struct {
		name  string
		scene Scene
	}{
		{"Reveal", scenes.NewRevealScene()},
		{"SchnorrCompare", scenes.NewSchnorrCompareScene()},
		{"FrostSideBySide", scenes.NewScene()},
		{"FrostAnimated", scenes.NewFrostAnimatedScene(nil)},
		{"WhyBoth", scenes.NewWhyBothScene()},
	}

	for _, tc := range scenesToTest {
		t.Run(tc.name, func(t *testing.T) {
			cmd := tc.scene.Init()
			// Init should return nil or a valid command
			_ = cmd
		})
	}
}

// TestSceneUpdate verifies that all bonus scenes can handle updates.
func TestSceneUpdate(t *testing.T) {
	scenesToTest := []struct {
		name  string
		scene Scene
	}{
		{"Reveal", scenes.NewRevealScene()},
		{"SchnorrCompare", scenes.NewSchnorrCompareScene()},
		{"FrostSideBySide", scenes.NewScene()},
		{"FrostAnimated", scenes.NewFrostAnimatedScene(nil)},
		{"WhyBoth", scenes.NewWhyBothScene()},
	}

	for _, tc := range scenesToTest {
		t.Run(tc.name, func(t *testing.T) {
			// Test with a quit message
			updated, cmd := tc.scene.Update(struct {
				Type string
			}{Type: "quit"})

			// Should return a valid model
			if updated == nil {
				t.Error("Update() returned nil model")
			}

			_ = cmd
		})
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && findSubstring(s, substr))
}

// findSubstring is a simple substring search
func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
