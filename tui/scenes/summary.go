package scenes

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// SummaryScene represents Scene 14: Final Ceremony Summary
// It provides a comprehensive summary of the entire signing ceremony.
type SummaryScene struct {
	currentStep int
	maxSteps    int
	config      *Config
	styles      *Styles
}

// NewSummaryScene creates a new ceremony summary scene
func NewSummaryScene(config *Config, styles *Styles) *SummaryScene {
	return &SummaryScene{
		currentStep: 0,
		maxSteps:    2,
		config:      config,
		styles:      styles,
	}
}

// Init initializes the scene
func (s *SummaryScene) Init() tea.Cmd {
	return nil
}

// Update handles events in the scene
func (s *SummaryScene) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
func (s *SummaryScene) Render() string {
	var builder strings.Builder

	// Header
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		MarginBottom(1).
		Foreground(lipgloss.Color("226")) // Yellow

	builder.WriteString(headerStyle.Render("Scene 14: Ceremony Summary"))
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
		builder.WriteString(s.renderSummary())
	case 1:
		builder.WriteString(s.renderVerifyCommand())
	case 2:
		builder.WriteString(s.renderWhatWeProved())
	}

	// Navigation hint
	navigationStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("243"))
	builder.WriteString("\n\n")
	builder.WriteString(navigationStyle.Render("[←/→ or h/l to navigate] [q to quit]"))

	return builder.String()
}

// renderSummary renders the ceremony summary with all key values
func (s *SummaryScene) renderSummary() string {
	var builder strings.Builder

	sharedStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("226")).
		Bold(true)

	labelStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("252"))

	separatorStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241"))

	builder.WriteString(sharedStyle.Render("Ceremony Complete"))
	builder.WriteString("\n")
	builder.WriteString(separatorStyle.Render(strings.Repeat("─", 60)))
	builder.WriteString("\n\n")

	// Combined Public Key
	builder.WriteString(labelStyle.Render("Combined Public Key:"))
	builder.WriteString("\n")
	builder.WriteString("  ")
	builder.WriteString(sharedStyle.Render(s.formatHex(s.config.Ceremony.CombinedPubHex)))
	builder.WriteString("\n\n")

	// Message
	builder.WriteString(labelStyle.Render("Message:"))
	builder.WriteString("\n")
	builder.WriteString("  ")
	builder.WriteString(labelStyle.Render(s.config.Ceremony.MessageText))
	builder.WriteString("\n\n")

	// Message Hash
	builder.WriteString(labelStyle.Render("Message Hash (SHA-256):"))
	builder.WriteString("\n")
	builder.WriteString("  ")
	builder.WriteString(labelStyle.Render(s.formatHex(s.config.Ceremony.MessageHash)))
	builder.WriteString("\n\n")

	// Signature
	builder.WriteString(separatorStyle.Render(strings.Repeat("─", 60)))
	builder.WriteString("\n\n")

	builder.WriteString(sharedStyle.Render("ECDSA Signature (r, s):"))
	builder.WriteString("\n\n")

	builder.WriteString(labelStyle.Render("  r = "))
	builder.WriteString(sharedStyle.Render(s.formatHex(s.config.Ceremony.SignatureRHex)))
	builder.WriteString("\n\n")

	builder.WriteString(labelStyle.Render("  s = "))
	builder.WriteString(sharedStyle.Render(s.formatHex(s.config.Ceremony.SignatureSHex)))
	builder.WriteString("\n\n")

	// Verification status
	if s.config.Ceremony.Valid {
		validStyle := lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("46")) // Green
		builder.WriteString(validStyle.Render("  Status: VALID"))
	} else {
		invalidStyle := lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("196")) // Red
		builder.WriteString(invalidStyle.Render("  Status: INVALID"))
	}
	builder.WriteString("\n")

	return builder.String()
}

// renderVerifyCommand renders the verification command
func (s *SummaryScene) renderVerifyCommand() string {
	var builder strings.Builder

	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("226"))

	separatorStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241"))

	labelStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("252"))

	builder.WriteString(headerStyle.Render("Verify This Ceremony"))
	builder.WriteString("\n")
	builder.WriteString(separatorStyle.Render(strings.Repeat("─", 60)))
	builder.WriteString("\n\n")

	builder.WriteString(labelStyle.Render("To independently verify this signature, run:"))
	builder.WriteString("\n\n")

	// Build verify command with real values
	cmdStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("243")).
		Background(lipgloss.Color("236")).
		Padding(1, 2)

	verifyCmd := fmt.Sprintf("tss-ceremony --verify \\\n  --pubkey %s \\\n  --sig-r %s \\\n  --sig-s %s \\\n  --message \"%s\"",
		s.config.Ceremony.CombinedPubHex,
		s.config.Ceremony.SignatureRHex,
		s.config.Ceremony.SignatureSHex,
		s.config.Ceremony.MessageText,
	)

	builder.WriteString(cmdStyle.Render(verifyCmd))
	builder.WriteString("\n\n")

	explanationStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("243"))

	builder.WriteString(explanationStyle.Render("This command performs standard ECDSA verification on secp256k1."))
	builder.WriteString("\n")
	builder.WriteString(explanationStyle.Render("The signature is indistinguishable from one produced by a single signer."))
	builder.WriteString("\n")

	return builder.String()
}

// renderWhatWeProved renders the "what we proved" summary
func (s *SummaryScene) renderWhatWeProved() string {
	var builder strings.Builder

	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("226"))

	separatorStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241"))

	explanationStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("252")).
		MarginLeft(2)

	partyAStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("81")).
		Bold(true)

	partyBStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("213")).
		Bold(true)

	phantomStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("196")).
		Bold(true)

	builder.WriteString(headerStyle.Render("What This Ceremony Proved"))
	builder.WriteString("\n")
	builder.WriteString(separatorStyle.Render(strings.Repeat("─", 60)))
	builder.WriteString("\n\n")

	builder.WriteString(explanationStyle.Render("1. Two parties can jointly produce a valid ECDSA signature"))
	builder.WriteString("\n")
	builder.WriteString(explanationStyle.Render("   without either party ever holding the full private key."))
	builder.WriteString("\n\n")

	builder.WriteString(explanationStyle.Render("2. "))
	builder.WriteString(partyAStyle.Render("Party A"))
	builder.WriteString(explanationStyle.Render(" kept their secret share private."))
	builder.WriteString("\n")
	builder.WriteString(explanationStyle.Render("   "))
	builder.WriteString(partyBStyle.Render("Party B"))
	builder.WriteString(explanationStyle.Render(" kept their secret share private."))
	builder.WriteString("\n\n")

	builder.WriteString(explanationStyle.Render("3. The "))
	builder.WriteString(phantomStyle.Render("phantom private key"))
	builder.WriteString(explanationStyle.Render(" (a + b) was NEVER computed."))
	builder.WriteString("\n")
	builder.WriteString(explanationStyle.Render("   It exists only as a mathematical concept."))
	builder.WriteString("\n\n")

	builder.WriteString(explanationStyle.Render("4. The resulting signature is a standard ECDSA signature"))
	builder.WriteString("\n")
	builder.WriteString(explanationStyle.Render("   on secp256k1 — verifiable by any standard tool."))
	builder.WriteString("\n\n")

	builder.WriteString(explanationStyle.Render("5. This is the power of the DKLS23 threshold protocol:"))
	builder.WriteString("\n")
	builder.WriteString(explanationStyle.Render("   security through distribution, compatibility through"))
	builder.WriteString("\n")
	builder.WriteString(explanationStyle.Render("   standard cryptographic output."))
	builder.WriteString("\n")

	return builder.String()
}

// formatHex formats a hex string in 8-character groups
func (s *SummaryScene) formatHex(hexStr string) string {
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
func (s *SummaryScene) View() string {
	return s.Render()
}

// Narrator returns the narrator text for this scene
func (s *SummaryScene) Narrator() string {
	narrators := []string{
		"The ceremony is complete. Two parties have jointly produced a valid ECDSA signature on secp256k1 without either party ever knowing the full private key. The signature is mathematically identical to one produced by a single signer.",
		"You can verify this signature independently using the command shown. The verification uses standard ECDSA math — no special threshold-aware tools are needed. This is what makes DKLS practical for real-world use.",
		"This ceremony demonstrated the core insight of threshold cryptography: security does not require trust in a single party. By distributing the private key across multiple parties, we eliminate single points of failure while maintaining full compatibility with existing systems.",
	}

	if s.currentStep >= len(narrators) {
		return narrators[len(narrators)-1]
	}
	return narrators[s.currentStep]
}
