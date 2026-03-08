// Package tui provides the terminal user interface for the TSS ceremony.
package tui

import (
	"github.com/charmbracelet/lipgloss"
)

// NewStyles creates a new Style configuration with default styling.
func NewStyles() Style {
	return Style{
		Title: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#61afef")).
			MarginBottom(1),

		Description: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#d19a66")).
			MarginBottom(1),

		Success: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#98c379")).
			Bold(true),

		Error: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#e06c75")).
			Bold(true),

		Info: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#56b6c2")),

		Warning: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#e5c07b")).
			Bold(true),

		Code: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#abb2bf")).
			Background(lipgloss.Color("#3c3c3c")).
			Padding(0, 1),
	}
}

// Base styles for common UI elements
var (
	// Border styles
	Border = lipgloss.RoundedBorder()

	// Box style for containers
	Box = lipgloss.NewStyle().
		Border(Border).
		BorderForeground(lipgloss.Color("#3e4451")).
		Padding(1, 2)

	// Highlight style for selected items
	Highlight = lipgloss.NewStyle().
		Background(lipgloss.Color("#3e4451")).
		Foreground(lipgloss.Color("#ffffff"))

	// Dimmed style for inactive elements
	Dimmed = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#5c6370"))

	// Bold style for emphasis
	Bold = lipgloss.NewStyle().Bold(true)

	// Italic style for notes
	Italic = lipgloss.NewStyle().Italic(true)
)
