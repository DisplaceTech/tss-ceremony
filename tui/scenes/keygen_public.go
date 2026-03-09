package scenes

import (
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/decred/dcrd/dcrec/secp256k1/v4"
)

// PublicShareScene represents the public share computation and exchange (Scene 3)
type PublicShareScene struct {
	config      *Config
	styles      *Styles
	phase       int // 0: computing A, 1: computing B, 2: exchange, 3: complete
	step        int // animation step within phase
	started     bool
	duration    time.Duration
	partyAPub   string // Hex string of Party A's public share
	partyBPub   string // Hex string of Party B's public share
	secretA     []byte // Party A's secret (for computing public share)
	secretB     []byte // Party B's secret (for computing public share)
}

// NewPublicShareScene creates a new public share scene
func NewPublicShareScene(config *Config, styles *Styles) *PublicShareScene {
	return &PublicShareScene{
		config:   config,
		styles:   styles,
		phase:    0,
		step:     0,
		started:  false,
		duration: getStepDuration(config.Speed),
	}
}

// SetSecrets sets the secret shares for both parties (called by parent model)
func (s *PublicShareScene) SetSecrets(secretA, secretB []byte) {
	s.secretA = secretA
	s.secretB = secretB
}

// ComputePublicShares computes the public shares from secrets using protocol logic
func (s *PublicShareScene) ComputePublicShares() error {
	if s.secretA != nil && len(s.secretA) == 32 {
		// Compute Party A's public share: A = a × G
		privKey := secp256k1.PrivKeyFromBytes(s.secretA)
		if privKey != nil {
			pubKey := privKey.PubKey()
			s.partyAPub = formatPublicKeyHex(pubKey)
		}
	}
	if s.secretB != nil && len(s.secretB) == 32 {
		// Compute Party B's public share: B = b × G
		privKey := secp256k1.PrivKeyFromBytes(s.secretB)
		if privKey != nil {
			pubKey := privKey.PubKey()
			s.partyBPub = formatPublicKeyHex(pubKey)
		}
	}
	return nil
}

// formatPublicKeyHex formats a public key as hex string with 8-char groups
func formatPublicKeyHex(pubKey *secp256k1.PublicKey) string {
	// Use uncompressed format for display (0x04 prefix + x + y)
	bytes := pubKey.SerializeUncompressed()
	hexStr := hex.EncodeToString(bytes)
	// Format with spaces every 8 characters
	var result strings.Builder
	for i, c := range hexStr {
		if i > 0 && i%8 == 0 {
			result.WriteRune(' ')
		}
		result.WriteRune(c)
	}
	return result.String()
}

// Init initializes the scene
func (s *PublicShareScene) Init() tea.Cmd {
	s.started = true
	s.phase = 0
	s.step = 0
	s.partyAPub = ""
	s.partyBPub = ""
	return tea.Tick(s.duration, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

// Update handles events in the scene
func (s *PublicShareScene) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
		s.step++
		switch s.phase {
		case 0:
			// Computing Party A's public share
			if s.step >= 10 {
				s.phase = 1
				s.step = 0
			}
		case 1:
			// Computing Party B's public share
			if s.step >= 10 {
				s.phase = 2
				s.step = 0
			}
		case 2:
			// Exchange animation
			if s.step >= 15 {
				s.phase = 3
				s.step = 0
			}
		case 3:
			// Complete, keep ticking for auto-advance
		}
		return s, tea.Tick(s.duration, func(t time.Time) tea.Msg {
			return tickMsg(t)
		})
	}
	return s, nil
}

// Render renders the scene view
func (s *PublicShareScene) Render() string {
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
	builder.WriteString(headerStyle.Render("Public Share Computation") + "\n\n")

	// Separator
	builder.WriteString(separatorStyle.Render(strings.Repeat("─", 50)) + "\n\n")

	// Scalar multiplication animation
	builder.WriteString(labelStyle.Render("Scalar Multiplication:") + "\n\n")

	switch s.phase {
	case 0:
		// Computing Party A's public share
		builder.WriteString(s.renderScalarMult("a", "A", PartyAColor, s.step, 10) + "\n\n")
		builder.WriteString(s.renderPlaceholder("b", "B", PartyBColor) + "\n\n")
	case 1:
		// Computing Party B's public share
		builder.WriteString(s.renderCompletedScalarMult("a", "A", PartyAColor) + "\n\n")
		builder.WriteString(s.renderScalarMult("b", "B", PartyBColor, s.step, 10) + "\n\n")
	case 2:
		// Exchange animation
		builder.WriteString(s.renderExchangeAnimation(s.step, 15) + "\n\n")
	case 3:
		// Complete
		builder.WriteString(s.renderCompletedScalarMult("a", "A", PartyAColor) + "\n\n")
		builder.WriteString(s.renderCompletedScalarMult("b", "B", PartyBColor) + "\n\n")
	}

	// Status
	var status string
	if s.phase == 0 {
		status = "Computing Party A's public share: A = a × G"
	} else if s.phase == 1 {
		status = "Computing Party B's public share: B = b × G"
	} else if s.phase == 2 {
		status = "Exchanging public shares..."
	} else {
		status = "Public shares exchanged!"
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

// renderScalarMult renders the scalar multiplication animation
func (s *PublicShareScene) renderScalarMult(secret, public, color string, step, maxSteps int) string {
	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("  %s = %s × G", public, secret))
	
	// Progress bar
	arrowStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("226"))
	builder.WriteString("\n  ")
	for i := 0; i < maxSteps; i++ {
		if i < step {
			builder.WriteString(arrowStyle.Render("→"))
		} else {
			builder.WriteString("·")
		}
	}
	
	// Result - public share displayed in Yellow/Gold
	publicColor := lipgloss.NewStyle().Foreground(lipgloss.Color("226")) // Yellow/Gold
	if step >= maxSteps {
		// Use actual computed public share if available, otherwise generate placeholder
		displayValue := s.partyAPub
		if displayValue == "" {
			displayValue = s.generateRandomHex(128)
		}
		builder.WriteString(fmt.Sprintf("\n  %s%s%s = %s", publicColor.Render(public), Reset, Reset, displayValue))
	}
	return builder.String()
}

// renderCompletedScalarMult renders a completed scalar multiplication
func (s *PublicShareScene) renderCompletedScalarMult(secret, public, color string) string {
	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("  %s = %s × G", public, secret))
	// Public share displayed in Yellow/Gold
	publicColor := lipgloss.NewStyle().Foreground(lipgloss.Color("226")) // Yellow/Gold
	// Use actual computed public share if available, otherwise generate placeholder
	displayValue := s.partyAPub
	if displayValue == "" {
		displayValue = s.partyBPub
	}
	if displayValue == "" {
		displayValue = s.generateRandomHex(128)
	}
	builder.WriteString(fmt.Sprintf("\n  ↓ (scalar multiplication)\n  %s%s%s = %s", publicColor.Render(public), Reset, Reset, displayValue))
	return builder.String()
}

// renderPlaceholder renders a placeholder for not-yet-computed share
func (s *PublicShareScene) renderPlaceholder(secret, public, color string) string {
	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("  %s = %s × G", public, secret))
	builder.WriteString("\n  (waiting...)")
	return builder.String()
}

// renderExchangeAnimation renders the exchange animation
func (s *PublicShareScene) renderExchangeAnimation(step, maxSteps int) string {
	var builder strings.Builder
	
	// Party A's public share (in Yellow/Gold)
	publicColor := lipgloss.NewStyle().Foreground(lipgloss.Color("226")) // Yellow/Gold
	displayA := s.partyAPub
	if displayA == "" {
		displayA = s.generateRandomHex(128)
	}
	builder.WriteString(fmt.Sprintf("  %sA%s = a × G\n  %s%s%s = %s\n", 
		publicColor.Render(PartyAColor+"A"+Reset), Reset,
		publicColor.Render(PartyAColor+"A"+Reset), Reset, Reset, displayA))
	
	// Exchange arrows - animated bidirectional arrows
	arrowStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("226")) // Yellow/Gold for public exchange
	builder.WriteString("  ")
	for i := 0; i < maxSteps; i++ {
		if i < step {
			// Alternate between → and ← to show bidirectional exchange
			if i%2 == 0 {
				builder.WriteString(arrowStyle.Render("→"))
			} else {
				builder.WriteString(arrowStyle.Render("←"))
			}
		} else {
			builder.WriteString("·")
		}
	}
	builder.WriteString("\n")
	
	// Party B's public share (in Yellow/Gold)
	displayB := s.partyBPub
	if displayB == "" {
		displayB = s.generateRandomHex(128)
	}
	builder.WriteString(fmt.Sprintf("  %sB%s = b × G\n  %s%s%s = %s", 
		publicColor.Render(PartyBColor+"B"+Reset), Reset,
		publicColor.Render(PartyBColor+"B"+Reset), Reset, Reset, displayB))
	
	return builder.String()
}

// generateRandomHex generates a random hex string
func (s *PublicShareScene) generateRandomHex(length int) string {
	hexChars := "0123456789abcdef"
	var builder strings.Builder
	for i := 0; i < length; i++ {
		builder.WriteByte(hexChars[getRandomInt(16)])
		// Add space every 8 characters
		if (i+1)%8 == 0 && i < length-1 {
			builder.WriteString(" ")
		}
	}
	return builder.String()
}

// View renders the scene view (required by tea.Model interface)
func (s *PublicShareScene) View() string {
	return s.Render()
}

// Narrator returns the narrator text for this scene
func (s *PublicShareScene) Narrator() string {
	if s.phase == 0 {
		return "Party A computes their public share by multiplying their secret by the generator point G. This is scalar multiplication on the elliptic curve."
	} else if s.phase == 1 {
		return "Party B computes their public share the same way. Scalar multiplication is easy to compute but hard to reverse - this is the one-way property of elliptic curves."
	} else if s.phase == 2 {
		return "Now both parties exchange their public shares. Unlike secrets, public shares can be shared openly - anyone can see them."
	}
	return "Both parties now have each other's public shares. These will be used to compute the combined public key."
}
