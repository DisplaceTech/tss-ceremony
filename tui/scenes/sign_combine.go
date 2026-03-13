package scenes

import (
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// CombineSigScene represents Scene 11: Signature Assembly
// It shows partial signatures converging and combining into the final signature.
type CombineSigScene struct {
	config   *Config
	styles   *Styles
	phase    int // 0: show_partials, 1: combining_animation, 2: show_result
	step     int
	started  bool
	duration time.Duration
}

// NewCombineSigScene creates a new signature assembly scene
func NewCombineSigScene(config *Config, styles *Styles) *CombineSigScene {
	return &CombineSigScene{
		config:   config,
		styles:   styles,
		phase:    0,
		step:     0,
		started:  false,
		duration: getStepDuration(config.Speed),
	}
}

// Init initializes the scene
func (s *CombineSigScene) Init() tea.Cmd {
	s.started = true
	s.phase = 0
	s.step = 0
	if s.config.FixedMode {
		ResetFixedCounter()
	}
	return tea.Tick(s.duration, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

// Update handles events in the scene
func (s *CombineSigScene) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return s, tea.Quit
		case "enter", "right", "l", " ", "n", "j", "down":
			return s, nil
		}

	case tickMsg:
		s.step++
		if s.phase == 0 && s.step >= 6 {
			s.phase = 1
			s.step = 0
		} else if s.phase == 1 && s.step >= 10 {
			s.phase = 2
			s.step = 0
		}
		if s.phase < 2 {
			return s, tea.Tick(s.duration, func(t time.Time) tea.Msg {
				return tickMsg(t)
			})
		}
	}
	return s, nil
}

// Render renders the scene view
func (s *CombineSigScene) Render() string {
	var builder strings.Builder

	// Header
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("226")).
		MarginBottom(1)

	builder.WriteString(headerStyle.Render("Scene 11: Signature Assembly"))
	builder.WriteString("\n\n")

	separatorStyle := lipgloss.NewStyle().
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

	// Phase 0: Show the partial signatures
	if s.phase >= 0 {
		builder.WriteString(partyAStyle.Render("  s_a (Party A)"))
		builder.WriteString("\n")
		builder.WriteString("  ")
		builder.WriteString(partyAStyle.Render(s.formatHex(s.config.Ceremony.PartialSigAHex)))
		builder.WriteString("\n\n")

		builder.WriteString(partyBStyle.Render("  s_b (Party B)"))
		builder.WriteString("\n")
		builder.WriteString("  ")
		builder.WriteString(partyBStyle.Render(s.formatHex(s.config.Ceremony.PartialSigBHex)))
		builder.WriteString("\n\n")
	}

	// Phase 1: Combining animation
	if s.phase == 1 {
		builder.WriteString(separatorStyle.Render(strings.Repeat("─", 60)))
		builder.WriteString("\n\n")

		// Show converging arrows
		padA := 10 - s.step
		padB := 10 - s.step
		if padA < 0 {
			padA = 0
		}
		if padB < 0 {
			padB = 0
		}

		builder.WriteString(strings.Repeat(" ", padA+2))
		builder.WriteString(partyAStyle.Render("s_a ──►"))
		builder.WriteString("\n")

		builder.WriteString(strings.Repeat(" ", 12))
		combiningDots := ""
		for i := 0; i < s.step; i++ {
			combiningDots += "◆"
		}
		builder.WriteString(sharedStyle.Render(combiningDots))
		builder.WriteString("\n")

		builder.WriteString(strings.Repeat(" ", padB+2))
		builder.WriteString(partyBStyle.Render("s_b ──►"))
		builder.WriteString("\n\n")

		builder.WriteString(sharedStyle.Render("  s = s_a + s_b  (mod n)"))
		builder.WriteString("\n\n")
	}

	// Phase 2: Show final result
	if s.phase == 2 {
		builder.WriteString(separatorStyle.Render(strings.Repeat("─", 60)))
		builder.WriteString("\n\n")

		builder.WriteString(sharedStyle.Render("  s = s_a + s_b  (mod n)"))
		builder.WriteString("\n\n")

		builder.WriteString(separatorStyle.Render(strings.Repeat("─", 60)))
		builder.WriteString("\n\n")

		builder.WriteString(sharedStyle.Render("  Final ECDSA Signature (r, s):"))
		builder.WriteString("\n\n")

		labelStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("252"))

		builder.WriteString(labelStyle.Render("  r = "))
		builder.WriteString(sharedStyle.Render(s.formatHex(s.config.Ceremony.SignatureRHex)))
		builder.WriteString("\n\n")

		builder.WriteString(labelStyle.Render("  s = "))
		builder.WriteString(sharedStyle.Render(s.formatHex(s.config.Ceremony.SignatureSHex)))
		builder.WriteString("\n\n")

		explanationStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("243"))

		builder.WriteString(explanationStyle.Render("This is a standard ECDSA signature, indistinguishable"))
		builder.WriteString("\n")
		builder.WriteString(explanationStyle.Render("from one produced by a single signer with the full key."))
		builder.WriteString("\n\n")
	}

	// Status
	var status string
	switch s.phase {
	case 0:
		status = "Preparing partial signatures..."
	case 1:
		status = "Combining partial signatures..."
	default:
		status = "Signature assembled!"
	}
	statusStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("243"))
	builder.WriteString(statusStyle.Render(status))
	builder.WriteString("\n\n")

	hintStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("243"))
	builder.WriteString(hintStyle.Render("Press Enter to continue..."))

	return builder.String()
}

// formatHex formats a hex string in 8-character groups
func (s *CombineSigScene) formatHex(hexStr string) string {
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
func (s *CombineSigScene) View() string {
	return s.Render()
}

// Narrator returns the narrator text for this scene
func (s *CombineSigScene) Narrator() string {
	switch s.phase {
	case 0:
		return "Both partial signatures are ready. Party A has s_a and Party B has s_b. These will now be combined to form the final ECDSA signature."
	case 1:
		return "The partial signatures are being combined using modular addition: s = s_a + s_b mod n. This is the beauty of the additive MtA shares — simple addition produces the correct result."
	default:
		return "The final signature (r, s) is now complete. This is a standard ECDSA signature on secp256k1 — the same format used by Bitcoin and Ethereum. No one can tell it was produced by two parties instead of one."
	}
}
