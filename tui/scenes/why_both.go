package scenes

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// WhyBothScene represents the "Why Both Exist" scene (Scene 19)
// It explains the practical reasons for maintaining both ECDSA and FROST/Schnorr
// in the ecosystem, including infrastructure inertia and compatibility concerns.
type WhyBothScene struct {
	currentStep int
	maxSteps    int
}

// NewWhyBothScene creates a new "Why Both Exist" scene
func NewWhyBothScene() *WhyBothScene {
	return &WhyBothScene{
		currentStep: 0,
		maxSteps:    5,
	}
}

// Init initializes the scene
func (s *WhyBothScene) Init() tea.Cmd {
	return nil
}

// Update handles events in the scene
func (s *WhyBothScene) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
func (s *WhyBothScene) Render() string {
	// Define styles
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		MarginBottom(1).
		Foreground(lipgloss.Color("226")) // Yellow

	ecdsaStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("6")) // Cyan
	schnorrStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("202")) // Magenta
	separatorStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")) // Dark gray
	explanationStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("252")). // White
		MarginLeft(2)
	highlightStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("226")).
		Bold(true)

	// Build the view
	var builder string

	// Header
	builder += headerStyle.Render("Why Both ECDSA and Schnorr/FROST Exist") + "\n\n"

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
		builder += explanationStyle.Render("The Reality: Both Protocols Coexist") + "\n\n"
		builder += separatorStyle.Render("────────────────────────────────────────────────────────────────────────────────") + "\n\n"
		builder += highlightStyle.Render("Infrastructure Inertia") + "\n\n"
		builder += explanationStyle.Render("Bitcoin launched in 2009 with ECDSA.") + "\n"
		builder += explanationStyle.Render("Ethereum launched in 2015 with ECDSA.") + "\n"
		builder += explanationStyle.Render("Trillions of dollars in value secured by ECDSA.") + "\n"
		builder += explanationStyle.Render("Decades of tooling, libraries, and expertise built around ECDSA.") + "\n\n"
		builder += separatorStyle.Render("────────────────────────────────────────────────────────────────────────────────") + "\n\n"
		builder += explanationStyle.Render("You cannot simply 'replace' ECDSA with Schnorr.") + "\n"
		builder += explanationStyle.Render("The ecosystem is too large and too valuable.") + "\n"

	case 1:
		builder += highlightStyle.Render("Bitcoin's Partial Adoption of Schnorr") + "\n\n"
		builder += explanationStyle.Render("Bitcoin Taproot (2021) introduced Schnorr signatures") + "\n"
		builder += explanationStyle.Render("for new transaction types, but:") + "\n\n"
		builder += lipgloss.NewStyle().Foreground(lipgloss.Color("243")).Render("  • Legacy addresses still use ECDSA") + "\n"
		builder += lipgloss.NewStyle().Foreground(lipgloss.Color("243")).Render("  • Most wallets default to ECDSA for compatibility") + "\n"
		builder += lipgloss.NewStyle().Foreground(lipgloss.Color("243")).Render("  • Schnorr is optional, not mandatory") + "\n\n"
		builder += separatorStyle.Render("────────────────────────────────────────────────────────────────────────────────") + "\n\n"
		builder += explanationStyle.Render("This gradual adoption pattern is typical.") + "\n"
		builder += explanationStyle.Render("New protocols complement, not replace, existing ones.") + "\n"

	case 2:
		builder += highlightStyle.Render("Ethereum's Path Forward") + "\n\n"
		builder += explanationStyle.Render("Ethereum's roadmap includes BLS signatures for") + "\n"
		builder += explanationStyle.Render("consensus (Beacon Chain), but ECDSA remains") + "\n"
		builder += explanationStyle.Render("for account transactions.") + "\n\n"
		builder += lipgloss.NewStyle().Foreground(lipgloss.Color("243")).Render("  • ECDSA: Account abstraction, EVM compatibility") + "\n"
		builder += lipgloss.NewStyle().Foreground(lipgloss.Color("243")).Render("  • BLS: Consensus aggregation, staking") + "\n"
		builder += lipgloss.NewStyle().Foreground(lipgloss.Color("243")).Render("  • Schnorr: Not currently in roadmap") + "\n\n"
		builder += separatorStyle.Render("────────────────────────────────────────────────────────────────────────────────") + "\n\n"
		builder += explanationStyle.Render("Different use cases favor different signature schemes.") + "\n"
		builder += explanationStyle.Render("Ethereum uses the right tool for each job.") + "\n"

	case 3:
		builder += highlightStyle.Render("The Threshold Signing Landscape") + "\n\n"
		builder += schnorrStyle.Render("FROST/Schnorr:") + "\n"
		builder += explanationStyle.Render("  • New hardware wallets adopting FROST") + "\n"
		builder += explanationStyle.Render("  • Threshold signing services prefer FROST") + "\n"
		builder += explanationStyle.Render("  • Academic research focuses on FROST variants") + "\n"
		builder += explanationStyle.Render("  • Future-proof for new protocols") + "\n\n"
		builder += ecdsaStyle.Render("DKLS/ECDSA:") + "\n"
		builder += explanationStyle.Render("  • Required for Bitcoin multisig upgrades") + "\n"
		builder += explanationStyle.Render("  • Essential for Ethereum account abstraction") + "\n"
		builder += explanationStyle.Render("  • Legacy hardware wallet compatibility") + "\n"
		builder += explanationStyle.Render("  • Regulatory compliance with existing standards") + "\n"

	case 4:
		builder += highlightStyle.Render("Practical Considerations for Adoption") + "\n\n"
		builder += separatorStyle.Render("────────────────────────────────────────────────────────────────────────────────") + "\n\n"
		builder += explanationStyle.Render("When choosing a threshold signing protocol:") + "\n\n"
		builder += lipgloss.NewStyle().Foreground(lipgloss.Color("243")).Render("  1. What assets do you need to sign?") + "\n"
		builder += lipgloss.NewStyle().Foreground(lipgloss.Color("243")).Render("     - Bitcoin/Ethereum → ECDSA (DKLS)") + "\n"
		builder += lipgloss.NewStyle().Foreground(lipgloss.Color("243")).Render("     - New chains → Consider Schnorr (FROST)") + "\n\n"
		builder += lipgloss.NewStyle().Foreground(lipgloss.Color("243")).Render("  2. What's your infrastructure?") + "\n"
		builder += lipgloss.NewStyle().Foreground(lipgloss.Color("243")).Render("     - Existing ECDSA tooling → DKLS") + "\n"
		builder += lipgloss.NewStyle().Foreground(lipgloss.Color("243")).Render("     - Greenfield project → FROST") + "\n\n"
		builder += lipgloss.NewStyle().Foreground(lipgloss.Color("243")).Render("  3. What's your timeline?") + "\n"
		builder += lipgloss.NewStyle().Foreground(lipgloss.Color("243")).Render("     - Immediate needs → ECDSA compatibility") + "\n"
		builder += lipgloss.NewStyle().Foreground(lipgloss.Color("243")).Render("     - Long-term planning → Consider FROST") + "\n"

	case 5:
		builder += highlightStyle.Render("The Future: Coexistence and Convergence") + "\n\n"
		builder += separatorStyle.Render("────────────────────────────────────────────────────────────────────────────────") + "\n\n"
		builder += explanationStyle.Render("Both ECDSA and Schnorr will continue to exist:") + "\n\n"
		builder += explanationStyle.Render("  ✓ ECDSA for legacy compatibility") + "\n"
		builder += explanationStyle.Render("  ✓ Schnorr for new deployments") + "\n"
		builder += explanationStyle.Render("  ✓ DKLS bridges ECDSA to threshold signing") + "\n"
		builder += explanationStyle.Render("  ✓ FROST provides simple threshold Schnorr") + "\n\n"
		builder += separatorStyle.Render("────────────────────────────────────────────────────────────────────────────────") + "\n\n"
		builder += highlightStyle.Render("The ceremony you've seen (DKLS) enables threshold ECDSA") + "\n"
		builder += highlightStyle.Render("for the ecosystems that need it most.") + "\n"
		builder += highlightStyle.Render("FROST offers simplicity for new protocols.") + "\n"
		builder += highlightStyle.Render("Both have their place in the cryptographic landscape.") + "\n"

	default:
		s.currentStep = 0
	}

	builder += "\n\n" + lipgloss.NewStyle().Foreground(lipgloss.Color("243")).Render("[←/→ or h/l to navigate] [q to quit]")

	return builder
}

// View renders the scene view (required by tea.Model interface)
func (s *WhyBothScene) View() string {
	return s.Render()
}

// Narrator returns the narrator text for this scene
func (s *WhyBothScene) Narrator() string {
	narrators := []string{
		"Welcome to the final scene. Here we explore why both ECDSA and Schnorr/FROST coexist in the ecosystem. The answer lies in infrastructure inertia and practical compatibility concerns.",
		"Bitcoin's partial adoption of Schnorr via Taproot shows the pattern: new protocols complement existing ones. Legacy addresses still use ECDSA, and most wallets default to ECDSA for compatibility.",
		"Ethereum uses BLS for consensus but keeps ECDSA for transactions. Different use cases favor different signature schemes. The ecosystem uses the right tool for each job.",
		"In threshold signing, FROST/Schnorr is preferred for new deployments, while DKLS/ECDSA is essential for Bitcoin and Ethereum compatibility. Each has its place.",
		"When choosing a protocol, consider: what assets do you need to sign, what infrastructure exists, and what's your timeline. ECDSA for immediate compatibility, FROST for greenfield projects.",
		"The future is coexistence. ECDSA for legacy compatibility, Schnorr for new deployments. DKLS bridges ECDSA to threshold signing for ecosystems that need it. Both protocols have their place.",
	}

	if s.currentStep >= len(narrators) {
		s.currentStep = len(narrators) - 1
	}
	return narrators[s.currentStep]
}
