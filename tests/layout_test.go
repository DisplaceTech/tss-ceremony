package tests

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/DisplaceTech/tss-ceremony/tui"
	"github.com/DisplaceTech/tss-ceremony/tui/scenes"
)

// TestLayoutVisualRegression_TitleScreen tests that the title screen renders correctly
func TestLayoutVisualRegression_TitleScreen(t *testing.T) {
	config := &scenes.Config{
		FixedMode: true,
		Message:   "Test Message",
		Speed:     "normal",
		NoColor:   true,
	}

	styles := scenes.NewStyles(config.NoColor)
	scene := scenes.NewTitleScene(config, styles)
	scene.Init()

	rendered := scene.Render()

	expectedElements := []string{
		"DKLS23 Threshold ECDSA Signing Ceremony",
		"FIXED MODE",
		"Message: Test Message",
		"Press Enter to begin the ceremony...",
	}

	for _, element := range expectedElements {
		if !strings.Contains(rendered, element) {
			t.Errorf("Expected rendered output to contain %q, but it was missing", element)
		}
	}

	lines := strings.Split(rendered, "\n")
	if len(lines) < 10 {
		t.Errorf("Expected at least 10 lines of ASCII art, got %d", len(lines))
	}
}

// TestLayoutVisualRegression_HeaderFooter tests that header and footer are rendered correctly
func TestLayoutVisualRegression_HeaderFooter(t *testing.T) {
	config := &scenes.Config{
		FixedMode: true,
		Message:   "Test",
		Speed:     "normal",
		NoColor:   true,
	}

	model := tui.NewModel(config)
	view := model.View()

	if !strings.Contains(view, "Scene") || !strings.Contains(view, "/") {
		t.Error("Expected header to contain scene number in format 'Scene X/Y'")
	}

	if !strings.Contains(view, "[") || !strings.Contains(view, "]") {
		t.Error("Expected footer to contain bracketed navigation info")
	}

	if !strings.Contains(view, "Key bindings") {
		t.Error("Expected footer to contain key bindings information")
	}
}

// TestLayoutVisualRegression_BorderRendering tests that border characters appear in the output
func TestLayoutVisualRegression_BorderRendering(t *testing.T) {
	config := &scenes.Config{
		FixedMode: true,
		Message:   "Test",
		Speed:     "normal",
		NoColor:   true,
	}

	model := tui.NewModel(config)
	view := model.View()

	// The narrator panel uses rounded borders; check for any box-drawing characters
	hasBorderChar := strings.Contains(view, "│") ||
		strings.Contains(view, "┌") ||
		strings.Contains(view, "╭") ||
		strings.Contains(view, "─")

	if !hasBorderChar {
		t.Error("Expected at least one border/box-drawing character in output")
	}
}

// TestLayoutVisualRegression_NarratorPanel tests that the narrator panel is rendered
func TestLayoutVisualRegression_NarratorPanel(t *testing.T) {
	config := &scenes.Config{
		FixedMode: true,
		Message:   "Test",
		Speed:     "normal",
		NoColor:   true,
	}

	model := tui.NewModel(config)
	view := model.View()

	if !strings.Contains(view, "Narrator:") {
		t.Error("Expected narrator panel to contain 'Narrator:' prefix")
	}

	if strings.Contains(view, "Narrator: ") && strings.TrimSpace(view[strings.Index(view, "Narrator:")+11:]) == "" {
		t.Error("Expected narrator text to be non-empty")
	}
}

// TestLayoutVisualRegression_ThreeColumnStructure tests that the three-column layout is present
// in scenes that implement it (e.g. keygen scenes), not just the title screen
func TestLayoutVisualRegression_ThreeColumnStructure(t *testing.T) {
	config := &scenes.Config{
		FixedMode: true,
		Message:   "Test",
		Speed:     "normal",
		NoColor:   true,
	}

	styles := scenes.NewStyles(config.NoColor)
	// Use a scene known to have column structure (PublicShareScene)
	scene := scenes.NewPublicShareScene(config, styles)
	view := scene.Render()

	verticalBars := strings.Count(view, "│")
	if verticalBars < 4 {
		t.Errorf("Expected at least 4 vertical bars for three-column layout, got %d", verticalBars)
	}
}

// TestLayoutVisualRegression_MinimumSize tests that the layout respects minimum size requirements
func TestLayoutVisualRegression_MinimumSize(t *testing.T) {
	spec := tui.DefaultLayoutSpec()

	if spec.MinWidth < 80 || spec.MinHeight < 24 {
		t.Errorf("Expected minimum width >= 80 and height >= 24, got %dx%d", spec.MinWidth, spec.MinHeight)
	}

	result := tui.ValidateTerminalSize(spec.MinWidth, spec.MinHeight, spec)
	if !result.IsValid {
		t.Errorf("Expected minimum size %dx%d to be valid, got mismatches: %v",
			spec.MinWidth, spec.MinHeight, result.Mismatches)
	}

	invalidResult := tui.ValidateTerminalSize(spec.MinWidth-1, spec.MinHeight, spec)
	if invalidResult.IsValid {
		t.Error("Expected width below minimum to be invalid")
	}

	invalidResult = tui.ValidateTerminalSize(spec.MinWidth, spec.MinHeight-1, spec)
	if invalidResult.IsValid {
		t.Error("Expected height below minimum to be invalid")
	}
}

// TestLayoutVisualRegression_LayoutStructure tests that the layout structure is valid
func TestLayoutVisualRegression_LayoutStructure(t *testing.T) {
	config := &scenes.Config{
		FixedMode: true,
		Message:   "Test",
		Speed:     "normal",
		NoColor:   true,
	}

	model := tui.NewModel(config)
	view := model.View()

	spec := tui.DefaultLayoutSpec()
	result := tui.ValidateLayoutStructure(view, spec)

	if !result.IsValid {
		t.Errorf("Expected valid layout structure, got mismatches: %v", result.Mismatches)
	}
}

// TestLayoutVisualRegression_ColorStyling tests that color styling is applied correctly
func TestLayoutVisualRegression_ColorStyling(t *testing.T) {
	configWithColor := &scenes.Config{
		FixedMode: true,
		Message:   "Test",
		Speed:     "normal",
		NoColor:   false,
	}

	stylesWithColor := scenes.NewStyles(configWithColor.NoColor)
	modelWithColor := tui.NewModel(configWithColor)
	viewWithColor := modelWithColor.View()

	configNoColor := &scenes.Config{
		FixedMode: true,
		Message:   "Test",
		Speed:     "normal",
		NoColor:   true,
	}

	stylesNoColor := scenes.NewStyles(configNoColor.NoColor)
	modelNoColor := tui.NewModel(configNoColor)
	viewNoColor := modelNoColor.View()

	if viewWithColor == "" {
		t.Error("Expected view with colors to be non-empty")
	}
	if viewNoColor == "" {
		t.Error("Expected view without colors to be non-empty")
	}
	if stylesWithColor == nil {
		t.Error("Expected styles with color to be non-nil")
	}
	if stylesNoColor == nil {
		t.Error("Expected styles without color to be non-nil")
	}
}

// TestLayoutVisualRegression_LipglossIntegration tests that lipgloss styling works correctly
func TestLayoutVisualRegression_LipglossIntegration(t *testing.T) {
	style := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("6")).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("241"))

	rendered := style.Render("Test Content")

	if rendered == "" {
		t.Error("Expected lipgloss style to produce non-empty output")
	}
}

// TestLayoutVisualRegression_SceneTransitions tests that scene transitions work correctly
func TestLayoutVisualRegression_SceneTransitions(t *testing.T) {
	config := &scenes.Config{
		FixedMode: true,
		Message:   "Test",
		Speed:     "normal",
		NoColor:   true,
	}

	model := tui.NewModel(config)
	initialScene := model.GetCurrentScene()

	// Scene 0 is the initial scene; just verify model is valid
	if initialScene != 0 {
		t.Errorf("Expected initial scene to be 0, got %d", initialScene)
	}

	count := model.GetSceneCount()
	if count != 20 {
		t.Errorf("Expected 20 scenes, got %d", count)
	}
}

// TestLayoutVisualRegression_NarratorContent tests that narrator content is meaningful
func TestLayoutVisualRegression_NarratorContent(t *testing.T) {
	config := &scenes.Config{
		FixedMode: true,
		Message:   "Test",
		Speed:     "normal",
		NoColor:   true,
	}

	model := tui.NewModel(config)
	view := model.View()

	narratorIndex := strings.Index(view, "Narrator:")
	if narratorIndex == -1 {
		t.Error("Expected to find 'Narrator:' in view")
		return
	}

	narratorText := view[narratorIndex+11:]
	narratorText = strings.TrimSpace(narratorText)

	if len(narratorText) < 10 {
		t.Errorf("Expected narrator text to be at least 10 characters, got %d", len(narratorText))
	}

	if strings.Contains(narratorText, "implement me") {
		t.Error("Expected narrator text to not contain placeholder text")
	}
}

// TestLayoutVisualRegression_ResponsiveResizing tests that the layout handles different sizes
func TestLayoutVisualRegression_ResponsiveResizing(t *testing.T) {
	spec := tui.DefaultLayoutSpec()

	testSizes := []struct {
		width  int
		height int
		valid  bool
	}{
		{80, 24, true},
		{100, 30, true},
		{79, 24, false},
		{80, 23, false},
		{40, 20, false},
		{200, 100, true},
	}

	for _, test := range testSizes {
		result := tui.ValidateTerminalSize(test.width, test.height, spec)
		if test.valid && !result.IsValid {
			t.Errorf("Expected size %dx%d to be valid, got mismatches: %v",
				test.width, test.height, result.Mismatches)
		}
		if !test.valid && result.IsValid {
			t.Errorf("Expected size %dx%d to be invalid, but validation passed",
				test.width, test.height)
		}
	}
}

// TestLayoutVisualRegression_HexGrouping tests that hex values are grouped correctly
func TestLayoutVisualRegression_HexGrouping(t *testing.T) {
	spec := tui.DefaultLayoutSpec()

	if spec.HexGroupSize != 8 {
		t.Errorf("Expected hex group size to be 8, got %d", spec.HexGroupSize)
	}

	hexValue := "0123456789ABCDEF0123456789ABCDEF"
	grouped := ""
	for i := 0; i < len(hexValue); i += spec.HexGroupSize {
		end := i + spec.HexGroupSize
		if end > len(hexValue) {
			end = len(hexValue)
		}
		grouped += hexValue[i:end]
		if i+spec.HexGroupSize < len(hexValue) {
			grouped += " "
		}
	}

	expected := "01234567 89ABCDEF 01234567 89ABCDEF"
	if grouped != expected {
		t.Errorf("Expected hex grouping %q, got %q", expected, grouped)
	}
}

// TestLayoutVisualRegression_ColorScheme tests that the color scheme is correctly defined
func TestLayoutVisualRegression_ColorScheme(t *testing.T) {
	spec := tui.DefaultLayoutSpec()

	if spec.PartyAColor == "" {
		t.Error("Expected PartyAColor to be defined")
	}
	if spec.PartyBColor == "" {
		t.Error("Expected PartyBColor to be defined")
	}
	if spec.SharedColor == "" {
		t.Error("Expected SharedColor to be defined")
	}
	if spec.PhantomColor == "" {
		t.Error("Expected PhantomColor to be defined")
	}
	if spec.NarratorColor == "" {
		t.Error("Expected NarratorColor to be defined")
	}
	if spec.PartyAColor != "6" {
		t.Errorf("Expected PartyAColor to be '6' (cyan), got %q", spec.PartyAColor)
	}
	if spec.PartyBColor != "5" {
		t.Errorf("Expected PartyBColor to be '5' (magenta), got %q", spec.PartyBColor)
	}
	if spec.SharedColor != "3" {
		t.Errorf("Expected SharedColor to be '3' (yellow), got %q", spec.SharedColor)
	}
}

// TestLayoutVisualRegression_ColumnWidths tests that column widths are correctly defined
func TestLayoutVisualRegression_ColumnWidths(t *testing.T) {
	spec := tui.DefaultLayoutSpec()

	if spec.LeftColumnWidth <= 0 {
		t.Error("Expected LeftColumnWidth to be positive")
	}
	if spec.SharedColumnWidth <= 0 {
		t.Error("Expected SharedColumnWidth to be positive")
	}
	if spec.RightColumnWidth <= 0 {
		t.Error("Expected RightColumnWidth to be positive")
	}

	totalWidth := spec.LeftColumnWidth + spec.SharedColumnWidth + spec.RightColumnWidth
	if totalWidth > spec.MinWidth {
		t.Errorf("Expected total column width %d to not exceed minimum width %d",
			totalWidth, spec.MinWidth)
	}
}

// TestLayoutVisualRegression_ValidationReportFormatting tests that validation reports are formatted correctly
func TestLayoutVisualRegression_ValidationReportFormatting(t *testing.T) {
	passedResult := tui.ValidationResult{
		IsValid:    true,
		TotalLines: 10,
		TotalChars: 100,
		Mismatches: []tui.Mismatch{},
		SpecName:   "Test Spec",
	}

	passedReport := tui.FormatMismatchReport(passedResult)
	if !strings.Contains(passedReport, "PASSED") {
		t.Error("Expected passed report to contain 'PASSED'")
	}
	if !strings.Contains(passedReport, "Test Spec") {
		t.Error("Expected passed report to contain spec name")
	}

	failedResult := tui.ValidationResult{
		IsValid:    false,
		TotalLines: 10,
		TotalChars: 100,
		Mismatches: []tui.Mismatch{
			{Line: 0, Column: 5, Expected: 'A', Actual: 'B', Context: "test context"},
		},
		SpecName: "Test Spec",
	}

	failedReport := tui.FormatMismatchReport(failedResult)
	if !strings.Contains(failedReport, "FAILED") {
		t.Error("Expected failed report to contain 'FAILED'")
	}
	if !strings.Contains(failedReport, "Line 0, Column 5") {
		t.Error("Expected failed report to contain mismatch coordinates")
	}
}

// TestLayoutVisualRegression_EmptySpecHandling tests that empty specs are handled correctly
func TestLayoutVisualRegression_EmptySpecHandling(t *testing.T) {
	result := tui.ValidateLayout("Some text", "", "Empty Spec")
	if len(result.Mismatches) != 0 {
		t.Errorf("Expected empty spec to pass validation, got mismatches: %v", result.Mismatches)
	}
}

// TestLayoutVisualRegression_EmptyRenderedHandling tests that empty rendered output is handled correctly
func TestLayoutVisualRegression_EmptyRenderedHandling(t *testing.T) {
	result := tui.ValidateLayout("", "Some text", "Non-empty Spec")
	if result.IsValid {
		t.Error("Expected validation to fail for empty rendered output")
	}
}

// TestLayoutVisualRegression_MismatchDetection tests that mismatches are detected correctly
func TestLayoutVisualRegression_MismatchDetection(t *testing.T) {
	rendered := "Line 1\nLine 2\nLine 3"
	spec := "Line 1\nLine X\nLine 3"

	result := tui.ValidateLayout(rendered, spec, "Multiline Test")

	if result.IsValid {
		t.Error("Expected validation to fail for multiline mismatch")
	}
	if len(result.Mismatches) != 1 {
		t.Errorf("Expected exactly 1 mismatch, got %d", len(result.Mismatches))
	}

	mismatch := result.Mismatches[0]
	if mismatch.Line != 1 || mismatch.Column != 5 {
		t.Errorf("Expected mismatch at line 1, column 5, got line %d, column %d", mismatch.Line, mismatch.Column)
	}
}

// TestLayoutVisualRegression_ContextExtraction tests that context is extracted correctly for mismatches
func TestLayoutVisualRegression_ContextExtraction(t *testing.T) {
	rendered := "The quick brown fox jumps over the lazy dog"
	spec := "The quick brown fox jumps over the LAZY dog"

	result := tui.ValidateLayout(rendered, spec, "Context Test")

	if !result.IsValid {
		if len(result.Mismatches) == 0 {
			t.Error("Expected at least one mismatch")
			return
		}
		mismatch := result.Mismatches[0]
		if mismatch.Context == "" {
			t.Error("Expected mismatch context to be non-empty")
		}
	}
}
