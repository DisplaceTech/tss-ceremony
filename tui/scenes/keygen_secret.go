package scenes

import (
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// SecretGenScene represents the secret generation animation (Scene 2)
type SecretGenScene struct {
	config      *Config
	styles      *Styles
	partyAChars []rune
	partyBChars []rune
	partyAIndex int
	partyBIndex int
	phase       int // 0: Party A, 1: Party B, 2: complete
	started     bool
	duration    time.Duration
}

// NewSecretGenScene creates a new secret generation scene
func NewSecretGenScene(config *Config, styles *Styles) *SecretGenScene {
	return &SecretGenScene{
		config:      config,
		styles:      styles,
		partyAChars: make([]rune, 64), // 32 bytes = 64 hex chars
		partyBChars: make([]rune, 64),
		partyAIndex: 0,
		partyBIndex: 0,
		phase:       0,
		started:     false,
		duration:    getCharDuration(config.Speed),
	}
}

// Init initializes the scene
func (s *SecretGenScene) Init() tea.Cmd {
	s.started = true
	s.phase = 0
	s.partyAIndex = 0
	s.partyBIndex = 0
	return tea.Tick(s.duration, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

// Update handles events in the scene
func (s *SecretGenScene) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return s, tea.Quit
		case "enter", "right", "l", " ", "n", "j", "down":
			// Advance to next scene - handled by parent model
			return s, nil
		}

	case tickMsg:
		// Animation tick
		if s.phase == 0 && s.partyAIndex < 64 {
			s.partyAChars[s.partyAIndex] = s.getRandomHexChar()
			s.partyAIndex++
			if s.partyAIndex >= 64 {
				s.phase = 1
			}
		} else if s.phase == 1 && s.partyBIndex < 64 {
			s.partyBChars[s.partyBIndex] = s.getRandomHexChar()
			s.partyBIndex++
			if s.partyBIndex >= 64 {
				s.phase = 2
			}
		} else if s.phase == 2 {
			// Animation complete, keep ticking for auto-advance
		}
		return s, tea.Tick(s.duration, func(t time.Time) tea.Msg {
			return tickMsg(t)
		})
	}
	return s, nil
}

// getRandomHexChar returns a random hex character
func (s *SecretGenScene) getRandomHexChar() rune {
	hexChars := "0123456789abcdef"
	return rune(hexChars[getRandomInt(16)])
}

// Render renders the scene view
func (s *SecretGenScene) Render() string {
	// Build the view
	var builder strings.Builder

	// Styles
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("226")). // Yellow
		MarginBottom(1)

	labelStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")). // Gray
		MarginRight(2)

	separatorStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241"))

	// Header
	builder.WriteString(headerStyle.Render("Secret Generation") + "\n\n")

	// Separator
	builder.WriteString(separatorStyle.Render(strings.Repeat("─", 50)) + "\n\n")

	// Party A secret
	builder.WriteString(labelStyle.Render("Party A Secret (a):") + "\n")
	builder.WriteString(s.renderHexChars(s.partyAChars, s.partyAIndex, PartyAColor) + "\n\n")

	// Party B secret
	builder.WriteString(labelStyle.Render("Party B Secret (b):") + "\n")
	builder.WriteString(s.renderHexChars(s.partyBChars, s.partyBIndex, PartyBColor) + "\n\n")

	// Status
	var status string
	if s.phase == 0 {
		status = "Generating Party A's secret..."
	} else if s.phase == 1 {
		status = "Generating Party B's secret..."
	} else {
		status = "Secrets generated!"
	}
	statusStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("243"))
	builder.WriteString(statusStyle.Render(status) + "\n\n")

	// Navigation hint
	hintStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("243"))
	builder.WriteString(hintStyle.Render("Press Enter to continue..."))

	return builder.String()
}

// renderHexChars renders hex characters with color, showing progress
func (s *SecretGenScene) renderHexChars(chars []rune, index int, color string) string {
	var builder strings.Builder
	for i := 0; i < 64; i++ {
		if i < index {
			// Show the character with color
			if !s.styles.NoColor {
				builder.WriteString(color)
			}
			builder.WriteRune(chars[i])
			if !s.styles.NoColor {
				builder.WriteString(Reset)
			}
		} else {
			// Show placeholder
			builder.WriteRune('.')
		}
		// Add space every 8 characters
		if (i+1)%8 == 0 && i < 63 {
			builder.WriteString(" ")
		}
	}
	return builder.String()
}

// View renders the scene view (required by tea.Model interface)
func (s *SecretGenScene) View() string {
	return s.Render()
}

// Narrator returns the narrator text for this scene
func (s *SecretGenScene) Narrator() string {
	if s.phase == 0 {
		return "Party A generates a random 256-bit secret key. This secret will never leave Party A's device."
	} else if s.phase == 1 {
		return "Party B generates their own random 256-bit secret key. This secret will never leave Party B's device."
	}
	return "Both parties now have their private secrets. These secrets will never be revealed to each other or any third party."
}
