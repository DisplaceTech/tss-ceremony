package scenes

import (
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ImpossibilityScene represents Scene 13: Security Proof Visualization.
// It illustrates why forging a signature without both secret shares is
// computationally impossible, using locked-box and missing-piece metaphors.
type ImpossibilityScene struct {
	config      *Config
	styles      *Styles
	currentStep int
	maxSteps    int
	tick        int
	duration    time.Duration
	// Interactive forgery attempt state
	forgeryAttempted bool
	forgeryStep      int
	forgeryMaxSteps  int
}

// NewImpossibilityScene creates a new impossibility scene.
func NewImpossibilityScene(config *Config, styles *Styles) *ImpossibilityScene {
	return &ImpossibilityScene{
		config:        config,
		styles:        styles,
		maxSteps:      5,
		duration:      getSceneDuration(config.Speed),
		forgeryMaxSteps: 3,
	}
}

// Init initializes the scene.
func (s *ImpossibilityScene) Init() tea.Cmd {
	return tea.Tick(s.duration, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

// Update handles events.
func (s *ImpossibilityScene) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return s, tea.Quit
		case "right", "l", " ":
			if s.currentStep < s.maxSteps {
				s.currentStep++
				// Reset forgery state when entering forgery step
				if s.currentStep == 4 {
					s.forgeryAttempted = false
					s.forgeryStep = 0
				}
			}
		case "left", "h":
			if s.currentStep > 0 {
				s.currentStep--
			}
		case "f":
			// Interactive forgery attempt - only works in step 4
			if s.currentStep == 4 && !s.forgeryAttempted {
				s.forgeryAttempted = true
				s.forgeryStep = 1
			} else if s.currentStep == 4 && s.forgeryAttempted && s.forgeryStep < s.forgeryMaxSteps {
				s.forgeryStep++
			}
		}
	case tickMsg:
		s.tick++
		return s, tea.Tick(s.duration, func(t time.Time) tea.Msg {
			return tickMsg(t)
		})
	}
	return s, nil
}

// Render renders the scene view.
func (s *ImpossibilityScene) Render() string {
	var b strings.Builder

	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("226")).MarginBottom(1)
	b.WriteString(headerStyle.Render("Security Proof: Impossibility of Forgery"))
	b.WriteString("\n\n")

	// Progress dots
	for i := 0; i <= s.maxSteps; i++ {
		if i == s.currentStep {
			b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("226")).Render("●"))
		} else if i < s.currentStep {
			b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("226")).Render("○"))
		} else {
			b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render("○"))
		}
	}
	b.WriteString("\n\n")

	switch s.currentStep {
	case 0:
		b.WriteString(s.renderLockedBoxes())
	case 1:
		b.WriteString(s.renderMissingPiece())
	case 2:
		b.WriteString(s.renderAttackerView())
	case 3:
		b.WriteString(s.renderMathProof())
	case 4:
		b.WriteString(s.renderInteractiveForgery())
	case 5:
		b.WriteString(s.renderConclusion())
	}

	b.WriteString("\n\n")
	b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("243")).Render("[←/→ or h/l to navigate] [f to attempt forgery] [q to quit]"))
	return b.String()
}

// renderLockedBoxes shows the two secret shares as locked boxes.
func (s *ImpossibilityScene) renderLockedBoxes() string {
	var b strings.Builder

	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("226"))
	partyAStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("81"))  // cyan
	partyBStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("213")) // magenta
	dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("243"))

	b.WriteString(titleStyle.Render("The Secret Shares Are Locked Away") + "\n\n")

	b.WriteString(partyAStyle.Render("  Party A") + "                    " + partyBStyle.Render("Party B") + "\n")
	b.WriteString(partyAStyle.Render("  ┌─────────────┐") + "          " + partyBStyle.Render("┌─────────────┐") + "\n")
	b.WriteString(partyAStyle.Render("  │  🔒  share_a │") + "          " + partyBStyle.Render("│  🔒  share_b │") + "\n")
	b.WriteString(partyAStyle.Render("  │  ??????????  │") + "          " + partyBStyle.Render("│  ??????????  │") + "\n")
	b.WriteString(partyAStyle.Render("  └─────────────┘") + "          " + partyBStyle.Render("└─────────────┘") + "\n\n")

	b.WriteString(dimStyle.Render("  share_a + share_b = private_key  (but neither party ever sees this sum)") + "\n\n")
	b.WriteString(dimStyle.Render("  Each share alone is uniformly random — it reveals nothing about the key.") + "\n")
	b.WriteString(dimStyle.Render("  An attacker holding only one box cannot reconstruct the private key.") + "\n")
	return b.String()
}

// renderMissingPiece shows the puzzle-piece metaphor for an incomplete signature.
func (s *ImpossibilityScene) renderMissingPiece() string {
	var b strings.Builder

	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("226"))
	redStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("196"))
	dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("243"))
	goodStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("82"))

	b.WriteString(titleStyle.Render("A Signature Needs Both Pieces") + "\n\n")

	b.WriteString("  Full signature puzzle:\n\n")
	b.WriteString(goodStyle.Render("  ┌──────────────┬──────────────┐") + "\n")
	b.WriteString(goodStyle.Render("  │  partial_A   │  partial_B   │") + "\n")
	b.WriteString(goodStyle.Render("  │  (from A)    │  (from B)    │") + "\n")
	b.WriteString(goodStyle.Render("  └──────────────┴──────────────┘") + "\n")
	b.WriteString(goodStyle.Render("           ↓  combine  ↓") + "\n")
	b.WriteString(goodStyle.Render("       ✓  Valid ECDSA Signature") + "\n\n")

	b.WriteString("  Attacker's attempt (missing one piece):\n\n")
	b.WriteString(redStyle.Render("  ┌──────────────┬──────────────┐") + "\n")
	b.WriteString(redStyle.Render("  │  partial_A   │    ??????    │") + "\n")
	b.WriteString(redStyle.Render("  │  (from A)    │   MISSING    │") + "\n")
	b.WriteString(redStyle.Render("  └──────────────┴──────────────┘") + "\n")
	b.WriteString(redStyle.Render("           ↓  combine  ↓") + "\n")
	b.WriteString(redStyle.Render("       ✗  Invalid / Unverifiable") + "\n\n")

	b.WriteString(dimStyle.Render("  Without both partial signatures the final (r, s) pair fails verification.") + "\n")
	return b.String()
}

// renderAttackerView shows what an attacker can and cannot see.
func (s *ImpossibilityScene) renderAttackerView() string {
	var b strings.Builder

	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("226"))
	redStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("196"))
	greenStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("82"))
	dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("243"))

	b.WriteString(titleStyle.Render("Attacker's View of the Protocol") + "\n\n")

	b.WriteString(greenStyle.Render("  What the attacker CAN observe:") + "\n")
	b.WriteString(dimStyle.Render("    • Public key Q = share_a·G + share_b·G") + "\n")
	b.WriteString(dimStyle.Render("    • Commitment values exchanged in round 1") + "\n")
	b.WriteString(dimStyle.Render("    • OT messages (encrypted, information-theoretically hidden)") + "\n")
	b.WriteString(dimStyle.Render("    • Final signature (r, s)") + "\n\n")

	b.WriteString(redStyle.Render("  What the attacker CANNOT learn:") + "\n")
	b.WriteString(dimStyle.Render("    • share_a  — never transmitted in plaintext") + "\n")
	b.WriteString(dimStyle.Render("    • share_b  — never transmitted in plaintext") + "\n")
	b.WriteString(dimStyle.Render("    • nonce_a, nonce_b  — per-session, never reused") + "\n")
	b.WriteString(dimStyle.Render("    • private_key = share_a + share_b  — never reconstructed") + "\n\n")

	b.WriteString(redStyle.Render("  ⚠  Forging a new signature requires solving ECDLP:") + "\n")
	b.WriteString(dimStyle.Render("    Given Q, find k such that Q = k·G") + "\n")
	b.WriteString(dimStyle.Render("    This is believed computationally infeasible on secp256k1.") + "\n")
	return b.String()
}

// renderMathProof shows the mathematical argument for security.
func (s *ImpossibilityScene) renderMathProof() string {
	var b strings.Builder

	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("226"))
	mathStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("214"))
	dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("243"))

	b.WriteString(titleStyle.Render("Why The Math Prevents Forgery") + "\n\n")

	b.WriteString(mathStyle.Render("  ECDSA signing equation:") + "\n")
	b.WriteString(mathStyle.Render("    s = k⁻¹ · (h + r · privkey)  mod n") + "\n\n")

	b.WriteString(mathStyle.Render("  In DKLS the nonce k = k_a · k_b and privkey = a + b:") + "\n")
	b.WriteString(mathStyle.Render("    s = (k_a·k_b)⁻¹ · (h + r·(a+b))  mod n") + "\n\n")

	b.WriteString(dimStyle.Render("  Party A contributes:  k_a⁻¹  and  a") + "\n")
	b.WriteString(dimStyle.Render("  Party B contributes:  k_b⁻¹  and  b") + "\n")
	b.WriteString(dimStyle.Render("  Neither learns the other's values — MtA + OT ensure this.") + "\n\n")

	b.WriteString(mathStyle.Render("  To forge without share_b an attacker must know b, i.e. solve:") + "\n")
	b.WriteString(mathStyle.Render("    b·G = Q - a·G   →   discrete log problem  (infeasible)") + "\n\n")

	b.WriteString(dimStyle.Render("  Security reduces to the hardness of ECDLP on secp256k1,") + "\n")
	b.WriteString(dimStyle.Render("  the same assumption that secures Bitcoin and Ethereum.") + "\n")
	return b.String()
}

// renderConclusion summarises the security guarantees.
func (s *ImpossibilityScene) renderConclusion() string {
	var b strings.Builder

	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("226"))
	greenStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("82"))
	dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("243"))
	phantomStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("196"))

	b.WriteString(titleStyle.Render("Security Guarantees Summary") + "\n\n")

	b.WriteString("  ┌──────────────────────────────────────────────────────────┐\n")
	b.WriteString("  │                                                          │\n")
	b.WriteString("  │  " + greenStyle.Render("✓  Unforgeability") + "                                        │\n")
	b.WriteString("  │     No single party can forge a signature alone.         │\n")
	b.WriteString("  │                                                          │\n")
	b.WriteString("  │  " + greenStyle.Render("✓  Privacy of Secret Shares") + "                               │\n")
	b.WriteString("  │     OT + commitments prevent share leakage.              │\n")
	b.WriteString("  │                                                          │\n")
	b.WriteString("  │  " + greenStyle.Render("✓  Standard Verification") + "                                  │\n")
	b.WriteString("  │     Output is a normal ECDSA (r, s) pair.                │\n")
	b.WriteString("  │                                                          │\n")
	b.WriteString("  │  " + greenStyle.Render("✓  Nonce Freshness") + "                                        │\n")
	b.WriteString("  │     Per-session nonces prevent signature reuse attacks.  │\n")
	b.WriteString("  │                                                          │\n")
	b.WriteString("  └──────────────────────────────────────────────────────────┘\n\n")

	b.WriteString(phantomStyle.Render("  The phantom key (combined private key) is never materialised.") + "\n")
	b.WriteString(dimStyle.Render("  Security inherits directly from the ECDLP hardness assumption.") + "\n")
	return b.String()
}

// renderInteractiveForgery shows an interactive forgery attempt demonstration.
func (s *ImpossibilityScene) renderInteractiveForgery() string {
	var b strings.Builder

	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("226"))
	redStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("196"))
	greenStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("82"))
	dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("243"))
	cyanStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("81"))
	magentaStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("213"))

	b.WriteString(titleStyle.Render("Interactive Forgery Attempt") + "\n\n")

	b.WriteString(dimStyle.Render("  Press 'f' to attempt forging a signature with only Party A's share.") + "\n")
	b.WriteString(dimStyle.Render("  Watch what happens when the attacker tries to complete the signature...\n\n") + "\n")

	b.WriteString("  ┌────────────────────────────────────────────────────────────┐\n")
	b.WriteString("  │  ATTACKER'S FORGERY ATTEMPT                                │\n")
	b.WriteString("  └────────────────────────────────────────────────────────────┘\n\n")

	if !s.forgeryAttempted {
		// Initial state - waiting for user to attempt forgery
		b.WriteString("  Step 1: Attacker intercepts Party A's partial signature\n\n")
		b.WriteString(cyanStyle.Render("    partial_A = " + dimStyle.Render("a1b2c3d4e5f6a7b8...") + "\n\n"))
		b.WriteString("  Step 2: Attacker needs Party B's partial signature\n\n")
		b.WriteString(magentaStyle.Render("    partial_B = " + redStyle.Render("????????????????...") + "  (MISSING!)") + "\n\n")
		b.WriteString(redStyle.Render("  [Press 'f' to attempt forgery without partial_B]") + "\n")
	} else {
		// Forgery attempted - show the failure progression
		b.WriteString("  Step 1: Attacker has Party A's partial signature\n\n")
		b.WriteString(cyanStyle.Render("    partial_A = a1b2c3d4e5f6a7b8...") + "\n\n")

		if s.forgeryStep >= 1 {
			b.WriteString("  Step 2: Attacker tries to guess partial_B\n\n")
			b.WriteString(redStyle.Render("    partial_B = GUESSING...") + "\n\n")
		}

		if s.forgeryStep >= 2 {
			b.WriteString("  Step 3: Attacker attempts to combine\n\n")
			b.WriteString(redStyle.Render("    Combining partial_A + guessed_partial_B...") + "\n\n")
		}

		if s.forgeryStep >= 3 {
			b.WriteString("  ┌────────────────────────────────────────────────────────────┐\n")
			b.WriteString("  │  " + redStyle.Render("✗ FORGERY FAILED!") + "                                              │\n")
			b.WriteString("  └────────────────────────────────────────────────────────────┘\n\n")
			b.WriteString(redStyle.Render("    Error: Invalid signature - verification failed") + "\n\n")
			b.WriteString(dimStyle.Render("    The forged (r, s) pair does not satisfy:") + "\n")
			b.WriteString(dimStyle.Render("      s·G = h·G + r·Q") + "\n\n")
			b.WriteString(greenStyle.Render("    ✓ Security preserved - forgery is computationally infeasible!") + "\n")
		}
	}

	b.WriteString("\n")
	b.WriteString(dimStyle.Render("  Why it failed: Without share_b, the attacker cannot compute") + "\n")
	b.WriteString(dimStyle.Render("  the correct partial_B. Each guess has probability 1/2^256 of") + "\n")
	b.WriteString(dimStyle.Render("  being correct — effectively zero.") + "\n")

	return b.String()
}

// View implements tea.Model.
func (s *ImpossibilityScene) View() string {
	return s.Render()
}

// Narrator returns the narrator text for the current step.
func (s *ImpossibilityScene) Narrator() string {
	narrators := []string{
		"Each party holds one secret share, locked away in its own box. A share alone is uniformly random and reveals nothing about the combined private key. An attacker who steals one box gains no useful information.",
		"A valid ECDSA signature requires both partial contributions to be combined. Without Party B's piece the puzzle is incomplete — the resulting (r, s) pair will fail every standard verification check.",
		"An attacker watching the network sees only public keys, commitments, and encrypted OT messages. The secret shares and per-session nonces are never transmitted in plaintext, and the private key is never reconstructed anywhere.",
		"The security proof reduces to the Elliptic Curve Discrete Logarithm Problem. Forging a signature without a secret share is equivalent to solving ECDLP on secp256k1 — the same hardness assumption that secures Bitcoin and Ethereum.",
		"Try pressing 'f' to attempt a forgery. You'll see that without Party B's partial signature, the attacker cannot produce a valid (r, s) pair. Each guess has probability 1/2^256 of being correct — effectively zero.",
		"In summary: unforgeability, share privacy, standard ECDSA output, and fresh per-session nonces together guarantee that the DKLS protocol is secure against a malicious adversary controlling one party.",
	}
	if s.currentStep >= len(narrators) {
		return narrators[len(narrators)-1]
	}
	return narrators[s.currentStep]
}
