package scenes

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// FrostAnimatedScene represents the Animated FROST Signing scene (Scene 18)
// It animates the FROST signing process step-by-step, reusing animation components
// from the main ceremony to show how FROST collapses complexity.
type FrostAnimatedScene struct {
	config      *Config
	currentStep int
	maxSteps    int
}

// NewFrostAnimatedScene creates a new FROST animated scene
func NewFrostAnimatedScene(config *Config) *FrostAnimatedScene {
	return &FrostAnimatedScene{
		config:      config,
		currentStep: 0,
		maxSteps:    5,
	}
}

// Init initializes the scene
func (s *FrostAnimatedScene) Init() tea.Cmd {
	if s.config != nil && s.config.FixedMode {
		ResetFixedCounter()
		s.currentStep = 0
	}
	return nil
}

// Update handles events in the scene
func (s *FrostAnimatedScene) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
func (s *FrostAnimatedScene) Render() string {
	// Define styles
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		MarginBottom(1).
		Foreground(lipgloss.Color("226")) // Yellow

	partyAStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("6")) // Cyan
	partyBStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("202")) // Magenta
	sharedStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("226")) // Yellow
	separatorStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")) // Dark gray
	equationStyle := lipgloss.NewStyle().
		Bold(true).
		Margin(0, 2)
	highlightStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("226")).
		Bold(true)

	// Build the view
	var builder string

	// Header
	builder += headerStyle.Render("FROST Signing: Animated Step-by-Step") + "\n\n"

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
	builder += progress + "\n\n"

	// Content based on current step
	switch s.currentStep {
	case 0:
		builder += highlightStyle.Render("Step 1: Key Generation (DKG)") + "\n\n"
		builder += separatorStyle.Render("────────────────────────────────────────────────────────────────────────────────") + "\n\n"
		builder += partyAStyle.Render("Party A:") + "\n"
		builder += "  • Generates secret share xₐ\n"
		builder += "  • Creates public share Pₐ = xₐ·G\n"
		builder += "  • Broadcasts Pₐ to Party B\n"
		builder += "\n" + partyBStyle.Render("Party B:") + "\n"
		builder += "  • Generates secret share xᵦ\n"
		builder += "  • Creates public share Pᵦ = xᵦ·G\n"
		builder += "  • Broadcasts Pᵦ to Party A\n"
		builder += "\n" + sharedStyle.Render("Combined:") + "\n"
		builder += equationStyle.Render("  P = Pₐ + Pᵦ = (xₐ + xᵦ)·G = x·G") + "\n"
		builder += equationStyle.Render("  x = xₐ + xᵦ (never revealed)") + "\n"

	case 1:
		builder += highlightStyle.Render("Step 2: Nonce Generation") + "\n\n"
		builder += separatorStyle.Render("────────────────────────────────────────────────────────────────────────────────") + "\n\n"
		builder += partyAStyle.Render("Party A:") + "\n"
		builder += "  • Generates nonce share kₐ\n"
		builder += "  • Computes Rₐ = kₐ·G\n"
		builder += "  • Broadcasts Rₐ (no OT needed!)\n"
		builder += "\n" + partyBStyle.Render("Party B:") + "\n"
		builder += "  • Generates nonce share kᵦ\n"
		builder += "  • Computes Rᵦ = kᵦ·G\n"
		builder += "  • Broadcasts Rᵦ (no OT needed!)\n"
		builder += "\n" + sharedStyle.Render("Combined:") + "\n"
		builder += equationStyle.Render("  R = Rₐ + Rᵦ = (kₐ + kᵦ)·G = k·G") + "\n"
		builder += equationStyle.Render("  k = kₐ + kᵦ (never revealed)") + "\n"
		builder += "\n" + lipgloss.NewStyle().Foreground(lipgloss.Color("243")).Render("Note: FROST uses simple broadcast, no Oblivious Transfer!") + "\n"

	case 2:
		builder += highlightStyle.Render("Step 3: Compute Challenge") + "\n\n"
		builder += separatorStyle.Render("────────────────────────────────────────────────────────────────────────────────") + "\n\n"
		builder += sharedStyle.Render("Both parties compute:") + "\n"
		builder += equationStyle.Render("  e = H(R, message, P)") + "\n"
		builder += "\n" + lipgloss.NewStyle().Foreground(lipgloss.Color("243")).Render("Where:") + "\n"
		builder += lipgloss.NewStyle().Foreground(lipgloss.Color("243")).Render("  R = combined nonce point (Rₐ + Rᵦ)") + "\n"
		builder += lipgloss.NewStyle().Foreground(lipgloss.Color("243")).Render("  message = data to sign") + "\n"
		builder += lipgloss.NewStyle().Foreground(lipgloss.Color("243")).Render("  P = combined public key") + "\n"
		builder += "\n" + separatorStyle.Render("────────────────────────────────────────────────────────────────────────────────") + "\n\n"
		builder += highlightStyle.Render("The challenge e is deterministic and public!") + "\n"

	case 3:
		builder += highlightStyle.Render("Step 4: Partial Signatures") + "\n\n"
		builder += separatorStyle.Render("────────────────────────────────────────────────────────────────────────────────") + "\n\n"
		builder += partyAStyle.Render("Party A computes:") + "\n"
		builder += equationStyle.Render("  sₐ = kₐ + e·xₐ") + "\n"
		builder += "  • Simple addition and multiplication\n"
		builder += "  • No MtA needed!\n"
		builder += "  • Broadcasts sₐ to Party B\n"
		builder += "\n" + partyBStyle.Render("Party B computes:") + "\n"
		builder += equationStyle.Render("  sᵦ = kᵦ + e·xᵦ") + "\n"
		builder += "  • Simple addition and multiplication\n"
		builder += "  • No MtA needed!\n"
		builder += "  • Broadcasts sᵦ to Party A\n"
		builder += "\n" + separatorStyle.Render("────────────────────────────────────────────────────────────────────────────────") + "\n\n"
		builder += highlightStyle.Render("No complex cryptographic protocols required!") + "\n"

	case 4:
		builder += highlightStyle.Render("Step 5: Aggregation") + "\n\n"
		builder += separatorStyle.Render("────────────────────────────────────────────────────────────────────────────────") + "\n\n"
		builder += sharedStyle.Render("Combine partial signatures:") + "\n"
		builder += equationStyle.Render("  s = sₐ + sᵦ") + "\n"
		builder += equationStyle.Render("    = (kₐ + e·xₐ) + (kᵦ + e·xᵦ)") + "\n"
		builder += equationStyle.Render("    = (kₐ + kᵦ) + e·(xₐ + xᵦ)") + "\n"
		builder += equationStyle.Render("    = k + e·x") + "\n"
		builder += "\n" + sharedStyle.Render("Final signature:") + "\n"
		builder += equationStyle.Render("  (R, s) where R = Rₐ + Rᵦ") + "\n"
		builder += "\n" + separatorStyle.Render("────────────────────────────────────────────────────────────────────────────────") + "\n\n"
		builder += highlightStyle.Render("Simple addition produces valid Schnorr signature!") + "\n"

	case 5:
		builder += highlightStyle.Render("Step 6: Verification") + "\n\n"
		builder += separatorStyle.Render("────────────────────────────────────────────────────────────────────────────────") + "\n\n"
		builder += sharedStyle.Render("Anyone can verify:") + "\n"
		builder += equationStyle.Render("  1. Compute e = H(R, message, P)") + "\n"
		builder += equationStyle.Render("  2. Compute R' = s·G - e·P") + "\n"
		builder += equationStyle.Render("  3. Check: R' == R") + "\n"
		builder += "\n" + separatorStyle.Render("────────────────────────────────────────────────────────────────────────────────") + "\n\n"
		builder += highlightStyle.Render("FROST Summary:") + "\n"
		builder += "  ✓ No Oblivious Transfer\n"
		builder += "  ✓ No Multiplication-in-the-Exponent\n"
		builder += "  ✓ Simple partial signature formula\n"
		builder += "  ✓ Direct aggregation by addition\n"
		builder += "  ✓ Standard Schnorr verification\n"
		builder += "\n" + lipgloss.NewStyle().Foreground(lipgloss.Color("243")).Render("FROST collapses ECDSA complexity into elegant simplicity!") + "\n"

	default:
		s.currentStep = 0
	}

	builder += "\n\n" + lipgloss.NewStyle().Foreground(lipgloss.Color("243")).Render("[←/→ or h/l to navigate] [q to quit]")

	return builder
}

// View renders the scene view (required by tea.Model interface)
func (s *FrostAnimatedScene) View() string {
	return s.Render()
}

// Narrator returns the narrator text for this scene
func (s *FrostAnimatedScene) Narrator() string {
	narrators := []string{
		"Step 1: Key Generation. FROST uses the same DKG protocol as DKLS. Each party generates a secret share and broadcasts their public share. The combined public key is the sum of individual shares.",
		"Step 2: Nonce Generation. Unlike DKLS, FROST does not require Oblivious Transfer. Each party simply broadcasts their nonce point Rᵢ. The combined nonce is R = ΣRᵢ.",
		"Step 3: Compute Challenge. Both parties compute the same challenge e = H(R, message, P). This is deterministic and public, requiring no coordination.",
		"Step 4: Partial Signatures. Each party computes sᵢ = kᵢ + e·xᵢ. This is simple arithmetic - no MtA protocol needed! The linearity of Schnorr makes this trivial.",
		"Step 5: Aggregation. Simply add the partial signatures: s = sₐ + sᵦ = k + e·x. The algebra works out perfectly due to Schnorr's linear structure.",
		"Step 6: Verification. The final signature (R, s) is verified using standard Schnorr verification. No special threshold verification needed - it's a regular Schnorr signature!",
	}

	if s.currentStep >= len(narrators) {
		s.currentStep = len(narrators) - 1
	}
	return narrators[s.currentStep]
}

// GetStepInfo returns formatted step information for display
func (s *FrostAnimatedScene) GetStepInfo() string {
	steps := []string{
		"Key Generation (DKG)",
		"Nonce Generation",
		"Compute Challenge",
		"Partial Signatures",
		"Aggregation",
		"Verification",
	}
	if s.currentStep >= len(steps) {
		return steps[len(steps)-1]
	}
	return fmt.Sprintf("Step %d: %s", s.currentStep+1, steps[s.currentStep])
}
