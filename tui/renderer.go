package tui

import (
	"regexp"
	"strings"
)

// ansiEscapeRegexp matches ANSI escape sequences (colors, bold, dim, etc.)
var ansiEscapeRegexp = regexp.MustCompile(`\x1b\[[0-9;]*[a-zA-Z]`)

// StripANSI removes all ANSI escape sequences from a string.
// Used when --no-color is active to ensure clean plain-text output.
func StripANSI(s string) string {
	return ansiEscapeRegexp.ReplaceAllString(s, "")
}

// Renderer handles rendering of TUI output with optional color stripping.
type Renderer struct {
	noColor bool
}

// NewRenderer creates a new Renderer.
// If noColor is true, all output is stripped of ANSI escape sequences.
func NewRenderer(noColor bool) *Renderer {
	return &Renderer{noColor: noColor}
}

// Render processes a string, stripping ANSI codes if no-color mode is active.
func (r *Renderer) Render(s string) string {
	if r.noColor {
		return StripANSI(s)
	}
	return s
}

// RenderLines renders each line of a multi-line string through the renderer.
func (r *Renderer) RenderLines(s string) string {
	if !r.noColor {
		return s
	}
	lines := strings.Split(s, "\n")
	for i, line := range lines {
		lines[i] = StripANSI(line)
	}
	return strings.Join(lines, "\n")
}
