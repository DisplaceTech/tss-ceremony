package scenes

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// RevealScene represents "The Reveal" scene (Scene 15)
// It explains why ECDSA requires the complex k⁻¹·x term and demonstrates
// the algebraic difference between ECDSA and simpler schemes like Schnorr.
type RevealScene struct {
	currentStep int
	maxSteps    int
}

// NewRevealScene creates a new reveal scene
func NewRevealScene() *RevealScene {
	return &RevealScene{
		currentStep: 0,
		maxSteps:    5,
	}
}

// Init initializes the scene
func (s *RevealScene) Init() tea.Cmd {
	return nil
}

// Update handles events in the scene
func (s *RevealScene) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
func (s *RevealScene) Render() string {
	// Define styles
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		MarginBottom(1).
		Foreground(lipgloss.Color("226")) // Yellow

	equationStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("214")). // Light blue
		Margin(1, 2).
		Bold(true)

	highlightStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("226")). // Yellow
		Bold(true)

	explanationStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("252")). // White
		MarginLeft(2)

	separatorStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")) // Dark gray

	// Build the view
	var builder string

	// Header
	builder += headerStyle.Render("The Reveal: Why ECDSA Needs k⁻¹·x") + "\n\n"

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
		builder += explanationStyle.Render("ECDSA signature generation requires computing:") + "\n\n"
		builder += equationStyle.Render("s = k⁻¹ · (z + r·x) mod n") + "\n\n"
		builder += explanationStyle.Render("Where:") + "\n"
		builder += explanationStyle.Render("  k = random nonce (kept secret)") + "\n"
		builder += explanationStyle.Render("  z = hash of message") + "\n"
		builder += explanationStyle.Render("  r = x-coordinate of k·G") + "\n"
		builder += explanationStyle.Render("  x = private key") + "\n"
		builder += "\n" + highlightStyle.Render("The k⁻¹·x term is the problem!") + "\n"
		builder += explanationStyle.Render("In threshold signing, x is split among parties.") + "\n"
		builder += explanationStyle.Render("Each party has a share xᵢ, but computing k⁻¹·x requires") + "\n"
		builder += explanationStyle.Render("combining shares while keeping k secret.") + "\n"

	case 1:
		builder += explanationStyle.Render("The Challenge: Computing k⁻¹·x in Threshold") + "\n\n"
		builder += explanationStyle.Render("Party A has share xₐ, Party B has share xᵦ") + "\n"
		builder += explanationStyle.Render("We need: k⁻¹·(xₐ + xᵦ) = k⁻¹·xₐ + k⁻¹·xᵦ") + "\n\n"
		builder += highlightStyle.Render("But k is also split!") + "\n"
		builder += explanationStyle.Render("Each party generates nonce share kᵢ") + "\n"
		builder += explanationStyle.Render("Combined nonce: k = kₐ + kᵦ") + "\n"
		builder += explanationStyle.Render("We need k⁻¹ = (kₐ + kᵦ)⁻¹") + "\n\n"
		builder += separatorStyle.Render("────────────────────────────────────────────────────────────────────────────────") + "\n\n"
		builder += highlightStyle.Render("Problem: (kₐ + kᵦ)⁻¹ ≠ kₐ⁻¹ + kᵦ⁻¹") + "\n"
		builder += explanationStyle.Render("You cannot simply invert individual shares!") + "\n"

	case 2:
		builder += explanationStyle.Render("Why Schnorr is Simpler") + "\n\n"
		builder += equationStyle.Render("Schnorr: s = k + e·x") + "\n\n"
		builder += explanationStyle.Render("Where:") + "\n"
		builder += explanationStyle.Render("  k = random nonce") + "\n"
		builder += explanationStyle.Render("  e = hash(R, message, public key)") + "\n"
		builder += explanationStyle.Render("  x = private key") + "\n\n"
		builder += highlightStyle.Render("No inverse needed!") + "\n"
		builder += explanationStyle.Render("Each party computes: sᵢ = kᵢ + e·xᵢ") + "\n"
		builder += explanationStyle.Render("Aggregate: s = Σsᵢ = Σkᵢ + e·Σxᵢ = k + e·x") + "\n"
		builder += explanationStyle.Render("Simple addition works perfectly!") + "\n"

	case 3:
		builder += explanationStyle.Render("The ECDSA Threshold Solution: MtA") + "\n\n"
		builder += explanationStyle.Render("To compute k⁻¹·xᵢ, we use") + "\n"
		builder += highlightStyle.Render("Multiplication-in-the-Exponent (MtA)") + "\n\n"
		builder += explanationStyle.Render("MtA Protocol:") + "\n"
		builder += explanationStyle.Render("  1. Party has secret a, wants to compute g^(a·b)") + "\n"
		builder += explanationStyle.Render("  2. Other party has secret b") + "\n"
		builder += explanationStyle.Render("  3. Neither reveals their secret") + "\n"
		builder += explanationStyle.Render("  4. Result: g^(a·b) computed jointly") + "\n\n"
		builder += separatorStyle.Render("────────────────────────────────────────────────────────────────────────────────") + "\n\n"
		builder += explanationStyle.Render("For ECDSA: a = k⁻¹, b = xᵢ") + "\n"
		builder += explanationStyle.Render("Result: g^(k⁻¹·xᵢ) enables partial signature") + "\n"

	case 4:
		builder += explanationStyle.Render("Why ECDSA Has This Complexity") + "\n\n"
		builder += highlightStyle.Render("Historical Design Decision") + "\n\n"
		builder += explanationStyle.Render("ECDSA was designed in the 1990s before") + "\n"
		builder += explanationStyle.Render("threshold signing was a major concern.") + "\n\n"
		builder += explanationStyle.Render("The formula s = k⁻¹·(z + r·x) provides:") + "\n"
		builder += explanationStyle.Render("  ✓ Strong security properties") + "\n"
		builder += explanationStyle.Render("  ✓ Resistance to certain attacks") + "\n"
		builder += explanationStyle.Render("  ✓ Compatibility with existing math libraries") + "\n\n"
		builder += separatorStyle.Render("────────────────────────────────────────────────────────────────────────────────") + "\n\n"
		builder += explanationStyle.Render("But it makes threshold signing HARD.") + "\n"

	case 5:
		builder += explanationStyle.Render("The Trade-off Summary") + "\n\n"
		builder += separatorStyle.Render("────────────────────────────────────────────────────────────────────────────────") + "\n\n"
		builder += "\n" + lipgloss.NewStyle().Foreground(lipgloss.Color("6")).Bold(true).Render("ECDSA (with DKLS):") + "\n"
		builder += explanationStyle.Render("  ✓ Bitcoin/Ethereum compatible") + "\n"
		builder += explanationStyle.Render("  ✓ Existing infrastructure works") + "\n"
		builder += explanationStyle.Render("  ✗ Complex threshold protocol") + "\n"
		builder += explanationStyle.Render("  ✗ Requires MtA and OT") + "\n"
		builder += explanationStyle.Render("  ✗ More communication rounds") + "\n\n"
		builder += lipgloss.NewStyle().Foreground(lipgloss.Color("202")).Bold(true).Render("Schnorr (with FROST):") + "\n"
		builder += explanationStyle.Render("  ✓ Simple threshold protocol") + "\n"
		builder += explanationStyle.Render("  ✓ No MtA or OT needed") + "\n"
		builder += explanationStyle.Render("  ✓ Fewer communication rounds") + "\n"
		builder += explanationStyle.Render("  ✗ New signature format") + "\n"
		builder += explanationStyle.Render("  ✗ Requires protocol adoption") + "\n"

	default:
		s.currentStep = 0
	}

	builder += "\n\n" + lipgloss.NewStyle().Foreground(lipgloss.Color("243")).Render("[←/→ or h/l to navigate] [q to quit]")

	return builder
}

// View renders the scene view (required by tea.Model interface)
func (s *RevealScene) View() string {
	return s.Render()
}

// Narrator returns the narrator text for this scene
func (s *RevealScene) Narrator() string {
	narrators := []string{
		"Welcome to The Reveal. This scene explains why ECDSA threshold signing is so complex. The culprit is the k⁻¹·x term in the signature formula.",
		"The fundamental challenge: in threshold ECDSA, both the private key x and the nonce k are split among parties. Computing k⁻¹ when k is split is impossible with simple arithmetic.",
		"Compare this to Schnorr signatures, which use a simple additive formula. No inverse is needed, making threshold signing straightforward with FROST.",
		"To overcome ECDSA's complexity, DKLS uses Multiplication-in-the-Exponent (MtA). This cryptographic protocol allows parties to jointly compute g^(k⁻¹·xᵢ) without revealing their secrets.",
		"ECDSA's complex formula was designed in the 1990s before threshold signing was a priority. It provides strong security but makes distributed signing difficult.",
		"The trade-off: ECDSA with DKLS gives you compatibility with Bitcoin and Ethereum, but at the cost of protocol complexity. Schnorr with FROST is simpler but requires new infrastructure.",
	}

	if s.currentStep >= len(narrators) {
		s.currentStep = len(narrators) - 1
	}
	return narrators[s.currentStep]
}
