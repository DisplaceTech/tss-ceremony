package scenes

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// MtAScene represents Scene 9: MtA Conversion Result
// It displays the Multiplicative-to-Additive conversion result showing
// how alpha and beta are produced such that α + β = k_a · k_b mod n.
type MtAScene struct {
	currentStep int
	maxSteps    int
	config      *Config
	styles      *Styles
}

// NewMtAScene creates a new MtA conversion result scene
func NewMtAScene(config *Config, styles *Styles) *MtAScene {
	return &MtAScene{
		currentStep: 0,
		maxSteps:    1,
		config:      config,
		styles:      styles,
	}
}

// Init initializes the scene
func (s *MtAScene) Init() tea.Cmd {
	return nil
}

// Update handles events in the scene
func (s *MtAScene) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return s, tea.Quit
		case "right", "l", " ":
			if s.currentStep < s.maxSteps {
				s.currentStep++
			}
		case "left", "h":
			if s.currentStep > 0 {
				s.currentStep--
			}
		}
	}
	return s, nil
}

// Render renders the scene view
func (s *MtAScene) Render() string {
	var builder strings.Builder

	// Header
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		MarginBottom(1).
		Foreground(lipgloss.Color("226")) // Yellow

	builder.WriteString(headerStyle.Render("Scene 9: Multiplicative-to-Additive (MtA) Result"))
	builder.WriteString("\n\n")

	// Progress indicator
	progress := ""
	for i := 0; i <= s.maxSteps; i++ {
		if i == s.currentStep {
			progress += lipgloss.NewStyle().Foreground(lipgloss.Color("226")).Render("●")
		} else if i < s.currentStep {
			progress += lipgloss.NewStyle().Foreground(lipgloss.Color("226")).Render("○")
		} else {
			progress += lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render("○")
		}
	}
	builder.WriteString(progress)
	builder.WriteString("\n\n")

	// Content based on current step
	switch s.currentStep {
	case 0:
		builder.WriteString(s.renderConcept())
	case 1:
		builder.WriteString(s.renderValues())
	}

	// Navigation hint
	navigationStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("243"))
	builder.WriteString("\n\n")
	builder.WriteString(navigationStyle.Render("[←/→ or h/l to navigate] [q to quit]"))

	return builder.String()
}

// renderConcept renders the MtA concept explanation
func (s *MtAScene) renderConcept() string {
	var builder strings.Builder

	explanationStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("252")).
		MarginLeft(2)

	equationStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("226")).
		Bold(true).
		Margin(1, 2)

	separatorStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241"))

	builder.WriteString(explanationStyle.Render("The MtA protocol converts a multiplicative sharing"))
	builder.WriteString("\n")
	builder.WriteString(explanationStyle.Render("into an additive sharing:"))
	builder.WriteString("\n\n")

	builder.WriteString(equationStyle.Render("α + β = k_a · k_b  (mod n)"))
	builder.WriteString("\n\n")

	builder.WriteString(separatorStyle.Render(strings.Repeat("─", 60)))
	builder.WriteString("\n\n")

	partyAStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("81")).
		Bold(true)

	partyBStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("213")).
		Bold(true)

	builder.WriteString(explanationStyle.Render("Each party receives one additive share:"))
	builder.WriteString("\n\n")
	builder.WriteString("  ")
	builder.WriteString(partyAStyle.Render("Party A"))
	builder.WriteString(explanationStyle.Render(" receives α (alpha)"))
	builder.WriteString("\n")
	builder.WriteString("  ")
	builder.WriteString(partyBStyle.Render("Party B"))
	builder.WriteString(explanationStyle.Render(" receives β (beta)"))
	builder.WriteString("\n\n")

	builder.WriteString(explanationStyle.Render("Neither party learns the other's share, yet their"))
	builder.WriteString("\n")
	builder.WriteString(explanationStyle.Render("shares sum to the product of the nonce shares."))

	return builder.String()
}

// renderValues renders the actual MtA values from the ceremony
func (s *MtAScene) renderValues() string {
	var builder strings.Builder

	labelStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241"))

	partyAStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("81")).
		Bold(true)

	partyBStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("213")).
		Bold(true)

	sharedStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("226")).
		Bold(true)

	separatorStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241"))

	// Alpha value
	builder.WriteString(partyAStyle.Render("Party A's share: α (alpha)"))
	builder.WriteString("\n")
	builder.WriteString(labelStyle.Render("  "))
	alphaHex := s.formatHex(s.config.Ceremony.AlphaHex)
	builder.WriteString(partyAStyle.Render(alphaHex))
	builder.WriteString("\n\n")

	// Beta value
	builder.WriteString(partyBStyle.Render("Party B's share: β (beta)"))
	builder.WriteString("\n")
	builder.WriteString(labelStyle.Render("  "))
	betaHex := s.formatHex(s.config.Ceremony.BetaHex)
	builder.WriteString(partyBStyle.Render(betaHex))
	builder.WriteString("\n\n")

	// Separator
	builder.WriteString(separatorStyle.Render(strings.Repeat("─", 60)))
	builder.WriteString("\n\n")

	// Equation reminder
	builder.WriteString(sharedStyle.Render("  α + β = k_a · k_b  (mod n)"))
	builder.WriteString("\n\n")

	explanationStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("243"))

	builder.WriteString(explanationStyle.Render("These additive shares will be used to compute"))
	builder.WriteString("\n")
	builder.WriteString(explanationStyle.Render("each party's partial signature in the next step."))

	return builder.String()
}

// formatHex formats a hex string in 8-character groups
func (s *MtAScene) formatHex(hexStr string) string {
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

// View renders the scene view (required by tea.Model interface)
func (s *MtAScene) View() string {
	return s.Render()
}

// Narrator returns the narrator text for this scene
func (s *MtAScene) Narrator() string {
	narrators := []string{
		"The Multiplicative-to-Additive (MtA) protocol is the key innovation that enables threshold ECDSA. It converts the product k_a · k_b into additive shares α and β, so that α + β equals the product. This lets each party compute a partial signature independently.",
		"Here are the actual MtA output values from this ceremony. Party A holds α and Party B holds β. Neither party knows the other's share, preserving the security of the protocol. These shares will feed directly into the partial signature computation.",
	}

	if s.currentStep >= len(narrators) {
		return narrators[len(narrators)-1]
	}
	return narrators[s.currentStep]
}
