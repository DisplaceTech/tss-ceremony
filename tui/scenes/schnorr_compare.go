package scenes

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// SchnorrCompareScene represents the Side-by-Side Schnorr vs ECDSA Comparison scene (Scene 16)
// It displays the mathematical equations for both Schnorr and ECDSA side-by-side,
// highlighting the structural differences.
type SchnorrCompareScene struct {
	currentStep int
	maxSteps    int
}

// NewSchnorrCompareScene creates a new Schnorr comparison scene
func NewSchnorrCompareScene() *SchnorrCompareScene {
	return &SchnorrCompareScene{
		currentStep: 0,
		maxSteps:    4,
	}
}

// Init initializes the scene
func (s *SchnorrCompareScene) Init() tea.Cmd {
	return nil
}

// Update handles events in the scene
func (s *SchnorrCompareScene) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
func (s *SchnorrCompareScene) Render() string {
	// Define styles
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		MarginBottom(1).
		Foreground(lipgloss.Color("226")) // Yellow

	schnorrStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("202")) // Magenta
	ecdsaStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("6")) // Cyan
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
	builder += headerStyle.Render("Schnorr vs ECDSA: Mathematical Comparison") + "\n\n"

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

	// Column headers
	builder += schnorrStyle.Render("Schnorr Signature") + "  " +
		separatorStyle.Render("│") + "  " +
		ecdsaStyle.Render("ECDSA Signature") + "\n"
	builder += separatorStyle.Render("────────────────────────────────────────────────────────────────────────────────") + "\n\n"

	// Content based on current step
	switch s.currentStep {
	case 0:
		// Signature generation equations
		builder += schnorrStyle.Render("Generation:") + "\n"
		builder += equationStyle.Render("  R = k·G") + "  " +
			separatorStyle.Render("│") + "  " +
			equationStyle.Render("  R = k·G") + "\n"
		builder += equationStyle.Render("  e = H(R, m, P)") + "  " +
			separatorStyle.Render("│") + "  " +
			equationStyle.Render("  r = x-coordinate of R") + "\n"
		builder += equationStyle.Render("  s = k + e·x") + "  " +
			separatorStyle.Render("│") + "  " +
			equationStyle.Render("  s = k⁻¹·(z + r·x)") + "\n"
		builder += "\n" + highlightStyle.Render("Key difference: Schnorr uses addition, ECDSA uses multiplication by inverse!") + "\n"

	case 1:
		// Signature verification equations
		builder += schnorrStyle.Render("Verification:") + "\n"
		builder += equationStyle.Render("  R' = s·G - e·P") + "  " +
			separatorStyle.Render("│") + "  " +
			equationStyle.Render("  u₁ = z·s⁻¹ mod n") + "\n"
		builder += equationStyle.Render("  Check: R' == R") + "  " +
			separatorStyle.Render("│") + "  " +
			equationStyle.Render("  u₂ = r·s⁻¹ mod n") + "\n"
		builder += equationStyle.Render("") + "  " +
			separatorStyle.Render("│") + "  " +
			equationStyle.Render("  R' = u₁·G + u₂·P") + "\n"
		builder += equationStyle.Render("") + "  " +
			separatorStyle.Render("│") + "  " +
			equationStyle.Render("  Check: r == x-coordinate of R'") + "\n"
		builder += "\n" + highlightStyle.Render("Schnorr verification is simpler: one linear combination!") + "\n"

	case 2:
		// Threshold signing comparison
		builder += schnorrStyle.Render("Threshold (FROST):") + "\n"
		builder += equationStyle.Render("  sᵢ = kᵢ + e·xᵢ") + "  " +
			separatorStyle.Render("│") + "  " +
			equationStyle.Render("  Requires MtA for k⁻¹·xᵢ") + "\n"
		builder += equationStyle.Render("  s = Σsᵢ") + "  " +
			separatorStyle.Render("│") + "  " +
			equationStyle.Render("  Requires OT for nonce privacy") + "\n"
		builder += equationStyle.Render("  = Σkᵢ + e·Σxᵢ") + "  " +
			separatorStyle.Render("│") + "  " +
			equationStyle.Render("  Complex partial sig formula") + "\n"
		builder += equationStyle.Render("  = k + e·x") + "  " +
			separatorStyle.Render("│") + "  " +
			equationStyle.Render("  Phantom key derivation") + "\n"
		builder += "\n" + highlightStyle.Render("FROST: simple addition. DKLS: complex cryptographic protocols!") + "\n"

	case 3:
		// Security properties
		builder += schnorrStyle.Render("Security Properties:") + "\n"
		builder += equationStyle.Render("  ✓ Proven secure in ROM") + "  " +
			separatorStyle.Render("│") + "  " +
			equationStyle.Render("  ✓ Proven secure in ROM") + "\n"
		builder += equationStyle.Render("  ✓ Linear structure") + "  " +
			separatorStyle.Render("│") + "  " +
			equationStyle.Render("  ✓ Non-linear structure") + "\n"
		builder += equationStyle.Render("  ✓ Simple threshold") + "  " +
			separatorStyle.Render("│") + "  " +
			equationStyle.Render("  ✗ Complex threshold") + "\n"
		builder += equationStyle.Render("  ✗ New standard") + "  " +
			separatorStyle.Render("│") + "  " +
			equationStyle.Render("  ✓ Legacy compatibility") + "\n"
		builder += "\n" + highlightStyle.Render("Both are secure! ECDSA trades simplicity for compatibility.") + "\n"

	case 4:
		// Summary
		builder += highlightStyle.Render("Summary: The Structural Difference") + "\n\n"
		builder += schnorrStyle.Render("Schnorr: s = k + e·x") + "\n"
		builder += "  • Linear equation" + "\n"
		builder += "  • Additive homomorphism" + "\n"
		builder += "  • Natural threshold support" + "\n"
		builder += "\n" + ecdsaStyle.Render("ECDSA: s = k⁻¹·(z + r·x)") + "\n"
		builder += "  • Non-linear equation" + "\n"
		builder += "  • Multiplicative inverse" + "\n"
		builder += "  • Requires complex threshold protocols" + "\n"
		builder += "\n" + separatorStyle.Render("────────────────────────────────────────────────────────────────────────────────") + "\n"
		builder += highlightStyle.Render("The k⁻¹ term is what makes ECDSA threshold signing hard!") + "\n"

	default:
		s.currentStep = 0
	}

	builder += "\n\n" + lipgloss.NewStyle().Foreground(lipgloss.Color("243")).Render("[←/→ or h/l to navigate] [q to quit]")

	return builder
}

// View renders the scene view (required by tea.Model interface)
func (s *SchnorrCompareScene) View() string {
	return s.Render()
}

// Narrator returns the narrator text for this scene
func (s *SchnorrCompareScene) Narrator() string {
	narrators := []string{
		"Welcome to the mathematical comparison. Here we see the core equations for both Schnorr and ECDSA signatures side-by-side.",
		"Verification equations show the difference clearly. Schnorr uses a single linear combination, while ECDSA requires two scalar multiplications.",
		"In threshold signing, Schnorr's linear structure allows simple addition of partial signatures. ECDSA's non-linear structure requires complex protocols like MtA and OT.",
		"Both schemes are provably secure in the Random Oracle Model. The difference is structural: Schnorr is linear, ECDSA is non-linear due to the k⁻¹ term.",
		"The summary: Schnorr's s = k + e·x is naturally threshold-friendly. ECDSA's s = k⁻¹·(z + r·x) requires complex workarounds for threshold signing.",
	}

	if s.currentStep >= len(narrators) {
		s.currentStep = len(narrators) - 1
	}
	return narrators[s.currentStep]
}
