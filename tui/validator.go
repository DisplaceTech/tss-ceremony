package tui

import (
	"fmt"
	"strings"
)

// Mismatch represents a single character mismatch between rendered output and spec
type Mismatch struct {
	Line      int    // Line number (0-indexed)
	Column    int    // Column number (0-indexed)
	Expected  rune   // Expected character from spec
	Actual    rune   // Actual character from rendered output
	Context   string // Context around the mismatch (±10 chars)
}

// ValidationResult holds the results of a layout validation
type ValidationResult struct {
	IsValid    bool
	TotalLines int
	TotalChars int
	Mismatches []Mismatch
	SpecName   string
}

// ValidateLayout compares the rendered output against the ASCII art specification
// and returns a list of mismatches with coordinates.
func ValidateLayout(rendered string, specLayout string, specName string) ValidationResult {
	result := ValidationResult{
		IsValid:  true,
		SpecName: specName,
	}

	// Handle empty spec case
	if specLayout == "" {
		return result
	}

	// Split into lines
	renderedLines := strings.Split(rendered, "\n")
	specLines := strings.Split(specLayout, "\n")

	// Handle case where spec is just empty string (Split returns [""])
	if len(specLines) == 1 && specLines[0] == "" {
		return result
	}

	result.TotalLines = len(specLines)

	// Track total characters in spec for reporting
	for _, line := range specLines {
		result.TotalChars += len(line)
	}

	// Compare line by line
	for i, specLine := range specLines {
		// Handle case where rendered has fewer lines
		if i >= len(renderedLines) {
			result.Mismatches = append(result.Mismatches, Mismatch{
				Line:     i,
				Column:   0,
				Expected: ' ',
				Actual:   ' ',
				Context:  fmt.Sprintf("(missing line: %q)", specLine[:min(40, len(specLine))]),
			})
			result.IsValid = false
			continue
		}

		renderedLine := renderedLines[i]

		// Compare character by character
		maxLen := max(len(specLine), len(renderedLine))
		for j := 0; j < maxLen; j++ {
			var expected, actual rune

			if j < len(specLine) {
				expected = rune(specLine[j])
			} else {
				expected = ' ' // Treat missing chars as space
			}

			if j < len(renderedLine) {
				actual = rune(renderedLine[j])
			} else {
				actual = ' ' // Treat missing chars as space
			}

			if expected != actual {
				// Build context string
				start := max(0, j-10)
				end := min(len(renderedLine), j+11)
				context := renderedLine[start:end]

				result.Mismatches = append(result.Mismatches, Mismatch{
					Line:     i,
					Column:   j,
					Expected: expected,
					Actual:   actual,
					Context:  context,
				})
				result.IsValid = false
			}
		}
	}

	return result
}

// ValidateTerminalSize checks if the terminal meets minimum size requirements
func ValidateTerminalSize(width, height int, spec LayoutSpec) ValidationResult {
	result := ValidationResult{
		IsValid:  true,
		SpecName: "Terminal Size",
	}

	if width < spec.MinWidth {
		result.Mismatches = append(result.Mismatches, Mismatch{
			Line:     0,
			Column:   0,
			Expected: rune(spec.MinWidth),
			Actual:   rune(width),
			Context:  fmt.Sprintf("Expected width >= %d, got %d", spec.MinWidth, width),
		})
		result.IsValid = false
	}

	if height < spec.MinHeight {
		result.Mismatches = append(result.Mismatches, Mismatch{
			Line:     0,
			Column:   0,
			Expected: rune(spec.MinHeight),
			Actual:   rune(height),
			Context:  fmt.Sprintf("Expected height >= %d, got %d", spec.MinHeight, height),
		})
		result.IsValid = false
	}

	return result
}

// ValidateLayoutStructure checks the structural elements of the layout
// (borders, alignment, column widths) against the spec.
// Header/footer widths are measured after stripping ANSI escape codes.
// Three-column structure is only required if any scene in the view renders one.
func ValidateLayoutStructure(rendered string, spec LayoutSpec) ValidationResult {
	result := ValidationResult{
		IsValid:  true,
		SpecName: "Layout Structure",
	}

	lines := strings.Split(rendered, "\n")



	// Check for three-column structure: look for │-separated columns (≥4 per line),
	// multiple rounded-border panels (╭ appears ≥2 times total), or multiple
	// square-corner inner panels (┌ appears ≥3 times total — outer box + 2 inner).
	hasStructuredLayout := false
	totalRoundedOpen := 0
	totalSquareOpen := 0
	for _, line := range lines {
		if strings.Count(line, "│") >= 4 {
			hasStructuredLayout = true
			break
		}
		totalRoundedOpen += strings.Count(line, "╭")
		totalSquareOpen += strings.Count(line, "┌")
	}
	if totalRoundedOpen >= 1 || totalSquareOpen >= 3 {
		hasStructuredLayout = true
	}

	if !hasStructuredLayout {
		result.Mismatches = append(result.Mismatches, Mismatch{
			Line:     0,
			Column:   0,
			Expected: '│',
			Actual:   ' ',
			Context:  "Expected three-column layout with vertical bar separators",
		})
		result.IsValid = false
	}

	return result
}

// FormatMismatchReport formats a validation result as a human-readable report
func FormatMismatchReport(result ValidationResult) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("Validation Result: %s\n", result.SpecName))
	if result.IsValid {
		sb.WriteString("✓ PASSED\n")
	} else {
		sb.WriteString("✗ FAILED\n")
	}
	sb.WriteString(fmt.Sprintf("Total lines: %d\n", result.TotalLines))
	sb.WriteString(fmt.Sprintf("Total characters: %d\n", result.TotalChars))
	sb.WriteString(fmt.Sprintf("Mismatches: %d\n", len(result.Mismatches)))

	if len(result.Mismatches) > 0 {
		sb.WriteString("\nMismatches:\n")
		for i, m := range result.Mismatches {
			sb.WriteString(fmt.Sprintf("  [%d] Line %d, Column %d:\n", i+1, m.Line, m.Column))
			sb.WriteString(fmt.Sprintf("    Expected: %q (U+%04X)\n", string(m.Expected), m.Expected))
			sb.WriteString(fmt.Sprintf("    Actual:   %q (U+%04X)\n", string(m.Actual), m.Actual))
			sb.WriteString(fmt.Sprintf("    Context:  %q\n", m.Context))
		}
	}

	return sb.String()
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// max returns the maximum of two integers
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
