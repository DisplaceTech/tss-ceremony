package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// LayoutManager handles layout calculations and resize events
type LayoutManager struct {
	spec          LayoutSpec
	currentWidth  int
	currentHeight int
	minWidth      int
	minHeight     int
}

// NewLayoutManager creates a new layout manager with default spec
func NewLayoutManager() *LayoutManager {
	return &LayoutManager{
		spec:          DefaultLayoutSpec(),
		minWidth:      80,
		minHeight:     24,
		currentWidth:  80,
		currentHeight: 24,
	}
}

// SetSpec sets the layout specification
func (lm *LayoutManager) SetSpec(s LayoutSpec) {
	lm.spec = s
}

// GetSpec returns the current layout specification
func (lm *LayoutManager) GetSpec() LayoutSpec {
	return lm.spec
}

// Resize updates the layout manager with new terminal dimensions,
// enforcing minimum size requirements.
func (lm *LayoutManager) Resize(width, height int) {
	if width < lm.minWidth {
		width = lm.minWidth
	}
	if height < lm.minHeight {
		height = lm.minHeight
	}
	lm.currentWidth = width
	lm.currentHeight = height
}

// GetDimensions returns the current terminal dimensions (enforced minimum).
func (lm *LayoutManager) GetDimensions() (width, height int) {
	return lm.currentWidth, lm.currentHeight
}

// CalculateLayout returns the calculated layout dimensions for rendering.
func (lm *LayoutManager) CalculateLayout() LayoutDimensions {
	return LayoutDimensions{
		HeaderHeight:      lm.spec.HeaderHeight,
		FooterHeight:      lm.spec.FooterHeight,
		NarratorHeight:    lm.spec.NarratorHeight,
		ContentHeight:     lm.spec.ContentHeight,
		LeftColumnWidth:   lm.spec.LeftColumnWidth,
		SharedColumnWidth: lm.spec.SharedColumnWidth,
		RightColumnWidth:  lm.spec.RightColumnWidth,
		TotalWidth:        lm.currentWidth,
		TotalHeight:       lm.currentHeight,
	}
}

// LayoutDimensions holds the calculated layout dimensions.
type LayoutDimensions struct {
	HeaderHeight      int
	FooterHeight      int
	NarratorHeight    int
	ContentHeight     int
	LeftColumnWidth   int
	SharedColumnWidth int
	RightColumnWidth  int
	TotalWidth        int
	TotalHeight       int
}

// ValidateTerminalSize checks if the terminal meets minimum size requirements.
func (lm *LayoutManager) ValidateTerminalSize(width, height int) ValidationResult {
	return ValidateTerminalSize(width, height, lm.spec)
}

// RenderLayout renders the three-column layout structure with Unicode box-drawing borders.
func (lm *LayoutManager) RenderLayout(leftContent, sharedContent, rightContent string) string {
	dims := lm.CalculateLayout()

	// Use lipgloss border style for the outer container.
	_ = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(lm.spec.BorderForeground))

	var sb strings.Builder

	// Top border
	sb.WriteString("┌" + strings.Repeat("─", dims.LeftColumnWidth) + "┬" +
		strings.Repeat("─", dims.SharedColumnWidth) + "┬" +
		strings.Repeat("─", dims.RightColumnWidth) + "┐\n")

	leftLines := strings.Split(leftContent, "\n")
	sharedLines := strings.Split(sharedContent, "\n")
	rightLines := strings.Split(rightContent, "\n")

	maxRows := layoutMax3(len(leftLines), len(sharedLines), len(rightLines))
	for row := 0; row < maxRows; row++ {
		sb.WriteString("│")
		sb.WriteString(layoutPadRight(rowAt(leftLines, row), dims.LeftColumnWidth))
		sb.WriteString("│")
		sb.WriteString(layoutPadRight(rowAt(sharedLines, row), dims.SharedColumnWidth))
		sb.WriteString("│")
		sb.WriteString(layoutPadRight(rowAt(rightLines, row), dims.RightColumnWidth))
		sb.WriteString("│\n")
	}

	// Bottom border
	sb.WriteString("└" + strings.Repeat("─", dims.LeftColumnWidth) + "┴" +
		strings.Repeat("─", dims.SharedColumnWidth) + "┴" +
		strings.Repeat("─", dims.RightColumnWidth) + "┘\n")

	return sb.String()
}

// rowAt returns the string at index i, or "" if out of range.
func rowAt(lines []string, i int) string {
	if i < len(lines) {
		return lines[i]
	}
	return ""
}

// layoutPadRight pads s to exactly width characters.
func layoutPadRight(s string, width int) string {
	if len(s) >= width {
		return s[:width]
	}
	return s + strings.Repeat(" ", width-len(s))
}

// layoutMax3 returns the maximum of three integers.
func layoutMax3(a, b, c int) int {
	if a >= b && a >= c {
		return a
	}
	if b >= c {
		return b
	}
	return c
}

// LayoutValidator validates the rendered output against the layout specification.
type LayoutValidator struct {
	spec LayoutSpec
}

// NewLayoutValidator creates a new layout validator.
func NewLayoutValidator() *LayoutValidator {
	return &LayoutValidator{spec: DefaultLayoutSpec()}
}

// SetSpec sets the layout specification for validation.
func (lv *LayoutValidator) SetSpec(s LayoutSpec) {
	lv.spec = s
}

// Validate validates the rendered output against the specification.
func (lv *LayoutValidator) Validate(rendered string) ValidationResult {
	lines := strings.Split(rendered, "\n")

	result := ValidationResult{
		IsValid:  true,
		SpecName: "Layout Validation",
	}

	if len(lines) < lv.spec.MinHeight {
		result.Mismatches = append(result.Mismatches, Mismatch{
			Line:    0,
			Column:  0,
			Context: fmt.Sprintf("Expected at least %d lines, got %d", lv.spec.MinHeight, len(lines)),
		})
		result.IsValid = false
	}

	if len(lines) > 0 && len(lines[0]) < lv.spec.MinWidth {
		result.Mismatches = append(result.Mismatches, Mismatch{
			Line:    0,
			Column:  0,
			Context: fmt.Sprintf("Header too short: %d chars", len(lines[0])),
		})
		result.IsValid = false
	}

	hasThreeColumns := false
	for _, line := range lines {
		if strings.Count(line, "│") >= 4 {
			hasThreeColumns = true
			break
		}
	}
	if !hasThreeColumns {
		result.Mismatches = append(result.Mismatches, Mismatch{
			Line:    0,
			Column:  0,
			Context: "Expected three-column layout with vertical bar separators",
		})
		result.IsValid = false
	}

	return result
}
