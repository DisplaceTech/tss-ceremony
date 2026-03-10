package tui

import (
	"fmt"
	"strings"
)

// Styles holds styling information for the TUI
type Styles struct {
	NoColor bool
}

// NewStyles creates a new Styles instance
func NewStyles(noColor bool) *Styles {
	return &Styles{
		NoColor: noColor,
	}
}

// ANSI color codes
const (
	Reset     = "\033[0m"
	Bold      = "\033[1m"
	Dim       = "\033[2m"
	Underline = "\033[4m"

	// Foreground colors
	FGBlack   = "\033[30m"
	FGRed     = "\033[31m"
	FGGreen   = "\033[32m"
	FGYellow  = "\033[33m"
	FGBlue    = "\033[34m"
	FGMagenta = "\033[35m"
	FGCyan    = "\033[36m"
	FGWhite   = "\033[37m"

	// Bright foreground colors
	FGBrightRed    = "\033[91m"
	FGBrightGreen  = "\033[92m"
	FGBrightYellow = "\033[93m"
	FGBrightBlue   = "\033[94m"
	FGBrightMagenta = "\033[95m"
	FGBrightCyan   = "\033[96m"
	FGBrightWhite  = "\033[97m"

	// Background colors
	BGBlack   = "\033[40m"
	BGRed     = "\033[41m"
	BGGreen   = "\033[42m"
	BGYellow  = "\033[43m"
	BGBlue    = "\033[44m"
	BGMagenta = "\033[45m"
	BGCyan    = "\033[46m"
	BGWhite   = "\033[47m"
)

// Party colors
var (
	PartyAColor   = FGCyan
	PartyBColor   = FGMagenta
	SharedColor   = FGYellow
	PhantomColor  = FGRed
	NarratorColor = Dim
)

// ApplyColor applies a color code to text, respecting NoColor setting
func (s *Styles) ApplyColor(text string, color string) string {
	if s.NoColor {
		return text
	}
	return color + text + Reset
}

// PartyA applies Party A's color (cyan)
func (s *Styles) PartyA(text string) string {
	return s.ApplyColor(text, PartyAColor)
}

// PartyB applies Party B's color (magenta)
func (s *Styles) PartyB(text string) string {
	return s.ApplyColor(text, PartyBColor)
}

// Shared applies shared color (yellow)
func (s *Styles) Shared(text string) string {
	return s.ApplyColor(text, SharedColor)
}

// Phantom applies phantom key color (red)
func (s *Styles) Phantom(text string) string {
	return s.ApplyColor(text, PhantomColor)
}

// Narrator applies narrator text color (dim)
func (s *Styles) Narrator(text string) string {
	return s.ApplyColor(text, NarratorColor)
}

// Bold applies bold formatting
func (s *Styles) Bold(text string) string {
	if s.NoColor {
		return text
	}
	return Bold + text + Reset
}

// Dim applies dim formatting
func (s *Styles) Dim(text string) string {
	if s.NoColor {
		return text
	}
	return Dim + text + Reset
}

// Underline applies underline formatting
func (s *Styles) Underline(text string) string {
	if s.NoColor {
		return text
	}
	return Underline + text + Reset
}

// Box draws a simple box around text
func (s *Styles) Box(text string, width int) string {
	if width <= 0 {
		width = 40
	}

	lines := strings.Split(text, "\n")
	maxLen := 0
	for _, line := range lines {
		if len(line) > maxLen {
			maxLen = len(line)
		}
	}

	if maxLen > width {
		maxLen = width
	}

	border := strings.Repeat("─", maxLen)
	top := "┌" + border + "┐"
	bottom := "└" + border + "┘"

	var middle string
	for _, line := range lines {
		padded := fmt.Sprintf("%-*s", maxLen, line)
		middle += "│" + padded + "│\n"
	}

	return top + "\n" + middle + bottom
}

// Hex formats a byte slice as hex in 8-char groups
func (s *Styles) Hex(data []byte) string {
	hexStr := fmt.Sprintf("%x", data)
	var result strings.Builder
	for i := 0; i < len(hexStr); i += 8 {
		end := i + 8
		if end > len(hexStr) {
			end = len(hexStr)
		}
		if i > 0 {
			result.WriteString(" ")
		}
		result.WriteString(hexStr[i:end])
	}
	return result.String()
}

// Separator creates a visual separator line
func (s *Styles) Separator(width int, char string) string {
	if width <= 0 {
		width = 40
	}
	if char == "" {
		char = "─"
	}
	return strings.Repeat(char, width)
}

// Header creates a styled header
func (s *Styles) Header(text string) string {
	return s.Bold(text)
}

// SubHeader creates a styled sub-header
func (s *Styles) SubHeader(text string) string {
	return s.Dim(text)
}

// Highlight creates highlighted text
func (s *Styles) Highlight(text string) string {
	return s.Bold(text)
}

// Warning creates warning text
func (s *Styles) Warning(text string) string {
	return s.ApplyColor(text, FGBrightRed)
}

// Success creates success text
func (s *Styles) Success(text string) string {
	return s.ApplyColor(text, FGBrightGreen)
}

// Info creates info text
func (s *Styles) Info(text string) string {
	return s.ApplyColor(text, FGBrightBlue)
}
